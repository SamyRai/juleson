package core

import (
	"context"
	"fmt"
	"time"

	"github.com/SamyRai/juleson/internal/agent"
)

// review performs code review on changes.
func (a *CoreAgent) review(ctx context.Context, result *agent.Result) error {
	a.logger.Info("agent.review.start")

	var allChanges []agent.Change
	for _, task := range a.currentPlan {
		if task.Result != nil {
			allChanges = append(allChanges, task.Result.Changes...)
		}
	}

	if len(allChanges) == 0 {
		a.logger.Info("agent.review.no_changes")
		a.setState(agent.StateReflecting)
		return nil
	}

	reviewResult, err := a.reviewer.Review(ctx, allChanges)
	if err != nil {
		a.logger.Error("agent.review.failed", "error", err)
		return err
	}

	a.logger.Info("agent.review.complete",
		"decision", reviewResult.Decision,
		"score", reviewResult.Score,
		"comments", len(reviewResult.Comments))

	for i := range a.currentPlan {
		if a.currentPlan[i].Result != nil {
			a.currentPlan[i].Result.ReviewResult = reviewResult
		}
	}

	decision := agent.Decision{
		Timestamp:  time.Now(),
		State:      agent.StateReviewing,
		Type:       agent.DecisionTypeApprove,
		Reasoning:  reviewResult.Summary,
		Confidence: reviewResult.Score / PercentageScale,
	}

	if reviewResult.Decision == agent.ReviewDecisionReject {
		decision.Type = agent.DecisionTypeReject
		decision.Action = "Reject changes and retry"
		a.recordDecision(decision)

		for i := range a.currentPlan {
			if a.currentPlan[i].State != agent.TaskStateFailed {
				a.currentPlan[i].State = agent.TaskStatePending
			}
		}

		a.setState(agent.StatePlanning)
		return nil
	}

	if reviewResult.Decision == agent.ReviewDecisionRequestChanges {
		decision.Type = agent.DecisionTypeRequestChange
		decision.Action = "Request improvements"
		a.recordDecision(decision)
	}

	a.recordDecision(decision)
	a.setState(agent.StateReflecting)
	return nil
}

// reflect learns from the execution and decides next steps.
func (a *CoreAgent) reflect(ctx context.Context, result *agent.Result) error {
	a.logger.Info("agent.reflect.start")

	for _, task := range a.currentPlan {
		if task.Result != nil {
			result.Tasks = append(result.Tasks, *task.Result)
		}
	}

	learning := agent.Learning{
		Timestamp:  time.Now(),
		Context:    a.currentGoal.Description,
		Pattern:    "Jules tool execution",
		Lesson:     fmt.Sprintf("Completed %d tasks", len(result.Tasks)),
		Confidence: 0.7,
	}

	successCount := 0
	for _, taskResult := range result.Tasks {
		if taskResult.Success {
			successCount++
		}
	}

	if successCount == len(result.Tasks) {
		learning.Confidence = 0.9
		learning.Lesson += " - all tasks succeeded"
	}

	if err := a.memory.Store(ctx, learning); err != nil {
		a.logger.Warn("agent.reflect.store_learning_failed", "error", err)
	}

	result.Learnings = append(result.Learnings, learning)

	a.logger.Info("agent.reflect.complete", "learnings", len(result.Learnings))
	a.setState(agent.StateReflecting)
	return nil
}
