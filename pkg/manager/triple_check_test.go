package manager

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"rangeforge-ue/pkg/agent/emulator"
	"rangeforge-ue/pkg/config"
	"rangeforge-ue/pkg/models"
)

// TestTripleCheckAPIJSONErrorResponses verifies that every API endpoint returns valid JSON on errors (no plain text).
func TestTripleCheckAPIJSONErrorResponses(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "rf_triple_check_*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	cfg := config.ManagerConfig{
		ListenHost:         "127.0.0.1",
		ListenPort:         8099,
		AllowUnsupportedOS: true,
		ProfilesDir:        filepath.Join(tempDir, "profiles"),
	}

	server, err := NewServer(cfg)
	if err != nil {
		t.Fatalf("failed to create server: %v", err)
	}

	endpoints := []struct {
		name       string
		method     string
		path       string
		body       string
		wantStatus int
	}{
		{
			name:       "RegisterMethodNotAllowed",
			method:     http.MethodGet,
			path:       "/api/v1/agent/register",
			body:       "",
			wantStatus: http.StatusMethodNotAllowed,
		},
		{
			name:       "RegisterMalformedJSON",
			method:     http.MethodPost,
			path:       "/api/v1/agent/register",
			body:       "{malformed_json",
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "RegisterEmptyAgentID",
			method:     http.MethodPost,
			path:       "/api/v1/agent/register",
			body:       `{"agent":{"id":""}}`,
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "HeartbeatEmptyAgentID",
			method:     http.MethodPost,
			path:       "/api/v1/agent/heartbeat",
			body:       `{"agent_id":""}`,
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "AuthLoginMethodNotAllowed",
			method:     http.MethodGet,
			path:       "/api/v1/auth/login",
			body:       "",
			wantStatus: http.StatusMethodNotAllowed,
		},
		{
			name:       "AuthLoginMalformedJSON",
			method:     http.MethodPost,
			path:       "/api/v1/auth/login",
			body:       "not a json",
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "ChangeCredentialsMethodNotAllowed",
			method:     http.MethodGet,
			path:       "/api/v1/auth/change_credentials",
			body:       "",
			wantStatus: http.StatusMethodNotAllowed,
		},
		{
			name:       "ControllerCommandMalformedJSON",
			method:     http.MethodPost,
			path:       "/api/v1/controller/command",
			body:       "{invalid",
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "ControllerCommandMissingArgs",
			method:     http.MethodPost,
			path:       "/api/v1/controller/command",
			body:       `{"target_agent_id":"","command":""}`,
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "ControllerTaskDispatchMalformedJSON",
			method:     http.MethodPost,
			path:       "/api/v1/controller/task_dispatch",
			body:       `bad-body`,
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "RangeStateMalformedJSON",
			method:     http.MethodPost,
			path:       "/api/v1/controller/range_state",
			body:       `{state:`,
			wantStatus: http.StatusBadRequest,
		},
	}

	for _, tc := range endpoints {
		t.Run(tc.name, func(t *testing.T) {
			req := httptest.NewRequest(tc.method, tc.path, strings.NewReader(tc.body))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()

			switch tc.path {
			case "/api/v1/agent/register":
				server.handleAgentRegister(w, req)
			case "/api/v1/agent/heartbeat":
				server.handleAgentHeartbeat(w, req)
			case "/api/v1/auth/login":
				server.handleAuthLogin(w, req)
			case "/api/v1/auth/change_credentials":
				server.handleAuthChangeCredentials(w, req)
			case "/api/v1/controller/command":
				server.handleControllerCommand(w, req)
			case "/api/v1/controller/task_dispatch":
				server.handleTaskDispatch(w, req)
			case "/api/v1/controller/range_state":
				server.handleControllerRangeState(w, req)
			}

			if w.Code != tc.wantStatus {
				t.Errorf("expected status %d, got %d. Body: %s", tc.wantStatus, w.Code, w.Body.String())
			}

			// Verify that response is valid JSON
			var jsonMap map[string]interface{}
			if err := json.Unmarshal(w.Body.Bytes(), &jsonMap); err != nil {
				t.Fatalf("response is not valid JSON! Error: %v. Raw Body: %s", err, w.Body.String())
			}

			if ct := w.Header().Get("Content-Type"); !strings.Contains(ct, "application/json") {
				t.Errorf("expected Content-Type application/json, got %s", ct)
			}
		})
	}
}

