# AI Orchestration

AI orchestration uses Gemini to plan and coordinate multi-step development
workflows. It is separate from fixed workflows under `juleson orchestrate`.

## CLI Usage

```bash
juleson ai-orchestrate "Improve test coverage" \
  --source SOURCE_ID \
  --path . \
  --constraint "Do not change public APIs"
```

Flags:

```text
--source string
--path string
--constraint strings
--gemini-model string
--gemini-key string
--max-iterations int
--auto-approve
```

`--source` is required. `--gemini-key` can be omitted when `GEMINI_API_KEY` is set.
`--max-iterations` bounds the AI decision loop. By default Jules sessions require
plan approval; `--auto-approve` disables that approval gate for sessions created
by this command.

## MCP Tools

Gemini-backed MCP tools are registered when a Jules client is available and
`gemini.api_key` is configured in `juleson.yaml`:

- `plan_project_automation`
- `orchestrate_workflow`
- `manage_github_project`
- `synthesize_session_results`

## Flow

1. Analyze the project path.
2. Build a task plan from the requested goal and constraints.
3. Execute one step at a time.
4. Adapt the plan based on results.
5. Stop when the goal is complete or the iteration limit is reached.

When Gemini is configured, orchestration expects structured JSON from planning
and decision prompts. Malformed or unsupported responses fail the run instead of
falling back to a generic task.

## Use Carefully

`--auto-approve` skips manual approval gates. Use it only in repositories where
the workflow and token permissions are acceptable for unattended changes.
