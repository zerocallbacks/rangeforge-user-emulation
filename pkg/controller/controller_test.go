package controller

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"runtime"
	"testing"
	"time"

	"rangeforge-ue/pkg/models"
)

func TestControllerValidateEnvironment(t *testing.T) {
	err := ValidateEnvironment(false)
	if err != nil {
		t.Errorf("Expected no error on supported platform %s, got %v", runtime.GOOS, err)
	}
	errOverride := ValidateEnvironment(true)
	if errOverride != nil {
		t.Errorf("Expected no error when allowUnsupportedOS is true, got %v", errOverride)
	}
}

func TestControllerClient(t *testing.T) {
	mockAgents := []models.AgentInfo{
		{
			ID:              "agent-test-1",
			Hostname:        "win11-corp-01",
			OS:              "windows",
			PrimaryIP:       "192.168.10.55",
			AssignedPersona: "office_worker",
			Status:          models.AgentStatusOnline,
			LastHeartbeat:   time.Now().UTC(),
		},
	}

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		switch r.URL.Path {
		case "/api/v1/controller/agents":
			json.NewEncoder(w).Encode(map[string]interface{}{
				"count":  len(mockAgents),
				"agents": mockAgents,
			})
		case "/api/v1/controller/telemetry":
			json.NewEncoder(w).Encode(models.HostTelemetry{
				AgentID:   "agent-test-1",
				Hostname:  "win11-corp-01",
				Timestamp: time.Now().UTC(),
				CPU:       models.CPUStats{Percent: 24.5, NumCores: 8},
			})
		case "/api/v1/controller/command":
			json.NewEncoder(w).Encode(map[string]string{
				"task_id": "cmd-test-1234",
				"status":  "queued",
			})
		case "/api/v1/controller/command/status":
			json.NewEncoder(w).Encode(models.TaskResult{
				TaskID:   "cmd-test-1234",
				AgentID:  "agent-test-1",
				Status:   "success",
				ExitCode: 0,
				Stdout:   "RF-TEST-OUTPUT\n",
			})
		case "/api/v1/controller/persona":
			json.NewEncoder(w).Encode(map[string]string{"status": "updated"})
		case "/api/v1/controller/personas":
			json.NewEncoder(w).Encode(map[string]models.PersonaProfile{
				"office_worker": {Name: "office_worker", Description: "Office Worker"},
			})
		default:
			http.NotFound(w, r)
		}
	}))
	defer ts.Close()

	client := NewClient(ts.URL, 5)

	// Test GetAgents
	agents, err := client.GetAgents()
	if err != nil {
		t.Fatalf("GetAgents failed: %v", err)
	}
	if len(agents) != 1 || agents[0].Hostname != "win11-corp-01" {
		t.Errorf("Unexpected agents list: %+v", agents)
	}

	// Test GetTelemetry
	tel, err := client.GetTelemetry("agent-test-1")
	if err != nil {
		t.Fatalf("GetTelemetry failed: %v", err)
	}
	if tel.CPU.Percent != 24.5 {
		t.Errorf("Expected 24.5%% CPU, got %.1f", tel.CPU.Percent)
	}

	// Test SendCommand
	taskID, err := client.SendCommand("agent-test-1", "whoami", "powershell", 10)
	if err != nil {
		t.Fatalf("SendCommand failed: %v", err)
	}
	if taskID != "cmd-test-1234" {
		t.Errorf("Expected task ID cmd-test-1234, got %s", taskID)
	}

	// Test GetCommandStatus
	status, err := client.GetCommandStatus(taskID)
	if err != nil {
		t.Fatalf("GetCommandStatus failed: %v", err)
	}
	if status == nil || status.Stdout != "RF-TEST-OUTPUT\n" {
		t.Errorf("Unexpected command status: %+v", status)
	}

	// Test SetPersona
	err = client.SetPersona("agent-test-1", "developer")
	if err != nil {
		t.Fatalf("SetPersona failed: %v", err)
	}

	// Test GetPersonas
	personas, err := client.GetPersonas()
	if err != nil {
		t.Fatalf("GetPersonas failed: %v", err)
	}
	if len(personas) == 0 {
		t.Errorf("Expected non-empty personas map")
	}

	// Test CLI rendering
	cli := NewCLI(client)
	if err := cli.ListAgents(); err != nil {
		t.Errorf("ListAgents rendering failed: %v", err)
	}
	if err := cli.ShowHostStats("agent-test-1"); err != nil {
		t.Errorf("ShowHostStats rendering failed: %v", err)
	}
	if err := cli.ListPersonas(); err != nil {
		t.Errorf("ListPersonas rendering failed: %v", err)
	}
}
