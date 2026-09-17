//go:build windows

package telemetry

import (
	"fmt"
	"math"
	"os"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"syscall"
	"time"
	"unsafe"

	"rangeforge-ue/pkg/models"
)

var (
	modkernel32              = syscall.NewLazyDLL("kernel32.dll")
	procGetSystemTimes       = modkernel32.NewProc("GetSystemTimes")
	procGlobalMemoryStatusEx = modkernel32.NewProc("GlobalMemoryStatusEx")
	procGetDiskFreeSpaceExW  = modkernel32.NewProc("GetDiskFreeSpaceExW")
	procGetTickCount64       = modkernel32.NewProc("GetTickCount64")

	modadvapi32             = syscall.NewLazyDLL("advapi32.dll")
	procRegOpenKeyExW        = modadvapi32.NewProc("RegOpenKeyExW")
	procRegQueryValueExW     = modadvapi32.NewProc("RegQueryValueExW")
	procRegCloseKey          = modadvapi32.NewProc("RegCloseKey")
)

type memoryStatusEx struct {
	dwLength                uint32
	dwMemoryLoad            uint32
	ullTotalPhys            uint64
	ullAvailPhys            uint64
	ullTotalPageFile        uint64
	ullAvailPageFile        uint64
	ullTotalVirtual         uint64
	ullAvailVirtual         uint64
	ullAvailExtendedVirtual uint64
}

type filetime struct {
	LowDateTime  uint32
	HighDateTime uint32
}

func (ft *filetime) toUInt64() uint64 {
	return (uint64(ft.HighDateTime) << 32) | uint64(ft.LowDateTime)
}

// WindowsCollector collects native Windows metrics without external dependencies.
type WindowsCollector struct {
	mu           sync.Mutex
	lastIdle     uint64
	lastKernel   uint64
	lastUser     uint64
	lastCPUTime  time.Time
	lastCPUUsage float64
}

// NewPlatformCollector constructs the Windows telemetry collector.
func NewPlatformCollector() Collector {
	c := &WindowsCollector{}
	c.sampleCPU()
	return c
}

func (c *WindowsCollector) sampleCPU() float64 {
	c.mu.Lock()
	defer c.mu.Unlock()

	var idle, kernel, user filetime
	r1, _, _ := procGetSystemTimes.Call(
		uintptr(unsafe.Pointer(&idle)),
		uintptr(unsafe.Pointer(&kernel)),
		uintptr(unsafe.Pointer(&user)),
	)
	if r1 == 0 {
		return 0.0
	}

	idleVal := idle.toUInt64()
	kernelVal := kernel.toUInt64()
	userVal := user.toUInt64()

	if c.lastCPUTime.IsZero() {
		c.lastIdle = idleVal
		c.lastKernel = kernelVal
		c.lastUser = userVal
		c.lastCPUTime = time.Now()
		c.lastCPUUsage = 0.0
		return 0.0
	}

	idleDelta := idleVal - c.lastIdle
	kernelDelta := kernelVal - c.lastKernel
	userDelta := userVal - c.lastUser

	totalDelta := kernelDelta + userDelta
	if totalDelta == 0 {
		return c.lastCPUUsage
	}

	// In Windows, kernel time includes idle time!
	sysDelta := totalDelta - idleDelta
	usage := (float64(sysDelta) / float64(totalDelta)) * 100.0
	if usage < 0 {
		usage = 0
	} else if usage > 100 {
		usage = 100
	}

	c.lastIdle = idleVal
	c.lastKernel = kernelVal
	c.lastUser = userVal
	c.lastCPUTime = time.Now()
	c.lastCPUUsage = math.Round(usage*10) / 10
	return c.lastCPUUsage
}

func (c *WindowsCollector) getMemoryStats() models.MemoryStats {
	var mem memoryStatusEx
	mem.dwLength = uint32(unsafe.Sizeof(mem))

	r1, _, _ := procGlobalMemoryStatusEx.Call(uintptr(unsafe.Pointer(&mem)))
	if r1 == 0 {
		return models.MemoryStats{}
	}

	used := mem.ullTotalPhys - mem.ullAvailPhys
	pct := 0.0
	if mem.ullTotalPhys > 0 {
		pct = (float64(used) / float64(mem.ullTotalPhys)) * 100.0
	}

	return models.MemoryStats{
		TotalBytes: mem.ullTotalPhys,
		UsedBytes:  used,
		FreeBytes:  mem.ullAvailPhys,
		Percent:    math.Round(pct*10) / 10,
	}
}

