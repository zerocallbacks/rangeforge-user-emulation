================================================================================
RangeForge User Emulation Suite (rangeforge-ue)
Open-Source Cyber Range & Network User Activity Simulator
Version: 2.0.0-oss
Project URL: https://github.com/zerocallbacks/rangeforge-user-emulation
License: Apache 2.0
Target OS: Windows | Linux | FreeBSD (pfSense) | macOS
Language: Go 1.24+
================================================================================

1. OVERVIEW & PURPOSE
--------------------------------------------------------------------------------
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


2. ARCHITECTURE & ROLES
--------------------------------------------------------------------------------
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
  executes natively across Windows, Linux, FreeBSD, and macOS.


3. EMULATION ENGINES
--------------------------------------------------------------------------------
A. Web Browsing Emulation:
   - Corpus-driven intranet and internet URL browsing from corpus/web_corpus.json
   - Human dwell times using randomized Gaussian/Poisson think time (5-25s)
   - Rotates realistic modern User-Agent strings, Accept-Language, and cookies
   - Fetches accompanying assets (.css, .js, .png, favicon.ico) for realistic noise
   - Respects HTTP_PROXY, HTTPS_PROXY, and NO_PROXY settings

B. File Share & Storage Emulation (SMB / Network Drives):
   - Emulates office document reading and writing against shared network drives
   - Configurable read/write ratios (e.g. 70% read, 30% write)
   - Generates and cleans up temporary documents using realistic templates

C. Host-Based Activity Emulation:
   - Creates realistic temporary files in user folders (Documents, Downloads)
   - Benign background tools (whoami, hostname, ping, netstat, git)
   - Blue team noise controls: silent, low, standard, high

D. Custom Command Execution Engine:
   - Execute ad-hoc commands on targeted hosts:
     * Windows: powershell.exe -ExecutionPolicy Bypass or cmd.exe
     * Unix: /bin/bash or /bin/sh
   - Captures exit code, stdout, stderr, and duration; reports back to Manager


4. PRE-COMPILED RELEASE BINARIES
--------------------------------------------------------------------------------
Pre-compiled standalone binaries are located in the bin/ directory:
- bin/rangeforge-ue.exe       (Windows 64-bit amd64)
- bin/rangeforge-ue-linux     (Linux 64-bit amd64)
- bin/rangeforge-ue-freebsd   (FreeBSD amd64 / pfSense 2.7+)


5. QUICK START & USAGE
--------------------------------------------------------------------------------
Launching the Web UI Dashboard (Local Host):
  Windows:
    .\bin\rangeforge-ue.exe ui --port 8443

  Open browser at:
    https://127.0.0.1:8443/

Running the Central Manager:
  Linux / Unix:
    ./bin/rangeforge-ue-linux manager --host 0.0.0.0 --port 8443

Running the Host Agent:
  Windows Target:
    .\bin\rangeforge-ue.exe agent --manager https://10.0.0.1:8443 --persona office_worker

  Linux Target:
    ./bin/rangeforge-ue-linux agent --manager https://10.0.0.1:8443 --persona developer

Using the Controller CLI:
  List all active fleet agents:
    ./bin/rangeforge-ue-linux controller agents --manager https://10.0.0.1:8443

  Inspect live host telemetry:
    ./bin/rangeforge-ue-linux controller stats --manager https://10.0.0.1:8443 --agent host-123

  Execute remote command:
    ./bin/rangeforge-ue-linux controller exec --manager https://10.0.0.1:8443 --agent host-123 --cmd "whoami /all"

Standalone Evaluation Mode (No Manager Required):
  .\bin\rangeforge-ue.exe standalone


6. OUT-OF-THE-BOX PERSONAS
--------------------------------------------------------------------------------
- office_worker : Intranet portals, news, docs; Word/Excel drafts; whoami, hostname
- developer     : GitHub, StackOverflow, language docs; code shares; git, compiler
- sysadmin      : Documentation, admin portals; IT shares; netstat, routes
- executive     : Briefings, corporate metrics; financial memos; review notes


7. BUILDING FROM SOURCE
--------------------------------------------------------------------------------
Requires Go 1.24+:

Windows PowerShell:
  .\scripts\build.ps1 -Target all

Linux / Unix:
  chmod +x ./scripts/build.sh
  ./scripts/build.sh

Make:
  make build-all


8. LICENSE
--------------------------------------------------------------------------------
RangeForge User Emulation Suite is released under the Apache 2.0 License.
See LICENSE for full details.
