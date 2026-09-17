package config

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"

	"rangeforge-ue/pkg/models"
	"gopkg.in/yaml.v3"
)

// ManagerConfig configuration for Manager daemon.
type ManagerConfig struct {
	ListenHost          string `yaml:"listen_host" json:"listen_host"`
	ListenPort          int    `yaml:"listen_port" json:"listen_port"`
	HeartbeatTimeoutSec int    `yaml:"heartbeat_timeout_sec" json:"heartbeat_timeout_sec"`
	DefaultPersona      string `yaml:"default_persona" json:"default_persona"`
	ProfilesDir         string `yaml:"profiles_dir" json:"profiles_dir"`
	AllowUnsupportedOS  bool   `yaml:"allow_unsupported_os" json:"allow_unsupported_os"`
	TLSEnabled          bool   `yaml:"tls_enabled" json:"tls_enabled"`
	TLSCertPath         string `yaml:"tls_cert_path" json:"tls_cert_path"`
	TLSKeyPath          string `yaml:"tls_key_path" json:"tls_key_path"`
}

// AgentConfig configuration for Host Agent daemon.
type AgentConfig struct {
	ManagerURL          string   `yaml:"manager_url" json:"manager_url"`
	HeartbeatSec        int      `yaml:"heartbeat_sec" json:"heartbeat_sec"`
	BindInterface       string   `yaml:"bind_interface" json:"bind_interface"`
	Persona             string   `yaml:"persona" json:"persona"`
	WebCorpusFile       string   `yaml:"web_corpus_file" json:"web_corpus_file"`
	CredsFile           string   `yaml:"creds_file" json:"creds_file"`
	SharesFile          string   `yaml:"shares_file" json:"shares_file"`
	EnableWeb           bool     `yaml:"enable_web" json:"enable_web"`
	EnableShares        bool     `yaml:"enable_shares" json:"enable_shares"`
	EnableHostActivity  bool     `yaml:"enable_host_activity" json:"enable_host_activity"`
	Tags                []string `yaml:"tags" json:"tags"`
	AllowUnsupportedOS  bool     `yaml:"allow_unsupported_os" json:"allow_unsupported_os"`
}

// ControllerConfig configuration for Controller CLI.
type ControllerConfig struct {
	ManagerURL         string `yaml:"manager_url" json:"manager_url"`
	TimeoutSec         int    `yaml:"timeout_sec" json:"timeout_sec"`
	AllowUnsupportedOS bool   `yaml:"allow_unsupported_os" json:"allow_unsupported_os"`
}

// DefaultManagerConfig returns production defaults for Unix manager.
func DefaultManagerConfig() ManagerConfig {
	return ManagerConfig{
		ListenHost:          "0.0.0.0",
		ListenPort:          8443,
		HeartbeatTimeoutSec: 30,
		DefaultPersona:      "office_worker",
		ProfilesDir:         "./configs/profiles",
		AllowUnsupportedOS:  false,
		TLSEnabled:          true,
		TLSCertPath:         "./certs/server.crt",
		TLSKeyPath:          "./certs/server.key",
	}
}

// DefaultAgentConfig returns standard defaults for cross-platform agent.
func DefaultAgentConfig() AgentConfig {
	return AgentConfig{
		ManagerURL:         "https://127.0.0.1:8443",
		HeartbeatSec:       5,
		BindInterface:      "", // Auto-detect primary interface
		Persona:            "office_worker",
		EnableWeb:          true,
		EnableShares:       true,
		EnableHostActivity: true,
		Tags:               []string{"range-host"},
		AllowUnsupportedOS: false,
	}
}

// DefaultControllerConfig returns standard defaults for Unix controller.
func DefaultControllerConfig() ControllerConfig {
	return ControllerConfig{
		ManagerURL:         "https://127.0.0.1:8443",
		TimeoutSec:         10,
		AllowUnsupportedOS: false,
	}
}

