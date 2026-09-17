## Description
Provide a concise explanation of what this pull request introduces, fixes, or refactors.

## Related Issues
Closes #(issue) or Relates to #(issue)

## Type of Change
- [ ] Bug fix (non-breaking change which fixes an issue)
- [ ] New feature (non-breaking change which adds emulation functionality)
- [ ] Breaking change (fix or feature that would cause existing agent/manager setups not to work)
- [ ] Performance optimization / Resource reduction
- [ ] Documentation update

## Scope & Platform Validation
- [ ] Windows (tested on Windows 10/11 or Server)
- [ ] Linux (tested on Ubuntu, Debian, CentOS/Rocky, or Alpine)
- [ ] FreeBSD / pfSense
- [ ] macOS (Darwin amd64 / arm64)

## Pre-Submission Quality Checklist
- [ ] My code adheres to the project's Go style guidelines (`gofmt`, `go vet`).
- [ ] I have executed `go test -v ./...` and all tests pass without errors.
- [ ] I have added appropriate unit or integration tests for new functionality.
- [ ] I have updated relevant documentation (`README.md`, config comments, or docs).
- [ ] I have verified that **zero credentials, private keys, or compiled binaries** are committed in this PR.
- [ ] I have verified that any background tasks properly clean up temporary artifacts on shutdown.
