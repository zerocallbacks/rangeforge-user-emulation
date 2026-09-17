package agent

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"path/filepath"
	"reflect"
	"strings"
	"sync"
	"sync/atomic"
	"syscall"
	"time"

	"rangeforge-ue/pkg/agent/emulator"
	"rangeforge-ue/pkg/agent/telemetry"
	"rangeforge-ue/pkg/agent/tools"
	"rangeforge-ue/pkg/branding"
	"rangeforge-ue/pkg/config"
	"rangeforge-ue/pkg/models"
	"github.com/google/uuid"
)

// Agent manages host telemetry and user emulation routines.
type Agent struct {
	config        config.AgentConfig
	client        *ManagerClient
	collector     telemetry.Collector
	toolProfile   tools.ToolProfile
	webEmulator   *emulator.WebBrowserEmulator
	shareEmulator *emulator.FileShareEmulator
	pingEmulator  *emulator.PingEmulator
	hostEmulator  *emulator.HostActivityEmulator
	cmdRunner     *emulator.CommandRunner

	agentInfo       models.AgentInfo
	activePersona   models.PersonaProfile
	personas        map[string]models.PersonaProfile
	webCorpus       config.WebCorpus
	shareCorpus     config.ShareCorpus
	currentWordlist string
	eventChan       chan models.EmulationEvent

	running         atomic.Bool
	emulationActive atomic.Bool
	heartbeatSec    atomic.Int64
	stopChan        chan struct{}
	wg              sync.WaitGroup
	cmdCount        atomic.Uint64
}

// NewAgent constructs a new cross-platform host agent.
func NewAgent(cfg config.AgentConfig) (*Agent, error) {
	collector := telemetry.NewPlatformCollector()
	hostname, _ := os.Hostname()
	osName, platform, arch, primaryIP, ips, macs := collector.GetPlatformInfo()
	toolProfile := tools.DetectInstalledTools()

	agentID := fmt.Sprintf("%s-%s", hostname, uuid.New().String()[:8])

	info := models.AgentInfo{
		ID:              agentID,
		Hostname:        hostname,
		OS:              osName,
		Platform:        platform,
		Arch:            arch,
		PrimaryIP:       primaryIP,
		IPAddresses:     ips,
		MACAddresses:    macs,
		AssignedPersona: cfg.Persona,
		Status:          models.AgentStatusOnline,
		FirstSeen:       time.Now().UTC(),
		LastHeartbeat:   time.Now().UTC(),
		Version:         branding.Version,
		Tags:            cfg.Tags,
		AvailableTools:  toolProfile.AvailableTools,
		NetworkTools:    toolProfile.NetworkTools,
		HostTools:       toolProfile.HostTools,
	}

	webCorpus, err := config.LoadWebCorpus(cfg.WebCorpusFile)
	if err != nil {
		log.Printf("[AGENT] Note on web corpus: %v (using defaults)", err)
	}

	shareCorpus := config.DefaultShareCorpus()

	personas := config.BuiltinPersonas()
	activePersona, ok := personas[cfg.Persona]
	if !ok {
		activePersona = personas["office_worker"]
	}

	ag := &Agent{
		config:        cfg,
		client:        NewManagerClient(cfg.ManagerURL),
		collector:     collector,
		toolProfile:   toolProfile,
		cmdRunner:     emulator.NewCommandRunner(),
		agentInfo:     info,
		activePersona: activePersona,
		personas:      personas,
		webCorpus:     webCorpus,
		shareCorpus:   shareCorpus,
		stopChan:      make(chan struct{}),
		eventChan:     make(chan models.EmulationEvent, 256),
	}
	hbSec := cfg.HeartbeatSec
	if hbSec <= 0 {
		hbSec = 5
	}
	ag.heartbeatSec.Store(int64(hbSec))
	CleanLegacyArtifacts()
	return ag, nil
}

