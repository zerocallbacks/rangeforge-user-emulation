# Changelog

All notable changes to the **RangeForge User Emulation Suite (`rangeforge-ue`)** will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

---

## [2.0.0-oss] - 2026-09-17

### Added
- **Single-Binary Multi-Role Architecture**: Unified `manager`, `agent`, `controller`, `ui`, and `standalone` commands in a single, zero-dependency Go binary.
- **Cross-Platform Host Agent**: Native support for Windows (Win 10/11, Server 2016–2025), Linux (Ubuntu, Debian, RHEL, CentOS, Alpine), FreeBSD 14+ (pfSense 2.7+ / OPNsense), and macOS (Apple Silicon and Intel).
- **Interactive Cyber Console Web UI**:
  - Live SVG sparklines plotting Fleet Average CPU and RAM usage with 15-point smoothing.
  - Interactive SVG Network Topology Map with live link state and gateway visualization.
  - Host Telemetry Inspector modal rendering CPU, RAM, Disk progress bars, MAC address, and MTU.
  - Dual theme engine supporting **Ocean Sapphire** and **Carbon Black** with persistent preferences.
  - One-click Executive Summary Brief generator exporting ready-to-use Markdown reports.
- **Realistic Emulation Engines**:
  - **Web Browsing**: Human think time with randomized Gaussian/Poisson dwell jitter, realistic modern browser user-agent rotation, and automatic static asset retrieval (`.css`, `.js`, `favicon.ico`).
  - **SMB / Network Shares**: Authenticated share traversal, customizable read/write distribution ratios (e.g. 95/5 or 70/30), and automatic temporary business document generation and cleanup.
  - **Host Activity & Safe Processes**: Synthetic office file operations (`.docx`, `.xlsx`, `.pdf`) in user profile folders with automated cleanup on shutdown.
  - **ICMP Ping Emulator**: Realistic background network keepalives and routing noise.
  - **Stealth Command Dispatcher**: Remote command execution via PowerShell (`-NoProfile`, shell history suppression) on Windows and Bash/sh on Unix, returning status and duration to the Manager.
- **Dynamic TLS & Multi-Network SANs**:
  - Zero-configuration high-security ECDSA P-256 self-signed certificate generation on startup.
  - Automatic SAN enumeration across all physical and virtual interfaces, hostnames, and range wildcards (`*.local`, `*.lan`, `*.corp`, `*.range.local`).
  - Dual HTTP/HTTPS multiplexing on port 8443 for seamless browser and agent communication.
- **Robust Offline Resilience**:
  - Outbound-only agent architecture requiring zero inbound listening ports on endpoints or firewalls.
  - In-memory telemetry sample buffering (up to 50 samples) with exponential reconnect backoff during router or firewall reloads.
- **CI/CD & Automation**:
  - GitHub Actions multi-OS test matrix (`ubuntu-latest`, `windows-latest`, `macos-latest`).
  - Automated multi-arch release pipeline building Windows, Linux, FreeBSD, and macOS binaries with SHA-256 checksums.

[2.0.0-oss]: https://github.com/zerocallbacks/rangeforge-user-emulation/releases/tag/v2.0.0-oss
