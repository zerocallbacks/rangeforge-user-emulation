package branding

import (
	"fmt"
	"strings"
)

const (
	ToolName   = "RangeForge User Emulation Suite (rangeforge-ue)"
	Version    = "2.0.0-oss"
	Tagline    = "Open-Source Cyber Range & Network User Activity Simulator"
	ProjectURL = "https://github.com/rangeforge/rangeforge-ue"
)

// ANSI Color Codes
const (
	Reset     = "\033[0m"
	Bold      = "\033[1m"
	Red       = "\033[31m"
	Green     = "\033[32m"
	Yellow    = "\033[33m"
	Blue      = "\033[34m"
	Magenta   = "\033[35m"
	Cyan      = "\033[36m"
	White     = "\033[37m"
	BrightRed = "\033[91m"
	BrightCyn = "\033[96m"
)

// ASCIIArt represents the stylized RangeForge open-source banner.
const ASCIIArt = `
 ██████╗  █████╗ ███╗   ██╗ ██████╗ ███████╗███████╗ ██████╗ ██████╗  ██████╗ ███████╗
 ██╔══██╗██╔══██╗████╗  ██║██╔════╝ ██╔════╝██╔════╝██╔═══██╗██╔══██╗██╔════╝ ██╔════╝
 ██████╔╝███████║██╔██╗ ██║██║  ███╗█████╗  █████╗  ██║   ██║██████╔╝██║  ███╗█████╗  
 ██╔══██╗██╔══██║██║╚██╗██║██║   ██║██╔══╝  ██╔══╝  ██║   ██║██╔══██╗██║   ██║██╔══╝  
 ██║  ██║██║  ██║██║ ╚████║╚██████╔╝███████╗██║     ╚██████╔╝██║  ██║╚██████╔╝███████╗
 ╚═╝  ╚═╝╚═╝  ╚═╝╚═╝  ╚═══╝ ╚═════╝ ╚══════╝╚═╝      ╚═════╝ ╚═╝  ╚═╝ ╚═════╝ ╚══════╝
                     U S E R   E M U L A T I O N   S U I T E
`

// PrintBanner outputs the RangeForge open-source header to stdout with ANSI formatting.
func PrintBanner(role string) {
	fmt.Println(Cyan + Bold + ASCIIArt + Reset)
	fmt.Printf("%s%s%s - %s%s%s\n", Bold, ToolName, Reset, Yellow, Version, Reset)
	fmt.Printf("%s%s%s (%s)\n", Magenta, Tagline, Reset, ProjectURL)
	fmt.Println(strings.Repeat("═", 86))
	if role != "" {
		fmt.Printf("%s[ROLE: %s]%s Starting in %s mode...\n", Green+Bold, strings.ToUpper(role), Reset, role)
		fmt.Println(strings.Repeat("─", 86))
	}
}

// GetHeader returns a single-line branded header string.
func GetHeader() string {
	return fmt.Sprintf("%s v%s | %s (%s)", ToolName, Version, Tagline, ProjectURL)
}
