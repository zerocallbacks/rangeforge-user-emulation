#!/usr/bin/env bash
# ==============================================================================
# RangeForge User Emulation Suite - Unix Build & Cross-Compile Script
# Open-Source Cyber Range Platform
# ==============================================================================

set -e

OUT_DIR="bin"
mkdir -p "$OUT_DIR"

echo "================================================================="
echo " RangeForge User Emulation Suite (Open-Source)"
echo " Build & Packaging Automation"
echo "================================================================="

echo "[BUILD] Compiling Native Unix Binary ($OUT_DIR/rangeforge-ue)..."
go build -trimpath -ldflags "-s -w" -o "$OUT_DIR/rangeforge-ue" ./cmd/rangeforge-ue

echo "[BUILD] Cross-compiling Linux amd64 ($OUT_DIR/rangeforge-ue-linux)..."
GOOS=linux GOARCH=amd64 go build -trimpath -ldflags "-s -w" -o "$OUT_DIR/rangeforge-ue-linux" ./cmd/rangeforge-ue

echo "[BUILD] Cross-compiling Windows amd64 ($OUT_DIR/rangeforge-ue.exe)..."
GOOS=windows GOARCH=amd64 go build -trimpath -ldflags "-s -w" -o "$OUT_DIR/rangeforge-ue.exe" ./cmd/rangeforge-ue

echo "[BUILD] Cross-compiling FreeBSD / pfSense ($OUT_DIR/rangeforge-ue-freebsd)..."
GOOS=freebsd GOARCH=amd64 go build -trimpath -ldflags "-s -w" -o "$OUT_DIR/rangeforge-ue-freebsd" ./cmd/rangeforge-ue

echo "All targets compiled successfully into '$OUT_DIR/'"
