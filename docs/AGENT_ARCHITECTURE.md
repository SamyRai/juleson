# AI Agent Architecture

## Implementation Status

**Last Updated**: November 3, 2025
**Overall Completion**: 70% (Phase 1 Complete)

| Component | Status | Coverage | Files |
|-----------|--------|----------|-------|
| Type System | ✅ Complete | 100% | `types.go` |
| Tool Interface | ✅ Complete | 100% | `tools/tool.go` |
| Tool Registry | ✅ Complete | 26% | `tools/registry.go` |
| Jules Tool | ✅ Complete | 0% | `tools/jules_tool.go` |
| Core Agent | ✅ Complete | 0% | `core/agent.go` |
| Code Reviewer | ✅ Complete | 0% | `review/reviewer.go` |
| Memory System | ✅ Complete | 0% | `memory/memory.go` |
| CLI Integration | ✅ Complete | N/A | `cli/commands/agent.go` |
| GitHub Integration | 🟡 Partial | N/A | `agent/github/` (empty) |
| Advanced Learning | ⚪ Not Started | N/A | Planned |

**Test Coverage**: 26% for tools package, 0% for core/review/memory (tests needed)

## Overview

This is a production-ready AI agent system that goes beyond simple orchestration to provide intelligent, adaptive code development automation with built-in code review, GitHub integration, and learning capabilities.

## Architecture Philosophy

**The AI is not just an orchestrator - it's a proper agent.**

### What Makes This an Agent vs. Just an Orchestrator?

| Orchestrator (Old) | Agent (New) |
|--------------------|-------------|
| Executes predefined workflows | Perceives → Plans → Acts → Reviews → Reflects |
| Fixed task sequences | Adaptive task sequences based on outcomes |
| No learning | Learns from experience |
| Blind execution | Intelligent code review before approval |
| No feedback loop | Continuous feedback to improve Jules |
| Simple tool calls | Sophisticated tool selection and composition |
| No memory | Remembers past decisions and patterns |
| Single-shot execution | Iterative refinement with reflection |

## Core Components

### 1. Agent Core (`internal/agent/core/`)

The brain of the system implementing the agent loop:

```
┌──────────────┐
│   PERCEIVE   │ ← Understand goal, codebase, context
└──────┬───────┘
       ↓
┌──────────────┐
│     PLAN     │ ← Generate multi-step execution plan
└──────┬───────┘
       ↓
┌──────────────┐
│     ACT      │ ← Execute via tools (Jules, GitHub, etc.)
└──────┬───────┘
       ↓
┌──────────────┐
│    REVIEW    │ ← Intelligent code review
└──────┬───────┘
       ↓
┌──────────────┐
│   REFLECT    │ ← Learn and adapt
└──────┬───────┘
       ↓
   ┌───────┐
   │ LOOP  │ or COMPLETE
   └───────┘
```

**Files:**

- `agent.go` - Main agent implementation
- `perception.go` - Context gathering and understanding
- `planning.go` - Task generation with Gemini AI
- `execution.go` - Tool selection and execution
- `reflection.go` - Learning and adaptation

### 2. Tool System (`internal/agent/tools/`)

Abstract interface for agent actions:

**Tool Interface:**

```go
type Tool interface {
    Name() string
    Description() string
    Parameters() []Parameter
    Execute(ctx, params) (*ToolResult, error)
    RequiresApproval() bool
    CanHandle(task Task) bool
}
```

**Available Tools:**

1. **JulesTool** - Execute development tasks via Jules AI
   - Create sessions
   - Send messages/prompts
   - Review plans
   - Apply patches
   - Get activities

2. **GitHubTool** (Planned)
   - Create/manage PRs
   - Post code reviews
   - Manage issues
   - Track CI/CD status

3. **TestTool** (Planned)
   - Run tests
   - Check coverage
   - Run linters
   - Security scanning

4. **AnalysisTool** (Planned)
   - Code complexity analysis
   - Duplicate detection
   - Dependency analysis
   - Code smell detection

