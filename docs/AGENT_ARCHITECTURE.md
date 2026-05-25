# Agent Architecture

The agent package coordinates goals, planning, tool execution, review, memory,
retry behavior, checkpoints, telemetry, and validation.

## Package Layout

- `internal/agent/types.go`: goal, state, result, and task types.
- `internal/agent/core/agent.go`: agent loop and state transitions.
- `internal/agent/core/planner.go`: Gemini-backed planning support.
- `internal/agent/core/retry.go`: retry, backoff, circuit breaker, and rate limit helpers.
- `internal/agent/core/checkpoint.go`: checkpoint persistence and resume support.
- `internal/agent/core/telemetry.go`: in-process metrics and trace spans.
- `internal/agent/core/validator.go`: goal constraint validation.
- `internal/agent/tools`: tool interface, registry, Jules tool, Docker tool, and analyzer tool.
- `internal/agent/review`: code review support.
- `internal/agent/memory`: memory storage.

## Agent Loop

The core loop follows these stages:

1. Perceive the goal and context.
2. Plan the task sequence.
3. Execute tasks through registered tools.
4. Review the result.
5. Reflect and update memory.

The CLI entrypoint is:

```bash
juleson agent execute "Goal" --source SOURCE_ID
juleson agent status
```

Useful flags:

```text
--source string
--priority string
--constraint strings
--dry-run
--strictness string
--max-iterations int
```

## Safety Defaults

- Use `--dry-run` to inspect planned behavior.
- Pass constraints explicitly with `--constraint`.
- Keep source IDs explicit for session-backed work.
- Prefer small goals that map to one reviewable change.

## Current Limits

- GitHub-specific agent package code is not split into `internal/agent/github`.
- Some packages still need broader unit coverage.
- Review and memory behavior should be treated as evolving pre-1.0 behavior.
