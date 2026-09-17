package models

import "time"

// TaskType defines the category of emulation task.
type TaskType string

const (
	TaskTypeWebBrowse     TaskType = "web_browse"
	TaskTypeFileShare     TaskType = "file_share"
	TaskTypeHostActivity  TaskType = "host_activity"
	TaskTypeCustomCommand TaskType = "custom_command"
	TaskTypeDNSQuery      TaskType = "dns_query"
)

// AgentStatus defines the operational health state of an agent.
type AgentStatus string

const (
	AgentStatusOnline  AgentStatus = "online"
	AgentStatusOffline AgentStatus = "offline"
	AgentStatusBusy    AgentStatus = "busy"
	AgentStatusPaused  AgentStatus = "paused"
)

// AgentInfo holds persistent identification and registration info.
type AgentInfo struct {
	ID              string      `json:"id"`
	Hostname        string      `json:"hostname"`
	OS              string      `json:"os"`              // windows, linux, freebsd, darwin
	Platform        string      `json:"platform"`        // e.g. Windows 11 Enterprise, Ubuntu 24.04
	Arch            string      `json:"arch"`            // amd64, arm64
	PrimaryIP       string      `json:"primary_ip"`
	IPAddresses     []string    `json:"ip_addresses"`
	MACAddresses    []string    `json:"mac_addresses"`
	AssignedPersona string      `json:"assigned_persona"`
	Status          AgentStatus `json:"status"`
	FirstSeen       time.Time   `json:"first_seen"`
	LastHeartbeat   time.Time   `json:"last_heartbeat"`
	Version         string      `json:"version"`
	Tags            []string    `json:"tags"`
	AvailableTools  []string    `json:"available_tools,omitempty"`
	NetworkTools    []string    `json:"network_tools,omitempty"`
	HostTools       []string    `json:"host_tools,omitempty"`
}

// HostTelemetry holds periodic system telemetry data sent by agents.
type HostTelemetry struct {
	AgentID           string                  `json:"agent_id"`
	Hostname          string                  `json:"hostname"`
	Timestamp         time.Time               `json:"timestamp"`
	UptimeSeconds     uint64                  `json:"uptime_seconds"`
	CPU               CPUStats                `json:"cpu"`
	Memory            MemoryStats             `json:"memory"`
	Disk              DiskStats               `json:"disk"`
	NetworkInterfaces []NetworkInterfaceStats `json:"network_interfaces"`
	ProcessCount      int                     `json:"process_count"`
	AvailableTools    []string                `json:"available_tools,omitempty"`
	EmulationStats    EmulationMetrics        `json:"emulation_stats"`
}

// CPUStats captures CPU usage metrics.
type CPUStats struct {
	Percent  float64 `json:"percent"`
	NumCores int     `json:"num_cores"`
	Model    string  `json:"model,omitempty"`
}

// MemoryStats captures RAM statistics in bytes.
type MemoryStats struct {
	TotalBytes uint64  `json:"total_bytes"`
	UsedBytes  uint64  `json:"used_bytes"`
	FreeBytes  uint64  `json:"free_bytes"`
	Percent    float64 `json:"percent"`
}

// DiskStats captures disk space metrics for the primary partition.
type DiskStats struct {
	MountPoint string  `json:"mount_point"`
	TotalBytes uint64  `json:"total_bytes"`
	UsedBytes  uint64  `json:"used_bytes"`
	FreeBytes  uint64  `json:"free_bytes"`
	Percent    float64 `json:"percent"`
}

// NetworkInterfaceStats details an interface's addressing and traffic.
type NetworkInterfaceStats struct {
	Name      string   `json:"name"`
	IPv4      []string `json:"ipv4"`
	IPv6      []string `json:"ipv6"`
	MAC       string   `json:"mac"`
	IsUp      bool     `json:"is_up"`
	BytesSent uint64   `json:"bytes_sent,omitempty"`
	BytesRecv uint64   `json:"bytes_recv,omitempty"`
}

