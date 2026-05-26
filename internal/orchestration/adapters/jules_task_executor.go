package adapters

import (
	"context"
	"fmt"
	"sort"
	"strings"
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
		Prompt:              buildJulesPrompt(task, execution),
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

func buildJulesPrompt(task domain.Task, execution domain.ExecutionContext) string {
	var b strings.Builder
	basePrompt := firstNonEmpty(task.Prompt, task.Description, task.Name)
	if basePrompt != "" {
		b.WriteString(strings.TrimSpace(basePrompt))
		b.WriteString("\n\n")
	}

	b.WriteString("## Juleson Context\n")
	writeLine(&b, "Goal", execution.Goal.Description)
	writeList(&b, "Constraints", execution.Goal.Constraints)
	writeLine(&b, "Task", firstNonEmpty(task.Name, task.ID))
	writeLine(&b, "Task type", task.Type)
	writeLine(&b, "Priority", string(task.Priority))
	writeMap(&b, "Task context", task.Context)
	writeGoalContext(&b, execution.Goal.Context)
	writeProjectContext(&b, execution.Project)
	writeCompletedTasks(&b, execution.Completed)
	writeMap(&b, "Execution values", execution.Values)

	b.WriteString("\n## Engineering Guidelines\n")
	for _, guideline := range []string{
		"Make the smallest correct change that satisfies the goal.",
		"Inspect relevant files and usages before editing; follow the existing architecture and style.",
		"Do not touch secrets, credentials, authentication settings, key rotation, or production data unless the task explicitly asks for it.",
		"Maintain backward compatibility unless the goal explicitly allows a breaking change.",
		"Keep side effects explicit and error handling consistent with the surrounding code.",
		"Add or update focused tests when behavior changes, and update existing docs or comments when public behavior changes.",
		"Run the relevant format, lint, or test commands when possible, then report results and any residual risks.",
	} {
		b.WriteString("- ")
		b.WriteString(guideline)
		b.WriteByte('\n')
	}

	return strings.TrimSpace(b.String())
}

func writeGoalContext(b *strings.Builder, context domain.GoalContext) {
	if context.ProjectPath == "" && context.SourceID == "" && context.Repository == "" &&
		context.Branch == "" && len(context.RelatedIssues) == 0 && len(context.RelatedPRs) == 0 &&
		len(context.Values) == 0 {
		return
	}
	b.WriteString("\n### Goal Context\n")
	writeLine(b, "Project path", context.ProjectPath)
	writeLine(b, "Source ID", context.SourceID)
	writeLine(b, "Repository", context.Repository)
	writeLine(b, "Branch", context.Branch)
	writeList(b, "Related issues", context.RelatedIssues)
	writeList(b, "Related PRs", context.RelatedPRs)
	writeMap(b, "Values", context.Values)
}

func writeProjectContext(b *strings.Builder, project *domain.ProjectContext) {
	if project == nil {
		return
	}
	b.WriteString("\n### Project Context\n")
	writeLine(b, "Project path", project.ProjectPath)
	writeLine(b, "Project name", project.ProjectName)
	writeLine(b, "Project type", project.ProjectType)
	writeList(b, "Languages", project.Languages)
	writeList(b, "Frameworks", project.Frameworks)
	writeLine(b, "Architecture", project.Architecture)
	writeLine(b, "Complexity", project.Complexity)
	writeLine(b, "Git status", project.GitStatus)
	writeLine(b, "Branch", project.Branch)
	writeMap(b, "Dependencies", project.Dependencies)
	writeMap(b, "Values", project.Values)
	if project.Quality != nil {
		b.WriteString("- Quality: ")
		b.WriteString(fmt.Sprintf("coverage %.2f, complexity %.2f, maintainability %.2f, security issues %d, code smells %d",
			project.Quality.TestCoverage,
			project.Quality.CodeComplexity,
			project.Quality.Maintainability,
			project.Quality.SecurityIssues,
			project.Quality.CodeSmells,
		))
		b.WriteByte('\n')
	}
}

func writeCompletedTasks(b *strings.Builder, completed []domain.TaskResult) {
	if len(completed) == 0 {
		return
	}
	b.WriteString("\n### Completed Tasks\n")
	for _, task := range completed {
		name := firstNonEmpty(task.TaskName, task.TaskID)
		if name == "" {
			continue
		}
		status := "failed"
		if task.Success {
			status = "succeeded"
		}
		b.WriteString("- ")
		b.WriteString(name)
		b.WriteString(": ")
		b.WriteString(status)
		if task.Output != "" {
			b.WriteString(" - ")
			b.WriteString(task.Output)
		}
		b.WriteByte('\n')
	}
}

func writeLine(b *strings.Builder, label, value string) {
	if value == "" {
		return
	}
	b.WriteString("- ")
	b.WriteString(label)
	b.WriteString(": ")
	b.WriteString(value)
	b.WriteByte('\n')
}

func writeList(b *strings.Builder, label string, values []string) {
	if len(values) == 0 {
		return
	}
	cleaned := make([]string, 0, len(values))
	for _, value := range values {
		if strings.TrimSpace(value) != "" {
			cleaned = append(cleaned, value)
		}
	}
	if len(cleaned) == 0 {
		return
	}
	writeLine(b, label, strings.Join(cleaned, ", "))
}

func writeMap(b *strings.Builder, label string, values map[string]string) {
	if len(values) == 0 {
		return
	}
	keys := make([]string, 0, len(values))
	for key, value := range values {
		if strings.TrimSpace(key) != "" && strings.TrimSpace(value) != "" {
			keys = append(keys, key)
		}
	}
	if len(keys) == 0 {
		return
	}
	sort.Strings(keys)
	parts := make([]string, 0, len(keys))
	for _, key := range keys {
		parts = append(parts, fmt.Sprintf("%s=%s", key, values[key]))
	}
	writeLine(b, label, strings.Join(parts, ", "))
}

func resultWithError(result *domain.TaskResult, err error) *domain.TaskResult {
	result.Success = false
	result.Error = err
	result.EndTime = time.Now()
	result.Duration = result.EndTime.Sub(result.StartTime)
	return result
}
