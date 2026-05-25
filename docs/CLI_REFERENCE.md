# CLI Reference

This reference reflects the current Cobra command tree from `cmd/juleson`.
Core Jules commands (`setup`, `sessions`, `sources`, `activities`,
`completion`, and `version`) are composed through the internal Jules CLI package;
Juleson-specific extensions such as GitHub, templates, agent, and dev commands
are registered by the application layer.

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
| `config` | Manage Juleson configuration |
| `dev` | Build, test, lint, format, and release helpers |
| `execute` | Execute templates and automation tasks |
| `github` | Manage GitHub integration |
| `init` | Initialize a project for Jules automation |
| `orchestrate` | Run predefined multi-task Jules workflows |
| `official` | Bridge to the official Jules CLI when installed |
| `pr` | Manage PRs created by Jules sessions |
| `sessions` | Manage Jules sessions |
| `setup` | Run first-time setup |
| `sources` | Manage Jules sources |
| `sync` | Sync a project with a remote repository |
| `template` | Manage templates |
| `version` | Print version information |

## Config And Setup

```bash
juleson config validate
juleson setup [flags]
```

`config validate` validates the effective configuration and checks for hard errors
(e.g., invalid port or concurrency limits) and reports missing credentials as warnings.
It never prints API keys or other secrets to output.

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
juleson sessions create SOURCE_ID "Prompt text" --require-plan-approval
juleson sessions create . --prompt-file task.md --title "Fix failing tests"
juleson sessions create --no-source "Prompt text"
juleson sessions batch SOURCE_ID task.md --parallel 3 --batch-id batch-20260525 --group-title "Fix CI"
juleson sessions watch SESSION_ID --follow-activities --since 2026-05-25T10:00:00Z --cursor-output .juleson.cursor
juleson sessions watch SESSION_ID --wake-on-status-change --initial-state PLANNING
juleson sessions watch SESSION_ID --wake-on-agent-message --since 2026-05-25T10:00:00Z
juleson sessions get SESSION_ID
juleson sessions approve SESSION_ID
juleson sessions message SESSION_ID "Follow-up text"
juleson sessions apply SESSION_ID PROJECT_PATH
juleson sessions apply SESSION_ID PROJECT_PATH --activity-id ACTIVITY_ID --artifact-index 0
juleson sessions apply SESSION_ID PROJECT_PATH --confirm --allow-base-mismatch
juleson sessions artifacts list SESSION_ID
juleson sessions outputs SESSION_ID
juleson sessions delete SESSION_ID --force
juleson sessions preview SESSION_ID
juleson sessions preview-activity SESSION_ID ACTIVITY_ID
juleson sessions download SESSION_ID OUTPUT_DIR
juleson sessions download-activity SESSION_ID ACTIVITY_ID OUTPUT_DIR

juleson activities list SESSION_ID
juleson activities list SESSION_ID --since 2026-05-25T10:00:00Z --cursor-output .juleson.cursor
juleson activities get SESSION_ID ACTIVITY_ID
juleson official remote new --parallel 3
juleson official remote pull SESSION_ID
```

`sessions create` accepts either `github/owner/repo` or
`sources/github/owner/repo`. `--no-source` creates a repoless Jules session by
omitting `sourceContext`. Source-backed sessions also accept `--title`,
`--starting-branch`, `--require-plan-approval`, `--automation-mode`, and
`--prompt-file`. If `--starting-branch` is omitted for a GitHub source, Juleson
reads the connected source metadata and uses the default branch. Passing `.` as
the source asks Juleson to infer the connected Jules source from the local git
`origin` remote; ambiguous matches fail with the candidate source names.

`sessions batch` creates 1-5 parallel sessions for one source and prompt or task
file. Batch sessions require plan approval by default and include a `batch_id`,
optional `group_title`, and run index in each prompt because the REST API has no
documented bulk-create endpoint.

`sessions watch` polls until a session completes, fails, needs user action, or
surfaces session outputs. With `--follow-activities`, it uses the activity
`createTime` cursor for client-side filtering and prints the next cursor for
resumable watches. Use
`--wake-on-status-change` to stop on the next state transition from
`--initial-state` or the first observed state. Use `--wake-on-agent-message` to
stop when Jules posts a new agent message after `--since`; without `--since`,
the first poll establishes the activity baseline.
When a completed session has no pull request output and only empty changeset
artifacts, watch reports that no retrievable deliverable was produced instead
of directing operators to apply an empty patch.

`sessions apply` dry-runs by default. Use `--confirm` to apply patches; dirty
worktrees are blocked unless `--allow-dirty` is passed. `--activity-id` and
`--artifact-index` apply one changeset at a time. If an artifact includes
`baseCommitId`, dry-runs warn on mismatch and real apply blocks unless
`--allow-base-mismatch` is passed.

`sessions artifacts list` prints an artifact manifest with activity ID, artifact
index, type, changed files, base commit, suggested commit message, media MIME
type, and bash exit code. `sessions outputs` prints documented session outputs
such as Jules-created pull requests. Completed sessions can validly expose no
retrievable deliverables; in that case artifact manifests show empty changesets
and outputs report that no supported documented payloads were found.

`sessions delete` calls the Jules API delete endpoint. Without `--force`, it
asks for the exact session ID before deleting. Session cancel is not exposed by
the Jules API v1alpha reference used by this project.

`official remote ...` and `official tui ...` hand off to the official `jules`
binary when it is installed. They are optional parity bridges for exact
`remote new --parallel`, `remote pull`, and TUI diff-review behavior; REST
commands remain the default Juleson path.

`sessions preview` and `sessions download` use documented activity artifacts:
git patches from `changeSet`, command output from `bashOutput`, and decoded
base64 media from `media`.

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

`--dry-run` analyzes and plans the requested goal, prints the planned tasks, and
does not create, reuse, or mutate Jules sessions. Real `agent execute` runs
create Jules sessions with plan approval required by default. `--strictness`
accepts `low`, `medium`, or `high`; invalid values fail before orchestration
starts. `agent status` reports configured runtime capabilities, including Jules,
Gemini, review, memory, checkpointing, and dry-run planning availability.

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

- `JULES_API_KEY`: accepted directly by config loading and required for Jules API calls.
- `GITHUB_TOKEN`: read by `juleson setup --non-interactive`, then saved to config.
- `GEMINI_API_KEY`: read by `juleson ai-orchestrate --gemini-key` fallback. For
  MCP Gemini tools, save `gemini.api_key` in `juleson.yaml`.

Other settings should be configured in `juleson.yaml`.
