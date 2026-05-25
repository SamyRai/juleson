package app

import (
	"context"
	"fmt"
	"time"

	"github.com/SamyRai/juleson/internal/orchestration/domain"
	"github.com/SamyRai/juleson/internal/orchestration/ports"
)

const defaultAIMaxIterations = 20

const (
	aiAnalyzeNodeName     = "ai.analyze"
	aiPlanNodeName        = "ai.plan"
	aiDecideNodeName      = "ai.decide"
	aiRouteNodeName       = "ai.route"
	aiExecuteTaskNodeName = "ai.executeTask"
	aiReviewNodeName      = "ai.review"
	aiAdaptPlanNodeName   = "ai.adaptPlan"
	aiCompleteNodeName    = "ai.complete"
)

type AIWorkflowRunOptions struct {
	MaxIterations  int
	ApprovalPolicy domain.ApprovalPolicy
}

type AIWorkflowRunnerDeps struct {
	ProjectAnalyzer ports.ProjectAnalyzer
	Planner         ports.Planner
	DecisionMaker   ports.DecisionMaker
	TaskExecutor    ports.TaskExecutor
	Reviewer        ports.Reviewer
	MemoryStore     ports.MemoryStore
	CheckpointStore ports.CheckpointStore
	ProgressSink    ports.ProgressSink
	Clock           ports.Clock
	MaxIterations   int
}

type AIWorkflowRunner struct {
	deps  AIWorkflowRunnerDeps
	clock ports.Clock
}

func NewAIWorkflowRunner(deps AIWorkflowRunnerDeps) *AIWorkflowRunner {
	return &AIWorkflowRunner{deps: deps, clock: clockOrDefault(deps.Clock)}
}

func (r *AIWorkflowRunner) Run(ctx context.Context, goal domain.Goal) (*domain.Result, error) {
	return r.RunWithOptions(ctx, goal, AIWorkflowRunOptions{})
}

func (r *AIWorkflowRunner) RunWithOptions(ctx context.Context, goal domain.Goal, opts AIWorkflowRunOptions) (*domain.Result, error) {
	if goal.Description == "" {
		return nil, fmt.Errorf("goal description cannot be empty")
	}
	if r.deps.Planner == nil {
		return nil, fmt.Errorf("planner is required")
	}
	if r.deps.DecisionMaker == nil {
		return nil, fmt.Errorf("decision maker is required")
	}
	if r.deps.TaskExecutor == nil {
		return nil, fmt.Errorf("task executor is required")
	}

	start := r.clock.Now()
	maxIterations := opts.MaxIterations
	if maxIterations == 0 {
		maxIterations = r.deps.MaxIterations
	}
	if maxIterations == 0 {
		maxIterations = defaultAIMaxIterations
	}
	state := &appRunState{
		goal:          goal,
		aiOptions:     opts,
		startedAt:     start,
		maxIterations: maxIterations,
		result: &domain.Result{
			Goal:  goal,
			State: domain.StateAnalyzing,
			Tasks: make([]domain.TaskResult, 0),
		},
		execution: domain.ExecutionContext{
			Goal:           goal,
			StartedAt:      start,
			ApprovalPolicy: approvalPolicyOrDefault(opts.ApprovalPolicy),
		},
	}

	graph, err := newAgentGraph(aiAnalyzeNodeName, map[string]graphNode{
		aiAnalyzeNodeName:     r.aiAnalyzeNode,
		aiPlanNodeName:        r.aiPlanNode,
		aiDecideNodeName:      r.aiDecideNode,
		aiRouteNodeName:       r.aiRouteDecisionNode,
		aiExecuteTaskNodeName: r.aiExecuteTaskNode,
		aiReviewNodeName:      r.aiReviewNode,
		aiAdaptPlanNodeName:   r.aiAdaptPlanNode,
		aiCompleteNodeName:    r.aiCompleteNode,
	})
	if err != nil {
		return r.finishWithCheckpoint(ctx, state.result, start, err, state.execution, "failed"), err
	}

	if err := graph.run(ctx, state); err != nil {
		return r.finishWithCheckpoint(ctx, state.result, start, err, state.execution, "failed"), err
	}
	return state.result, nil
}

func (r *AIWorkflowRunner) aiAnalyzeNode(ctx context.Context, state *appRunState) (string, error) {
	if r.deps.ProjectAnalyzer == nil || state.goal.Context.ProjectPath == "" {
		return aiPlanNodeName, nil
	}
	project, err := r.deps.ProjectAnalyzer.AnalyzeProject(ctx, state.goal.Context.ProjectPath)
	if err != nil {
		return "", err
	}
	state.project = project
	state.execution.Project = project
	return aiPlanNodeName, nil
}

