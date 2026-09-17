# TLS Certificates Directory

This directory is intended for RangeForge Manager TLS certificates (`server.crt` and `server.key`).

### Automatic Certificate Generation
If no certificates are present, RangeForge Manager automatically creates high-security ECDSA P-256 self-signed certificates on startup using `pkg/tlsutil`. The dynamically generated certificate includes Subject Alternative Names (SANs) for:
- `localhost` and `127.0.0.1`
- All active local network interfaces (IPv4 and IPv6)
- System hostname and FQDN
- Common cyber range internal domain patterns (`*.local`, `*.lan`, `*.corp`, `*.internal`, `*.home.arpa`, `*.range.local`)

### Custom Enterprise / CA Certificates
To use your own enterprise or cyber range CA certificates, place them in this folder with the following filenames:
- `server.crt` (PEM-encoded X.509 certificate chain)
- `server.key` (PEM-encoded private key)

> **Security Note**: Never commit private keys (`*.key`) or production certificates to version control. This folder is configured with `.gitignore` to prevent accidental credential leakage.
