# RangeForge User Emulation Suite (`rangeforge-ue`)
## Operator, Architecture & Configuration Reference Manual

**Version:** `2.0.0-oss`  
**License:** Apache License 2.0  
**Target Audience:** Cyber Range Architects, SOC Analysts, Detection Engineers, Exercise Directors, Red/Blue/Purple Teams

---

## Table of Contents
1. [Executive Summary & Core Principles](#1-executive-summary--core-principles)
2. [System Architecture & Wire Protocol](#2-system-architecture--wire-protocol)
   - [2.1 High-Level Architecture](#21-high-level-architecture)
   - [2.2 Connection Handshake & Lifecycle Protocol](#22-connection-handshake--lifecycle-protocol)
   - [2.3 Dual-Protocol TLS/HTTP Multiplexing](#23-dual-protocol-tlshttp-multiplexing)
   - [2.4 Offline Fault-Tolerance & Buffer Queue](#24-offline-fault-tolerance--buffer-queue)
3. [Operational Roles & CLI Command Reference](#3-operational-roles--cli-command-reference)
   - [3.1 Web UI Mode (`ui`)](#31-web-ui-mode-ui)
   - [3.2 Dedicated Coordinator Mode (`manager`)](#32-dedicated-coordinator-mode-manager)
   - [3.3 Target Endpoint Daemon (`agent`)](#33-target-endpoint-daemon-agent)
   - [3.4 Operator CLI Client (`controller`)](#34-operator-cli-client-controller)
   - [3.5 Self-Contained Evaluation (`standalone`)](#35-self-contained-evaluation-standalone)
4. [Configuration Files & Syntax Specification](#4-configuration-files--syntax-specification)
   - [4.1 Manager Configuration (`configs/manager.yaml`)](#41-manager-configuration-configsmanageryaml)
   - [4.2 Windows Host Agent (`configs/agent_windows.yaml`)](#42-windows-host-agent-configsagent_windowsyaml)
   - [4.3 Unix Host Agent (`configs/agent_unix.yaml`)](#43-unix-host-agent-configsagent_unixyaml)
   - [4.4 Persona Emulation Profiles (`configs/profiles/*.yaml`)](#44-persona-emulation-profiles-configsprofilesyaml)
   - [4.5 Network Topology Map (`configs/topology.json`)](#45-network-topology-map-configstopologyjson)
   - [4.6 Cyber Range Scenarios (`configs/ranges.json`)](#46-cyber-range-scenarios-configsrangesjson)
   - [4.7 Operator Authentication (`configs/admin_auth.json`)](#47-operator-authentication-configsadmin_authjson)
   - [4.8 Emulation Corpora (`corpus/`)](#48-emulation-corpora-corpus)
5. [Network Topologies & Firewall Traversal](#5-network-topologies--firewall-traversal)
   - [5.1 Single-Interface Edge Traversal (pfSense / OPNsense)](#51-single-interface-edge-traversal-pfsense--opnsense)
   - [5.2 Outbound-Only State Engine & NAT Boundaries](#52-outbound-only-state-engine--nat-boundaries)
   - [5.3 Enterprise Proxies & PAC Configurations](#53-enterprise-proxies--pac-configurations)
   - [5.4 Air-Gapped & Isolated Range Operations](#54-air-gapped--isolated-range-operations)
6. [Emulation Engines & Behavioral Realism](#6-emulation-engines--behavioral-realism)
   - [6.1 Web Browsing Engine (Gaussian Jitter & Asset Bursts)](#61-web-browsing-engine)
   - [6.2 SMB / CIFS / NFS File Share Engine](#62-smb--cifs--nfs-file-share-engine)
   - [6.3 Benign Host Activity & Safe Process Execution](#63-benign-host-activity--safe-process-execution)
   - [6.4 Remote Command Execution Engine](#64-remote-command-execution-engine)
   - [6.5 Hardware & OS Telemetry Subsystem](#65-hardware--os-telemetry-subsystem)
7. [Detection Engineering & SIEM / EDR Noise Calibration](#7-detection-engineering--siem--edr-noise-calibration)
   - [7.1 Windows Event Log Telemetry Mapping](#71-windows-event-log-telemetry-mapping)
   - [7.2 Linux Auditd & eBPF Event Generation](#72-linux-auditd--ebpf-event-generation)
   - [7.3 Noise Density Matrix (Silent to Stress)](#73-noise-density-matrix-silent-to-stress)
8. [Interactive Web Operations Console](#8-interactive-web-operations-console)
   - [8.1 Real-Time SVG Sparklines & Smoothing](#81-real-time-svg-sparklines--smoothing)
   - [8.2 Fleet Roster & Host Inspector Modal](#82-fleet-roster--host-inspector-modal)
   - [8.3 Interactive SVG Network Topology Visualizer](#83-interactive-svg-network-topology-visualizer)
   - [8.4 Dual Cyber-Ops Theme System](#84-dual-cyber-ops-theme-system)
9. [Comprehensive REST API Reference](#9-comprehensive-rest-api-reference)
10. [Operational Security & Anti-Forensics Guarantees](#10-operational-security--anti-forensics-guarantees)
11. [Troubleshooting & Operator Diagnostics Runbook](#11-troubleshooting--operator-diagnostics-runbook)

---

## 1. Executive Summary & Core Principles

The **RangeForge User Emulation Suite (`rangeforge-ue`)** addresses the most pervasive weakness in cyber defense exercises: **the absence of realistic background network and host activity**. When a training range is completely idle, detecting unauthorized activity requires little skill; SOC analysts simply look for the only IP address communicating across the subnet.

RangeForge generates authentic, continuous background activity that mirrors human users across enterprise workstations, servers, and routers:
- **Zero External Dependencies**: Built entirely with Go's standard library and lightweight YAML parsing. Requires no Python, Node.js, external database, or third-party packages.
- **Outbound-Only HTTPS**: Host agents initiate all connections over outbound HTTPS (port 8443). They open **zero inbound listening ports**, traversing stateful firewalls, pfSense routers, and NAT gateways effortlessly.
- **Zero Persistent Footprint**: Temporary files generated to simulate user file share activity are tracked in memory and automatically unlinked upon agent shutdown.
- **Cross-Platform Uniformity**: Identical configuration structures, persona definitions, and telemetry contracts run seamlessly on Windows, Linux, FreeBSD, and macOS.

---

## 2. System Architecture & Wire Protocol

### 2.1 High-Level Architecture

RangeForge is organized into three operational tiers:

```
                            ┌────────────────────────────────────────┐
                            │        OPERATOR / CONTROLLER           │
                            │      - CLI Management Console          │
                            │      - REST Automation Scripts         │
                            └───────────────────┬────────────────────┘
                                                │ REST API / CLI
                                                ▼
                            ┌────────────────────────────────────────┐
                            │          RANGEFORGE MANAGER            │
                            │   - Port 8443 HTTPS Listener           │
                            │   - Agent Registry & Heartbeat Monitor │
                            │   - Interactive Web UI Operations      │
                            │   - Remote Command Queue               │
                            └───────────────────┬────────────────────┘
                                                │ HTTPS POST (Port 8443)
                                                │ Outbound From Endpoints
                     ┌──────────────────────────┴──────────────────────────┐
                     │                                                     │
                     ▼                                                     ▼
     ┌───────────────────────────────┐                     ┌───────────────────────────────┐
     │      HOST AGENT: WINDOWS      │                     │       HOST AGENT: UNIX        │
     │  - Windows 10/11/Server       │                     │  - Linux (Ubuntu/RHEL/Debian) │
     │  - WMI / Native Win32 Stats   │                     │  - FreeBSD 14 / pfSense 2.7+  │
     │  - SMB / Web / Process Noise  │                     │  - /proc, sysctl, Net Stats   │
     └───────────────────────────────┘                     └───────────────────────────────┘
```

### 2.2 Connection Handshake & Lifecycle Protocol

All communications between Host Agents and the Manager occur over outbound HTTPS:

```
Agent Endpoint                                                      Central Manager
      │                                                                    │
      ├────────────────── 1. POST /api/v1/agent/register ─────────────────►│
      │   Payload: Hostname, OS, Arch, IPs, MAC, Tags, Tools, Persona      │
      │                                                                    │
      │◄───────────────── 2. HTTP 200 OK (Registration Ack) ───────────────┤
      │   Response: AgentID, HeartbeatIntervalSec, ActivePersonaConfig     │
      │                                                                    │
      │                                                                    │
      ├──── 3. POST /api/v1/agent/heartbeat (every HeartbeatIntervalSec) ──►│
      │   Payload: AgentID, CPU%, RAM%, Disk%, NetworkStats, Status        │
      │                                                                    │
      │◄─── 4. HTTP 200 OK (Heartbeat Ack + Pending Directives) ───────────┤
      │   Response: PersonaUpdates, QueuedCommands, ConfigReloadSignals    │
      │                                                                    │
      ▼                                                                    ▼
```

1. **Initial Registration**: When launched, the Agent queries its local hardware, discovers functional command-line tools (`ping`, `git`, `powershell`, `netstat`, `curl`), and sends a registration request.
2. **Registration Acknowledgment**: The Manager assigns a unique `agent_id` (or recognizes an existing persistent ID) and returns the assigned persona parameters.
3. **Periodic Telemetry Beacon**: Every $N$ seconds (default: 5s), the Agent posts real-time CPU, RAM, disk, and network stats.
4. **Command & Directive Piggybacking**: Rather than maintaining separate polling channels, pending administrative commands and persona transitions are piggybacked onto the heartbeat response, minimizing network footprint.

### 2.3 Dual-Protocol TLS/HTTP Multiplexing

To prevent operator lockouts caused by strict browser certificate warnings or scripts lacking TLS flags, RangeForge implements a custom connection listener (`pkg/tlsutil/`):
- When a connection arrives on port 8443, the listener inspects the first byte without consuming it (`PEEK 1 byte`).
- If the first byte is `0x16` (TLS Handshake Record), the connection is passed to the Go standard library `crypto/tls` engine.
- If the first byte corresponds to an ASCII HTTP method (`G` for `GET`, `P` for `POST`, `H` for `HEAD`), the connection is processed directly as plaintext HTTP.
- Both secure HTTPS and plaintext HTTP operate seamlessly on the exact same port (8443).

### 2.4 Offline Fault-Tolerance & Buffer Queue

If an edge firewall, router reboot, or network partition temporarily severs communication between an Agent and the Manager:
- The Agent does **not** terminate or crash.
- Telemetry samples are enqueued in an in-memory ring buffer (up to 50 samples).
- When connectivity is restored, buffered telemetry is flushed to the Manager in chronological order.
- Background user emulation (browsing, file operations, ping audits) continues running uninterrupted based on the last known persona profile.

---

## 3. Operational Roles & CLI Command Reference

The `rangeforge-ue` binary supports five primary operational modes:

### 3.1 Web UI Mode (`ui`)

Launches the Central Manager HTTPS server, serves the single-page web console, and automatically starts an internal Host Agent representing the local machine.

```powershell
# Windows
.\bin\rangeforge-ue.exe ui --host 0.0.0.0 --port 8443
```

```bash
# Linux / Unix
./bin/rangeforge-ue-linux ui --host 0.0.0.0 --port 8443
```

#### Command Flags:
| Flag | Type | Default | Description |
| :--- | :--- | :--- | :--- |
| `--host` | string | `"0.0.0.0"` | Network interface address to bind. |
| `--port` | integer | `8443` | TCP port for HTTPS REST API and Web Console. |
| `--config` | string | `"./configs/manager.yaml"` | Path to Manager YAML configuration file. |
| `--cert` | string | `"./certs/server.crt"` | Path to TLS certificate (auto-generated if missing). |
| `--key` | string | `"./certs/server.key"` | Path to TLS private key (auto-generated if missing). |

---

### 3.2 Dedicated Coordinator Mode (`manager`)

Launches the Central Manager service without registering a local host agent. Recommended for dedicated orchestrator nodes in production cyber ranges.

```bash
./bin/rangeforge-ue-linux manager --host 0.0.0.0 --port 8443 --config ./configs/manager.yaml
```

#### Command Flags:
| Flag | Type | Default | Description |
| :--- | :--- | :--- | :--- |
| `--host` | string | `"0.0.0.0"` | Network interface address to bind. |
| `--port` | integer | `8443` | TCP port for HTTPS REST API and Web Console. |
| `--config` | string | `"./configs/manager.yaml"` | Path to Manager YAML configuration file. |

---

### 3.3 Target Endpoint Daemon (`agent`)

Runs the user emulation daemon on a target endpoint. Collects system metrics and generates background traffic.

```powershell
# Windows Host
.\bin\rangeforge-ue.exe agent --manager https://10.0.0.1:8443 --persona office_worker --config ./configs/agent_windows.yaml
```

```bash
# Linux Host
./bin/rangeforge-ue-linux agent --manager https://10.0.0.1:8443 --persona developer --config ./configs/agent_unix.yaml
```

```bash
# FreeBSD / pfSense Gateway
./bin/rangeforge-ue-freebsd agent --manager https://10.0.0.1:8443 --persona sysadmin --config ./configs/agent_unix.yaml
```

#### Command Flags:
| Flag | Type | Default | Description |
| :--- | :--- | :--- | :--- |
| `--manager` | string | `"https://127.0.0.1:8443"` | URL of the central RangeForge Manager server. |
| `--persona` | string | `"office_worker"` | Baseline persona profile (`office_worker`, `developer`, `sysadmin`, `executive`). |
| `--config` | string | `""` | Path to Agent YAML configuration file (overrides CLI defaults). |
| `--tag` | string | `""` | Optional metadata tag to attach to endpoint (e.g. `subnet-finance`). |

---

### 3.4 Operator CLI Client (`controller`)

Allows exercise directors and instructors to interact with the fleet via command line.

#### Subcommands:
```bash
# 1. List all active registered endpoints
rangeforge-ue controller agents --manager https://10.0.0.1:8443

# 2. Inspect real-time metrics for a specific endpoint
rangeforge-ue controller stats --manager https://10.0.0.1:8443 --agent <agent-id>

# 3. Dispatch an ad-hoc shell command to an endpoint
rangeforge-ue controller exec --manager https://10.0.0.1:8443 --agent <agent-id> --cmd "whoami /priv"

# 4. Dynamically reassign an endpoint's persona
rangeforge-ue controller persona --manager https://10.0.0.1:8443 --agent <agent-id> --set-persona sysadmin
```

---

### 3.5 Self-Contained Evaluation (`standalone`)

Runs user emulation directly in the current terminal without communicating with any central coordinator. Ideal for testing emulation engines or generating local noise.

```powershell
.\bin\rangeforge-ue.exe standalone --persona office_worker
```

---

## 4. Configuration Files & Syntax Specification

Every configuration file in RangeForge uses clean YAML or JSON with complete syntax validation.

### 4.1 Manager Configuration (`configs/manager.yaml`)

Defines network binding, TLS parameters, heartbeat timeouts, and persona directories.

```yaml
listen_host: "0.0.0.0"
listen_port: 8443
tls_enabled: true
tls_cert_path: "./certs/server.crt"
tls_key_path: "./certs/server.key"
heartbeat_timeout_sec: 30
default_persona: "office_worker"
profiles_dir: "./configs/profiles"
allow_unsupported_os: false
```

#### Syntax Reference Table:
| Parameter | Type | Req | Default | Valid Values | Description |
| :--- | :--- | :--- | :--- | :--- | :--- |
| `listen_host` | string | Yes | `"0.0.0.0"` | IP address | Interface address to bind. `"0.0.0.0"` binds all adapters; `"127.0.0.1"` restricts to localhost. |
| `listen_port` | integer| Yes | `8443` | `1024–65535` | TCP port for HTTPS REST API and Web Console. |
| `tls_enabled` | bool | Yes | `true` | `true`, `false` | Enables TLS encryption. When false, runs plaintext HTTP only. |
| `tls_cert_path`| string | No | `"./certs/server.crt"` | File path | Path to PEM-encoded X.509 certificate. Auto-generated if absent. |
| `tls_key_path` | string | No | `"./certs/server.key"` | File path | Path to PEM-encoded private key. Auto-generated if absent. |
| `heartbeat_timeout_sec` | integer | Yes | `30` | `10–300` | Seconds without an agent check-in before marking status as offline. |
| `default_persona` | string | Yes | `"office_worker"` | Persona name | Profile assigned to new endpoints that register without specifying a persona. |
| `profiles_dir` | string | Yes | `"./configs/profiles"` | Directory path | Directory containing persona YAML definitions. |
| `allow_unsupported_os` | bool | No | `false` | `true`, `false` | Permits running Manager on non-Linux OS without warnings. |

---

### 4.2 Windows Host Agent (`configs/agent_windows.yaml`)

Configures an agent executing on a Windows workstation or server:

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

#### Syntax Reference Table:
| Parameter | Type | Req | Default | Valid Values | Description |
| :--- | :--- | :--- | :--- | :--- | :--- |
| `manager_url` | string | Yes | `"https://10.0.0.1:8443"` | Valid URL | Address of central coordinator. |
| `heartbeat_sec`| integer| Yes | `5` | `1–60` | Telemetry transmission interval in seconds. |
| `bind_interface` | string | No | `""` | Interface name | Specific adapter name (e.g. `"Ethernet 2"`). Empty string auto-detects. |
| `persona` | string | Yes | `"office_worker"` | Persona name | Baseline persona to initialize upon startup. |
| `web_corpus_file` | string | Yes | `"./corpus/web_corpus.json"` | File path | JSON file containing target URLs and User-Agents. |
| `creds_file` | string | Yes | `"./corpus/user_creds.json"` | File path | Mock credential database for simulated activity. |
| `shares_file` | string | No | `""` | File path | Path to custom network share targets list. |
| `enable_web` | bool | Yes | `true` | `true`, `false` | Master toggle for HTTP/S web browsing engine. |
| `enable_shares`| bool | Yes | `true` | `true`, `false` | Master toggle for SMB/CIFS network share engine. |
| `enable_host_activity` | bool | Yes | `true` | `true`, `false` | Master toggle for benign diagnostic processes. |
| `tags` | list | No | `[...]` | String array | Descriptive tags for filtering and topology mapping. |

---

### 4.3 Unix Host Agent (`configs/agent_unix.yaml`)

Configures an agent executing on Linux, FreeBSD (pfSense), or macOS:

```yaml
manager_url: "https://10.0.0.1:8443"
heartbeat_sec: 5
bind_interface: ""
persona: "developer"
web_corpus_file: "./corpus/web_corpus.json"
creds_file: "./corpus/user_creds.json"
shares_file: ""
enable_web: true
enable_shares: true
enable_host_activity: true
tags:
  - "linux-server"
  - "development-vlan"
```

---

### 4.4 Persona Emulation Profiles (`configs/profiles/*.yaml`)

Persona profiles model authentic user roles with distinct operational cadences.

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
    - "annual compliance training login"

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

#### Detailed Parameter Breakdown:
- **`web_browsing.requests_per_minute_min` / `max`**: Integer range governing how many pages the agent navigates per minute. 2–6 represents calm human reading; 15–30 represents active research.
- **`web_browsing.dwell_time_min_sec` / `max_sec`**: The "think time" between consecutive page clicks. Calculated with random Gaussian jitter to avoid periodic spikes.
- **`web_browsing.fetch_assets`**: When true, parses HTML bodies and initiates parallel GET requests for embedded `.css`, `.js`, and `.png` assets, producing authentic multi-packet flows on network sensors.
- **`file_share.read_ratio` / `write_ratio`**: Floats summing to 1.0. A `read_ratio` of `0.95` ensures that 95% of share operations are benign file reads, preventing disk saturation.
- **`host_activity.safe_processes`**: Array of non-destructive diagnostic binaries executed without elevation to generate realistic process creation telemetry.
- **`host_activity.simulate_office_docs`**: Authors temporary draft files (`.docx`, `.xlsx`, `.pdf`) in user temporary folders and purges them upon completion.

---

### 4.5 Network Topology Map (`configs/topology.json`)

Defines cyber range subnets and gateway routers displayed in the Web UI Interactive SVG Topology Map:

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

### 4.6 Cyber Range Scenarios (`configs/ranges.json`)

Defines exercise intensity multipliers and network latency characteristics:

```json
{
  "range_id": "enterprise-cyber-range-01",
  "name": "Corporate Enterprise Range",
  "intensity": "Standard (1.0x)",
  "jitter_pct": 30,
  "latency_ms": 15,
  "dns_servers": ["10.0.0.1", "1.1.1.1"]
}
```

- **`intensity`**: Global activity multiplier (`"Low (0.5x)"`, `"Standard (1.0x)"`, `"High (2.5x)"`, `"Stress (5.0x)"`).
- **`jitter_pct`**: Randomization percentage applied to dwell times (e.g. `30` applies $\pm 30\%$ variance).
- **`latency_ms`**: Simulated artificial network latency in milliseconds.

---

### 4.7 Operator Authentication (`configs/admin_auth.json`)

Stores salted cryptographic credentials for administrative operator access:

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
- To generate a new hash via PowerShell:
  ```powershell
  $salt = "ebad3cddb04d1052ab0e66937c9c6031"
  $pass = "NewSecurePassword"
  $bytes = [System.Text.Encoding]::UTF8.GetBytes("$salt`:$pass")
  $hash = [System.Security.Cryptography.SHA256]::Create().ComputeHash($bytes)
  [BitConverter]::ToString($hash).Replace("-","").ToLower()
  ```

---

### 4.8 Emulation Corpora (`corpus/`)

- **`corpus/web_corpus.json`**: List of target domains, intranet portals, simulated search engines, and modern browser User-Agent strings.
- **`corpus/user_creds.json`**: Synthetic enterprise accounts, department affiliations, email addresses, and security roles.
- **`corpus/wordlist.txt`**: Dictionary of business and technical terminology used by the document authoring engine to generate realistic file names (e.g. `Q3_Financial_Review_Draft.docx`).

---

## 5. Network Topologies & Firewall Traversal

### 5.1 Single-Interface Edge Traversal (pfSense / OPNsense)

RangeForge is uniquely suited for multi-tier cyber ranges partitioned by pfSense or OPNsense firewalls:

```
[ Workstations Subnet: 10.0.10.0/24 ]
           │
           │ Outbound HTTPS (Port 8443)
           ▼
┌─────────────────────────────────────────┐
│     pfSense Firewall / Router Gateway   │
│     WAN: 192.168.1.1  LAN: 10.0.10.1    │
│     Rule: Allow LAN -> Manager:8443     │
└────────────────────┬────────────────────┘
                     │
                     │ Outbound NAT / Routed
                     ▼
[ Central Management Subnet: 10.0.0.0/24 ]
           │
           ▼
[ RangeForge Manager: 10.0.0.5:8443 ]
```

#### Firewall Rule Requirement:
Only **one outbound rule** is required on the edge router:
- **Action**: Pass
- **Interface**: LAN
- **Address Family**: IPv4
- **Protocol**: TCP
- **Source**: `LAN net` (or target subnet)
- **Destination**: `Manager_IP` (Port `8443`)

Because connections are stateful and outbound-only, no inbound port forwarding (NAT) or pinholes into the endpoint subnets are ever needed.

### 5.2 Outbound-Only State Engine & NAT Boundaries

When agents reside behind NAT gateways, their source IP address is translated. RangeForge automatically extracts both:
1. **Reported Internal IP**: Collected directly from endpoint adapters via `net.Interfaces()`.
2. **Observed Remote IP**: Extracted from the incoming TCP socket (`req.RemoteAddr`).
Both addresses are stored in the Agent Registry and displayed in the Web UI Fleet Roster.

### 5.3 Enterprise Proxies & PAC Configurations

Host Agents automatically honor system proxy settings via standard environment variables:
```bash
export HTTP_PROXY="http://proxy.corp.local:8080"
export HTTPS_PROXY="http://proxy.corp.local:8080"
export NO_PROXY="localhost,127.0.0.1,10.0.0.0/8"
```
Range internal subnets can be exempted from proxy inspection by listing them in `NO_PROXY`.

### 5.4 Air-Gapped & Isolated Range Operations

RangeForge is completely self-contained:
- Zero CDN dependencies: All JavaScript, CSS styling, and SVG assets are bundled directly inside the Go binary.
- Requires no internet access or external package registries.
- Can be deployed onto completely isolated, air-gapped networks via USB or offline disk image.

---

## 6. Emulation Engines & Behavioral Realism

### 6.1 Web Browsing Engine

The web emulation engine generates human-like HTTP/HTTPS traffic:
- **Poisson Dwell Times**: Dwell time between requests is randomized around the configured mean using a Poisson distribution.
- **Modern User-Agents**: Rotates real-world browser headers:
  - `Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/124.0.0.0 Safari/537.36`
  - `Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/17.4 Safari/605.1.15`
  - `Mozilla/5.0 (X11; Linux x86_64; rv:125.0) Gecko/20100101 Firefox/125.0`
- **Asset Waterfall**: Upon loading an HTML page, the engine parses asset tags (`<link rel="stylesheet">`, `<script src="...">`, `<img src="...">`) and issues parallel HTTP GET requests to simulate full page rendering.

### 6.2 SMB / CIFS / NFS File Share Engine

Generates authentic file server traffic:
- Connects to designated network shares (or local test mounts).
- Reads existing files according to `read_ratio`.
- Periodically authors new office document drafts (`.docx`, `.xlsx`, `.txt`) using terms from `corpus/wordlist.txt`.
- All written files are recorded in an in-memory tracking structure and deleted upon shutdown.

### 6.3 Benign Host Activity & Safe Process Execution

Executes realistic diagnostic and administrative tools without elevation:
- **Windows**: Executes `whoami.exe`, `hostname.exe`, `netstat.exe -an`, `ipconfig.exe /all`, `systeminfo.exe`.
- **Unix**: Executes `whoami`, `uname -a`, `netstat -tuln`, `df -h`, `uptime`, `git status`.
- Execution cadences are staggered to avoid periodic spikes on host CPU and disk.

### 6.4 Remote Command Execution Engine

Enables cyber range operators to trigger ad-hoc activities on specific endpoints:
- **Windows**: Executes via `powershell.exe -ExecutionPolicy Bypass -NoProfile -NonInteractive -Command "<cmd>"`.
- **History Suppression**: Execution bypasses `PSReadLine` history to avoid altering operator history logs.
- **Output Capture**: Stdout, Stderr, exit code, and execution duration in milliseconds are returned directly to the Manager.

### 6.5 Hardware & OS Telemetry Subsystem

Collects low-overhead system metrics every heartbeat cycle:
- **CPU Utilization**: Derived from CPU tick counters across all cores.
- **Memory Consumption**: Total RAM, Available RAM, and percentage utilized.
- **Disk Storage**: Free space and usage percentage of the system drive.
- **Network Interface Counters**: Bytes sent and received per second.
- **Installed Tools Inventory**: Discovers availability of CLI tools (`powershell`, `cmd`, `bash`, `ping`, `curl`, `git`, `netstat`, `nmap`).

---

## 7. Detection Engineering & SIEM / EDR Noise Calibration

### 7.1 Windows Event Log Telemetry Mapping

RangeForge actions generate standard Windows Event Logs and Sysmon records, allowing blue teams to test correlation rules against authentic baseline noise:

| Emulation Action | Windows Security Event ID | Sysmon Event ID | Target Log Provider |
| :--- | :--- | :--- | :--- |
| Benign Process Execution (`whoami`, `hostname`) | `4688` (Process Creation) | `1` (Process Create) | `Microsoft-Windows-Sysmon/Operational` |
| Web Browsing / HTTP GET Requests | — | `3` (Network Connection) | `Microsoft-Windows-Sysmon/Operational` |
| Temporary Office Document Authoring | `4663` (File System Access) | `11` (File Create) | `Microsoft-Windows-Sysmon/Operational` |
| Intranet Domain Resolution | — | `22` (DNS Query) | `Microsoft-Windows-Sysmon/Operational` |
| SMB Share Connection (`net use`) | `5140` (Share Accessed) | `3` (Network Connection) | `Security` |

### 7.2 Linux Auditd & eBPF Event Generation

On Linux endpoints, RangeForge generates standard system calls captured by `auditd`, `Falco`, or eBPF agents:
- `execve` / `execveat`: Triggered during diagnostic command execution.
- `connect` / `socket`: Triggered during web browsing and ping emulation.
- `openat` / `write`: Triggered during file share operations and temporary document creation.

### 7.3 Noise Density Matrix (Silent to Stress)

Blue teams can tune the noise level to calibrate SIEM alerting thresholds:

| Level | Intensity Multiplier | Web RPM | Dwell Jitter | Process Interval | Operational Objective |
| :--- | :--- | :--- | :--- | :--- | :--- |
| **Silent** | `0.0x` | 0 | None | Disabled | Pure stealth evaluation; verify if red team activity is spotted with zero background noise. |
| **Low** | `0.5x` | 1–3 | 30–90s | 60s | Small office / branch environment; modest baseline noise. |
| **Standard** | `1.0x` | 4–10 | 10–30s | 15s | Typical enterprise workday; authentic alert volume. |
| **High** | `2.5x` | 15–30 | 5–15s | 5s | High-density trading floor or dev shop; heavy traffic masking attacks. |
| **Stress** | `5.0x` | 30–60 | 1–5s | 2s | SOC stress test; evaluates SIEM pipeline throughput under load. |

---

## 8. Interactive Web Operations Console

The built-in web operations console is served directly from the Manager on port 8443:

### 8.1 Real-Time SVG Sparklines & Smoothing
- Displays real-time Fleet Average CPU and RAM usage.
- Uses 15-point historical smoothing to render clean, responsive SVG sparklines without client-side charting libraries.

### 8.2 Fleet Roster & Host Inspector Modal
- 8-column tabular view displaying Hostname, Operating System, Internal IP, External IP, Assigned Persona, Status, CPU %, and RAM %.
- Clicking any host opens the **Host Inspector Modal**, revealing:
  - Hardware specifications (Cores, Total RAM, Free Disk).
  - Network details (MAC address, MTU, Subnet).
  - Detected binary inventory (`git`, `ping`, `curl`, `powershell`, `bash`).
  - Raw JSON telemetry payload.

### 8.3 Interactive SVG Network Topology Visualizer
- Renders an interactive network graph mapping active endpoints to their respective gateway routers and subnets based on `configs/topology.json`.

### 8.4 Dual Cyber-Ops Theme System
- **Ocean Sapphire**: Cobalt deep blue styling tailored for modern cyber operations centers.
- **Carbon Black**: High-contrast slate and black theme designed for low-light control rooms.
- User theme selection is preserved across browser sessions via `localStorage`.

---

## 9. Comprehensive REST API Reference

The Manager exposes a full REST API for automation and integration with external cyber range orchestration frameworks:

### 9.1 Authentication (`POST /api/v1/auth/login`)
Authenticates an operator session and issues an HTTP cookie.

```bash
curl -k -X POST https://127.0.0.1:8443/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"username": "admin", "password": "rangeforge"}'
```

**Response (HTTP 200 OK):**
```json
{
  "status": "success",
  "message": "Authenticated successfully",
  "username": "admin"
}
```

---

### 9.2 Agent Registration (`POST /api/v1/agent/register`)
Invoked by host agents during startup.

```bash
curl -k -X POST https://127.0.0.1:8443/api/v1/agent/register \
  -H "Content-Type: application/json" \
  -d '{
    "hostname": "win11-finance-01",
    "os": "windows",
    "arch": "amd64",
    "ips": ["10.0.10.45"],
    "mac": "00:50:56:B2:1A:09",
    "tags": ["workstation", "finance"],
    "persona": "office_worker",
    "tools": ["ping", "whoami", "powershell", "netstat"]
  }'
```

**Response (HTTP 200 OK):**
```json
{
  "status": "registered",
  "agent_id": "agent-win11-finance-01-9a4f",
  "heartbeat_interval_sec": 5,
  "assigned_persona": "office_worker"
}
```

---

### 9.3 Ingest Telemetry Heartbeat (`POST /api/v1/agent/heartbeat`)
Transmits periodic hardware and network metrics.

```bash
curl -k -X POST https://127.0.0.1:8443/api/v1/agent/heartbeat \
  -H "Content-Type: application/json" \
  -d '{
    "agent_id": "agent-win11-finance-01-9a4f",
    "cpu_pct": 14.2,
    "ram_pct": 42.8,
    "disk_pct": 58.1,
    "net_bytes_sent": 1420500,
    "net_bytes_recv": 8940200,
    "status": "active"
  }'
```

**Response (HTTP 200 OK):**
```json
{
  "status": "acknowledged",
  "pending_commands": [],
  "persona_update": null
}
```

---

### 9.4 List Fleet Endpoints (`GET /api/v1/controller/agents`)
Returns complete inventory of active and offline endpoints.

```bash
curl -k -X GET https://127.0.0.1:8443/api/v1/controller/agents
```

**Response (HTTP 200 OK):**
```json
[
  {
    "agent_id": "agent-win11-finance-01-9a4f",
    "hostname": "win11-finance-01",
    "os": "windows",
    "arch": "amd64",
    "ips": ["10.0.10.45"],
    "persona": "office_worker",
    "status": "online",
    "last_seen_sec": 2,
    "cpu_pct": 14.2,
    "ram_pct": 42.8
  }
]
```

---

### 9.5 Dispatch Remote Command (`POST /api/v1/controller/command`)
Queues an administrative command for execution on a specific agent.

```bash
curl -k -X POST https://127.0.0.1:8443/api/v1/controller/command \
  -H "Content-Type: application/json" \
  -d '{
    "agent_id": "agent-win11-finance-01-9a4f",
    "command": "whoami /all"
  }'
```

**Response (HTTP 200 OK):**
```json
{
  "task_id": "task-78192a",
  "status": "queued",
  "agent_id": "agent-win11-finance-01-9a4f"
}
```

---

### 9.6 Live Persona Reassignment (`POST /api/v1/controller/persona`)
Transitions an endpoint to a new persona without restarting the agent process.

```bash
curl -k -X POST https://127.0.0.1:8443/api/v1/controller/persona \
  -H "Content-Type: application/json" \
  -d '{
    "agent_id": "agent-win11-finance-01-9a4f",
    "persona": "sysadmin"
  }'
```

**Response (HTTP 200 OK):**
```json
{
  "status": "success",
  "agent_id": "agent-win11-finance-01-9a4f",
  "new_persona": "sysadmin"
}
```

---

## 10. Operational Security & Anti-Forensics Guarantees

RangeForge is built for responsible, non-destructive simulation:
1. **Automated Ephemeral Cleanup**: Every synthetic file created on local disks or SMB shares is tracked in memory. Upon agent shutdown (or receiving a termination signal), all created files are unlinked.
2. **Shell History Suppression**: Commands executed via the remote command dispatcher use PowerShell `-NoProfile` and avoid altering the user's `ConsoleHost_history.txt` or bash history.
3. **Unprivileged Execution**: Agents never require root or Administrator privileges. They execute within standard user security contexts.

---

## 11. Troubleshooting & Operator Diagnostics Runbook

### Scenario 1: Browser Displays TLS Certificate Warning
- **Symptom**: Browser warns that the connection is not private (`NET::ERR_CERT_AUTHORITY_INVALID`).
- **Root Cause**: RangeForge auto-generates a self-signed X.509 certificate on initial launch.
- **Remediation**: Click **Advanced -> Proceed to 127.0.0.1 (unsafe)**, or navigate directly to plaintext HTTP: `http://127.0.0.1:8443/`. Alternatively, install your organization's internal CA certificate into `certs/server.crt` and `certs/server.key`.

### Scenario 2: Host Agent Cannot Reach Manager
- **Symptom**: Agent logs `[ERROR] Failed to connect to manager: dial tcp 10.0.0.1:8443: connectex: A connection attempt failed`.
- **Root Cause**: Firewall blocking port 8443 or Manager service is not listening.
- **Remediation**:
  1. Verify Manager is running: `curl -k https://<manager-ip>:8443/api/v1/controller/agents`.
  2. Verify firewall allows outbound TCP port 8443 from agent subnet.
  3. Ensure Manager was started with `--host 0.0.0.0` rather than `127.0.0.1`.

### Scenario 3: Agents Show Status 'Offline' in Web UI
- **Symptom**: Agent displays as offline after 30 seconds.
- **Root Cause**: Agent process terminated or heartbeat packets are being dropped by an intermediate router.
- **Remediation**: Inspect agent terminal output for errors. Verify whether network jitter exceeds `heartbeat_timeout_sec` in `configs/manager.yaml`.

### Scenario 4: Command Execution Returns Access Denied
- **Symptom**: Remote command runner fails with permission denied.
- **Root Cause**: Command attempted an administrative operation requiring elevation.
- **Remediation**: RangeForge agents intentionally run as unprivileged users. Ensure commands target non-elevated diagnostics (e.g. `ipconfig`, `whoami`, `netstat`).
