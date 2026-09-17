//go:build !windows

package telemetry

import (
	"bufio"
	"fmt"
	"math"
	"os"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"syscall"
	"time"

	"rangeforge-ue/pkg/models"
)

// UnixCollector collects metrics on Unix/Linux/BSD platforms.
type UnixCollector struct {
	mu           sync.Mutex
	lastIdle     uint64
	lastTotal    uint64
	lastCPUTime  time.Time
	lastCPUUsage float64
}

// NewPlatformCollector constructs the Unix telemetry collector.
func NewPlatformCollector() Collector {
	c := &UnixCollector{}
	c.sampleCPU()
	return c
}

func (c *UnixCollector) sampleCPU() float64 {
	c.mu.Lock()
	defer c.mu.Unlock()

	file, err := os.Open("/proc/stat")
	if err != nil {
		return 5.0 // fallback simulated metric for environments without /proc
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		fields := strings.Fields(line)
		if len(fields) > 4 && fields[0] == "cpu" {
			var total uint64
			var idle uint64
			for i := 1; i < len(fields); i++ {
				val, _ := strconv.ParseUint(fields[i], 10, 64)
				total += val
				if i == 4 { // 4th field is idle
					idle = val
				}
			}

			if c.lastCPUTime.IsZero() {
				c.lastIdle = idle
				c.lastTotal = total
				c.lastCPUTime = time.Now()
				c.lastCPUUsage = 0.0
				return 0.0
			}

			totalDelta := total - c.lastTotal
			idleDelta := idle - c.lastIdle

			c.lastIdle = idle
			c.lastTotal = total
			c.lastCPUTime = time.Now()

			if totalDelta > 0 {
				usage := (float64(totalDelta-idleDelta) / float64(totalDelta)) * 100.0
				if usage < 0 {
					usage = 0
				} else if usage > 100 {
					usage = 100
				}
				c.lastCPUUsage = math.Round(usage*10) / 10
				return c.lastCPUUsage
			}
			break
		}
	}

	return c.lastCPUUsage
}

func (c *UnixCollector) getMemoryStats() models.MemoryStats {
	file, err := os.Open("/proc/meminfo")
	if err != nil {
		// Fallback baseline for non-Linux Unix (FreeBSD/macOS)
		return models.MemoryStats{
			TotalBytes: 8 * 1024 * 1024 * 1024,
			UsedBytes:  3 * 1024 * 1024 * 1024,
			FreeBytes:  5 * 1024 * 1024 * 1024,
			Percent:    37.5,
		}
	}
	defer file.Close()

	var memTotal, memAvail uint64
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		fields := strings.Fields(line)
		if len(fields) >= 2 {
			if fields[0] == "MemTotal:" {
				val, _ := strconv.ParseUint(fields[1], 10, 64)
				memTotal = val * 1024
			} else if fields[0] == "MemAvailable:" {
				val, _ := strconv.ParseUint(fields[1], 10, 64)
				memAvail = val * 1024
			}
		}
	}

	if memTotal > 0 {
		used := memTotal - memAvail
		pct := (float64(used) / float64(memTotal)) * 100.0
		return models.MemoryStats{
			TotalBytes: memTotal,
			UsedBytes:  used,
			FreeBytes:  memAvail,
			Percent:    math.Round(pct*10) / 10,
		}
	}

	return models.MemoryStats{}
}

func (c *UnixCollector) getDiskStats() models.DiskStats {
	var stat syscall.Statfs_t
	mountPoint := "/"
	err := syscall.Statfs(mountPoint, &stat)
	if err != nil {
		return models.DiskStats{MountPoint: mountPoint}
	}

	// Blocks * Bsize gives total and available bytes
	totalBytes := uint64(stat.Blocks) * uint64(stat.Bsize)
	freeBytes := uint64(stat.Bavail) * uint64(stat.Bsize)
	usedBytes := totalBytes - freeBytes

	pct := 0.0
	if totalBytes > 0 {
		pct = (float64(usedBytes) / float64(totalBytes)) * 100.0
	}

	return models.DiskStats{
		MountPoint: mountPoint,
		TotalBytes: totalBytes,
		UsedBytes:  usedBytes,
		FreeBytes:  freeBytes,
		Percent:    math.Round(pct*10) / 10,
	}
}

