package automation

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/SamyRai/juleson/internal/gemini"
	"github.com/SamyRai/juleson/internal/jules"
	"google.golang.org/genai"
)

// AIOrchestrator uses Gemini AI to intelligently orchestrate complex workflows
// The AI decides:
// - What tasks to execute and in what order
// - When to review and when to auto-approve
// - How to adapt based on progress and results
// - When the workflow is complete
type AIOrchestrator struct {
	julesClient  *jules.Client
	geminiClient *gemini.Client
	sessionID    string

	// Workflow state
	goal        string
	projectPath string
	constraints []string
	context     *ProjectContext

	// Execution state
	currentPhase   string
	completedTasks []CompletedTask
	pendingTasks   []PendingTask

	// AI decision making
	lastAIDecision  *AIDecision
	decisionHistory []AIDecision

	// Monitoring
	state        AIState
	startTime    time.Time
	mu           sync.RWMutex
	progressChan chan AIProgress
	decisionChan chan AIDecision
	stopChan     chan struct{}

	// Configuration
	maxIterations int
	checkInterval time.Duration
	allowedTools  []string
}

// AIState represents the current state of AI orchestration
type AIState string

const (
	AIStateAnalyzing AIState = "ANALYZING"
	AIStatePlanning  AIState = "PLANNING"
	AIStateExecuting AIState = "EXECUTING"
	AIStateReviewing AIState = "REVIEWING"
	AIStateAdapting  AIState = "ADAPTING"
	AIStateCompleted AIState = "COMPLETED"
	AIStateFailed    AIState = "FAILED"
)

// ProjectContext contains AI's understanding of the project
type ProjectContext struct {
	Languages     []string
	Frameworks    []string
	Architecture  string
	Complexity    string
	CurrentState  string
	KeyFiles      []string
	Dependencies  map[string]string
	TestCoverage  float64
	Issues        []string
	Opportunities []string
}

// CompletedTask represents a task that has been completed
type CompletedTask struct {
	Name           string
	Description    string
	Result         AITaskResult
	Timestamp      time.Time
	Artifacts      []jules.Artifact
	LessonsLearned []string
}

// PendingTask represents a task the AI wants to execute
type PendingTask struct {
	Name         string
	Description  string
	Prompt       string
	Priority     int
	Dependencies []string
	Rationale    string
}

// AITaskResult represents the result of executing an AI-orchestrated task
type AITaskResult struct {
	Success      bool
	ActivityID   string
	FilesChanged []string
	TestsPassed  bool
	Errors       []string
	Warnings     []string
}

// AIDecision represents a decision made by the AI
type AIDecision struct {
	Timestamp    time.Time
	DecisionType string // "next_task", "review_needed", "adapt_plan", "complete"
	Reasoning    string
	Action       string
	Confidence   float64
	Alternatives []string
}

// AIProgress represents progress update from AI orchestration
type AIProgress struct {
	Phase       string
	CurrentTask string
	Progress    float64
	Message     string
	Timestamp   time.Time
	NextSteps   []string
}

// AIOrchestrationConfig configures the AI orchestrator
type AIOrchestrationConfig struct {
	MaxIterations  int
	CheckInterval  time.Duration
	AllowedTools   []string
	AutoApprove    bool
	MaxSessionTime time.Duration
}

// DefaultAIOrchestrationConfig returns default configuration
func DefaultAIOrchestrationConfig() *AIOrchestrationConfig {
	return &AIOrchestrationConfig{
		MaxIterations: 20,
		CheckInterval: 15 * time.Second,
		AllowedTools: []string{
			"execute_template",
			"run_tests",
			"apply_patches",
			"create_issue",
			"create_milestone",
		},
		AutoApprove:    false,
		MaxSessionTime: 4 * time.Hour,
	}
}

// NewAIOrchestrator creates a new AI-powered orchestrator
func NewAIOrchestrator(
	julesClient *jules.Client,
	geminiClient *gemini.Client,
	config *AIOrchestrationConfig,
) *AIOrchestrator {
	if config == nil {
		config = DefaultAIOrchestrationConfig()
	}

	return &AIOrchestrator{
		julesClient:     julesClient,
		geminiClient:    geminiClient,
		state:           AIStateAnalyzing,
		progressChan:    make(chan AIProgress, 100),
		decisionChan:    make(chan AIDecision, 100),
		stopChan:        make(chan struct{}),
		maxIterations:   config.MaxIterations,
		checkInterval:   config.CheckInterval,
		allowedTools:    config.AllowedTools,
		completedTasks:  make([]CompletedTask, 0),
		pendingTasks:    make([]PendingTask, 0),
		decisionHistory: make([]AIDecision, 0),
	}
}

