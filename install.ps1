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

# 2. Copy compiled binary to ~/.seppy/bin/seppy.exe
$scriptDir = Split-Path -Parent $MyInvocation.MyCommand.Path
$sourceExe = Join-Path $scriptDir "seppy.exe"
if (!(Test-Path $sourceExe)) {
    $sourceExe = Join-Path $scriptDir "..\seppy.exe"
}

if (Test-Path $sourceExe) {
    Copy-Item -Path $sourceExe -Destination (Join-Path $binDir "seppy.exe") -Force
    Copy-Item -Path $sourceExe -Destination (Join-Path $binDir "setup.exe") -Force
    Write-Host "  [OK] Binary installed to $binDir" -ForegroundColor Green
} else {
    Write-Host "  [!] Warning: seppy.exe not found in $scriptDir. Build it first with 'go build'." -ForegroundColor Yellow
}

# 3. Reflect local skills & docs into ~/.seppy
$templateDir = Join-Path $scriptDir "template"
if (!(Test-Path $templateDir)) {
    $templateDir = Join-Path $scriptDir "..\template"
}

if (Test-Path $templateDir) {
    $localSkills = Join-Path $templateDir ".agents\skills"
    if (Test-Path $localSkills) {
        Copy-Item -Path "$localSkills\*" -Destination $cacheDir -Recurse -Force -ErrorAction SilentlyContinue
        Write-Host "  [OK] Synced local skills to $cacheDir" -ForegroundColor Green
    }
    $localDocs = Join-Path $templateDir "docs"
    if (Test-Path $localDocs) {
        Copy-Item -Path "$localDocs\*" -Destination $docsDir -Recurse -Force -ErrorAction SilentlyContinue
        Write-Host "  [OK] Synced local markdown docs to $docsDir" -ForegroundColor Green
    }
}

# 4. Create default config.json if missing
if (!(Test-Path $configFile)) {
    $defaultConfig = @'
{
  "default_unselect_all": true,
  "custom_skills_commands": [
    {
      "name": "npx skills add (Vercel Find-Skills)",
      "command": "npx skills add https://github.com/vercel-labs/skills --skill find-skills"
    }
  ],
  "custom_npm_packages": [
    "Framer Motion (Animations)",
    "Zustand (State Management)"
  ]
}
'@
    Set-Content -Path $configFile -Value $defaultConfig -Encoding UTF8
    Write-Host "  [OK] Created user config at $configFile" -ForegroundColor Green
}

# 5. Add ~/.seppy/bin to User PATH
$userPath = [System.Environment]::GetEnvironmentVariable("Path", "User")
if ($userPath -notlike "*$binDir*") {
    [System.Environment]::SetEnvironmentVariable("Path", $userPath + ";$binDir", "User")
    Write-Host "  [OK] Added $binDir to User PATH" -ForegroundColor Green
}

# 6. Add seppy shortcut function to PowerShell Profile
if (!(Test-Path -Path $PROFILE)) {
    New-Item -Type File -Path $PROFILE -Force | Out-Null
}

$profileShortcut = @'

# Seppy CLI Global Shortcuts
function seppy {
    Start-Process -FilePath "$HOME\.seppy\bin\seppy.exe"
    Stop-Process -Id $PID -Force
}

function setup {
    Start-Process -FilePath "$HOME\.seppy\bin\seppy.exe"
    Stop-Process -Id $PID -Force
}
'@

$profileContent = Get-Content -Path $PROFILE -Raw -ErrorAction SilentlyContinue
if ($profileContent -notlike "*seppy.exe*") {
    Add-Content -Path $PROFILE -Value $profileShortcut
    Write-Host "  [OK] Added 'seppy' command shortcut to PowerShell Profile ($PROFILE)" -ForegroundColor Green
}

Write-Host "`n  🎉 Seppy installation complete! Close and reopen your terminal, then type 'seppy'.`n" -ForegroundColor Cyan