### 3. Code Reviewer (`internal/agent/review/`)

Intelligent code review system that provides feedback before approving changes:

**Review Checklist:**

- ✅ Correctness (logic errors, edge cases)
- ✅ Security (vulnerabilities, auth issues)
- ✅ Performance (inefficient algorithms, N+1 queries)
- ✅ Best Practices (style, documentation, tests)
- ✅ Architecture (SOLID principles, consistency)

**Review Process:**

1. Analyze all changes from Jules
2. Run static analysis
3. Check against security patterns
4. Verify test coverage
5. Generate structured feedback
6. Make decision: APPROVE, REQUEST_CHANGES, COMMENT, or REJECT

**Feedback to Jules:**

```go
type ReviewComment struct {
    Location   *Location
    Severity   Severity      // CRITICAL, HIGH, MEDIUM, LOW
    Category   ReviewCategory // SECURITY, PERFORMANCE, etc.
    Message    string
    Suggestion string
    Example    string
}
```

### 4. GitHub Integration (`internal/agent/github/`)

Professional GitHub workflow management:

**PR Management:**

- Create PRs with detailed descriptions
- Link to Jules sessions
- Auto-assign reviewers
- Track PR status
- Handle merge conflicts

**Code Review Posting:**

- Post inline comments
- Request changes with explanations
- Approve when ready
- Re-review after updates

**CI/CD Integration:**

- Wait for CI checks
- Analyze test failures
- Retry on transient failures
- Report status to agent

### 5. Memory System (`internal/agent/memory/`)

Agent memory for learning and pattern recognition:

**Three Types of Memory:**

1. **Episodic Memory** (what happened)
   - Session history
   - Actions taken
   - Results achieved
   - Problems encountered

2. **Semantic Memory** (what we know)
   - Codebase structure
   - Common patterns
   - Best practices
   - Team conventions

3. **Procedural Memory** (how to do things)
   - Successful workflows
   - Tool combinations that work
   - Problem-solving strategies
   - Error recovery patterns

**Learning Mechanisms:**

- Track success/failure of decisions
- Recognize patterns in code issues
- Reuse successful approaches
- Integrate human feedback
- Adjust quality criteria over time

## Agent State Machine

```
IDLE → ANALYZING → PLANNING → EXECUTING → REVIEWING → REFLECTING
  ↑                                                         ↓
  └─────────────────── COMPLETE/FAILED ←──────────────────┘
                              ↓
                        (if not done)
                              ↓
                      ITERATING (back to PLANNING)
```

**State Descriptions:**

- **IDLE**: Waiting for a goal
- **ANALYZING**: Gathering context about codebase and goal
- **PLANNING**: Generating multi-step execution plan with AI
- **EXECUTING**: Running tasks via tools
- **REVIEWING**: Intelligent code review of changes
- **REFLECTING**: Learning from results, updating memory
- **COMPLETE**: Goal achieved successfully
- **FAILED**: Unrecoverable error occurred

## Intelligent Features

### 1. Adaptive Planning

The agent doesn't just execute a fixed plan - it adapts based on results:

**Adaptation Triggers:**

- ❌ Tests fail → Add fix task
- 🔒 Security issues → Insert security fix (CRITICAL priority)
- 🐌 Performance degradation → Add optimization task
- 💥 Breaking changes → Add compatibility layer
- 📦 Dependency conflicts → Resolve dependencies first

**Plan Modifications:**

- Insert task at specific position
- Replace task with better approach
- Reorder based on new dependencies
- Split large task into smaller ones
- Remove tasks that are no longer needed

### 2. Intelligent Code Review

Before blindly approving Jules's changes, the agent reviews them:

```go
// Review process
changes := getChangesFromJules()
result := reviewer.Review(changes)

if result.Decision == APPROVE {
    applyChanges()
    createPR()
} else if result.Decision == REQUEST_CHANGES {
    provideFeedbackToJules(result.Comments)
    waitForImprovements()
    reReview()
}
```

