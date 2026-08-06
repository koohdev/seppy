# =====================================================================
# Seppy CLI Installer Script for Windows (PowerShell)
# =====================================================================

$ErrorActionPreference = "Stop"

Write-Host "`n    ▄████▄ ██████ █████▄ █████▄ ██  ██" -ForegroundColor Red
Write-Host "    ██▄▄▄▄ ██▄▄   ██▄▄██ ██▄▄██ ▀████▀" -ForegroundColor Red
Write-Host "    ▄▄▄▄██ ██████ ██     ██       ██`n" -ForegroundColor Red
Write-Host "    Installing Seppy CLI All-in-One Project Generator..." -ForegroundColor Yellow

$seppyHome = Join-Path $HOME ".seppy"
$binDir = Join-Path $seppyHome "bin"
$docsDir = Join-Path $seppyHome "docs"
$cacheDir = Join-Path $seppyHome "cache\skills"
$configFile = Join-Path $seppyHome "config.json"

# 1. Create directory structure
New-Item -ItemType Directory -Force -Path $binDir | Out-Null
New-Item -ItemType Directory -Force -Path $docsDir | Out-Null
New-Item -ItemType Directory -Force -Path $cacheDir | Out-Null

# 2. Build Go binary and install to ~/.seppy/bin/
$scriptDir = Split-Path -Parent $MyInvocation.MyCommand.Path
$srcDir = Join-Path $scriptDir "src"

if (Test-Path (Join-Path $srcDir "main.go")) {
    Write-Host "  Building Seppy CLI engine from Go source..." -ForegroundColor Cyan
    Push-Location $srcDir
    & "go" build -ldflags="-s -w" -o (Join-Path $binDir "seppy.exe") main.go
    Copy-Item -Path (Join-Path $binDir "seppy.exe") -Destination (Join-Path $binDir "setup.exe") -Force
    Pop-Location
    Write-Host "  [OK] Binary compiled and installed to $binDir" -ForegroundColor Green
} else {
    Write-Host "  [!] Error: src/main.go not found." -ForegroundColor Red
}

# 3. Create default config.json if missing
if (!(Test-Path $configFile)) {
    $defaultConfig = @'
{
  "default_unselect_all": true,
  "custom_skills_commands": [],
  "custom_npm_packages": []
}
'@
    Set-Content -Path $configFile -Value $defaultConfig -Encoding UTF8
    Write-Host "  [OK] Created user config at $configFile" -ForegroundColor Green
}

# 5. Add ~/.seppy/bin to User PATH
$userPath = [System.Environment]::GetEnvironmentVariable("Path", "User")
if ($userPath -notlike "*$binDir*") {
    [System.Environment]::SetEnvironmentVariable("Path", $userPath + ";$binDir", "User")
    $env:Path = "$env:Path;$binDir"
    Write-Host "  [OK] Added $binDir to User PATH" -ForegroundColor Green
}

Write-Host "`n  Seppy installation complete! Reopen PowerShell or run 'seppy'.`n" -ForegroundColor Cyan
