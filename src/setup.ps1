param(
    [Parameter(Position=0)]
    [string]$ProjectName = ""
)

[console]::InputEncoding = [console]::OutputEncoding = New-Object System.Text.UTF8Encoding

$RootDir = $PSScriptRoot
$ParentDir = Split-Path $RootDir

function Write-Header {
    Write-Host "    â–„â–ˆâ–ˆâ–ˆâ–ˆâ–„ â–ˆâ–ˆ â–„â–ˆâ–€ â–ˆâ–ˆ â–ˆâ–ˆâ–ˆâ–ˆâ–ˆâ–„  â–„â–ˆâ–ˆâ–ˆâ–ˆâ–„ Â©" -ForegroundColor Red
    Write-Host "    â–ˆâ–ˆâ–„â–„â–ˆâ–ˆ â–ˆâ–ˆâ–ˆâ–ˆ   â–ˆâ–ˆ â–ˆâ–ˆâ–„â–„â–ˆâ–ˆâ–„ â–ˆâ–ˆâ–„â–„â–ˆâ–ˆ " -ForegroundColor Red
    Write-Host "    â–ˆâ–ˆ  â–ˆâ–ˆ â–ˆâ–ˆ â–€â–ˆâ–„ â–ˆâ–ˆ â–ˆâ–ˆ   â–ˆâ–ˆ â–ˆâ–ˆ  â–ˆâ–ˆ                                             [+] 1988.`n" -ForegroundColor Red
}

Clear-Host
Write-Header

if (-not $ProjectName) {
    $ProjectName = Read-Host "Enter app name (default: my-awesome-app)"
    if ([string]::IsNullOrWhiteSpace($ProjectName)) {
        $ProjectName = "my-awesome-app"
    }
}

function Get-InteractiveMultiSelect {
    param(
        [string]$Title,
        [string[]]$Options,
        [string]$Subtitle = "Use Up/Down arrows to navigate, Space to toggle, 'A' to select all, 'N' to select none, Enter to finish."
    )
    if ($Options.Length -eq 0) { return @() }
    
    $selected = @($true) * $Options.Length
    $currentIndex = 0

    $done = $false
    while (-not $done) {
        Clear-Host
        Write-Header
        Write-Host "=== $Title ===" -ForegroundColor Cyan
        Write-Host "$Subtitle`n" -ForegroundColor DarkGray

        for ($i = 0; $i -lt $Options.Length; $i++) {
            $check = if ($selected[$i]) { "[X]" } else { "[ ]" }
            if ($i -eq $currentIndex) {
                Write-Host "> $check $($Options[$i])" -ForegroundColor Black -BackgroundColor White
            } else {
                Write-Host "  $check $($Options[$i])"
            }
        }

        # Read a key press securely without echoing to the screen
        $key = [System.Console]::ReadKey($true)

        switch ($key.Key) {
            'UpArrow' {
                if ($currentIndex -gt 0) { $currentIndex-- }
            }
            'DownArrow' {
                if ($currentIndex -lt ($Options.Length - 1)) { $currentIndex++ }
            }
            'Spacebar' {
                $selected[$currentIndex] = -not $selected[$currentIndex]
            }
            'Enter' {
                $done = $true
            }
            'A' {
                for ($i = 0; $i -lt $selected.Length; $i++) { $selected[$i] = $true }
            }
            'N' {
                for ($i = 0; $i -lt $selected.Length; $i++) { $selected[$i] = $false }
            }
        }
    }
    
    $result = @()
    for ($i = 0; $i -lt $Options.Length; $i++) {
        if ($selected[$i]) {
            $result += $Options[$i]
        }
    }
    Clear-Host
    Write-Header
    return $result
}

$AvailableSkills = @(
    "Prettier & Tailwind Canonical Classes",
    "Lenis (Smooth Scrolling)",
    "Lucide React (Icons)"
)

$AvailableTemplateItems = @(
    ".specify folder (Design & rules)",
    "docs folder (Documentation)",
    "AGENTS.md (Agent overview)",
    "skills-lock.json (Skill dependencies)"
)

# Dynamically fetch Agent Skills
$TemplateDir = Join-Path $RootDir "template"
$SkillsDir = Join-Path $TemplateDir ".agents\skills"
$AvailableAgentSkills = @()
if (Test-Path $SkillsDir) {
    $AvailableAgentSkills = Get-ChildItem -Path $SkillsDir -Directory | Select-Object -ExpandProperty Name
}

