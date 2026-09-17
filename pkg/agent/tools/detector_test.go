package tools

import (
	"runtime"
	"testing"
)

func TestDetectInstalledTools(t *testing.T) {
	profile := DetectInstalledTools()

	if len(profile.AvailableTools) == 0 {
		t.Errorf("Expected to detect at least one installed tool in environment")
	}

	if runtime.GOOS == "windows" {
		if !profile.CanCMD && !profile.CanPowerShell {
			t.Errorf("Expected either cmd or powershell on Windows")
		}
		if profile.PrimaryShell == "" {
			t.Errorf("Expected primary shell on Windows")
		}
	}

	if !profile.CanWeb {
		t.Errorf("Expected CanWeb to be true")
	}
}
