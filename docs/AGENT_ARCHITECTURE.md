# Agent Architecture

Agent orchestration is now owned by `internal/orchestration`. The older
`internal/agent` package remains as compatibility surface for agent-specific
types, tools, review, and memory code that has not yet been retired.

The extraction target is `internal/orchestration`: domain types, ports,
application runners, and concrete adapters. It is structured so the orchestration
core can move into its own module without carrying CLI, MCP, Jules SDK,
Gemini, analyzer, filesystem, or template implementation details with it.

## Package Layout

- `internal/orchestration/domain`: pure goal, plan, task, decision, workflow,
  progress, result, review, checkpoint, session, and project types.
- `internal/orchestration/ports`: consumer-owned interfaces used by the
  application layer.
- `internal/orchestration/app`: `AgentRunner`, `TemplateRunner`,
  `SessionWorkflowRunner`, `AIWorkflowRunner`, and a small private graph runner
  used to make agent control flow explicit.
- `internal/orchestration/adapters`: Jules, Gemini, analyzer, template, source
  matching, clock, progress, and tool execution adapters.
- `internal/orchestration`: runtime construction facade exposed as `orchestration.NewRuntime(deps)`.
- `internal/agent/types.go`: legacy goal, state, result, and task types.
- `internal/agent/core/agent.go`: legacy agent loop and state transitions.
- `internal/agent/core/planner.go`: Gemini-backed planning support.
- `internal/agent/core/retry.go`: retry, backoff, circuit breaker, and rate limit helpers.
- `internal/agent/core/checkpoint.go`: checkpoint persistence and resume support.
- `internal/agent/core/telemetry.go`: in-process metrics and trace spans.
- `internal/agent/core/validator.go`: goal constraint validation.
- `internal/agent/tools`: tool interface, registry, Jules tool, Docker tool, and analyzer tool.
- `internal/agent/review`: code review support.
- `internal/agent/memory`: memory storage.

## Extraction Boundary

Import rules are enforced by `internal/architecture` tests:

- `internal/orchestration/domain` imports only the standard library.
- `internal/orchestration/ports` imports only the standard library and `domain`.
- `internal/orchestration/app` imports only the standard library, `domain`, and `ports`.
- Concrete systems live in `internal/orchestration/adapters`.

The service container is the composition root for runtime construction. CLI and
MCP paths should call the runtime instead of manually constructing legacy
`core.NewAgent`, `automation.NewEngine`, `automation.NewAIOrchestrator`, or
`automation.NewSessionOrchestrator` instances. CLI orchestration command
constructors receive runtime factories from the composition root; active
orchestration paths should use `internal/orchestration/domain` and presentation
DTOs instead of legacy `internal/automation` workflow or execution-result types.

## Agent Loop

The application runner uses a simple in-process graph runner inspired by
LangGraph's node/state routing model. It is intentionally not a full LangGraph
port: there are no reducers, nested graphs, streaming, interrupts, or external
graph dependencies.

The standard agent graph follows these stages:

1. Perceive the goal and context.
2. Plan the task sequence.
3. Execute tasks through registered tools.
4. Review the result.
5. Reflect and update memory.

The AI workflow runner uses the same graph primitive with a decision loop:
analyze, plan, decide, route the decision, then execute, review, adapt, complete,
or abort.

`TemplateRunner` also uses the private graph for its sequential template
execution path: load template, analyze project, render and order tasks, execute,
write outputs, and complete. `SessionWorkflowRunner` intentionally remains
imperative for now because session phases support parallel execution and
`ContinueOnError`; modeling those correctly would require graph features the
private runner deliberately does not have.

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

- Use `--dry-run` to inspect analyzed and scheduled task plans without creating
  or reusing Jules sessions.
- Pass constraints explicitly with `--constraint`.
- Keep source IDs explicit for session-backed work.
- Agent-created Jules sessions require plan approval by default.
- `--strictness` is validated before orchestration starts and is carried through
  execution context for review adapters.
- Runner checkpoints are saved through `ports.CheckpointStore` after planning,
  after each task, and at final success or failure when a store is configured.
  AI workflows also checkpoint decisions and review passes.
- The default service container wires a local JSON checkpoint store using
  `automation.checkpoint_path`.
- Prefer small goals that map to one reviewable change.

## Current Limits

- Some legacy `internal/agent` and `internal/automation` APIs remain for tests
  and older callers.
- Review, memory, checkpoint resume, and telemetry adapters need follow-up extraction from legacy packages.
- MCP and CLI adapters should continue moving toward runtime-only construction
  as legacy commands are retired.