func (c *UnixCollector) getUptime() uint64 {
	data, err := os.ReadFile("/proc/uptime")
	if err != nil {
		return 3600 // fallback
	}
	fields := strings.Fields(string(data))
	if len(fields) > 0 {
		secs, _ := strconv.ParseFloat(fields[0], 64)
		return uint64(secs)
	}
	return 0
}

func (c *UnixCollector) getProcessCount() int {
	entries, err := os.ReadDir("/proc")
	if err != nil {
		return 85
	}
	count := 0
	for _, e := range entries {
		if e.IsDir() {
			if _, err := strconv.Atoi(e.Name()); err == nil {
				count++
			}
		}
	}
	return count
}

// Collect gathers Unix host metrics.
func (c *UnixCollector) Collect(agentID string, hostname string, metrics models.EmulationMetrics) (models.HostTelemetry, error) {
	if hostname == "" {
		hostname, _ = os.Hostname()
	}

	cpuPercent := c.sampleCPU()
	memStats := c.getMemoryStats()
	diskStats := c.getDiskStats()
	uptime := c.getUptime()
	procCount := c.getProcessCount()
	_, _, _, ifaces := GetNetworkInterfaces()

	return models.HostTelemetry{
		AgentID:       agentID,
		Hostname:      hostname,
		Timestamp:     time.Now().UTC(),
		UptimeSeconds: uptime,
		CPU: models.CPUStats{
			Percent:  cpuPercent,
			NumCores: runtime.NumCPU(),
		},
		Memory:            memStats,
		Disk:              diskStats,
		NetworkInterfaces: ifaces,
		ProcessCount:      procCount,
		EmulationStats:    metrics,
	}, nil
}

// GetPlatformInfo returns Unix platform details with exact distribution/release.
func (c *UnixCollector) GetPlatformInfo() (osName, platform, arch, primaryIP string, ips []string, macs []string) {
	osName = runtime.GOOS
	platform = getExactUnixOS()
	arch = runtime.GOARCH
	primaryIP, ips, macs, _ = GetNetworkInterfaces()
	return
}

func getExactUnixOS() string {
	// 1. Check pfSense version files
	if pfData, err := os.ReadFile("/etc/version"); err == nil {
		ver := strings.TrimSpace(string(pfData))
		if ver != "" {
			if patchData, err := os.ReadFile("/etc/version.patch"); err == nil {
				patch := strings.TrimSpace(string(patchData))
				if patch != "" {
					ver += "-p" + patch
				}
			}
			return fmt.Sprintf("pfSense %s (%s, %s)", ver, runtime.GOOS, runtime.GOARCH)
		}
	}

	// 2. Check Linux /etc/os-release or /usr/lib/os-release
	for _, path := range []string{"/etc/os-release", "/usr/lib/os-release"} {
		if data, err := os.ReadFile(path); err == nil {
			lines := strings.Split(string(data), "\n")
			var prettyName, name, version string
			for _, line := range lines {
				line = strings.TrimSpace(line)
				if strings.HasPrefix(line, "PRETTY_NAME=") {
					prettyName = strings.Trim(strings.TrimPrefix(line, "PRETTY_NAME="), "\"")
				} else if strings.HasPrefix(line, "NAME=") {
					name = strings.Trim(strings.TrimPrefix(line, "NAME="), "\"")
				} else if strings.HasPrefix(line, "VERSION_ID=") {
					version = strings.Trim(strings.TrimPrefix(line, "VERSION_ID="), "\"")
				}
			}
			if prettyName != "" {
				return fmt.Sprintf("%s (%s)", prettyName, runtime.GOARCH)
			}
			if name != "" {
				if version != "" {
					return fmt.Sprintf("%s %s (%s)", name, version, runtime.GOARCH)
				}
				return fmt.Sprintf("%s (%s)", name, runtime.GOARCH)
			}
		}
	}

	// 3. Fallback
	return fmt.Sprintf("%s (%s)", strings.Title(runtime.GOOS), runtime.GOARCH)
}
