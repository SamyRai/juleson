package automation

import (
	"context"
	"fmt"
	"time"

	"github.com/SamyRai/go-jules"
)

// Start initiates the workflow orchestration.
func (o *SessionOrchestrator) Start(ctx context.Context, sourceID string) error {
	o.mu.Lock()
	o.state = StateRunning
	o.startTime = time.Now()
	o.mu.Unlock()

	initialPrompt := o.buildInitialPrompt()

	session, err := o.client.Sessions().Create(ctx, &jules.CreateSessionRequest{
		Prompt: initialPrompt,
		SourceContext: &jules.SourceContext{
			Source: fmt.Sprintf("sources/%s", sourceID),
		},
		RequirePlanApproval: !o.autoApprove,
	})
	if err != nil {
		o.setState(StateFailed)
		return fmt.Errorf("failed to create session: %w", err)
	}

	o.mu.Lock()
	o.sessionID = session.ID
	o.mu.Unlock()

	o.monitor = jules.NewSessionMonitor(o.client, session.ID).
		WithInterval(o.checkInterval).
		WithMaxWait(o.maxSessionAge).
		OnProgress(o.handleProgress)

	o.sendProgress(0, 0, "Session created", 0)

	return o.executeWorkflow(ctx)
}

// executeWorkflow executes all workflow phases.
func (o *SessionOrchestrator) executeWorkflow(ctx context.Context) error {
	o.totalPhases = len(o.workflow.Phases)

	for i, phase := range o.workflow.Phases {
		select {
		case <-ctx.Done():
			o.setState(StateCancelled)
			return fmt.Errorf("workflow cancelled: %w", ctx.Err())
		default:
		}

		o.mu.Lock()
		o.currentPhase = i
		o.mu.Unlock()

		o.sendProgress(i, 0, fmt.Sprintf("Starting phase: %s", phase.Name), 0)

		if err := o.checkPrerequisites(phase.Prerequisites); err != nil {
			return fmt.Errorf("phase %d prerequisites not met: %w", i, err)
		}

		result, err := o.executePhase(ctx, i, phase)
		if err != nil && !phase.ContinueOnError {
			o.setState(StateFailed)
			return fmt.Errorf("phase %d failed: %w", i, err)
		}

		if o.workflow.OnPhaseComplete != nil {
			if err := o.workflow.OnPhaseComplete(i, result); err != nil {
				return fmt.Errorf("phase completion callback failed: %w", err)
			}
		}

		o.sendProgress(i, len(phase.Tasks), fmt.Sprintf("Phase completed: %s", phase.Name), 100)
	}

	o.setState(StateCompleted)

	workflowResult := o.buildWorkflowResult()
	if o.workflow.OnWorkflowComplete != nil {
		if err := o.workflow.OnWorkflowComplete(workflowResult); err != nil {
			return fmt.Errorf("workflow completion callback failed: %w", err)
		}
	}

	return nil
}
