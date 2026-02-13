# setup.ps1 — one-shot Windows bootstrap for sidecar
# Usage: irm https://raw.githubusercontent.com/marcus/sidecar/main/scripts/setup.ps1 | iex
#
# What it does:
#   1. Checks prerequisites (Go ≥ 1.22, git, psmux).
#   2. Installs sidecar via "go install".
#   3. Ensures %GOPATH%\bin is on the user PATH.
#   4. Prints next-steps.

Set-StrictMode -Version Latest
$ErrorActionPreference = "Stop"

function Write-Banner {
    Write-Host ""
    Write-Host "╔══════════════════════════════════════╗" -ForegroundColor Cyan
    Write-Host "║        sidecar — Windows setup        ║" -ForegroundColor Cyan
    Write-Host "╚══════════════════════════════════════╝" -ForegroundColor Cyan
    Write-Host ""
}

function Test-CommandExists([string]$cmd) {
    $null -ne (Get-Command $cmd -ErrorAction SilentlyContinue)
}

function Assert-Go {
    if (-not (Test-CommandExists "go")) {
        Write-Host "ERROR: Go is not installed or not on PATH." -ForegroundColor Red
        Write-Host "Install Go from https://go.dev/dl/ and re-run this script."
        exit 1
    }
    $ver = (go version) -replace '^go version go', '' -replace '\s.*', ''
    $parts = $ver.Split('.')
    $major = [int]$parts[0]
    $minor = [int]$parts[1]
    if ($major -lt 1 -or ($major -eq 1 -and $minor -lt 22)) {
        Write-Host "ERROR: Go $ver found, but sidecar requires Go ≥ 1.22." -ForegroundColor Red
        exit 1
    }
    Write-Host "  ✓ Go $ver" -ForegroundColor Green
}

function Assert-Git {
    if (-not (Test-CommandExists "git")) {
        Write-Host "  ⚠ git not found (optional, needed for worktree features)" -ForegroundColor Yellow
    } else {
        Write-Host "  ✓ git" -ForegroundColor Green
    }
}

function Assert-Psmux {
    if (-not (Test-CommandExists "psmux")) {
        Write-Host "  ⚠ psmux not found (install for terminal multiplexer support)" -ForegroundColor Yellow
    } else {
        Write-Host "  ✓ psmux" -ForegroundColor Green
    }
}

function Ensure-GoBinOnPath {
    $gobin = $env:GOBIN
    if (-not $gobin) {
        $gopath = $env:GOPATH
        if (-not $gopath) {
            $gopath = Join-Path $env:USERPROFILE "go"
        }
        $gobin = Join-Path $gopath "bin"
    }

    $userPath = [Environment]::GetEnvironmentVariable("Path", "User")
    if ($userPath -notlike "*$gobin*") {
        Write-Host "  Adding $gobin to user PATH..." -ForegroundColor Yellow
        [Environment]::SetEnvironmentVariable("Path", "$userPath;$gobin", "User")
        $env:Path = "$env:Path;$gobin"
        Write-Host "  ✓ PATH updated (restart your terminal if sidecar is not found)" -ForegroundColor Green
    } else {
        Write-Host "  ✓ $gobin already on PATH" -ForegroundColor Green
    }
}

function Install-Sidecar {
    Write-Host ""
    Write-Host "Installing sidecar..." -ForegroundColor Cyan
    & go install github.com/marcus/sidecar/cmd/sidecar@latest
    if ($LASTEXITCODE -ne 0) {
        Write-Host "ERROR: go install failed (exit $LASTEXITCODE)." -ForegroundColor Red
        exit 1
    }
    Write-Host "  ✓ sidecar installed" -ForegroundColor Green
}

# ── main ──────────────────────────────────────────────────
Write-Banner

Write-Host "Checking prerequisites..." -ForegroundColor Cyan
Assert-Go
Assert-Git
Assert-Psmux

Ensure-GoBinOnPath
Install-Sidecar

Write-Host ""
Write-Host "All done! Run 'sidecar' to get started." -ForegroundColor Green
Write-Host ""
