package emulator

import (
	"context"
	"fmt"
	"math/rand"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"rangeforge-ue/pkg/models"
)

// Realistic business document templates for zero-footprint host activity
var corporateDocBases = []string{
	"Q3_Financial_Review",
	"Annual_Budget_Allocation",
	"Department_Expenses_Summary",
	"Vendor_Rate_Card",
	"Staff_Directory_Update",
	"Project_Milestone_Plan",
	"Client_Proposal_Draft",
	"Employee_Onboarding_Checklist",
	"Security_Policy_Review",
	"Executive_Summary_Notes",
	"Operations_Meeting_Minutes",
	"System_Architecture_Plan",
	"Quarterly_Tax_Forecast",
	"Product_Roadmap_Draft",
}

// HostActivityEmulator simulates on-host user activity strictly in user profile folders
// (Documents, Downloads, Desktop) with complete artifact cleanup and Blue Team noise controls.
type HostActivityEmulator struct {
	config             models.HostActivityConfig
	running            atomic.Bool
	stopChan           chan struct{}
	wg                 sync.WaitGroup
	actionsCount       atomic.Uint64
	filesCreated       atomic.Uint64
	filesDeleted       atomic.Uint64
	rng                *rand.Rand
	rngMu              sync.Mutex
	onEvent            func(protocol, target, status string, durationMs int64)

	// User directories for realistic, zero-footprint storage
	userDirs           []string

	// Thread-safe registry of all synthetic files created on disk
	createdFilesMu     sync.Mutex
	createdFiles       map[string]time.Time

	// Blue Team noise & binary polling control
	noiseMu            sync.RWMutex
	noiseLevel         string
	toolIntervalSec    int
	pauseToolExecution bool
	lastToolExec       time.Time
}

// SetEventCallback registers a callback for live event logging.
func (h *HostActivityEmulator) SetEventCallback(cb func(protocol, target, status string, durationMs int64)) {
	h.onEvent = cb
}

// GetUserDirs resolves genuine user profile locations (Documents, Downloads, Desktop).
func GetUserDirs() []string {
	home, err := os.UserHomeDir()
	if err != nil || home == "" {
		home = os.Getenv("USERPROFILE")
		if home == "" {
			home = os.Getenv("HOME")
		}
	}
	if home == "" {
		home = os.TempDir()
	}

	targets := []string{
		filepath.Join(home, "Documents"),
		filepath.Join(home, "Downloads"),
		filepath.Join(home, "Desktop"),
	}

	var valid []string
	for _, d := range targets {
		if fi, err := os.Stat(d); err == nil && fi.IsDir() {
			valid = append(valid, d)
		} else {
			// Ensure standard user folders exist if accessible
			if err := os.MkdirAll(d, 0755); err == nil {
				valid = append(valid, d)
			}
		}
	}
	if len(valid) == 0 {
		valid = append(valid, home)
	}
	return valid
}

// NewHostActivityEmulator creates an instance of HostActivityEmulator.
func NewHostActivityEmulator(cfg models.HostActivityConfig) *HostActivityEmulator {
	userDirs := GetUserDirs()

	noiseLevel := cfg.NoiseLevel
	if noiseLevel == "" {
		noiseLevel = "balanced"
	}

	toolInterval := cfg.ToolIntervalSec
	if toolInterval <= 0 {
		switch noiseLevel {
		case "stealth":
			toolInterval = 180
		case "active":
			toolInterval = 20
		default:
			toolInterval = 60
		}
	}

	return &HostActivityEmulator{
		config:             cfg,
		stopChan:           make(chan struct{}),
		userDirs:           userDirs,
		createdFiles:       make(map[string]time.Time),
		rng:                rand.New(rand.NewSource(time.Now().UnixNano())),
		noiseLevel:         noiseLevel,
		toolIntervalSec:    toolInterval,
		pauseToolExecution: cfg.PauseToolExecution,
	}
}

// SetNoiseControl updates the noise level, tool interval, and pause toggle dynamically.
func (h *HostActivityEmulator) SetNoiseControl(noiseLevel string, toolIntervalSec int, pauseTools bool) {
	h.noiseMu.Lock()
	defer h.noiseMu.Unlock()
	if noiseLevel != "" {
		h.noiseLevel = noiseLevel
	}
	if toolIntervalSec > 0 {
		h.toolIntervalSec = toolIntervalSec
	}
	h.pauseToolExecution = pauseTools
}

