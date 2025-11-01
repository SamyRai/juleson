# Security Policy

## Supported Versions

We release patches for security vulnerabilities for the following versions:

| Version | Supported          |
| ------- | ------------------ |
| 1.x.x   | :white_check_mark: |
| < 1.0   | :x:                |

## Reporting a Vulnerability

We take security seriously. If you discover a security vulnerability, please follow these steps:

### 1. **Do Not** Open a Public Issue

Please do not create a public GitHub issue for security vulnerabilities.

### 2. Report Privately

Send an email to: **<security@glpx.pro>**

Include the following information:

- Description of the vulnerability
- Steps to reproduce the issue
- Potential impact
- Suggested fix (if any)
- Your contact information

### 3. Response Timeline

- **Initial Response**: Within 48 hours
- **Status Update**: Within 7 days
- **Fix Timeline**: Varies based on severity
  - Critical: 1-3 days
  - High: 1-2 weeks
  - Medium: 2-4 weeks
  - Low: Best effort

### 4. Disclosure Policy

- We will acknowledge receipt of your vulnerability report
- We will investigate and validate the issue
- We will develop a fix and coordinate disclosure
- We will credit you for the discovery (unless you prefer to remain anonymous)

## Security Best Practices

### API Keys and Secrets

- **Never** commit API keys, passwords, or secrets to the repository
- Use environment variables for sensitive configuration
- Use `configs/Juleson.example.yaml` as a template
- Keep `configs/Juleson.yaml` in `.gitignore`

### Configuration

```bash
# Set environment variables
export JULES_API_KEY="your-api-key"

# Or use configuration file
cp configs/Juleson.example.yaml configs/Juleson.yaml
# Edit configs/Juleson.yaml and add your API key
```

### Running Securely

- Keep dependencies up to date: `go get -u ./...`
- Review code before executing templates
- Use least privilege for file system operations
- Validate input from external sources

## Known Security Considerations

### 1. Template Execution

Templates execute operations on your file system. Review templates before executing:

```bash
# Review template before execution
./bin/juleson template view modular-restructure

# Execute with dry-run first
./bin/juleson execute --template modular-restructure --project ./test --dry-run
```

### 2. API Communications

All communications with Jules API use HTTPS. Ensure your system's CA certificates are up to date.

### 3. File System Access

The tool requires read/write access to your project directories. Be cautious
when running on sensitive codebases.

## Security Updates

Security updates will be released as patch versions and announced via:

- GitHub Security Advisories
- Release notes
- Email notifications (for reported vulnerabilities)

## Scope

This security policy applies to:

- All code in the `Juleson` repository
- Official binary releases
- Docker images (if applicable)

Out of scope:

- Third-party integrations
- User-created custom templates
- Forked repositories

## Bug Bounty

We currently do not offer a bug bounty program, but we greatly appreciate responsible disclosure.

## Recognition

We maintain a [SECURITY_CONTRIBUTORS.md](SECURITY_CONTRIBUTORS.md) file to
acknowledge security researchers who have helped improve the project's
security.

---

Thank you for helping keep Juleson and our users safe!
