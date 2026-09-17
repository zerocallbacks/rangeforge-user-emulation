package tools

import (
	"os/exec"
	"runtime"
)

// ToolProfile contains the inventory of available binaries on the host system.
type ToolProfile struct {
	AvailableTools  []string `json:"available_tools"`
	NetworkTools    []string `json:"network_tools"`
	HostTools       []string `json:"host_tools"`
	CanPing         bool     `json:"can_ping"`
	CanWeb          bool     `json:"can_web"`
	CanSMB          bool     `json:"can_smb"`
	CanPowerShell   bool     `json:"can_powershell"`
	CanBash         bool     `json:"can_bash"`
	CanCMD          bool     `json:"can_cmd"`
	CanGit          bool     `json:"can_git"`
	CanPython       bool     `json:"can_python"`
	CanDiagnostics  bool     `json:"can_diagnostics"`
	CanNetworkTools bool     `json:"can_network_tools"`
	CanRemoteSync   bool     `json:"can_remote_sync"`
	PrimaryShell    string   `json:"primary_shell"`
}

// DetectInstalledTools inspects the local host environment and identifies functional utilities.
func DetectInstalledTools() ToolProfile {
	candidates := []string{
		"powershell",
		"cmd",
		"bash",
		"sh",
		"zsh",
		"ping",
		"curl",
		"wget",
		"netstat",
		"tracert",
		"traceroute",
		"nslookup",
		"dig",
		"net",
		"smbclient",
		"git",
		"python",
		"python3",
		"wmic",
		"systeminfo",
		"whoami",
		"ssh",
		"scp",
		"sftp",
		"rsync",
		"robocopy",
		"ipconfig",
		"ifconfig",
		"ip",
		"arp",
		"route",
		"hostname",
		"tasklist",
		"ps",
		"netsh",
		"findstr",
		"grep",
		"tar",
		"zip",
		"unzip",
		"node",
		"npm",
		"docker",
		"sc",
	}

	var found []string
	var netTools []string
	var hostTools []string
	toolMap := make(map[string]bool)

	isNetTool := map[string]bool{
		"ping": true, "curl": true, "wget": true, "netstat": true, "tracert": true,
		"traceroute": true, "nslookup": true, "dig": true, "net": true, "smbclient": true,
		"ipconfig": true, "ifconfig": true, "ip": true, "arp": true, "route": true,
		"ssh": true, "scp": true, "sftp": true, "rsync": true,
	}

	for _, name := range candidates {
		if path, err := exec.LookPath(name); err == nil && path != "" {
			found = append(found, name)
			toolMap[name] = true
			if isNetTool[name] {
				netTools = append(netTools, name)
			} else {
				hostTools = append(hostTools, name)
			}
		}
	}

	profile := ToolProfile{
		AvailableTools:  found,
		NetworkTools:    netTools,
		HostTools:       hostTools,
		CanPing:         toolMap["ping"],
		CanWeb:          true, // Go's net/http is always functional internally; curl/wget add capability
		CanSMB:          toolMap["net"] || toolMap["smbclient"] || runtime.GOOS == "windows",
		CanPowerShell:   toolMap["powershell"],
		CanBash:         toolMap["bash"],
		CanCMD:          toolMap["cmd"],
		CanGit:          toolMap["git"],
		CanPython:       toolMap["python"] || toolMap["python3"],
		CanDiagnostics:  toolMap["netstat"] || toolMap["tracert"] || toolMap["traceroute"] || toolMap["nslookup"] || toolMap["dig"],
		CanNetworkTools: toolMap["ipconfig"] || toolMap["ifconfig"] || toolMap["ip"] || toolMap["arp"] || toolMap["route"],
		CanRemoteSync:   toolMap["robocopy"] || toolMap["rsync"] || toolMap["scp"] || toolMap["sftp"],
	}

	// Determine primary shell
	if runtime.GOOS == "windows" {
		if profile.CanPowerShell {
			profile.PrimaryShell = "powershell"
		} else {
			profile.PrimaryShell = "cmd"
		}
	} else {
		if profile.CanBash {
			profile.PrimaryShell = "bash"
		} else {
			profile.PrimaryShell = "sh"
		}
	}

	return profile
}
