# CLI Reference

This reference reflects the current Cobra command tree from `cmd/juleson`.

## Global Usage

```bash
juleson [command]
```

Available commands:

| Command | Purpose |
| --- | --- |
| `actions` | Manage GitHub Actions workflows and runs |
| `activities` | Manage Jules session activities |
| `agent` | Run agent-based development tasks |
| `ai-orchestrate` | Let Gemini plan and run a multi-step workflow |
| `analyze` | Analyze project structure and context |
| `completion` | Generate shell completion scripts |
| `dev` | Build, test, lint, format, and release helpers |
| `execute` | Execute templates and automation tasks |
| `github` | Manage GitHub integration |
| `init` | Initialize a project for Jules automation |
| `orchestrate` | Run predefined multi-task Jules workflows |
| `pr` | Manage PRs created by Jules sessions |
| `sessions` | Manage Jules sessions |
| `setup` | Run first-time setup |
| `sources` | Manage Jules sources |
| `sync` | Sync a project with a remote repository |
| `template` | Manage templates |
| `version` | Print version information |

## Setup

```bash
juleson setup [flags]
```

Flags:

```text
--non-interactive   Run setup without prompts
--skip-completion   Skip shell completion installation
--skip-github       Skip GitHub configuration
--skip-jules        Skip Jules API configuration
```

## Sources And Sessions

```bash
juleson sources list
juleson sources get SOURCE_ID

juleson sessions list
juleson sessions status
juleson sessions create SOURCE_ID "Prompt text"
juleson sessions get SESSION_ID
juleson sessions approve SESSION_ID
juleson sessions message SESSION_ID "Follow-up text"
juleson sessions preview SESSION_ID
juleson sessions preview-activity SESSION_ID ACTIVITY_ID
juleson sessions download SESSION_ID OUTPUT_DIR
juleson sessions download-activity SESSION_ID ACTIVITY_ID OUTPUT_DIR

juleson activities list SESSION_ID
juleson activities get SESSION_ID ACTIVITY_ID
```

Session cancel/delete commands are not present in the current CLI.

## Templates

```bash
juleson template list [category]
juleson template show TEMPLATE_NAME
juleson template search QUERY
juleson template create TEMPLATE_NAME CATEGORY DESCRIPTION

juleson execute template TEMPLATE_NAME PROJECT_PATH
juleson execute template-with-params TEMPLATE_NAME PROJECT_PATH key=value
```

## GitHub And Pull Requests

```bash
juleson github login
juleson github status
juleson github repos --limit 20
juleson github current
juleson github search QUERY --limit 30 --sort stars --order desc

juleson pr list --limit 10
juleson pr get SESSION_ID
juleson pr diff SESSION_ID
juleson pr merge SESSION_ID --method squash
```

Pull request flags:

```text
pr list:
  -l, --limit int

pr merge:
  -m, --method string           merge, squash, or rebase
  -c, --commit-message string   custom merge or squash message
```

## GitHub Actions

```bash
juleson actions workflows list [owner/repo]
juleson actions workflows get WORKFLOW_ID_OR_FILE [owner/repo]
juleson actions workflows trigger WORKFLOW_ID_OR_FILE [owner/repo]

juleson actions runs list [owner/repo]
juleson actions runs get RUN_ID [owner/repo]
juleson actions runs rerun RUN_ID [owner/repo]
juleson actions runs cancel RUN_ID [owner/repo]
juleson actions runs logs RUN_ID [owner/repo]

juleson actions jobs list RUN_ID [owner/repo]
juleson actions jobs get JOB_ID [owner/repo]
juleson actions jobs rerun JOB_ID [owner/repo]
juleson actions jobs logs JOB_ID [owner/repo]

juleson actions artifacts list [owner/repo]
juleson actions artifacts download ARTIFACT_ID [owner/repo]
juleson actions artifacts delete ARTIFACT_ID [owner/repo]

juleson actions cache list [owner/repo]
juleson actions cache delete [owner/repo]
```

Most Actions subcommands accept `--repo owner/repo`. Listing commands also expose
filters such as workflow, status, branch, run ID, cache key, cache ID, and Git ref.

## Agent And Orchestration

```bash
juleson agent execute "Goal" --source SOURCE_ID
juleson agent status

juleson ai-orchestrate "Goal" --source SOURCE_ID --path .
juleson orchestrate api-modernization --source SOURCE_ID
juleson orchestrate microservices-migration --source SOURCE_ID
juleson orchestrate custom workflow.yaml --source SOURCE_ID
```

`agent execute` flags:

```text
--source string
--priority string
--constraint strings
--dry-run
--strictness string
--max-iterations int
```

`ai-orchestrate` flags:

```text
--source string
--path string
--constraint strings
--gemini-model string
--gemini-key string
--max-iterations int
--auto-approve
```

## Project And Git Sync

```bash
juleson analyze [project-path]
juleson init [project-path]
juleson sync [project-path] [remote] --branch main --pull
juleson sync [project-path] [remote] --branch main --push
```

## Development Commands

```bash
juleson dev build [--all|--cli|--mcp] [--race] [--version dev]
juleson dev test [--race] [--cover] [--short] [--run PATTERN]
juleson dev lint [--fix] [--fast] [--timeout 5m]
juleson dev fmt [--gofumpt]
juleson dev clean [--all|--cache|--modcache|--testcache]
juleson dev mod tidy
juleson dev mod download
juleson dev mod verify
juleson dev mod vendor
juleson dev mod graph
juleson dev mod why PACKAGE
juleson dev check
juleson dev install [--path DIR] [--skip-checks]
juleson dev release --version VERSION
```

## Completion

```bash
juleson completion bash
juleson completion zsh
juleson completion fish
juleson completion powershell
```

## Environment Variables

- `JULES_API_KEY`: required for Jules API calls.
- `GITHUB_TOKEN`: required for GitHub commands and GitHub MCP tools.
- `GEMINI_API_KEY`: required for Gemini-backed orchestration.
- `JULES_BASE_URL`: optional Jules API base URL override.
