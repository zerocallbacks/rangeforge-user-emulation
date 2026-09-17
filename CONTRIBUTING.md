# Contributing to RangeForge User Emulation Suite

Thank you for your interest in contributing to the **RangeForge User Emulation Suite (`rangeforge-ue`)**! We welcome contributions from cyber range engineers, security practitioners, developers, and educators.

---

## 1. Code of Conduct

All contributors and participants agree to abide by our [Code of Conduct](CODE_OF_CONDUCT.md). Please treat all community members with respect and courtesy.

---

## 2. Getting Started

### Prerequisites
- **Go**: Version 1.24 or later
- **Git**: 2.30+
- **Make** (optional, recommended on Linux/macOS)
- **PowerShell** (optional, recommended on Windows)

### Fork and Clone
```bash
# 1. Fork the repository on GitHub
# 2. Clone your personal fork
git clone https://github.com/<your-username>/rangeforge-user-emulation.git
cd rangeforge-user-emulation

# 3. Add upstream remote
git remote add upstream https://github.com/zerocallbacks/rangeforge-user-emulation.git
```

---

## 3. Development Workflow

### Branching Model
Always create a descriptive feature or fix branch from `main`:
```bash
git checkout -b feature/new-emulation-engine
# or
git checkout -b fix/tls-handshake-retry
```

### Building the Project
```bash
# Using Go directly:
go build ./cmd/rangeforge-ue

# Using Make:
make build-all

# Using PowerShell (Windows):
.\scripts\build.ps1 -Target all

# Using Bash (Linux/macOS):
./scripts/build.sh
```

### Running Tests
All contributions must pass the entire test suite:
```bash
# Run all package unit and integration tests
go test -v ./...

# Run tests with race detection (Linux/macOS)
go test -v -race ./...
```

---

## 4. Code Standards & Best Practices

1. **Format Code**: Run `go fmt ./...` before committing.
2. **Lint Cleanliness**: Run `go vet ./...` to detect structural errors.
3. **Cross-Platform Compatibility**:
   - Do NOT use OS-specific packages without build tags (e.g. `//go:build windows` or `//go:build !windows`).
   - Remember that `rangeforge-ue` targets Windows, Linux, FreeBSD (pfSense), and macOS.
4. **Security & Sensitive Data**:
   - **NEVER** commit TLS private keys, production passwords, or API tokens.
   - Use `corpus/` files for synthetic test data only.
   - Emulation engines must strictly adhere to the defined noise profiles and clean up any synthetic files upon termination.
5. **No Binary Bloat**:
   - Never commit compiled binaries (`.exe`, ELF, Mach-O) to Git.

---

## 5. Adding New Personas

Personas are defined in `configs/profiles/<persona_name>.yaml`. When adding a new persona:
1. Provide a realistic YAML profile in `configs/profiles/`.
2. Include realistic think times (Gaussian/Poisson jitter) and dwell intervals.
3. Test loading the profile via `rangeforge-ue standalone --config configs/profiles/<name>.yaml`.
4. Document the persona's intended cyber range role in `README.md`.

---

## 6. Submitting a Pull Request (PR)

1. Ensure your branch is rebased onto the latest `upstream/main`:
   ```bash
   git fetch upstream
   git rebase upstream/main
   ```
2. Push your branch to your GitHub fork:
   ```bash
   git push origin feature/new-emulation-engine
   ```
3. Open a Pull Request on GitHub against the `main` branch.
4. Fill out the [Pull Request Template](.github/PULL_REQUEST_TEMPLATE.md).
5. Ensure all CI automated checks pass.
6. A maintainer will review your pull request, request feedback or approval, and merge once verified.
