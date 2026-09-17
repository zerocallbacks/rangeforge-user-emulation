# ==============================================================================
# RangeForge User Emulation Suite - GitHub Repository Setup Helper
# ==============================================================================
param(
    [string]$RemoteUrl = ""
)

Write-Host "=================================================================" -ForegroundColor Cyan
Write-Host " RangeForge User Emulation Suite - GitHub Repository Initializer" -ForegroundColor Yellow
Write-Host "=================================================================" -ForegroundColor Cyan

# Ensure we are running inside the repository directory
$RepoDir = Split-Path -Parent (Split-Path -Parent $MyInvocation.MyCommand.Path)
Set-Location $RepoDir
Write-Host "[INFO] Working directory: $RepoDir" -ForegroundColor Gray

# 1. Verify Git is installed
try {
    $gitVersion = git --version
    Write-Host "[CHECK] $gitVersion found." -ForegroundColor Green
} catch {
    Write-Host "[ERROR] Git is not installed or not in PATH. Please install Git." -ForegroundColor Red
    exit 1
}

# 2. Initialize Git repository if not already initialized
if (-not (Test-Path ".git")) {
    Write-Host "[GIT] Initializing new Git repository..." -ForegroundColor Cyan
    git init
    git branch -M main
} else {
    Write-Host "[GIT] Git repository already initialized." -ForegroundColor Yellow
}

# 3. Check Git author identity
$userName = git config user.name
$userEmail = git config user.email
if (-not $userName -or -not $userEmail) {
    Write-Host "`n[GIT] Git user identity is not configured yet." -ForegroundColor Yellow
    if (-not $userName) {
        $userName = Read-Host "Enter your Git Name (e.g., John Doe)"
        if ($userName) { git config user.name $userName }
    }
    if (-not $userEmail) {
        $userEmail = Read-Host "Enter your Git Email (e.g., john@example.com)"
        if ($userEmail) { git config user.email $userEmail }
    }
}

# 4. Verify .gitignore protection
Write-Host "[SECURITY] Verifying that binary & credential exclusions are active..." -ForegroundColor Cyan
$ignoredBins = git status --ignored -s bin 2>$null
if ($ignoredBins) {
    Write-Host "[VERIFIED] Compiled binaries in bin/ are properly ignored." -ForegroundColor Green
}

# 5. Stage all repository files
Write-Host "[GIT] Staging files for initial open-source commit..." -ForegroundColor Cyan
git add .

# 5. Display status
Write-Host "`nStaged Files Summary:" -ForegroundColor Yellow
git status --short

# 6. Prompt to commit
$commitMsg = "feat: initial open-source release of RangeForge User Emulation Suite v2.0.0-oss"
Write-Host "`n[GIT] Ready to commit: '$commitMsg'" -ForegroundColor Cyan
$doCommit = Read-Host "Create initial commit now? (Y/N)"
if ($doCommit -match "^[Yy]$") {
    git commit -m $commitMsg
    Write-Host "[SUCCESS] Initial commit created on branch 'main'!" -ForegroundColor Green

    # 7. Add GitHub remote if provided
    if (-not $RemoteUrl) {
        $RemoteUrl = Read-Host "Enter your GitHub repository URL (or press Enter to skip, e.g. https://github.com/user/rangeforge-ue.git)"
    }

    if ($RemoteUrl) {
        git remote remove origin 2>$null
        git remote add origin $RemoteUrl
        Write-Host "[REMOTE] Added origin: $RemoteUrl" -ForegroundColor Green
        Write-Host "`nTo push your code to GitHub, run:" -ForegroundColor Cyan
        Write-Host "  git push -u origin main`n" -ForegroundColor White
        Write-Host "To publish the initial release binaries via GitHub Actions, push a version tag:" -ForegroundColor Cyan
        Write-Host "  git tag v2.0.0-oss" -ForegroundColor White
        Write-Host "  git push origin v2.0.0-oss`n" -ForegroundColor White
    } else {
        Write-Host "`nWhen you have created your repository on GitHub, run:" -ForegroundColor Cyan
        Write-Host "  git remote add origin <your-repo-url>" -ForegroundColor White
        Write-Host "  git push -u origin main" -ForegroundColor White
    }
} else {
    Write-Host "[ABORTED] Commit skipped. You can commit whenever you are ready using 'git commit -m ...'." -ForegroundColor Yellow
}