func (r *AIWorkflowRunner) aiPlanNode(ctx context.Context, state *appRunState) (string, error) {
	state.result.State = domain.StatePlanning
	if err := reportProgress(ctx, r.deps.ProgressSink, domain.Progress{
		State:     domain.StatePlanning,
		Message:   "Planning AI workflow",
		Progress:  20,
		Timestamp: r.clock.Now(),
	}); err != nil {
		return "", err
	}
	plan, err := r.deps.Planner.Plan(ctx, state.goal, state.project)
	if err != nil {
		return "", err
	}
	state.plan = plan
	state.result.Plan = plan
	state.execution.Plan = plan
	if err := r.saveCheckpoint(ctx, state.result, state.execution, "planned"); err != nil {
		return "", err
	}
	return aiDecideNodeName, nil
}

func (r *AIWorkflowRunner) aiDecideNode(ctx context.Context, state *appRunState) (string, error) {
	if state.iteration >= state.maxIterations {
		return "", fmt.Errorf("max iterations (%d) reached", state.maxIterations)
	}
	state.iteration++
	state.execution.Iteration = state.iteration
	state.execution.Completed = append([]domain.TaskResult(nil), state.result.Tasks...)

	decision, err := r.deps.DecisionMaker.Decide(ctx, state.execution)
	if err != nil {
		return "", err
	}
	if decision == nil {
		return "", fmt.Errorf("decision maker returned nil decision")
	}
	state.decision = decision
	state.execution.Decisions = append(state.execution.Decisions, *decision)
	if r.deps.MemoryStore != nil {
		if err := r.deps.MemoryStore.RecordDecision(ctx, *decision); err != nil {
			return "", err
		}
	}
	if err := r.saveCheckpoint(ctx, state.result, state.execution, "decision"); err != nil {
		return "", err
	}
	return aiRouteNodeName, nil
}

func (r *AIWorkflowRunner) aiRouteDecisionNode(ctx context.Context, state *appRunState) (string, error) {
	switch state.decision.Type {
	case domain.DecisionTypeNextTask:
		task, err := selectTask(state.plan.Tasks, state.result.Tasks, state.decision.TaskID)
		if err != nil {
			return "", err
		}
		state.selectedTask = task
		return aiExecuteTaskNodeName, nil
	case domain.DecisionTypeReviewNeeded:
		return aiReviewNodeName, nil
	case domain.DecisionTypeAdaptPlan:
		return aiAdaptPlanNodeName, nil
	case domain.DecisionTypeComplete:
		return aiCompleteNodeName, nil
	case domain.DecisionTypeAbort:
		state.err = fmt.Errorf("AI workflow aborted: %s", state.decision.Reasoning)
		return graphFailNode, nil
	default:
		state.err = fmt.Errorf("unsupported decision type %q", state.decision.Type)
		return graphFailNode, nil
	}
}

func (r *AIWorkflowRunner) aiExecuteTaskNode(ctx context.Context, state *appRunState) (string, error) {
	state.result.State = domain.StateExecuting
	if err := reportProgress(ctx, r.deps.ProgressSink, domain.Progress{
		State:          domain.StateExecuting,
		CurrentTask:    state.selectedTask.Name,
		CompletedTasks: len(state.result.Tasks),
		TotalTasks:     len(state.plan.Tasks),
		Progress:       percent(len(state.result.Tasks), len(state.plan.Tasks)),
		Message:        "Executing AI workflow task",
		Timestamp:      r.clock.Now(),
	}); err != nil {
		return "", err
	}
	taskResult, err := r.deps.TaskExecutor.ExecuteTask(ctx, state.selectedTask, state.execution)
	if taskResult != nil {
		state.result.Tasks = append(state.result.Tasks, *taskResult)
		state.result.Artifacts = append(state.result.Artifacts, taskResult.Artifacts...)
	}
	state.execution.Completed = append([]domain.TaskResult(nil), state.result.Tasks...)
	if err != nil {
		return "", err
	}
	if err := r.saveCheckpoint(ctx, state.result, state.execution, "task"); err != nil {
		return "", err
	}
	return aiDecideNodeName, nil
}