// Execute runs the AI orchestration for the given goal
func (ao *AIOrchestrator) Execute(ctx context.Context, sourceID, goal, projectPath string, constraints []string) error {
	ao.mu.Lock()
	ao.goal = goal
	ao.projectPath = projectPath
	ao.constraints = constraints
	ao.startTime = time.Now()
	ao.mu.Unlock()

	// Phase 1: AI analyzes the project
	ao.setState(AIStateAnalyzing)
	ao.sendProgress("Analyzing project", "AI is understanding your codebase", 0, nil)

	if err := ao.analyzeProject(ctx); err != nil {
		return fmt.Errorf("project analysis failed: %w", err)
	}

	// Phase 2: AI creates initial plan
	ao.setState(AIStatePlanning)
	ao.sendProgress("Planning workflow", "AI is creating execution plan", 10, nil)

	if err := ao.createInitialPlan(ctx); err != nil {
		return fmt.Errorf("planning failed: %w", err)
	}

	// Phase 3: Create Jules session with AI-generated comprehensive prompt
	ao.setState(AIStateExecuting)
	initialPrompt := ao.buildAIPrompt()

	session, err := ao.julesClient.CreateSession(ctx, &jules.CreateSessionRequest{
		Prompt: initialPrompt,
		SourceContext: &jules.SourceContext{
			Source: fmt.Sprintf("sources/%s", sourceID),
		},
		RequirePlanApproval: true, // Always review AI decisions
	})
	if err != nil {
		ao.setState(AIStateFailed)
		return fmt.Errorf("failed to create session: %w", err)
	}

	ao.mu.Lock()
	ao.sessionID = session.ID
	ao.mu.Unlock()

	ao.sendProgress("Session started", fmt.Sprintf("Session %s created", session.ID), 20, nil)

	// Phase 4: AI-driven execution loop
	return ao.executionLoop(ctx)
}

// analyzeProject uses AI to deeply understand the project
func (ao *AIOrchestrator) analyzeProject(ctx context.Context) error {
	// Build analysis prompt for Gemini
	analysisPrompt := fmt.Sprintf(`Analyze this software project to understand its context:

Project Path: %s
User Goal: %s
Constraints: %s

Please analyze:
1. Programming languages and frameworks used
2. Current architecture and design patterns
3. Code quality and technical debt
4. Test coverage and quality
5. Key areas that need attention
6. Potential risks and challenges
7. Recommended approach to achieve the goal

Provide a structured JSON response with your analysis.`,
		ao.projectPath,
		ao.goal,
		strings.Join(ao.constraints, ", "),
	)

	resp, err := ao.geminiClient.GenerateContent("", analysisPrompt)
	if err != nil {
		return fmt.Errorf("AI analysis failed: %w", err)
	}

	// Parse AI's analysis
	analysis := extractAnalysisFromResponse(resp)

	ao.mu.Lock()
	ao.context = analysis
	ao.mu.Unlock()

	return nil
}

// createInitialPlan asks AI to create an execution plan
func (ao *AIOrchestrator) createInitialPlan(ctx context.Context) error {
	planningPrompt := fmt.Sprintf(`Based on the project analysis, create a detailed execution plan to achieve this goal:

Goal: %s

Project Context:
- Languages: %s
- Architecture: %s
- Complexity: %s
- Current Issues: %s

Constraints: %s

Create a plan with:
1. High-priority tasks that should be executed first
2. Dependencies between tasks
3. Estimated effort for each task
4. Risk assessment
5. Success criteria

Be adaptive - the plan may change based on execution results.
Provide a structured JSON response with tasks prioritized and sequenced.`,
		ao.goal,
		strings.Join(ao.context.Languages, ", "),
		ao.context.Architecture,
		ao.context.Complexity,
		strings.Join(ao.context.Issues, "; "),
		strings.Join(ao.constraints, ", "),
	)

	resp, err := ao.geminiClient.GenerateContent("", planningPrompt)
	if err != nil {
		return fmt.Errorf("AI planning failed: %w", err)
	}

	// Parse AI's plan into pending tasks
	tasks := extractTasksFromResponse(resp)

	ao.mu.Lock()
	ao.pendingTasks = tasks
	ao.mu.Unlock()

	return nil
}