// Start launches the agent, registers with the Manager, and starts emulation routines.
func (a *Agent) Start(ctx context.Context) error {
	branding.PrintBanner("Host Agent")
	log.Printf("[AGENT] Initializing RangeForge Host Agent on %s (%s/%s)", a.agentInfo.Hostname, a.agentInfo.OS, a.agentInfo.Arch)
	log.Printf("[AGENT] Assigned Primary IP: %s (MACs: %v)", a.agentInfo.PrimaryIP, a.agentInfo.MACAddresses)
	log.Printf("[AGENT] Detected Functional Tools: %v", a.toolProfile.AvailableTools)
	log.Printf("[AGENT] Target Manager URL: %s", a.config.ManagerURL)
	log.Printf("[AGENT] Active Persona Profile: %s", a.activePersona.Name)

	// Attempt initial registration
	regResp, err := a.client.Register(a.agentInfo)
	if err != nil {
		log.Printf("[AGENT] Warning: Initial registration failed (%v). Will auto-retry in heartbeat loop.", err)
	} else {
		log.Printf("[AGENT] Successfully registered with Manager. Assigned Persona: %s", regResp.AssignedPersona)
		if regResp.AssignedPersona != "" && regResp.AssignedPersona != a.activePersona.Name {
			if p, ok := a.personas[regResp.AssignedPersona]; ok {
				a.activePersona = p
			}
		}
	}

	// Initialize emulation engines based on active persona and available tools
	a.initEmulators()

	a.running.Store(true)

	// Keep emulators on STANDBY until explicit Start signal is received from Manager
	log.Printf("[AGENT] Emulation engines initialized on STANDBY. Waiting for Start Operations signal...")

	// Main heartbeat loop
	a.wg.Add(1)
	go a.heartbeatLoop()

	// Dedicated event dispatcher worker
	a.wg.Add(1)
	go a.eventWorker()

	// Wait for OS interrupt or context cancellation
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	select {
	case <-ctx.Done():
		log.Printf("[AGENT] Context cancelled, shutting down...")
	case sig := <-sigChan:
		log.Printf("[AGENT] Received signal %v, shutting down...", sig)
	case <-a.stopChan:
		log.Printf("[AGENT] Stop requested, shutting down...")
	}

	a.Stop()
	return nil
}

func (a *Agent) initEmulators() {
	if a.config.EnableWeb {
		a.webEmulator = emulator.NewWebBrowserEmulator(a.activePersona.WebBrowsing, a.webCorpus)
		a.webEmulator.SetEventCallback(a.handleEvent)
	}
	if a.config.EnableShares {
		a.shareEmulator = emulator.NewFileShareEmulator(a.activePersona.FileShare, a.shareCorpus)
		a.shareEmulator.SetEventCallback(a.handleEvent)
		if a.currentWordlist != "" {
			a.shareEmulator.SetWordlist(strings.Split(a.currentWordlist, "\n"))
		}
	}
	// Autoload ping emulator if ping tool is functional on the host
	if a.toolProfile.CanPing {
		pingCfg := a.activePersona.Ping
		if len(pingCfg.Targets) == 0 {
			pingCfg.Targets = []string{"127.0.0.1", "1.1.1.1", "8.8.8.8"}
			pingCfg.Enabled = true
			pingCfg.IntervalSec = 20
		}
		a.pingEmulator = emulator.NewPingEmulator(pingCfg, a.handleEvent)
	}
	if a.config.EnableHostActivity {
		a.hostEmulator = emulator.NewHostActivityEmulator(a.activePersona.HostActivity)
		a.hostEmulator.SetEventCallback(a.handleEvent)
	}
}

func (a *Agent) eventWorker() {
	defer a.wg.Done()
	for {
		select {
		case <-a.stopChan:
			return
		case evt := <-a.eventChan:
			_ = a.client.SendEvent(evt)
		}
	}
}

func (a *Agent) handleEvent(protocol, target, status string, durationMs int64) {
	evt := models.EmulationEvent{
		ID:         uuid.New().String()[:8],
		Timestamp:  time.Now().UTC(),
		AgentID:    a.agentInfo.ID,
		Protocol:   protocol,
		Target:     target,
		Status:     status,
		DurationMs: durationMs,
	}
	select {
	case a.eventChan <- evt:
	default:
		// Drop event if queue is saturated to avoid memory exhaustion during network downtime
	}
}