// Start launches the host simulation loop.
func (h *HostActivityEmulator) Start() {
	if !h.config.Enabled {
		return
	}
	if h.running.Swap(true) {
		return
	}

	h.stopChan = make(chan struct{})
	h.wg.Add(1)
	go h.simulationLoop()
}

// Stop terminates the host activity loop and automatically cleans up all created files.
func (h *HostActivityEmulator) Stop() {
	if h.running.Swap(false) {
		close(h.stopChan)
		h.wg.Wait()
		// Guarantee complete removal of any synthetic files created in user directories
		h.CleanupCreatedFiles()
	}
}

// GetStats returns actions count.
func (h *HostActivityEmulator) GetStats() uint64 {
	return h.actionsCount.Load()
}

// GetDetailedStats returns total actions, files created, and files deleted.
func (h *HostActivityEmulator) GetDetailedStats() (actions, created, deleted, active uint64) {
	h.createdFilesMu.Lock()
	activeCount := uint64(len(h.createdFiles))
	h.createdFilesMu.Unlock()
	return h.actionsCount.Load(), h.filesCreated.Load(), h.filesDeleted.Load(), activeCount
}

// CleanupCreatedFiles purges every synthetic file created by this emulator on target hosts.
func (h *HostActivityEmulator) CleanupCreatedFiles() {
	h.createdFilesMu.Lock()
	toDelete := make([]string, 0, len(h.createdFiles))
	for p := range h.createdFiles {
		toDelete = append(toDelete, p)
	}
	h.createdFiles = make(map[string]time.Time)
	h.createdFilesMu.Unlock()

	for _, p := range toDelete {
		start := time.Now()
		if err := os.Remove(p); err == nil || os.IsNotExist(err) {
			h.filesDeleted.Add(1)
			h.actionsCount.Add(1)
			dur := time.Since(start).Milliseconds()
			if h.onEvent != nil {
				h.onEvent("FILE_DELETE", filepath.Base(p), "PURGED", dur)
			}
		}
	}
}

func (h *HostActivityEmulator) simulationLoop() {
	defer h.wg.Done()

	interval := h.config.IntervalSec
	if interval <= 0 {
		interval = 20
	}

	for {
		select {
		case <-h.stopChan:
			return
		default:
		}

		h.performHostAction()

		h.rngMu.Lock()
		jitter := interval + h.rng.Intn(interval/2+1) - (interval / 4)
		h.rngMu.Unlock()
		if jitter < 5 {
			jitter = 5
		}

		select {
		case <-h.stopChan:
			return
		case <-time.After(time.Duration(jitter) * time.Second):
		}
	}
}

func (h *HostActivityEmulator) performHostAction() {
	h.noiseMu.RLock()
	pauseTools := h.pauseToolExecution
	noiseLevel := h.noiseLevel
	toolInterval := h.toolIntervalSec
	lastExec := h.lastToolExec
	h.noiseMu.RUnlock()

	h.rngMu.Lock()
	actionType := h.rng.Intn(2)
	h.rngMu.Unlock()

	// If tools are explicitly paused, or if it's stealth mode, prioritize benign file activity
	if pauseTools {
		actionType = 0
	}

	switch actionType {
	case 0:
		if h.config.SimulateOfficeDocs || h.config.TempFileOperations || h.config.UserFileOperations {
			h.simulateUserFileWork()
		}
	case 1:
		if !pauseTools && len(h.config.SafeProcesses) > 0 {
			// Check if sufficient time has elapsed based on Blue Team tool execution interval
			minPacing := time.Duration(toolInterval) * time.Second
			if noiseLevel == "stealth" && minPacing < 120*time.Second {
				minPacing = 120 * time.Second
			}
			if time.Since(lastExec) >= minPacing {
				h.simulateProcessExecution()
			} else {
				// Too frequent; fallback to quiet user file work so Blue Team logs aren't spammed
				h.simulateUserFileWork()
			}
		} else {
			h.simulateUserFileWork()
		}
	}
}

func (h *HostActivityEmulator) pickUserDirectory() string {
	if len(h.userDirs) == 0 {
		return os.TempDir()
	}
	h.rngMu.Lock()
	d := h.userDirs[h.rng.Intn(len(h.userDirs))]
	h.rngMu.Unlock()
	return d
}

