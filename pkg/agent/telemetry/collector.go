package telemetry

import (
	"net"
	"strings"

	"rangeforge-ue/pkg/models"
)

// Collector provides an interface for host metrics collection.
type Collector interface {
	Collect(agentID string, hostname string, metrics models.EmulationMetrics) (models.HostTelemetry, error)
	GetPlatformInfo() (osName, platform, arch, primaryIP string, ips []string, macs []string)
}

// GetNetworkInterfaces gathers active network interfaces, IP addresses, and MAC addresses.
func GetNetworkInterfaces() (primaryIP string, allIPs []string, allMACs []string, ifaceStats []models.NetworkInterfaceStats) {
	ifaces, err := net.Interfaces()
	if err != nil {
		return "127.0.0.1", []string{"127.0.0.1"}, nil, nil
	}

	for _, iface := range ifaces {
		// Skip loopback interfaces that have no traffic
		isLoopback := (iface.Flags & net.FlagLoopback) != 0
		isUp := (iface.Flags & net.FlagUp) != 0

		addrs, err := iface.Addrs()
		if err != nil {
			continue
		}

		stat := models.NetworkInterfaceStats{
			Name: iface.Name,
			MAC:  iface.HardwareAddr.String(),
			IsUp: isUp,
		}

		if stat.MAC != "" {
			allMACs = append(allMACs, stat.MAC)
		}

		for _, addr := range addrs {
			ipNet, ok := addr.(*net.IPNet)
			if !ok {
				continue
			}
			ip := ipNet.IP
			if ip == nil {
				continue
			}

			if ip.To4() != nil {
				ipStr := ip.String()
				stat.IPv4 = append(stat.IPv4, ipStr)
				allIPs = append(allIPs, ipStr)

				// Determine primary IP (first non-loopback, non-link-local IPv4)
				if primaryIP == "" && !isLoopback && !ip.IsLoopback() && !ip.IsLinkLocalUnicast() {
					primaryIP = ipStr
				}
			} else {
				ipStr := ip.String()
				if !strings.HasPrefix(ipStr, "fe80:") {
					stat.IPv6 = append(stat.IPv6, ipStr)
				}
			}
		}

		if len(stat.IPv4) > 0 || len(stat.IPv6) > 0 {
			ifaceStats = append(ifaceStats, stat)
		}
	}

	if primaryIP == "" && len(allIPs) > 0 {
		primaryIP = allIPs[0]
	}
	if primaryIP == "" {
		primaryIP = "127.0.0.1"
	}

	return primaryIP, allIPs, allMACs, ifaceStats
}
