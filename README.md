# RangeForge User Emulation Suite (`rangeforge-ue`)

### Open-Source Cyber Range & Network User Activity Simulator

[![License](https://img.shields.io/badge/License-Apache%202.0-blue.svg)](LICENSE)
[![Version](https://img.shields.io/badge/version-2.0.0--oss-brightgreen.svg)](https://github.com/zerocallbacks/rangeforge-user-emulation/releases)
[![Build](https://img.shields.io/badge/build-passing-brightgreen.svg)](#)
[![OS](https://img.shields.io/badge/OS-Windows%20%7C%20Linux%20%7C%20FreeBSD%20(pfSense)%20%7C%20macOS-informational.svg)](#)
[![Go](https://img.shields.io/badge/Go-1.24+-00ADD8.svg?logo=go)](https://golang.org)

---

## 1. Overview & Purpose

In cyber range training, red vs. blue team exercises, and security operations simulations, a major challenge is the lack of realistic background network and host activity. Without authentic user activity:
- **Blue Teams & SOC Analysts** can trivially isolate red team activity because idle networks produce zero baseline noise.
- **Red Teams** lack realistic host traffic, browsing activity, and file modifications to test operational security and evasion.
- **Network Sensors (IDS/IPS/SIEM/EDR)** cannot be accurately calibrated against normal user behavior.

**RangeForge User Emulation Suite (`rangeforge-ue`)** is a single, zero-dependency, cross-platform open-source Go binary designed to generate high-fidelity background user and network activity across any cyber range topology:
- **Windows Hosts & Servers**: Windows 10, 11, Server 2016, 2019, 2022, and 2025.
- **Linux Hosts & Servers**: Ubuntu, Debian, RHEL, CentOS, Rocky Linux, Alpine.
- **FreeBSD & pfSense Firewalls**: FreeBSD amd64 native target for single-interface edge routing.
- **macOS Workstations**: Apple Silicon (arm64) and Intel (amd64).

> [!NOTE]
> For the complete configuration manual, syntax reference, and architecture guide, see the [Comprehensive Manual (MANUAL.md)](MANUAL.md).

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
                                               │ HTTPS / TLS (Outbound Only)
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

### Roles:
- **Manager**: Central orchestrator providing the REST API, fleet heartbeat coordination, dynamic TLS listener, and interactive Web Console.
- **Controller**: Operator CLI console to inspect fleet metrics, dispatch remote tasks, or tune personas on the fly.
- **Host Agent**: Fully cross-platform agent running on target endpoints. Requires **zero inbound listening ports**—all communication is outbound HTTPS to the Manager.

---

## 3. Pre-Compiled Release Binaries

Pre-compiled standalone executables are provided directly in the [`bin/`](bin/) directory:

| Binary | Target Platform | Architecture | Description |
| :--- | :--- | :--- | :--- |
| [`bin/rangeforge-ue.exe`](bin/rangeforge-ue.exe) | Windows (10/11/Server) | `amd64` | Native Windows executable supporting all roles |
| [`bin/rangeforge-ue-linux`](bin/rangeforge-ue-linux) | Linux (glibc/musl) | `amd64` | Static Linux binary (Ubuntu, RHEL, Debian, Alpine) |
| [`bin/rangeforge-ue-freebsd`](bin/rangeforge-ue-freebsd) | FreeBSD / pfSense 2.7+ | `amd64` | Native FreeBSD target for pfSense & OPNsense routers |

---

## 4. Quick Start & Execution Modes

### A. Launching the Web UI Dashboard (Local Host)

Start the Manager Web UI and a local Host Agent over HTTPS on port 8443:

```powershell
# Windows
.\bin\rangeforge-ue.exe ui --port 8443
```

```bash
# Linux / Unix
./bin/rangeforge-ue-linux ui --port 8443
```

Open your browser at: **`https://127.0.0.1:8443/`** *(or `http://127.0.0.1:8443/`)*

- **Default Username**: `admin`
- **Default Password**: `rangeforge`

> [!TIP]
> The Web UI automatically multiplexes TLS and plaintext HTTP on the same port, allowing direct browser access without certificate configuration errors.

---

### B. Running the Central Coordinator (Manager)

On your range coordinator node:

```bash
./bin/rangeforge-ue-linux manager --host 0.0.0.0 --port 8443
```

### C. Running Endpoint Host Agents

Deploy the Host Agent on target virtual machines:

```powershell
# Windows Target
.\bin\rangeforge-ue.exe agent --manager https://10.0.0.1:8443 --persona office_worker
```

```bash
# Linux / Unix Target
./bin/rangeforge-ue-linux agent --manager https://10.0.0.1:8443 --persona developer
```

### D. Using the Controller CLI

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

### E. Standalone Evaluation Mode (No Manager Required)

Run local user emulation immediately without deploying a central server:

```powershell
.\bin\rangeforge-ue.exe standalone
```

---

## 5. Out-of-the-Box Persona Profiles

Personas model distinct organizational roles with authentic behavioral cadences:

| Persona | Role Intent | Web Targets | Share Ratio | Host Actions |
| :--- | :--- | :--- | :--- | :--- |
| **`office_worker`** | Knowledge Worker | Intranet portals, news, wikis | 95% Read / 5% Write | Word/Excel drafts, `whoami`, `hostname` |
| **`developer`** | Software Engineer | GitHub, StackOverflow, API docs | 70% Read / 30% Write | Code shares, `git`, compiler checks |
| **`sysadmin`** | IT Administrator | Router panels, tech docs | 60% Read / 40% Write | `netstat`, `route`, gateway ping audits |
| **`executive`** | Corporate Leadership | Briefings, financial metrics | 98% Read / 2% Write | Strategic memos, executive review notes |

Detailed YAML syntax for every profile is documented in [`configs/profiles/`](configs/profiles/) and [MANUAL.md](MANUAL.md).

---

## 6. Emulation Engines

### A. Web Browsing Emulation
- **Human Think Time**: Uses randomized Gaussian/Poisson dwell times between visits (e.g. 5–45 seconds) rather than robotic polling.
- **Authentic Signatures**: Rotates modern User-Agent strings (Chrome 124+, Edge, Firefox), Accept-Language, and cookie jars.
- **Asset Fetching**: Automatically fetches accompanying static assets (`.css`, `.js`, `.png`) to generate authentic packet flows for network sensors.
- **Proxy Support**: Respects standard `HTTP_PROXY`, `HTTPS_PROXY`, and `NO_PROXY` environment variables.

### B. File Share & Storage Emulation (SMB / CIFS / NFS)
- Interacts with configured network shares (`\\server\share` or `/mnt/shares`).
- Configurable read vs. write ratios to prevent disk bloat.
- Authors temporary office documents from business templates and automatically purges all created files upon agent shutdown.

### C. Host Activity & Safe Process Execution
- Executes benign diagnostic tools (`whoami`, `hostname`, `ping`, `netstat`, `git`) matching the assigned persona.
- Tunable noise levels (`silent`, `low`, `standard`, `high`) to calibrate SIEM and EDR detection rules.

### D. Stealth Command Execution Engine
- Allows range instructors to execute ad-hoc commands on targeted hosts:
  - **Windows**: Executes via `powershell.exe -ExecutionPolicy Bypass` with `-NoProfile` and shell history suppression.
  - **Unix**: Executes via `/bin/bash` or `/bin/sh`.
- Captures exit codes, stdout, stderr, and duration, reporting results back to the Controller in real-time.

---

## 7. Building from Source

RangeForge compiles with zero external C library dependencies using Go 1.24+:

```powershell
# Windows PowerShell
.\scripts\build.ps1 -Target all
```

```bash
# Linux / Unix Bash
chmod +x ./scripts/build.sh
./scripts/build.sh
```

```bash
# Via Makefile
make build-all
```

---

## 8. Documentation & Manual

For complete operational details, refer to:
- [**MANUAL.md**](MANUAL.md): Complete technical manual, configuration file schemas, REST API documentation, network topology planning, and troubleshooting guide.
- [**configs/**](configs/): Fully commented YAML and JSON configuration templates.
- [**LICENSE**](LICENSE): Apache 2.0 Open Source License.