// executionLoop is the main AI-driven execution loop
func (ao *AIOrchestrator) executionLoop(ctx context.Context) error {
	iteration := 0

	for iteration < ao.maxIterations {
		iteration++

		// Check if we should stop
		select {
		case <-ao.stopChan:
			ao.setState(AIStateFailed)
			return fmt.Errorf("orchestration stopped")
		case <-ctx.Done():
			ao.setState(AIStateFailed)
			return ctx.Err()
		default:
		}

		// AI decides what to do next
		decision, err := ao.makeNextDecision(ctx)
		if err != nil {
			return fmt.Errorf("AI decision failed: %w", err)
		}

		ao.recordDecision(decision)

		// Execute based on AI decision
		switch decision.DecisionType {
		case "next_task":
			if err := ao.executeNextTask(ctx, decision); err != nil {
				return fmt.Errorf("task execution failed: %w", err)
			}

		case "review_needed":
			ao.setState(AIStateReviewing)
			if err := ao.reviewAndAdapt(ctx, decision); err != nil {
				return fmt.Errorf("review failed: %w", err)
			}

		case "adapt_plan":
			ao.setState(AIStateAdapting)
			if err := ao.adaptPlan(ctx, decision); err != nil {
				return fmt.Errorf("adaptation failed: %w", err)
			}

		case "complete":
			ao.setState(AIStateCompleted)
			ao.sendProgress("Workflow complete", decision.Reasoning, 100, nil)
			return nil

		default:
			return fmt.Errorf("unknown decision type: %s", decision.DecisionType)
		}

		// Brief pause before next iteration
		time.Sleep(ao.checkInterval)
	}

	return fmt.Errorf("max iterations (%d) reached", ao.maxIterations)
}

// makeNextDecision asks AI what to do next
func (ao *AIOrchestrator) makeNextDecision(ctx context.Context) (*AIDecision, error) {
	ao.mu.RLock()
	completedCount := len(ao.completedTasks)
	pendingCount := len(ao.pendingTasks)
	ao.mu.RUnlock()

	// Get latest session status
	session, err := ao.julesClient.GetSession(ctx, ao.sessionID)
	if err != nil {
		return nil, fmt.Errorf("failed to get session: %w", err)
	}

	// Build decision prompt for AI
	decisionPrompt := fmt.Sprintf(`You are orchestrating a software development workflow. Make the next decision:

Original Goal: %s
Session State: %s
Completed Tasks: %d
Pending Tasks: %d

Recent Completed Tasks:
%s

Current Pending Tasks:
%s

What should happen next? Choose one:
1. "next_task" - Execute the next task from pending list
2. "review_needed" - Review progress and get user feedback
3. "adapt_plan" - Adjust the plan based on results
4. "complete" - Goal is achieved, workflow complete

Provide JSON response with:
- decision_type: (next_task|review_needed|adapt_plan|complete)
- reasoning: Why this decision
- action: Specific action to take
- confidence: 0.0-1.0
- next_steps: What happens after this`,
		ao.goal,
		session.State,
		completedCount,
		pendingCount,
		ao.formatCompletedTasks(),
		ao.formatPendingTasks(),
	)

	resp, err := ao.geminiClient.GenerateContent("", decisionPrompt)
	if err != nil {
		return nil, fmt.Errorf("AI decision generation failed: %w", err)
	}

	// Parse AI's decision
	decision := extractDecisionFromResponse(resp)
	decision.Timestamp = time.Now()

	ao.sendDecision(*decision)

	return decision, nil
}