func (h *HostActivityEmulator) simulateUserFileWork() {
	targetDir := h.pickUserDirectory()

	h.rngMu.Lock()
	baseName := corporateDocBases[h.rng.Intn(len(corporateDocBases))]
	exts := []string{"docx", "xlsx", "pdf", "txt"}
	ext := exts[h.rng.Intn(len(exts))]
	isLockFile := h.config.SimulateOfficeDocs && (h.rng.Intn(3) == 0)
	h.rngMu.Unlock()

	var fileName string
	var payload []byte

	currentUsername := os.Getenv("USERNAME")
	if currentUsername == "" {
		currentUsername = os.Getenv("USER")
	}
	if currentUsername == "" {
		currentUsername = "user"
	}

	if isLockFile {
		// Realistic Microsoft Office lock file (e.g. ~$Q3_Financial_Review.docx)
		// Genuine Office lock files are padded buffers containing the current Windows username
		fileName = fmt.Sprintf("~$%s.%s", baseName, ext)
		lockContent := fmt.Sprintf("%-162s", currentUsername)
		payload = []byte(lockContent)
	} else {
		// Realistic business document name and content
		fileName = fmt.Sprintf("%s.%s", baseName, ext)
		content := fmt.Sprintf(
			"CONFIDENTIAL - INTERNAL CORPORATE USE ONLY\n\n"+
				"DOCUMENT: %s\n"+
				"DEPARTMENT: Operations & Planning\n"+
				"REVIEW DATE: %s\n"+
				"AUTHOR: %s\n\n"+
				"EXECUTIVE SUMMARY:\n"+
				"This operational summary compiles current metrics, project milestones, and resource planning data for the ongoing fiscal period.\n\n"+
				"KEY TOPICS & INITIATIVES:\n"+
				"- Departmental budget allocation and quarterly forecast alignment.\n"+
				"- Infrastructure maintenance windows and security compliance checks.\n"+
				"- Cross-team operational objectives and milestone deliverables.\n\n"+
				"Status: Review Completed\n",
			fileName,
			time.Now().Format("January 02, 2006"),
			currentUsername,
		)
		payload = []byte(content)
	}

	filePath := filepath.Join(targetDir, fileName)

	start := time.Now()
	if err := os.WriteFile(filePath, payload, 0644); err == nil {
		dur := time.Since(start).Milliseconds()
		h.actionsCount.Add(1)
		h.filesCreated.Add(1)

		// Record in thread-safe registry
		h.createdFilesMu.Lock()
		h.createdFiles[filePath] = time.Now()

		// Bounded rotation: keep at most 6 synthetic files active in user folders
		var oldestPath string
		if len(h.createdFiles) > 6 {
			var oldestTime time.Time
			for p, t := range h.createdFiles {
				if oldestPath == "" || t.Before(oldestTime) {
					oldestPath = p
					oldestTime = t
				}
			}
			if oldestPath != "" {
				delete(h.createdFiles, oldestPath)
			}
		}
		h.createdFilesMu.Unlock()

		if h.onEvent != nil {
			h.onEvent("FILE_WRITE", filepath.Join(filepath.Base(targetDir), fileName), "COMPLETED", dur)
		}

		// Delete rotated oldest file
		if oldestPath != "" {
			delStart := time.Now()
			_ = os.Remove(oldestPath)
			h.filesDeleted.Add(1)
			h.actionsCount.Add(1)
			delDur := time.Since(delStart).Milliseconds()
			if h.onEvent != nil {
				h.onEvent("FILE_ROTATE", filepath.Base(oldestPath), "PURGED", delDur)
			}
		}
	}
}

func (h *HostActivityEmulator) simulateProcessExecution() {
	h.rngMu.Lock()
	if len(h.config.SafeProcesses) == 0 {
		h.rngMu.Unlock()
		return
	}
	procCmd := h.config.SafeProcesses[h.rng.Intn(len(h.config.SafeProcesses))]
	h.rngMu.Unlock()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	parts := strings.Fields(procCmd)
	if len(parts) == 0 {
		return
	}

	binary := parts[0]
	args := parts[1:]

	if binary == "ping" {
		if runtime.GOOS == "windows" {
			args = []string{"-n", "1", "127.0.0.1"}
		} else {
			args = []string{"-c", "1", "127.0.0.1"}
		}
	}

	start := time.Now()
	cmd := exec.CommandContext(ctx, binary, args...)
	_ = cmd.Run()
	duration := time.Since(start).Milliseconds()

	h.noiseMu.Lock()
	h.lastToolExec = time.Now()
	h.noiseMu.Unlock()

	h.actionsCount.Add(1)
	if h.onEvent != nil {
		h.onEvent("HOST_PROC", procCmd, "SUCCESS", duration)
	}
}