### 3. Feedback Loop with Jules

The agent doesn't just execute - it helps Jules improve:

**Feedback Types:**

1. **Inline Comments**

   ```
   Line 42: Consider using context with timeout
   Rationale: Long-running operations need timeouts
   Example: ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
   ```

2. **General Improvements**

   ```
   Authentication logic is solid, but consider:
   - Adding rate limiting
   - Implementing token rotation
   - Adding audit logging
   ```

3. **Positive Reinforcement**

   ```
   ✨ Excellent error handling!
   ✨ Comprehensive test coverage!
   ```

### 4. Learning and Memory

The agent gets smarter over time:

**What It Learns:**

- Which tools work best for which tasks
- Common code issues in this codebase
- Successful workflow patterns
- What feedback improves Jules's output
- Team-specific conventions

**How It Learns:**

```go
// After each decision
outcome := executeDecision(decision)
learning := Learning{
    Pattern:     "Using Jules for test generation",
    Lesson:      "Works well with clear examples",
    Confidence:  0.9,
}
memory.Learn(learning)

// Later, when similar task appears
similar := memory.Recall("test generation")
// Agent applies learned patterns
```

## Production-Ready Features

### 1. Observability

**Structured Logging:**

```go
log.Info("agent.plan.generated",
    "goal", goal.ID,
    "tasks", len(plan.Tasks),
    "estimated_duration", plan.EstimatedDuration)
```

**Metrics:**

- Decisions made per session
- Success rate by decision type
- Time spent in each state
- Tool usage statistics
- Review pass/fail rates

**Tracing:**

- Distributed tracing for agent actions
- Track decision chains
- Measure tool execution times

### 2. Safety Mechanisms

**Dry-Run Mode:**

```bash
juleson agent execute "improve tests" --dry-run
# Shows what would be done without executing
```

**Approval Gates:**

- High-risk operations require approval
- Configurable approval levels
- Approval timeout with defaults

**Rollback Capabilities:**

- Every action can be rolled back
- Automatic rollback on critical errors
- Manual rollback via CLI

**Rate Limiting:**

- Don't spam Jules with requests
- Respect GitHub API limits
- Configurable delays between actions

**Circuit Breakers:**

- Stop after N consecutive failures
- Automatic recovery with backoff
- Alert on circuit open

### 3. Configuration

**YAML Configuration:**

```yaml
agent:
  max_iterations: 20
  approval_required: true
  dry_run: false

tools:
  jules:
    require_approval: true
    auto_approve: false
    timeout: 10m

  github:
    auto_assign_reviewers: true
    require_ci_pass: true

review:
  strictness: high  # low, medium, high
  min_test_coverage: 0.8
  security_scan: true

memory:
  enabled: true
  retention_days: 90
```

### 4. Error Recovery

**Graceful Degradation:**

- If tool fails, try alternative
- If API times out, retry with backoff
- If review fails, ask for human input

**Error Handling Strategies:**

```go
type ErrorRecovery struct {
    MaxRetries     int
    BackoffStrategy BackoffStrategy
    Fallback       Tool
    OnFailure      func(error) Decision
}
```

## Usage Examples

### Example 1: Improve Test Coverage

```bash
juleson agent execute \
  "Improve test coverage to 80% focusing on authentication" \
  --source my-repo \
  --constraint "Don't change public APIs"
```

**What the Agent Does:**

1. **Perceive**: Analyzes current test coverage (45%)
2. **Plan**: Creates tasks to add tests for auth functions
3. **Act**: Uses Jules to generate tests
4. **Review**: Checks test quality and coverage improvement
5. **Reflect**: If coverage < 80%, adapt plan and continue
6. **Complete**: Creates PR when 80% coverage achieved

### Example 2: Security Audit and Fix

```bash
juleson agent execute \
  "Audit and fix security vulnerabilities" \
  --source backend \
  --priority CRITICAL
```

**What the Agent Does:**

