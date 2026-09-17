# ==============================================================================
# RangeForge User Emulation Suite - Windows Build & PowerShell Build Script
# Open-Source Cyber Range Platform
# ==============================================================================

param(
    [string]$Target = "all",
    [string]$OutDir = "bin"
)

Write-Host "=================================================================" -ForegroundColor Cyan
Write-Host " RangeForge User Emulation Suite (Open-Source)" -ForegroundColor Yellow
Write-Host " Build & Packaging Automation" -ForegroundColor White
Write-Host "=================================================================" -ForegroundColor Cyan

if (-not (Test-Path $OutDir)) {
    New-Item -ItemType Directory -Path $OutDir | Out-Null
}

function Build-Windows {
    Write-Host "[BUILD] Compiling Windows Agent/Binary ($OutDir/rangeforge-ue.exe)..." -ForegroundColor Green
    $env:GOOS = "windows"
    $env:GOARCH = "amd64"
    go build -trimpath -ldflags "-s -w" -o "$OutDir/rangeforge-ue.exe" ./cmd/rangeforge-ue
    if ($LASTEXITCODE -eq 0) {
        Write-Host "[SUCCESS] Created $OutDir/rangeforge-ue.exe" -ForegroundColor Green
    } else {
        Write-Host "[FAILED] Windows build failed" -ForegroundColor Red
    }
}

function Build-Linux {
    Write-Host "[BUILD] Compiling Linux Manager/Controller/Agent ($OutDir/rangeforge-ue-linux)..." -ForegroundColor Green
    $env:GOOS = "linux"
    $env:GOARCH = "amd64"
    go build -trimpath -ldflags "-s -w" -o "$OutDir/rangeforge-ue-linux" ./cmd/rangeforge-ue
    if ($LASTEXITCODE -eq 0) {
        Write-Host "[SUCCESS] Created $OutDir/rangeforge-ue-linux" -ForegroundColor Green
    } else {
        Write-Host "[FAILED] Linux build failed" -ForegroundColor Red
    }
}

function Build-FreeBSD {
    Write-Host "[BUILD] Compiling FreeBSD/pfSense Binary ($OutDir/rangeforge-ue-freebsd)..." -ForegroundColor Green
    $env:GOOS = "freebsd"
    $env:GOARCH = "amd64"
    go build -trimpath -ldflags "-s -w" -o "$OutDir/rangeforge-ue-freebsd" ./cmd/rangeforge-ue
    if ($LASTEXITCODE -eq 0) {
        Write-Host "[SUCCESS] Created $OutDir/rangeforge-ue-freebsd" -ForegroundColor Green
    } else {
        Write-Host "[FAILED] FreeBSD build failed" -ForegroundColor Red
    }
}

switch ($Target.ToLower()) {
    "windows" { Build-Windows }
    "linux"   { Build-Linux }
    "freebsd" { Build-FreeBSD }
    "all" {
        Build-Windows
        Build-Linux
        Build-FreeBSD
        Write-Host "All targets compiled into $OutDir/" -ForegroundColor Cyan
    }
    default {
        Write-Host "Unknown target: $Target. Use: windows, linux, freebsd, or all." -ForegroundColor Red
    }
}
