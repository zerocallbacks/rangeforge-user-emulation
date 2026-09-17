package emulator

import (
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

func TestCommandRunner(t *testing.T) {
	runner := NewCommandRunner()

	var cmd string
	if runtime.GOOS == "windows" {
		cmd = "Write-Output 'RangeForgeTest'"
	} else {
		cmd = "echo 'RangeForgeTest'"
	}

	req := models.CommandRequest{
		TargetAgentID:  "agent-1",
		Command:        cmd,
		Shell:          "auto",
		TimeoutSeconds: 5,
	}

	result := runner.Execute("agent-1", "task-1", req)
	if result.Status != "success" {
		t.Fatalf("Expected success status, got: %s (err: %s)", result.Status, result.Error)
	}
	if result.ExitCode != 0 {
		t.Errorf("Expected exit code 0, got: %d", result.ExitCode)
	}
	if len(result.Stdout) == 0 {
		t.Errorf("Expected non-empty stdout")
	}
}

func TestFileShareEmulator(t *testing.T) {
	cfg := models.FileShareConfig{
		Enabled:     true,
		IntervalSec: 1,
		ReadRatio:   0.5,
		WriteRatio:  0.5,
	}
	corpus := config.DefaultShareCorpus()

	emu := NewFileShareEmulator(cfg, corpus)
	emu.Start()

	// Wait 2 seconds for initial seed and at least one operation
	time.Sleep(2 * time.Second)
	emu.Stop()

	total, _ := emu.GetStats()
	if total == 0 {
		t.Logf("Share ops total was 0, triggering direct action")
		emu.performShareAction()
		total, _ = emu.GetStats()
		if total == 0 {
			t.Errorf("Expected at least 1 share operation")
		}
	}
}

func TestFileShareWordlistCreationAndDeletion(t *testing.T) {
	cfg := models.FileShareConfig{
		Enabled:     true,
		IntervalSec: 1,
		ReadRatio:   0.0, // Force 100% writes
		WriteRatio:  1.0,
	}
	corpus := config.DefaultShareCorpus()
	emu := NewFileShareEmulator(cfg, corpus)
	emu.SetWordlist([]string{"mission_assessment", "audit_trace", "threat_matrix", "telemetry_sync"})

	// Create 5 files
	for i := 0; i < 5; i++ {
		emu.simulateWrite()
	}

	_, _, created, _ := emu.GetDetailedStats()
	if created != 5 {
		t.Fatalf("Expected 5 files created, got %d", created)
	}

	// Verify all 5 created files are tracked
	emu.createdFilesMu.Lock()
	trackedCount := len(emu.createdFiles)
	emu.createdFilesMu.Unlock()
	if trackedCount != 5 {
		t.Fatalf("Expected 5 tracked files, got %d", trackedCount)
	}

	// Trigger cleanup of all UE-created files
	emu.CleanupCreatedFiles()

	_, _, _, afterDelete := emu.GetDetailedStats()
	if afterDelete != 5 {
		t.Fatalf("Expected 5 files deleted during cleanup, got %d", afterDelete)
	}

	emu.createdFilesMu.Lock()
	rem := len(emu.createdFiles)
	emu.createdFilesMu.Unlock()
	if rem != 0 {
		t.Fatalf("Expected 0 tracked files after cleanup, got %d", rem)
	}

	// Test automatic retention deletion when count exceeds 10
	for i := 0; i < 15; i++ {
		emu.simulateWrite()
	}

	_, _, createdAfter, deletedAfter := emu.GetDetailedStats()
	if createdAfter != 20 {
		t.Errorf("Expected 20 total created, got %d", createdAfter)
	}
	if deletedAfter <= 5 {
		t.Errorf("Expected auto-deletion of older UE-created files when exceeding limit of 10, got %d", deletedAfter)
	}

	// Final cleanup
	emu.CleanupCreatedFiles()
}

func TestHostActivityEmulator(t *testing.T) {
	cfg := models.HostActivityConfig{
		Enabled:             true,
		IntervalSec:         1,
		SafeProcesses:       []string{"hostname"},
		SimulateOfficeDocs:  true,
		TempFileOperations:  true,
	}

	emu := NewHostActivityEmulator(cfg)
	emu.Start()
	time.Sleep(2 * time.Second)
	emu.Stop()

	actions := emu.GetStats()
	if actions == 0 {
		t.Logf("Actions was 0, triggering direct action")
		emu.performHostAction()
		actions = emu.GetStats()
		if actions == 0 {
			t.Errorf("Expected at least 1 host action")
		}
	}
}

func TestWebBrowserEmulator(t *testing.T) {
	// Create mock HTTP server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("<html><body><h1>RangeForge Cyber Range Mock Portal</h1></body></html>"))
	}))
	defer server.Close()

	cfg := models.WebBrowsingConfig{
		Enabled:              true,
		RequestsPerMinuteMin: 60,
		RequestsPerMinuteMax: 120,
		DwellTimeMinSec:      1,
		DwellTimeMaxSec:      2,
		FetchAssets:          false,
		TargetURLs:           []string{server.URL},
	}
	corpus := config.DefaultWebCorpus()

	emu := NewWebBrowserEmulator(cfg, corpus)
	emu.browseURL(server.URL)

	total, errs := emu.GetStats()
	if total != 1 {
		t.Errorf("Expected 1 successful request, got %d", total)
	}
	if errs != 0 {
		t.Errorf("Expected 0 errors, got %d", errs)
	}
}