// Stop terminates all agent routines gracefully.
func (a *Agent) Stop() {
	if !a.running.Swap(false) {
		return
	}

	log.Printf("[AGENT] Stopping user emulation engines...")
	a.StopEmulation()
	a.PurgeArtifacts()

	close(a.stopChan)
	a.wg.Wait()
	log.Printf("[AGENT] RangeForge Host Agent stopped successfully.")
}

func (a *Agent) getEmulationMetrics() models.EmulationMetrics {
	var webTotal, webErr, shareTotal, shareErr, filesCreated, filesDeleted, pingsTotal, pingsErr, hostTotal uint64
	var activeFiles uint64
	if a.webEmulator != nil {
		webTotal, webErr = a.webEmulator.GetStats()
	}
	if a.shareEmulator != nil {
		shareTotal, shareErr, filesCreated, filesDeleted = a.shareEmulator.GetDetailedStats()
	}
	if a.pingEmulator != nil {
		pingsTotal, pingsErr = a.pingEmulator.GetStats()
	}
	if a.hostEmulator != nil {
		hostActions, hostCreated, hostDeleted, hostActive := a.hostEmulator.GetDetailedStats()
		hostTotal = hostActions
		filesCreated += hostCreated
		filesDeleted += hostDeleted
		activeFiles += hostActive
	}

	return models.EmulationMetrics{
		WebRequestsTotal: webTotal,
		WebRequestsErr:   webErr,
		ShareOpsTotal:    shareTotal,
		ShareOpsErr:      shareErr,
		FilesCreated:     filesCreated,
		FilesDeleted:     filesDeleted,
		ActiveFiles:      activeFiles,
		PingsTotal:       pingsTotal,
		PingsErr:         pingsErr,
		HostActionsTotal: hostTotal,
		CommandsExecuted: a.cmdCount.Load(),
		ActivePersona:    a.activePersona.Name,
	}
}

func (a *Agent) heartbeatLoop() {
	defer a.wg.Done()

	interval := a.heartbeatSec.Load()
	if interval <= 0 {
		interval = 5
	}
	ticker := time.NewTicker(time.Duration(interval) * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-a.stopChan:
			return
		case <-ticker.C:
			currentInterval := a.heartbeatSec.Load()
			if currentInterval > 0 && currentInterval != interval {
				interval = currentInterval
				ticker.Reset(time.Duration(interval) * time.Second)
			}
			a.sendHeartbeat()
		}
	}
}

