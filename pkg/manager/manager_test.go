package manager

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"

	"rangeforge-ue/pkg/config"
	"rangeforge-ue/pkg/models"
)

func TestMain(m *testing.M) {
	code := m.Run()
	_ = os.RemoveAll("corpus")
	_ = os.RemoveAll("configs")
	os.Exit(code)
}

// setupTestServer initializes an isolated Manager server instance with test configuration.
func setupTestServer(t *testing.T) (*Server, *httptest.Server, func()) {
	t.Helper()
	tempDir, err := os.MkdirTemp("", "rf_test_*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}

	cfg := config.DefaultManagerConfig()
	cfg.AllowUnsupportedOS = true
	cfg.ProfilesDir = filepath.Join(tempDir, "profiles")
	_ = os.MkdirAll(cfg.ProfilesDir, 0755)

	srv, err := NewServer(cfg)
	if err != nil {
		os.RemoveAll(tempDir)
		t.Fatalf("NewServer failed: %v", err)
	}

	ts := httptest.NewServer(srv.httpServer.Handler)

	cleanup := func() {
		ts.Close()
		os.RemoveAll(tempDir)
	}

	return srv, ts, cleanup
}

// -----------------------------------------------------------------------------
// 1. ENVIRONMENT VALIDATION FUNCTION (3 TEST CASES)
// -----------------------------------------------------------------------------
func TestValidateEnvironment(t *testing.T) {
	// Case 1: Validate environment succeeds natively without requiring allowUnsupportedOS
	t.Run("NativePlatformValidation", func(t *testing.T) {
		err := ValidateEnvironment(false)
		if err != nil {
			t.Errorf("Expected no error on supported platform %s, got: %v", runtime.GOOS, err)
		}
	})

	// Case 2: Permit execution when allowUnsupportedOS is true
	t.Run("PermitWithAllowUnsupportedOS", func(t *testing.T) {
		err := ValidateEnvironment(true)
		if err != nil {
			t.Errorf("Expected no error with allowUnsupportedOS=true, got: %v", err)
		}
	})

	// Case 3: Verify consistent idempotency across repeated invocations
	t.Run("IdempotentValidation", func(t *testing.T) {
		for i := 0; i < 3; i++ {
			err := ValidateEnvironment(true)
			if err != nil {
				t.Fatalf("Iteration %d: expected no error, got: %v", i, err)
			}
		}
	})
}

// -----------------------------------------------------------------------------
// 2. AGENT REGISTRATION FUNCTION (3 TEST CASES)
// -----------------------------------------------------------------------------
func TestAgentRegistration(t *testing.T) {
	_, ts, cleanup := setupTestServer(t)
	defer cleanup()

	// Case 1: Standard Windows Agent registration with explicit persona
	t.Run("WindowsAgentWithPersona", func(t *testing.T) {
		regReq := models.RegisterRequest{
			Agent: models.AgentInfo{
				ID:              "agent-win-01",
				Hostname:        "DESKTOP-MIL-01",
				OS:              "windows",
				Platform:        "Windows 11 Enterprise",
				Arch:            "amd64",
				PrimaryIP:       "192.168.68.101",
				AssignedPersona: "office_worker",
			},
		}
		data, _ := json.Marshal(regReq)
		resp, err := http.Post(ts.URL+"/api/v1/agent/register", "application/json", bytes.NewReader(data))
		if err != nil || resp.StatusCode != http.StatusOK {
			t.Fatalf("Register failed: err=%v, code=%d", err, resp.StatusCode)
		}
		defer resp.Body.Close()

		var res models.RegisterResponse
		json.NewDecoder(resp.Body).Decode(&res)
		if !res.Success || res.AssignedPersona != "office_worker" {
			t.Errorf("Expected success and office_worker persona, got: %+v", res)
		}
	})

	// Case 2: Linux / pfSense Agent registration with default persona fallback
	t.Run("LinuxAgentDefaultPersonaFallback", func(t *testing.T) {
		regReq := models.RegisterRequest{
			Agent: models.AgentInfo{
				ID:        "agent-lin-02",
				Hostname:  "srv-ops-02",
				OS:        "linux",
				Platform:  "Ubuntu 24.04 LTS",
				Arch:      "amd64",
				PrimaryIP: "192.168.68.102",
			},
		}
		data, _ := json.Marshal(regReq)
		resp, err := http.Post(ts.URL+"/api/v1/agent/register", "application/json", bytes.NewReader(data))
		if err != nil || resp.StatusCode != http.StatusOK {
			t.Fatalf("Register failed: err=%v, code=%d", err, resp.StatusCode)
		}
		defer resp.Body.Close()

		var res models.RegisterResponse
		json.NewDecoder(resp.Body).Decode(&res)
		if !res.Success || res.AssignedPersona == "" {
			t.Errorf("Expected fallback persona assignment, got: %+v", res)
		}
	})

	// Case 3: Re-registration idempotency & metadata update
	t.Run("ReregistrationIdempotency", func(t *testing.T) {
		regReq := models.RegisterRequest{
			Agent: models.AgentInfo{
				ID:        "agent-win-01",
				Hostname:  "DESKTOP-MIL-01-RENAMED",
				OS:        "windows",
				PrimaryIP: "192.168.68.199",
			},
		}
		data, _ := json.Marshal(regReq)
		resp, err := http.Post(ts.URL+"/api/v1/agent/register", "application/json", bytes.NewReader(data))
		if err != nil || resp.StatusCode != http.StatusOK {
			t.Fatalf("Re-register failed: err=%v, code=%d", err, resp.StatusCode)
		}
		resp.Body.Close()

		// Verify updated metadata in controller listing
		listResp, err := http.Get(ts.URL + "/api/v1/controller/agents")
		if err != nil {
			t.Fatalf("List agents failed: %v", err)
		}
		defer listResp.Body.Close()

		var listResult struct {
			Agents []models.AgentInfo `json:"agents"`
		}
		json.NewDecoder(listResp.Body).Decode(&listResult)

		var found bool
		for _, a := range listResult.Agents {
			if a.ID == "agent-win-01" {
				found = true
				if a.Hostname != "DESKTOP-MIL-01-RENAMED" || a.PrimaryIP != "192.168.68.199" {
					t.Errorf("Expected updated hostname and IP, got: %+v", a)
				}
			}
		}
		if !found {
			t.Errorf("Re-registered agent not found in controller listing")
		}
	})
}

// -----------------------------------------------------------------------------
// 3. AGENT HEARTBEAT & TELEMETRY FUNCTION (3 TEST CASES)
// -----------------------------------------------------------------------------
func TestAgentHeartbeatAndTelemetry(t *testing.T) {
	_, ts, cleanup := setupTestServer(t)
	defer cleanup()

	// Register agent first
	regReq := models.RegisterRequest{
		Agent: models.AgentInfo{
			ID:       "agent-hb-01",
			Hostname: "node-c2-hb",
			OS:       "windows",
		},
	}
	regData, _ := json.Marshal(regReq)
	resp, _ := http.Post(ts.URL+"/api/v1/agent/register", "application/json", bytes.NewReader(regData))
	resp.Body.Close()

	// Case 1: Full hardware telemetry reporting (CPU, RAM, Disk, Network)
	t.Run("ReportFullHardwareTelemetry", func(t *testing.T) {
		hbReq := models.HeartbeatRequest{
			AgentID: "agent-hb-01",
			Telemetry: models.HostTelemetry{
				AgentID:   "agent-hb-01",
				Hostname:  "node-c2-hb",
				Timestamp: time.Now().UTC(),
				CPU: models.CPUStats{
					Percent:  35.5,
					NumCores: 8,
				},
				Memory: models.MemoryStats{
					TotalBytes: 16 * 1024 * 1024 * 1024,
					UsedBytes:  8 * 1024 * 1024 * 1024,
					FreeBytes:  8 * 1024 * 1024 * 1024,
					Percent:    50.0,
				},
				Disk: models.DiskStats{
					MountPoint: "C:\\",
					TotalBytes: 500 * 1024 * 1024 * 1024,
					UsedBytes:  200 * 1024 * 1024 * 1024,
					FreeBytes:  300 * 1024 * 1024 * 1024,
					Percent:    40.0,
				},
				NetworkInterfaces: []models.NetworkInterfaceStats{
					{Name: "Ethernet0", IPv4: []string{"192.168.68.50"}, MAC: "00:15:5D:AA:BB:CC", IsUp: true},
				},
				EmulationStats: models.EmulationMetrics{
					WebRequestsTotal: 120,
					ShareOpsTotal:    45,
					CommandsExecuted: 3,
				},
			},
		}
		data, _ := json.Marshal(hbReq)
		hbResp, err := http.Post(ts.URL+"/api/v1/agent/heartbeat", "application/json", bytes.NewReader(data))
		if err != nil || hbResp.StatusCode != http.StatusOK {
			t.Fatalf("Heartbeat failed: err=%v, code=%d", err, hbResp.StatusCode)
		}
		defer hbResp.Body.Close()

		var hbRes models.HeartbeatResponse
		json.NewDecoder(hbResp.Body).Decode(&hbRes)
		if !hbRes.Acknowledge {
			t.Errorf("Expected heartbeat acknowledgment")
		}

		// Verify telemetry retrieval via Controller API
		telResp, err := http.Get(ts.URL + "/api/v1/controller/telemetry?agent_id=agent-hb-01")
		if err != nil || telResp.StatusCode != http.StatusOK {
			t.Fatalf("Get telemetry failed: %v", err)
		}
		defer telResp.Body.Close()

		var retrieved models.HostTelemetry
		json.NewDecoder(telResp.Body).Decode(&retrieved)
		if retrieved.CPU.Percent != 35.5 || retrieved.Disk.Percent != 40.0 {
			t.Errorf("Telemetry mismatch: %+v", retrieved)
		}
	})

	// Case 2: Heartbeat response delivering operational state & queued tasks
	t.Run("HeartbeatDeliversStateAndTasks", func(t *testing.T) {
		// Queue a command for agent-hb-01
		cmdReq := models.CommandRequest{
			TargetAgentID:  "agent-hb-01",
			Command:        "whoami",
			Shell:          "powershell",
			TimeoutSeconds: 15,
		}
		cmdData, _ := json.Marshal(cmdReq)
		cResp, err := http.Post(ts.URL+"/api/v1/controller/command", "application/json", bytes.NewReader(cmdData))
		if err != nil || cResp.StatusCode != http.StatusOK {
			t.Fatalf("Queue command failed: err=%v, code=%d", err, cResp.StatusCode)
		}
		cResp.Body.Close()

		// Heartbeat from agent
		hbReq := models.HeartbeatRequest{
			AgentID: "agent-hb-01",
			Telemetry: models.HostTelemetry{
				AgentID:   "agent-hb-01",
				Timestamp: time.Now().UTC(),
			},
		}
		data, _ := json.Marshal(hbReq)
		hbResp, err := http.Post(ts.URL+"/api/v1/agent/heartbeat", "application/json", bytes.NewReader(data))
		if err != nil {
			t.Fatalf("Heartbeat failed: %v", err)
		}
		defer hbResp.Body.Close()

		var res models.HeartbeatResponse
		json.NewDecoder(hbResp.Body).Decode(&res)
		if len(res.Tasks) != 1 {
			t.Fatalf("Expected 1 task delivered in heartbeat, got %d", len(res.Tasks))
		}
		if res.Tasks[0].Parameters["command"] != "whoami" {
			t.Errorf("Expected task command 'whoami', got: %v", res.Tasks[0].Parameters["command"])
		}
	})

	// Case 3: Heartbeat response delivering active persona profile & web corpus
	t.Run("HeartbeatDeliversCorpusAndPersonaProfile", func(t *testing.T) {
		hbReq := models.HeartbeatRequest{
			AgentID: "agent-hb-01",
			Telemetry: models.HostTelemetry{
				AgentID:   "agent-hb-01",
				Timestamp: time.Now().UTC(),
			},
		}
		data, _ := json.Marshal(hbReq)
		hbResp, err := http.Post(ts.URL+"/api/v1/agent/heartbeat", "application/json", bytes.NewReader(data))
		if err != nil {
			t.Fatalf("Heartbeat failed: %v", err)
		}
		defer hbResp.Body.Close()

		var res models.HeartbeatResponse
		json.NewDecoder(hbResp.Body).Decode(&res)
		if res.Persona == "" {
			t.Errorf("Expected assigned persona in heartbeat response")
		}
	})
}

// -----------------------------------------------------------------------------
// 4. SINGLE COMMAND DISPATCH & EXECUTION (3 TEST CASES)
// -----------------------------------------------------------------------------
func TestSingleCommandDispatch(t *testing.T) {
	_, ts, cleanup := setupTestServer(t)
	defer cleanup()

	// Register test agent
	regReq := models.RegisterRequest{
		Agent: models.AgentInfo{ID: "agent-cmd-01", Hostname: "node-cmd", OS: "windows"},
	}
	regData, _ := json.Marshal(regReq)
	r, _ := http.Post(ts.URL+"/api/v1/agent/register", "application/json", bytes.NewReader(regData))
	r.Body.Close()

	var taskID string

	// Case 1: Dispatch valid administrative command
	t.Run("DispatchValidCommand", func(t *testing.T) {
		cmdReq := models.CommandRequest{
			TargetAgentID:  "agent-cmd-01",
			Command:        "ipconfig /all",
			Shell:          "powershell",
			TimeoutSeconds: 20,
		}
		data, _ := json.Marshal(cmdReq)
		resp, err := http.Post(ts.URL+"/api/v1/controller/command", "application/json", bytes.NewReader(data))
		if err != nil || resp.StatusCode != http.StatusOK {
			t.Fatalf("Dispatch failed: err=%v, code=%d", err, resp.StatusCode)
		}
		defer resp.Body.Close()

		var res map[string]string
		json.NewDecoder(resp.Body).Decode(&res)
		taskID = res["task_id"]
		if taskID == "" || res["status"] != "queued" {
			t.Fatalf("Expected queued status and valid task_id, got: %+v", res)
		}
	})

	// Case 2: Reject dispatch missing target or command
	t.Run("RejectInvalidCommandPayload", func(t *testing.T) {
		badReq := models.CommandRequest{Command: ""}
		data, _ := json.Marshal(badReq)
		resp, err := http.Post(ts.URL+"/api/v1/controller/command", "application/json", bytes.NewReader(data))
		if err != nil || resp.StatusCode != http.StatusBadRequest {
			t.Errorf("Expected 400 Bad Request for empty command, got: %d", resp.StatusCode)
		}
		resp.Body.Close()
	})

	// Case 3: Report task completion result and poll status
	t.Run("ReportResultAndPollStatus", func(t *testing.T) {
		if taskID == "" {
			t.Fatalf("No taskID from previous test")
		}

		// Agent reports result
		resReq := models.TaskResult{
			TaskID:     taskID,
			AgentID:    "agent-cmd-01",
			Status:     "success",
			ExitCode:   0,
			Stdout:     "Windows IP Configuration...",
			DurationMs: 45,
		}
		resData, _ := json.Marshal(resReq)
		resp, err := http.Post(ts.URL+"/api/v1/agent/task_result", "application/json", bytes.NewReader(resData))
		if err != nil || resp.StatusCode != http.StatusOK {
			t.Fatalf("Report result failed: err=%v, code=%d", err, resp.StatusCode)
		}
		resp.Body.Close()

		// Controller polls status
		statusResp, err := http.Get(ts.URL + "/api/v1/controller/command/status?task_id=" + taskID)
		if err != nil || statusResp.StatusCode != http.StatusOK {
			t.Fatalf("Poll status failed: %v", err)
		}
		defer statusResp.Body.Close()

		var pollResult models.TaskResult
		json.NewDecoder(statusResp.Body).Decode(&pollResult)
		if pollResult.Status != "success" || pollResult.ExitCode != 0 || pollResult.Stdout != "Windows IP Configuration..." {
			t.Errorf("Status mismatch: %+v", pollResult)
		}
	})
}

// -----------------------------------------------------------------------------
// 5. BATCH COMMAND DISPATCH FUNCTION (3 TEST CASES)
// -----------------------------------------------------------------------------
func TestBatchCommandDispatch(t *testing.T) {
	_, ts, cleanup := setupTestServer(t)
	defer cleanup()

	// Register 3 agents: 2 Windows, 1 Linux
	agents := []models.AgentInfo{
		{ID: "win-node-1", Hostname: "win-1", OS: "windows"},
		{ID: "win-node-2", Hostname: "win-2", OS: "windows"},
		{ID: "lin-node-1", Hostname: "lin-1", OS: "linux"},
	}
	for _, a := range agents {
		data, _ := json.Marshal(models.RegisterRequest{Agent: a})
		r, _ := http.Post(ts.URL+"/api/v1/agent/register", "application/json", bytes.NewReader(data))
		r.Body.Close()
	}

	// Case 1: Broadcast command to ALL agents
	t.Run("BroadcastToAllAgents", func(t *testing.T) {
		batchReq := map[string]interface{}{
			"target_scope":    "ALL",
			"command":         "hostname",
			"shell":           "auto",
			"timeout_seconds": 15,
		}
		data, _ := json.Marshal(batchReq)
		resp, err := http.Post(ts.URL+"/api/v1/controller/commands/batch", "application/json", bytes.NewReader(data))
		if err != nil || resp.StatusCode != http.StatusOK {
			t.Fatalf("Batch dispatch failed: err=%v, code=%d", err, resp.StatusCode)
		}
		defer resp.Body.Close()

		var res struct {
			DispatchedCount int `json:"dispatched_count"`
			Tasks           []struct {
				AgentID string `json:"agent_id"`
				TaskID  string `json:"task_id"`
				Status  string `json:"status"`
			} `json:"tasks"`
		}
		json.NewDecoder(resp.Body).Decode(&res)
		if res.DispatchedCount != 3 || len(res.Tasks) != 3 {
			t.Errorf("Expected 3 dispatched tasks, got %d", res.DispatchedCount)
		}
	})

	// Case 2: Broadcast command to ALL_WINDOWS only
	t.Run("BroadcastToAllWindowsAgents", func(t *testing.T) {
		batchReq := map[string]interface{}{
			"target_scope":    "ALL_WINDOWS",
			"command":         "whoami",
			"shell":           "powershell",
			"timeout_seconds": 15,
		}
		data, _ := json.Marshal(batchReq)
		resp, err := http.Post(ts.URL+"/api/v1/controller/commands/batch", "application/json", bytes.NewReader(data))
		if err != nil || resp.StatusCode != http.StatusOK {
			t.Fatalf("Batch dispatch failed: err=%v, code=%d", err, resp.StatusCode)
		}
		defer resp.Body.Close()

		var res struct {
			DispatchedCount int `json:"dispatched_count"`
		}
		json.NewDecoder(resp.Body).Decode(&res)
		if res.DispatchedCount != 2 {
			t.Errorf("Expected 2 dispatched tasks for Windows nodes, got %d", res.DispatchedCount)
		}
	})

	// Case 3: Broadcast to explicit list of agent IDs
	t.Run("BroadcastToExplicitList", func(t *testing.T) {
		batchReq := map[string]interface{}{
			"target_agent_ids": []string{"win-node-1", "lin-node-1"},
			"command":          "uptime",
			"shell":            "auto",
			"timeout_seconds":  10,
		}
		data, _ := json.Marshal(batchReq)
		resp, err := http.Post(ts.URL+"/api/v1/controller/commands/batch", "application/json", bytes.NewReader(data))
		if err != nil || resp.StatusCode != http.StatusOK {
			t.Fatalf("Explicit batch dispatch failed: %v", err)
		}
		defer resp.Body.Close()

		var res struct {
			DispatchedCount int `json:"dispatched_count"`
		}
		json.NewDecoder(resp.Body).Decode(&res)
		if res.DispatchedCount != 2 {
			t.Errorf("Expected 2 dispatched tasks, got %d", res.DispatchedCount)
		}
	})
}

// -----------------------------------------------------------------------------
// 6. OPERATIONAL STATE MACHINE & KILLSWITCH (3 TEST CASES)
// -----------------------------------------------------------------------------
func TestOperationalStateMachineAndKillswitch(t *testing.T) {
	_, ts, cleanup := setupTestServer(t)
	defer cleanup()

	// Case 1: Transition from Standby to Running to Paused
	t.Run("StateTransitions", func(t *testing.T) {
		states := []string{"running", "paused", "stopped"}
		for _, st := range states {
			payload := map[string]string{"state": st}
			data, _ := json.Marshal(payload)
			resp, err := http.Post(ts.URL+"/api/v1/controller/ranges/state", "application/json", bytes.NewReader(data))
			if err != nil || resp.StatusCode != http.StatusOK {
				t.Fatalf("Set state '%s' failed: %v", st, err)
			}
			resp.Body.Close()

			// Query range state
			rResp, err := http.Get(ts.URL + "/api/v1/controller/ranges")
			if err != nil {
				t.Fatalf("Get ranges failed: %v", err)
			}
			var rangeData struct {
				CurrentState string `json:"current_state"`
			}
			json.NewDecoder(rResp.Body).Decode(&rangeData)
			rResp.Body.Close()

			if rangeData.CurrentState != st {
				t.Errorf("Expected state '%s', got '%s'", st, rangeData.CurrentState)
			}
		}
	})

	// Case 2: Emergency Killswitch Trigger and System Audit Event
	t.Run("EmergencyKillswitchHalt", func(t *testing.T) {
		payload := map[string]string{"state": "emergency_stop"}
		data, _ := json.Marshal(payload)
		resp, err := http.Post(ts.URL+"/api/v1/controller/ranges/state", "application/json", bytes.NewReader(data))
		if err != nil || resp.StatusCode != http.StatusOK {
			t.Fatalf("Trigger emergency halt failed: %v", err)
		}
		resp.Body.Close()

		// Verify system event recorded
		evResp, err := http.Get(ts.URL + "/api/v1/controller/events")
		if err != nil {
			t.Fatalf("Get events failed: %v", err)
		}
		defer evResp.Body.Close()

		var events []models.EmulationEvent
		json.NewDecoder(evResp.Body).Decode(&events)

		var haltEventFound bool
		for _, ev := range events {
			if ev.Protocol == "SYSTEM" && ev.Target == "FLEET-HALT" && ev.Status == "EMERGENCY_STOP" {
				haltEventFound = true
				break
			}
		}
		if !haltEventFound {
			t.Errorf("Expected SYSTEM FLEET-HALT audit event after emergency stop")
		}
	})

	// Case 3: Resume from Emergency Stop
	t.Run("ResumeFromEmergencyStop", func(t *testing.T) {
		payload := map[string]string{"state": "running"}
		data, _ := json.Marshal(payload)
		resp, err := http.Post(ts.URL+"/api/v1/controller/ranges/state", "application/json", bytes.NewReader(data))
		if err != nil || resp.StatusCode != http.StatusOK {
			t.Fatalf("Resume operations failed: %v", err)
		}
		resp.Body.Close()

		// Verify system event recorded
		evResp, err := http.Get(ts.URL + "/api/v1/controller/events")
		if err != nil {
			t.Fatalf("Get events failed: %v", err)
		}
		defer evResp.Body.Close()

		var events []models.EmulationEvent
		json.NewDecoder(evResp.Body).Decode(&events)

		var resumeEventFound bool
		for _, ev := range events {
			if ev.Protocol == "SYSTEM" && ev.Target == "FLEET-OPERATIONS" && ev.Status == "RESUMED" {
				resumeEventFound = true
				break
			}
		}
		if !resumeEventFound {
			t.Errorf("Expected SYSTEM FLEET-OPERATIONS RESUMED audit event")
		}
	})
}

// -----------------------------------------------------------------------------
// 7. ENCLAVE PARTITION SWITCHING (3 TEST CASES)
// -----------------------------------------------------------------------------
func TestEnclavePartitionSwitching(t *testing.T) {
	_, ts, cleanup := setupTestServer(t)
	defer cleanup()

	enclaves := []string{
		"Enclave-Alpha (Corporate)",
		"Enclave-Bravo (Industrial/SCADA)",
		"Enclave-Charlie (Transit & DMZ)",
	}

	for i, enc := range enclaves {
		t.Run(fmt.Sprintf("SwitchToEnclave_%d", i+1), func(t *testing.T) {
			payload := map[string]string{"enclave": enc}
			data, _ := json.Marshal(payload)
			resp, err := http.Post(ts.URL+"/api/v1/controller/ranges/state", "application/json", bytes.NewReader(data))
			if err != nil || resp.StatusCode != http.StatusOK {
				t.Fatalf("Set enclave failed: %v", err)
			}
			resp.Body.Close()

			rResp, err := http.Get(ts.URL + "/api/v1/controller/ranges")
			if err != nil {
				t.Fatalf("Get ranges failed: %v", err)
			}
			defer rResp.Body.Close()

			var rangeData struct {
				CurrentEnclave string                   `json:"current_enclave"`
				Ranges         []map[string]interface{} `json:"ranges"`
			}
			json.NewDecoder(rResp.Body).Decode(&rangeData)
			if rangeData.CurrentEnclave != enc {
				t.Errorf("Expected current_enclave '%s', got '%s'", enc, rangeData.CurrentEnclave)
			}
		})
	}
}

// -----------------------------------------------------------------------------
// 8. WEB CORPUS DYNAMIC MODELING (3 TEST CASES)
// -----------------------------------------------------------------------------
func TestWebCorpusDynamicModeling(t *testing.T) {
	_, ts, cleanup := setupTestServer(t)
	defer cleanup()

	// Case 1: GET default web corpus
	t.Run("GetDefaultWebCorpus", func(t *testing.T) {
		resp, err := http.Get(ts.URL + "/api/v1/controller/corpus")
		if err != nil || resp.StatusCode != http.StatusOK {
			t.Fatalf("Get corpus failed: %v", err)
		}
		defer resp.Body.Close()
		var corpus config.WebCorpus
		json.NewDecoder(resp.Body).Decode(&corpus)
	})

	// Case 2: POST update corpus and verify live broadcast
	t.Run("UpdateAndBroadcastCorpus", func(t *testing.T) {
		updated := config.WebCorpus{
			IntranetPortals: []string{"http://portal.range.local", "http://wiki.corp.local"},
			InternetSites:   []string{"https://en.wikipedia.org/wiki/Computer_security", "https://cisa.gov/"},
			SearchQueries:   []string{"range exercise", "syslog alert"},
		}
		data, _ := json.Marshal(updated)
		resp, err := http.Post(ts.URL+"/api/v1/controller/corpus", "application/json", bytes.NewReader(data))
		if err != nil || resp.StatusCode != http.StatusOK {
			t.Fatalf("Save corpus failed: %v", err)
		}
		resp.Body.Close()

		// Read back via controller
		getResp, err := http.Get(ts.URL + "/api/v1/controller/corpus")
		if err != nil {
			t.Fatalf("Get corpus failed: %v", err)
		}
		defer getResp.Body.Close()

		var verify config.WebCorpus
		json.NewDecoder(getResp.Body).Decode(&verify)
		if len(verify.IntranetPortals) != 2 || len(verify.InternetSites) != 2 {
			t.Errorf("Corpus persistence mismatch: %+v", verify)
		}
	})

	// Case 3: Safe handling of empty payload
	t.Run("SafeHandlingOfEmptyCorpus", func(t *testing.T) {
		empty := config.WebCorpus{}
		data, _ := json.Marshal(empty)
		resp, err := http.Post(ts.URL+"/api/v1/controller/corpus", "application/json", bytes.NewReader(data))
		if err != nil || resp.StatusCode != http.StatusOK {
			t.Fatalf("Empty corpus save failed: %v", err)
		}
		resp.Body.Close()
	})
}

// -----------------------------------------------------------------------------
// 9. SCENARIO WORDLIST MANAGEMENT (3 TEST CASES)
// -----------------------------------------------------------------------------
func TestScenarioWordlistManagement(t *testing.T) {
	_, ts, cleanup := setupTestServer(t)
	defer cleanup()

	// Case 1: GET wordlist
	t.Run("GetWordlist", func(t *testing.T) {
		resp, err := http.Get(ts.URL + "/api/v1/controller/wordlist")
		if err != nil || resp.StatusCode != http.StatusOK {
			t.Fatalf("Get wordlist failed: %v", err)
		}
		defer resp.Body.Close()
	})

	// Case 2: POST save new multiline keywords
	t.Run("SaveMultilineWordlist", func(t *testing.T) {
		newWords := "enterprise_opsec\nquarterly_audit\nrange_telemetry\nransomware_drill"
		payload := map[string]string{"words": newWords}
		data, _ := json.Marshal(payload)
		resp, err := http.Post(ts.URL+"/api/v1/controller/wordlist", "application/json", bytes.NewReader(data))
		if err != nil || resp.StatusCode != http.StatusOK {
			t.Fatalf("Save wordlist failed: %v", err)
		}
		resp.Body.Close()

		// Read back
		getResp, err := http.Get(ts.URL + "/api/v1/controller/wordlist")
		if err != nil {
			t.Fatalf("Get wordlist failed: %v", err)
		}
		defer getResp.Body.Close()

		var res map[string]string
		json.NewDecoder(getResp.Body).Decode(&res)
		if res["words"] != newWords {
			t.Errorf("Wordlist mismatch: expected '%s', got '%s'", newWords, res["words"])
		}
	})

	// Case 3: Wordlist updates broadcast live across server state
	t.Run("WordlistBroadcastsInServer", func(t *testing.T) {
		// Register and heartbeat to verify wordlist is sent
		regData, _ := json.Marshal(models.RegisterRequest{Agent: models.AgentInfo{ID: "agent-wl-1", OS: "windows"}})
		r, _ := http.Post(ts.URL+"/api/v1/agent/register", "application/json", bytes.NewReader(regData))
		r.Body.Close()

		hbData, _ := json.Marshal(models.HeartbeatRequest{AgentID: "agent-wl-1"})
		hbResp, err := http.Post(ts.URL+"/api/v1/agent/heartbeat", "application/json", bytes.NewReader(hbData))
		if err != nil {
			t.Fatalf("Heartbeat failed: %v", err)
		}
		defer hbResp.Body.Close()

		var hbRes models.HeartbeatResponse
		json.NewDecoder(hbResp.Body).Decode(&hbRes)
		if !strings.Contains(hbRes.Wordlist, "enterprise_opsec") {
			t.Errorf("Expected updated wordlist in heartbeat response, got: %s", hbRes.Wordlist)
		}
	})
}

// -----------------------------------------------------------------------------
// 10. PERSONA PROFILE MODELING & ASSIGNMENT (3 TEST CASES)
// -----------------------------------------------------------------------------
func TestPersonaProfileModelingAndAssignment(t *testing.T) {
	_, ts, cleanup := setupTestServer(t)
	defer cleanup()

	// Case 1: GET all personas
	t.Run("GetAllPersonas", func(t *testing.T) {
		resp, err := http.Get(ts.URL + "/api/v1/controller/personas")
		if err != nil || resp.StatusCode != http.StatusOK {
			t.Fatalf("Get personas failed: %v", err)
		}
		defer resp.Body.Close()
		var personas map[string]models.PersonaProfile
		json.NewDecoder(resp.Body).Decode(&personas)
	})

	// Case 2: Save customized persona with modified parameters
	t.Run("SaveCustomizedPersona", func(t *testing.T) {
		custom := models.PersonaProfile{
			Name:        "c2_operator",
			Description: "Custom Enterprise Operations Lead",
			WebBrowsing: models.WebBrowsingConfig{
				Enabled:              true,
				RequestsPerMinuteMin: 10,
				RequestsPerMinuteMax: 30,
				DwellTimeMinSec:      5,
				DwellTimeMaxSec:      15,
				TargetURLs:           []string{"http://portal.range.local"},
			},
			FileShare: models.FileShareConfig{
				Enabled:     true,
				IntervalSec: 10,
				ReadRatio:   0.7,
				WriteRatio:  0.3,
			},
		}
		data, _ := json.Marshal(custom)
		resp, err := http.Post(ts.URL+"/api/v1/controller/persona/save", "application/json", bytes.NewReader(data))
		if err != nil || resp.StatusCode != http.StatusOK {
			t.Fatalf("Save persona failed: %v", err)
		}
		resp.Body.Close()

		// Verify retrieval
		getResp, err := http.Get(ts.URL + "/api/v1/controller/personas")
		if err != nil {
			t.Fatalf("Get personas failed: %v", err)
		}
		defer getResp.Body.Close()

		var personas map[string]models.PersonaProfile
		json.NewDecoder(getResp.Body).Decode(&personas)
		p, exists := personas["c2_operator"]
		if !exists || p.WebBrowsing.RequestsPerMinuteMax != 30 {
			t.Errorf("Persona save mismatch: %+v", p)
		}
	})

	// Case 3: Assign persona override to agent and verify heartbeat delivery
	t.Run("AssignPersonaOverrideAndVerifyHeartbeat", func(t *testing.T) {
		// Register agent
		regData, _ := json.Marshal(models.RegisterRequest{
			Agent: models.AgentInfo{ID: "agent-persona-test", OS: "windows", AssignedPersona: "office_worker"},
		})
		r, _ := http.Post(ts.URL+"/api/v1/agent/register", "application/json", bytes.NewReader(regData))
		r.Body.Close()

		// Override to c2_operator
		overrideReq := map[string]string{
			"agent_id": "agent-persona-test",
			"persona":  "c2_operator",
		}
		data, _ := json.Marshal(overrideReq)
		oResp, err := http.Post(ts.URL+"/api/v1/controller/persona", "application/json", bytes.NewReader(data))
		if err != nil || oResp.StatusCode != http.StatusOK {
			t.Fatalf("Persona override failed: %v", err)
		}
		oResp.Body.Close()

		// Heartbeat should return c2_operator
		hbData, _ := json.Marshal(models.HeartbeatRequest{AgentID: "agent-persona-test"})
		hbResp, err := http.Post(ts.URL+"/api/v1/agent/heartbeat", "application/json", bytes.NewReader(hbData))
		if err != nil {
			t.Fatalf("Heartbeat failed: %v", err)
		}
		defer hbResp.Body.Close()

		var hbRes models.HeartbeatResponse
		json.NewDecoder(hbResp.Body).Decode(&hbRes)
		if hbRes.Persona != "c2_operator" {
			t.Errorf("Expected c2_operator persona in heartbeat, got: %s", hbRes.Persona)
		}
	})
}

// -----------------------------------------------------------------------------
// 11. NETWORK TOPOLOGY & ROUTING CRUD (3 TEST CASES)
// -----------------------------------------------------------------------------
func TestNetworkTopologyAndGatewaysCRUD(t *testing.T) {
	_, ts, cleanup := setupTestServer(t)
	defer cleanup()

	// Case 1: Provision edge pfSense router
	t.Run("ProvisionEdgeRouter", func(t *testing.T) {
		topo := map[string]interface{}{
			"routers": []map[string]string{
				{
					"hostname":      "pfsense-edge-01.corp.local",
					"platform_role": "pfSense Firewall Gateway",
					"interfaces":    "WAN (em0), LAN (em1)",
					"control_ip":    "192.168.68.1",
					"external_ip":   "100.64.1.1",
					"routing_state": "active",
				},
			},
			"policies": []map[string]string{},
		}
		data, _ := json.Marshal(topo)
		resp, err := http.Post(ts.URL+"/api/v1/controller/topology", "application/json", bytes.NewReader(data))
		if err != nil || resp.StatusCode != http.StatusOK {
			t.Fatalf("Save topology failed: %v", err)
		}
		resp.Body.Close()

		// Read back
		getResp, err := http.Get(ts.URL + "/api/v1/controller/topology")
		if err != nil {
			t.Fatalf("Get topology failed: %v", err)
		}
		defer getResp.Body.Close()

		var verify struct {
			Routers []map[string]interface{} `json:"routers"`
		}
		json.NewDecoder(getResp.Body).Decode(&verify)
		if len(verify.Routers) != 1 || verify.Routers[0]["hostname"] != "pfsense-edge-01.corp.local" {
			t.Errorf("Topology router mismatch: %+v", verify)
		}
	})

	// Case 2: Define subnet reachability policy
	t.Run("DefineSubnetReachabilityPolicy", func(t *testing.T) {
		topo := map[string]interface{}{
			"routers": []map[string]string{
				{"hostname": "pfsense-edge-01.corp.local"},
			},
			"policies": []map[string]string{
				{
					"enclave_name":   "Enclave-Alpha (Corporate)",
					"subnet_cidr":    "192.168.68.0/24",
					"permitted_dest": "Internet (0.0.0.0/0)",
					"designated_dns": "1.1.1.1",
					"vocabulary":     "corp.local",
				},
			},
		}
		data, _ := json.Marshal(topo)
		resp, err := http.Post(ts.URL+"/api/v1/controller/topology", "application/json", bytes.NewReader(data))
		if err != nil || resp.StatusCode != http.StatusOK {
			t.Fatalf("Save policy failed: %v", err)
		}
		resp.Body.Close()

		// Verify policy
		getResp, err := http.Get(ts.URL + "/api/v1/controller/topology")
		if err != nil {
			t.Fatalf("Get topology failed: %v", err)
		}
		defer getResp.Body.Close()

		var verify struct {
			Policies []map[string]interface{} `json:"policies"`
		}
		json.NewDecoder(getResp.Body).Decode(&verify)
		if len(verify.Policies) != 1 || verify.Policies[0]["subnet_cidr"] != "192.168.68.0/24" {
			t.Errorf("Topology policy mismatch: %+v", verify)
		}
	})

	// Case 3: Clear/Delete topology items
	t.Run("DeleteTopologyItems", func(t *testing.T) {
		cleared := map[string]interface{}{
			"routers":  []string{},
			"policies": []string{},
		}
		data, _ := json.Marshal(cleared)
		resp, err := http.Post(ts.URL+"/api/v1/controller/topology", "application/json", bytes.NewReader(data))
		if err != nil || resp.StatusCode != http.StatusOK {
			t.Fatalf("Clear topology failed: %v", err)
		}
		resp.Body.Close()
	})
}

// -----------------------------------------------------------------------------
// 12. EXERCISE ACTION SCHEDULES CRUD (3 TEST CASES)
// -----------------------------------------------------------------------------
func TestActionSchedulesEngineCRUD(t *testing.T) {
	_, ts, cleanup := setupTestServer(t)
	defer cleanup()

	// Case 1: Create timed emulation schedule
	t.Run("CreateTimedSchedule", func(t *testing.T) {
		schedules := []map[string]interface{}{
			{
				"id":          "sched-day-01",
				"name":        "Daytime Normal Shift",
				"time_window": "08:00 - 17:00 UTC",
				"intensity":   1.5,
				"profile":     "office_worker",
				"enabled":     true,
				"status":      "active",
			},
		}
		data, _ := json.Marshal(schedules)
		resp, err := http.Post(ts.URL+"/api/v1/controller/schedules", "application/json", bytes.NewReader(data))
		if err != nil || resp.StatusCode != http.StatusOK {
			t.Fatalf("Save schedule failed: %v", err)
		}
		resp.Body.Close()

		// Read back
		getResp, err := http.Get(ts.URL + "/api/v1/controller/schedules")
		if err != nil {
			t.Fatalf("Get schedules failed: %v", err)
		}
		defer getResp.Body.Close()

		var verify []map[string]interface{}
		json.NewDecoder(getResp.Body).Decode(&verify)
		if len(verify) != 1 || verify[0]["id"] != "sched-day-01" {
			t.Errorf("Schedule mismatch: %+v", verify)
		}
	})

	// Case 2: Toggle schedule enabled state
	t.Run("ToggleScheduleState", func(t *testing.T) {
		schedules := []map[string]interface{}{
			{
				"id":          "sched-day-01",
				"name":        "Daytime Normal Shift",
				"time_window": "08:00 - 17:00 UTC",
				"intensity":   1.5,
				"profile":     "office_worker",
				"enabled":     false,
				"status":      "standby",
			},
		}
		data, _ := json.Marshal(schedules)
		resp, err := http.Post(ts.URL+"/api/v1/controller/schedules", "application/json", bytes.NewReader(data))
		if err != nil || resp.StatusCode != http.StatusOK {
			t.Fatalf("Toggle schedule failed: %v", err)
		}
		resp.Body.Close()

		getResp, err := http.Get(ts.URL + "/api/v1/controller/schedules")
		if err != nil {
			t.Fatalf("Get schedules failed: %v", err)
		}
		defer getResp.Body.Close()

		var verify []map[string]interface{}
		json.NewDecoder(getResp.Body).Decode(&verify)
		if verify[0]["enabled"] != false {
			t.Errorf("Expected enabled=false after toggle")
		}
	})

	// Case 3: Delete schedule
	t.Run("DeleteSchedule", func(t *testing.T) {
		empty := []interface{}{}
		data, _ := json.Marshal(empty)
		resp, err := http.Post(ts.URL+"/api/v1/controller/schedules", "application/json", bytes.NewReader(data))
		if err != nil || resp.StatusCode != http.StatusOK {
			t.Fatalf("Delete schedule failed: %v", err)
		}
		resp.Body.Close()

		getResp, err := http.Get(ts.URL + "/api/v1/controller/schedules")
		if err != nil {
			t.Fatalf("Get schedules failed: %v", err)
		}
		defer getResp.Body.Close()

		var verify []interface{}
		json.NewDecoder(getResp.Body).Decode(&verify)
		if len(verify) != 0 {
			t.Errorf("Expected empty schedule list after delete, got %d", len(verify))
		}
	})
}

// -----------------------------------------------------------------------------
// 13. SYNTHETIC IDENTITY DIRECTORY (3 TEST CASES)
// -----------------------------------------------------------------------------
func TestSyntheticIdentitiesDirectory(t *testing.T) {
	_, ts, cleanup := setupTestServer(t)
	defer cleanup()

	// Case 1: GET default identities
	t.Run("GetIdentities", func(t *testing.T) {
		resp, err := http.Get(ts.URL + "/api/v1/controller/identities")
		if err != nil || resp.StatusCode != http.StatusOK {
			t.Fatalf("Get identities failed: %v", err)
		}
		resp.Body.Close()
	})

	// Case 2: Save custom credentials list
	t.Run("SaveCustomIdentities", func(t *testing.T) {
		creds := []map[string]string{
			{"username": "secops_lead", "domain": "corp.local", "role": "Security Operations Lead"},
			{"username": "net_operator", "domain": "corp.local", "role": "Network Defense"},
		}
		data, _ := json.Marshal(creds)
		resp, err := http.Post(ts.URL+"/api/v1/controller/identities", "application/json", bytes.NewReader(data))
		if err != nil || resp.StatusCode != http.StatusOK {
			t.Fatalf("Save identities failed: %v", err)
		}
		resp.Body.Close()

		// Read back
		getResp, err := http.Get(ts.URL + "/api/v1/controller/identities")
		if err != nil {
			t.Fatalf("Get identities failed: %v", err)
		}
		defer getResp.Body.Close()

		var verify []map[string]string
		json.NewDecoder(getResp.Body).Decode(&verify)
		if len(verify) != 2 || verify[0]["username"] != "secops_lead" {
			t.Errorf("Identities mismatch: %+v", verify)
		}
	})

	// Case 3: Empty credentials list handling
	t.Run("EmptyIdentitiesHandling", func(t *testing.T) {
		empty := []interface{}{}
		data, _ := json.Marshal(empty)
		resp, err := http.Post(ts.URL+"/api/v1/controller/identities", "application/json", bytes.NewReader(data))
		if err != nil || resp.StatusCode != http.StatusOK {
			t.Fatalf("Empty identities failed: %v", err)
		}
		resp.Body.Close()
	})
}

// -----------------------------------------------------------------------------
// 14. EMULATION EVENTS STREAMING & FILTERING (3 TEST CASES)
// -----------------------------------------------------------------------------
func TestEmulationEventsStreamingAndFiltering(t *testing.T) {
	_, ts, cleanup := setupTestServer(t)
	defer cleanup()

	// Case 1: Ingest events across multiple protocols
	t.Run("IngestMultipleProtocolEvents", func(t *testing.T) {
		events := []models.EmulationEvent{
			{ID: "ev-1", Timestamp: time.Now().UTC(), AgentID: "win-01", Protocol: "HTTP", Target: "https://intranet.corp.local", Status: "SUCCESS", DurationMs: 120},
			{ID: "ev-2", Timestamp: time.Now().UTC(), AgentID: "win-01", Protocol: "SMB", Target: "\\\\server\\share\\budget.docx", Status: "SUCCESS", DurationMs: 45},
			{ID: "ev-3", Timestamp: time.Now().UTC(), AgentID: "lin-01", Protocol: "ICMP", Target: "8.8.8.8", Status: "SUCCESS", DurationMs: 12},
			{ID: "ev-4", Timestamp: time.Now().UTC(), AgentID: "lin-01", Protocol: "CMD", Target: "whoami", Status: "SUCCESS", DurationMs: 5},
		}

		for _, ev := range events {
			data, _ := json.Marshal(ev)
			resp, err := http.Post(ts.URL+"/api/v1/agent/event", "application/json", bytes.NewReader(data))
			if err != nil || resp.StatusCode != http.StatusOK {
				t.Fatalf("Post event failed: %v", err)
			}
			resp.Body.Close()
		}
	})

	// Case 2: Query events list and verify retrieval
	t.Run("QueryEventsList", func(t *testing.T) {
		resp, err := http.Get(ts.URL + "/api/v1/controller/events")
		if err != nil || resp.StatusCode != http.StatusOK {
			t.Fatalf("Get events failed: %v", err)
		}
		defer resp.Body.Close()

		var retrieved []models.EmulationEvent
		json.NewDecoder(resp.Body).Decode(&retrieved)
		if len(retrieved) < 4 {
			t.Errorf("Expected at least 4 events, got %d", len(retrieved))
		}
	})

	// Case 3: System state audit events
	t.Run("SystemStateEventsAudit", func(t *testing.T) {
		// Setting emergency_stop generates a SYSTEM event
		payload, _ := json.Marshal(map[string]string{"state": "emergency_stop"})
		r, _ := http.Post(ts.URL+"/api/v1/controller/ranges/state", "application/json", bytes.NewReader(payload))
		r.Body.Close()

		resp, err := http.Get(ts.URL + "/api/v1/controller/events")
		if err != nil {
			t.Fatalf("Get events failed: %v", err)
		}
		defer resp.Body.Close()

		var retrieved []models.EmulationEvent
		json.NewDecoder(resp.Body).Decode(&retrieved)

		var systemEventCount int
		for _, ev := range retrieved {
			if ev.Protocol == "SYSTEM" {
				systemEventCount++
			}
		}
		if systemEventCount == 0 {
			t.Errorf("Expected SYSTEM audit events recorded")
		}
	})
}

// -----------------------------------------------------------------------------
// 15. FLEET STORAGE CAPACITY METRICS (3 TEST CASES)
// -----------------------------------------------------------------------------
func TestFleetStorageCapacityCalculations(t *testing.T) {
	_, ts, cleanup := setupTestServer(t)
	defer cleanup()

	// Case 1: Zero-agent fleet reporting safe 0 GB / 0% without division-by-zero panics
	t.Run("ZeroAgentFleetStorageSafety", func(t *testing.T) {
		resp, err := http.Get(ts.URL + "/api/v1/controller/agents")
		if err != nil || resp.StatusCode != http.StatusOK {
			t.Fatalf("Get agents failed: %v", err)
		}
		defer resp.Body.Close()

		var res struct {
			FleetStorage struct {
				TotalGB float64 `json:"total_gb"`
				FreeGB  float64 `json:"free_gb"`
				UsedGB  float64 `json:"used_gb"`
				UsedPct float64 `json:"used_pct"`
			} `json:"fleet_storage"`
		}
		json.NewDecoder(resp.Body).Decode(&res)
		if res.FleetStorage.TotalGB != 0 || res.FleetStorage.UsedPct != 0 {
			t.Errorf("Expected 0 GB and 0%% for empty fleet, got: %+v", res.FleetStorage)
		}
	})

	// Case 2: Single agent storage reporting with accurate GB conversion and percent used
	t.Run("SingleAgentStorageCalculation", func(t *testing.T) {
		regReq := models.RegisterRequest{
			Agent: models.AgentInfo{ID: "agent-storage-01", Hostname: "box-1", OS: "windows"},
		}
		regData, _ := json.Marshal(regReq)
		r, _ := http.Post(ts.URL+"/api/v1/agent/register", "application/json", bytes.NewReader(regData))
		r.Body.Close()

		hbReq := models.HeartbeatRequest{
			AgentID: "agent-storage-01",
			Telemetry: models.HostTelemetry{
				AgentID:   "agent-storage-01",
				Timestamp: time.Now().UTC(),
				Disk: models.DiskStats{
					TotalBytes: 100 * 1024 * 1024 * 1024, // 100 GB
					UsedBytes:  25 * 1024 * 1024 * 1024,  // 25 GB
					FreeBytes:  75 * 1024 * 1024 * 1024,  // 75 GB
					Percent:    25.0,
				},
			},
		}
		hbData, _ := json.Marshal(hbReq)
		h, _ := http.Post(ts.URL+"/api/v1/agent/heartbeat", "application/json", bytes.NewReader(hbData))
		h.Body.Close()

		resp, err := http.Get(ts.URL + "/api/v1/controller/agents")
		if err != nil {
			t.Fatalf("Get agents failed: %v", err)
		}
		defer resp.Body.Close()

		var res struct {
			FleetStorage struct {
				TotalGB float64 `json:"total_gb"`
				FreeGB  float64 `json:"free_gb"`
				UsedGB  float64 `json:"used_gb"`
				UsedPct float64 `json:"used_pct"`
			} `json:"fleet_storage"`
		}
		json.NewDecoder(resp.Body).Decode(&res)
		if res.FleetStorage.TotalGB != 100.0 || res.FleetStorage.UsedGB != 25.0 || res.FleetStorage.UsedPct != 25.0 {
			t.Errorf("Storage metric calculation mismatch: %+v", res.FleetStorage)
		}
	})

	// Case 3: Multi-agent storage aggregation across heterogeneous filesystems
	t.Run("MultiAgentStorageAggregation", func(t *testing.T) {
		regReq2 := models.RegisterRequest{
			Agent: models.AgentInfo{ID: "agent-storage-02", Hostname: "box-2", OS: "linux"},
		}
		regData2, _ := json.Marshal(regReq2)
		r2, _ := http.Post(ts.URL+"/api/v1/agent/register", "application/json", bytes.NewReader(regData2))
		r2.Body.Close()

		hbReq2 := models.HeartbeatRequest{
			AgentID: "agent-storage-02",
			Telemetry: models.HostTelemetry{
				AgentID:   "agent-storage-02",
				Timestamp: time.Now().UTC(),
				Disk: models.DiskStats{
					TotalBytes: 300 * 1024 * 1024 * 1024, // 300 GB
					UsedBytes:  150 * 1024 * 1024 * 1024, // 150 GB
					FreeBytes:  150 * 1024 * 1024 * 1024, // 150 GB
					Percent:    50.0,
				},
			},
		}
		hbData2, _ := json.Marshal(hbReq2)
		h2, _ := http.Post(ts.URL+"/api/v1/agent/heartbeat", "application/json", bytes.NewReader(hbData2))
		h2.Body.Close()

		resp, err := http.Get(ts.URL + "/api/v1/controller/agents")
		if err != nil {
			t.Fatalf("Get agents failed: %v", err)
		}
		defer resp.Body.Close()

		var res struct {
			FleetStorage struct {
				TotalGB float64 `json:"total_gb"`
				FreeGB  float64 `json:"free_gb"`
				UsedGB  float64 `json:"used_gb"`
				UsedPct float64 `json:"used_pct"`
			} `json:"fleet_storage"`
		}
		json.NewDecoder(resp.Body).Decode(&res)

		// 100 + 300 = 400 GB total, 25 + 150 = 175 GB used
		if res.FleetStorage.TotalGB != 400.0 || res.FleetStorage.UsedGB != 175.0 {
			t.Errorf("Aggregated fleet storage mismatch: %+v", res.FleetStorage)
		}
	})
}

// -----------------------------------------------------------------------------
// 16. DASHBOARD HTTP & STATIC ASSET SERVING (3 TEST CASES)
// -----------------------------------------------------------------------------
func TestDashboardServingAndStaticAssets(t *testing.T) {
	_, ts, cleanup := setupTestServer(t)
	defer cleanup()

	// Case 1: GET `/` returns HTTP 200 with complete RangeForge HTML
	t.Run("GetDashboardRoot", func(t *testing.T) {
		resp, err := http.Get(ts.URL + "/")
		if err != nil || resp.StatusCode != http.StatusOK {
			t.Fatalf("GET / failed: err=%v, code=%d", err, resp.StatusCode)
		}
		defer resp.Body.Close()

		body, err := io.ReadAll(resp.Body)
		if err != nil {
			t.Fatalf("Read body failed: %v", err)
		}
		html := string(body)

		if !strings.Contains(html, "RANGEFORGE // COMMAND CENTER") {
			t.Errorf("Missing RangeForge Command Center Title in HTML")
		}
		if strings.Contains(strings.ToLower(html), "military") {
			t.Errorf("HTML must NOT contain any military references")
		}
		if !strings.Contains(html, "RANGEFORGE") {
			t.Errorf("Missing RangeForge branding in HTML")
		}
		if !strings.Contains(html, "HOST INVENTORY") {
			t.Errorf("Missing Host Inventory module in HTML")
		}
	})

	// Case 2: GET `/dashboard` returns identical dashboard HTML
	t.Run("GetDashboardPath", func(t *testing.T) {
		resp, err := http.Get(ts.URL + "/dashboard")
		if err != nil || resp.StatusCode != http.StatusOK {
			t.Fatalf("GET /dashboard failed: err=%v, code=%d", err, resp.StatusCode)
		}
		defer resp.Body.Close()

		body, _ := io.ReadAll(resp.Body)
		html := string(body)
		if !strings.Contains(html, "c2-main-viewport") {
			t.Errorf("Expected C2 main viewport container in /dashboard")
		}
	})

	// Case 3: GET `/assets/brand-icon.png` static asset delivery
	t.Run("GetBrandIconAsset", func(t *testing.T) {
		resp, err := http.Get(ts.URL + "/assets/brand-icon.png")
		if err != nil {
			t.Fatalf("GET brand icon failed: %v", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusNotFound {
			t.Errorf("Unexpected status code: %d", resp.StatusCode)
		}
		if resp.StatusCode == http.StatusOK {
			ct := resp.Header.Get("Content-Type")
			if ct != "image/png" {
				t.Errorf("Expected Content-Type image/png, got: %s", ct)
			}
		}
	})
}

func TestRangeOperationsPersistenceAndEnrichment(t *testing.T) {
	s, ts, cleanup := setupTestServer(t)
	defer cleanup()

	// Case 1: GET /api/v1/controller/ranges returns default ranges
	t.Run("GetDefaultRanges", func(t *testing.T) {
		resp, err := http.Get(ts.URL + "/api/v1/controller/ranges")
		if err != nil || resp.StatusCode != http.StatusOK {
			t.Fatalf("GET ranges failed: %v", err)
		}
		defer resp.Body.Close()

		var result struct {
			CurrentState string       `json:"current_state"`
			Ranges       []RangeModel `json:"ranges"`
		}
		if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
			t.Fatalf("Decode ranges failed: %v", err)
		}
		if len(result.Ranges) < 1 {
			t.Errorf("Expected at least 1 range, got %d", len(result.Ranges))
		}
		if !strings.Contains(result.Ranges[0].Name, "Range") {
			t.Errorf("Expected range name containing 'Range', got '%s'", result.Ranges[0].Name)
		}
	})

	// Case 2: POST /api/v1/controller/ranges persists custom range
	t.Run("PostCustomRange", func(t *testing.T) {
		newRange := RangeModel{
			ID:              "RANGE-5",
			Badge:           "CY",
			Name:            "Enterprise Evaluation Lab",
			State:           "stopped",
			EndpointsOnline: 3,
			EndpointsTotal:  6,
			Intensity:       "Medium",
			AvgCPU:          32.5,
			AvgRAM:          41.0,
			PrimaryCIDR:     "10.99.0.0/16",
		}
		data, _ := json.Marshal(newRange)
		resp, err := http.Post(ts.URL+"/api/v1/controller/ranges", "application/json", bytes.NewReader(data))
		if err != nil || resp.StatusCode != http.StatusOK {
			t.Fatalf("POST range failed: %v", err)
		}
		resp.Body.Close()

		// Verify on subsequent GET
		getResp, err := http.Get(ts.URL + "/api/v1/controller/ranges")
		if err != nil || getResp.StatusCode != http.StatusOK {
			t.Fatalf("Subsequent GET ranges failed: %v", err)
		}
		defer getResp.Body.Close()

		var getResult struct {
			Ranges []RangeModel `json:"ranges"`
		}
		json.NewDecoder(getResp.Body).Decode(&getResult)
		var found bool
		for _, r := range getResult.Ranges {
			if r.Name == "Enterprise Evaluation Lab" {
				found = true
				if r.PrimaryCIDR != "10.99.0.0/16" {
					t.Errorf("Expected CIDR 10.99.0.0/16, got %s", r.PrimaryCIDR)
				}
				break
			}
		}
		if !found {
			t.Errorf("Custom range was not found in persisted ranges")
		}
	})

	// Case 3: Live telemetry dynamic enrichment when agent reports
	t.Run("LiveAgentEnrichment", func(t *testing.T) {
		agent := models.AgentInfo{
			ID:        "agent-range-enrich-01",
			Hostname:  "test-primary-host",
			OS:        "windows",
			Status:    models.AgentStatusOnline,
			FirstSeen: time.Now().UTC(),
		}
		s.registry.Register(agent)
		_ = s.registry.UpdateHeartbeat(agent.ID, models.HostTelemetry{
			AgentID:   agent.ID,
			Hostname:  agent.Hostname,
			Timestamp: time.Now().UTC(),
			CPU:       models.CPUStats{Percent: 88.5},
			Memory:    models.MemoryStats{Percent: 62.0},
		})

		resp, err := http.Get(ts.URL + "/api/v1/controller/ranges")
		if err != nil || resp.StatusCode != http.StatusOK {
			t.Fatalf("GET ranges failed: %v", err)
		}
		defer resp.Body.Close()

		var enriched struct {
			Ranges []RangeModel `json:"ranges"`
		}
		json.NewDecoder(resp.Body).Decode(&enriched)
		if len(enriched.Ranges) == 0 {
			t.Fatalf("No ranges returned")
		}
		// Active single range should have updated CPU from live agent
		if enriched.Ranges[0].AvgCPU != 88.5 {
			t.Errorf("Expected live enriched CPU 88.5, got %v", enriched.Ranges[0].AvgCPU)
		}
	})
}

func TestIntensityConfiguration(t *testing.T) {
	_, ts, cleanup := setupTestServer(t)
	defer cleanup()

	// Case 1: GET default intensity
	t.Run("GetDefaultIntensity", func(t *testing.T) {
		resp, err := http.Get(ts.URL + "/api/v1/controller/intensity")
		if err != nil || resp.StatusCode != http.StatusOK {
			t.Fatalf("GET intensity failed: %v", err)
		}
		defer resp.Body.Close()

		var res map[string]string
		json.NewDecoder(resp.Body).Decode(&res)
		if res["intensity"] == "" {
			t.Errorf("Expected non-empty default intensity")
		}
	})

	// Case 2: POST update intensity
	t.Run("UpdateIntensity", func(t *testing.T) {
		payload := map[string]string{"intensity": "High (2.0x)"}
		data, _ := json.Marshal(payload)
		resp, err := http.Post(ts.URL+"/api/v1/controller/intensity", "application/json", bytes.NewReader(data))
		if err != nil || resp.StatusCode != http.StatusOK {
			t.Fatalf("POST intensity failed: %v", err)
		}
		resp.Body.Close()

		// Verify GET returns updated
		getResp, err := http.Get(ts.URL + "/api/v1/controller/intensity")
		if err != nil || getResp.StatusCode != http.StatusOK {
			t.Fatalf("Verify intensity failed: %v", err)
		}
		defer getResp.Body.Close()

		var res map[string]string
		json.NewDecoder(getResp.Body).Decode(&res)
		if res["intensity"] != "High (2.0x)" {
			t.Errorf("Expected 'High (2.0x)', got '%s'", res["intensity"])
		}
	})

	// Case 3: Ranges reflect updated intensity
	t.Run("RangesReflectUpdatedIntensity", func(t *testing.T) {
		resp, err := http.Get(ts.URL + "/api/v1/controller/ranges")
		if err != nil || resp.StatusCode != http.StatusOK {
			t.Fatalf("GET ranges failed: %v", err)
		}
		defer resp.Body.Close()

		var result struct {
			Intensity string       `json:"intensity"`
			Ranges    []RangeModel `json:"ranges"`
		}
		json.NewDecoder(resp.Body).Decode(&result)
		if result.Intensity != "High (2.0x)" {
			t.Errorf("Expected range controller intensity 'High (2.0x)', got '%s'", result.Intensity)
		}
		if len(result.Ranges) > 0 && result.Ranges[0].Intensity != "High" {
			t.Errorf("Expected range intensity 'High', got '%s'", result.Ranges[0].Intensity)
		}
	})
}

// -----------------------------------------------------------------------------
// 19. ADMIN AUTHENTICATION & CREDENTIAL MANAGEMENT (7 TEST CASES)
// -----------------------------------------------------------------------------
func TestAdminAuthenticationAndCredentialManagement(t *testing.T) {
	srv, ts, cleanup := setupTestServer(t)
	defer cleanup()

	// Case 1: Initial auth status returns unauthenticated
	t.Run("InitialAuthStatusUnauthenticated", func(t *testing.T) {
		resp, err := http.Get(ts.URL + "/api/v1/auth/status")
		if err != nil || resp.StatusCode != http.StatusOK {
			t.Fatalf("GET /api/v1/auth/status failed: %v", err)
		}
		defer resp.Body.Close()

		var res struct {
			Authenticated  bool   `json:"authenticated"`
			Username       string `json:"username"`
			ConfiguredUser string `json:"configured_user"`
		}
		json.NewDecoder(resp.Body).Decode(&res)
		if res.Authenticated {
			t.Errorf("Expected unauthenticated status initially")
		}
		if res.ConfiguredUser != "admin" {
			t.Errorf("Expected configured user 'admin', got '%s'", res.ConfiguredUser)
		}
	})

	// Case 2: Login with invalid password returns HTTP 401
	t.Run("LoginWithInvalidCredentials", func(t *testing.T) {
		payload := []byte(`{"username":"admin","password":"WrongPassword123!"}`)
		resp, err := http.Post(ts.URL+"/api/v1/auth/login", "application/json", bytes.NewBuffer(payload))
		if err != nil {
			t.Fatalf("POST /api/v1/auth/login failed: %v", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusUnauthorized {
			t.Errorf("Expected HTTP 401, got %d", resp.StatusCode)
		}
	})

	// Case 3: Login with valid credentials succeeds and returns session token + cookie
	var sessionToken string
	t.Run("LoginWithValidCredentials", func(t *testing.T) {
		payload := []byte(`{"username":"admin","password":"rangeforge"}`)
		resp, err := http.Post(ts.URL+"/api/v1/auth/login", "application/json", bytes.NewBuffer(payload))
		if err != nil {
			t.Fatalf("POST /api/v1/auth/login failed: %v", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			t.Fatalf("Expected HTTP 200, got %d", resp.StatusCode)
		}

		var res struct {
			Status   string `json:"status"`
			Token    string `json:"token"`
			Username string `json:"username"`
		}
		json.NewDecoder(resp.Body).Decode(&res)
		if res.Status != "authenticated" || res.Token == "" || res.Username != "admin" {
			t.Errorf("Unexpected login response: %+v", res)
		}
		sessionToken = res.Token

		// Check cookie
		hasCookie := false
		for _, c := range resp.Cookies() {
			if c.Name == "rf_auth_token" && c.Value == res.Token {
				hasCookie = true
			}
		}
		if !hasCookie {
			t.Errorf("Missing rf_auth_token cookie in login response")
		}
	})

	// Case 4: Auth status with valid token returns authenticated
	t.Run("AuthStatusWithValidToken", func(t *testing.T) {
		req, _ := http.NewRequest("GET", ts.URL+"/api/v1/auth/status", nil)
		req.Header.Set("Authorization", "Bearer "+sessionToken)
		resp, err := http.DefaultClient.Do(req)
		if err != nil || resp.StatusCode != http.StatusOK {
			t.Fatalf("GET /api/v1/auth/status failed: %v", err)
		}
		defer resp.Body.Close()

		var res struct {
			Authenticated bool   `json:"authenticated"`
			Username      string `json:"username"`
		}
		json.NewDecoder(resp.Body).Decode(&res)
		if !res.Authenticated || res.Username != "admin" {
			t.Errorf("Expected authenticated=true for user admin, got %+v", res)
		}
	})

	// Case 5: Change credentials with wrong current password fails
	t.Run("ChangeCredentialsWrongOldPassword", func(t *testing.T) {
		payload := []byte(`{
			"old_password": "IncorrectPassword!",
			"new_username": "secops_admin",
			"new_password": "NewSecretPass2026!"
		}`)
		resp, err := http.Post(ts.URL+"/api/v1/auth/change_credentials", "application/json", bytes.NewBuffer(payload))
		if err != nil {
			t.Fatalf("POST /api/v1/auth/change_credentials failed: %v", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusUnauthorized {
			t.Errorf("Expected HTTP 401 for wrong old password, got %d", resp.StatusCode)
		}
	})

	// Case 6: Change credentials with valid old password updates username & password
	t.Run("ChangeCredentialsSuccess", func(t *testing.T) {
		payload := []byte(`{
			"old_password": "rangeforge",
			"new_username": "rf_admin_lead",
			"new_password": "LeadEnterprise2026#"
		}`)
		resp, err := http.Post(ts.URL+"/api/v1/auth/change_credentials", "application/json", bytes.NewBuffer(payload))
		if err != nil {
			t.Fatalf("POST /api/v1/auth/change_credentials failed: %v", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			t.Fatalf("Expected HTTP 200, got %d", resp.StatusCode)
		}

		var res struct {
			Status   string `json:"status"`
			Username string `json:"username"`
			Token    string `json:"token"`
		}
		json.NewDecoder(resp.Body).Decode(&res)
		if res.Status != "success" || res.Username != "rf_admin_lead" || res.Token == "" {
			t.Errorf("Unexpected change_credentials response: %+v", res)
		}

		// Verify new login works with new credentials
		loginPayload := []byte(`{"username":"rf_admin_lead","password":"LeadEnterprise2026#"}`)
		loginResp, err := http.Post(ts.URL+"/api/v1/auth/login", "application/json", bytes.NewBuffer(loginPayload))
		if err != nil || loginResp.StatusCode != http.StatusOK {
			t.Fatalf("Login with new credentials failed: %v, code=%d", err, loginResp.StatusCode)
		}
		defer loginResp.Body.Close()

		// Verify old credentials now fail
		oldLoginPayload := []byte(`{"username":"admin","password":"rangeforge"}`)
		oldResp, err := http.Post(ts.URL+"/api/v1/auth/login", "application/json", bytes.NewBuffer(oldLoginPayload))
		if err != nil || oldResp.StatusCode != http.StatusUnauthorized {
			t.Errorf("Expected old credentials to fail with HTTP 401, got %d", oldResp.StatusCode)
		}
		defer oldResp.Body.Close()
	})

	// Case 7: Logout invalidates session
	t.Run("LogoutInvalidatesSession", func(t *testing.T) {
		req, _ := http.NewRequest("POST", ts.URL+"/api/v1/auth/logout", nil)
		req.Header.Set("Authorization", "Bearer "+sessionToken)
		resp, err := http.DefaultClient.Do(req)
		if err != nil || resp.StatusCode != http.StatusOK {
			t.Fatalf("POST /api/v1/auth/logout failed: %v", err)
		}
		defer resp.Body.Close()

		// Verify session token is no longer valid
		statusReq, _ := http.NewRequest("GET", ts.URL+"/api/v1/auth/status", nil)
		statusReq.Header.Set("Authorization", "Bearer "+sessionToken)
		statusResp, err := http.DefaultClient.Do(statusReq)
		if err != nil || statusResp.StatusCode != http.StatusOK {
			t.Fatalf("GET /api/v1/auth/status failed: %v", err)
		}
		defer statusResp.Body.Close()

		var statusRes struct {
			Authenticated bool `json:"authenticated"`
		}
		json.NewDecoder(statusResp.Body).Decode(&statusRes)
		if statusRes.Authenticated {
			t.Errorf("Session should be invalidated after logout")
		}
	})
	_ = srv
}

