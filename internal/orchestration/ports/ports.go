package ports

import (
	"context"
	"time"

	"github.com/SamyRai/juleson/internal/orchestration/domain"
)

type ProjectAnalyzer interface {
	AnalyzeProject(ctx context.Context, projectPath string) (*domain.ProjectContext, error)
}

type Planner interface {
	Plan(ctx context.Context, goal domain.Goal, project *domain.ProjectContext) (*domain.Plan, error)
	AdaptPlan(ctx context.Context, execution domain.ExecutionContext, reason string) (*domain.Plan, error)
}

type DecisionMaker interface {
	Decide(ctx context.Context, execution domain.ExecutionContext) (*domain.Decision, error)
}

type TaskExecutor interface {
	ExecuteTask(ctx context.Context, task domain.Task, execution domain.ExecutionContext) (*domain.TaskResult, error)
}

type Reviewer interface {
	Review(ctx context.Context, execution domain.ExecutionContext) (*domain.ReviewResult, error)
}

type MemoryStore interface {
	RecordDecision(ctx context.Context, decision domain.Decision) error
	RecordResult(ctx context.Context, result domain.Result) error
}

type CheckpointStore interface {
	SaveCheckpoint(ctx context.Context, checkpoint domain.Checkpoint) error
	LoadCheckpoint(ctx context.Context, id string) (*domain.Checkpoint, error)
}

type SessionGateway interface {
	ListSources(ctx context.Context, limit int) ([]domain.Source, error)
	FindReusableSession(ctx context.Context, title string) (*domain.Session, error)
	CreateSession(ctx context.Context, request domain.SessionRequest) (*domain.Session, error)
	GetSession(ctx context.Context, sessionID string) (*domain.Session, error)
}

type TemplateStore interface {
	LoadTemplate(ctx context.Context, name string) (*domain.Template, error)
}

type PromptRenderer interface {
	RenderPrompt(ctx context.Context, template string, values map[string]string) (string, error)
}

type SourceMatcher interface {
	MatchSource(ctx context.Context, project domain.ProjectContext, sources []domain.Source) (*domain.Source, error)
}

type OutputWriter interface {
	WriteOutputs(ctx context.Context, template domain.Template, result domain.Result) ([]string, error)
}

type ProgressSink interface {
	ReportProgress(ctx context.Context, progress domain.Progress) error
}

type Clock interface {
	Now() time.Time
	Sleep(ctx context.Context, duration time.Duration) error
}
