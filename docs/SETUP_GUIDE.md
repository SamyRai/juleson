# Juleson Setup Guide

## Quick Start

The easiest way to get started with Juleson is to run the setup wizard:

```bash
juleson setup
```

This interactive wizard will guide you through:

1. **Shell Completion Installation** - Auto-complete commands and flags
2. **Jules API Configuration** - Set up your Jules API credentials
3. **GitHub Integration** - Connect your GitHub account (optional)
4. **Configuration Validation** - Verify everything is working

## Installation Methods

### Method 1: Interactive Setup (Recommended)

Run the setup wizard and follow the prompts:

```bash
juleson setup
```

The wizard will:

- Detect your shell (bash, zsh, fish)
- Offer to install shell completion
- Guide you through credential setup
- Save your configuration
- Validate the setup

### Method 2: Non-Interactive Setup

For automated deployments or CI/CD:

```bash
# Set environment variables
export JULES_API_KEY="your-api-key"
export GITHUB_TOKEN="your-github-token"

# Run non-interactive setup
juleson setup --non-interactive
```

### Method 3: Manual Configuration

Create `~/.juleson.yaml` or `./configs/juleson.yaml`:

```yaml
jules:
  api_key: "your-jules-api-key"
  base_url: "https://jules.googleapis.com/v1alpha"
  timeout: "30s"
  retry_attempts: 3

github:
  token: "your-github-token"
  default_org: ""
  pr:
    default_merge_method: "squash"
    auto_delete_branch: true
  discovery:
    enabled: true
    use_git_remote: true
    cache_ttl: "5m"
```

## Setup Options

### Skip Specific Steps

```bash
# Skip shell completion installation
juleson setup --skip-completion

# Skip GitHub configuration
juleson setup --skip-github

# Skip Jules API configuration
juleson setup --skip-jules

# Combine multiple skip flags
juleson setup --skip-completion --skip-github
```

### Environment Variables

The setup command respects these environment variables:

- `JULES_API_KEY` - Jules API key
- `GITHUB_TOKEN` - GitHub Personal Access Token
- `SHELL` - Shell type (detected automatically)

## Shell Completion

### Bash

#### Linux

```bash
# Install completion
juleson completion bash | sudo tee /etc/bash_completion.d/juleson

# Or use setup wizard
juleson setup

# Reload shell
source ~/.bashrc
```

#### macOS

```bash
# Install bash-completion if needed
brew install bash-completion

# Install juleson completion
juleson completion bash > $(brew --prefix)/etc/bash_completion.d/juleson

# Add to ~/.bash_profile
[[ -r "$(brew --prefix)/etc/profile.d/bash_completion.sh" ]] && . "$(brew --prefix)/etc/profile.d/bash_completion.sh"

# Reload shell
source ~/.bash_profile
```

### Zsh

```bash
# Option 1: Use setup wizard (recommended)
juleson setup

# Option 2: Manual installation
mkdir -p ~/.zfunc
juleson completion zsh > ~/.zfunc/_juleson

# Add to ~/.zshrc
fpath=(~/.zfunc $fpath)
autoload -Uz compinit && compinit

# Reload shell
source ~/.zshrc
```

### Fish

```bash
# Option 1: Use setup wizard (recommended)
juleson setup

# Option 2: Manual installation
juleson completion fish > ~/.config/fish/completions/juleson.fish

# Reload shell
source ~/.config/fish/config.fish
```

### PowerShell

```powershell
# Add to your PowerShell profile
juleson completion powershell | Out-String | Invoke-Expression

# Or save to profile permanently
juleson completion powershell >> $PROFILE
```

## Jules API Setup

### Getting Your API Key