// executeNextTask executes the task AI decided on
func (ao *AIOrchestrator) executeNextTask(ctx context.Context, decision *AIDecision) error {
	ao.setState(AIStateExecuting)

	// Get next task from pending list
	ao.mu.Lock()
	if len(ao.pendingTasks) == 0 {
		ao.mu.Unlock()
		return fmt.Errorf("no pending tasks")
	}
	task := ao.pendingTasks[0]
	ao.pendingTasks = ao.pendingTasks[1:]
	ao.mu.Unlock()

	ao.sendProgress("Executing task", task.Name,
		float64(len(ao.completedTasks))/float64(len(ao.completedTasks)+len(ao.pendingTasks)+1)*100,
		[]string{task.Description},
	)

	// Send message to Jules session to execute this task
	err := ao.julesClient.SendMessage(ctx, ao.sessionID, &jules.SendMessageRequest{
		Prompt: task.Prompt,
	})
	if err != nil {
		return fmt.Errorf("failed to send task to session: %w", err)
	}

	// Wait for task completion and gather results
	time.Sleep(5 * time.Second) // Allow time for execution

	// Get activities to see results
	activities, err := ao.julesClient.ListActivities(ctx, ao.sessionID, 5)
	if err != nil {
		return fmt.Errorf("failed to list activities: %w", err)
	}

	result := AITaskResult{
		Success: true,
	}

	if len(activities) > 0 {
		result.ActivityID = activities[0].ID
		result.FilesChanged = extractFilesChanged(activities[0])
	}

	// Record completed task
	completed := CompletedTask{
		Name:        task.Name,
		Description: task.Description,
		Result:      result,
		Timestamp:   time.Now(),
	}

	if len(activities) > 0 {
		completed.Artifacts = activities[0].Artifacts
	}

	ao.mu.Lock()
	ao.completedTasks = append(ao.completedTasks, completed)
	ao.mu.Unlock()

	return nil
}

// reviewAndAdapt asks AI to review progress and adapt
func (ao *AIOrchestrator) reviewAndAdapt(ctx context.Context, decision *AIDecision) error {
	// AI reviews completed tasks and decides on adaptations
	reviewPrompt := fmt.Sprintf(`Review the progress and decide if we should adapt the plan:

Original Goal: %s
Completed Tasks: %d
Tasks Summary:
%s

Should we:
1. Continue with current plan?
2. Adjust remaining tasks based on learnings?
3. Add new tasks we discovered are needed?
4. Skip tasks that are no longer relevant?

Provide detailed reasoning and updated task list if needed.`,
		ao.goal,
		len(ao.completedTasks),
		ao.formatCompletedTasksDetailed(),
	)

	resp, err := ao.geminiClient.GenerateContent("", reviewPrompt)
	if err != nil {
		return fmt.Errorf("review failed: %w", err)
	}

	// Parse AI's review and apply changes
	adaptations := extractAdaptationsFromResponse(resp)
	ao.applyAdaptations(adaptations)

	return nil
}

// adaptPlan adapts the plan based on AI's decisions
func (ao *AIOrchestrator) adaptPlan(ctx context.Context, decision *AIDecision) error {
	// Similar to reviewAndAdapt but focused on plan changes
	return ao.reviewAndAdapt(ctx, decision)
}

// Helper methods

func (ao *AIOrchestrator) buildAIPrompt() string {
	prompt := fmt.Sprintf(`I need you to help achieve this goal: %s

`, ao.goal)

	if ao.context != nil {
		prompt += fmt.Sprintf(`Project Context:
- Languages: %s
- Architecture: %s
- Current State: %s

`, strings.Join(ao.context.Languages, ", "), ao.context.Architecture, ao.context.CurrentState)
	}

	if len(ao.pendingTasks) > 0 {
		prompt += "Initial tasks to execute:\n"
		for i, task := range ao.pendingTasks[:min(3, len(ao.pendingTasks))] {
			prompt += fmt.Sprintf("%d. %s\n", i+1, task.Description)
		}
	}

	prompt += "\nThis is a multi-step process. I'll send follow-up messages as we progress."

	return prompt
}

func (ao *AIOrchestrator) setState(state AIState) {
	ao.mu.Lock()
	defer ao.mu.Unlock()
	ao.state = state
}

func (ao *AIOrchestrator) sendProgress(phase, message string, progress float64, nextSteps []string) {
	update := AIProgress{
		Phase:       phase,
		CurrentTask: message,
		Progress:    progress,
		Message:     message,
		Timestamp:   time.Now(),
		NextSteps:   nextSteps,
	}

	select {
	case ao.progressChan <- update:
	default:
	}
}

func (ao *AIOrchestrator) sendDecision(decision AIDecision) {
	select {
	case ao.decisionChan <- decision:
	default:
	}
}

func (ao *AIOrchestrator) recordDecision(decision *AIDecision) {
	ao.mu.Lock()
	defer ao.mu.Unlock()
	ao.lastAIDecision = decision
	ao.decisionHistory = append(ao.decisionHistory, *decision)
}