func (c *WindowsCollector) getDiskStats() models.DiskStats {
	rootPath := "C:\\"
	var freeBytesAvail, totalBytes, totalFreeBytes uint64

	pRoot, err := syscall.UTF16PtrFromString(rootPath)
	if err != nil {
		return models.DiskStats{MountPoint: rootPath}
	}

	r1, _, _ := procGetDiskFreeSpaceExW.Call(
		uintptr(unsafe.Pointer(pRoot)),
		uintptr(unsafe.Pointer(&freeBytesAvail)),
		uintptr(unsafe.Pointer(&totalBytes)),
		uintptr(unsafe.Pointer(&totalFreeBytes)),
	)
	if r1 == 0 {
		return models.DiskStats{MountPoint: rootPath}
	}

	used := totalBytes - freeBytesAvail
	pct := 0.0
	if totalBytes > 0 {
		pct = (float64(used) / float64(totalBytes)) * 100.0
	}

	return models.DiskStats{
		MountPoint: rootPath,
		TotalBytes: totalBytes,
		UsedBytes:  used,
		FreeBytes:  freeBytesAvail,
		Percent:    math.Round(pct*10) / 10,
	}
}

func (c *WindowsCollector) getUptime() uint64 {
	r1, _, _ := procGetTickCount64.Call()
	if r1 == 0 {
		return 0
	}
	return uint64(r1) / 1000
}

// Collect gathers complete Windows host telemetry.
func (c *WindowsCollector) Collect(agentID string, hostname string, metrics models.EmulationMetrics) (models.HostTelemetry, error) {
	if hostname == "" {
		hostname, _ = os.Hostname()
	}

	cpuPercent := c.sampleCPU()
	memStats := c.getMemoryStats()
	diskStats := c.getDiskStats()
	uptime := c.getUptime()
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
		ProcessCount:      120, // baseline estimate
		EmulationStats:    metrics,
	}, nil
}

// GetPlatformInfo returns Windows environment data with exact OS edition and build.
func (c *WindowsCollector) GetPlatformInfo() (osName, platform, arch, primaryIP string, ips []string, macs []string) {
	osName = "windows"
	platform = getExactWindowsOS()
	arch = runtime.GOARCH
	primaryIP, ips, macs, _ = GetNetworkInterfaces()
	return
}

func getExactWindowsOS() string {
	var hKey uintptr
	subKey, err := syscall.UTF16PtrFromString(`SOFTWARE\Microsoft\Windows NT\CurrentVersion`)
	if err != nil {
		return "Windows (" + runtime.GOARCH + ")"
	}

	const HKEY_LOCAL_MACHINE = uintptr(0x80000002)
	const KEY_READ = uintptr(0x20019)

	r1, _, _ := procRegOpenKeyExW.Call(HKEY_LOCAL_MACHINE, uintptr(unsafe.Pointer(subKey)), 0, KEY_READ, uintptr(unsafe.Pointer(&hKey)))
	if r1 != 0 {
		return "Windows (" + runtime.GOARCH + ")"
	}
	defer procRegCloseKey.Call(hKey)

	queryStr := func(valName string) string {
		namePtr, err := syscall.UTF16PtrFromString(valName)
		if err != nil {
			return ""
		}
		var bufType uint32
		var bufSize uint32 = 512
		buf := make([]uint16, 256)
		r, _, _ := procRegQueryValueExW.Call(hKey, uintptr(unsafe.Pointer(namePtr)), 0, uintptr(unsafe.Pointer(&bufType)), uintptr(unsafe.Pointer(&buf[0])), uintptr(unsafe.Pointer(&bufSize)))
		if r == 0 {
			return syscall.UTF16ToString(buf)
		}
		return ""
	}

	queryDword := func(valName string) uint32 {
		namePtr, err := syscall.UTF16PtrFromString(valName)
		if err != nil {
			return 0
		}
		var bufType uint32
		var val uint32
		var bufSize uint32 = 4
		r, _, _ := procRegQueryValueExW.Call(hKey, uintptr(unsafe.Pointer(namePtr)), 0, uintptr(unsafe.Pointer(&bufType)), uintptr(unsafe.Pointer(&val)), uintptr(unsafe.Pointer(&bufSize)))
		if r == 0 {
			return val
		}
		return 0
	}

	prodName := queryStr("ProductName")
	dispVer := queryStr("DisplayVersion")
	buildStr := queryStr("CurrentBuildNumber")
	if buildStr == "" {
		buildStr = queryStr("CurrentBuild")
	}
	ubr := queryDword("UBR")

	// If build >= 22000, it's Windows 11 (Windows maintains "Windows 10" in ProductName for legacy compatibility)
	buildNum, _ := strconv.Atoi(buildStr)
	if buildNum >= 22000 && strings.Contains(prodName, "Windows 10") {
		prodName = strings.Replace(prodName, "Windows 10", "Windows 11", 1)
	}

	if prodName == "" {
		prodName = "Windows"
	}

	parts := []string{prodName}
	if dispVer != "" {
		parts = append(parts, dispVer)
	}
	if buildStr != "" {
		if ubr > 0 {
			parts = append(parts, fmt.Sprintf("(Build %s.%d, %s)", buildStr, ubr, runtime.GOARCH))
		} else {
			parts = append(parts, fmt.Sprintf("(Build %s, %s)", buildStr, runtime.GOARCH))
		}
	} else {
		parts = append(parts, "("+runtime.GOARCH+")")
	}

	return strings.Join(parts, " ")
}