func (r *AIWorkflowRunner) aiReviewNode(ctx context.Context, state *appRunState) (string, error) {
	state.result.State = domain.StateReviewing
	if err := reportProgress(ctx, r.deps.ProgressSink, domain.Progress{
		State:          domain.StateReviewing,
		CompletedTasks: len(state.result.Tasks),
		TotalTasks:     len(state.plan.Tasks),
		Progress:       percent(len(state.result.Tasks), len(state.plan.Tasks)),
		Message:        "Reviewing AI workflow",
		Timestamp:      r.clock.Now(),
	}); err != nil {
		return "", err
	}
	if r.deps.Reviewer == nil {
		if err := r.saveCheckpoint(ctx, state.result, state.execution, "review"); err != nil {
			return "", err
		}
		return aiDecideNodeName, nil
	}
	if _, err := r.deps.Reviewer.Review(ctx, state.execution); err != nil {
		return "", err
	}
	if err := r.saveCheckpoint(ctx, state.result, state.execution, "review"); err != nil {
		return "", err
	}
	return aiDecideNodeName, nil
}

func (r *AIWorkflowRunner) aiAdaptPlanNode(ctx context.Context, state *appRunState) (string, error) {
	state.result.State = domain.StatePlanning
	adapted, err := r.deps.Planner.AdaptPlan(ctx, state.execution, state.decision.Reasoning)
	if err != nil {
		return "", err
	}
	state.plan = adapted
	state.result.Plan = adapted
	state.execution.Plan = adapted
	if err := r.saveCheckpoint(ctx, state.result, state.execution, "planned"); err != nil {
		return "", err
	}
	return aiDecideNodeName, nil
}

func (r *AIWorkflowRunner) aiCompleteNode(ctx context.Context, state *appRunState) (string, error) {
	state.result.State = domain.StateComplete
	state.result.Success = true
	state.result.Duration = r.clock.Now().Sub(state.startedAt)
	if r.deps.MemoryStore != nil {
		if err := r.deps.MemoryStore.RecordResult(ctx, *state.result); err != nil {
			return "", err
		}
	}
	if err := reportProgress(ctx, r.deps.ProgressSink, domain.Progress{
		State:          domain.StateComplete,
		CompletedTasks: len(state.result.Tasks),
		TotalTasks:     len(state.result.Tasks),
		Progress:       100,
		Message:        "AI workflow complete",
		Timestamp:      r.clock.Now(),
	}); err != nil {
		return "", err
	}
	if err := r.saveCheckpoint(ctx, state.result, state.execution, "complete"); err != nil {
		return "", err
	}
	return graphEndNode, nil
}

func approvalPolicyOrDefault(policy domain.ApprovalPolicy) domain.ApprovalPolicy {
	if policy.AutoApprove {
		policy.RequirePlanApproval = false
		return policy
	}
	if !policy.RequirePlanApproval {
		policy.RequirePlanApproval = true
	}
	return policy
}

func (r *AIWorkflowRunner) finish(result *domain.Result, start time.Time, err error) *domain.Result {
	result.State = domain.StateFailed
	result.Success = false
	result.Error = err
	result.Duration = r.clock.Now().Sub(start)
	return result
}

func (r *AIWorkflowRunner) finishWithCheckpoint(ctx context.Context, result *domain.Result, start time.Time, err error, execution domain.ExecutionContext, phase string) *domain.Result {
	result = r.finish(result, start, err)
	_ = r.saveCheckpoint(ctx, result, execution, phase)
	return result
}

func (r *AIWorkflowRunner) saveCheckpoint(ctx context.Context, result *domain.Result, execution domain.ExecutionContext, phase string) error {
	if r.deps.CheckpointStore == nil {
		return nil
	}
	if execution.Completed == nil {
		execution.Completed = append([]domain.TaskResult(nil), result.Tasks...)
	}
	if execution.Plan == nil {
		execution.Plan = result.Plan
	}
	checkpoint := domain.Checkpoint{
		ID:        checkpointID(result.Goal.ID, phase, execution.Iteration),
		GoalID:    result.Goal.ID,
		State:     result.State,
		Context:   execution,
		CreatedAt: r.clock.Now(),
		Metadata: map[string]string{
			"phase": phase,
		},
	}
	return r.deps.CheckpointStore.SaveCheckpoint(ctx, checkpoint)
}

func selectTask(tasks []domain.Task, completed []domain.TaskResult, taskID string) (domain.Task, error) {
	completedByID := map[string]bool{}
	for _, task := range completed {
		if task.TaskID != "" {
			completedByID[task.TaskID] = true
		}
		if task.TaskName != "" {
			completedByID[task.TaskName] = true
		}
	}
	if taskID != "" {
		for _, task := range tasks {
			if task.ID == taskID || task.Name == taskID {
				return task, nil
			}
		}
		return domain.Task{}, fmt.Errorf("task %q not found", taskID)
	}
	for _, task := range tasks {
		key := task.ID
		if key == "" {
			key = task.Name
		}
		if !completedByID[key] {
			return task, nil
		}
	}
	return domain.Task{}, fmt.Errorf("no pending tasks")
}
