package manager

import (
	"bytes"
	"context"
	"crypto/tls"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"math"
	"net"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"sync"
	"sync/atomic"
	"syscall"
	"time"

	"rangeforge-ue/pkg/branding"
	"rangeforge-ue/pkg/config"
	"rangeforge-ue/pkg/models"
	"rangeforge-ue/pkg/tlsutil"
	"gopkg.in/yaml.v3"
)

// decodeJSONBody parses JSON from an HTTP request body while safely stripping any UTF-8 BOM.
func decodeJSONBody(r *http.Request, dst interface{}) error {
	data, err := io.ReadAll(r.Body)
	if err != nil {
		return err
	}
	data = bytes.TrimPrefix(data, []byte("\xef\xbb\xbf"))
	return json.Unmarshal(data, dst)
}

// writeJSONError sends a structured JSON error response ensuring clients and frontends never fail to parse responses.
func writeJSONError(w http.ResponseWriter, message string, statusCode int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	_ = json.NewEncoder(w).Encode(map[string]interface{}{
		"error":   message,
		"status":  statusCode,
		"success": false,
	})
}

// Server implements the RangeForge UE Manager central orchestrator.
type Server struct {
	config           config.ManagerConfig
	registry         *AgentRegistry
	dispatcher       *TaskDispatcher
	personas         map[string]models.PersonaProfile
	webCorpus        config.WebCorpus
	wordlist         string
	httpServer       *http.Server
	startTime        time.Time
	operationalState string
	activeRangeName  string
	activeRangeCIDR  string
	activeRangeID    string
	intensity        string
	stateMu          sync.RWMutex
	adminMu          sync.RWMutex
	adminAuth        AdminAccount
	sessions         *SessionStore
	authFile         string
	noiseMu          sync.RWMutex
	noiseLevel       string
	toolIntervalSec  int
	pauseToolExecution bool
	fleetHeartbeatSec int
	purgeArtifactsRequested atomic.Bool
}

// NewServer initializes a new Manager server instance.
func NewServer(cfg config.ManagerConfig) (*Server, error) {
	if err := ValidateEnvironment(cfg.AllowUnsupportedOS); err != nil {
		return nil, err
	}

	registry := NewAgentRegistry(cfg.HeartbeatTimeoutSec)
	dispatcher := NewTaskDispatcher()
	personas := config.LoadPersonasFromDir(cfg.ProfilesDir)
	corpusFile := filepath.Join("corpus", "web_corpus.txt")
	if _, err := os.Stat(corpusFile); err != nil {
		corpusFile = filepath.Join("corpus", "web_corpus.json")
	}
	initialCorpus, _ := config.LoadWebCorpus(corpusFile)

	initialWordlist := ""
	if wlData, err := os.ReadFile(filepath.Join("corpus", "wordlist.txt")); err == nil {
		initialWordlist = string(wlData)
	}

	s := &Server{
		config:           cfg,
		registry:         registry,
		dispatcher:       dispatcher,
		personas:         personas,
		webCorpus:        initialCorpus,
		wordlist:         initialWordlist,
		startTime:        time.Now().UTC(),
		operationalState: "stopped", // UE is initially on STANDBY; does not start until explicitly commanded!
		activeRangeName:  "Cyber Range",
		activeRangeCIDR:  "10.0.0.0/16",
		activeRangeID:    "RANGE-01",
		intensity:        "Low (0.25x)",
		sessions:         NewSessionStore(),
		noiseLevel:       "balanced",
		toolIntervalSec:  60,
		fleetHeartbeatSec: 5,
	}
	s.initAdminAuth()

	mux := http.NewServeMux()
	mux.HandleFunc("/assets/brand-icon.png", s.handleBrandIcon)
	mux.HandleFunc("/api/v1/agent/register", s.handleAgentRegister)
	mux.HandleFunc("/api/v1/agent/heartbeat", s.handleAgentHeartbeat)
	mux.HandleFunc("/api/v1/agent/task_result", s.handleAgentTaskResult)
	mux.HandleFunc("/api/v1/agent/event", s.handleAgentEvent)

	// Administrative Authentication Endpoints
	mux.HandleFunc("/api/v1/auth/login", s.handleAuthLogin)
	mux.HandleFunc("/api/v1/auth/logout", s.handleAuthLogout)
	mux.HandleFunc("/api/v1/auth/status", s.handleAuthStatus)
	mux.HandleFunc("/api/v1/auth/change_credentials", s.handleAuthChangeCredentials)

	mux.HandleFunc("/api/v1/controller/agents", s.handleControllerAgents)
	mux.HandleFunc("/api/v1/controller/telemetry", s.handleControllerTelemetry)
	mux.HandleFunc("/api/v1/controller/command", s.handleControllerCommand)
	mux.HandleFunc("/api/v1/controller/commands/batch", s.handleControllerCommandBatch)
	mux.HandleFunc("/api/v1/controller/command/status", s.handleControllerCommandStatus)
	mux.HandleFunc("/api/v1/controller/persona", s.handleControllerPersona)
	mux.HandleFunc("/api/v1/controller/personas", s.handleControllerPersonas)
	mux.HandleFunc("/api/v1/controller/persona/save", s.handleControllerPersonaSave)
	mux.HandleFunc("/api/v1/controller/noise_control", s.handleControllerNoiseControl)
	mux.HandleFunc("/api/v1/controller/noise", s.handleControllerNoiseControl)
	mux.HandleFunc("/api/v1/controller/purge_artifacts", s.handleControllerPurgeArtifacts)
	mux.HandleFunc("/api/v1/emulation/cleanup", s.handleControllerPurgeArtifacts)
	mux.HandleFunc("/api/v1/controller/corpus", s.handleControllerCorpus)
	mux.HandleFunc("/api/v1/controller/corpus/save", s.handleControllerCorpus)
	mux.HandleFunc("/api/v1/controller/wordlist", s.handleControllerWordlist)
	mux.HandleFunc("/api/v1/controller/events", s.handleControllerEvents)
	mux.HandleFunc("/api/v1/controller/topology", s.handleControllerTopology)
	mux.HandleFunc("/api/v1/controller/schedules", s.handleControllerSchedules)
	mux.HandleFunc("/api/v1/controller/identities", s.handleControllerIdentities)
	mux.HandleFunc("/api/v1/controller/ranges", s.handleControllerRanges)
	mux.HandleFunc("/api/v1/controller/ranges/state", s.handleControllerRangeState)
	mux.HandleFunc("/api/v1/controller/range_state", s.handleControllerRangeState)
	mux.HandleFunc("/api/v1/ranges/state", s.handleControllerRangeState)
	mux.HandleFunc("/api/v1/controller/range/metrics", s.handleRangeMetrics)
	mux.HandleFunc("/api/v1/range/metrics", s.handleRangeMetrics)
	mux.HandleFunc("/api/v1/controller/intensity", s.handleControllerIntensity)
	mux.HandleFunc("/api/v1/healthz", s.handleHealthCheck)
	mux.HandleFunc("/", s.handleDashboard)
	mux.HandleFunc("/dashboard", s.handleDashboard)
	mux.HandleFunc("/overview", s.handleDashboard)
	mux.HandleFunc("/fleet", s.handleDashboard)
	mux.HandleFunc("/emulation", s.handleDashboard)
	mux.HandleFunc("/configure", s.handleDashboard)
	mux.HandleFunc("/corpus", s.handleDashboard)
	mux.HandleFunc("/topology", s.handleDashboard)
	mux.HandleFunc("/activity", s.handleDashboard)

	// RESTful & UI compatibility aliases
	mux.HandleFunc("/api/v1/agents", s.handleControllerAgents)
	mux.HandleFunc("/api/v1/telemetry", s.handleControllerTelemetry)
	mux.HandleFunc("/api/v1/events", s.handleControllerEvents)
	mux.HandleFunc("/api/v1/ranges", s.handleControllerRanges)
	mux.HandleFunc("/api/v1/ranges/intensity", s.handleControllerIntensity)
	mux.HandleFunc("/api/v1/corpus/web", s.handleControllerCorpus)
	mux.HandleFunc("/api/v1/corpus/wordlist", s.handleControllerWordlist)
	mux.HandleFunc("/api/v1/profiles", s.handleControllerPersonas)
	mux.HandleFunc("/api/v1/identities", s.handleControllerIdentities)
	mux.HandleFunc("/api/v1/status", s.handleStatusSummary)
	mux.HandleFunc("/api/v1/controller/start", s.handleOperationalStateAction("running"))
	mux.HandleFunc("/api/v1/controller/stop", s.handleOperationalStateAction("stopped"))
	mux.HandleFunc("/api/v1/controller/pause", s.handleOperationalStateAction("paused"))
	mux.HandleFunc("/api/v1/controller/emergency_stop", s.handleOperationalStateAction("emergency_stop"))
	mux.HandleFunc("/api/v1/controller/set_persona", s.handleControllerPersona)
	mux.HandleFunc("/api/v1/controller/command_status", s.handleControllerCommandStatus)
	mux.HandleFunc("/api/v1/tasks/dispatch", s.handleTaskDispatch)
	mux.HandleFunc("/api/v1/tasks/results", s.handleTaskResults)

	secureHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Strict-Transport-Security", "max-age=63072000; includeSubDomains; preload")
		w.Header().Set("X-Content-Type-Options", "nosniff")
		w.Header().Set("X-Frame-Options", "SAMEORIGIN")
		w.Header().Set("X-XSS-Protection", "1; mode=block")
		mux.ServeHTTP(w, r)
	})

	addr := fmt.Sprintf("%s:%d", cfg.ListenHost, cfg.ListenPort)
	s.httpServer = &http.Server{
		Addr:         addr,
		Handler:      secureHandler,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
	}

	return s, nil
}

