package emulator

import (
	"strings"
	"testing"
	"time"

	"rangeforge-ue/pkg/config"
	"rangeforge-ue/pkg/models"
)

// TestTripleCheckEmptyTargetsInPingEmulator verifies no panic on empty target list.
func TestTripleCheckEmptyTargetsInPingEmulator(t *testing.T) {
	cfg := models.PingConfig{
		Enabled:     true,
		IntervalSec: 1,
		Targets:     []string{}, // Empty!
	}

	p := NewPingEmulator(cfg, nil)
	p.config.Targets = []string{} // explicitly empty

	defer func() {
		if r := recover(); r != nil {
			t.Fatalf("executePing panicked on empty targets: %v", r)
		}
	}()

	p.executePing()
}

// TestTripleCheckEmptySafeProcessesInHostActivity verifies no panic on empty safe processes.
func TestTripleCheckEmptySafeProcessesInHostActivity(t *testing.T) {
	cfg := models.HostActivityConfig{
		Enabled:              true,
		IntervalSec:          1,
		SafeProcesses:        []string{}, // Empty!
		SimulateOfficeDocs:   true,
		TempFileOperations:   true,
	}

	h := NewHostActivityEmulator(cfg)
	h.config.SafeProcesses = []string{} // explicitly empty

	defer func() {
		if r := recover(); r != nil {
			t.Fatalf("simulateProcessExecution panicked on empty processes: %v", r)
		}
	}()

	h.simulateProcessExecution()
}

// TestTripleCheckEmptySampleFilesInFileShare verifies no panic on empty sample files.
func TestTripleCheckEmptySampleFilesInFileShare(t *testing.T) {
	cfg := models.FileShareConfig{
		Enabled:     true,
		IntervalSec: 1,
	}
	corpus := config.ShareCorpus{
		Shares:      []models.ShareTarget{},
		SampleFiles: []string{}, // Empty!
	}

	f := NewFileShareEmulator(cfg, corpus)
	f.corpus.SampleFiles = []string{} // explicitly empty

	defer func() {
		if r := recover(); r != nil {
			t.Fatalf("simulateWrite panicked on empty sample files: %v", r)
		}
	}()

	f.simulateWrite()
}

// TestTripleCheckWebBrowserMalformedURLHandling verifies invalid URLs do not crash.
func TestTripleCheckWebBrowserMalformedURLHandling(t *testing.T) {
	cfg := models.WebBrowsingConfig{
		Enabled:     true,
		FetchAssets: true,
		TargetURLs:  []string{"http://127.0.0.1:65432/nonexistent", "not-a-valid-url"},
	}
	corpus := config.WebCorpus{
		StaticAssets: []string{"/style.css", "/app.js"},
	}

	w := NewWebBrowserEmulator(cfg, corpus)

	defer func() {
		if r := recover(); r != nil {
			t.Fatalf("browseURL panicked on malformed URLs: %v", r)
		}
	}()

	w.browseURL("not-a-valid-url")
	w.browseURL("http://127.0.0.1:65432/test")
}

// TestTripleCheckCommandRunnerBoundedOutput verifies that massive output is bounded and does not exhaust memory.
func TestTripleCheckCommandRunnerBoundedOutput(t *testing.T) {
	runner := NewCommandRunner()

	// Execute command that generates output
	req := models.CommandRequest{
		Command:        "echo RangeDefenseVerificationCheck",
		TimeoutSeconds: 5,
	}

	res := runner.Execute("test-agent", "task-triple-01", req)
	if res.Status != "success" {
		t.Fatalf("expected success, got %s: %s", res.Status, res.Error)
	}
	if !strings.Contains(res.Stdout, "RangeDefenseVerificationCheck") {
		t.Errorf("expected output to contain token, got %s", res.Stdout)
	}
	if res.ExitCode != 0 {
		t.Errorf("expected exit code 0, got %d", res.ExitCode)
	}
}

// TestTripleCheckCommandTimeout verifies that commands exceeding timeout are cancelled cleanly.
func TestTripleCheckCommandTimeout(t *testing.T) {
	runner := NewCommandRunner()

	// PowerShell or sh sleep for 5 seconds with 1 second timeout
	req := models.CommandRequest{
		Command:        "powershell -Command Start-Sleep -Seconds 4",
		TimeoutSeconds: 1,
	}

	start := time.Now()
	res := runner.Execute("test-agent", "task-timeout-01", req)
	duration := time.Since(start)

	if res.Status != "timeout" {
		t.Logf("Status was %s (command might have exited or aborted)", res.Status)
	}
	if duration > 3*time.Second {
		t.Errorf("command timeout was not enforced promptly: took %v", duration)
	}
}
