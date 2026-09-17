package config

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"rangeforge-ue/pkg/models"
)

// WebCorpus holds target URLs, domains, and keywords for realistic web emulation.
type WebCorpus struct {
	IntranetPortals []string `json:"intranet_portals"`
	InternetSites   []string `json:"internet_sites"`
	SearchQueries   []string `json:"search_queries"`
	StaticAssets    []string `json:"static_assets"`
	UserAgents      []string `json:"user_agents"`
}

// UserCredential represents simulated cyber range credentials.
type UserCredential struct {
	Username string `json:"username"`
	Password string `json:"password"`
	Domain   string `json:"domain"`
	Role     string `json:"role"`
}

// ShareCorpus represents a list of target shares and operations.
type ShareCorpus struct {
	Shares      []models.ShareTarget `json:"shares"`
	SampleFiles []string             `json:"sample_files"`
}

// DefaultWebCorpus provides realistic background browsing traffic targets.
func DefaultWebCorpus() WebCorpus {
	return WebCorpus{
		IntranetPortals: []string{
			"http://portal.range.local",
			"http://wiki.corp.local/kb/index.php",
			"http://jira.corp.local/secure/Dashboard.jspa",
			"http://hr.range.local/benefits",
			"http://files.corp.local/public",
		},
		InternetSites: []string{
			"https://en.wikipedia.org/wiki/Computer_security",
			"https://en.wikipedia.org/wiki/Information_security",
			"https://en.wikipedia.org/wiki/Network_security",
			"https://www.cnn.com",
			"https://www.reuters.com",
			"https://news.ycombinator.com",
			"https://github.com/trending",
			"https://stackoverflow.com/questions",
			"https://docs.python.org/3/",
			"https://golang.org/doc/",
			"https://www.microsoft.com",
			"https://aws.amazon.com/documentation/",
		},
		SearchQueries: []string{
			"quarterly budget analysis template",
			"network configuration best practices",
			"cyber range operations checklist",
			"powershell script error 0x80070005",
			"linux sysadmin daily checklist",
			"pfsense firewall port forward setup",
			"remote desktop connection troubleshooting",
			"office 365 spreadsheet formulas",
		},
		StaticAssets: []string{
			"/css/bootstrap.min.css",
			"/css/app.css",
			"/js/jquery.min.js",
			"/js/app.js",
			"/images/logo.png",
			"/images/banner.jpg",
			"/favicon.ico",
			"/fonts/roboto-bold.woff2",
		},
		UserAgents: []string{
			"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/124.0.0.0 Safari/537.36",
			"Mozilla/5.0 (Windows NT 10.0; Win64; x64; rv:125.0) Gecko/20100101 Firefox/125.0",
			"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/124.0.0.0 Safari/537.36 Edg/124.0.0.0",
			"Mozilla/5.0 (X11; Linux x86_64; rv:125.0) Gecko/20100101 Firefox/125.0",
			"Mozilla/5.0 (X11; Ubuntu; Linux x86_64; rv:124.0) Gecko/20100101 Firefox/124.0",
		},
	}
}

// DefaultUserCredentials provides standard range credentials.
func DefaultUserCredentials() []UserCredential {
	return []UserCredential{
		{Username: "jdoe", Password: "Password123!", Domain: "RANGE", Role: "Executive"},
		{Username: "asmith", Password: "Spring2026Secure!", Domain: "RANGE", Role: "Finance"},
		{Username: "bjones", Password: "WinterCyber2026!", Domain: "RANGE", Role: "Developer"},
		{Username: "administrator", Password: "RangeAdmin2026#", Domain: "RANGE", Role: "SysAdmin"},
		{Username: "guest", Password: "Welcome2026!", Domain: "RANGE", Role: "Auditor"},
	}
}

func DefaultShareCorpus() ShareCorpus {
	return ShareCorpus{
		Shares: []models.ShareTarget{
			{Path: `\\fileserver.internal\CompanyShared`, Username: "asmith", Password: "Spring2026Secure!", Domain: "RANGE"},
			{Path: `\\fileserver.internal\Accounting`, Username: "asmith", Password: "Spring2026Secure!", Domain: "RANGE"},
			{Path: `\\fileserver.internal\Engineering`, Username: "bjones", Password: "WinterCyber2026!", Domain: "RANGE"},
			{Path: `/mnt/shares/public`, Username: "asmith", Password: "Spring2026Secure!", Domain: "RANGE"},
		},
		SampleFiles: []string{
			"Q1_Budget_Final.xlsx",
			"Staff_Meeting_Minutes.docx",
			"Network_Topology_v3.pdf",
			"Asset_Inventory_2026.csv",
			"Incident_Response_Plan_Draft.docx",
			"readme_operations.txt",
		},
	}
}

// LoadWebCorpus reads a web corpus file (plain text 1 site per line or JSON) or returns default.
func LoadWebCorpus(filePath string) (WebCorpus, error) {
	if filePath == "" {
		return DefaultWebCorpus(), nil
	}
	data, err := os.ReadFile(filePath)
	if err != nil {
		return DefaultWebCorpus(), fmt.Errorf("reading corpus file %s: %w", filePath, err)
	}
	data = bytes.TrimPrefix(data, []byte("\xef\xbb\xbf"))

	// If it's a JSON file or starts with '{', attempt JSON unmarshal
	trimmed := bytes.TrimSpace(data)
	if len(trimmed) > 0 && trimmed[0] == '{' {
		var corpus WebCorpus
		if err := json.Unmarshal(data, &corpus); err == nil {
			return corpus, nil
		}
	}

	// Plain text format: strictly one site per line (http:// or https://)
	lines := strings.Split(string(data), "\n")
	var urls []string
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "http://") || strings.HasPrefix(line, "https://") {
			urls = append(urls, line)
		}
	}

	if len(urls) > 0 {
		defaults := DefaultWebCorpus()
		return WebCorpus{
			InternetSites: urls,
			SearchQueries: defaults.SearchQueries,
			StaticAssets:  defaults.StaticAssets,
			UserAgents:    defaults.UserAgents,
		}, nil
	}

	return DefaultWebCorpus(), nil
}

// LoadUserCredentials reads a credentials JSON file or returns default.
func LoadUserCredentials(filePath string) ([]UserCredential, error) {
	if filePath == "" {
		return DefaultUserCredentials(), nil
	}
	data, err := os.ReadFile(filePath)
	if err != nil {
		return DefaultUserCredentials(), fmt.Errorf("reading creds file %s: %w", filePath, err)
	}
	data = bytes.TrimPrefix(data, []byte("\xef\xbb\xbf"))
	var creds []UserCredential
	if err := json.Unmarshal(data, &creds); err != nil {
		return DefaultUserCredentials(), fmt.Errorf("parsing creds json %s: %w", filePath, err)
	}
	return creds, nil
}
