package adapters

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/SamyRai/juleson/internal/llm"
	"github.com/SamyRai/juleson/internal/orchestration/domain"
)

type LLMPlanner struct {
	llmProvider llm.Provider
}

func NewLLMPlanner(provider llm.Provider) *LLMPlanner {
	return &LLMPlanner{llmProvider: provider}
}

func (p *LLMPlanner) Plan(ctx context.Context, goal domain.Goal, project *domain.ProjectContext) (*domain.Plan, error) {
	if p.llmProvider == nil {
		return fallbackPlan(goal), nil
	}
	response, err := p.llmProvider.GenerateContent(ctx, llm.Request{Prompt: buildPlanningPrompt(goal, project)})
	if err != nil {
		return nil, err
	}
	tasks, err := parseTasks(response)
	if err != nil {
		return nil, fmt.Errorf("parse Gemini plan: %w", err)
	}
	return &domain.Plan{
		ID:        goal.ID,
		Goal:      goal,
		Tasks:     tasks,
		Reasoning: response.Text,
		CreatedAt: time.Now(),
	}, nil
}

func (p *LLMPlanner) AdaptPlan(ctx context.Context, execution domain.ExecutionContext, reason string) (*domain.Plan, error) {
	if execution.Goal.Description == "" {
		return nil, fmt.Errorf("goal description cannot be empty")
	}
	goal := execution.Goal
	if p.llmProvider == nil {
		return fallbackPlan(goal), nil
	}
	prompt := fmt.Sprintf("Adapt this plan for goal %q because: %s\nCompleted tasks: %d", goal.Description, reason, len(execution.Completed))
	response, err := p.llmProvider.GenerateContent(ctx, llm.Request{Prompt: prompt})
	if err != nil {
		return nil, err
	}
	tasks, err := parseTasks(response)
	if err != nil {
		return nil, fmt.Errorf("parse Gemini adapted plan: %w", err)
	}
	return &domain.Plan{
		ID:        goal.ID,
		Goal:      goal,
		Tasks:     tasks,
		Reasoning: response.Text,
		CreatedAt: time.Now(),
	}, nil
}

type LLMDecisionMaker struct {
	llmProvider llm.Provider
}

func NewLLMDecisionMaker(provider llm.Provider) *LLMDecisionMaker {
	return &LLMDecisionMaker{llmProvider: provider}
}

func (d *LLMDecisionMaker) Decide(ctx context.Context, execution domain.ExecutionContext) (*domain.Decision, error) {
	if d.llmProvider == nil {
		return fallbackDecision(execution), nil
	}
	response, err := d.llmProvider.GenerateContent(ctx, llm.Request{Prompt: buildDecisionPrompt(execution)})
	if err != nil {
		return nil, err
	}
	decision, err := parseDecision(response)
	if err != nil {
		return nil, fmt.Errorf("parse Gemini decision: %w", err)
	}
	decision.Timestamp = time.Now()
	return decision, nil
}

func buildPlanningPrompt(goal domain.Goal, project *domain.ProjectContext) string {
	var projectSummary string
	if project != nil {
		projectSummary = fmt.Sprintf("Project: %s\nLanguages: %s\nArchitecture: %s\nComplexity: %s",
			project.ProjectName,
			strings.Join(project.Languages, ", "),
			project.Architecture,
			project.Complexity,
		)
	}
	return fmt.Sprintf(`Create an execution plan for this software engineering goal.

Goal: %s
Constraints: %s
%s

Return JSON in this shape:
{"tasks":[{"id":"task-id","name":"Task name","description":"What to do","prompt":"Prompt for execution","tool":"jules","dependencies":[]}]}`,
		goal.Description,
		strings.Join(goal.Constraints, ", "),
		projectSummary,
	)
}

func buildDecisionPrompt(execution domain.ExecutionContext) string {
	return fmt.Sprintf(`Choose the next orchestration decision for goal %q.
Completed tasks: %d
Total planned tasks: %d

Return JSON: {"decision_type":"next_task|review_needed|adapt_plan|complete|abort","task_id":"optional","reasoning":"why","confidence":0.8}`,
		execution.Goal.Description,
		len(execution.Completed),
		len(execution.Plan.Tasks),
	)
}