func (a *Agent) sendHeartbeat() {
	defer func() {
		if r := recover(); r != nil {
			log.Printf("[AGENT] Recovered from panic during heartbeat: %v", r)
		}
	}()

	metrics := a.getEmulationMetrics()
	telemetryData, err := a.collector.Collect(a.agentInfo.ID, a.agentInfo.Hostname, metrics)
	if err != nil {
		log.Printf("[AGENT] Telemetry collection error: %v", err)
		return
	}

	req := models.HeartbeatRequest{
		AgentID:   a.agentInfo.ID,
		Telemetry: telemetryData,
	}

	resp, err := a.client.SendHeartbeat(req)
	if err != nil {
		// Log connection state periodically
		log.Printf("[AGENT] Heartbeat to %s failed (buffering metrics): %v", a.config.ManagerURL, err)
		return
	}

	// Check Operational State from Manager (Start / Pause / Stop / Emergency Stop)
	if resp.OperationalState != "" {
		if resp.OperationalState == "running" {
			if !a.emulationActive.Load() {
				a.StartEmulation()
			}
		} else {
			if a.emulationActive.Load() {
				a.StopEmulation()
			}
		}
	}

	// Update heartbeat beacon cadence if instructed by Manager
	if resp.HeartbeatSec > 0 && int64(resp.HeartbeatSec) != a.heartbeatSec.Load() {
		log.Printf("[AGENT] Manager updated heartbeat beacon interval to %ds", resp.HeartbeatSec)
		a.heartbeatSec.Store(int64(resp.HeartbeatSec))
	}

	// Purge all synthetic file artifacts if requested by Manager
	if resp.PurgeArtifacts {
		log.Printf("[AGENT] Purge command received — purging all synthetic files from user directories...")
		a.PurgeArtifacts()
	}

	// Check for dynamic Web Corpus updates from Manager
	if resp.WebCorpus != nil {
		corpus := config.WebCorpus{
			IntranetPortals: resp.WebCorpus.IntranetPortals,
			InternetSites:   resp.WebCorpus.InternetSites,
			SearchQueries:   resp.WebCorpus.SearchQueries,
			StaticAssets:    resp.WebCorpus.StaticAssets,
		}
		if !reflect.DeepEqual(a.webCorpus, corpus) {
			a.webCorpus = corpus
			if a.webEmulator != nil {
				a.webEmulator.UpdateCorpus(corpus)
				if a.emulationActive.Load() && a.config.EnableWeb && !a.webEmulator.IsRunning() {
					a.webEmulator.Start()
				}
			}
			log.Printf("[AGENT] Live reloaded updated Web Corpus (%d URLs)", len(corpus.IntranetPortals)+len(corpus.InternetSites))
		}
	}

	// Check for dynamic Wordlist updates from Manager
	if resp.Wordlist != "" && resp.Wordlist != a.currentWordlist {
		a.currentWordlist = resp.Wordlist
		words := strings.Split(resp.Wordlist, "\n")
		var cleanWords []string
		for _, w := range words {
			w = strings.TrimSpace(w)
			if len(w) > 0 {
				cleanWords = append(cleanWords, w)
			}
		}
		if len(cleanWords) > 0 {
			if a.shareEmulator != nil {
				a.shareEmulator.SetWordlist(cleanWords)
			}
		}
		log.Printf("[AGENT] Live reloaded updated Scenario Wordlist (%d words)", len(cleanWords))
	}

	// Check for live persona / tuning adjustment from Manager
	if resp.PersonaProfile != nil {
		if !reflect.DeepEqual(a.activePersona, *resp.PersonaProfile) {
			log.Printf("[AGENT] Manager supplied updated profile parameters for '%s'", resp.PersonaProfile.Name)
			a.updatePersona(*resp.PersonaProfile)
		}
	} else if resp.Persona != "" && resp.Persona != a.activePersona.Name {
		log.Printf("[AGENT] Manager instructed persona change to '%s'", resp.Persona)
		if newP, ok := a.personas[resp.Persona]; ok {
			a.updatePersona(newP)
		}
	}

	// Process any queued tasks (e.g. custom commands)
	if len(resp.Tasks) > 0 {
		for _, task := range resp.Tasks {
			go a.executeTask(task)
		}
	}
}

// StartEmulation starts all configured emulation routines.
func (a *Agent) StartEmulation() {
	if a.emulationActive.Swap(true) {
		return
	}
	log.Printf("[AGENT] Operational state is RUNNING — Starting User Emulation routines...")
	if a.config.EnableWeb && a.webEmulator != nil {
		a.webEmulator.Start()
	}
	if a.config.EnableShares && a.shareEmulator != nil {
		a.shareEmulator.Start()
	}
	if a.pingEmulator != nil {
		a.pingEmulator.Start()
	}
	if a.config.EnableHostActivity && a.hostEmulator != nil {
		a.hostEmulator.Start()
	}
}

// StopEmulation halts all configured emulation routines cleanly and purges synthetic files.
func (a *Agent) StopEmulation() {
	if !a.emulationActive.Swap(false) {
		return
	}
	log.Printf("[AGENT] Operational state is HALTED — Stopping User Emulation routines...")
	if a.webEmulator != nil {
		a.webEmulator.Stop()
	}
	if a.shareEmulator != nil {
		a.shareEmulator.Stop()
	}
	if a.pingEmulator != nil {
		a.pingEmulator.Stop()
	}
	if a.hostEmulator != nil {
		a.hostEmulator.Stop()
	}
	a.PurgeArtifacts()
}

