# Contributing

## Development Setup

Requirements:

- Go 1.25 or newer
- Git
- Jules API access for integration flows that call Jules
- GitHub token for Jules-created pull request command testing

Clone and prepare:

```bash
git clone https://github.com/SamyRai/juleson.git
cd juleson
go mod download
```

Build:

```bash
go build -o bin/juleson ./cmd/juleson
go build -o bin/jsn ./cmd/juleson
go build -o bin/builder ./cmd/builder
```

Run tests:

```bash
go test ./...
```

Run local quality checks:

```bash
juleson dev check
```

## Configuration For Local Testing

Use environment variables for non-interactive setup, or put credentials in an
untracked local config file:

```bash
export JULES_API_KEY="..."
export GITHUB_TOKEN="..."
juleson setup --non-interactive
```

Optional config file:

```bash
cp configs/juleson.example.yaml configs/juleson.yaml
```

Do not commit local config files or credentials.

## Pull Requests

1. Create a focused branch.
2. Keep changes scoped to one behavior or documentation area.
3. Add or update tests for behavior changes.
4. Update docs for user-facing CLI, config, workflow, or API changes.
5. Run the relevant local checks before opening the PR.

For documentation-only changes, run:

```bash
markdownlint '**/*.md'
```

## Code Style

- Use `gofmt -s`.
- Keep package responsibilities narrow.
- Prefer small interfaces at boundaries.
- Return contextual errors.
- Do not log or commit secrets.

## Documentation Style

- Keep root Markdown limited to `README.md`, `CONTRIBUTING.md`, `SECURITY.md`, and `LICENSE`.
- Put user, operator, and architecture docs under `docs/`.
- Use direct technical language.
- Keep commands runnable and aligned with current code.
- Update [docs/README.md](docs/README.md) when adding or moving docs.

## Release Notes

Update [docs/CHANGELOG.md](docs/CHANGELOG.md) for user-visible changes.