// EmulationMetrics tracks activity generation volume on the host.
type EmulationMetrics struct {
	WebRequestsTotal uint64 `json:"web_requests_total"`
	WebRequestsErr   uint64 `json:"web_requests_err"`
	ShareOpsTotal    uint64 `json:"share_ops_total"`
	ShareOpsErr      uint64 `json:"share_ops_err"`
	FilesCreated     uint64 `json:"files_created,omitempty"`
	FilesDeleted     uint64 `json:"files_deleted,omitempty"`
	ActiveFiles      uint64 `json:"active_files,omitempty"`
	PingsTotal       uint64 `json:"pings_total"`
	PingsErr         uint64 `json:"pings_err"`
	HostActionsTotal uint64 `json:"host_actions_total"`
	CommandsExecuted uint64 `json:"commands_executed"`
	ActivePersona    string `json:"active_persona"`
}

// EmulationEvent records a single simulated user action for real-time telemetry streaming.
type EmulationEvent struct {
	ID         string    `json:"id"`
	Timestamp  time.Time `json:"timestamp"`
	AgentID    string    `json:"agent_id"`
	Protocol   string    `json:"protocol"` // HTTP, HTTPS, SMB, ICMP, CMD
	Target     string    `json:"target"`
	Status     string    `json:"status"` // SUCCESS, FAILED, TIMEOUT
	DurationMs int64     `json:"duration_ms"`
	Details    string    `json:"details,omitempty"`
}

// EmulationTask represents an instruction dispatched from Manager to an Agent.
type EmulationTask struct {
	TaskID         string                 `json:"task_id"`
	Type           TaskType               `json:"type"`
	Parameters     map[string]interface{} `json:"parameters"`
	CreatedAt      time.Time              `json:"created_at"`
	TimeoutSeconds int                    `json:"timeout_seconds"`
}

// TaskResult captures execution status, output, and telemetry from a completed task.
type TaskResult struct {
	TaskID      string    `json:"task_id"`
	AgentID     string    `json:"agent_id"`
	Type        TaskType  `json:"type"`
	Status      string    `json:"status"` // success, failed, timeout
	ExitCode    int       `json:"exit_code"`
	Stdout      string    `json:"stdout,omitempty"`
	Stderr      string    `json:"stderr,omitempty"`
	DurationMs  int64     `json:"duration_ms"`
	CompletedAt time.Time `json:"completed_at"`
	Error       string    `json:"error,omitempty"`
}

// PersonaProfile defines behavioral parameters for a realistic user simulation persona.
type PersonaProfile struct {
	Name         string             `json:"name" yaml:"name"`
	Description  string             `json:"description" yaml:"description"`
	WebBrowsing  WebBrowsingConfig  `json:"web_browsing" yaml:"web_browsing"`
	FileShare    FileShareConfig    `json:"file_share" yaml:"file_share"`
	Ping         PingConfig         `json:"ping" yaml:"ping"`
	HostActivity HostActivityConfig `json:"host_activity" yaml:"host_activity"`
}

// PingConfig specifies ICMP keepalive network simulation.
type PingConfig struct {
	Enabled     bool     `json:"enabled" yaml:"enabled"`
	IntervalSec int      `json:"interval_sec" yaml:"interval_sec"`
	Targets     []string `json:"targets" yaml:"targets"`
}

// WebBrowsingConfig specifies HTTP/HTTPS emulation behavior.
type WebBrowsingConfig struct {
	Enabled               bool     `json:"enabled" yaml:"enabled"`
	RequestsPerMinuteMin  int      `json:"requests_per_minute_min" yaml:"requests_per_minute_min"`
	RequestsPerMinuteMax  int      `json:"requests_per_minute_max" yaml:"requests_per_minute_max"`
	DwellTimeMinSec       int      `json:"dwell_time_min_sec" yaml:"dwell_time_min_sec"`
	DwellTimeMaxSec       int      `json:"dwell_time_max_sec" yaml:"dwell_time_max_sec"`
	FetchAssets           bool     `json:"fetch_assets" yaml:"fetch_assets"` // css, js, images
	UserAgents            []string `json:"user_agents" yaml:"user_agents"`
	TargetURLs            []string `json:"target_urls" yaml:"target_urls"`
	SearchKeywords        []string `json:"search_keywords" yaml:"search_keywords"`
}

