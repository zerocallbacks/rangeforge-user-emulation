package controller

import (
	"fmt"
	"strings"
	"time"

	"rangeforge-ue/pkg/branding"
	"rangeforge-ue/pkg/models"
)

// CLI provides formatted terminal outputs and interactive workflows for cyber range operators.
type CLI struct {
	client *Client
}

// NewCLI constructs a CLI handler.
func NewCLI(client *Client) *CLI {
	return &CLI{client: client}
}

// ListAgents prints a formatted ASCII table of all registered agents in the cyber range.
func (c *CLI) ListAgents() error {
	agents, err := c.client.GetAgents()
	if err != nil {
		return err
	}

	branding.PrintBanner("Controller")
	fmt.Printf("%s[CYBER RANGE FLEET STATUS]%s Discovered %d Connected Host Agent(s)\n\n",
		branding.Bold+branding.Cyan, branding.Reset, len(agents))

	if len(agents) == 0 {
		fmt.Printf("%sNo agents registered yet. Ensure 'rangeforge-ue agent' is running on target hosts.%s\n",
			branding.Yellow, branding.Reset)
		return nil
	}

	fmt.Printf("%-24s %-16s %-12s %-16s %-15s %-10s %-12s\n",
		"AGENT ID", "HOSTNAME", "OS/ARCH", "PRIMARY IP", "PERSONA", "STATUS", "LAST SEEN")
	fmt.Println(strings.Repeat("─", 110))

	for _, a := range agents {
		statusColor := branding.Green
		if a.Status != models.AgentStatusOnline {
			statusColor = branding.Red
		}

		timeAgo := time.Since(a.LastHeartbeat).Round(time.Second)
		timeAgoStr := fmt.Sprintf("%vs ago", int(timeAgo.Seconds()))
		if timeAgo.Seconds() > 120 {
			timeAgoStr = a.LastHeartbeat.Format("15:04:05")
		}

		osArch := fmt.Sprintf("%s/%s", a.OS, a.Arch)

		fmt.Printf("%-24s %-16s %-12s %-16s %-15s %s%-10s%s %-12s\n",
			a.ID, a.Hostname, osArch, a.PrimaryIP, a.AssignedPersona,
			statusColor, a.Status, branding.Reset, timeAgoStr)
	}
	fmt.Println()
	return nil
}

// ShowHostStats displays detailed telemetry for a target host.
func (c *CLI) ShowHostStats(agentID string) error {
	tel, err := c.client.GetTelemetry(agentID)
	if err != nil {
		return err
	}

	branding.PrintBanner("Host Telemetry")
	fmt.Printf("%s=== TELEMETRY REPORT: %s (%s) ===%s\n",
		branding.Bold+branding.Cyan, tel.Hostname, tel.AgentID, branding.Reset)
	fmt.Printf("Timestamp: %s | Host Uptime: %d seconds\n\n",
		tel.Timestamp.Format(time.RFC1123), tel.UptimeSeconds)

	// CPU & Memory
	fmt.Printf("%s[SYSTEM HARDWARE]%s\n", branding.Yellow+branding.Bold, branding.Reset)
	fmt.Printf("  CPU Usage:       %.1f%% (%d Cores)\n", tel.CPU.Percent, tel.CPU.NumCores)
	fmt.Printf("  Memory Usage:    %.1f%% (%.2f GB used / %.2f GB total)\n",
		tel.Memory.Percent,
		float64(tel.Memory.UsedBytes)/(1024*1024*1024),
		float64(tel.Memory.TotalBytes)/(1024*1024*1024))
	fmt.Printf("  Disk Space:      %.1f%% (%.2f GB free / %.2f GB total on %s)\n",
		tel.Disk.Percent,
		float64(tel.Disk.FreeBytes)/(1024*1024*1024),
		float64(tel.Disk.TotalBytes)/(1024*1024*1024),
		tel.Disk.MountPoint)
	fmt.Printf("  Active Procs:    %d\n\n", tel.ProcessCount)

	// Network Interfaces
	fmt.Printf("%s[NETWORK INTERFACES & SHARED TOPOLOGY]%s\n", branding.Yellow+branding.Bold, branding.Reset)
	for _, iface := range tel.NetworkInterfaces {
		state := "DOWN"
		if iface.IsUp {
			state = "UP"
		}
		fmt.Printf("  • %-12s [%s] MAC: %s\n", iface.Name, state, iface.MAC)
		if len(iface.IPv4) > 0 {
			fmt.Printf("      IPv4: %s\n", strings.Join(iface.IPv4, ", "))
		}
		if len(iface.IPv6) > 0 {
			fmt.Printf("      IPv6: %s\n", strings.Join(iface.IPv6, ", "))
		}
	}
	fmt.Println()

	// Emulation Activity Metrics
	fmt.Printf("%s[USER EMULATION ACTIVITY METRICS]%s\n", branding.Yellow+branding.Bold, branding.Reset)
	fmt.Printf("  Active Persona:  %s%s%s\n", branding.Green+branding.Bold, tel.EmulationStats.ActivePersona, branding.Reset)
	fmt.Printf("  Web Requests:    %d total (%d errors)\n", tel.EmulationStats.WebRequestsTotal, tel.EmulationStats.WebRequestsErr)
	fmt.Printf("  Share Actions:   %d total (%d errors)\n", tel.EmulationStats.ShareOpsTotal, tel.EmulationStats.ShareOpsErr)
	fmt.Printf("  Host Actions:    %d local actions executed\n", tel.EmulationStats.HostActionsTotal)
	fmt.Printf("  Custom Commands: %d admin commands completed\n\n", tel.EmulationStats.CommandsExecuted)

	return nil
}

