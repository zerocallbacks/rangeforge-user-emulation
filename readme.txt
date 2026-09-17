================================================================================
RangeForge User Emulation Suite (rangeforge-ue)
Open-Source Cyber Range & Network User Activity Simulator
Version: 2.0.0-oss
Project URL: https://github.com/zerocallbacks/rangeforge-user-emulation
License: Apache 2.0
Target OS: Windows | Linux | FreeBSD (pfSense) | macOS
Language: Go 1.24+
================================================================================

TABLE OF CONTENTS
--------------------------------------------------------------------------------
1. Overview & Purpose
2. Architecture & Roles
3. Pre-Compiled Release Binaries
4. Quick Start & Execution Modes
5. Detailed Configuration Syntax & Parameter Manual
   - Manager Configuration (configs/manager.yaml)
   - Agent Configurations (configs/agent_windows.yaml & agent_unix.yaml)
   - Persona Emulation Profiles (configs/profiles/*.yaml)
   - Network Topology (configs/topology.json)
   - Cyber Range Scenarios (configs/ranges.json)
   - Administrative Authentication (configs/admin_auth.json)
   - Web & User Credential Corpus (corpus/)
6. Emulation Engines & Behavioral Realism
7. Building from Source
8. License & Open-Source Information


================================================================================
1. OVERVIEW & PURPOSE
================================================================================
In high-fidelity cyber ranges, red vs. blue team exercises, and security
operations training, a pervasive challenge is the lack of realistic background
network and host activity. Without normal user emulation, defensive blue teams
can trivially pinpoint red team activity, and red teams lack realistic host
environments and network noise to test operational security.

RangeForge User Emulation Suite (rangeforge-ue) is a single, zero-dependency,
cross-platform open-source Go binary that delivers realistic user emulation
across any network environment:
- Windows Hosts & Servers (Windows 10, 11, Server 2016/2019/2022/2025)
- Linux Hosts & Servers (Ubuntu, Debian, RHEL, CentOS, Rocky, Alpine)
- FreeBSD & pfSense Routers (FreeBSD amd64 binary target, single-interface traversal)
- macOS Workstations (Darwin amd64 & arm64 native support)


================================================================================
2. ARCHITECTURE & ROLES
================================================================================
rangeforge-ue operates as a single unified binary supporting three roles:

  [RANGEFORGE CONTROLLER]  (CLI Console / Operator Dashboard)
            |
            v (REST API / CLI)
  [RANGEFORGE MANAGER]     (Central Orchestrator, Port 8443 HTTPS)
            |
            v (HTTPS / TLS over shared network or pfSense)
  +---------+-----------------------------+
  |                                       |
  v                                       v
[RANGEFORGE HOST AGENT]                 [RANGEFORGE HOST AGENT]
  (Windows Target)                        (Linux / Unix Target)

- Manager: Central orchestrator providing the REST API, fleet heartbeat
  coordination, dynamic TLS listener, and web console.
- Controller: Operator CLI console to inspect fleet metrics, dispatch remote
  tasks, or tune personas on the fly.
- Host Agent: Cross-platform agent running on target endpoints. Compiles and
  executes natively across Windows, Linux, FreeBSD, and macOS. Communicates
  outbound-only over HTTPS (zero inbound listening ports on endpoints).


================================================================================
3. PRE-COMPILED RELEASE BINARIES
================================================================================
Pre-compiled standalone binaries are located in the bin/ directory:
- bin/rangeforge-ue.exe       (Windows 64-bit amd64)
- bin/rangeforge-ue-linux     (Linux 64-bit amd64)
- bin/rangeforge-ue-freebsd   (FreeBSD amd64 / pfSense 2.7+)


================================================================================
4. QUICK START & EXECUTION MODES
================================================================================
Launching the Interactive Web UI Dashboard (Recommended Local Test):
  Windows:
    .\bin\rangeforge-ue.exe ui --port 8443

  Linux / macOS:
    ./bin/rangeforge-ue-linux ui --port 8443

  Open your web browser at:
    https://127.0.0.1:8443/  (or http://127.0.0.1:8443/)

  Default Web UI Credentials:
    Username: admin
    Password: rangeforge

Running the Central Coordinator Manager:
  Linux / Unix:
    ./bin/rangeforge-ue-linux manager --host 0.0.0.0 --port 8443

Running the Endpoint Host Agent:
  Windows Endpoint:
    .\bin\rangeforge-ue.exe agent --manager https://10.0.0.1:8443 --persona office_worker

  Linux / Unix Endpoint:
    ./bin/rangeforge-ue-linux agent --manager https://10.0.0.1:8443 --persona developer

Using the Controller CLI:
  List all active fleet agents:
    ./bin/rangeforge-ue-linux controller agents --manager https://10.0.0.1:8443

  Inspect live host telemetry:
    ./bin/rangeforge-ue-linux controller stats --manager https://10.0.0.1:8443 --agent host-123

  Execute remote command on an agent:
    ./bin/rangeforge-ue-linux controller exec --manager https://10.0.0.1:8443 --agent host-123 --cmd "whoami /all"

  Dynamically reassign a persona during a live scenario:
    ./bin/rangeforge-ue-linux controller persona --manager https://10.0.0.1:8443 --agent host-123 --set-persona sysadmin

Standalone Evaluation Mode (Local Emulation, No Manager Required):
  .\bin\rangeforge-ue.exe standalone


================================================================================
5. DETAILED CONFIGURATION SYNTAX & PARAMETER MANUAL
================================================================================

--------------------------------------------------------------------------------
A. MANAGER CONFIGURATION (configs/manager.yaml)
--------------------------------------------------------------------------------
Controls the central coordinator server, TLS binding, and fleet timeouts.

Syntax and Parameters:
- listen_host (string, default "0.0.0.0"):
    The IP address the manager binds to. Use "0.0.0.0" to listen across all
    interfaces, or "127.0.0.1" for local-only testing.
- listen_port (integer, default 8443):
    TCP port for the unified HTTPS/HTTP service. RangeForge automatically
    multiplexes TLS and plaintext HTTP requests on this single port.
- tls_enabled (boolean, default true):
    Enforces TLS encryption for all agent and API communication.
- tls_cert_path (string, default "./certs/server.crt"):
    Filepath to PEM certificate. If absent, RangeForge automatically creates
    a high-security self-signed ECDSA P-256 certificate covering all local
    interfaces and range domain names.
- tls_key_path (string, default "./certs/server.key"):
    Filepath to PEM private key.
- heartbeat_timeout_sec (integer, default 30):
    Elapsed seconds without a check-in before an agent is flagged as offline.
- default_persona (string, default "office_worker"):
    Persona assigned to newly connecting agents that do not specify one.
- profiles_dir (string, default "./configs/profiles"):
    Directory where persona YAML files are located.
- allow_unsupported_os (boolean, default false):
    Guard preventing Manager execution on non-Unix production nodes. Set to
    true for local Windows evaluation.


--------------------------------------------------------------------------------
B. AGENT CONFIGURATIONS (configs/agent_windows.yaml & agent_unix.yaml)
--------------------------------------------------------------------------------
Controls the endpoint daemon running on each target machine.

Syntax and Parameters:
- manager_url (string, e.g. "https://10.0.0.1:8443"):
    Base HTTPS URL of the manager coordinator. The agent connects outbound-only.
- heartbeat_sec (integer, default 5):
    Interval in seconds between telemetry transmissions (CPU, RAM, disk, IP, tools).
- bind_interface (string, default ""):
    Network adapter name to bind. Leave empty ("") to auto-detect the default route.
- persona (string, default "office_worker"):
    Name of the emulation persona to load (matches filename in profiles without .yaml).
- web_corpus_file (string, default "./corpus/web_corpus.json"):
    JSON file defining web URLs, search queries, and User-Agents.
- creds_file (string, default "./corpus/user_creds.json"):
    JSON file containing simulated user identities and roles.
- shares_file (string, default ""):
    Optional JSON file defining network shares to traverse.
- enable_web (boolean, default true):
    Master toggle for HTTP/HTTPS web browsing traffic.
- enable_shares (boolean, default true):
    Master toggle for SMB/CIFS/NFS network drive traffic.
- enable_host_activity (boolean, default true):
    Master toggle for local benign process execution and office document drafts.
- tags (list of strings):
    Labels attached to telemetry (e.g. ["windows-11", "finance-vlan"]).


--------------------------------------------------------------------------------
C. PERSONA EMULATION PROFILES (configs/profiles/*.yaml)
--------------------------------------------------------------------------------
Personas govern human realism, traffic density, and host behaviors.
Four out-of-the-box personas are provided:
1. office_worker.yaml : High-dwell intranet browsing, Word/Excel document drafts
2. developer.yaml     : Fast technical browsing (GitHub/StackOverflow), code shares
3. sysadmin.yaml      : Network diagnostics (netstat/route/ping), audit logs
4. executive.yaml     : Executive memos, board reviews, financial summaries

Common Profile Syntax:
  name: <string>
    Unique identifier for the profile.
  description: <string>
    Human-readable explanation of persona intent.

  web_browsing:
    enabled: <boolean>
      Enable or disable web emulation for this profile.
    requests_per_minute_min: <integer>
      Lower bound of web requests per minute.
    requests_per_minute_max: <integer>
      Upper bound of web requests per minute.
    dwell_time_min_sec: <integer>
      Minimum human think time in seconds before following a link.
    dwell_time_max_sec: <integer>
      Maximum human think time in seconds before following a link.
    fetch_assets: <boolean>
      When true, downloads accompanying CSS/JS/images to generate realistic noise.
    user_agents: <list of strings>
      Custom User-Agent header list (empty [] rotates modern Chrome/Edge/Firefox).
    target_urls: <list of strings>
      URLs browsed by this persona.
    search_keywords: <list of strings>
      Search queries entered by this persona.

  file_share:
    enabled: <boolean>
      Enable or disable SMB/network drive emulation.
    interval_sec: <integer>
      Interval between share operations.
    shares: <list of strings>
      Target UNC paths (\\\\server\\share) or Unix mount paths (/mnt/share).
    read_ratio: <float 0.0 to 1.0>
      Percentage of operations that read files (e.g. 0.95 for 95%).
    write_ratio: <float 0.0 to 1.0>
      Percentage of operations that author/update files (e.g. 0.05 for 5%).

  ping:
    enabled: <boolean>
      Enable periodic ICMP keepalives.
    interval_sec: <integer>
      Seconds between ping bursts.
    targets: <list of strings>
      IP addresses or hostnames to ping.

  host_activity:
    enabled: <boolean>
      Enable host-level benign execution.
    interval_sec: <integer>
      Seconds between benign process invocations.
    safe_processes: <list of strings>
      Non-destructive commands executed matching the persona (e.g. whoami, hostname).
    simulate_office_docs: <boolean>
      Creates realistic temporary office draft files in user folders.
    temp_file_operations: <boolean>
      Creates temporary scratch/lock files with automatic cleanup on exit.


--------------------------------------------------------------------------------
D. NETWORK TOPOLOGY (configs/topology.json)
--------------------------------------------------------------------------------
Defines the subnet architecture rendered visually on the Web UI Topology Map.

Schema: Array of Subnet Gateway Objects:
  [
    {
      "name": "Edge Gateway",          // Gateway / router display name
      "ip": "10.0.0.1",                // Gateway router IP address
      "subnet": "10.0.0.0/16",         // Subnet block in CIDR notation
      "description": "Core Uplink DNS" // Scenario role description
    }
  ]


--------------------------------------------------------------------------------
E. CYBER RANGE SCENARIOS (configs/ranges.json)
--------------------------------------------------------------------------------
Defines exercise parameters, intensity multipliers, and DNS/gateway mapping.

Key Parameters:
- id: Scenario code (e.g. "RANGE-01")
- name: Human-readable scenario name
- state: Current state ("running", "paused", "stopped")
- intensity: Traffic pacing ("Low (0.5x)", "Standard (1.0x)", "High (2.5x)", "Stress (5.0x)")
- primary_cidr: Main range address block (e.g. "10.0.0.0/16")
- gateway_ip: Primary gateway router IP
- dns_server: Exercise DNS server
- domain_controller: Active Directory domain controller IP
- search_domain: Internal domain suffix (e.g. "range.local")
- subnets: List of child subnets with purpose and gateway definitions
- jitter_pct: Percentage variance applied to human dwell intervals (e.g. 30%)
- latency_ms: Simulated network latency in milliseconds


--------------------------------------------------------------------------------
F. ADMINISTRATIVE AUTHENTICATION (configs/admin_auth.json)
--------------------------------------------------------------------------------
Stores hashed credentials for Web UI and REST API operator access.

Schema:
  {
    "username": "admin",
    "password_hash": "<sha256_hash_of_salt_and_password>",
    "salt": "<16_byte_hex_salt>",
    "updated_at": "2026-09-07T18:30:00Z"
  }

Default Credentials:
  Username: admin
  Password: rangeforge

Password Hashing Algorithm:
  SHA-256(salt + ":" + password)


--------------------------------------------------------------------------------
G. CORPUS DIRECTORY (corpus/)
--------------------------------------------------------------------------------
- corpus/web_corpus.json:
    Contains lists of intranet portals, internet sites, search queries,
    static asset extensions, and browser User-Agent strings.
- corpus/user_creds.json:
    Mock user accounts with roles and assigned personas for realistic logging.
- corpus/wordlist.txt:
    Keywords used to generate realistic synthetic document titles
    (e.g. quarterly_budget.docx, incident_review.xlsx).


================================================================================
6. EMULATION ENGINES & BEHAVIORAL REALISM
================================================================================
A. Web Browsing Engine:
   - Evaluates URLs with randomized dwell times modeled on human Poisson/Gaussian
     distributions.
   - Rotates authentic modern User-Agents and HTTP headers (Accept-Language, Sec-Fetch).
   - Fully respects HTTP_PROXY, HTTPS_PROXY, and NO_PROXY environment variables.

B. File Share & Storage Engine:
   - Emulates SMB, CIFS, and local user directory activity.
   - Configurable read vs. write ratios prevent excessive disk bloat.
   - All synthetic files created are tracked in memory and automatically purged
     when the agent is halted.

C. Host Activity Engine:
   - Runs benign diagnostic commands without elevation.
   - Provides tunable execution noise levels (silent, low, standard, high).

D. Stealth Command Execution Engine:
   - Allows operators to execute ad-hoc commands via PowerShell or Bash.
   - Windows commands use -NoProfile and suppress shell history logging.
   - Streams exit code, stdout, stderr, and execution duration back to Manager.


================================================================================
7. BUILDING FROM SOURCE
================================================================================
Requires Go 1.24+:

Windows PowerShell:
  .\scripts\build.ps1 -Target all

Linux / Unix:
  chmod +x ./scripts/build.sh
  ./scripts/build.sh

Make:
  make build-all


================================================================================
8. LICENSE
================================================================================
RangeForge User Emulation Suite is released under the Apache 2.0 License.
See LICENSE for full details.