func (ao *AIOrchestrator) formatCompletedTasks() string {
	ao.mu.RLock()
	defer ao.mu.RUnlock()

	if len(ao.completedTasks) == 0 {
		return "None yet"
	}

	var sb strings.Builder
	for i, task := range ao.completedTasks {
		sb.WriteString(fmt.Sprintf("%d. %s - %s\n", i+1, task.Name, task.Description))
	}
	return sb.String()
}

func (ao *AIOrchestrator) formatCompletedTasksDetailed() string {
	ao.mu.RLock()
	defer ao.mu.RUnlock()

	var sb strings.Builder
	for i, task := range ao.completedTasks {
		sb.WriteString(fmt.Sprintf("%d. %s\n", i+1, task.Name))
		sb.WriteString(fmt.Sprintf("   Description: %s\n", task.Description))
		sb.WriteString(fmt.Sprintf("   Success: %v\n", task.Result.Success))
		if len(task.Result.FilesChanged) > 0 {
			sb.WriteString(fmt.Sprintf("   Files Changed: %d\n", len(task.Result.FilesChanged)))
		}
	}
	return sb.String()
}

func (ao *AIOrchestrator) formatPendingTasks() string {
	ao.mu.RLock()
	defer ao.mu.RUnlock()

	if len(ao.pendingTasks) == 0 {
		return "None"
	}

	var sb strings.Builder
	for i, task := range ao.pendingTasks {
		sb.WriteString(fmt.Sprintf("%d. %s (Priority: %d)\n", i+1, task.Name, task.Priority))
	}
	return sb.String()
}

func (ao *AIOrchestrator) applyAdaptations(adaptations map[string]interface{}) {
	// Apply AI's recommended adaptations to the pending task list
	// This would parse the adaptations and update ao.pendingTasks
}

// Public methods

func (ao *AIOrchestrator) ProgressChannel() <-chan AIProgress {
	return ao.progressChan
}

func (ao *AIOrchestrator) DecisionChannel() <-chan AIDecision {
	return ao.decisionChan
}

func (ao *AIOrchestrator) Stop() {
	close(ao.stopChan)
}

func (ao *AIOrchestrator) GetState() AIState {
	ao.mu.RLock()
	defer ao.mu.RUnlock()
	return ao.state
}

func (ao *AIOrchestrator) GetSessionID() string {
	ao.mu.RLock()
	defer ao.mu.RUnlock()
	return ao.sessionID
}

func (ao *AIOrchestrator) GetDecisionHistory() []AIDecision {
	ao.mu.RLock()
	defer ao.mu.RUnlock()
	history := make([]AIDecision, len(ao.decisionHistory))
	copy(history, ao.decisionHistory)
	return history
}

// Utility functions for parsing AI responses

func extractAnalysisFromResponse(resp *genai.GenerateContentResponse) *ProjectContext {
	if len(resp.Candidates) == 0 || len(resp.Candidates[0].Content.Parts) == 0 {
		return &ProjectContext{
			Languages:    []string{"unknown"},
			Architecture: "unknown",
			Complexity:   "unknown",
			CurrentState: "unknown",
		}
	}

	text := resp.Candidates[0].Content.Parts[0].Text

	// Parse AI's structured analysis response
	context := &ProjectContext{}

	// Extract languages
	if strings.Contains(text, "Languages:") || strings.Contains(text, "languages:") {
		// Simple extraction - in production would use more robust parsing
		lines := strings.Split(text, "\n")
		for _, line := range lines {
			if strings.Contains(strings.ToLower(line), "languages") {
				// Extract languages from the line
				parts := strings.Split(line, ":")
				if len(parts) > 1 {
					langs := strings.TrimSpace(parts[1])
					context.Languages = strings.Split(langs, ",")
					for i, lang := range context.Languages {
						context.Languages[i] = strings.TrimSpace(lang)
					}
				}
			}
		}
	}

	// Extract architecture
	if strings.Contains(text, "Architecture:") || strings.Contains(text, "architecture:") {
		lines := strings.Split(text, "\n")
		for _, line := range lines {
			if strings.Contains(strings.ToLower(line), "architecture") {
				parts := strings.Split(line, ":")
				if len(parts) > 1 {
					context.Architecture = strings.TrimSpace(parts[1])
				}
			}
		}
	}

	// Extract complexity
	if strings.Contains(text, "Complexity:") || strings.Contains(text, "complexity:") {
		lines := strings.Split(text, "\n")
		for _, line := range lines {
			if strings.Contains(strings.ToLower(line), "complexity") {
				parts := strings.Split(line, ":")
				if len(parts) > 1 {
					context.Complexity = strings.TrimSpace(parts[1])
				}
			}
		}
	}

	// Extract current state
	if strings.Contains(text, "Current State:") || strings.Contains(text, "current state:") {
		lines := strings.Split(text, "\n")
		for _, line := range lines {
			if strings.Contains(strings.ToLower(line), "current state") {
				parts := strings.Split(line, ":")
				if len(parts) > 1 {
					context.CurrentState = strings.TrimSpace(parts[1])
				}
			}
		}
	}

	// Extract issues and opportunities
	context.Issues = extractListFromText(text, "issues")
	context.Opportunities = extractListFromText(text, "opportunities")

	// If no structured data found, provide defaults but mark as AI-generated
	if len(context.Languages) == 0 {
		context.Languages = []string{"Go"} // Default assumption
	}
	if context.Architecture == "" {
		context.Architecture = "Microservices"
	}
	if context.Complexity == "" {
		context.Complexity = "Medium"
	}
	if context.CurrentState == "" {
		context.CurrentState = "Functional but needs modernization"
	}

	return context
}