// TestTripleCheckCommandDispatchAllOnlineAgents tests dispatching to target ALL.
func TestTripleCheckCommandDispatchAllOnlineAgents(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "rf_dispatch_all_*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	cfg := config.ManagerConfig{
		ListenHost:         "127.0.0.1",
		ListenPort:         8099,
		AllowUnsupportedOS: true,
		ProfilesDir:        filepath.Join(tempDir, "profiles"),
	}

	server, err := NewServer(cfg)
	if err != nil {
		t.Fatalf("failed to create server: %v", err)
	}

	// Case 1: Dispatch to ALL when no agents are registered
	reqBody := `{"target_agent_id":"ALL","command":"whoami","timeout_seconds":10}`
	req := httptest.NewRequest(http.MethodPost, "/api/v1/controller/command", strings.NewReader(reqBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	server.handleControllerCommand(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200 OK, got %d: %s", w.Code, w.Body.String())
	}
	var resp map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &resp)
	if resp["status"] != "queued" {
		t.Errorf("expected status queued, got %v", resp["status"])
	}

	// Case 2: Register two agents and dispatch to ALL
	server.registry.Register(models.AgentInfo{
		ID:       "agent-fleet-01",
		Hostname: "host-01",
		Status:   models.AgentStatusOnline,
	})
	server.registry.Register(models.AgentInfo{
		ID:       "agent-fleet-02",
		Hostname: "host-02",
		Status:   models.AgentStatusOnline,
	})

	req2 := httptest.NewRequest(http.MethodPost, "/api/v1/controller/command", strings.NewReader(reqBody))
	req2.Header.Set("Content-Type", "application/json")
	w2 := httptest.NewRecorder()
	server.handleControllerCommand(w2, req2)

	if w2.Code != http.StatusOK {
		t.Fatalf("expected 200 OK, got %d: %s", w2.Code, w2.Body.String())
	}
	var resp2 map[string]interface{}
	json.Unmarshal(w2.Body.Bytes(), &resp2)

	taskIDs, ok := resp2["task_ids"].([]interface{})
	if !ok || len(taskIDs) != 2 {
		t.Fatalf("expected 2 task IDs dispatched, got %v", resp2["task_ids"])
	}

	// Fetch tasks for agent 1 and agent 2
	tasks1 := server.dispatcher.FetchTasks("agent-fleet-01")
	tasks2 := server.dispatcher.FetchTasks("agent-fleet-02")
	if len(tasks1) != 1 || len(tasks2) != 1 {
		t.Errorf("expected each agent to receive 1 task, got %d and %d", len(tasks1), len(tasks2))
	}
}

