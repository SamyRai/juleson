# Installation Guide

Juleson provides multiple installation methods for Linux, macOS, and Windows. Choose the
method that best fits your needs.

## Table of Contents

- [Quick Install](#quick-install)
- [Pre-built Binaries](#method-1-pre-built-binaries-recommended)
- [Go Install](#method-2-go-install)
- [Build from Source](#method-3-build-from-source)
- [Package Managers](#method-4-package-managers)
- [Configuration](#configuration)
- [Verification](#verification)
- [Troubleshooting](#troubleshooting)

## Quick Install

### Linux/macOS (Quick Install)

```bash
# Using pre-built binaries
curl -L https://github.com/SamyRai/juleson/releases/latest/download/install.sh | bash

# Or using Go (requires Go 1.23+)
go install github.com/SamyRai/juleson/cmd/juleson@latest
go install github.com/SamyRai/juleson/cmd/jules-mcp@latest
```

### Windows (Quick Install)

```powershell
# Using PowerShell
irm https://github.com/SamyRai/juleson/releases/latest/download/install.ps1 | iex

# Or using Go (requires Go 1.23+)
go install github.com/SamyRai/juleson/cmd/juleson@latest
go install github.com/SamyRai/juleson/cmd/jules-mcp@latest
```

## Method 1: Pre-built Binaries (Recommended)

Download pre-built binaries from the [GitHub Releases](https://github.com/SamyRai/juleson/releases) page.

### Linux (x86_64)

```bash
# Download binaries
curl -L -o juleson-linux-amd64.tar.gz https://github.com/SamyRai/juleson/releases/latest/download/juleson-linux-amd64.tar.gz
curl -L -o jules-mcp-linux-amd64.tar.gz https://github.com/SamyRai/juleson/releases/latest/download/jules-mcp-linux-amd64.tar.gz

# Extract
tar -xzf juleson-linux-amd64.tar.gz
tar -xzf jules-mcp-linux-amd64.tar.gz

# Install to /usr/local/bin (requires sudo)
sudo mv juleson /usr/local/bin/
sudo mv jules-mcp /usr/local/bin/

# Make executable
sudo chmod +x /usr/local/bin/juleson
sudo chmod +x /usr/local/bin/jules-mcp

# Verify
juleson --version
jules-mcp --version
```

### Linux (ARM64)

```bash
# Download ARM64 binaries
curl -L -o juleson-linux-arm64.tar.gz https://github.com/SamyRai/juleson/releases/latest/download/juleson-linux-arm64.tar.gz
curl -L -o jules-mcp-linux-arm64.tar.gz https://github.com/SamyRai/juleson/releases/latest/download/jules-mcp-linux-arm64.tar.gz

# Extract and install
tar -xzf juleson-linux-arm64.tar.gz
tar -xzf jules-mcp-linux-arm64.tar.gz
sudo mv juleson jules-mcp /usr/local/bin/
sudo chmod +x /usr/local/bin/juleson /usr/local/bin/jules-mcp
```

### macOS (Intel)

```bash
# Download Intel binaries
curl -L -o juleson-darwin-amd64.tar.gz https://github.com/SamyRai/juleson/releases/latest/download/juleson-darwin-amd64.tar.gz
curl -L -o jules-mcp-darwin-amd64.tar.gz https://github.com/SamyRai/juleson/releases/latest/download/jules-mcp-darwin-amd64.tar.gz

# Extract
tar -xzf juleson-darwin-amd64.tar.gz
tar -xzf jules-mcp-darwin-amd64.tar.gz

# Install to /usr/local/bin
sudo mv juleson /usr/local/bin/
sudo mv jules-mcp /usr/local/bin/
sudo chmod +x /usr/local/bin/juleson /usr/local/bin/jules-mcp

# Verify
juleson --version
jules-mcp --version
```

### macOS (Apple Silicon - M1/M2/M3)

```bash
# Download ARM64 binaries for Apple Silicon
curl -L -o juleson-darwin-arm64.tar.gz https://github.com/SamyRai/juleson/releases/latest/download/juleson-darwin-arm64.tar.gz
curl -L -o jules-mcp-darwin-arm64.tar.gz https://github.com/SamyRai/juleson/releases/latest/download/jules-mcp-darwin-arm64.tar.gz

# Extract and install
tar -xzf juleson-darwin-arm64.tar.gz
tar -xzf jules-mcp-darwin-arm64.tar.gz
sudo mv juleson jules-mcp /usr/local/bin/
sudo chmod +x /usr/local/bin/juleson /usr/local/bin/jules-mcp
```

### Windows (x86_64)

```powershell
# Download binaries
Invoke-WebRequest -Uri "https://github.com/SamyRai/juleson/releases/latest/download/juleson-windows-amd64.zip" -OutFile "juleson.zip"
Invoke-WebRequest -Uri "https://github.com/SamyRai/juleson/releases/latest/download/jules-mcp-windows-amd64.zip" -OutFile "jules-mcp.zip"

# Extract (requires PowerShell 5.0+)
Expand-Archive -Path "juleson.zip" -DestinationPath "C:\Program Files\juleson"
Expand-Archive -Path "jules-mcp.zip" -DestinationPath "C:\Program Files\juleson"

# Add to PATH (run as Administrator)
[Environment]::SetEnvironmentVariable(
    "Path",
    [Environment]::GetEnvironmentVariable("Path", "Machine") + ";C:\Program Files\juleson",
    "Machine"
)

# Restart PowerShell and verify
juleson --version
jules-mcp --version
```

## Method 2: Go Install

Requires Go 1.23 or higher.

### All Platforms

```bash
# Install CLI
go install github.com/SamyRai/juleson/cmd/juleson@latest

# Install MCP Server
go install github.com/SamyRai/juleson/cmd/jules-mcp@latest

# Verify (binaries installed to $GOPATH/bin or $HOME/go/bin)
juleson --version
jules-mcp --version
```

### Ensure Go bin is in PATH

#### Linux/macOS

```bash
# Add to ~/.bashrc, ~/.zshrc, or ~/.profile
export PATH="$PATH:$(go env GOPATH)/bin"

# Reload shell configuration
source ~/.bashrc  # or ~/.zshrc or ~/.profile
```

#### Windows

```powershell
# Add GOPATH\bin to PATH
$goPath = go env GOPATH
[Environment]::SetEnvironmentVariable(
    "Path",
    $env:Path + ";$goPath\bin",
    "User"
)
```

## Method 3: Build from Source

### Prerequisites

- Go 1.23 or higher
- Git
- Make (Linux/macOS) or PowerShell (Windows)

### Linux/macOS (Build from Source)

```bash
# Clone the repository
git clone https://github.com/SamyRai/juleson.git
cd juleson

# Download dependencies
go mod download

# Build using Make
make build

# Install to system
sudo make install

# Or use the dev command for custom installation
./bin/juleson dev install --path /usr/local/bin

# Verify
juleson --version
jules-mcp --version
```

### Windows (Build from Source)

```powershell
# Clone the repository
git clone https://github.com/SamyRai/juleson.git
cd juleson

# Download dependencies
go mod download

# Build binaries
go build -o bin\juleson.exe .\cmd\juleson\main.go
go build -o bin\jules-mcp.exe .\cmd\jules-mcp\main.go

# Copy to a directory in PATH
New-Item -ItemType Directory -Force -Path "C:\Program Files\juleson"
Copy-Item bin\juleson.exe "C:\Program Files\juleson\"
Copy-Item bin\jules-mcp.exe "C:\Program Files\juleson\"

# Add to PATH
[Environment]::SetEnvironmentVariable(
    "Path",
    $env:Path + ";C:\Program Files\juleson",
    "Machine"
)
```

## Method 4: Package Managers

### Homebrew (macOS/Linux)

```bash
# Tap the repository (once published)
brew tap SamyRai/juleson

# Install
brew install juleson

# Verify
juleson --version
jules-mcp --version
```

### Snap (Linux)

```bash
# Install from Snap Store (once published)
sudo snap install juleson

# Verify
juleson --version
```

### Chocolatey (Windows)

```powershell
# Install from Chocolatey (once published)
choco install juleson

# Verify
juleson --version
```

### AUR (Arch Linux)

```bash
# Using yay or another AUR helper (once published)
yay -S juleson

# Verify
juleson --version
```

## Configuration

After installation, configure your Jules API key:

### Linux/macOS (Configuration)

#### Environment Variable (Temporary)

```bash
export JULES_API_KEY="your-jules-api-key-here"
```

#### Environment Variable (Permanent - Linux/macOS)

```bash
# Add to ~/.bashrc, ~/.zshrc, or ~/.profile
echo 'export JULES_API_KEY="your-jules-api-key-here"' >> ~/.bashrc
source ~/.bashrc
```

#### Configuration File (Linux/macOS)

```bash
# Create config directory
mkdir -p ~/.config/juleson

# Create config file
cat > ~/.config/juleson/config.yaml << EOF
jules:
  api_key: "your-jules-api-key-here"
  base_url: "https://api.jules.ai"
  timeout: 30
  retry_attempts: 3
EOF
```

### Windows (Configuration)

#### Environment Variable (PowerShell - Temporary)

```powershell
$env:JULES_API_KEY = "your-jules-api-key-here"
```

#### Environment Variable (Permanent)

```powershell
[Environment]::SetEnvironmentVariable(
    "JULES_API_KEY",
    "your-jules-api-key-here",
    "User"
)
```

#### Configuration File

```powershell
# Create config directory
New-Item -ItemType Directory -Force -Path "$env:APPDATA\juleson"

# Create config file
@"
jules:
  api_key: "your-jules-api-key-here"
  base_url: "https://api.jules.ai"
  timeout: 30
  retry_attempts: 3
"@ | Out-File -FilePath "$env:APPDATA\juleson\config.yaml" -Encoding UTF8
```

## Verification

### Test Installation

```bash
# Check version
juleson --version
jules-mcp --version

# View help
juleson --help
jules-mcp --help

# List available commands
juleson template list

# Initialize a config file
juleson init
```

### Test MCP Server

```bash
# The MCP server communicates via stdio
# It should start without errors
jules-mcp
# Press Ctrl+C to stop
```

## Troubleshooting

### Common Issues

#### 1. "command not found"

**Linux/macOS:**

```bash
# Check if binary is in PATH
which juleson

# If not found, add to PATH
export PATH="$PATH:/usr/local/bin"

# Or if using Go install
export PATH="$PATH:$(go env GOPATH)/bin"
```

**Windows:**

```powershell
# Check if binary is in PATH
where.exe juleson

# If not found, add directory to PATH
$currentPath = [Environment]::GetEnvironmentVariable("Path", "User")
[Environment]::SetEnvironmentVariable("Path", "$currentPath;C:\Program Files\juleson", "User")
```

#### 2. "permission denied"

**Linux/macOS:**

```bash
# Make binary executable
chmod +x /path/to/juleson
chmod +x /path/to/jules-mcp

# Or reinstall with proper permissions
sudo mv juleson /usr/local/bin/
sudo chmod +x /usr/local/bin/juleson
```

**Windows:**

```powershell
# Run PowerShell as Administrator
# Right-click PowerShell → Run as Administrator
```

#### 3. "API key required"

```bash
# Set the API key
export JULES_API_KEY="your-api-key-here"

# Or create a config file (see Configuration section)
juleson init
```

#### 4. "go: command not found" (for go install method)

Install Go from the official website:

- **Linux**: <https://golang.org/dl/> → Download → Extract to `/usr/local`
- **macOS**: `brew install go` or download from <https://golang.org/dl/>
- **Windows**: Download installer from <https://golang.org/dl/>

#### 5. Certificate/SSL Errors

**Linux:**

```bash
# Update CA certificates
sudo apt-get update && sudo apt-get install ca-certificates
# or
sudo yum install ca-certificates
```

**macOS:**

```bash
# Update certificates
brew install ca-certificates
```

**Windows:**

```powershell
# Update Windows
# SSL certificates are managed by Windows Update
```

#### 6. Architecture Mismatch

Check your system architecture and download the appropriate binary:

```bash
# Linux/macOS
uname -m
# x86_64 = amd64
# aarch64 or arm64 = arm64

# Windows
echo $env:PROCESSOR_ARCHITECTURE
# AMD64 = amd64
```

### Platform-Specific Notes

#### Linux

- Works on all major distributions (Ubuntu, Debian, CentOS, Fedora, Arch, etc.)
- Requires glibc 2.27+ (most modern distributions)
- For older distributions, build from source

#### macOS

- Compatible with macOS 10.15 (Catalina) and later
- Intel (x86_64) and Apple Silicon (ARM64) binaries available
- May require allowing the app in System Preferences → Security & Privacy on first run

#### Windows (Platform Notes)

- Requires Windows 10 or later (Windows 11 recommended)
- Works in PowerShell, Command Prompt, and Windows Terminal
- Windows Defender may flag the binary - add an exception if needed

## Getting Help

- **Documentation**: <https://github.com/SamyRai/juleson/tree/master/docs>
- **Issues**: <https://github.com/SamyRai/juleson/issues>
- **Discussions**: <https://github.com/SamyRai/juleson/discussions>
- **MCP Server Guide**: [docs/MCP_SERVER_USAGE.md](./MCP_SERVER_USAGE.md)

## Next Steps

After successful installation:

1. **Configure your API key** (see Configuration section)
2. **Initialize a project**: `juleson init`
3. **List templates**: `juleson template list`
4. **Analyze a project**: `juleson analyze`
5. **Read the documentation**: [README.md](../README.md)

## Uninstallation

### Linux/macOS (Uninstallation)

```bash
# Remove binaries
sudo rm /usr/local/bin/juleson
sudo rm /usr/local/bin/jules-mcp

# Remove configuration
rm -rf ~/.config/juleson

# If installed via Go
rm $(go env GOPATH)/bin/juleson
rm $(go env GOPATH)/bin/jules-mcp
```

### Windows (Uninstallation)

```powershell
# Remove binaries
Remove-Item "C:\Program Files\juleson\juleson.exe"
Remove-Item "C:\Program Files\juleson\jules-mcp.exe"

# Remove from PATH
$path = [Environment]::GetEnvironmentVariable("Path", "Machine")
$newPath = ($path.Split(';') | Where-Object { $_ -ne "C:\Program Files\juleson" }) -join ';'
[Environment]::SetEnvironmentVariable("Path", $newPath, "Machine")

# Remove configuration
Remove-Item -Recurse -Force "$env:APPDATA\juleson"
```
