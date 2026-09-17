# RangeForge User Emulation Suite (`rangeforge-ue`)
## Operator & Technical Reference Manual

---

## Table of Contents
1. [Architecture & Component Interaction](#1-architecture--component-interaction)
2. [Operating Roles & Modes](#2-operating-roles--modes)
3. [Configuration Reference & Syntax Guide](#3-configuration-reference--syntax-guide)
   - [Manager Server Configuration (`configs/manager.yaml`)](#a-manager-server-configuration)
   - [Host Agent Configurations (`configs/agent_windows.yaml` & `agent_unix.yaml`)](#b-host-agent-configurations)
   - [Persona Emulation Profiles (`configs/profiles/*.yaml`)](#c-persona-emulation-profiles)
   - [Interactive Network Topology (`configs/topology.json`)](#d-interactive-network-topology)
   - [Cyber Range Scenarios (`configs/ranges.json`)](#e-cyber-range-scenarios)
   - [Administrative Authentication (`configs/admin_auth.json`)](#f-administrative-authentication)
   - [Web & Identity Corpus (`corpus/`)](#g-web--identity-corpus)
4. [Multi-Network Deployment & Firewall Traversal](#4-multi-network-deployment--firewall-traversal)
5. [Emulation Engines & Realism Calibration](#5-emulation-engines--realism-calibration)
6. [Interactive Web UI Dashboard](#6-interactive-web-ui-dashboard)
7. [Controller CLI & REST API Reference](#7-controller-cli--rest-api-reference)
8. [Operational Security & Zero-Artifact Guarantees](#8-operational-security--zero-artifact-guarantees)
9. [Troubleshooting & Frequently Asked Questions](#9-troubleshooting--frequently-asked-questions)

---

## 1. Architecture & Component Interaction

RangeForge is engineered as a **single, unified, zero-dependency Go binary**. The binary dynamically activates specific internal engines based on the invoked subcommand:

```
+-----------------------------------------------------------------------------+
|                          RANGEFORGE UNIFIED BINARY                          |
|                                                                             |
|  [manager]              [controller]              [agent]      [standalone] |
|     |                        |                       |              |       |
|     v                        v                       v              v       |
|  Orchestrator           CLI Console             Host Daemon    Local Runner |
|  REST API & UI          Fleet Inspection        Telemetry      Self-Contained|
|  TLS Listener           Command Dispatch        Emulators      No Coordinator|
+-----------------------------------------------------------------------------+
```

### Communication Flow
1. **Outbound-Only HTTPS Handshake**: Host Agents initiate an HTTPS POST to `/api/v1/agent/register` on the Manager (port 8443). Agents require **no inbound listening sockets**.
2. **Dynamic Heartbeat & Telemetry Synchronization**: Agents send periodic heartbeats (default every 5s) transmitting CPU utilization, RAM consumption, disk space, network adapter IPs, and an inventory of functional local binaries (e.g. `ping`, `git`, `powershell`, `netstat`).
3. **Task & Command Dispatching**: Operators issue remote commands or persona reassignments via the Controller CLI or Web Dashboard. Tasks are queued on the Manager and fetched by agents on subsequent check-ins or dispatched over active streams.
4. **Offline Resilient Buffering**: If a router, firewall, or switch drops connectivity, agents buffer up to 50 telemetry samples in RAM and retry with exponential backoff.

---

## 2. Operating Roles & Modes

### `ui` (Web Dashboard & Local Range Host)
Launches the Central Manager HTTPS listener, serves the single-page operations console, and registers a local Host Agent on the host machine.
```powershell
.\bin\rangeforge-ue.exe ui --port 8443
```

### `manager` (Dedicated Central Coordinator)
Runs the coordinator daemon without launching a local host agent. Intended for dedicated server nodes in multi-subnet cyber ranges.
```bash
./bin/rangeforge-ue-linux manager --host 0.0.0.0 --port 8443
```

### `agent` (Target Endpoint Daemon)
Executes natively on Windows, Linux, FreeBSD, or macOS targets. Collects local system telemetry and autonomously generates user activity.
```powershell
.\bin\rangeforge-ue.exe agent --manager https://10.0.0.1:8443 --persona office_worker
```

### `controller` (Operator CLI)
Command-line client for exercise directors to inspect telemetry, list fleet endpoints, and dispatch remote commands.
```bash
./bin/rangeforge-ue-linux controller agents --manager https://10.0.0.1:8443
```

### `standalone` (Evaluation Mode)
Executes full user emulation routines locally in memory without communicating with a central coordinator.
```powershell
.\bin\rangeforge-ue.exe standalone
```

---

## 3. Configuration Reference & Syntax Guide

### A. Manager Server Configuration
**File:** [`configs/manager.yaml`](configs/manager.yaml)

```yaml
# Network interface binding (0.0.0.0 binds all interfaces)
listen_host: "0.0.0.0"

# TCP port for unified HTTPS/HTTP service
listen_port: 8443

# Cryptographic transport encryption
tls_enabled: true
tls_cert_path: "./certs/server.crt"
tls_key_path: "./certs/server.key"

# Heartbeat timeout in seconds before marking agent as offline
heartbeat_timeout_sec: 30

# Default persona profile automatically assigned to new endpoints
default_persona: "office_worker"

# Directory where persona YAML profiles are loaded from
profiles_dir: "./configs/profiles"

# Set to true when evaluating Manager on Windows workstations
allow_unsupported_os: false
```

#### Parameter Breakdown:
| Key | Type | Default | Description |
| :--- | :--- | :--- | :--- |
| `listen_host` | string | `"0.0.0.0"` | Network address to bind. `"0.0.0.0"` binds all interfaces; `"127.0.0.1"` restricts access to local host. |
| `listen_port` | integer | `8443` | Port for HTTPS REST API and Web Console. |
| `tls_enabled` | boolean | `true` | When true, enforces TLS 1.2/1.3 encryption. |
| `tls_cert_path` | string | `"./certs/server.crt"` | Path to X.509 certificate. Auto-generated if missing. |
| `tls_key_path` | string | `"./certs/server.key"` | Path to private key. Auto-generated if missing. |
| `heartbeat_timeout_sec` | integer | `30` | Seconds without an agent check-in before marking offline. |
| `default_persona` | string | `"office_worker"` | Persona assigned to newly registering endpoints. |
| `profiles_dir` | string | `"./configs/profiles"` | Directory containing persona YAML files. |
| `allow_unsupported_os` | boolean | `false` | Permits running Manager on non-Unix nodes for testing. |

---

### B. Host Agent Configurations
**Files:** [`configs/agent_windows.yaml`](configs/agent_windows.yaml) and [`configs/agent_unix.yaml`](configs/agent_unix.yaml)

```yaml
manager_url: "https://10.0.0.1:8443"
heartbeat_sec: 5
bind_interface: ""
persona: "office_worker"
web_corpus_file: "./corpus/web_corpus.json"
creds_file: "./corpus/user_creds.json"
shares_file: ""
enable_web: true
enable_shares: true
enable_host_activity: true
tags:
  - "windows-workstation"
  - "finance-vlan"
  - "cyber-training-range"
```

#### Parameter Breakdown:
| Key | Type | Default | Description |
| :--- | :--- | :--- | :--- |
| `manager_url` | string | `"https://10.0.0.1:8443"` | Outbound HTTPS address of central Manager server. |
| `heartbeat_sec` | integer | `5` | Cadence in seconds between telemetry check-ins. |
| `bind_interface` | string | `""` | Interface to bind (e.g. `"Ethernet"`, `"eth0"`). Leave empty for auto-detection. |
| `persona` | string | `"office_worker"` | Baseline persona profile loaded from Manager. |
| `web_corpus_file` | string | `"./corpus/web_corpus.json"` | JSON file containing URLs and User-Agents. |
| `creds_file` | string | `"./corpus/user_creds.json"` | Mock credentials used in simulated files and logs. |
| `enable_web` | boolean | `true` | Master switch for web browsing emulation engine. |
| `enable_shares` | boolean | `true` | Master switch for SMB/CIFS file share emulation. |
| `enable_host_activity`| boolean | `true` | Master switch for local benign process and file emulation. |
| `tags` | list | `[...]` | String labels attached to telemetry for filtering in API. |

---

### C. Persona Emulation Profiles
**Directory:** [`configs/profiles/`](configs/profiles/)

Personas define the behavior, noise density, and operational focus of endpoints.

#### Full YAML Schema:
```yaml
name: "office_worker"
description: "Knowledge worker persona focused on intranet portals and document reviews"

web_browsing:
  enabled: true
  requests_per_minute_min: 2
  requests_per_minute_max: 6
  dwell_time_min_sec: 15
  dwell_time_max_sec: 45
  fetch_assets: true
  user_agents: []
  target_urls:
    - https://intranet.corp.local
    - http://portal.range.local
  search_keywords:
    - "quarterly expense report template"

file_share:
  enabled: true
  interval_sec: 15
  shares: []
  read_ratio: 0.95
  write_ratio: 0.05

ping:
  enabled: true
  interval_sec: 12
  targets:
    - 127.0.0.1
    - 1.1.1.1

host_activity:
  enabled: true
  interval_sec: 15
  safe_processes:
    - whoami
    - hostname
    - netstat
    - ipconfig
  simulate_office_docs: true
  temp_file_operations: true
```

#### Key Parameter Descriptions:
- **`requests_per_minute_min` / `max`**: Controls request cadence. Low values (2–6) mimic deliberate human browsing; higher values (15–30) mimic automated tools or active researchers.
- **`dwell_time_min_sec` / `max_sec`**: The "think time" between consecutive page navigations. Modeled with Poisson/Gaussian random jitter.
- **`fetch_assets`**: When true, requests static CSS, JS, and image files linked to target pages to create multi-packet bursts on network sensors.
- **`read_ratio` / `write_ratio`**: Decimal percentages (summing to 1.0) governing share interactions. `read_ratio: 0.95` means 95% of share actions are read-only.
- **`safe_processes`**: List of non-destructive diagnostic tools executed without elevation to create realistic EDR/SIEM telemetry.
- **`simulate_office_docs`**: Authors realistic temporary draft files (`.docx`, `.xlsx`, `.pdf`) in user profile folders with automated cleanup.

---

### D. Interactive Network Topology
**File:** [`configs/topology.json`](configs/topology.json)

Defines subnet gateways rendered in the Web UI Interactive SVG Topology Map:

```json
[
  {
    "name": "Edge Gateway",
    "ip": "10.0.0.1",
    "subnet": "10.0.0.0/16",
    "description": "Uplink and Core DNS gateway router"
  },
  {
    "name": "Workstations Gateway",
    "ip": "10.0.10.1",
    "subnet": "10.0.10.0/24",
    "description": "User Workstations Subnet"
  },
  {
    "name": "Core Servers Gateway",
    "ip": "10.0.20.1",
    "subnet": "10.0.20.0/24",
    "description": "Enterprise Core Services & Intranet"
  }
]
```

---

### E. Cyber Range Scenarios
**File:** [`configs/ranges.json`](configs/ranges.json)

Defines high-level scenario parameters, DNS servers, and intensity multipliers:
- **`intensity`**: Global activity multiplier (`"Low (0.5x)"`, `"Standard (1.0x)"`, `"High (2.5x)"`, `"Stress (5.0x)"`).
- **`jitter_pct`**: Randomization percentage applied to dwell times (e.g. `30` applies $\pm 30\%$ variance).
- **`latency_ms`**: Simulated network latency in milliseconds.

---

### F. Administrative Authentication
**File:** [`configs/admin_auth.json`](configs/admin_auth.json)

Stores salted SHA-256 credentials for Web UI and REST API operator access:

```json
{
  "username": "admin",
  "password_hash": "d8af6cb3782ae42b4e4ae03b19ac5776809b29fbd0a3316ea0d75399b24d7e16",
  "salt": "ebad3cddb04d1052ab0e66937c9c6031",
  "updated_at": "2026-09-07T18:30:00Z"
}
```

- **Default Credentials**: `admin` / `rangeforge`
- **Hash Algorithm**: `hex(SHA-256(salt + ":" + password))`
- Credentials can be changed dynamically via the Web UI Settings modal or by updating this file.

---

### G. Web & Identity Corpus
**Directory:** [`corpus/`](corpus/)

- **`web_corpus.json`**: Reference URLs, intranet portals, search phrases, and browser User-Agents.
- **`user_creds.json`**: Simulated user accounts, domains, roles, and default persona assignments.
- **`wordlist.txt`**: Scenario dictionary used by the document authoring engine to name temporary files.

---

## 4. Multi-Network Deployment & Firewall Traversal

RangeForge is optimized for challenging network topologies:

### A. pfSense & Router Traversal
- **Outbound-Only HTTPS**: Host Agents require **zero inbound listening ports**. Communication flows outbound to the Manager on port 8443.
- Firewalls only require an outbound stateful filter rule allowing HTTPS from endpoint subnets to the coordinator.

### B. Enterprise Proxies
- Respects standard `HTTP_PROXY`, `HTTPS_PROXY`, and `NO_PROXY` environment variables automatically via Go's `http.ProxyFromEnvironment`.
- Range-internal subnets (e.g. `10.0.0.0/8`, `*.range.local`) can be exempted via `NO_PROXY`.

### C. Air-Gapped & Offline Networks
- Zero external dependencies. All assets, web console logic, and synthetic corpora are bundled within the binary.
- Requires no internet connectivity.

---

## 5. Emulation Engines & Realism Calibration

### Blue Team Noise Calibration Matrix
Adjust execution noise in persona YAML profiles to match training objectives:

| Noise Level | RPM | Dwell Time | Process Interval | Training Scenario |
| :--- | :--- | :--- | :--- | :--- |
| **Silent** | 0–1 | 60–120s | Disabled | Stealth evaluation; test if red team is noticed without any background cover |
| **Low** | 2–5 | 20–60s | 60s | Baseline corporate office; modest background noise |
| **Standard** | 5–15 | 10–30s | 15s | Standard enterprise workday; realistic SIEM alert volume |
| **High** | 15–30 | 4–15s | 5s | High-stress environment; heavy noise masking offensive operations |

---

## 6. Interactive Web UI Dashboard

Access the operations console at: **`https://127.0.0.1:8443/`**

### Features:
- **Real-Time Telemetry**: Live SVG sparklines plotting Fleet Average CPU and RAM with 15-point historical smoothing.
- **Fleet Roster**: 8-column host table displaying OS version, network IPs, assigned persona, and live telemetry.
- **Interactive SVG Topology Map**: Visual representation of subnets, edge routers, and connected agents.
- **Emulation Studio**: Live tuner for Web RPM, dwell jitter, SMB read/write ratios, and safe host processes.
- **Host Inspector Modal**: Deep-dive inspection of CPU cores, memory usage, MAC address, MTU, and installed tools.
- **Integrated Command Console**: Dispatch PowerShell or Bash commands to any endpoint directly from the browser.
- **Dual Themes**: Toggle between **Ocean Sapphire** (cobalt deep blue cyber ops console) and **Carbon Black** (stealth dark operations with slate borders). Preferences persist in `localStorage`.

---

## 7. Controller CLI & REST API Reference

### Controller Commands
```bash
# List all active fleet endpoints
rangeforge-ue controller agents --manager https://10.0.0.1:8443

# Inspect host telemetry
rangeforge-ue controller stats --manager https://10.0.0.1:8443 --agent <agent-id>

# Dispatch remote command
rangeforge-ue controller exec --manager https://10.0.0.1:8443 --agent <agent-id> --cmd "whoami /all"

# Reassign persona live
rangeforge-ue controller persona --manager https://10.0.0.1:8443 --agent <agent-id> --set-persona sysadmin
```

### Core REST API Endpoints
| Method | Path | Description |
| :--- | :--- | :--- |
| `POST` | `/api/v1/agent/register` | Endpoint agent registration |
| `POST` | `/api/v1/agent/heartbeat` | Ingest periodic telemetry |
| `GET` | `/api/v1/controller/agents` | Retrieve fleet inventory |
| `GET` | `/api/v1/controller/telemetry` | Retrieve real-time CPU/RAM metrics |
| `POST` | `/api/v1/controller/command` | Queue remote command for execution |
| `POST` | `/api/v1/controller/persona` | Dynamically update assigned persona |
| `GET` | `/api/v1/controller/topology` | Fetch current network topology |
| `POST` | `/api/v1/auth/login` | Authenticate administrative operator |

---

## 8. Operational Security & Zero-Artifact Guarantees

RangeForge is engineered with strict operational security guardrails:

1. **Zero Persistent Artifacts**: All temporary files created during emulation (drafts, locks, logs) are tracked in memory. When emulation halts or the agent process exits, all synthetic files are automatically deleted.
2. **Stealth Command Execution**: Remote PowerShell commands on Windows are executed with `-NoProfile` and history suppression. Commands are never appended to user shell history files (`ConsoleHost_history.txt`).
3. **Privilege-Free Operation**: Host Agents do not require `Administrator` or `root` privileges. They operate cleanly within standard unprivileged user security contexts.

---

## 9. Troubleshooting & Frequently Asked Questions

### Q: Why do I see a TLS certificate warning in my browser?
**A:** RangeForge generates a high-security self-signed certificate on first startup. Click **Advanced -> Proceed** in your browser, or install your organization's CA certificate into `certs/server.crt` and `certs/server.key`.

### Q: Can I run RangeForge over plain HTTP?
**A:** Yes. The unified listener multiplexes TLS and HTTP on the same port (8443). You can browse to `http://127.0.0.1:8443/` directly.

### Q: How do I add custom browsing URLs?
**A:** Add target URLs to `configs/profiles/<persona>.yaml` under `target_urls`, or update `corpus/web_corpus.json`.

### Q: How do I change the admin password?
**A:** Click the **Admin** indicator in the Web UI navigation bar to open the **Change Credentials** dialog, or generate a new salted SHA-256 hash in `configs/admin_auth.json`.