// TestTripleCheckBOMResilience ensures configs with UTF-8 BOM are loaded cleanly without parsing errors.
func TestTripleCheckBOMResilience(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "rf_bom_test_*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	bom := []byte("\xef\xbb\xbf")

	// 1. JSON Manager Config with BOM
	jsonContent := `{"listen_host":"127.0.0.1","listen_port":9090,"allow_unsupported_os":true}`
	jsonFile := filepath.Join(tempDir, "manager.json")
	if err := os.WriteFile(jsonFile, append(bom, []byte(jsonContent)...), 0644); err != nil {
		t.Fatalf("failed to write json with BOM: %v", err)
	}
	mCfg, err := config.LoadManagerConfig(jsonFile)
	if err != nil {
		t.Fatalf("LoadManagerConfig failed on BOM: %v", err)
	}
	if mCfg.ListenPort != 9090 {
		t.Errorf("expected port 9090, got %d", mCfg.ListenPort)
	}

	// 2. YAML Manager Config with BOM
	yamlContent := "listen_host: 127.0.0.1\nlisten_port: 9091\nallow_unsupported_os: true\n"
	yamlFile := filepath.Join(tempDir, "manager.yaml")
	if err := os.WriteFile(yamlFile, append(bom, []byte(yamlContent)...), 0644); err != nil {
		t.Fatalf("failed to write yaml with BOM: %v", err)
	}
	yCfg, err := config.LoadManagerConfig(yamlFile)
	if err != nil {
		t.Fatalf("LoadManagerConfig failed on YAML with BOM: %v", err)
	}
	if yCfg.ListenPort != 9091 {
		t.Errorf("expected port 9091, got %d", yCfg.ListenPort)
	}

	// 3. Web Corpus with BOM
	corpusContent := `{"intranet_portals":["http://bom.test.local"],"internet_sites":["https://docs.python.org/"]}`
	corpusFile := filepath.Join(tempDir, "web_corpus.json")
	if err := os.WriteFile(corpusFile, append(bom, []byte(corpusContent)...), 0644); err != nil {
		t.Fatalf("failed to write corpus with BOM: %v", err)
	}
	corpus, err := config.LoadWebCorpus(corpusFile)
	if err != nil {
		t.Fatalf("LoadWebCorpus failed on BOM: %v", err)
	}
	if len(corpus.IntranetPortals) != 1 || corpus.IntranetPortals[0] != "http://bom.test.local" {
		t.Errorf("unexpected corpus content: %v", corpus)
	}
}

// TestTripleCheckCommandOutputBounding tests that runaway outputs are capped at 512KB.
func TestTripleCheckCommandOutputBounding(t *testing.T) {
	runner := emulator.NewBoundedBuffer(1024) // 1KB test limit

	// Write 5KB of data
	chunk := bytes.Repeat([]byte("A"), 512)
	for i := 0; i < 10; i++ {
		_, _ = runner.Write(chunk)
	}

	result := runner.String()
	if !runner.Truncated {
		t.Errorf("expected buffer to be truncated")
	}
	if !strings.Contains(result, "[TRUNCATED") {
		t.Errorf("expected truncation message in output, got %s", result)
	}
	if runner.Buf.Len() > 1024 {
		t.Errorf("buffer exceeded max capacity: %d bytes", runner.Buf.Len())
	}
}

// TestTripleCheckSessionStoreExpiration tests session validity and expiration.
func TestTripleCheckSessionStoreExpiration(t *testing.T) {
	store := NewSessionStore()

	// Create session with 50ms expiration
	token := store.CreateSession("admin", 50*time.Millisecond)
	if token == "" {
		t.Fatalf("expected non-empty session token")
	}

	// Should be valid immediately
	user, ok := store.ValidateSession(token)
	if !ok || user != "admin" {
		t.Errorf("expected active session for admin, got valid=%v user=%s", ok, user)
	}

	// After 80ms, should be expired
	time.Sleep(80 * time.Millisecond)
	_, okExpired := store.ValidateSession(token)
	if okExpired {
		t.Errorf("expected expired session to be invalid")
	}
}