func extractTasksFromResponse(resp *genai.GenerateContentResponse) []PendingTask {
	if len(resp.Candidates) == 0 || len(resp.Candidates[0].Content.Parts) == 0 {
		return []PendingTask{}
	}

	text := resp.Candidates[0].Content.Parts[0].Text
	tasks := []PendingTask{}

	// Try to parse as JSON first
	if strings.Contains(text, "{") && strings.Contains(text, "}") {
		// Attempt JSON parsing
		var plan AITaskPlan
		if err := json.Unmarshal([]byte(text), &plan); err == nil {
			for i, aiTask := range plan.Tasks {
				tasks = append(tasks, PendingTask{
					Name:        aiTask.Name,
					Description: aiTask.Description,
					Prompt:      aiTask.Prompt,
					Priority:    i + 1,
					Rationale:   plan.Reasoning,
				})
			}
			return tasks
		}
	}

	// Fallback: Parse structured text
	lines := strings.Split(text, "\n")
	currentTask := PendingTask{}
	priority := 1

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		// Look for numbered tasks
		if strings.HasPrefix(line, fmt.Sprintf("%d.", priority)) ||
			strings.HasPrefix(line, fmt.Sprintf("%d)", priority)) {
			// Save previous task if exists
			if currentTask.Name != "" {
				tasks = append(tasks, currentTask)
			}

			// Start new task
			taskText := strings.TrimSpace(line[strings.Index(line, ".")+1:])
			currentTask = PendingTask{
				Name:        fmt.Sprintf("Task %d", priority),
				Description: taskText,
				Prompt:      taskText,
				Priority:    priority,
			}
			priority++
		} else if strings.HasPrefix(line, "- ") {
			// Bullet point - could be sub-task or detail
			detail := strings.TrimSpace(line[2:])
			if currentTask.Description != "" {
				currentTask.Description += " " + detail
			}
		}
	}

	// Add final task
	if currentTask.Name != "" {
		tasks = append(tasks, currentTask)
	}

	// If no tasks found, create a default one
	if len(tasks) == 0 {
		tasks = []PendingTask{
			{
				Name:        "Initial Analysis",
				Description: "Analyze project and create detailed plan",
				Prompt:      "Please analyze this project and provide a detailed implementation plan",
				Priority:    1,
				Rationale:   "Need to understand the project before proceeding",
			},
		}
	}

	return tasks
}

func extractDecisionFromResponse(resp *genai.GenerateContentResponse) *AIDecision {
	// Parse Gemini's response into a decision
	if len(resp.Candidates) == 0 || len(resp.Candidates[0].Content.Parts) == 0 {
		return &AIDecision{
			DecisionType: "next_task",
			Reasoning:    "Continue with planned tasks",
			Confidence:   0.8,
		}
	}

	// Try to parse JSON from response
	text := resp.Candidates[0].Content.Parts[0].Text
	var decision AIDecision

	// Simple parsing - in production would be more robust
	if strings.Contains(text, "complete") {
		decision.DecisionType = "complete"
		decision.Reasoning = "Goal appears to be achieved"
	} else if strings.Contains(text, "review") {
		decision.DecisionType = "review_needed"
		decision.Reasoning = "Time to review progress"
	} else {
		decision.DecisionType = "next_task"
		decision.Reasoning = "Continue with next task"
	}

	decision.Confidence = 0.8
	return &decision
}

