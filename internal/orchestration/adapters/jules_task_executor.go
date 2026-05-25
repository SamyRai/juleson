package adapters

import (
	"context"
	"fmt"
	"time"

	"github.com/SamyRai/juleson/internal/orchestration/domain"
	"github.com/SamyRai/juleson/internal/orchestration/ports"
)

const (
	defaultSourceListLimit = 100
	defaultBranchName      = "main"
	defaultAutomationMode  = "AUTO_CREATE_PR"
)

type JulesTaskExecutor struct {
	gateway       ports.SessionGateway
	sourceMatcher ports.SourceMatcher
}

func NewJulesTaskExecutor(gateway ports.SessionGateway, sourceMatcher ports.SourceMatcher) *JulesTaskExecutor {
	return &JulesTaskExecutor{
		gateway:       gateway,
		sourceMatcher: sourceMatcher,
	}
}

func (e *JulesTaskExecutor) ExecuteTask(ctx context.Context, task domain.Task, execution domain.ExecutionContext) (*domain.TaskResult, error) {
	start := time.Now()
	result := &domain.TaskResult{
		TaskID:    task.ID,
		TaskName:  task.Name,
		TaskType:  task.Type,
		StartTime: start,
		Tool:      firstNonEmpty(task.Tool, "jules"),
		Metrics:   map[string]any{},
	}
	result.Metrics["dry_run"] = execution.DryRun

	title := taskTitle(task)
	if execution.DryRun {
		result.Output = fmt.Sprintf("Dry run: would create or reuse Jules session %q", title)
		result.Success = true
		result.Metrics["session_action"] = "dry_run"
		result.EndTime = time.Now()
		result.Duration = result.EndTime.Sub(start)
		return result, nil
	}
	if e.gateway == nil {
		err := fmt.Errorf("session gateway is required")
		result.Metrics["session_action"] = "failed_safety_check"
		return resultWithError(result, err), err
	}

	reusable, err := e.gateway.FindReusableSession(ctx, title)
	if err != nil {
		return resultWithError(result, err), err
	}
	if reusable != nil {
		result.SessionID = reusable.ID
		result.Output = fmt.Sprintf("Reused existing session: %s", reusable.URL)
		result.Success = true
		result.Metrics["session_action"] = "reused"
		result.EndTime = time.Now()
		result.Duration = result.EndTime.Sub(start)
		return result, nil
	}

	source, err := e.resolveSource(ctx, execution)
	if err != nil {
		return resultWithError(result, err), err
	}
	branch := firstNonEmpty(execution.Goal.Context.Branch, defaultBranchName)
	requirePlanApproval := requirePlanApproval(execution.ApprovalPolicy)
	session, err := e.gateway.CreateSession(ctx, domain.SessionRequest{
		Prompt:              firstNonEmpty(task.Prompt, task.Description),
		Title:               title,
		Source:              *source,
		Branch:              branch,
		RequirePlanApproval: requirePlanApproval,
		AutomationMode:      defaultAutomationMode,
	})
	if err != nil {
		return resultWithError(result, err), err
	}

	result.SessionID = session.ID
	result.Output = fmt.Sprintf("Session created: %s", session.URL)
	result.Success = true
	result.Metrics["session_action"] = "created"
	result.Metrics["require_plan_approval"] = requirePlanApproval
	result.EndTime = time.Now()
	result.Duration = result.EndTime.Sub(start)
	return result, nil
}

func requirePlanApproval(policy domain.ApprovalPolicy) bool {
	if policy.AutoApprove {
		return false
	}
	return true
}

func (e *JulesTaskExecutor) resolveSource(ctx context.Context, execution domain.ExecutionContext) (*domain.Source, error) {
	if execution.Goal.Context.SourceID != "" {
		return &domain.Source{Name: execution.Goal.Context.SourceID, ID: execution.Goal.Context.SourceID}, nil
	}
	sources, err := e.gateway.ListSources(ctx, defaultSourceListLimit)
	if err != nil {
		return nil, err
	}
	project := domain.ProjectContext{}
	if execution.Project != nil {
		project = *execution.Project
	}
	if project.Values == nil {
		project.Values = map[string]string{}
	}
	if execution.Goal.Context.Repository != "" {
		project.Values["repository"] = execution.Goal.Context.Repository
	}
	if e.sourceMatcher == nil {
		if len(sources) == 0 {
			return nil, fmt.Errorf("no sources available")
		}
		return &sources[0], nil
	}
	return e.sourceMatcher.MatchSource(ctx, project, sources)
}

func taskTitle(task domain.Task) string {
	taskType := firstNonEmpty(task.Type, "jules")
	description := firstNonEmpty(task.Description, task.Name)
	return fmt.Sprintf("Execute %s task: %s", taskType, description)
}

func resultWithError(result *domain.TaskResult, err error) *domain.TaskResult {
	result.Success = false
	result.Error = err
	result.EndTime = time.Now()
	result.Duration = result.EndTime.Sub(result.StartTime)
	return result
}