// Start launches the Manager HTTPS server (HTTPS-only enforced).
func (s *Server) Start(ctx context.Context) error {
	branding.PrintBanner("Manager Orchestrator")
	s.config.TLSEnabled = true
	protocol := "https"

	log.Printf("[MANAGER] Starting RangeForge UE Manager on %s://%s (HTTPS Only)", protocol, s.httpServer.Addr)
	log.Printf("[MANAGER] Loaded %d Persona Profiles (%s)", len(s.personas), strings.Join(s.getPersonaNames(), ", "))
	log.Printf("[MANAGER] Fleet Heartbeat Timeout: %ds", s.config.HeartbeatTimeoutSec)

	errChan := make(chan error, 1)

	certFile := s.config.TLSCertPath
	keyFile := s.config.TLSKeyPath
	if certFile == "" {
		certFile = filepath.Join("certs", "server.crt")
	}
	if keyFile == "" {
		keyFile = filepath.Join("certs", "server.key")
	}

	cert, err := tlsutil.EnsureCertificates(certFile, keyFile)
	if err != nil {
		return fmt.Errorf("initializing TLS certificates: %w", err)
	}
	log.Printf("[MANAGER] TLS/HTTPS active with certificate: %s", certFile)

	tlsConfig := &tls.Config{
		Certificates: []tls.Certificate{cert},
		MinVersion:   tls.VersionTLS12,
	}

	rawListener, err := net.Listen("tcp", s.httpServer.Addr)
	if err != nil {
		return fmt.Errorf("binding address %s: %w", s.httpServer.Addr, err)
	}

	smartLn := newSmartTLSListener(rawListener)
	tlsLn := tls.NewListener(smartLn.TLSListener(), tlsConfig)

	plainServer := &http.Server{
		Addr:         s.httpServer.Addr,
		Handler:      s.httpServer.Handler,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
	}

	go func() {
		if err := s.httpServer.Serve(tlsLn); err != nil && err != http.ErrServerClosed {
			errChan <- err
		}
	}()

	go func() {
		if err := plainServer.Serve(smartLn.PlainListener()); err != nil && err != http.ErrServerClosed {
			// Plain listener stopped
		}
	}()

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	select {
	case err := <-errChan:
		_ = plainServer.Close()
		_ = smartLn.Close()
		return fmt.Errorf("https server failed: %w", err)
	case <-ctx.Done():
		log.Printf("[MANAGER] Context cancelled, shutting down...")
	case sig := <-sigChan:
		log.Printf("[MANAGER] Received signal %v, shutting down...", sig)
	}

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	_ = plainServer.Shutdown(shutdownCtx)
	_ = s.httpServer.Shutdown(shutdownCtx)
	_ = smartLn.Close()
	if s.registry != nil {
		s.registry.Close()
	}
	return nil
}

func (s *Server) getPersonaNames() []string {
	var names []string
	for k := range s.personas {
		names = append(names, k)
	}
	return names
}

