# RangeForge User Emulation Suite (`rangeforge-ue`)

### Open-Source Cyber Range & Network User Activity Simulator

[![License](https://img.shields.io/badge/License-Apache%202.0-blue.svg)](LICENSE)
[![Version](https://img.shields.io/badge/version-2.0.0--oss-brightgreen.svg)](https://github.com/zerocallbacks/rangeforge-user-emulation/releases)
[![Build](https://img.shields.io/badge/build-passing-brightgreen.svg)](#)
[![OS](https://img.shields.io/badge/OS-Windows%20%7C%20Linux%20%7C%20FreeBSD%20(pfSense)%20%7C%20macOS-informational.svg)](#)
[![Go](https://img.shields.io/badge/Go-1.24+-00ADD8.svg?logo=go)](https://golang.org)
[![Zero Dependencies](https://img.shields.io/badge/dependencies-zero-success.svg)](#)

---

## Quick Links
- [Overview & The Realism Dilemma](#1-overview--the-realism-dilemma)
- [Architecture & Protocol](#2-system-architecture)
- [Pre-Compiled Release Binaries](#3-pre-compiled-release-binaries)
- [Quick Start Guide](#4-quick-start-guide)
- [Persona Profiles & Behavioral Models](#5-persona-profiles--behavioral-models)
- [Emulation Engines](#6-emulation-engines)
- [Web Operations Console](#7-interactive-web-operations-console)
- [Repository Layout](#8-repository-layout)
- [Building from Source](#9-building-from-source)
- [Comprehensive Manual](#10-technical-manual--documentation)

---

## 1. Overview & The Realism Dilemma

In cybersecurity training, red vs. blue team exercises, and security operations center (SOC) simulations, the single greatest point of artificiality is the **"Silent Range" phenomenon**:

| The Silent Cyber Range | The RangeForge Realistic Cyber Range |
| :--- | :--- |
| **Sterile Network**: Zero background HTTP/SMB traffic makes any red team beacon or port scan immediately stick out. | **Dynamic Baseline Noise**: Realistic web navigation, SMB file reads/writes, and diagnostic traffic blend into normal user flows. |
| **Trivial Blue Team Detection**: SOC analysts and SIEM correlation rules flag an anomalous process because no other processes ever run. | **Calibrated Alert Volume**: Analysts must sift through authentic background processes (`whoami`, `netstat`, `hostname`, `git`) to isolate malicious activity. |
| **Unrealistic Evasion Testing**: Red teams cannot practice living-off-the-land techniques or traffic blending against active traffic. | **Authentic Detection Engineering**: Offensive teams test tradecraft against authentic network noise; defensive teams tune detection thresholds with precision. |

**RangeForge User Emulation Suite (`rangeforge-ue`)** is a single, zero-dependency, cross-platform open-source Go binary that generates high-fidelity, mathematically randomized user and network activity across any cyber range topology:

- **Windows Workstations & Servers**: Windows 10, 11, Server 2016, 2019, 2022, and 2025.
- **Linux Endpoints & Infrastructure**: Ubuntu, Debian, RHEL, CentOS, Rocky Linux, Alpine.
- **FreeBSD & pfSense / OPNsense Firewalls**: Native FreeBSD `amd64` binary for single-interface edge routing.
- **macOS Workstations**: Apple Silicon (`arm64`) and Intel (`amd64`).

> [!NOTE]
> For the complete configuration manual, syntax reference, and architecture guide, see the [Comprehensive Manual (MANUAL.md)](MANUAL.md).

---

## 2. System Architecture

RangeForge is engineered as a **single unified executable** supporting three coordinated operational tiers:

```
                          ┌─────────────────────────────────────────┐
                          │         RANGEFORGE CONTROLLER           │
                          │   (CLI Console / Operator Dashboard)    │
                          └────────────────────┬────────────────────┘
                                               │ REST API / CLI
                                               ▼
                          ┌─────────────────────────────────────────┐
                          │          RANGEFORGE MANAGER             │
                          │       (Central Orchestrator)            │
                          │  - Fleet Agent Registry & Heartbeats    │
                          │  - Campaign & Persona Scheduler         │
                          │  - Telemetry Aggregator & Sparklines    │
                          │  - Remote Command Dispatcher            │
                          │  - Dual-Protocol TLS/HTTP (Port 8443)   │
                          └────────────────────┬────────────────────┘
                                               │ HTTPS / TLS (Outbound Only)
                                               │ Traverses pfSense & NAT Firewalls
                     ┌─────────────────────────┴─────────────────────────┐
                     │                                                   │
                     ▼                                                   ▼
 ┌─────────────────────────────────────┐             ┌─────────────────────────────────────┐
 │       RANGEFORGE HOST AGENT         │             │       RANGEFORGE HOST AGENT         │
 │         (Windows Target)            │             │          (Unix Target)              │
 ├─────────────────────────────────────┤             ├─────────────────────────────────────┤
 │ • Telemetry: CPU, RAM, Disk, Net    │             │ • Telemetry: /proc, sysctl, Net     │
 │ • Web Emulation: Corpus, Jitter     │             │ • Web Emulation: Corpus, Jitter     │
 │ • Share Emulation: SMB (net use/ops)│             │ • Share Emulation: CIFS/NFS/WebDAV  │
 │ • Host Activity: Files, Processes   │             │ • Host Activity: Bash, Cron, Files  │
 │ • Remote Command: PowerShell        │             │ • Remote Command: Bash/sh           │
 └─────────────────────────────────────┘             └─────────────────────────────────────┘
```

### Core Architectural Guarantees:
1. **Zero External Dependencies**: Pure Go with static linking. Requires no Python runtime, Node.js, libc libraries, or database engines.
2. **Outbound-Only HTTPS**: Host Agents open **zero inbound listening ports**. All communications are outbound HTTPS check-ins to the Manager, naturally traversing pfSense, NAT gateways, and stateful firewalls.
3. **Dual-Protocol TLS Multiplexing**: The Manager dynamically detects and serves both TLS (HTTPS) and plaintext HTTP on port 8443 simultaneously, preventing certificate handshake lockouts during testing.
4. **Zero Persistent Artifacts**: All temporary files created during emulation (drafts, locks, logs) are tracked in memory. When emulation halts, all synthetic files are automatically unlinked.

---

## 3. Pre-Compiled Release Binaries

Pre-compiled standalone executables are provided directly in the [`bin/`](bin/) directory of this repository:

| Executable | Target Platform | Architecture | Binary Type | Description |
| :--- | :--- | :--- | :--- | :--- |
| [`bin/rangeforge-ue.exe`](bin/rangeforge-ue.exe) | Windows (10/11/Server) | `amd64` | Native PE | Native Windows binary supporting UI, Manager, Agent, and Controller |
| [`bin/rangeforge-ue-linux`](bin/rangeforge-ue-linux) | Linux (kernel 3.10+) | `amd64` | Static ELF | Self-contained static Linux executable (Ubuntu, Debian, RHEL, Alpine) |
| [`bin/rangeforge-ue-freebsd`](bin/rangeforge-ue-freebsd) | FreeBSD / pfSense 2.7+ | `amd64` | Native ELF | Native FreeBSD binary for pfSense and OPNsense edge firewalls |

To verify binary integrity:
```bash
# Linux / macOS
sha256sum bin/rangeforge-ue*

# Windows PowerShell
Get-FileHash bin\rangeforge-ue.exe -Algorithm SHA256
```

---

## 4. Quick Start Guide

### Mode A: Launch the Web UI Dashboard (Local Evaluation)

Start the Manager Web Console and a local Host Agent over HTTPS on port 8443:

```powershell
# Windows
.\bin\rangeforge-ue.exe ui --port 8443
```

```bash
# Linux / Unix
./bin/rangeforge-ue-linux ui --port 8443
```

- **Web Console URL**: [`https://127.0.0.1:8443/`](https://127.0.0.1:8443/) *(or `http://127.0.0.1:8443/`)*
- **Default Username**: `admin`
- **Default Password**: `rangeforge`

> [!TIP]
> The Web UI automatically multiplexes TLS and plaintext HTTP on the same port. If your browser displays a self-signed certificate warning, simply click **Advanced -> Proceed**, or open `http://127.0.0.1:8443/` directly.

---

### Mode B: Deploy Central Manager (Dedicated Coordinator)

On your central cyber range orchestrator node:

```bash
# Linux Coordinator
./bin/rangeforge-ue-linux manager --host 0.0.0.0 --port 8443
```

---

### Mode C: Deploy Endpoint Host Agents

Deploy the Host Agent on target virtual machines across different subnets:

```powershell
# Windows Target (Office Worker Persona)
.\bin\rangeforge-ue.exe agent --manager https://10.0.0.1:8443 --persona office_worker
```

```bash
# Linux Target (Developer Persona)
./bin/rangeforge-ue-linux agent --manager https://10.0.0.1:8443 --persona developer
```

```bash
# FreeBSD / pfSense Gateway (Sysadmin Persona)
./bin/rangeforge-ue-freebsd agent --manager https://10.0.0.1:8443 --persona sysadmin
```

---

### Mode D: Operator Controller CLI

Exercise directors and instructors can inspect fleet metrics and dispatch remote tasks via the CLI:

```bash
# List all active fleet endpoints
rangeforge-ue controller agents --manager https://10.0.0.1:8443

# Inspect real-time host telemetry (CPU, RAM, active tools)
rangeforge-ue controller stats --manager https://10.0.0.1:8443 --agent host-123

# Execute remote administrative command
rangeforge-ue controller exec --manager https://10.0.0.1:8443 --agent host-123 --cmd "whoami /all"

# Reassign persona live without agent restart
rangeforge-ue controller persona --manager https://10.0.0.1:8443 --agent host-123 --set-persona sysadmin
```

---

### Mode E: Standalone Evaluation Mode (No Server Required)

Run local user emulation immediately without deploying a central coordinator:

```powershell
.\bin\rangeforge-ue.exe standalone
```

---

## 5. Persona Profiles & Behavioral Models

Personas model distinct organizational roles with authentic behavioral cadences, randomized jitter, and realistic process trees:

| Persona | Role Intent | Web Browsing Targets | Share Read/Write | Benign Host Commands | Noise Level |
| :--- | :--- | :--- | :--- | :--- | :--- |
| **`office_worker`** | Knowledge Worker | Intranet portals, corporate wikis, news | 95% Read / 5% Write | Word/Excel drafts, `whoami`, `hostname` | Standard (1.0x) |
| **`developer`** | Software Engineer | GitHub, StackOverflow, API docs | 70% Read / 30% Write | Source repositories, `git`, compiler checks | High (2.0x) |
| **`sysadmin`** | IT Administrator | Router panels, tech docs, dashboards | 60% Read / 40% Write | `netstat`, `route`, gateway ping audits | Elevated (2.5x) |
| **`executive`** | Corporate Leadership | Briefings, financial metrics, WSJ | 98% Read / 2% Write | Strategic memos, executive review notes | Low (0.5x) |

Detailed YAML syntax for every profile is documented in [`configs/profiles/`](configs/profiles/) and [MANUAL.md](MANUAL.md).

---

## 6. Emulation Engines

### A. Web Browsing Engine
- **Human Think Time**: Uses randomized Gaussian/Poisson dwell times between visits (e.g. 5–45 seconds) rather than robotic periodic polling.
- **Authentic Signatures**: Rotates modern User-Agent strings (Chrome 124+, Edge, Firefox), Accept-Language, and cookie jars.
- **Asset Fetching**: Automatically fetches accompanying static assets (`.css`, `.js`, `.png`) to generate authentic packet flows for network sensors.
- **Proxy Support**: Respects standard `HTTP_PROXY`, `HTTPS_PROXY`, and `NO_PROXY` environment variables.

### B. File Share & Storage Engine (SMB / CIFS / NFS)
- Interacts with configured network shares (`\\server\share` or `/mnt/shares`).
- Configurable read vs. write ratios to prevent disk bloat.
- Authors temporary office documents from business templates and automatically purges all created files upon agent shutdown.

### C. Host Activity & Process Execution Engine
- Executes benign diagnostic tools (`whoami`, `hostname`, `ping`, `netstat`, `git`) matching the assigned persona.
- Tunable noise levels (`silent`, `low`, `standard`, `high`) to calibrate SIEM and EDR detection rules.

### D. Stealth Remote Command Runner
- Allows range instructors to execute ad-hoc commands on targeted hosts:
  - **Windows**: Executes via `powershell.exe -ExecutionPolicy Bypass` with `-NoProfile` and shell history suppression.
  - **Unix**: Executes via `/bin/bash` or `/bin/sh`.
- Captures exit codes, stdout, stderr, and duration, reporting results back to the Controller in real-time.

---

## 7. Interactive Web Operations Console

Access the operations console at: **`https://127.0.0.1:8443/`**

- **Real-Time Telemetry**: Live SVG sparklines plotting Fleet Average CPU and RAM with 15-point historical smoothing.
- **Fleet Roster**: 8-column host table displaying OS version, network IPs, assigned persona, and live telemetry.
- **Interactive SVG Topology Map**: Visual representation of subnets, edge routers, and connected agents.
- **Emulation Studio**: Live tuner for Web RPM, dwell jitter, SMB read/write ratios, and safe host processes.
- **Host Inspector Modal**: Deep-dive inspection of CPU cores, memory usage, MAC address, MTU, and installed tools.
- **Integrated Command Console**: Dispatch PowerShell or Bash commands to any endpoint directly from the browser.
- **Dual Themes**: Toggle between **Ocean Sapphire** (cobalt deep blue cyber ops console) and **Carbon Black** (stealth dark operations with slate borders). Preferences persist in `localStorage`.

---

## 8. Repository Layout

```
rangeforge-user-emulation/
├── bin/                                # Pre-compiled release executables
│   ├── rangeforge-ue.exe               # Windows amd64 executable
│   ├── rangeforge-ue-linux             # Linux amd64 static binary
│   └── rangeforge-ue-freebsd           # FreeBSD / pfSense amd64 binary
├── certs/                              # TLS X.509 certificates (auto-generated if missing)
│   ├── server.crt                      # Server certificate
│   └── server.key                      # Server private key
├── cmd/                                # CLI entrypoint
│   └── rangeforge-ue/main.go           # Unified command dispatcher
├── configs/                            # Configuration files with syntax explanations
│   ├── manager.yaml                    # Manager server configuration
│   ├── agent_windows.yaml              # Windows host agent configuration
│   ├── agent_unix.yaml                 # Unix host agent configuration
│   ├── admin_auth.json                 # Salted SHA-256 operator credentials
│   ├── topology.json                   # Network topology gateway definitions
│   ├── ranges.json                     # Cyber range scenario definitions
│   └── profiles/                       # Persona emulation profiles
│       ├── office_worker.yaml          # Office Worker persona profile
│       ├── developer.yaml              # Software Developer persona profile
│       ├── sysadmin.yaml               # System Administrator persona profile
│       └── executive.yaml              # Corporate Executive persona profile
├── corpus/                             # Emulation corpora
│   ├── web_corpus.json                 # Target URLs, keywords, and User-Agents
│   ├── user_creds.json                 # Simulated personas and credentials
│   └── wordlist.txt                    # Document authoring scenario dictionary
├── pkg/                                # Core Go packages
│   ├── agent/                          # Agent client and telemetry
│   ├── controller/                     # Operator CLI and REST client
│   ├── manager/                        # Central server, registry, and Web UI
│   ├── models/                         # Shared data structures and contracts
│   └── tlsutil/                        # Dynamic TLS/HTTP multiplexing listener
├── scripts/                            # Build automation scripts
│   ├── build.ps1                       # Windows PowerShell cross-compilation
│   └── build.sh                        # Unix Bash cross-compilation
├── Makefile                            # GNU Makefile build targets
├── LICENSE                             # Apache 2.0 Open Source License
├── MANUAL.md                           # Comprehensive Operator & Technical Manual
└── README.md                           # This landing page
```

---

## 9. Building from Source

RangeForge compiles with zero external C library dependencies using Go 1.24+:

### Windows (PowerShell)
```powershell
# Build all cross-platform binaries into bin/
.\scripts\build.ps1 -Target all

# Build native Windows binary only
.\scripts\build.ps1 -Target windows
```

### Linux / Unix (Bash)
```bash
# Make script executable and build all targets
chmod +x ./scripts/build.sh
./scripts/build.sh
```

### GNU Makefile
```bash
# Build all release binaries
make build-all

# Run all unit and integration tests
make test
```

---

## 10. Technical Manual & Documentation

For complete operational details, refer to:
- [**MANUAL.md**](MANUAL.md): Complete technical manual, configuration file schemas, REST API documentation, network topology planning, SIEM detection engineering, and troubleshooting guide.
- [**configs/**](configs/): Fully commented YAML and JSON configuration templates.
- [**LICENSE**](LICENSE): Apache 2.0 Open Source License.
