package app

import "github.com/SamyRai/juleson/internal/orchestration/ports"

type RuntimeDeps struct {
	ProjectAnalyzer ports.ProjectAnalyzer
	Planner         ports.Planner
	DecisionMaker   ports.DecisionMaker
	TaskExecutor    ports.TaskExecutor
	Reviewer        ports.Reviewer
	MemoryStore     ports.MemoryStore
	CheckpointStore ports.CheckpointStore
	SessionGateway  ports.SessionGateway
	TemplateStore   ports.TemplateStore
	PromptRenderer  ports.PromptRenderer
	SourceMatcher   ports.SourceMatcher
	OutputWriter    ports.OutputWriter
	ProgressSink    ports.ProgressSink
	Clock           ports.Clock
}

type Runtime struct {
	deps RuntimeDeps
}

type RuntimeCapabilities struct {
	ProjectAnalysis bool
	Planning        bool
	TaskExecution   bool
	Review          bool
	Memory          bool
	Checkpointing   bool
	DryRunPlanning  bool
}

func NewRuntime(deps RuntimeDeps) *Runtime {
	return &Runtime{deps: deps}
}

func (r *Runtime) Capabilities() RuntimeCapabilities {
	return RuntimeCapabilities{
		ProjectAnalysis: r.deps.ProjectAnalyzer != nil,
		Planning:        r.deps.Planner != nil,
		TaskExecution:   r.deps.TaskExecutor != nil,
		Review:          r.deps.Reviewer != nil,
		Memory:          r.deps.MemoryStore != nil,
		Checkpointing:   r.deps.CheckpointStore != nil,
		DryRunPlanning:  r.deps.Planner != nil,
	}
}

func (r *Runtime) AgentRunner() *AgentRunner {
	return NewAgentRunner(AgentRunnerDeps{
		ProjectAnalyzer: r.deps.ProjectAnalyzer,
		Planner:         r.deps.Planner,
		TaskExecutor:    r.deps.TaskExecutor,
		Reviewer:        r.deps.Reviewer,
		MemoryStore:     r.deps.MemoryStore,
		CheckpointStore: r.deps.CheckpointStore,
		ProgressSink:    r.deps.ProgressSink,
		Clock:           r.deps.Clock,
	})
}

func (r *Runtime) TemplateRunner() *TemplateRunner {
	return NewTemplateRunner(TemplateRunnerDeps{
		ProjectAnalyzer: r.deps.ProjectAnalyzer,
		TemplateStore:   r.deps.TemplateStore,
		PromptRenderer:  r.deps.PromptRenderer,
		TaskExecutor:    r.deps.TaskExecutor,
		OutputWriter:    r.deps.OutputWriter,
		ProgressSink:    r.deps.ProgressSink,
		Clock:           r.deps.Clock,
	})
}

func (r *Runtime) SessionWorkflowRunner() *SessionWorkflowRunner {
	return NewSessionWorkflowRunner(SessionWorkflowRunnerDeps{
		TaskExecutor: r.deps.TaskExecutor,
		ProgressSink: r.deps.ProgressSink,
		Clock:        r.deps.Clock,
	})
}

func (r *Runtime) AIWorkflowRunner() *AIWorkflowRunner {
	return NewAIWorkflowRunner(AIWorkflowRunnerDeps{
		ProjectAnalyzer: r.deps.ProjectAnalyzer,
		Planner:         r.deps.Planner,
		DecisionMaker:   r.deps.DecisionMaker,
		TaskExecutor:    r.deps.TaskExecutor,
		Reviewer:        r.deps.Reviewer,
		MemoryStore:     r.deps.MemoryStore,
		ProgressSink:    r.deps.ProgressSink,
		Clock:           r.deps.Clock,
	})
}
