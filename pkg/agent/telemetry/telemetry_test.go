package telemetry

import (
	"testing"

	"rangeforge-ue/pkg/models"
)

func TestGetNetworkInterfaces(t *testing.T) {
	primaryIP, allIPs, allMACs, ifaces := GetNetworkInterfaces()
	if primaryIP == "" {
		t.Errorf("Expected primary IP to not be empty")
	}
	if len(allIPs) == 0 {
		t.Errorf("Expected at least one detected IP address")
	}
	_ = allMACs
	_ = ifaces
}

func TestCollector(t *testing.T) {
	collector := NewPlatformCollector()
	if collector == nil {
		t.Fatalf("Failed to initialize platform collector")
	}

	osName, platform, arch, primaryIP, ips, macs := collector.GetPlatformInfo()
	if osName == "" || platform == "" || arch == "" {
		t.Errorf("Incomplete platform info: os=%s, platform=%s, arch=%s", osName, platform, arch)
	}
	if primaryIP == "" || len(ips) == 0 {
		t.Errorf("Missing IP info: primary=%s, ips=%v", primaryIP, ips)
	}
	_ = macs

	metrics := models.EmulationMetrics{
		WebRequestsTotal: 10,
		ShareOpsTotal:    5,
		HostActionsTotal: 8,
		CommandsExecuted: 2,
		ActivePersona:    "office_worker",
	}

	tel, err := collector.Collect("test-agent-01", "testhost", metrics)
	if err != nil {
		t.Fatalf("Collect failed: %v", err)
	}

	if tel.AgentID != "test-agent-01" {
		t.Errorf("Expected agent ID test-agent-01, got %s", tel.AgentID)
	}
	if tel.Hostname != "testhost" {
		t.Errorf("Expected hostname testhost, got %s", tel.Hostname)
	}
	if tel.CPU.NumCores <= 0 {
		t.Errorf("Expected positive CPU core count, got %d", tel.CPU.NumCores)
	}
	if tel.Memory.TotalBytes <= 0 {
		t.Errorf("Expected positive total memory, got %d", tel.Memory.TotalBytes)
	}
	if tel.EmulationStats.WebRequestsTotal != 10 {
		t.Errorf("Expected 10 web requests, got %d", tel.EmulationStats.WebRequestsTotal)
	}
}