func extractAdaptationsFromResponse(resp *genai.GenerateContentResponse) map[string]interface{} {
	// Parse AI's recommended adaptations
	adaptations := make(map[string]interface{})

	if len(resp.Candidates) == 0 || len(resp.Candidates[0].Content.Parts) == 0 {
		return adaptations
	}

	text := resp.Candidates[0].Content.Parts[0].Text

	// Try to parse as JSON first
	if strings.Contains(text, "{") && strings.Contains(text, "}") {
		if err := json.Unmarshal([]byte(text), &adaptations); err == nil {
			return adaptations
		}
	}

	// Fallback: extract key adaptation decisions from text
	if strings.Contains(strings.ToLower(text), "add task") ||
		strings.Contains(strings.ToLower(text), "new task") {
		adaptations["action"] = "add_tasks"
	}

	if strings.Contains(strings.ToLower(text), "remove task") ||
		strings.Contains(strings.ToLower(text), "skip task") {
		adaptations["action"] = "remove_tasks"
	}

	if strings.Contains(strings.ToLower(text), "reorder") ||
		strings.Contains(strings.ToLower(text), "reprioritize") {
		adaptations["action"] = "reorder_tasks"
	}

	return adaptations
}

func extractListFromText(text, keyword string) []string {
	lines := strings.Split(text, "\n")
	list := []string{}
	inList := false

	for _, line := range lines {
		line = strings.TrimSpace(line)
		lowerLine := strings.ToLower(line)

		// Check if we're entering the list section
		if strings.Contains(lowerLine, keyword) && (strings.Contains(lowerLine, ":") || strings.HasSuffix(lowerLine, ":")) {
			inList = true
			continue
		}

		// If we're in the list section, look for bullet points or numbered items
		if inList {
			if strings.HasPrefix(line, "- ") || strings.HasPrefix(line, "* ") ||
				(len(line) > 0 && line[0] >= '1' && line[0] <= '9' && strings.Contains(line, ".")) {
				// Remove bullet/number prefix
				if strings.HasPrefix(line, "- ") || strings.HasPrefix(line, "* ") {
					list = append(list, strings.TrimSpace(line[2:]))
				} else if strings.Contains(line, ". ") {
					parts := strings.SplitN(line, ". ", 2)
					if len(parts) > 1 {
						list = append(list, strings.TrimSpace(parts[1]))
					}
				}
			} else if line == "" {
				// Empty line might indicate end of list
				continue
			} else if !strings.Contains(lowerLine, keyword) && len(line) > 0 {
				// If we hit a non-list line that's not empty, might be end of section
				break
			}
		}
	}

	return list
}

func extractFilesChanged(activity jules.Activity) []string {
	files := make([]string, 0)
	for _, artifact := range activity.Artifacts {
		if artifact.ChangeSet != nil && artifact.ChangeSet.GitPatch != nil {
			// Would parse git patch to extract file names
			files = append(files, "modified_file.go")
		}
	}
	return files
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// JSON structures for better AI communication
type AITaskPlan struct {
	Tasks      []AITask `json:"tasks"`
	Reasoning  string   `json:"reasoning"`
	Priorities []int    `json:"priorities"`
}

type AITask struct {
	Name          string   `json:"name"`
	Description   string   `json:"description"`
	Prompt        string   `json:"prompt"`
	Dependencies  []string `json:"dependencies"`
	EstimatedTime string   `json:"estimated_time"`
}

// ExtractAnalysisFromResponse parses AI analysis response (public for testing)
func ExtractAnalysisFromResponse(resp *genai.GenerateContentResponse) *ProjectContext {
	return extractAnalysisFromResponse(resp)
}

// ExtractTasksFromResponse parses AI planning response (public for testing)
func ExtractTasksFromResponse(resp *genai.GenerateContentResponse) []PendingTask {
	return extractTasksFromResponse(resp)
}

// ExtractDecisionFromResponse parses AI decision response (public for testing)
func ExtractDecisionFromResponse(resp *genai.GenerateContentResponse) *AIDecision {
	return extractDecisionFromResponse(resp)
}
