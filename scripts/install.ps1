param(
    [string]$Version = "latest",
    [string]$InstallDir = "$env:USERPROFILE\.juleson\bin",
    [string]$Repo = "SamyRai/juleson",
    [string]$BaseUrl = $env:JULESON_INSTALL_BASE_URL,
    [switch]$NoPathUpdate,
    [switch]$Help
)

Set-StrictMode -Version Latest
$ErrorActionPreference = "Stop"

function Show-Usage {
    @"
Install the latest Juleson release binaries.

Usage:
  .\install.ps1 [-Version <tag|latest>] [-InstallDir <path>] [-Repo <owner/repo>] [-BaseUrl <url>] [-NoPathUpdate]

Examples:
  irm https://github.com/SamyRai/juleson/releases/latest/download/install.ps1 | iex
  .\install.ps1 -InstallDir "$env:USERPROFILE\bin"
  .\install.ps1 -Version v1.0.0 -NoPathUpdate
"@
}

if ($Help) {
    Show-Usage
    exit 0
}

if ([string]::IsNullOrWhiteSpace($Version) -or [string]::IsNullOrWhiteSpace($InstallDir) -or [string]::IsNullOrWhiteSpace($Repo)) {
    throw "Version, install directory, and repository must be non-empty."
}
$installRoot = [System.IO.Path]::GetPathRoot($InstallDir)
if ($InstallDir -ne $installRoot) {
    $InstallDir = $InstallDir.TrimEnd('\', '/')
}

$arch = switch ($env:PROCESSOR_ARCHITECTURE) {
    "AMD64" { "amd64" }
    "ARM64" { "arm64" }
    default { throw "Unsupported Windows architecture: $env:PROCESSOR_ARCHITECTURE" }
}

if (-not [string]::IsNullOrWhiteSpace($BaseUrl)) {
    $baseUrl = $BaseUrl.TrimEnd("/")
} elseif ($Version -eq "latest") {
    $baseUrl = "https://github.com/$Repo/releases/latest/download"
} else {
    $baseUrl = "https://github.com/$Repo/releases/download/$Version"
}

$tempDir = Join-Path ([System.IO.Path]::GetTempPath()) ("juleson-install-" + [System.Guid]::NewGuid().ToString("N"))
New-Item -ItemType Directory -Path $tempDir | Out-Null

try {
    New-Item -ItemType Directory -Force -Path $InstallDir | Out-Null

    foreach ($binary in @("juleson", "jules-mcp")) {
        $asset = "$binary-windows-$arch.zip"
        $archive = Join-Path $tempDir $asset
        $extractDir = Join-Path $tempDir $binary

        Write-Host "Downloading $asset..."
        Invoke-WebRequest -Uri "$baseUrl/$asset" -OutFile $archive
        Expand-Archive -Path $archive -DestinationPath $extractDir -Force

        $source = Join-Path $extractDir "$binary.exe"
        if (-not (Test-Path $source)) {
            throw "Release asset $asset did not contain $binary.exe."
        }

        Copy-Item -Path $source -Destination (Join-Path $InstallDir "$binary.exe") -Force
    }

    if (-not $NoPathUpdate) {
        $currentPath = [Environment]::GetEnvironmentVariable("Path", "User")
        $pathEntries = $currentPath -split ";" | Where-Object { $_ }
        if ($pathEntries -notcontains $InstallDir) {
            $newPath = if ([string]::IsNullOrWhiteSpace($currentPath)) { $InstallDir } else { "$currentPath;$InstallDir" }
            [Environment]::SetEnvironmentVariable("Path", $newPath, "User")
            Write-Host "Added $InstallDir to your user PATH. Restart your shell before running juleson."
        }
    }

    Write-Host "Installed juleson and jules-mcp to $InstallDir"
} finally {
    Remove-Item -Recurse -Force $tempDir -ErrorAction SilentlyContinue
}
