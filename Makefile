# ==============================================================================
# RangeForge User Emulation Suite Build Automation
# Open-Source Cyber Range Platform
# ==============================================================================

BINARY_NAME=rangeforge-ue
BIN_DIR=bin
VERSION=2.0.0-oss

.PHONY: all build build-all test clean run-standalone

all: test build

build:
	@echo "Building native $(BINARY_NAME)..."
	go build -trimpath -ldflags "-s -w -X 'rangeforge-ue/pkg/branding.Version=$(VERSION)'" -o $(BIN_DIR)/$(BINARY_NAME) ./cmd/rangeforge-ue

build-all:
	@echo "Building Windows, Linux, and FreeBSD (pfSense) targets..."
	@mkdir -p $(BIN_DIR)
	GOOS=windows GOARCH=amd64 go build -trimpath -ldflags "-s -w" -o $(BIN_DIR)/$(BINARY_NAME).exe ./cmd/rangeforge-ue
	GOOS=linux GOARCH=amd64 go build -trimpath -ldflags "-s -w" -o $(BIN_DIR)/$(BINARY_NAME)-linux ./cmd/rangeforge-ue
	GOOS=freebsd GOARCH=amd64 go build -trimpath -ldflags "-s -w" -o $(BIN_DIR)/$(BINARY_NAME)-freebsd ./cmd/rangeforge-ue

test:
	@echo "Running full test suite..."
	go test -v ./...

clean:
	@echo "Cleaning artifacts..."
	rm -rf $(BIN_DIR)

run-standalone: build
	./$(BIN_DIR)/$(BINARY_NAME) standalone