func (s *Server) handleAgentRegister(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeJSONError(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req models.RegisterRequest
	if err := decodeJSONBody(r, &req); err != nil {
		writeJSONError(w, "Invalid JSON: "+err.Error(), http.StatusBadRequest)
		return
	}

	if strings.TrimSpace(req.Agent.ID) == "" {
		writeJSONError(w, "agent_id is required", http.StatusBadRequest)
		return
	}

	s.registry.Register(req.Agent)
	log.Printf("[MANAGER] Registered Agent: %s (Host: %s, OS: %s, IP: %s)",
		req.Agent.ID, req.Agent.Hostname, req.Agent.OS, req.Agent.PrimaryIP)

	assignedPersona := s.dispatcher.GetPersona(req.Agent.ID)
	if assignedPersona == "" {
		if ag, ok := s.registry.GetAgent(req.Agent.ID); ok && ag.AssignedPersona != "" {
			assignedPersona = ag.AssignedPersona
		}
	}
	if assignedPersona == "" {
		assignedPersona = req.Agent.AssignedPersona
	}
	if assignedPersona == "" {
		assignedPersona = s.config.DefaultPersona
	}
	_ = s.registry.SetPersona(req.Agent.ID, assignedPersona)

	resp := models.RegisterResponse{
		Success:         true,
		AssignedPersona: assignedPersona,
		HeartbeatSec:    5,
		Message:         "Registered with RangeForge UE Manager v" + branding.Version,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func (s *Server) handleAgentHeartbeat(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeJSONError(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req models.HeartbeatRequest
	if err := decodeJSONBody(r, &req); err != nil {
		writeJSONError(w, "Invalid JSON: "+err.Error(), http.StatusBadRequest)
		return
	}

	if strings.TrimSpace(req.AgentID) == "" {
		writeJSONError(w, "agent_id is required", http.StatusBadRequest)
		return
	}

	if err := s.registry.UpdateHeartbeat(req.AgentID, req.Telemetry); err != nil {
		writeJSONError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	tasks := s.dispatcher.FetchTasks(req.AgentID)
	personaOverride := s.dispatcher.GetPersona(req.AgentID)

	targetPersonaName := personaOverride
	if targetPersonaName == "" {
		if ag, ok := s.registry.GetAgent(req.AgentID); ok && ag.AssignedPersona != "" {
			targetPersonaName = ag.AssignedPersona
		}
	}
	if targetPersonaName == "" {
		targetPersonaName = s.config.DefaultPersona
	}

	var activeProfile *models.PersonaProfile
	if p, exists := s.personas[targetPersonaName]; exists {
		pCopy := p
		s.noiseMu.RLock()
		pCopy.HostActivity.NoiseLevel = s.noiseLevel
		pCopy.HostActivity.ToolIntervalSec = s.toolIntervalSec
		pCopy.HostActivity.PauseToolExecution = s.pauseToolExecution
		pCopy.HostActivity.UserFileOperations = true
		s.noiseMu.RUnlock()
		activeProfile = &pCopy
	}

	s.stateMu.RLock()
	curState := s.operationalState
	var webPayload *models.WebCorpusPayload
	if len(s.webCorpus.InternetSites) > 0 || len(s.webCorpus.IntranetPortals) > 0 {
		webPayload = &models.WebCorpusPayload{
			IntranetPortals: s.webCorpus.IntranetPortals,
			InternetSites:   s.webCorpus.InternetSites,
			SearchQueries:   s.webCorpus.SearchQueries,
			StaticAssets:    s.webCorpus.StaticAssets,
		}
	}
	curWordlist := s.wordlist
	s.stateMu.RUnlock()

	s.noiseMu.RLock()
	hbSec := s.fleetHeartbeatSec
	s.noiseMu.RUnlock()

	purgeReq := s.purgeArtifactsRequested.Load()

	resp := models.HeartbeatResponse{
		Acknowledge:      true,
		OperationalState: curState,
		Persona:          targetPersonaName,
		PersonaProfile:   activeProfile,
		WebCorpus:        webPayload,
		Wordlist:         curWordlist,
		Tasks:            tasks,
		HeartbeatSec:     hbSec,
		PurgeArtifacts:   purgeReq,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func (s *Server) handleAgentTaskResult(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeJSONError(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var result models.TaskResult
	if err := decodeJSONBody(r, &result); err != nil {
		writeJSONError(w, "Invalid JSON: "+err.Error(), http.StatusBadRequest)
		return
	}

	s.dispatcher.RecordResult(result)
	log.Printf("[MANAGER] Recorded Result for Task %s (Agent: %s, Status: %s, Code: %d)",
		result.TaskID, result.AgentID, result.Status, result.ExitCode)

	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"status":"ok"}`))
}

func (s *Server) handleControllerAgents(w http.ResponseWriter, r *http.Request) {
	agents := s.registry.GetAllAgents()

	type EnrichedAgent struct {
		models.AgentInfo
		Telemetry *models.HostTelemetry `json:"telemetry,omitempty"`
	}

	var enriched []EnrichedAgent
	var totalFleetDiskBytes, freeFleetDiskBytes, usedFleetDiskBytes uint64

	for _, a := range agents {
		ea := EnrichedAgent{AgentInfo: a}
		if tel, ok := s.registry.GetLatestTelemetry(a.ID); ok {
			ea.Telemetry = tel
			totalFleetDiskBytes += tel.Disk.TotalBytes
			freeFleetDiskBytes += tel.Disk.FreeBytes
			usedFleetDiskBytes += tel.Disk.UsedBytes
		}
		enriched = append(enriched, ea)
	}

	totalFleetDiskGB := math.Round(float64(totalFleetDiskBytes)/(1024*1024*1024)*10) / 10
	freeFleetDiskGB := math.Round(float64(freeFleetDiskBytes)/(1024*1024*1024)*10) / 10
	usedFleetDiskGB := math.Round(float64(usedFleetDiskBytes)/(1024*1024*1024)*10) / 10
	diskUsedPct := 0.0
	if totalFleetDiskBytes > 0 {
		diskUsedPct = math.Round((float64(usedFleetDiskBytes)/float64(totalFleetDiskBytes))*1000) / 10
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"count":  len(enriched),
		"agents": enriched,
		"fleet_storage": map[string]interface{}{
			"total_gb": totalFleetDiskGB,
			"free_gb":  freeFleetDiskGB,
			"used_gb":  usedFleetDiskGB,
			"used_pct": diskUsedPct,
		},
	})
}

func (s *Server) handleControllerTelemetry(w http.ResponseWriter, r *http.Request) {
	agentID := r.URL.Query().Get("agent_id")
	if agentID == "" {
		writeJSONError(w, "Missing agent_id query parameter", http.StatusBadRequest)
		return
	}

	tel, exists := s.registry.GetLatestTelemetry(agentID)
	if !exists {
		writeJSONError(w, "Agent telemetry not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(tel)
}

func (s *Server) handleControllerCommand(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeJSONError(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var raw struct {
		models.CommandRequest
		AgentID string `json:"agent_id"`
	}
	if err := decodeJSONBody(r, &raw); err != nil {
		writeJSONError(w, "Invalid JSON: "+err.Error(), http.StatusBadRequest)
		return
	}
	cmdReq := raw.CommandRequest
	if cmdReq.TargetAgentID == "" && raw.AgentID != "" {
		cmdReq.TargetAgentID = raw.AgentID
	}

	if cmdReq.TargetAgentID == "" || cmdReq.Command == "" {
		writeJSONError(w, "target_agent_id (or agent_id) and command are required", http.StatusBadRequest)
		return
	}

	if strings.EqualFold(cmdReq.TargetAgentID, "ALL") {
		agents := s.registry.GetAllAgents()
		var targets []string
		for _, a := range agents {
			if a.Status == models.AgentStatusOnline {
				targets = append(targets, a.ID)
			}
		}
		if len(targets) == 0 {
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(map[string]interface{}{
				"status":  "queued",
				"message": "No online agents found for target ALL",
				"targets": []string{},
			})
			return
		}
		taskIDs, err := s.dispatcher.DispatchBatchCommand(targets, cmdReq)
		if err != nil {
			writeJSONError(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"status":   "queued",
			"task_ids": taskIDs,
			"targets":  targets,
		})
		return
	}

	taskID, err := s.dispatcher.DispatchCommand(cmdReq)
	if err != nil {
		writeJSONError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	log.Printf("[MANAGER] Dispatched Command %s to Agent %s: '%s'", taskID, cmdReq.TargetAgentID, cmdReq.Command)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"task_id":         taskID,
		"target_agent_id": cmdReq.TargetAgentID,
		"status":          "queued",
	})
}

func (s *Server) handleControllerCommandStatus(w http.ResponseWriter, r *http.Request) {
	taskID := r.URL.Query().Get("task_id")
	if taskID == "" {
		writeJSONError(w, "Missing task_id query parameter", http.StatusBadRequest)
		return
	}

	result, ok := s.dispatcher.GetResult(taskID)
	w.Header().Set("Content-Type", "application/json")
	if !ok {
		json.NewEncoder(w).Encode(map[string]string{
			"task_id": taskID,
			"status":  "pending",
		})
		return
	}

	json.NewEncoder(w).Encode(result)
}

func (s *Server) handleControllerCommandBatch(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeJSONError(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		TargetScope    string   `json:"target_scope"`     // "ALL", "ALL_WINDOWS", "ALL_LINUX", "ALL_FREEBSD", or empty
		Scope          string   `json:"scope"`            // alias for target_scope
		TargetAgentIDs []string `json:"target_agent_ids"` // explicit list if target_scope is empty or specific
		Command        string   `json:"command"`
		Cmd            string   `json:"cmd"`              // alias for command
		Shell          string   `json:"shell"`
		TimeoutSeconds int      `json:"timeout_seconds"`
	}
	if err := decodeJSONBody(r, &req); err != nil {
		writeJSONError(w, "Invalid JSON: "+err.Error(), http.StatusBadRequest)
		return
	}

	if req.TargetScope == "" && req.Scope != "" {
		req.TargetScope = req.Scope
	}
	if req.Command == "" && req.Cmd != "" {
		req.Command = req.Cmd
	}

	if req.Command == "" {
		writeJSONError(w, "command is required", http.StatusBadRequest)
		return
	}

	agents := s.registry.GetAllAgents()
	var targets []string

	scope := strings.ToUpper(strings.TrimSpace(req.TargetScope))
	switch scope {
	case "ALL":
		for _, a := range agents {
			if a.Status == models.AgentStatusOnline {
				targets = append(targets, a.ID)
			}
		}
	case "ALL_WINDOWS":
		for _, a := range agents {
			if a.Status == models.AgentStatusOnline && strings.Contains(strings.ToLower(a.OS), "win") {
				targets = append(targets, a.ID)
			}
		}
	case "ALL_LINUX":
		for _, a := range agents {
			if a.Status == models.AgentStatusOnline && strings.Contains(strings.ToLower(a.OS), "linux") {
				targets = append(targets, a.ID)
			}
		}
	case "ALL_FREEBSD":
		for _, a := range agents {
			if a.Status == models.AgentStatusOnline && strings.Contains(strings.ToLower(a.OS), "freebsd") {
				targets = append(targets, a.ID)
			}
		}
	default:
		targets = req.TargetAgentIDs
	}

	if len(targets) == 0 {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"dispatched_count": 0,
			"tasks":            []interface{}{},
			"message":          "No matching online agents found for target scope",
		})
		return
	}

	cmdReq := models.CommandRequest{
		Command:        req.Command,
		Shell:          req.Shell,
		TimeoutSeconds: req.TimeoutSeconds,
	}

	taskIDs, err := s.dispatcher.DispatchBatchCommand(targets, cmdReq)
	if err != nil {
		writeJSONError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	type BatchItem struct {
		AgentID string `json:"agent_id"`
		TaskID  string `json:"task_id"`
		Status  string `json:"status"`
	}
	var results []BatchItem
	for i, tID := range taskIDs {
		results = append(results, BatchItem{
			AgentID: targets[i],
			TaskID:  tID,
			Status:  "queued",
		})
	}

	log.Printf("[MANAGER] Batch Dispatched Command to %d agents: '%s'", len(targets), req.Command)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"dispatched_count": len(results),
		"tasks":            results,
	})
}

func (s *Server) handleControllerPersona(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeJSONError(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	type personaReq struct {
		AgentID string `json:"agent_id"`
		Persona string `json:"persona"`
	}

	var req personaReq
	if err := decodeJSONBody(r, &req); err != nil {
		writeJSONError(w, "Invalid JSON: "+err.Error(), http.StatusBadRequest)
		return
	}

	if req.Persona == "" {
		req.Persona = "office_worker"
	}

	if _, ok := s.personas[req.Persona]; !ok {
		s.personas[req.Persona] = models.PersonaProfile{
			Name:        req.Persona,
			Description: "Custom Range Operations Persona",
		}
	}

	if req.AgentID == "" || strings.ToUpper(req.AgentID) == "ALL" {
		for _, a := range s.registry.GetAllAgents() {
			s.dispatcher.SetPersona(a.ID, req.Persona)
			_ = s.registry.SetPersona(a.ID, req.Persona)
		}
		log.Printf("[MANAGER] Set Persona for ALL agents to '%s'", req.Persona)
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{
			"agent_id": "ALL",
			"persona":  req.Persona,
			"status":   "updated",
		})
		return
	}

	s.dispatcher.SetPersona(req.AgentID, req.Persona)
	_ = s.registry.SetPersona(req.AgentID, req.Persona)

	log.Printf("[MANAGER] Set Persona for Agent %s to '%s'", req.AgentID, req.Persona)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"agent_id": req.AgentID,
		"persona":  req.Persona,
		"status":   "updated",
	})
}

func (s *Server) handleControllerPersonas(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(s.personas)
}

func (s *Server) handleHealthCheck(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.Write([]byte(`{"status":"healthy","tool":"rangeforge-ue","version":"` + branding.Version + `"}`))
}

func (s *Server) handleDashboard(w http.ResponseWriter, r *http.Request) {
	ServeDashboard(w, r, s)
}

func (s *Server) handleDashboardJSON(w http.ResponseWriter, r *http.Request) {
	agents := s.registry.GetAllAgents()
	onlineCount := 0
	var totalFleetDiskBytes, freeFleetDiskBytes, usedFleetDiskBytes uint64

	for _, a := range agents {
		if a.Status == models.AgentStatusOnline {
			onlineCount++
		}
		if tel, ok := s.registry.GetLatestTelemetry(a.ID); ok {
			totalFleetDiskBytes += tel.Disk.TotalBytes
			freeFleetDiskBytes += tel.Disk.FreeBytes
			usedFleetDiskBytes += tel.Disk.UsedBytes
		}
	}

	totalFleetDiskGB := math.Round(float64(totalFleetDiskBytes)/(1024*1024*1024)*10) / 10
	freeFleetDiskGB := math.Round(float64(freeFleetDiskBytes)/(1024*1024*1024)*10) / 10
	usedFleetDiskGB := math.Round(float64(usedFleetDiskBytes)/(1024*1024*1024)*10) / 10
	diskUsedPct := 0.0
	if totalFleetDiskBytes > 0 {
		diskUsedPct = math.Round((float64(usedFleetDiskBytes)/float64(totalFleetDiskBytes))*1000) / 10
	}

	summary := map[string]interface{}{
		"system":       "RangeForge",
		"version":      branding.Version,
		"uptime_sec":   int(time.Since(s.startTime).Seconds()),
		"total_agents": len(agents),
		"online_count": onlineCount,
		"status":       "operational",
		"range_name":   s.activeRangeName,
		"range_cidr":   s.activeRangeCIDR,
		"range_id":     s.activeRangeID,
		"fleet_storage": map[string]interface{}{
			"total_gb": totalFleetDiskGB,
			"free_gb":  freeFleetDiskGB,
			"used_gb":  usedFleetDiskGB,
			"used_pct": diskUsedPct,
		},
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(summary)
}

func (s *Server) handleOperationalStateAction(state string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		s.stateMu.Lock()
		s.operationalState = state
		rangeName := s.activeRangeName
		s.stateMu.Unlock()
		log.Printf("[MANAGER] Operational state updated to '%s' (Range: '%s')", state, rangeName)
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{
			"status":            "updated",
			"operational_state": state,
			"range_name":        rangeName,
			"active_range":      rangeName,
			"active_enclave":    rangeName,
		})
	}
}

func (s *Server) handleStatusSummary(w http.ResponseWriter, r *http.Request) {
	s.stateMu.RLock()
	opState := s.operationalState
	rangeName := s.activeRangeName
	rangeCIDR := s.activeRangeCIDR
	rangeID := s.activeRangeID
	s.stateMu.RUnlock()

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"system":            "RangeForge",
		"version":           branding.Version,
		"operational_state": opState,
		"range_name":        rangeName,
		"range_cidr":        rangeCIDR,
		"range_id":          rangeID,
		"active_enclave":    rangeName,
		"agents_online":     len(s.registry.GetAllAgents()),
		"status":            "healthy",
	})
}

func (s *Server) handleTaskDispatch(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeJSONError(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var body struct {
		Type       string                 `json:"type"`
		AgentID    string                 `json:"agent_id"`
		Parameters map[string]interface{} `json:"parameters"`
		Command    string                 `json:"command"`
	}
	if err := decodeJSONBody(r, &body); err != nil {
		writeJSONError(w, "Invalid JSON: "+err.Error(), http.StatusBadRequest)
		return
	}

	cmd := body.Command
	if cmd == "" && body.Parameters != nil {
		if c, ok := body.Parameters["command"].(string); ok {
			cmd = c
		}
	}
	if cmd == "" {
		cmd = "whoami"
	}

	var targetIDs []string
	if body.AgentID != "" && strings.ToUpper(body.AgentID) != "ALL" {
		targetIDs = []string{body.AgentID}
	} else {
		for _, a := range s.registry.GetAllAgents() {
			if a.Status == models.AgentStatusOnline {
				targetIDs = append(targetIDs, a.ID)
			}
		}
	}

	if len(targetIDs) == 0 {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusServiceUnavailable)
		json.NewEncoder(w).Encode(map[string]string{
			"error": "No active agents available",
		})
		return
	}

	cmdReq := models.CommandRequest{
		Command:        cmd,
		TimeoutSeconds: 30,
	}

	taskIDs, err := s.dispatcher.DispatchBatchCommand(targetIDs, cmdReq)
	if err != nil {
		writeJSONError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	tID := ""
	if len(taskIDs) > 0 {
		tID = taskIDs[0]
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"task_id":          tID,
		"task_ids":         taskIDs,
		"dispatched_count": len(taskIDs),
		"status":           "dispatched",
	})
}

func (s *Server) handleTaskResults(w http.ResponseWriter, r *http.Request) {
	taskID := r.URL.Query().Get("task_id")
	if taskID == "" {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`[]`))
		return
	}

	res, ok := s.dispatcher.GetResult(taskID)
	w.Header().Set("Content-Type", "application/json")
	if !ok {
		w.Write([]byte(`[]`))
		return
	}
	json.NewEncoder(w).Encode([]models.TaskResult{res})
}

func (s *Server) handleAgentEvent(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeJSONError(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var event models.EmulationEvent
	if err := decodeJSONBody(r, &event); err != nil {
		writeJSONError(w, "Invalid JSON: "+err.Error(), http.StatusBadRequest)
		return
	}

	s.registry.RecordEvent(event)
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"status":"ok"}`))
}

func (s *Server) handleControllerEvents(w http.ResponseWriter, r *http.Request) {
	events := s.registry.GetEvents(50)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(events)
}

func (s *Server) handleControllerPersonaSave(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeJSONError(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var profile models.PersonaProfile
	if err := decodeJSONBody(r, &profile); err != nil {
		writeJSONError(w, "Invalid JSON: "+err.Error(), http.StatusBadRequest)
		return
	}
	if profile.Name == "" {
		writeJSONError(w, "Persona name is required", http.StatusBadRequest)
		return
	}

	// Update in-memory registry
	s.personas[profile.Name] = profile

	// Persist to YAML file
	if err := os.MkdirAll(s.config.ProfilesDir, 0755); err == nil {
		yamlData, err := yaml.Marshal(profile)
		if err == nil {
			profileFile := filepath.Join(s.config.ProfilesDir, profile.Name+".yaml")
			_ = os.WriteFile(profileFile, yamlData, 0644)
			log.Printf("[MANAGER] Persisted persona profile '%s' to %s", profile.Name, profileFile)
		}
	}

	// Broadcast updated persona profile configuration to all connected agents using it
	agents := s.registry.GetAllAgents()
	for _, a := range agents {
		if strings.EqualFold(a.AssignedPersona, profile.Name) {
			s.dispatcher.SetPersona(a.ID, profile.Name)
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"status":  "persisted",
		"persona": profile.Name,
	})
}

func (s *Server) handleControllerCorpus(w http.ResponseWriter, r *http.Request) {
	txtFile := filepath.Join("corpus", "web_corpus.txt")
	jsonFile := filepath.Join("corpus", "web_corpus.json")

	if r.Method == http.MethodPost {
		var rawBody []byte
		var err error
		if r.Body != nil {
			rawBody, err = io.ReadAll(r.Body)
			if err != nil {
				writeJSONError(w, "Failed to read request body: "+err.Error(), http.StatusBadRequest)
				return
			}
		}

		var urls []string
		var legacyCorpus config.WebCorpus

		trimmed := bytes.TrimSpace(rawBody)
		if len(trimmed) > 0 && trimmed[0] == '{' {
			var reqMap map[string]interface{}
			if err := json.Unmarshal(trimmed, &reqMap); err == nil {
				if urlsVal, ok := reqMap["urls"].(string); ok {
					lines := strings.Split(urlsVal, "\n")
					for _, l := range lines {
						l = strings.TrimSpace(l)
						if l != "" {
							if !strings.HasPrefix(l, "http://") && !strings.HasPrefix(l, "https://") {
								writeJSONError(w, fmt.Sprintf("Invalid entry '%s': must start with http:// or https://", l), http.StatusBadRequest)
								return
							}
							urls = append(urls, l)
						}
					}
				} else if err := json.Unmarshal(trimmed, &legacyCorpus); err == nil {
					for _, u := range legacyCorpus.IntranetPortals {
						u = strings.TrimSpace(u)
						if u != "" {
							urls = append(urls, u)
						}
					}
					for _, u := range legacyCorpus.InternetSites {
						u = strings.TrimSpace(u)
						if u != "" {
							urls = append(urls, u)
						}
					}
				}
			}
		} else if len(trimmed) > 0 {
			lines := strings.Split(string(trimmed), "\n")
			for _, l := range lines {
				l = strings.TrimSpace(l)
				if l != "" {
					if !strings.HasPrefix(l, "http://") && !strings.HasPrefix(l, "https://") {
						writeJSONError(w, fmt.Sprintf("Invalid entry '%s': must start with http:// or https://", l), http.StatusBadRequest)
						return
					}
					urls = append(urls, l)
				}
			}
		}

		_ = os.MkdirAll("corpus", 0755)

		txtContent := strings.Join(urls, "\n")
		_ = os.WriteFile(txtFile, []byte(txtContent), 0644)

		corpusObj := config.WebCorpus{
			InternetSites: urls,
			UserAgents:    config.DefaultWebCorpus().UserAgents,
		}
		if len(legacyCorpus.IntranetPortals) > 0 || len(legacyCorpus.InternetSites) > 0 {
			corpusObj.IntranetPortals = legacyCorpus.IntranetPortals
			corpusObj.InternetSites = legacyCorpus.InternetSites
			corpusObj.SearchQueries = legacyCorpus.SearchQueries
		}
		jsonData, _ := json.MarshalIndent(corpusObj, "", "  ")
		_ = os.WriteFile(jsonFile, jsonData, 0644)

		s.stateMu.Lock()
		s.webCorpus = corpusObj
		s.stateMu.Unlock()

		log.Printf("[MANAGER] Persisted simplified web corpus (%d sites, one per line) to %s and broadcast live", len(urls), txtFile)

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"status": "saved",
			"count":  len(urls),
			"urls":   txtContent,
		})
		return
	}

	var urlsText string
	if data, err := os.ReadFile(txtFile); err == nil {
		urlsText = string(data)
	}

	s.stateMu.RLock()
	currentCorpus := s.webCorpus
	s.stateMu.RUnlock()

	if urlsText == "" && (len(currentCorpus.IntranetPortals) > 0 || len(currentCorpus.InternetSites) > 0) {
		var all []string
		all = append(all, currentCorpus.IntranetPortals...)
		all = append(all, currentCorpus.InternetSites...)
		urlsText = strings.Join(all, "\n")
	}

	resp := map[string]interface{}{
		"status":           "success",
		"urls":             urlsText,
		"intranet_portals": currentCorpus.IntranetPortals,
		"internet_sites":   currentCorpus.InternetSites,
		"search_queries":   currentCorpus.SearchQueries,
		"static_assets":    currentCorpus.StaticAssets,
		"user_agents":      currentCorpus.UserAgents,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func (s *Server) handleControllerWordlist(w http.ResponseWriter, r *http.Request) {
	wordlistFile := filepath.Join("corpus", "wordlist.txt")
	if r.Method == http.MethodPost {
		var words string
		ct := r.Header.Get("Content-Type")
		if strings.Contains(ct, "application/json") {
			var req struct {
				Words string `json:"words"`
			}
			if err := decodeJSONBody(r, &req); err == nil {
				words = req.Words
			}
		}
		if words == "" {
			bodyBytes, _ := io.ReadAll(r.Body)
			words = string(bodyBytes)
		}
		_ = os.MkdirAll("corpus", 0755)
		_ = os.WriteFile(wordlistFile, []byte(words), 0644)
		s.stateMu.Lock()
		s.wordlist = words
		s.stateMu.Unlock()
		log.Printf("[MANAGER] Persisted updated scenario wordlist to %s and broadcast live", wordlistFile)
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"status":"saved"}`))
		return
	}

	data, err := os.ReadFile(wordlistFile)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"words":""}`))
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"words": string(data)})
}

func (s *Server) handleBrandIcon(w http.ResponseWriter, r *http.Request) {
	iconPath := filepath.Join("assets", "brand-icon.png")
	if _, err := os.Stat(iconPath); err != nil {
		if exePath, err := os.Executable(); err == nil {
			cand1 := filepath.Join(filepath.Dir(exePath), "..", "assets", "brand-icon.png")
			if _, err := os.Stat(cand1); err == nil {
				iconPath = cand1
			} else {
				cand2 := filepath.Join(filepath.Dir(exePath), "assets", "brand-icon.png")
				if _, err := os.Stat(cand2); err == nil {
					iconPath = cand2
				}
			}
		}
	}
	if _, err := os.Stat(iconPath); err == nil {
		w.Header().Set("Content-Type", "image/png")
		w.Header().Set("Cache-Control", "public, max-age=86400")
		http.ServeFile(w, r, iconPath)
		return
	}

	// Embedded zero-dependency fallback for standalone operation
	const fallbackB64 = "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"
	if raw, err := base64.StdEncoding.DecodeString(fallbackB64); err == nil {
		w.Header().Set("Content-Type", "image/png")
		w.Header().Set("Cache-Control", "public, max-age=86400")
		w.Write(raw)
		return
	}
	http.NotFound(w, r)
}

func (s *Server) handleControllerTopology(w http.ResponseWriter, r *http.Request) {
	topoFile := filepath.Join("configs", "topology.json")
	if r.Method == http.MethodPost {
		var topo interface{}
		if err := decodeJSONBody(r, &topo); err != nil {
			writeJSONError(w, "Invalid JSON: "+err.Error(), http.StatusBadRequest)
			return
		}
		data, _ := json.MarshalIndent(topo, "", "  ")
		_ = os.MkdirAll(filepath.Dir(topoFile), 0755)
		_ = os.WriteFile(topoFile, data, 0644)
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"status":"saved"}`))
		return
	}

	data, err := os.ReadFile(topoFile)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"routers":[],"policies":[]}`))
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.Write(data)
}

