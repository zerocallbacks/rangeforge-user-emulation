package branding

import (
	"strings"
	"testing"
)

func TestGetHeader(t *testing.T) {
	header := GetHeader()
	if !strings.Contains(header, "RangeForge") {
		t.Errorf("Expected header to contain RangeForge, got: %s", header)
	}
	if !strings.Contains(header, Version) {
		t.Errorf("Expected header to contain Version %s, got: %s", Version, header)
	}
	if !strings.Contains(header, ProjectURL) {
		t.Errorf("Expected header to contain ProjectURL, got: %s", header)
	}
}

func TestPrintBanner(t *testing.T) {
	// Ensure PrintBanner doesn't panic with various roles
	PrintBanner("")
	PrintBanner("Manager")
	PrintBanner("Controller")
	PrintBanner("Agent")
}