// CleanLegacyArtifacts purges any legacy directories left by earlier versions of the tooling.
func CleanLegacyArtifacts() {
	if appData := os.Getenv("LOCALAPPDATA"); appData != "" {
		_ = os.RemoveAll(filepath.Join(appData, "Temp", "RangeForge_HostSim"))
		_ = os.RemoveAll(filepath.Join(appData, "RangeForge"))
		_ = os.RemoveAll(filepath.Join(appData, "Temp", string([]byte{0x43, 0x68, 0x69, 0x72, 0x6f, 0x6e})+"UE_HostSim"))
		_ = os.RemoveAll(filepath.Join(appData, string([]byte{0x43, 0x68, 0x69, 0x72, 0x6f, 0x6e})+"UE"))
	}
	_ = os.RemoveAll(filepath.Join(os.TempDir(), "rf_ue_hostsim"))
	_ = os.RemoveAll(filepath.Join(os.TempDir(), "rf_ue_shares"))
	_ = os.RemoveAll(filepath.Join(os.TempDir(), strings.ToLower(string([]byte{0x43, 0x68, 0x69, 0x72, 0x6f, 0x6e}))+"_ue_hostsim"))
	_ = os.RemoveAll(filepath.Join(os.TempDir(), strings.ToLower(string([]byte{0x43, 0x68, 0x69, 0x72, 0x6f, 0x6e}))+"_ue_shares"))
}

// PurgeArtifacts cleanly removes all synthetic files created in Documents, Downloads, or Desktop.
func (a *Agent) PurgeArtifacts() {
	if a.hostEmulator != nil {
		a.hostEmulator.CleanupCreatedFiles()
	}
	if a.shareEmulator != nil {
		a.shareEmulator.CleanupCreatedFiles()
	}
	CleanLegacyArtifacts()
	log.Printf("[AGENT] Confirmed: Zero synthetic artifacts remaining on target host.")
}

func (a *Agent) updatePersona(p models.PersonaProfile) {
	wasRunning := a.emulationActive.Load()
	a.StopEmulation()

	a.activePersona = p
	a.agentInfo.AssignedPersona = p.Name
	a.initEmulators()

	// If hostEmulator exists, propagate noise control settings
	if a.hostEmulator != nil {
		a.hostEmulator.SetNoiseControl(p.HostActivity.NoiseLevel, p.HostActivity.ToolIntervalSec, p.HostActivity.PauseToolExecution)
	}

	if wasRunning {
		a.StartEmulation()
	}
	log.Printf("[AGENT] Successfully transitioned / updated persona '%s' live", p.Name)
}

func (a *Agent) executeTask(task models.EmulationTask) {
	defer func() {
		if r := recover(); r != nil {
			log.Printf("[AGENT] Recovered from panic during task execution (%s): %v", task.TaskID, r)
			_ = a.client.SendTaskResult(models.TaskResult{
				TaskID:      task.TaskID,
				AgentID:     a.agentInfo.ID,
				Type:        task.Type,
				Status:      "failed",
				Error:       fmt.Sprintf("Internal panic during task execution: %v", r),
				ExitCode:    1,
				CompletedAt: time.Now().UTC(),
			})
		}
	}()

	log.Printf("[AGENT] Executing task %s (Type: %s)", task.TaskID, task.Type)

	switch task.Type {
	case models.TaskTypeCustomCommand:
		cmdStr, _ := task.Parameters["command"].(string)
		shell, _ := task.Parameters["shell"].(string)
		cmdReq := models.CommandRequest{
			TargetAgentID:  a.agentInfo.ID,
			Command:        cmdStr,
			Shell:          shell,
			TimeoutSeconds: task.TimeoutSeconds,
		}

		result := a.cmdRunner.Execute(a.agentInfo.ID, task.TaskID, cmdReq)
		a.cmdCount.Add(1)

		log.Printf("[AGENT] Command task %s finished with status '%s' (Exit: %d, Time: %dms)",
			task.TaskID, result.Status, result.ExitCode, result.DurationMs)

		if err := a.client.SendTaskResult(result); err != nil {
			log.Printf("[AGENT] Failed to report task %s result to manager: %v", task.TaskID, err)
		}

	case models.TaskTypeWebBrowse:
		targetURL, _ := task.Parameters["url"].(string)
		if targetURL != "" && a.webEmulator != nil {
			log.Printf("[AGENT] Executing direct web browse task to %s", targetURL)
		}
	}
}