1. Go to [Jules Settings](https://jules.ai/settings/api)
2. Create a new API key
3. Copy the key

### Configure via Setup

```bash
juleson setup
# Follow the prompts to enter your API key
```

### Configure via Environment

```bash
export JULES_API_KEY="your-api-key"
```

### Configure via Config File

Add to `~/.juleson.yaml`:

```yaml
jules:
  api_key: "your-jules-api-key"
```

## GitHub Integration

### Getting Your Personal Access Token

1. Go to [GitHub Settings > Tokens](https://github.com/settings/tokens)
2. Click "Generate new token (classic)"
3. Select scopes:
   - ✅ `repo` - Full control of private repositories
   - ✅ `workflow` - Update GitHub Actions workflows
   - ✅ `read:org` - Read org and team membership (optional)
4. Generate and copy the token

### Configure via Setup

```bash
juleson setup
# Follow the prompts to enter your GitHub token
```

### Configure via GitHub Login

```bash
juleson github login
```

### Configure via Environment

```bash
export GITHUB_TOKEN="ghp_your_token_here"
```

### Configure via Config File

Add to `~/.juleson.yaml`:

```yaml
github:
  token: "ghp_your_token_here"
```

## Verifying Your Setup

### Check Configuration

```bash
# Run setup validation
juleson setup --non-interactive

# Check GitHub integration
juleson github status

# List your repositories
juleson github repos --limit 5
```

### Test Jules Connection

```bash
# List your Jules sessions
juleson sessions list

# Create a test session
juleson sessions create "Analyze my project structure"
```

## Troubleshooting

### Shell Completion Not Working

**Bash:**

```bash
# Ensure bash-completion is installed
which bash-completion

# Reload completion
source /etc/bash_completion.d/juleson

# Check if function exists
type _juleson
```

**Zsh:**

```bash
# Check fpath includes .zfunc
echo $fpath

# Verify completion file exists
ls -la ~/.zfunc/_juleson

# Rebuild completion cache
rm -f ~/.zcompdump
compinit
```

**Fish:**

```bash
# Verify completion file exists
ls -la ~/.config/fish/completions/juleson.fish

# Rebuild completions
fish_update_completions
```

### Jules API Key Not Working

```bash
# Verify API key is set
juleson setup --skip-completion --skip-github

# Check config file
cat ~/.juleson.yaml | grep api_key

# Test with explicit key
JULES_API_KEY="your-key" juleson sessions list
```

### GitHub Token Issues

```bash
# Verify token is set
juleson github status

# Re-authenticate
juleson github login

# Test token manually
curl -H "Authorization: token YOUR_TOKEN" https://api.github.com/user
```

## Configuration Locations

Juleson looks for configuration in these locations (in order):

1. `./configs/juleson.yaml` - Project-specific config
2. `./.juleson.yaml` - Current directory config
3. `~/.juleson.yaml` - User home directory config
4. `/etc/juleson/juleson.yaml` - System-wide config

Environment variables override config file settings.

## Next Steps

After setup is complete:

```bash
# Analyze your project
juleson analyze

# View available templates
juleson template list

# Create an automation session
juleson sessions create "Refactor my CLI commands"

# View all available commands
juleson help
```

## Automated Setup Example

For CI/CD or automated deployments:

```bash
#!/bin/bash

# Set credentials
export JULES_API_KEY="${JULES_API_KEY}"
export GITHUB_TOKEN="${GITHUB_TOKEN}"

# Run non-interactive setup
juleson setup \
  --non-interactive \
  --skip-completion

# Verify setup
juleson github status
juleson sessions list
```

## Uninstall

To remove Juleson configuration:

```bash
# Remove config files
rm -f ~/.juleson.yaml
rm -rf ./configs/juleson.yaml

# Remove shell completions
rm -f ~/.zfunc/_juleson                          # Zsh
rm -f ~/.bash_completion.d/juleson               # Bash
rm -f ~/.config/fish/completions/juleson.fish    # Fish

# Remove environment variables from shell profile
# Edit ~/.zshrc, ~/.bashrc, or ~/.config/fish/config.fish
```

## Support

If you encounter issues during setup:

1. Check the [Troubleshooting](#troubleshooting) section
2. Run `juleson setup` again to reconfigure
3. Check configuration with `juleson github status`
4. View detailed help with `juleson setup --help`

For more help:

- [Configuration Guide](../configs/README.md)
- [GitHub Integration Guide](./GITHUB_CONFIGURATION_GUIDE.md)
- [CLI Reference](./CLI_REFERENCE.md)