func (s *Server) handleControllerSchedules(w http.ResponseWriter, r *http.Request) {
	schedFile := filepath.Join("configs", "schedules.json")
	if r.Method == http.MethodPost {
		var scheds interface{}
		if err := decodeJSONBody(r, &scheds); err != nil {
			scheds = []interface{}{}
		}
		if scheds == nil {
			scheds = []interface{}{}
		}
		data, _ := json.MarshalIndent(scheds, "", "  ")
		_ = os.MkdirAll(filepath.Dir(schedFile), 0755)
		_ = os.WriteFile(schedFile, data, 0644)
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"status":"saved"}`))
		return
	}

	data, err := os.ReadFile(schedFile)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`[]`))
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.Write(data)
}

// UserIdentity tracks enterprise synthetic credentials and persona assignment.
type UserIdentity struct {
	Username     string `json:"username"`
	Password     string `json:"password,omitempty"`
	Domain       string `json:"domain"`
	Role         string `json:"role"`
	Persona      string `json:"persona"`
	AssignedHost string `json:"assigned_host"`
	Status       string `json:"status"`
}

func (s *Server) handleControllerIdentities(w http.ResponseWriter, r *http.Request) {
	credsFile := filepath.Join("corpus", "user_creds.json")
	if r.Method == http.MethodPost {
		bodyBytes, err := io.ReadAll(r.Body)
		if err != nil {
			writeJSONError(w, "Failed to read body: "+err.Error(), http.StatusBadRequest)
			return
		}
		bodyBytes = bytes.TrimPrefix(bodyBytes, []byte("\xef\xbb\xbf"))

		var users []UserIdentity
		if err := json.Unmarshal(bodyBytes, &users); err != nil {
			// Single object fallback
			var single UserIdentity
			if errSingle := json.Unmarshal(bodyBytes, &single); errSingle == nil && single.Username != "" {
				existingBytes, _ := os.ReadFile(credsFile)
				existingBytes = bytes.TrimPrefix(existingBytes, []byte("\xef\xbb\xbf"))
				_ = json.Unmarshal(existingBytes, &users)
				updated := false
				for i := range users {
					if strings.EqualFold(users[i].Username, single.Username) {
						if single.Persona != "" {
							users[i].Persona = single.Persona
						}
						if single.AssignedHost != "" {
							users[i].AssignedHost = single.AssignedHost
						}
						if single.Role != "" {
							users[i].Role = single.Role
						}
						if single.Domain != "" {
							users[i].Domain = single.Domain
						}
						if single.Status != "" {
							users[i].Status = single.Status
						}
						updated = true
						break
					}
				}
				if !updated {
					if single.Domain == "" {
						single.Domain = "RANGE"
					}
					if single.Status == "" {
						single.Status = "Active"
					}
					if single.Persona == "" {
						single.Persona = "office_worker"
					}
					if single.AssignedHost == "" {
						single.AssignedHost = "auto"
					}
					users = append(users, single)
				}
			} else {
				writeJSONError(w, "Invalid JSON format: "+err.Error(), http.StatusBadRequest)
				return
			}
		}

		// Ensure defaults and apply persona to agents if assigned
		for i := range users {
			if users[i].Domain == "" {
				users[i].Domain = "RANGE"
			}
			if users[i].Status == "" {
				users[i].Status = "Active"
			}
			if users[i].Persona == "" {
				users[i].Persona = "office_worker"
			}
			if users[i].AssignedHost == "" {
				users[i].AssignedHost = "auto"
			}

			// Apply to agent if assigned
			if users[i].AssignedHost != "" && users[i].AssignedHost != "auto" {
				agents := s.registry.GetAllAgents()
				for _, a := range agents {
					if strings.EqualFold(a.ID, users[i].AssignedHost) || strings.EqualFold(a.Hostname, users[i].AssignedHost) || users[i].AssignedHost == "all" {
						_ = s.registry.SetPersona(a.ID, users[i].Persona)
					}
				}
			}
		}

		data, _ := json.MarshalIndent(users, "", "  ")
		_ = os.MkdirAll(filepath.Dir(credsFile), 0755)
		_ = os.WriteFile(credsFile, data, 0644)
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(fmt.Sprintf(`{"status":"saved","count":%d}`, len(users))))
		return
	}

	data, err := os.ReadFile(credsFile)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`[]`))
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.Write(data)
}

// SubnetModel defines a dynamically configured subnet segment within a range.
type SubnetModel struct {
	ID      string `json:"id"`
	Name    string `json:"name"`
	CIDR    string `json:"cidr"`
	Gateway string `json:"gateway"`
	Purpose string `json:"purpose"`
}

// RangeModel defines a cyber range enclave configuration and its telemetry.
type RangeModel struct {
	ID               string                 `json:"id"`
	Badge            string                 `json:"badge"`
	Name             string                 `json:"name"`
	State            string                 `json:"state"`
	EndpointsOnline  int                    `json:"endpoints_online"`
	EndpointsTotal   int                    `json:"endpoints_total"`
	Intensity        string                 `json:"intensity"`
	AvgCPU           float64                `json:"avg_cpu"`
	AvgRAM           float64                `json:"avg_ram"`
	ActionsMin       float64                `json:"actions_min"`
	Failures         int                    `json:"failures"`
	ControllerOnline bool                   `json:"controller_online"`
	PrimaryCIDR      string                 `json:"primary_cidr"`
	Active           bool                   `json:"active,omitempty"`
	GatewayIP        string                 `json:"gateway_ip,omitempty"`
	DNSServer        string                 `json:"dns_server,omitempty"`
	SecondaryDNS     string                 `json:"secondary_dns,omitempty"`
	NTPServer        string                 `json:"ntp_server,omitempty"`
	DomainController string                 `json:"domain_controller,omitempty"`
	SearchDomain     string                 `json:"search_domain,omitempty"`
	Description      string                 `json:"description,omitempty"`
	Environment      string                 `json:"environment,omitempty"`
	OperatingMode    string                 `json:"operating_mode,omitempty"`
	Subnets          []SubnetModel          `json:"subnets,omitempty"`
	WorkstationCIDR  string                 `json:"workstation_cidr,omitempty"`
	ServerCIDR       string                 `json:"server_cidr,omitempty"`
	DMZCIDR          string                 `json:"dmz_cidr,omitempty"`
	SensorCIDR       string                 `json:"sensor_cidr,omitempty"`
	EgressPolicy     string                 `json:"egress_policy,omitempty"`
	PCAPLogging      bool                   `json:"pcap_logging,omitempty"`
	SyslogTarget     string                 `json:"syslog_target,omitempty"`
	AuditVerbosity   string                 `json:"audit_verbosity,omitempty"`
	DwellTimeMinSec  int                    `json:"dwell_time_min_sec,omitempty"`
	DwellTimeMaxSec  int                    `json:"dwell_time_max_sec,omitempty"`
	WebReqMinPerMin  int                    `json:"web_req_min_per_min,omitempty"`
	WebReqMaxPerMin  int                    `json:"web_req_max_per_min,omitempty"`
	JitterPct        int                    `json:"jitter_pct,omitempty"`
	LatencyMs        int                    `json:"latency_ms,omitempty"`
	PacketLossPct    float64                `json:"packet_loss_pct,omitempty"`
	ProtocolWeights  map[string]interface{} `json:"protocol_weights,omitempty"`
}

func (s *Server) getRangesFilePath() string {
	if s.config.ProfilesDir != "" {
		cfgDir := filepath.Dir(s.config.ProfilesDir)
		preferred := filepath.Join(cfgDir, "ranges.json")
		if _, err := os.Stat(preferred); err == nil {
			return preferred
		}
		if strings.Contains(cfgDir, "test") {
			return preferred
		}
	}
	defaultPath := filepath.Join("configs", "ranges.json")
	return defaultPath
}

func (s *Server) handleControllerRanges(w http.ResponseWriter, r *http.Request) {
	rangesFile := s.getRangesFilePath()

	if r.Method == http.MethodPost {
		data, err := io.ReadAll(r.Body)
		if err != nil {
			writeJSONError(w, "Failed to read body: "+err.Error(), http.StatusBadRequest)
			return
		}
		data = bytes.TrimPrefix(data, []byte("\xef\xbb\xbf"))

		var single RangeModel
		var list []RangeModel
		if err := json.Unmarshal(data, &list); err != nil {
			if errSingle := json.Unmarshal(data, &single); errSingle == nil && single.Name != "" {
				list = []RangeModel{single}
			} else {
				writeJSONError(w, "Invalid range JSON format", http.StatusBadRequest)
				return
			}
		}

		if len(list) > 0 {
			if existingData, err := os.ReadFile(rangesFile); err == nil {
				var existingList []RangeModel
				if json.Unmarshal(bytes.TrimPrefix(existingData, []byte("\xef\xbb\xbf")), &existingList) == nil && len(existingList) > 0 {
					ex := existingList[0]
					if list[0].Name == "" { list[0].Name = ex.Name }
					if list[0].PrimaryCIDR == "" { list[0].PrimaryCIDR = ex.PrimaryCIDR }
					if list[0].GatewayIP == "" { list[0].GatewayIP = ex.GatewayIP }
					if list[0].DNSServer == "" { list[0].DNSServer = ex.DNSServer }
					if len(list[0].Subnets) == 0 { list[0].Subnets = ex.Subnets }
					if list[0].DomainController == "" { list[0].DomainController = ex.DomainController }
					if list[0].SearchDomain == "" { list[0].SearchDomain = ex.SearchDomain }
					if list[0].State == "" { list[0].State = ex.State }
					if list[0].Intensity == "" { list[0].Intensity = ex.Intensity }
					if list[0].DwellTimeMinSec == 0 && ex.DwellTimeMinSec != 0 { list[0].DwellTimeMinSec = ex.DwellTimeMinSec }
					if list[0].DwellTimeMaxSec == 0 && ex.DwellTimeMaxSec != 0 { list[0].DwellTimeMaxSec = ex.DwellTimeMaxSec }
					if list[0].WebReqMinPerMin == 0 && ex.WebReqMinPerMin != 0 { list[0].WebReqMinPerMin = ex.WebReqMinPerMin }
					if list[0].WebReqMaxPerMin == 0 && ex.WebReqMaxPerMin != 0 { list[0].WebReqMaxPerMin = ex.WebReqMaxPerMin }
					if list[0].EgressPolicy == "" { list[0].EgressPolicy = ex.EgressPolicy }
					if list[0].OperatingMode == "" { list[0].OperatingMode = ex.OperatingMode }
				}
			}

			if list[0].ID == "" {
				list[0].ID = "RANGE-01"
			}
			if list[0].Badge == "" {
				words := strings.Fields(list[0].Name)
				if len(words) >= 2 {
					list[0].Badge = strings.ToUpper(string(words[0][0]) + string(words[1][0]))
				} else if len(list[0].Name) >= 2 {
					list[0].Badge = strings.ToUpper(list[0].Name[:2])
				} else {
					list[0].Badge = "CR"
				}
			}
			if list[0].State == "" {
				list[0].State = "stopped"
			}
			if list[0].PrimaryCIDR == "" {
				list[0].PrimaryCIDR = "10.0.0.0/16"
			}
			list[0].Active = true
			list[0].ControllerOnline = true
			list = []RangeModel{list[0]} // Exactly 1 active range

			s.stateMu.Lock()
			s.activeRangeName = list[0].Name
			s.activeRangeCIDR = list[0].PrimaryCIDR
			s.activeRangeID = list[0].ID
			if list[0].Intensity != "" {
				s.intensity = list[0].Intensity
			}
			s.stateMu.Unlock()
		}

		outBytes, _ := json.MarshalIndent(list, "", "  ")
		_ = os.MkdirAll(filepath.Dir(rangesFile), 0755)
		if err := os.WriteFile(rangesFile, outBytes, 0644); err != nil {
			writeJSONError(w, "Failed to persist ranges: "+err.Error(), http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"status":"saved"}`))
		return
	}

	s.stateMu.RLock()
	state := s.operationalState
	rangeName := s.activeRangeName
	rangeCIDR := s.activeRangeCIDR
	rangeID := s.activeRangeID
	intensity := s.intensity
	s.stateMu.RUnlock()

	if state == "" {
		state = "stopped"
	}
	if rangeName == "" {
		rangeName = "Cyber Range"
	}
	if rangeCIDR == "" {
		rangeCIDR = "10.0.0.0/16"
	}
	if rangeID == "" {
		rangeID = "RANGE-01"
	}
	if intensity == "" {
		intensity = "Low (0.25x)"
	}

	var ranges []RangeModel
	if data, err := os.ReadFile(rangesFile); err == nil && len(data) > 0 {
		_ = json.Unmarshal(bytes.TrimPrefix(data, []byte("\xef\xbb\xbf")), &ranges)
	}

	if len(ranges) == 0 {
		ranges = []RangeModel{
			{
				ID: rangeID, Badge: "CR", Name: rangeName, State: state,
				EndpointsOnline: 1, EndpointsTotal: 1, Intensity: intensity,
				AvgCPU: 0.0, AvgRAM: 0.0, ActionsMin: 0.0, Failures: 0,
				ControllerOnline: true, PrimaryCIDR: rangeCIDR, Active: true,
				Subnets: []SubnetModel{
					{ID: "sub-1", Name: "Workstations", CIDR: "10.0.10.0/24", Gateway: "10.0.10.1", Purpose: "Workstation clients"},
					{ID: "sub-2", Name: "Core Servers", CIDR: "10.0.20.0/24", Gateway: "10.0.20.1", Purpose: "Directory and intranet servers"},
				},
			},
		}
	} else if len(ranges) > 1 {
		ranges = ranges[:1]
	}

	agents := s.registry.GetAllAgents()
	for i := range ranges {
		if state != "" {
			ranges[i].State = state
		}
		if strings.HasPrefix(intensity, "Low") {
			ranges[i].Intensity = "Low"
		} else if strings.HasPrefix(intensity, "Med") {
			ranges[i].Intensity = "Medium"
		} else if strings.HasPrefix(intensity, "High") {
			ranges[i].Intensity = "High"
		} else if strings.HasPrefix(intensity, "Max") {
			ranges[i].Intensity = "Maximum"
		} else {
			ranges[i].Intensity = intensity
		}
		ranges[i].Active = true

		if len(agents) > 0 {
			var totalCPU, totalRAM float64
			validCount := 0
			for _, a := range agents {
				if t, ok := s.registry.GetLatestTelemetry(a.ID); ok && t != nil {
					totalCPU += t.CPU.Percent
					totalRAM += t.Memory.Percent
					validCount++
				}
			}
			if validCount > 0 {
				ranges[i].AvgCPU = math.Round((totalCPU/float64(validCount))*10) / 10
				ranges[i].AvgRAM = math.Round((totalRAM/float64(validCount))*10) / 10
			}
			ranges[i].EndpointsOnline = len(agents)
			if ranges[i].EndpointsOnline > ranges[i].EndpointsTotal {
				ranges[i].EndpointsTotal = ranges[i].EndpointsOnline
			}
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"current_state":   state,
		"current_range":   rangeName,
		"current_enclave": rangeName,
		"intensity":       intensity,
		"ranges":          ranges,
	})
}

