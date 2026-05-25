# Security Policy

## Supported Versions

Juleson is pre-1.0. Security fixes are handled on `main` unless a maintained release branch exists.

## Report A Vulnerability

Do not open a public issue for vulnerabilities.

Email: <security@glpx.pro>

Include:

- affected version or commit
- description and impact
- reproduction steps
- relevant logs or proof of concept
- suggested fix, if available

Expected response:

- initial response within 48 hours
- status update within 7 days
- fix timeline based on severity and scope

## Secrets

- Do not commit API keys, tokens, passwords, private keys, or local config containing credentials.
- Prefer environment variables for setup inputs and untracked local config files
  for stored credentials.
- Keep local config files untracked.
- Redact credentials in logs, issue reports, and screenshots.

## Local Configuration

```bash
export JULES_API_KEY="..."
export GITHUB_TOKEN="..."
juleson setup --non-interactive
```

Optional config:

```bash
cp configs/juleson.example.yaml configs/juleson.yaml
```

## Runtime Considerations

- Review templates before executing them against a working tree.
- Use least-privilege GitHub tokens.
- Avoid `--auto-approve` unless unattended execution is acceptable.
- Inspect patches before applying session changes.
- Keep dependencies current and review security scan output in CI.