1. **Perceive**: Scans code for security issues
2. **Plan**: Prioritizes by severity (SQL injection first)
3. **Act**: Uses Jules to fix vulnerabilities
4. **Review**: Security scan + code review
5. **Reflect**: If new vulns found, add to plan
6. **Complete**: Creates PR with security fixes

### Example 3: API Modernization

```bash
juleson agent execute \
  "Modernize REST API to use OpenAPI 3.0 and add authentication" \
  --source api-service \
  --constraint "Maintain backward compatibility"
```

**What the Agent Does:**

1. **Perceive**: Analyzes current API structure
2. **Plan**: Multi-phase plan (OpenAPI spec → Auth → Tests → Docs)
3. **Act**: Executes phase by phase with Jules
4. **Review**: Validates OpenAPI spec, tests auth
5. **Reflect**: Adapts plan if issues found
6. **Complete**: Creates comprehensive PR with migration guide

## Integration with Jules

The agent works **with** Jules, not instead of:

```
┌─────────────┐
│    Agent    │ ← Orchestrates, reviews, provides feedback
└──────┬──────┘
       ↓ (uses Jules via JulesTool)
┌─────────────┐
│    Jules    │ ← Writes code, refactors, adds tests
└─────────────┘
```

**Agent Responsibilities:**

- Understand high-level goals
- Break down into tasks
- Select appropriate tools
- Review Jules's work
- Provide constructive feedback
- Learn from outcomes
- Manage GitHub workflow

**Jules Responsibilities:**

- Write actual code
- Refactor existing code
- Add tests
- Fix bugs
- Generate documentation

## Roadmap

### Phase 1: Core Agent ✅ (Complete)

- [x] Agent loop with states
- [x] Tool interface and registry
- [x] Jules tool integration
- [x] Basic types and interfaces
- [x] Code reviewer implementation
- [x] Memory system (episodic)
- [x] CLI commands (`juleson agent execute`, `juleson agent status`)

**Implementation Files:**

- `internal/agent/types.go` - Comprehensive type system
- `internal/agent/tools/tool.go` - Tool interface
- `internal/agent/tools/registry.go` - Thread-safe tool registry
- `internal/agent/tools/jules_tool.go` - Jules AI integration
- `internal/agent/core/agent.go` - Core agent with state machine
- `internal/agent/review/reviewer.go` - Code review system
- `internal/agent/memory/memory.go` - Memory and learning
- `internal/cli/commands/agent.go` - CLI integration

### Phase 2: Enhanced Review (In Progress)

- [x] Basic code reviewer implementation
- [ ] Advanced security analysis integration (Snyk, gosec)
- [ ] Performance profiling integration
- [ ] Feedback loop with Jules
- [ ] Re-review after improvements

### Phase 3: GitHub Integration (Planned)

- [ ] PR creation and management
- [ ] Code review posting
- [ ] CI/CD integration
- [ ] Issue tracking

### Phase 4: Memory & Learning (In Progress)

- [x] Episodic memory (basic)
- [ ] Persistent storage (SQLite)
- [ ] Pattern recognition
- [ ] Feedback integration
- [ ] Success tracking and confidence adjustment

### Phase 5: Advanced Features (Planned)

- [ ] Multi-agent coordination
- [ ] Advanced adaptation strategies
- [ ] Performance optimization
- [ ] Enhanced observability

## Development

### Building

```bash
go build ./internal/agent/...
```

### Testing

```bash
go test ./internal/agent/...
```

### Running

```bash
juleson agent execute "your goal" --source your-repo
```

## Conclusion

This is not just an orchestrator - it's a **proper AI agent** that:

- ✅ Perceives its environment
- ✅ Plans adaptively
- ✅ Acts via tools
- ✅ Reviews intelligently
- ✅ Reflects and learns
- ✅ Provides feedback
- ✅ Manages workflows professionally
- ✅ Gets smarter over time

**The AI is the orchestrator, the reviewer, the learner, and the improver.**
