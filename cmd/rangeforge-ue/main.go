package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"rangeforge-ue/pkg/agent"
	"rangeforge-ue/pkg/branding"
	"rangeforge-ue/pkg/config"
	"rangeforge-ue/pkg/controller"
	"rangeforge-ue/pkg/manager"
)

func main() {
	if len(os.Args) < 2 {
		printGeneralUsage()
		os.Exit(1)
	}

	command := strings.ToLower(os.Args[1])

	switch command {
	case "ui", "dashboard":
		runUI(os.Args[2:])
	case "manager":
		runManager(os.Args[2:])
	case "controller":
		runController(os.Args[2:])
	case "agent":
		runAgent(os.Args[2:])
	case "standalone":
		runStandalone(os.Args[2:])
	case "version", "-v", "--version":
		branding.PrintBanner("")
		os.Exit(0)
	case "help", "-h", "--help":
		printGeneralUsage()
		os.Exit(0)
	default:
		fmt.Printf("%s[ERROR]%s Unknown command: '%s'\n\n", branding.Red+branding.Bold, branding.Reset, command)
		printGeneralUsage()
		os.Exit(1)
	}
}

func printGeneralUsage() {
	branding.PrintBanner("")
	fmt.Printf("%sUSAGE:%s\n", branding.Bold+branding.Cyan, branding.Reset)
	fmt.Printf("  rangeforge-ue <command> [arguments...]\n\n")
	fmt.Printf("%sCORE ROLES & DASHBOARD:%s\n", branding.Bold+branding.Yellow, branding.Reset)
	fmt.Printf("  %sui%s          Launch Web UI Dashboard & Local Range Host (Interactive Browser UI)\n", branding.Cyan+branding.Bold, branding.Reset)
	fmt.Printf("  %smanager%s     Start Central Exercise Manager (Cross-platform Orchestrator)\n", branding.Green+branding.Bold, branding.Reset)
	fmt.Printf("  %scontroller%s  Operator CLI Console to inspect fleet & dispatch tasks\n", branding.Green+branding.Bold, branding.Reset)
	fmt.Printf("  %sagent%s       Host Emulation Daemon (Cross-platform: Windows, Linux, FreeBSD, macOS)\n", branding.Green+branding.Bold, branding.Reset)
	fmt.Printf("  %sstandalone%s  Run local user emulation without a manager (evaluation mode)\n", branding.Green+branding.Bold, branding.Reset)
	fmt.Printf("  %sversion%s     Display version and open-source project metadata\n\n", branding.Green+branding.Bold, branding.Reset)
	fmt.Printf("%sQUICK EXAMPLES:%s\n", branding.Bold+branding.Cyan, branding.Reset)
	fmt.Printf("  # Start Web UI Dashboard on this host (HTTPS default on port 8443)\n")
	fmt.Printf("  rangeforge-ue ui --port 8443\n\n")
	fmt.Printf("  # Start central manager on coordinator node (TLS/HTTPS enforced)\n")
	fmt.Printf("  rangeforge-ue manager --host 0.0.0.0 --port 8443\n\n")
	fmt.Printf("  # Run host agent on target host\n")
	fmt.Printf("  rangeforge-ue agent --manager https://10.0.0.1:8443 --persona office_worker\n\n")
	fmt.Printf("  # View all connected cyber range hosts from controller\n")
	fmt.Printf("  rangeforge-ue controller agents --manager https://10.0.0.1:8443\n\n")
	fmt.Printf("  # Dispatch custom command to host from controller\n")
	fmt.Printf("  rangeforge-ue controller exec --agent host-123 --cmd \"whoami /all\"\n\n")
}