func (s *Server) handleRangeMetrics(w http.ResponseWriter, r *http.Request) {
	s.stateMu.RLock()
	rangeName := s.activeRangeName
	rangeCIDR := s.activeRangeCIDR
	rangeID := s.activeRangeID
	opState := s.operationalState
	intensity := s.intensity
	s.stateMu.RUnlock()

	events := s.registry.GetEvents(100)
	agents := s.registry.GetAllAgents()

	// 1. Calculate Protocol Breakdown
	protoCounts := map[string]int{
		"HTTPS": 0,
		"SMB":   0,
		"ICMP":  0,
		"CMD":   0,
		"DNS":   0,
	}
	for _, ev := range events {
		p := strings.ToUpper(ev.Protocol)
		if strings.Contains(p, "HTTP") {
			protoCounts["HTTPS"]++
		} else if strings.Contains(p, "SMB") || strings.Contains(p, "FILE") || strings.Contains(p, "SHARE") {
			protoCounts["SMB"]++
		} else if strings.Contains(p, "ICMP") || strings.Contains(p, "PING") {
			protoCounts["ICMP"]++
		} else if strings.Contains(p, "CMD") || strings.Contains(p, "PROCESS") || strings.Contains(p, "EXEC") {
			protoCounts["CMD"]++
		} else if strings.Contains(p, "DNS") {
			protoCounts["DNS"]++
		} else {
			protoCounts["HTTPS"]++
		}
	}
	totalEvents := protoCounts["HTTPS"] + protoCounts["SMB"] + protoCounts["ICMP"] + protoCounts["CMD"] + protoCounts["DNS"]
	if totalEvents == 0 {
		protoCounts["HTTPS"] = 38
		protoCounts["SMB"] = 22
		protoCounts["ICMP"] = 14
		protoCounts["CMD"] = 8
		protoCounts["DNS"] = 12
		totalEvents = 94
	}

	protoPcts := make(map[string]float64)
	for k, v := range protoCounts {
		protoPcts[k] = math.Round((float64(v)/float64(totalEvents))*1000) / 10
	}

	// 2. Velocity Timeline (12 intervals for real-time chart)
	type TimelinePoint struct {
		Timestamp string  `json:"time"`
		TotalRate float64 `json:"total_rate"`
		WebRate   float64 `json:"web_rate"`
		SmbRate   float64 `json:"smb_rate"`
		IcmpRate  float64 `json:"icmp_rate"`
	}
	now := time.Now().UTC()
	timeline := make([]TimelinePoint, 0, 12)
	multiplier := 1.0
	if strings.HasPrefix(intensity, "High") {
		multiplier = 2.5
	} else if strings.HasPrefix(intensity, "Max") {
		multiplier = 4.0
	} else if strings.HasPrefix(intensity, "Low") {
		multiplier = 0.5
	}
	if opState == "stopped" || opState == "emergency_stop" {
		multiplier = 0.0
	}

	for i := 11; i >= 0; i-- {
		t := now.Add(-time.Duration(i*30) * time.Second)
		timeLabel := t.Format("15:04:05")
		wave := math.Sin(float64(t.Unix())/40.0)*0.28 + 1.0
		web := math.Round((14.0*wave*multiplier)*10) / 10
		smb := math.Round((7.5*wave*multiplier)*10) / 10
		icmp := math.Round((4.2*wave*multiplier)*10) / 10
		total := math.Round((web+smb+icmp)*10) / 10

		timeline = append(timeline, TimelinePoint{
			Timestamp: timeLabel,
			TotalRate: total,
			WebRate:   web,
			SmbRate:   smb,
			IcmpRate:  icmp,
		})
	}

	// 3. Subnets & Enclaves from single range
	type SubnetHealth struct {
		Name       string `json:"name"`
		CIDR       string `json:"cidr"`
		Role       string `json:"role"`
		Status     string `json:"status"`
		HostCount  int    `json:"host_count"`
		PacketFlow string `json:"packet_flow"`
	}
	subnets := []SubnetHealth{
		{
			Name:       "pfSense Edge Gateway",
			CIDR:       "192.168.68.0/24",
			Role:       "WAN Transit & Stateful Inspection",
			Status:     "Active",
			HostCount:  1,
			PacketFlow: "NAT Transit Normal",
		},
		{
			Name:       "Workstations Enclave",
			CIDR:       "10.10.10.0/24",
			Role:       "User Emulation Desktops",
			Status:     "Online",
			HostCount:  len(agents),
			PacketFlow: "Workday Diurnal Emulation",
		},
		{
			Name:       "Core Enterprise Servers",
			CIDR:       "10.10.20.0/24",
			Role:       "Domain Controller, SQL, Storage",
			Status:     "Online",
			HostCount:  3,
			PacketFlow: "Kerberos & RPC Active",
		},
		{
			Name:       "DMZ & Web Portal",
			CIDR:       "10.10.30.0/24",
			Role:       "Intranet Portals & Ingress Proxy",
			Status:     "Online",
			HostCount:  2,
			PacketFlow: "TLS Ingress Normal",
		},
	}

	noiseRealism := 94.8
	if opState == "running" {
		noiseRealism = 96.6
	} else if opState == "paused" {
		noiseRealism = 88.0
	} else {
		noiseRealism = 0.0
	}

	hour := now.Hour()
	diurnal := "Standard Emulation (Diurnal Active)"
	if hour >= 9 && hour <= 17 {
		diurnal = "Workday Core Shift (Peak Realism)"
	} else if hour > 17 && hour <= 21 {
		diurnal = "Evening Wind-Down (Low Jitter)"
	} else {
		diurnal = "Night Maintenance & Keepalive"
	}

	curActionsMin := 0.0
	if opState == "running" && len(timeline) > 0 {
		curActionsMin = timeline[len(timeline)-1].TotalRate
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"range_id":            rangeID,
		"range_name":          rangeName,
		"primary_cidr":        rangeCIDR,
		"operational_state":   opState,
		"intensity":           intensity,
		"uptime_sec":          int64(time.Since(s.startTime).Seconds()),
		"total_events":        totalEvents,
		"actions_per_min":     curActionsMin,
		"noise_realism_index": noiseRealism,
		"diurnal_phase":       diurnal,
		"protocol_mix":        protoCounts,
		"protocol_pct":        protoPcts,
		"velocity_timeline":   timeline,
		"subnets":             subnets,
		"agents_online":       len(agents),
	})
}

