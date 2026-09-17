#!/usr/bin/env bash
# ==============================================================================
# RangeForge User Emulation Suite - GitHub Repository Setup Helper (Unix)
# ==============================================================================
set -e

REPO_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "$REPO_DIR"

echo "================================================================="
echo " RangeForge User Emulation Suite - GitHub Repository Initializer"
echo "================================================================="
echo "[INFO] Working directory: $REPO_DIR"

# Check Git
if ! command -v git &>/dev/null; then
  echo "[ERROR] Git is not installed. Please install Git first."
  exit 1
fi

# Initialize Git
if [ ! -d ".git" ]; then
  echo "[GIT] Initializing Git repository..."
  git init
  git branch -M main
else
  echo "[GIT] Repository already initialized."
fi

# Check Git author identity
USER_NAME=$(git config user.name || true)
USER_EMAIL=$(git config user.email || true)
if [ -z "$USER_NAME" ]; then
  read -p "Enter your Git Name (e.g. John Doe): " USER_NAME
  [ -n "$USER_NAME" ] && git config user.name "$USER_NAME"
fi
if [ -z "$USER_EMAIL" ]; then
  read -p "Enter your Git Email (e.g. john@example.com): " USER_EMAIL
  [ -n "$USER_EMAIL" ] && git config user.email "$USER_EMAIL"
fi

# Stage files
echo "[GIT] Staging files..."
git add .

echo "Staged Files:"
git status --short

read -p "Create initial commit now? (y/n) " -n 1 -r
echo
if [[ $REPLY =~ ^[Yy]$ ]]; then
  git commit -m "feat: initial open-source release of RangeForge User Emulation Suite v2.0.0-oss"
  echo "[SUCCESS] Initial commit created on branch 'main'."
  
  read -p "Enter GitHub repository URL (or press Enter to skip): " REMOTE_URL
  if [ -n "$REMOTE_URL" ]; then
    git remote remove origin 2>/dev/null || true
    git remote add origin "$REMOTE_URL"
    echo "[REMOTE] Added origin: $REMOTE_URL"
    echo ""
    echo "To push your code to GitHub, run:"
    echo "  git push -u origin main"
    echo ""
    echo "To publish release binaries, push a version tag:"
    echo "  git tag v2.0.0-oss"
    echo "  git push origin v2.0.0-oss"
  fi
fi
