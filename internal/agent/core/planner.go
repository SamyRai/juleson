package core

import (
	"context"
	"fmt"
	"log/slog"
	"strings"
	"time"

	"github.com/SamyRai/juleson/internal/agent"
	"github.com/SamyRai/juleson/internal/analyzer"
	"github.com/SamyRai/juleson/internal/gemini"
)

// Planner generates execution plans using AI
type Planner struct {
	gemini *gemini.Client
	logger *slog.Logger
}

// NewPlanner creates a new AI-powered planner
func NewPlanner(geminiClient *gemini.Client, logger *slog.Logger) *Planner {
	if geminiClient == nil {
		panic("geminiClient cannot be nil")
	}
	if logger == nil {
		logger = slog.Default()
	}
	return &Planner{
		gemini: geminiClient,
		logger: logger,
	}
}

// GeneratePlan creates a multi-step execution plan using chain-of-thought reasoning
func (p *Planner) GeneratePlan(ctx context.Context, goal agent.Goal, codebaseContext string, projectContext *analyzer.ProjectContext) ([]agent.Task, *agent.Reasoning, error) {
	if goal.Description == "" {
		return nil, nil, fmt.Errorf("cannot generate plan: goal description is empty")
	}

	p.logger.Info("planner.generate_plan.start", "goal", goal.Description)

	// Build context-aware prompt
	prompt := p.buildPlanningPrompt(goal, codebaseContext, projectContext)

	// Use Gemini to generate plan with chain-of-thought
	response, err := p.gemini.Generate(ctx, prompt)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to generate plan: %w", err)
	}

	// Parse response into tasks and reasoning
	tasks, reasoning, err := p.parsePlanResponse(response, goal)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to parse plan: %w", err)
	}

	p.logger.Info("planner.generate_plan.complete", "tasks", len(tasks))
	return tasks, reasoning, nil
}

// AdaptPlan modifies an existing plan based on new information
func (p *Planner) AdaptPlan(ctx context.Context, currentPlan []agent.Task, adaptReason string, feedback string) ([]agent.Task, error) {
	if len(currentPlan) == 0 {
		return nil, fmt.Errorf("cannot adapt plan: currentPlan is empty")
	}
	if adaptReason == "" {
		return nil, fmt.Errorf("cannot adapt plan: adaptReason is empty")
	}

	p.logger.Info("planner.adapt_plan.start", "reason", adaptReason)

	prompt := p.buildAdaptationPrompt(currentPlan, adaptReason, feedback)

	response, err := p.gemini.Generate(ctx, prompt)
	if err != nil {
		return nil, fmt.Errorf("failed to adapt plan: %w", err)
	}

	// Parse adapted plan
	tasks, _, err := p.parsePlanResponse(response, agent.Goal{Description: adaptReason})
	if err != nil {
		return nil, fmt.Errorf("failed to parse adapted plan: %w", err)
	}

	p.logger.Info("planner.adapt_plan.complete", "new_tasks", len(tasks))
	return tasks, nil
}

// buildPlanningPrompt creates a detailed prompt for plan generation
func (p *Planner) buildPlanningPrompt(goal agent.Goal, codebaseContext string, projectContext *analyzer.ProjectContext) string {
	constraints := ""
	if len(goal.Constraints) > 0 {
		constraints = "\n\nCONSTRAINTS:\n" + strings.Join(goal.Constraints, "\n")
	}

	deadline := ""
	if goal.Deadline != nil {
		deadline = fmt.Sprintf("\n\nDEADLINE: %s", goal.Deadline.Format(time.RFC3339))
	}

	projectInfo := ""
	if projectContext != nil {
		projectInfo = fmt.Sprintf(`
PROJECT ANALYSIS:
Name: %s
Type: %s
Languages: %s
Frameworks: %s
Architecture: %s
Complexity: %s
Git Status: %s
Dependencies: %d packages`,
			projectContext.ProjectName,
			projectContext.ProjectType,
			strings.Join(projectContext.Languages, ", "),
			strings.Join(projectContext.Frameworks, ", "),
			projectContext.Architecture,
			projectContext.Complexity,
			projectContext.GitStatus,
			len(projectContext.Dependencies),
		)

		if projectContext.CodeQuality != nil {
			projectInfo += fmt.Sprintf(`
Code Quality:
- Test Coverage: %.1f%%
- Code Complexity: %.1f
- Maintainability: %.1f
- Security Issues: %d
- Code Smells: %d`,
				projectContext.CodeQuality.TestCoverage,
				projectContext.CodeQuality.CodeComplexity,
				projectContext.CodeQuality.Maintainability,
				len(projectContext.CodeQuality.SecurityIssues),
				len(projectContext.CodeQuality.CodeSmells),
			)
		}
	}

	return fmt.Sprintf(`You are an expert AI agent planner. Generate a detailed, step-by-step execution plan to achieve the following goal.

GOAL: %s
PRIORITY: %s%s%s%s

CODEBASE CONTEXT:
%s

INSTRUCTIONS:
1. Use chain-of-thought reasoning to break down the goal
2. Create specific, actionable tasks with clear success criteria
3. Identify dependencies between tasks
4. Consider edge cases and potential issues
5. Recommend appropriate tools for each task
6. Estimate complexity and risk for each task
7. Take into account the project analysis information when planning

OUTPUT FORMAT:
Provide your response in the following format:

REASONING:
[Your chain-of-thought reasoning process]

TASKS:
Task 1: [name]
Description: [detailed description]
Tool: [recommended tool: jules, github, test, analysis]
Dependencies: [comma-separated task numbers, or "none"]
Success Criteria: [clear criteria]
Risk: [LOW/MEDIUM/HIGH]
Estimated Duration: [estimate]

Task 2: ...

Generate a comprehensive plan now.`,
		goal.Description,
		goal.Priority,
		constraints,
		deadline,
		projectInfo,
		codebaseContext,
	)
}

