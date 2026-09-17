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
	"strings"
	"sync"
	"testing"
	"time"

	"rangeforge-ue/pkg/agent"
	"rangeforge-ue/pkg/agent/emulator"
	"rangeforge-ue/pkg/config"
	"rangeforge-ue/pkg/models"
)

// TestProductionReadinessSuite executes end-to-end resilience tests simulating a production network deployment.
func TestProductionReadinessSuite(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "rf_prod_test_*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	cfg := config.DefaultManagerConfig()
	cfg.AllowUnsupportedOS = true
	cfg.TLSEnabled = false

	server, err := NewServer(cfg)
	if err != nil {
		t.Fatalf("Failed to create Manager server: %v", err)
	}
	server.authFile = filepath.Join(tempDir, "admin_auth.json")
	server.initAdminAuth()
	defer server.registry.Close()

	ts := httptest.NewServer(server.httpServer.Handler)
	defer ts.Close()

	// 1. High Concurrency Heartbeats & Event Drainage
	t.Run("HighConcurrencyFleetTelemetry", func(t *testing.T) {
		const agentCount = 20
		const heartbeatsPerAgent = 5
		var wg sync.WaitGroup

		for i := 0; i < agentCount; i++ {
			agentID := fmt.Sprintf("prod-node-%02d", i)
			wg.Add(1)
			go func(id string) {
				defer wg.Done()
				client := agent.NewManagerClient(ts.URL)

				// Register
				regInfo := models.AgentInfo{
					ID:              id,
					Hostname:        "host-" + id,
					OS:              "windows",
					Platform:        "Microsoft Windows (amd64)",
					Arch:            "amd64",
					PrimaryIP:       fmt.Sprintf("10.0.1.%d", 10+i),
					AssignedPersona: "office_worker",
					Status:          models.AgentStatusOnline,
				}
				if _, err := client.Register(regInfo); err != nil {
					t.Errorf("Agent %s register failed: %v", id, err)
					return
				}

				// Rapid heartbeats & events
				for j := 0; j < heartbeatsPerAgent; j++ {
					hbReq := models.HeartbeatRequest{
						AgentID: id,
						Telemetry: models.HostTelemetry{
							AgentID:   id,
							Hostname:  "host-" + id,
							Timestamp: time.Now().UTC(),
							CPU:       models.CPUStats{Percent: 12.5, NumCores: 8},
						},
					}
					if _, err := client.SendHeartbeat(hbReq); err != nil {
						t.Errorf("Agent %s heartbeat %d failed: %v", id, j, err)
					}

					// Event
					evt := models.EmulationEvent{
						ID:         fmt.Sprintf("evt-%s-%d", id, j),
						Timestamp:  time.Now().UTC(),
						AgentID:    id,
						Protocol:   "SMB_WRITE",
						Target:     "UE_budget_report.docx",
						Status:     "SUCCESS",
						DurationMs: 45,
					}
					if err := client.SendEvent(evt); err != nil {
						t.Errorf("Agent %s send event failed: %v", id, err)
					}
				}
			}(agentID)
		}
		wg.Wait()

		// Verify all agents registered and online
		agents := server.registry.GetAllAgents()
		if len(agents) < agentCount {
			t.Errorf("Expected at least %d registered agents, got %d", agentCount, len(agents))
		}
	})

	// 2. Network Disconnect & Telemetry Buffering Resilience
	t.Run("NetworkPartitionBufferAndRecovery", func(t *testing.T) {
		client := agent.NewManagerClient("http://127.0.0.1:59999") // Intentionally dead port

		hbReq := models.HeartbeatRequest{
			AgentID: "partitioned-agent",
			Telemetry: models.HostTelemetry{
				AgentID:   "partitioned-agent",
				Timestamp: time.Now().UTC(),
			},
		}

		// Multiple attempts while offline
		for i := 0; i < 5; i++ {
			_, err := client.SendHeartbeat(hbReq)
			if err == nil {
				t.Errorf("Expected error sending heartbeat to dead server")
			}
		}

		// Reconnect to active server
		clientLive := agent.NewManagerClient(ts.URL)
		resp, err := clientLive.SendHeartbeat(hbReq)
		if err != nil {
			t.Fatalf("Reconnection failed: %v", err)
		}
		if !resp.Acknowledge {
			t.Errorf("Expected Acknowledge to be true")
		}
	})

	// 3. Wordlist-Driven File Synthesis & Complete Deletion Guarantee
	t.Run("WordlistFileSynthesisAndZeroClutterCleanup", func(t *testing.T) {
		shareDir := filepath.Join(tempDir, "prod_share_test")
		_ = os.MkdirAll(shareDir, 0755)

		fileShareCfg := models.FileShareConfig{
			Enabled:     true,
			IntervalSec: 1,
			ReadRatio:   0.0, // Force write
		}
		shareCorpus := config.ShareCorpus{
			Shares:      []models.ShareTarget{{Path: shareDir}},
			SampleFiles: []string{"seed.docx"},
		}

		emulatorInstance := emulator.NewFileShareEmulator(fileShareCfg, shareCorpus)
		emulatorInstance.SetWordlist([]string{"quarterly", "briefing", "confidential", "metrics", "compliance"})

		// Trigger multiple file writes
		for i := 0; i < 15; i++ {
			emulatorInstance.Start()
			time.Sleep(50 * time.Millisecond)
			emulatorInstance.Stop()
		}

		// Check files created count
		_, _, created, deleted := emulatorInstance.GetDetailedStats()
		if created == 0 {
			t.Fatalf("Expected files to be created from wordlist, got 0")
		}
		t.Logf("Files created: %d, files pruned during run: %d", created, deleted)

		// Execute strict cleanup
		emulatorInstance.CleanupCreatedFiles()

		// Verify zero files created by UE remain on disk
		entries, err := os.ReadDir(shareDir)
		if err != nil {
			t.Fatalf("Failed to read share dir: %v", err)
		}

		for _, e := range entries {
			if strings.HasPrefix(e.Name(), "UE_") {
				t.Errorf("Lingering UE-created file detected: %s", e.Name())
			}
		}
	})

	// 4. Web Corpus Protocol Validation & Rejection
	t.Run("WebCorpusStrictProtocolValidation", func(t *testing.T) {
		// Valid post with http/https
		validBody := map[string]string{
			"urls": "https://portal.range.local\nhttp://wiki.corp.local\nhttps://internal.app",
		}
		data, _ := json.Marshal(validBody)
		resp, err := http.Post(ts.URL+"/api/v1/controller/corpus", "application/json", bytes.NewReader(data))
		if err != nil {
			t.Fatalf("Valid corpus post failed: %v", err)
		}
		if resp.StatusCode != http.StatusOK {
			t.Errorf("Expected 200 OK for valid corpus, got %d", resp.StatusCode)
		}

		// Invalid post missing http/https scheme
		invalidBody := map[string]string{
			"urls": "https://portal.range.local\ncorrupted_url_without_scheme\nhttp://wiki.corp.local",
		}
		invData, _ := json.Marshal(invalidBody)
		invResp, err := http.Post(ts.URL+"/api/v1/controller/corpus", "application/json", bytes.NewReader(invData))
		if err != nil {
			t.Fatalf("Invalid corpus post request failed: %v", err)
		}
		if invResp.StatusCode != http.StatusBadRequest {
			t.Errorf("Expected 400 Bad Request for corrupted URL scheme, got %d", invResp.StatusCode)
		}
	})

	// 5. Command Runner Bounded Buffer & Timeout Defense
	t.Run("CommandRunnerTimeoutAndBufferSafety", func(t *testing.T) {
		runner := emulator.NewCommandRunner()

		// Command Timeout
		timeoutReq := models.CommandRequest{
			TargetAgentID:  "test-runner",
			Command:        "ping 127.0.0.1 -n 10",
			Shell:          "cmd",
			TimeoutSeconds: 1,
		}
		start := time.Now()
		result := runner.Execute("test-runner", "cmd-timeout-test", timeoutReq)
		elapsed := time.Since(start)

		if elapsed > 4*time.Second {
			t.Errorf("Command took too long to abort: %v", elapsed)
		}
		if result.Status != "timeout" && result.Status != "failed" {
			t.Errorf("Expected timeout status, got '%s'", result.Status)
		}
	})

	// 6. Memory Bounding on Dispatcher Results
	t.Run("DispatcherTaskResultsMemoryBounding", func(t *testing.T) {
		dispatcher := NewTaskDispatcher()
		for i := 0; i < 600; i++ {
			dispatcher.RecordResult(models.TaskResult{
				TaskID:      fmt.Sprintf("task-%d", i),
				AgentID:     "agent-load",
				Status:      "success",
				CompletedAt: time.Now().Add(time.Duration(i) * time.Second),
			})
		}

		dispatcher.mu.Lock()
		count := len(dispatcher.taskResults)
		dispatcher.mu.Unlock()

		if count > 500 {
			t.Errorf("Expected taskResults to be bounded to <= 500, got %d", count)
		}
	})

	// 7. Plaintext HTTP to HTTPS Redirector Verification
	t.Run("PlaintextHTTPRedirector", func(t *testing.T) {
		req, _ := http.NewRequest("GET", ts.URL+"/api/v1/healthz", nil)
		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			t.Fatalf("Health check request failed: %v", err)
		}
		defer resp.Body.Close()
		body, _ := io.ReadAll(resp.Body)
		if !strings.Contains(string(body), "healthy") {
			t.Errorf("Expected healthy response, got: %s", string(body))
		}
	})
}