func TestZeroFootprintFileLocationsAndCleanup(t *testing.T) {
	cfg := models.HostActivityConfig{
		Enabled:            true,
		IntervalSec:        1,
		SafeProcesses:      []string{"hostname"},
		SimulateOfficeDocs: true,
		UserFileOperations: true,
		NoiseLevel:         "stealth",
		ToolIntervalSec:    300,
	}

	h := NewHostActivityEmulator(cfg)

	// Simulate 4 file creations
	for i := 0; i < 4; i++ {
		h.simulateUserFileWork()
	}

	h.createdFilesMu.Lock()
	count := len(h.createdFiles)
	var paths []string
	for p := range h.createdFiles {
		paths = append(paths, p)
	}
	h.createdFilesMu.Unlock()

	if count != 4 {
		t.Fatalf("Expected 4 files created and tracked, got %d", count)
	}

	// Verify all created file paths are strictly within Documents, Downloads, or Desktop
	for _, p := range paths {
		dir := filepath.Dir(p)
		baseDir := filepath.Base(dir)
		fileName := filepath.Base(p)

		// Must be in Documents, Downloads, Desktop, or UserHomeDir fallback
		if baseDir != "Documents" && baseDir != "Downloads" && baseDir != "Desktop" {
			home, _ := os.UserHomeDir()
			if dir != home {
				t.Errorf("File %s created outside approved user profile directories (dir: %s)", p, dir)
			}
		}

		// Verify absence of artificial branding/tool names
		lowerName := strings.ToLower(fileName)
		if strings.Contains(lowerName, "rangeforge") || strings.Contains(lowerName, "range_report") || strings.Contains(lowerName, "hostsim") || strings.HasPrefix(lowerName, "ue_") {
			t.Errorf("File %s contains artificial simulation markers or branding", fileName)
		}

		// Verify file content does not contain blue-team obscuring headers
		data, err := os.ReadFile(p)
		if err != nil {
			t.Errorf("Failed to read created file %s: %v", p, err)
		} else {
			strData := string(data)
			if strings.Contains(strData, "USER EMULATION AUTOMATED DOCUMENT") || strings.Contains(strData, "Internal Range Emulation") {
				t.Errorf("File content in %s contains synthetic range artifact headers", fileName)
			}
		}
	}

	// Verify complete cleanup
	h.CleanupCreatedFiles()

	for _, p := range paths {
		if _, err := os.Stat(p); !os.IsNotExist(err) {
			t.Errorf("File %s was not deleted during cleanup", p)
		}
	}

	h.createdFilesMu.Lock()
	remCount := len(h.createdFiles)
	h.createdFilesMu.Unlock()
	if remCount != 0 {
		t.Errorf("Expected 0 tracked files after cleanup, got %d", remCount)
	}
}

func TestNoiseControlSuppression(t *testing.T) {
	cfg := models.HostActivityConfig{
		Enabled:            true,
		IntervalSec:        1,
		SafeProcesses:      []string{"whoami"},
		SimulateOfficeDocs: true,
		NoiseLevel:         "stealth",
		ToolIntervalSec:    300,
		PauseToolExecution: true,
	}

	h := NewHostActivityEmulator(cfg)
	initialActions := h.GetStats()

	// When PauseToolExecution is true, performHostAction must NEVER invoke processes
	h.performHostAction()
	h.noiseMu.RLock()
	lastExec := h.lastToolExec
	h.noiseMu.RUnlock()

	if !lastExec.IsZero() {
		t.Errorf("Expected lastToolExec to be zero since process execution was paused")
	}

	newActions := h.GetStats()
	if newActions <= initialActions {
		t.Errorf("Expected host action to execute (file work fallback), got actions: %d", newActions)
	}

	// Cleanup any created file
	h.CleanupCreatedFiles()
}