// FileShareConfig specifies SMB/share interaction simulation.
type FileShareConfig struct {
	Enabled     bool          `json:"enabled" yaml:"enabled"`
	IntervalSec int           `json:"interval_sec" yaml:"interval_sec"`
	Shares      []ShareTarget `json:"shares" yaml:"shares"`
	ReadRatio   float64       `json:"read_ratio" yaml:"read_ratio"`   // e.g. 0.80 read
	WriteRatio  float64       `json:"write_ratio" yaml:"write_ratio"` // e.g. 0.20 write
}

// ShareTarget specifies share path and credentials.
type ShareTarget struct {
	Path     string `json:"path" yaml:"path"` // e.g. \\fileserver\finance or /mnt/shares/finance
	Username string `json:"username" yaml:"username"`
	Password string `json:"password" yaml:"password"`
	Domain   string `json:"domain" yaml:"domain"`
}

// HostActivityConfig specifies local user actions (documents, safe processes, noise control).
type HostActivityConfig struct {
	Enabled            bool     `json:"enabled" yaml:"enabled"`
	IntervalSec        int      `json:"interval_sec" yaml:"interval_sec"`
	SafeProcesses      []string `json:"safe_processes" yaml:"safe_processes"` // benign tools like ping, whoami, calc
	SimulateOfficeDocs bool     `json:"simulate_office_docs" yaml:"simulate_office_docs"`
	TempFileOperations bool     `json:"temp_file_operations" yaml:"temp_file_operations"`
	UserFileOperations bool     `json:"user_file_operations,omitempty" yaml:"user_file_operations,omitempty"` // Strictly Documents, Downloads, Desktop
	NoiseLevel         string   `json:"noise_level,omitempty" yaml:"noise_level,omitempty"`                   // "stealth", "balanced", "active"
	ToolIntervalSec    int      `json:"tool_interval_sec,omitempty" yaml:"tool_interval_sec,omitempty"`       // Cadence for binary execution
	PauseToolExecution bool     `json:"pause_tool_execution,omitempty" yaml:"pause_tool_execution,omitempty"` // Mute process creation for EDR/Blue team
}

// CommandRequest represents a direct operator request to run a command on a host.
type CommandRequest struct {
	TargetAgentID  string `json:"target_agent_id"`
	Command        string `json:"command"`
	Shell          string `json:"shell"` // powershell, cmd, bash, sh, auto
	TimeoutSeconds int    `json:"timeout_seconds"`
}

// HeartbeatRequest is sent by an agent every few seconds.
type HeartbeatRequest struct {
	AgentID   string        `json:"agent_id"`
	Telemetry HostTelemetry `json:"telemetry"`
}

// WebCorpusPayload encapsulates distributed targets and search queries.
type WebCorpusPayload struct {
	IntranetPortals []string `json:"intranet_portals"`
	InternetSites   []string `json:"internet_sites"`
	SearchQueries   []string `json:"search_queries"`
	StaticAssets    []string `json:"static_assets"`
}

// HeartbeatResponse contains instructions returned by Manager.
type HeartbeatResponse struct {
	Acknowledge      bool              `json:"acknowledge"`
	OperationalState string            `json:"operational_state,omitempty"` // "running", "paused", "stopped", "emergency_stop"
	Persona          string            `json:"persona,omitempty"`
	PersonaProfile   *PersonaProfile   `json:"persona_profile,omitempty"`
	WebCorpus        *WebCorpusPayload `json:"web_corpus,omitempty"`
	Wordlist         string            `json:"wordlist,omitempty"`
	Tasks            []EmulationTask   `json:"tasks,omitempty"`
	HeartbeatSec     int               `json:"heartbeat_sec,omitempty"`
	PurgeArtifacts   bool              `json:"purge_artifacts,omitempty"`
}

// RegisterRequest is sent when an agent boots up.
type RegisterRequest struct {
	Agent AgentInfo `json:"agent"`
}

// RegisterResponse confirms registration and supplies initial configuration.
type RegisterResponse struct {
	Success         bool   `json:"success"`
	AssignedPersona string `json:"assigned_persona"`
	HeartbeatSec    int    `json:"heartbeat_sec"`
	Message         string `json:"message"`
}