// BuiltinPersonas returns pre-configured cyber range user profiles.
func BuiltinPersonas() map[string]models.PersonaProfile {
	return map[string]models.PersonaProfile{
		"office_worker": {
			Name:        "office_worker",
			Description: "Simulates typical enterprise office knowledge worker browsing intranet, working with docs, accessing file shares.",
			WebBrowsing: models.WebBrowsingConfig{
				Enabled:              true,
				RequestsPerMinuteMin: 4,
				RequestsPerMinuteMax: 12,
				DwellTimeMinSec:      5,
				DwellTimeMaxSec:      25,
				FetchAssets:          true,
				TargetURLs: []string{
					"http://portal.range.local",
					"https://en.wikipedia.org/wiki/Office_management",
					"https://www.cnn.com",
					"http://intranet.corp.local",
				},
				SearchKeywords: []string{
					"quarterly budget analysis",
					"travel expense report template",
					"system compliance checklist",
				},
			},
			FileShare: models.FileShareConfig{
				Enabled:     true,
				IntervalSec: 30,
				ReadRatio:   0.8,
				WriteRatio:  0.2,
			},
			HostActivity: models.HostActivityConfig{
				Enabled:            true,
				IntervalSec:        20,
				SafeProcesses:      []string{"whoami", "hostname"},
				SimulateOfficeDocs: true,
				TempFileOperations: true,
				UserFileOperations: true,
				NoiseLevel:         "balanced",
				ToolIntervalSec:    60,
				PauseToolExecution: false,
			},
		},
		"developer": {
			Name:        "developer",
			Description: "Simulates software developer checking code repos, technical documentation, API endpoints, git activity.",
			WebBrowsing: models.WebBrowsingConfig{
				Enabled:              true,
				RequestsPerMinuteMin: 6,
				RequestsPerMinuteMax: 18,
				DwellTimeMinSec:      3,
				DwellTimeMaxSec:      15,
				FetchAssets:          true,
				TargetURLs: []string{
					"https://github.com/trending",
					"https://stackoverflow.com/questions",
					"https://docs.python.org/3/",
					"https://golang.org/doc/",
					"http://jira.corp.local/secure/Dashboard.jspa",
				},
				SearchKeywords: []string{
					"golang net/http connection pool reuse",
					"git rebase interactive merge conflict",
					"docker build multi-stage alpine",
				},
			},
			FileShare: models.FileShareConfig{
				Enabled:     true,
				IntervalSec: 25,
				ReadRatio:   0.7,
				WriteRatio:  0.3,
			},
			HostActivity: models.HostActivityConfig{
				Enabled:            true,
				IntervalSec:        15,
				SafeProcesses:      []string{"whoami", "git --version", "go version"},
				SimulateOfficeDocs: false,
				TempFileOperations: true,
				UserFileOperations: true,
				NoiseLevel:         "balanced",
				ToolIntervalSec:    60,
				PauseToolExecution: false,
			},
		},
		"sysadmin": {
			Name:        "sysadmin",
			Description: "Simulates IT / Range administrator performing network diagnostics, querying servers, managing shares.",
			WebBrowsing: models.WebBrowsingConfig{
				Enabled:              true,
				RequestsPerMinuteMin: 5,
				RequestsPerMinuteMax: 15,
				DwellTimeMinSec:      4,
				DwellTimeMaxSec:      20,
				FetchAssets:          false,
				TargetURLs: []string{
					"http://portal.range.local",
					"https://docs.netgate.com/pfsense/en/latest/",
					"https://learn.microsoft.com/en-us/windows-server/",
				},
				SearchKeywords: []string{
					"pfsense gateway routing table troubleshooting",
					"active directory replication error 1722",
					"event log id 4625 brute force analysis",
				},
			},
			FileShare: models.FileShareConfig{
				Enabled:     true,
				IntervalSec: 15,
				ReadRatio:   0.6,
				WriteRatio:  0.4,
			},
			HostActivity: models.HostActivityConfig{
				Enabled:            true,
				IntervalSec:        15,
				SafeProcesses:      []string{"netstat -rn", "hostname", "uptime"},
				SimulateOfficeDocs: false,
				TempFileOperations: true,
				UserFileOperations: true,
				NoiseLevel:         "balanced",
				ToolIntervalSec:    60,
				PauseToolExecution: false,
			},
		},
	}
}

// LoadManagerConfig reads a file or returns default.
func LoadManagerConfig(path string) (ManagerConfig, error) {
	cfg := DefaultManagerConfig()
	if path == "" {
		return cfg, nil
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return cfg, err
	}
	data = bytes.TrimPrefix(data, []byte("\xef\xbb\xbf"))
	if strings.HasSuffix(path, ".json") {
		err = json.Unmarshal(data, &cfg)
	} else {
		err = yaml.Unmarshal(data, &cfg)
	}
	return cfg, err
}

// LoadAgentConfig reads a file or returns default.
func LoadAgentConfig(path string) (AgentConfig, error) {
	cfg := DefaultAgentConfig()
	if path == "" {
		return cfg, nil
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return cfg, err
	}
	data = bytes.TrimPrefix(data, []byte("\xef\xbb\xbf"))
	if strings.HasSuffix(path, ".json") {
		err = json.Unmarshal(data, &cfg)
	} else {
		err = yaml.Unmarshal(data, &cfg)
	}
	return cfg, err
}

// LoadControllerConfig reads a file or returns default.
func LoadControllerConfig(path string) (ControllerConfig, error) {
	cfg := DefaultControllerConfig()
	if path == "" {
		return cfg, nil
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return cfg, err
	}
	data = bytes.TrimPrefix(data, []byte("\xef\xbb\xbf"))
	if strings.HasSuffix(path, ".json") {
		err = json.Unmarshal(data, &cfg)
	} else {
		err = yaml.Unmarshal(data, &cfg)
	}
	return cfg, err
}

// LoadPersonaFile reads a single persona YAML or JSON file.
func LoadPersonaFile(path string) (models.PersonaProfile, error) {
	var profile models.PersonaProfile
	data, err := os.ReadFile(path)
	if err != nil {
		return profile, err
	}
	data = bytes.TrimPrefix(data, []byte("\xef\xbb\xbf"))
	if strings.HasSuffix(path, ".json") {
		err = json.Unmarshal(data, &profile)
	} else {
		err = yaml.Unmarshal(data, &profile)
	}
	return profile, err
}

// LoadPersonasFromDir loads all .yaml or .json files in the profiles directory.
func LoadPersonasFromDir(dir string) map[string]models.PersonaProfile {
	personas := BuiltinPersonas()
	if dir == "" {
		return personas
	}
	entries, err := os.ReadDir(dir)
	if err != nil {
		return personas
	}
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		ext := strings.ToLower(filepath.Ext(entry.Name()))
		if ext == ".yaml" || ext == ".yml" || ext == ".json" {
			p, err := LoadPersonaFile(filepath.Join(dir, entry.Name()))
			if err == nil && p.Name != "" {
				personas[p.Name] = p
			}
		}
	}
	return personas
}
