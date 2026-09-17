package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestDefaultConfigs(t *testing.T) {
	mgr := DefaultManagerConfig()
	if mgr.ListenPort != 8443 {
		t.Errorf("Expected port 8443, got %d", mgr.ListenPort)
	}
	if !mgr.TLSEnabled {
		t.Errorf("Expected TLSEnabled to be true")
	}

	agent := DefaultAgentConfig()
	if agent.HeartbeatSec != 5 {
		t.Errorf("Expected heartbeat 5s, got %d", agent.HeartbeatSec)
	}
	if agent.ManagerURL != "https://127.0.0.1:8443" {
		t.Errorf("Expected https://127.0.0.1:8443, got %s", agent.ManagerURL)
	}

	ctrl := DefaultControllerConfig()
	if ctrl.ManagerURL != "https://127.0.0.1:8443" {
		t.Errorf("Expected https://127.0.0.1:8443, got %s", ctrl.ManagerURL)
	}
}

func TestCorpusLoading(t *testing.T) {
	web := DefaultWebCorpus()
	if len(web.IntranetPortals) == 0 {
		t.Errorf("Expected non-empty intranet portals")
	}
	if len(web.InternetSites) == 0 {
		t.Errorf("Expected non-empty internet sites")
	}

	creds := DefaultUserCredentials()
	if len(creds) == 0 {
		t.Errorf("Expected non-empty user creds")
	}

	shares := DefaultShareCorpus()
	if len(shares.Shares) == 0 {
		t.Errorf("Expected non-empty shares list")
	}
}

func TestYAMLConfigRoundtrip(t *testing.T) {
	tmpDir := t.TempDir()
	mgrPath := filepath.Join(tmpDir, "manager.yaml")

	yamlContent := `
listen_host: "127.0.0.1"
listen_port: 9090
heartbeat_timeout_sec: 45
default_persona: "developer"
allow_unsupported_os: true
`
	if err := os.WriteFile(mgrPath, []byte(yamlContent), 0644); err != nil {
		t.Fatalf("failed to write test yaml: %v", err)
	}

	loaded, err := LoadManagerConfig(mgrPath)
	if err != nil {
		t.Fatalf("LoadManagerConfig error: %v", err)
	}

	if loaded.ListenPort != 9090 {
		t.Errorf("Expected port 9090, got %d", loaded.ListenPort)
	}
	if loaded.DefaultPersona != "developer" {
		t.Errorf("Expected developer, got %s", loaded.DefaultPersona)
	}
	if !loaded.AllowUnsupportedOS {
		t.Errorf("Expected allow_unsupported_os to be true")
	}
}

func TestBuiltinPersonas(t *testing.T) {
	personas := BuiltinPersonas()
	if len(personas) < 3 {
		t.Errorf("Expected at least 3 builtin personas, got %d", len(personas))
	}
	if _, ok := personas["office_worker"]; !ok {
		t.Errorf("office_worker persona missing")
	}
	if _, ok := personas["developer"]; !ok {
		t.Errorf("developer persona missing")
	}
	if _, ok := personas["sysadmin"]; !ok {
		t.Errorf("sysadmin persona missing")
	}
}