func runManager(args []string) {
	fs := flag.NewFlagSet("manager", flag.ExitOnError)
	configFile := fs.String("config", "", "Path to manager YAML/JSON config file")
	host := fs.String("host", "0.0.0.0", "Bind address for Manager API")
	port := fs.Int("port", 8443, "Port for Manager API (default 8443 for HTTPS)")
	timeout := fs.Int("timeout", 30, "Agent heartbeat timeout in seconds")
	persona := fs.String("persona", "office_worker", "Default persona profile for new agents")
	profilesDir := fs.String("profiles-dir", "./configs/profiles", "Directory containing custom persona profiles")
	allowUnsupportedOS := fs.Bool("allow-unsupported-os", false, "Allow running Manager on non-Unix environments (lab evaluation only)")
	tlsEnabled := fs.Bool("tls", true, "Enable TLS / HTTPS listener (HTTPS only)")
	certPath := fs.String("cert", "./certs/server.crt", "Path to TLS server certificate")
	keyPath := fs.String("key", "./certs/server.key", "Path to TLS server private key")

	_ = fs.Parse(args)

	cfg := config.DefaultManagerConfig()
	if *configFile != "" {
		var err error
		cfg, err = config.LoadManagerConfig(*configFile)
		if err != nil {
			fmt.Printf("%s[CONFIG ERROR]%s %v\n", branding.Red+branding.Bold, branding.Reset, err)
			os.Exit(1)
		}
	} else {
		cfg.ListenHost = *host
		cfg.ListenPort = *port
		cfg.HeartbeatTimeoutSec = *timeout
		cfg.DefaultPersona = *persona
		cfg.ProfilesDir = *profilesDir
		cfg.AllowUnsupportedOS = *allowUnsupportedOS
		cfg.TLSEnabled = *tlsEnabled
		cfg.TLSCertPath = *certPath
		cfg.TLSKeyPath = *keyPath
	}

	srv, err := manager.NewServer(cfg)
	if err != nil {
		fmt.Printf("%s[STARTUP ERROR]%s %v\n", branding.Red+branding.Bold, branding.Reset, err)
		os.Exit(1)
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	if err := srv.Start(ctx); err != nil {
		fmt.Printf("%s[SERVER ERROR]%s %v\n", branding.Red+branding.Bold, branding.Reset, err)
		os.Exit(1)
	}
}

func runController(args []string) {
	if len(args) == 0 {
		printControllerUsage()
		os.Exit(1)
	}

	subCmd := strings.ToLower(args[0])
	subArgs := args[1:]

	fs := flag.NewFlagSet("controller", flag.ExitOnError)
	managerURL := fs.String("manager", "https://127.0.0.1:8443", "Target Manager base URL")
	agentID := fs.String("agent", "", "Target Agent ID (or 'all' for broadcast commands)")
	cmd := fs.String("cmd", "", "Command to execute on agent")
	shell := fs.String("shell", "", "Shell to invoke (powershell, cmd, bash, sh; empty = auto)")
	setPersona := fs.String("set-persona", "", "Persona name to assign to the target agent")
	timeoutSec := fs.Int("timeout", 15, "Command execution timeout in seconds")
	outFormat := fs.String("format", "table", "Output format: table or json")

	_ = fs.Parse(subArgs)
	_ = outFormat

	client := controller.NewClient(*managerURL, *timeoutSec)
	cli := controller.NewCLI(client)

	switch subCmd {
	case "agents", "list":
		_ = cli.ListAgents()
	case "stats", "telemetry":
		if *agentID == "" {
			fmt.Printf("%s[ERROR]%s --agent <agent-id> is required for stats.\n", branding.Red+branding.Bold, branding.Reset)
			os.Exit(1)
		}
		_ = cli.ShowHostStats(*agentID)
	case "exec", "cmd":
		if *agentID == "" || *cmd == "" {
			fmt.Printf("%s[ERROR]%s Both --agent <agent-id> and --cmd \"<command>\" are required for exec.\n", branding.Red+branding.Bold, branding.Reset)
			os.Exit(1)
		}
		_ = cli.RunCommand(*agentID, *cmd, *shell, *timeoutSec)
	case "personas", "profiles":
		_ = cli.ListPersonas()
	case "persona", "set-persona":
		if *agentID == "" || *setPersona == "" {
			fmt.Printf("%s[ERROR]%s Both --agent <agent-id> and --set-persona <name> are required.\n", branding.Red+branding.Bold, branding.Reset)
			os.Exit(1)
		}
		_ = cli.SetPersona(*agentID, *setPersona)
	case "help", "-h":
		printControllerUsage()
	default:
		fmt.Printf("%s[ERROR]%s Unknown controller sub-command: '%s'\n", branding.Red+branding.Bold, branding.Reset, subCmd)
		printControllerUsage()
		os.Exit(1)
	}
}

func printControllerUsage() {
	branding.PrintBanner("Operator Controller")
	fmt.Printf("%sCONTROLLER SUB-COMMANDS:%s\n", branding.Bold+branding.Cyan, branding.Reset)
	fmt.Printf("  rangeforge-ue controller agents    [--manager URL]                 List active fleet agents\n")
	fmt.Printf("  rangeforge-ue controller stats     --agent ID [--manager URL]      Show live host metrics\n")
	fmt.Printf("  rangeforge-ue controller exec      --agent ID --cmd \"...\"         Run command on host\n")
	fmt.Printf("  rangeforge-ue controller personas  [--manager URL]                 List available personas\n")
	fmt.Printf("  rangeforge-ue controller persona   --agent ID --set-persona NAME   Assign persona to host\n")
}

func runAgent(args []string) {
	fs := flag.NewFlagSet("agent", flag.ExitOnError)
	configFile := fs.String("config", "", "Path to agent YAML/JSON config file")
	managerURL := fs.String("manager", "https://127.0.0.1:8443", "Manager central URL (HTTPS)")
	persona := fs.String("persona", "office_worker", "Assigned persona profile")
	heartbeat := fs.Int("heartbeat", 5, "Heartbeat interval in seconds")
	bindIface := fs.String("iface", "", "Network interface to bind (empty = auto)")
	webCorpusFile := fs.String("web-corpus", "", "Path to web_corpus.json")
	credsFile := fs.String("creds", "", "Path to user_creds.json")
	tags := fs.String("tags", "", "Comma-separated tags for this agent")

	_ = fs.Parse(args)

	cfg := config.DefaultAgentConfig()
	if *configFile != "" {
		var err error
		cfg, err = config.LoadAgentConfig(*configFile)
		if err != nil {
			fmt.Printf("%s[CONFIG ERROR]%s %v\n", branding.Red+branding.Bold, branding.Reset, err)
			os.Exit(1)
		}
	} else {
		cfg.ManagerURL = *managerURL
		cfg.Persona = *persona
		cfg.HeartbeatSec = *heartbeat
		cfg.BindInterface = *bindIface
		cfg.WebCorpusFile = *webCorpusFile
		cfg.CredsFile = *credsFile
		if *tags != "" {
			cfg.Tags = strings.Split(*tags, ",")
		}
	}

	ag, err := agent.NewAgent(cfg)
	if err != nil {
		fmt.Printf("%s[AGENT INIT ERROR]%s %v\n", branding.Red+branding.Bold, branding.Reset, err)
		os.Exit(1)
	}

	if err := ag.Start(context.Background()); err != nil {
		fmt.Printf("%s[AGENT ERROR]%s %v\n", branding.Red+branding.Bold, branding.Reset, err)
		os.Exit(1)
	}
}

func runStandalone(args []string) {
	branding.PrintBanner("Standalone Evaluation")
	fmt.Printf("%sRunning RangeForge User Emulation in standalone mode (no external manager required).%s\n",
		branding.Bold+branding.Cyan, branding.Reset)

	cfg := config.DefaultAgentConfig()
	cfg.ManagerURL = "http://localhost:9999" // Mock
	ag, err := agent.NewAgent(cfg)
	if err != nil {
		fmt.Printf("%s[STANDALONE ERROR]%s %v\n", branding.Red+branding.Bold, branding.Reset, err)
		os.Exit(1)
	}

	// In standalone, start emulation routines directly
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go func() {
		time.Sleep(1 * time.Second)
		fmt.Println("\nPress Ctrl+C to terminate standalone emulation.")
	}()

	_ = ag.Start(ctx)
}

func runUI(args []string) {
	branding.PrintBanner("Web UI & Range Host")
	fs := flag.NewFlagSet("ui", flag.ExitOnError)
	port := fs.Int("port", 8443, "Port for Manager Web UI (HTTPS default 8443)")
	host := fs.String("host", "127.0.0.1", "Host address for Manager Web UI")
	tlsEnabled := fs.Bool("tls", true, "Enforce HTTPS / TLS mode (certs stored in ./certs/)")
	certPath := fs.String("cert", "./certs/server.crt", "Path to TLS server certificate")
	keyPath := fs.String("key", "./certs/server.key", "Path to TLS server private key")
	persona := fs.String("persona", "office_worker", "Persona for the local host agent")
	_ = fs.Parse(args)

	cfg := config.DefaultManagerConfig()
	cfg.ListenHost = *host
	cfg.ListenPort = *port
	cfg.TLSEnabled = true // HTTPS only strictly enforced (*tlsEnabled accepted for CLI compatibility)
	_ = tlsEnabled
	cfg.TLSCertPath = *certPath
	cfg.TLSKeyPath = *keyPath
	cfg.AllowUnsupportedOS = true // Permitted for UI viewing on the host

	srv, err := manager.NewServer(cfg)
	if err != nil {
		fmt.Printf("%s[UI ERROR]%s %v\n", branding.Red+branding.Bold, branding.Reset, err)
		os.Exit(1)
	}

	go func() {
		if err := srv.Start(context.Background()); err != nil {
			fmt.Printf("%s[MANAGER ERROR]%s %v\n", branding.Red+branding.Bold, branding.Reset, err)
		}
	}()

	// Wait 500ms for server to bind
	time.Sleep(500 * time.Millisecond)

	protocol := "https"
	uiURL := fmt.Sprintf("%s://%s:%d/", protocol, *host, *port)
	fmt.Printf("\n%s╔══════════════════════════════════════════════════════════════════════╗%s\n", branding.Cyan+branding.Bold, branding.Reset)
	fmt.Printf("%s║  RANGEFORGE WEB UI DASHBOARD READY!                                  ║%s\n", branding.Cyan+branding.Bold, branding.Reset)
	fmt.Printf("%s║  Open your browser at: %s%-45s%s ║%s\n", branding.Cyan+branding.Bold, branding.Yellow+branding.Bold, uiURL, branding.Cyan+branding.Bold, branding.Reset)
	fmt.Printf("%s║  TLS Certificates:     %-45s %s║%s\n", branding.Cyan+branding.Bold, *certPath, branding.Cyan+branding.Bold, branding.Reset)
	fmt.Printf("%s║  Protocol Security:    %-45s %s║%s\n", branding.Cyan+branding.Bold, "HTTPS ONLY (TLS 1.3 / 1.2 Enforced)", branding.Cyan+branding.Bold, branding.Reset)
	fmt.Printf("%s╚══════════════════════════════════════════════════════════════════════╝%s\n\n", branding.Cyan+branding.Bold, branding.Reset)

	// Launch local agent so the user sees live host telemetry and user emulation
	agentCfg := config.DefaultAgentConfig()
	agentHost := *host
	if agentHost == "0.0.0.0" || agentHost == "" {
		agentHost = "127.0.0.1"
	}
	agentCfg.ManagerURL = fmt.Sprintf("https://%s:%d", agentHost, *port)
	agentCfg.Persona = *persona
	agentCfg.HeartbeatSec = 2

	ag, err := agent.NewAgent(agentCfg)
	if err != nil {
		fmt.Printf("%s[AGENT INIT NOTE]%s %v\n", branding.Yellow, branding.Reset, err)
	} else {
		go func() {
			_ = ag.Start(context.Background())
		}()
	}

	// Keep running until Ctrl+C
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
	<-sigChan
	fmt.Println("\nShutting down RangeForge UI & cleaning up User Emulation artifacts...")
	if ag != nil {
		ag.Stop()
	}
}