// RunCommand dispatches a command to an agent, polls for completion, and prints results.
func (c *CLI) RunCommand(agentID, command, shell string, timeoutSec int) error {
	taskID, err := c.client.SendCommand(agentID, command, shell, timeoutSec)
	if err != nil {
		return err
	}

	fmt.Printf("%s[DISPATCHED]%s Task %s sent to Agent %s: '%s'\n",
		branding.Cyan+branding.Bold, branding.Reset, taskID, agentID, command)
	fmt.Print("Waiting for execution on target host...")

	deadline := time.Now().Add(time.Duration(timeoutSec+5) * time.Second)
	for time.Now().Before(deadline) {
		time.Sleep(500 * time.Millisecond)
		fmt.Print(".")

		res, err := c.client.GetCommandStatus(taskID)
		if err != nil {
			continue
		}
		if res != nil {
			fmt.Println()
			fmt.Println(strings.Repeat("═", 72))
			fmt.Printf("%sExecution Result for Task %s%s (Host: %s, Time: %dms)\n",
				branding.Bold, res.TaskID, branding.Reset, res.AgentID, res.DurationMs)

			if res.Status == "success" {
				fmt.Printf("Status: %sSUCCESS%s (Exit Code: %d)\n", branding.Green+branding.Bold, branding.Reset, res.ExitCode)
			} else {
				fmt.Printf("Status: %s%s%s (Exit Code: %d)\n", branding.Red+branding.Bold, strings.ToUpper(res.Status), branding.Reset, res.ExitCode)
			}

			if res.Stdout != "" {
				fmt.Printf("\n%s[STDOUT]%s\n%s", branding.Cyan, branding.Reset, res.Stdout)
			}
			if res.Stderr != "" {
				fmt.Printf("\n%s[STDERR]%s\n%s", branding.Red, branding.Reset, res.Stderr)
			}
			if res.Error != "" {
				fmt.Printf("\n%s[ERROR]%s: %s\n", branding.Red+branding.Bold, branding.Reset, res.Error)
			}
			fmt.Println(strings.Repeat("═", 72))
			return nil
		}
	}

	fmt.Println("\nCommand execution timed out while waiting for agent response.")
	return nil
}

// SetPersona requests persona transition for an agent.
func (c *CLI) SetPersona(agentID, persona string) error {
	if err := c.client.SetPersona(agentID, persona); err != nil {
		return err
	}
	fmt.Printf("%s[SUCCESS]%s Persona for agent %s successfully updated to '%s'\n",
		branding.Green+branding.Bold, branding.Reset, agentID, persona)
	return nil
}

// ListPersonas displays all available persona definitions.
func (c *CLI) ListPersonas() error {
	personas, err := c.client.GetPersonas()
	if err != nil {
		return err
	}

	branding.PrintBanner("Personas")
	fmt.Printf("%sAvailable Cyber Range Emulation Personas:%s\n\n", branding.Bold+branding.Cyan, branding.Reset)

	for name, p := range personas {
		fmt.Printf("%s• %s%s: %s\n", branding.Yellow+branding.Bold, name, branding.Reset, p.Description)
		fmt.Printf("    Web:   Enabled=%v (RPM: %d-%d, Dwell: %d-%ds, Assets=%v)\n",
			p.WebBrowsing.Enabled, p.WebBrowsing.RequestsPerMinuteMin, p.WebBrowsing.RequestsPerMinuteMax,
			p.WebBrowsing.DwellTimeMinSec, p.WebBrowsing.DwellTimeMaxSec, p.WebBrowsing.FetchAssets)
		fmt.Printf("    Share: Enabled=%v (Interval: %ds, Read: %.0f%%, Write: %.0f%%)\n",
			p.FileShare.Enabled, p.FileShare.IntervalSec, p.FileShare.ReadRatio*100, p.FileShare.WriteRatio*100)
		fmt.Printf("    Host:  Enabled=%v (Interval: %ds, SafeProcs: %v)\n\n",
			p.HostActivity.Enabled, p.HostActivity.IntervalSec, p.HostActivity.SafeProcesses)
	}
	return nil
}