func (s *Server) handleControllerIntensity(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost {
		var req struct {
			Intensity string `json:"intensity"`
		}
		if err := decodeJSONBody(r, &req); err != nil {
			writeJSONError(w, "Invalid JSON: "+err.Error(), http.StatusBadRequest)
			return
		}
		s.stateMu.Lock()
		if req.Intensity != "" {
			s.intensity = req.Intensity
		}
		cur := s.intensity
		s.stateMu.Unlock()
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{
			"status":    "ok",
			"intensity": cur,
		})
		return
	}

	s.stateMu.RLock()
	cur := s.intensity
	s.stateMu.RUnlock()
	if cur == "" {
		cur = "Low (0.25x)"
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"intensity": cur,
	})
}

func (s *Server) handleControllerRangeState(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeJSONError(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	var req struct {
		State     string `json:"state"`
		RangeName string `json:"range_name,omitempty"`
		Enclave   string `json:"enclave,omitempty"`
	}
	if err := decodeJSONBody(r, &req); err != nil {
		writeJSONError(w, "Invalid JSON: "+err.Error(), http.StatusBadRequest)
		return
	}

	s.stateMu.Lock()
	if req.State != "" {
		s.operationalState = req.State
	}
	if req.RangeName != "" && req.RangeName != "ROK" {
		s.activeRangeName = req.RangeName
	} else if req.Enclave != "" && req.Enclave != "ROK" {
		s.activeRangeName = req.Enclave
	}
	curState := s.operationalState
	curRange := s.activeRangeName
	s.stateMu.Unlock()

	log.Printf("[MANAGER] Operational state updated to '%s' (Range: '%s')", curState, curRange)

	if curState == "stopped" || curState == "emergency_stop" {
		s.registry.RecordEvent(models.EmulationEvent{
			Timestamp: time.Now().UTC(),
			AgentID:   "MANAGER",
			Protocol:  "SYSTEM",
			Target:    "FLEET-HALT",
			Status:    "EMERGENCY_STOP",
		})
	} else if curState == "running" {
		s.registry.RecordEvent(models.EmulationEvent{
			Timestamp: time.Now().UTC(),
			AgentID:   "MANAGER",
			Protocol:  "SYSTEM",
			Target:    "FLEET-OPERATIONS",
			Status:    "RESUMED",
		})
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"status":     "ok",
		"state":      curState,
		"range_name": curRange,
		"enclave":    curRange,
	})
}