func fallbackPlan(goal domain.Goal) *domain.Plan {
	return &domain.Plan{
		ID:   goal.ID,
		Goal: goal,
		Tasks: []domain.Task{{
			ID:          "task-1",
			Name:        "Execute goal",
			Description: goal.Description,
			Prompt:      goal.Description,
			Priority:    goal.Priority,
			State:       domain.TaskStatePending,
		}},
		CreatedAt: time.Now(),
	}
}

func fallbackDecision(execution domain.ExecutionContext) *domain.Decision {
	if execution.Plan == nil || len(execution.Completed) >= len(execution.Plan.Tasks) {
		return &domain.Decision{
			Timestamp:  time.Now(),
			Type:       domain.DecisionTypeComplete,
			Reasoning:  "All planned tasks are complete",
			Confidence: 1,
		}
	}
	task := execution.Plan.Tasks[len(execution.Completed)]
	return &domain.Decision{
		Timestamp:  time.Now(),
		Type:       domain.DecisionTypeNextTask,
		TaskID:     firstNonEmpty(task.ID, task.Name),
		Reasoning:  "Execute the next pending task",
		Confidence: 1,
	}
}

func parseTasks(response *llm.Response) ([]domain.Task, error) {
	var wrapper struct {
		Tasks []struct {
			ID           string   `json:"id"`
			Name         string   `json:"name"`
			Description  string   `json:"description"`
			Prompt       string   `json:"prompt"`
			Tool         string   `json:"tool"`
			Dependencies []string `json:"dependencies"`
		} `json:"tasks"`
	}
	if err := json.Unmarshal([]byte(extractJSON(response.Text)), &wrapper); err != nil {
		return nil, err
	}
	if len(wrapper.Tasks) == 0 {
		return nil, fmt.Errorf("Gemini response did not include any tasks")
	}
	tasks := make([]domain.Task, 0, len(wrapper.Tasks))
	for i, task := range wrapper.Tasks {
		id := task.ID
		if id == "" {
			id = fmt.Sprintf("task-%d", i+1)
		}
		tasks = append(tasks, domain.Task{
			ID:           id,
			Name:         firstNonEmpty(task.Name, id),
			Description:  task.Description,
			Prompt:       firstNonEmpty(task.Prompt, task.Description),
			Tool:         task.Tool,
			Dependencies: task.Dependencies,
			State:        domain.TaskStatePending,
		})
	}
	return tasks, nil
}

func parseDecision(response *llm.Response) (*domain.Decision, error) {
	var payload struct {
		DecisionType string   `json:"decision_type"`
		TaskID       string   `json:"task_id"`
		Reasoning    string   `json:"reasoning"`
		Action       string   `json:"action"`
		Confidence   float64  `json:"confidence"`
		NextSteps    []string `json:"next_steps"`
	}
	if err := json.Unmarshal([]byte(extractJSON(response.Text)), &payload); err != nil {
		return nil, err
	}
	decisionType := normalizeDecisionType(payload.DecisionType)
	if decisionType == "" {
		return nil, fmt.Errorf("unsupported decision type %q", payload.DecisionType)
	}
	return &domain.Decision{
		Type:       decisionType,
		TaskID:     payload.TaskID,
		Reasoning:  payload.Reasoning,
		Action:     payload.Action,
		Confidence: payload.Confidence,
		NextSteps:  payload.NextSteps,
	}, nil
}

func normalizeDecisionType(value string) domain.DecisionType {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case "next_task", "next task":
		return domain.DecisionTypeNextTask
	case "review_needed", "review needed":
		return domain.DecisionTypeReviewNeeded
	case "adapt_plan", "adapt plan":
		return domain.DecisionTypeAdaptPlan
	case "complete":
		return domain.DecisionTypeComplete
	case "abort":
		return domain.DecisionTypeAbort
	default:
		return ""
	}
}

func extractJSON(value string) string {
	start := strings.Index(value, "{")
	end := strings.LastIndex(value, "}")
	if start >= 0 && end > start {
		return value[start : end+1]
	}
	return value
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if value != "" {
			return value
		}
	}
	return ""
}