// buildAdaptationPrompt creates a prompt for plan adaptation
func (p *Planner) buildAdaptationPrompt(currentPlan []agent.Task, reason string, feedback string) string {
	planSummary := ""
	for i, task := range currentPlan {
		planSummary += fmt.Sprintf("\nTask %d: %s (Status: %s)", i+1, task.Name, task.State)
	}

	return fmt.Sprintf(`You are adapting an execution plan based on new information.

CURRENT PLAN:%s

ADAPTATION REASON: %s

FEEDBACK: %s

INSTRUCTIONS:
1. Analyze what went wrong or what changed
2. Determine which tasks need to be modified, added, or removed
3. Maintain dependencies and task ordering
4. Provide reasoning for each change

OUTPUT FORMAT:
REASONING:
[Your analysis and reasoning]

ADAPTED TASKS:
Task 1: [name]
Description: [description]
Tool: [tool]
Dependencies: [dependencies]
Success Criteria: [criteria]
Risk: [risk]
Estimated Duration: [duration]

Generate the adapted plan now.`,
		planSummary,
		reason,
		feedback,
	)
}

// parsePlanResponse parses Gemini's response into structured tasks
func (p *Planner) parsePlanResponse(response string, goal agent.Goal) ([]agent.Task, *agent.Reasoning, error) {
	// Extract reasoning section
	reasoning := p.extractReasoning(response)

	// Extract tasks section
	tasksSection := p.extractSection(response, "TASKS:")
	if tasksSection == "" {
		return nil, nil, fmt.Errorf("no TASKS section found in response")
	}

	// Parse individual tasks
	tasks := p.parseTasks(tasksSection, goal)

	return tasks, reasoning, nil
}

// extractReasoning extracts the chain-of-thought reasoning
func (p *Planner) extractReasoning(response string) *agent.Reasoning {
	reasoningText := p.extractSection(response, "REASONING:")

	return &agent.Reasoning{
		ChainOfThought: []string{reasoningText},
		Confidence:     0.8, // Default confidence
		Alternatives:   []string{},
		SelectedPath:   reasoningText,
		Timestamp:      time.Now(),
	}
}

// extractSection extracts a section from the response
func (p *Planner) extractSection(response string, marker string) string {
	startIdx := strings.Index(response, marker)
	if startIdx == -1 {
		return ""
	}

	content := response[startIdx+len(marker):]

	// Find next section marker or end
	nextMarkers := []string{"TASKS:", "REASONING:", "ADAPTED TASKS:"}
	endIdx := len(content)

	for _, nextMarker := range nextMarkers {
		if idx := strings.Index(content, nextMarker); idx != -1 && idx < endIdx {
			endIdx = idx
		}
	}

	return strings.TrimSpace(content[:endIdx])
}

// parseTasks parses the tasks section into structured Task objects
func (p *Planner) parseTasks(tasksSection string, goal agent.Goal) []agent.Task {
	var tasks []agent.Task

	// Split by "Task N:" pattern
	lines := strings.Split(tasksSection, "\n")

	var currentTask *agent.Task
	taskCounter := 0

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		// Check if this is a new task
		if strings.HasPrefix(line, "Task ") && strings.Contains(line, ":") {
			// Save previous task
			if currentTask != nil {
				tasks = append(tasks, *currentTask)
			}

			// Start new task
			taskCounter++
			parts := strings.SplitN(line, ":", 2)
			taskName := strings.TrimSpace(parts[1])

			currentTask = &agent.Task{
				ID:           fmt.Sprintf("task-%d", taskCounter),
				Name:         taskName,
				Priority:     goal.Priority,
				State:        agent.TaskStatePending,
				Context:      make(map[string]interface{}),
				Dependencies: []string{},
			}
			continue
		}

		// Parse task attributes
		if currentTask != nil {
			if strings.HasPrefix(line, "Description:") {
				currentTask.Description = strings.TrimSpace(strings.TrimPrefix(line, "Description:"))
			} else if strings.HasPrefix(line, "Tool:") {
				currentTask.Tool = strings.ToLower(strings.TrimSpace(strings.TrimPrefix(line, "Tool:")))
			} else if strings.HasPrefix(line, "Dependencies:") {
				deps := strings.TrimSpace(strings.TrimPrefix(line, "Dependencies:"))
				if deps != "none" && deps != "" {
					currentTask.Dependencies = strings.Split(deps, ",")
					for i := range currentTask.Dependencies {
						currentTask.Dependencies[i] = strings.TrimSpace(currentTask.Dependencies[i])
					}
				}
			} else if strings.HasPrefix(line, "Success Criteria:") {
				criteria := strings.TrimSpace(strings.TrimPrefix(line, "Success Criteria:"))
				currentTask.Context["success_criteria"] = criteria
			} else if strings.HasPrefix(line, "Risk:") {
				risk := strings.TrimSpace(strings.TrimPrefix(line, "Risk:"))
				currentTask.Context["risk"] = risk
			} else if strings.HasPrefix(line, "Estimated Duration:") {
				duration := strings.TrimSpace(strings.TrimPrefix(line, "Estimated Duration:"))
				currentTask.Context["estimated_duration"] = duration
			}
		}
	}

	// Add last task
	if currentTask != nil {
		tasks = append(tasks, *currentTask)
	}

	// Build prompt for each task
	for i := range tasks {
		if tasks[i].Description != "" {
			tasks[i].Prompt = fmt.Sprintf("%s: %s", tasks[i].Name, tasks[i].Description)
		} else {
			tasks[i].Prompt = tasks[i].Name
		}
	}

	return tasks
}
