# RangeForge User Emulation Suite (`rangeforge-ue`)
### Open-Source Cyber Range & Network User Activity Simulator

[![License](https://img.shields.io/badge/License-Apache%202.0-blue.svg)](LICENSE)
[![Release](https://img.shields.io/badge/version-2.0.0--oss-brightgreen.svg)](https://github.com/zerocallbacks/rangeforge-user-emulation/releases)
[![Build Status](https://github.com/zerocallbacks/rangeforge-user-emulation/actions/workflows/ci.yml/badge.svg)](https://github.com/zerocallbacks/rangeforge-user-emulation/actions/workflows/ci.yml)
[![Target OS](https://img.shields.io/badge/OS-Windows%20%7C%20Linux%20%7C%20FreeBSD%20(pfSense)%20%7C%20macOS-green.svg)](#)
[![Language](https://img.shields.io/badge/Go-1.24+-00ADD8.svg)](https://golang.org)

---

## 1. Overview & Purpose
In high-fidelity cyber ranges, red vs. blue team exercises, and security operations training, a pervasive challenge is the lack of realistic, background network and host activity. Without normal user emulation, defensive blue teams can trivially pinpoint red team activity, and red teams lack realistic host environments and network noise to test operational security.

The **RangeForge User Emulation Suite (`rangeforge-ue`)** is a single, zero-dependency, cross-platform open-source Go binary that delivers realistic user emulation across any network environment:
- **Windows Hosts & Servers** (Windows 10, 11, Server 2016/2019/2022/2025)
- **Linux Hosts & Servers** (Ubuntu, Debian, RHEL, CentOS, Rocky, Alpine)
- **FreeBSD & pfSense Routers** (FreeBSD amd64 binary target, single-interface network traversal)
- **macOS Workstations** (Darwin amd64 & arm64 native support)

---

## 2. Architecture & Role Separation

`rangeforge-ue` operates as a single unified binary supporting three coordinated roles:

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
                          │  - Fleet Agent Registry & Heartbeat     │
                          │  - Campaign & Persona Scheduler         │
                          │  - Telemetry Aggregator                 │
                          │  - Remote Command Dispatcher            │
                          │  - Dynamic TLS / HTTPS (Port 8443)      │
                          └────────────────────┬────────────────────┘
                                               │ HTTPS / TLS
                                               │ Works over pfSense / Shared In-Game Network
                    ┌───────────────────────────┴───────────────────────────┐
                    │                                                       │
                    ▼                                                       ▼
┌─────────────────────────────────────┐                 ┌─────────────────────────────────────┐
│       RANGEFORGE HOST AGENT         │                 │       RANGEFORGE HOST AGENT         │
│         (Windows Host)              │                 │            (Unix Host)              │
├─────────────────────────────────────┤                 ├─────────────────────────────────────┤
│ • Telemetry: CPU, RAM, Disk, Net    │                 │ • Telemetry: /proc, sysctl, Net     │
│ • Web Emulation: Corpus, Jitter     │                 │ • Web Emulation: Corpus, Jitter     │
│ • Share Emulation: SMB (net use/ops)│                 │ • Share Emulation: CIFS/NFS/WebDAV  │
│ • Host Activity: Files, Processes   │                 │ • Host Activity: Bash, Cron, Files  │
│ • Custom Command Runner: PowerShell │                 │ • Custom Command Runner: Bash/sh    │
└─────────────────────────────────────┘                 └─────────────────────────────────────┘
```

### Roles
- **Manager**: Central orchestrator providing the REST API, fleet heartbeat coordination, dynamic TLS listener, and web console.
- **Controller**: Operator CLI console to inspect fleet metrics, dispatch remote tasks, or tune personas on the fly.
- **Host Agent**: Fully cross-platform agent running on target endpoints. Compiles and executes natively across Windows, Linux, FreeBSD, and macOS.

---

## 3. Emulation Engines & Realism

### A. Web Browsing Emulation
- **Corpus-Driven**: Reads intranet URLs (portals, wikis, ticketing systems) and external internet sites from `corpus/web_corpus.json`.
- **Human Think Time**: Uses randomized Gaussian/Poisson dwell times between visits (e.g. 5–25 seconds) rather than robotic polling.
- **Realistic Browser Signatures**: Rotates realistic modern User-Agent strings (Chrome 124+ on Win11, Edge, Firefox), Accept-Language, Sec-Fetch headers, and handles HTTP cookies.
- **Asset Fetching**: Automatically fetches accompanying static assets (`.css`, `.js`, `.png`, `favicon.ico`) to generate authentic packet patterns for network sensors.
- **Proxy Adaptability**: Respects standard `HTTP_PROXY`, `HTTPS_PROXY`, and `NO_PROXY` environment variables.

### B. File Share & Storage Emulation (SMB / Network Shares)
- Emulates realistic office document reading and authoring against shared network drives (`\\server\share` or `/mnt/shares`).
- Configurable read/write ratios (e.g. 70% read, 30% write).
- Automatically writes and cleans up temporary documents using realistic business templates.

### C. Host-Based Activity Emulation
- **Simulated Office Documents**: Creates realistic temporary files in user profile folders (`Documents`, `Downloads`, `Desktop`).
- **Benign Background Processes**: Spawns realistic background diagnostic tools (`whoami`, `hostname`, `ping`, `netstat`, `git`) matching the host's assigned persona.
- **Blue Team Noise Controls**: Tunable execution noise levels (`silent`, `low`, `standard`, `high`) to calibrate sensor alerts.

### D. Custom Command Execution Engine
- Allows range directors and instructors to execute ad-hoc administrative commands on targeted hosts:
  - **Windows**: Executes via `powershell.exe -ExecutionPolicy Bypass` or `cmd.exe`.
  - **Unix**: Executes via `/bin/bash` or `/bin/sh`.
- Captures exit code, stdout, stderr, and execution duration, streaming the result back to the Controller in real-time.

---

## 4. Multi-Network Adaptability & Robustness

RangeForge is engineered to be deployable across any network topology:

1. **Air-Gapped & Offline Networks**:
   - Zero external cloud dependencies.
   - Built-in embedded logo and asset fallbacks so standalone binaries operate without missing asset files.
   - Self-contained realistic user creds and web corpus.

2. **Enterprise & Proxied Networks**:
   - Built-in HTTP proxy support via `http.ProxyFromEnvironment`.
   - Bypasses local subnet addresses automatically via `NO_PROXY`.

3. **Dynamic Multi-Network TLS SANs**:
   - Self-signed certificate generation (`pkg/tlsutil`) automatically enumerates all local network interfaces (all IPv4 and IPv6 addresses), hostnames, FQDNs, and common domain wildcards (`*.local`, `*.lan`, `*.corp`, `*.internal`, `*.home.arpa`, `*.range.local`).
   - Eliminates certificate hostname mismatch errors across private subnets.

4. **Single-Interface & pfSense Firewalls**:
   - Agents require **zero inbound listening ports**—all communication is outbound HTTPS to the Manager.
   - Offline telemetry buffering: If a router or firewall reloads, agents buffer up to 50 telemetry samples in memory and retry with exponential backoff.

---

## 5. Quick Start & CLI Usage

### Downloading Pre-Compiled Binaries
Download the latest pre-compiled binary for your operating system from the **[GitHub Releases](https://github.com/zerocallbacks/rangeforge-user-emulation/releases)** page:
- `rangeforge-ue-windows-amd64.exe` (Windows 10/11/Server)
- `rangeforge-ue-linux-amd64` / `rangeforge-ue-linux-arm64` (Linux)
- `rangeforge-ue-freebsd-amd64` (FreeBSD / pfSense 2.7+)
- `rangeforge-ue-darwin-arm64` / `rangeforge-ue-darwin-amd64` (macOS)

### Launching the Web UI Dashboard (Local Host)
```powershell
# Start the Manager Web UI and local Host Agent over HTTPS on port 8443:
.\rangeforge-ue.exe ui --port 8443

# Open your browser at:
# https://127.0.0.1:8443/
```

### Dashboard Features & Themes
- **Dual Themes**: Toggle between **Ocean Sapphire** (cobalt deep blue cyber ops console) and **Carbon Black** (stealth dark operations with slate borders). Preferences are automatically saved in `localStorage`.
- **Fleet Telemetry**: Live interactive graphs of CPU, RAM, Disk, and Network traffic per agent.
- **Emulation Studio**: Live tuner for Web Requests Per Minute, dwell jitter, SMB read/write ratios, ICMP ping frequency & targets, and safe host processes.
- **Topology & Schedules**: Interactive visual network map and campaign timeline.
- **Integrated Command Console**: Dispatch PowerShell or Bash commands to any fleet host and view outputs directly in the browser terminal.

---

### Running the Central Manager
```bash
# Start central manager on coordinator node (binds 0.0.0.0, HTTPS default on 8443)
rangeforge-ue manager --host 0.0.0.0 --port 8443
```

### Running the Host Agent
```powershell
# Windows target
.\rangeforge-ue.exe agent --manager https://10.0.0.1:8443 --persona office_worker

# Linux / Unix target
./rangeforge-ue-linux agent --manager https://10.0.0.1:8443 --persona developer
```

### Using the Controller CLI
```bash
# List all active fleet agents
rangeforge-ue controller agents --manager https://10.0.0.1:8443

# Inspect live host telemetry
rangeforge-ue controller stats --manager https://10.0.0.1:8443 --agent host-123

# Execute remote command on a host
rangeforge-ue controller exec --manager https://10.0.0.1:8443 --agent host-123 --cmd "whoami /all"

# Reassign persona live
rangeforge-ue controller persona --manager https://10.0.0.1:8443 --agent host-123 --set-persona sysadmin
```

### Standalone Evaluation Mode
Run local emulation without a central manager:
```bash
rangeforge-ue standalone
```

---

## 6. Out-of-the-Box Personas

| Persona | Primary Focus | Web Targets | Share Ops | Host Actions |
| :--- | :--- | :--- | :--- | :--- |
| **`office_worker`** | Knowledge worker | Intranet portals, news, docs | Reads/writes Word & Excel drafts | Office temp locks, `whoami`, `hostname` |
| **`developer`** | Software engineering | GitHub, StackOverflow, language docs | Code & archive shares | Git queries, compiler checks |
| **`sysadmin`** | IT operations | Documentation, admin portals | IT audit shares | `netstat`, network route inspection |
| **`executive`** | Management | High-level briefings, corporate metrics | Financial memos & reports | Document reviews, executive notes |

---

## 7. Building from Source

Compile native binaries and cross-platform targets using Go 1.24+:

### Via PowerShell (Windows):
```powershell
.\scripts\build.ps1 -Target all
```

### Via Bash (Linux / Unix):
```bash
chmod +x ./scripts/build.sh
./scripts/build.sh
```

### Via Makefile:
```bash
make build-all
```

---

## 8. Community & Contributing

We welcome contributions from the security and developer community!
- **[Contributing Guidelines](CONTRIBUTING.md)**: Learn how to set up your environment, write tests, and submit PRs.
- **[Code of Conduct](CODE_OF_CONDUCT.md)**: Our standards for a welcoming community.
- **[Security Policy](SECURITY.md)**: Instructions for responsible vulnerability disclosure.
- **[Changelog](CHANGELOG.md)**: Release history and version notes.

---

## 9. License

RangeForge User Emulation Suite is licensed under the **Apache License 2.0**. See [LICENSE](LICENSE) for full details.