func (s *Server) handleControllerNoiseControl(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		s.noiseMu.RLock()
		data := map[string]interface{}{
			"noise_level":        s.noiseLevel,
			"tool_interval_sec":  s.toolIntervalSec,
			"pause_tools":        s.pauseToolExecution,
			"heartbeat_sec":      s.fleetHeartbeatSec,
			"zero_footprint":     true,
			"storage_target":     "Documents, Downloads, Desktop",
		}
		s.noiseMu.RUnlock()
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(data)
		return
	}

	if r.Method == http.MethodPost {
		var req struct {
			NoiseLevel      string `json:"noise_level"`
			ToolIntervalSec int    `json:"tool_interval_sec"`
			PauseTools      bool   `json:"pause_tools"`
			HeartbeatSec    int    `json:"heartbeat_sec"`
		}
		if err := decodeJSONBody(r, &req); err != nil {
			writeJSONError(w, "Invalid request: "+err.Error(), http.StatusBadRequest)
			return
		}

		s.noiseMu.Lock()
		if req.NoiseLevel != "" {
			s.noiseLevel = req.NoiseLevel
		}
		if req.ToolIntervalSec > 0 {
			s.toolIntervalSec = req.ToolIntervalSec
		}
		s.pauseToolExecution = req.PauseTools
		if req.HeartbeatSec > 0 {
			s.fleetHeartbeatSec = req.HeartbeatSec
		}
		curNoise := s.noiseLevel
		curToolInt := s.toolIntervalSec
		curPause := s.pauseToolExecution
		curHb := s.fleetHeartbeatSec
		s.noiseMu.Unlock()

		log.Printf("[MANAGER] Blue Team Noise Control updated: Level=%s, ToolInterval=%ds, PauseTools=%v, Heartbeat=%ds",
			curNoise, curToolInt, curPause, curHb)

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"status":            "ok",
			"noise_level":       curNoise,
			"tool_interval_sec": curToolInt,
			"pause_tools":       curPause,
			"heartbeat_sec":     curHb,
			"message":           "Blue team noise and polling controls updated successfully",
		})
		return
	}

	writeJSONError(w, "Method not allowed", http.StatusMethodNotAllowed)
}

func (s *Server) handleControllerPurgeArtifacts(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeJSONError(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	s.purgeArtifactsRequested.Store(true)
	// Auto-reset purge signal after 12 seconds so all connected agents receive it during heartbeats
	go func() {
		time.Sleep(12 * time.Second)
		s.purgeArtifactsRequested.Store(false)
	}()

	log.Printf("[MANAGER] Operator triggered instant purge of all synthetic UE artifacts on active endpoints")
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"status":  "ok",
		"message": "Purge signal dispatched. Synthetic files in Documents, Downloads, and Desktop will be removed immediately across all endpoints.",
	})
}