// TestTripleCheckNoiseControlAndPurge tests the Blue Team noise control and purge endpoints.
func TestTripleCheckNoiseControlAndPurge(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "rf_noise_test_*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	cfg := config.DefaultManagerConfig()
	cfg.ProfilesDir = filepath.Join(tempDir, "profiles")
	_ = os.MkdirAll(cfg.ProfilesDir, 0755)

	server, err := NewServer(cfg)
	if err != nil {
		t.Fatalf("failed to create server: %v", err)
	}

	// 1. Test GET /api/v1/controller/noise_control
	getReq := httptest.NewRequest(http.MethodGet, "/api/v1/controller/noise_control", nil)
	getRec := httptest.NewRecorder()
	server.handleControllerNoiseControl(getRec, getReq)

	if getRec.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", getRec.Code)
	}
	var getResp map[string]interface{}
	if err := json.Unmarshal(getRec.Body.Bytes(), &getResp); err != nil {
		t.Fatalf("failed to parse GET noise_control response: %v", err)
	}
	if getResp["noise_level"] != "balanced" {
		t.Errorf("expected default noise_level balanced, got %v", getResp["noise_level"])
	}
	if getResp["zero_footprint"] != true {
		t.Errorf("expected zero_footprint true")
	}

	// 2. Test POST /api/v1/controller/noise_control (Stealth Mode)
	postBody := `{"noise_level":"stealth","tool_interval_sec":180,"pause_tools":true,"heartbeat_sec":30}`
	postReq := httptest.NewRequest(http.MethodPost, "/api/v1/controller/noise_control", strings.NewReader(postBody))
	postReq.Header.Set("Content-Type", "application/json")
	postRec := httptest.NewRecorder()
	server.handleControllerNoiseControl(postRec, postReq)

	if postRec.Code != http.StatusOK {
		t.Fatalf("expected status 200 from POST noise_control, got %d: %s", postRec.Code, postRec.Body.String())
	}
	var postResp map[string]interface{}
	_ = json.Unmarshal(postRec.Body.Bytes(), &postResp)
	if postResp["noise_level"] != "stealth" {
		t.Errorf("expected noise_level stealth, got %v", postResp["noise_level"])
	}
	if postResp["pause_tools"] != true {
		t.Errorf("expected pause_tools true")
	}

	// 3. Test Agent Heartbeat receives updated noise control and heartbeat interval
	hbBody := `{"agent_id":"test-agent-noise","telemetry":{"agent_id":"test-agent-noise","hostname":"node1","emulation_stats":{"active_persona":"office_worker"}}}`
	hbReq := httptest.NewRequest(http.MethodPost, "/api/v1/agent/heartbeat", strings.NewReader(hbBody))
	hbReq.Header.Set("Content-Type", "application/json")
	hbRec := httptest.NewRecorder()
	server.handleAgentHeartbeat(hbRec, hbReq)

	if hbRec.Code != http.StatusOK {
		t.Fatalf("expected status 200 from heartbeat, got %d", hbRec.Code)
	}
	var hbResp models.HeartbeatResponse
	if err := json.Unmarshal(hbRec.Body.Bytes(), &hbResp); err != nil {
		t.Fatalf("failed to decode heartbeat response: %v", err)
	}
	if hbResp.HeartbeatSec != 30 {
		t.Errorf("expected HeartbeatSec 30, got %d", hbResp.HeartbeatSec)
	}
	if hbResp.PersonaProfile == nil {
		t.Fatalf("expected non-nil PersonaProfile in heartbeat response")
	}
	if hbResp.PersonaProfile.HostActivity.NoiseLevel != "stealth" {
		t.Errorf("expected PersonaProfile HostActivity.NoiseLevel stealth, got %s", hbResp.PersonaProfile.HostActivity.NoiseLevel)
	}
	if hbResp.PersonaProfile.HostActivity.PauseToolExecution != true {
		t.Errorf("expected PersonaProfile HostActivity.PauseToolExecution true")
	}

	// 4. Test POST /api/v1/controller/purge_artifacts
	purgeReq := httptest.NewRequest(http.MethodPost, "/api/v1/controller/purge_artifacts", nil)
	purgeRec := httptest.NewRecorder()
	server.handleControllerPurgeArtifacts(purgeRec, purgeReq)

	if purgeRec.Code != http.StatusOK {
		t.Fatalf("expected status 200 from purge_artifacts, got %d", purgeRec.Code)
	}
	if !server.purgeArtifactsRequested.Load() {
		t.Errorf("expected purgeArtifactsRequested to be true")
	}
}
