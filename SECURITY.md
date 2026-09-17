# Security Policy

## Supported Versions

We provide security updates and patches for the following versions of **RangeForge User Emulation Suite (`rangeforge-ue`)**:

| Version | Supported | Notes |
| :--- | :--- | :--- |
| `2.0.x-oss` | :white_check_mark: | Current Active Release |
| `< 2.0.0` | :x: | Legacy / Deprecated |

---

## Reporting a Vulnerability

We take the security of RangeForge seriously. If you believe you have discovered a security vulnerability or sensitive information exposure:

1. **Do NOT file a public issue.** Publicly disclosing vulnerabilities puts range operators and training networks at risk before patches are ready.
2. Report vulnerabilities privately via **[GitHub Private Vulnerability Reporting](https://github.com/zerocallbacks/rangeforge-user-emulation/security/advisories/new)**.
3. Alternatively, email the maintainers directly with details:
   - Vulnerability description and impact assessment
   - Target component (Manager, Host Agent, Controller, Web Console, or Dynamic TLS Generator)
   - Step-by-step reproduction instructions or proof of concept
   - Proposed mitigation or patch (if available)

### Response Commitments
- **Initial Response**: Within 48 hours of receiving your report.
- **Triage & Reproduction**: Within 5 business days.
- **Remediation & Advisory**: We will coordinate a release date and credit your contribution in the security advisory (unless you prefer anonymity).

---

## Operational Security in Cyber Ranges

RangeForge is intentionally designed to generate realistic network and endpoint activity for training, blue team defense drills, and range fidelity. Maintainers adhere to the following safety constraints:
- **Stealth Command Execution**: RangeForge requires mutual authentication (TLS + Session/API token) before accepting remote command payloads.
- **Synthetic Cleanup**: All temporary synthetic documents created by host activity engines are scoped to temporary paths and automatically purged upon agent stop.
- **Zero Inbound Agent Ports**: Host agents initiate all connections outbound to the Manager over HTTPS; agents do not listen on network sockets or expose external endpoints.