$SelectedSkills = Get-InteractiveMultiSelect -Title "Select NPM Dependencies to Install" -Options $AvailableSkills
$SelectedAgentSkills = Get-InteractiveMultiSelect -Title "Select Agent Skills to Copy (.agents/skills)" -Options $AvailableAgentSkills
$SelectedTemplates = Get-InteractiveMultiSelect -Title "Select Other Template Assets to Copy" -Options $AvailableTemplateItems

Write-Host "Starting Next.js project setup for '$ProjectName'..." -ForegroundColor Green

# 1. Project Creation
$TargetDir = Join-Path $ParentDir $ProjectName
Write-Host "`nRunning create-next-app..."
npx --yes create-next-app@latest $TargetDir --yes
Set-Location -Path $TargetDir -ErrorAction Stop

# 2. Copy Local Template Assets
if (Test-Path $TemplateDir) {
    Write-Host "`nCopying selected template assets into project..."
    
    # Agent Skills
    if ($SelectedAgentSkills.Count -gt 0) {
        $TargetSkillsDir = Join-Path $TargetDir ".agents\skills"
        New-Item -ItemType Directory -Force -Path $TargetSkillsDir | Out-Null
        foreach ($skill in $SelectedAgentSkills) {
            $src = Join-Path $SkillsDir $skill
            if (Test-Path $src) { Copy-Item -Path $src -Destination $TargetSkillsDir -Recurse -Force }
        }
    }

    # Other Template Assets
    if ($SelectedTemplates -contains ".specify folder (Design & rules)") {
        $src = Join-Path $TemplateDir ".specify"
        if (Test-Path $src) { Copy-Item -Path $src -Destination . -Recurse -Force }
    }
    if ($SelectedTemplates -contains "docs folder (Documentation)") {
        $src = Join-Path $TemplateDir "docs"
        if (Test-Path $src) { Copy-Item -Path $src -Destination . -Recurse -Force }
    }
    if ($SelectedTemplates -contains "AGENTS.md (Agent overview)") {
        $src = Join-Path $TemplateDir "AGENTS.md"
        if (Test-Path $src) { Copy-Item -Path $src -Destination . -Force }
    }
    if ($SelectedTemplates -contains "skills-lock.json (Skill dependencies)") {
        $src = Join-Path $TemplateDir "skills-lock.json"
        if (Test-Path $src) { Copy-Item -Path $src -Destination . -Force }
    }
}

# 3. Linter & Formatter Installation
Write-Host "`nInstalling dependencies..."
if ($SelectedSkills -contains "Prettier & Tailwind Canonical Classes") {
    Write-Host "Installing Prettier and formatters..." -ForegroundColor Cyan
    npm install -D prettier @laststance/tailwind-suggest-canonical-classes prettier-plugin-tailwindcss-canonical-classes --no-audit --no-fund

    # Update package.json to add formatting scripts
    $nodeScript = @'
const fs = require("fs");
const pkg = JSON.parse(fs.readFileSync("package.json", "utf8"));
pkg.scripts = {
  ...pkg.scripts,
  "lint:tailwind": "tailwind-suggest-canonical-classes \"src/**/*.{ts,tsx,js,jsx,html}\"",
  "format:fix": "prettier --write \"src/**/*.{ts,tsx,js,jsx,html}\""
};
fs.writeFileSync("package.json", JSON.stringify(pkg, null, 2));
'@
    node -e $nodeScript

    # Create .prettierrc
    @'
{
  "plugins": ["prettier-plugin-tailwindcss-canonical-classes"]
}
'@ | Out-File -FilePath .prettierrc -Encoding utf8
}

if ($SelectedSkills -contains "Lenis (Smooth Scrolling)") {
    Write-Host "Installing Lenis..." -ForegroundColor Cyan
    npm install lenis --no-audit --no-fund
}

if ($SelectedSkills -contains "Lucide React (Icons)") {
    Write-Host "Installing Lucide React..." -ForegroundColor Cyan
    npm install lucide-react --no-audit --no-fund
}

Write-Host "`nSetup Complete for '$ProjectName'!" -ForegroundColor Green

