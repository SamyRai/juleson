package jules

import (
	"context"
	"fmt"
	"time"

	"errors"
)

// SessionState represents the current state of a session
type SessionState string

const (
	SessionStatePlanning   SessionState = "PLANNING"
	SessionStateInProgress SessionState = "IN_PROGRESS"
	SessionStateCompleted  SessionState = "COMPLETED"
	SessionStateFailed     SessionState = "FAILED"
	SessionStateCancelled  SessionState = "CANCELLED"
)

// SessionStatus represents the current status of a session
type SessionStatus struct {
	Session   *Session
	State     SessionState
	IsActive  bool
	IsDone    bool
	IsSuccess bool
	Error     string
}

// SessionMonitor provides session monitoring capabilities
type SessionMonitor struct {
	client     *Client
	sessionID  string
	interval   time.Duration
	maxWait    time.Duration
	onProgress func(*SessionStatus)
	onComplete func(*SessionStatus)
}

// NewSessionMonitor creates a new session monitor
func NewSessionMonitor(client *Client, sessionID string) *SessionMonitor {
	return &SessionMonitor{
		client:    client,
		sessionID: sessionID,
		interval:  5 * time.Second,  // Check every 5 seconds
		maxWait:   30 * time.Minute, // Wait up to 30 minutes
	}
}

// WithInterval sets the polling interval
func (sm *SessionMonitor) WithInterval(interval time.Duration) *SessionMonitor {
	sm.interval = interval
	return sm
}

// WithMaxWait sets the maximum wait time
func (sm *SessionMonitor) WithMaxWait(maxWait time.Duration) *SessionMonitor {
	sm.maxWait = maxWait
	return sm
}

// OnProgress sets the progress callback
func (sm *SessionMonitor) OnProgress(callback func(*SessionStatus)) *SessionMonitor {
	sm.onProgress = callback
	return sm
}

// OnComplete sets the completion callback
func (sm *SessionMonitor) OnComplete(callback func(*SessionStatus)) *SessionMonitor {
	sm.onComplete = callback
	return sm
}

// WaitForCompletion waits for the session to complete
func (sm *SessionMonitor) WaitForCompletion(ctx context.Context) (*SessionStatus, error) {
	return sm.pollUntilComplete(ctx, true)
}

// PollUntilComplete polls the session until it reaches a terminal state
func (sm *SessionMonitor) PollUntilComplete(ctx context.Context) (*SessionStatus, error) {
	return sm.pollUntilComplete(ctx, true)
}

// pollUntilComplete implements the polling logic
func (sm *SessionMonitor) pollUntilComplete(ctx context.Context, continuous bool) (*SessionStatus, error) {
	ticker := time.NewTicker(sm.interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return nil, fmt.Errorf("context cancelled: %w", ctx.Err())
		case <-time.After(sm.maxWait):
			return nil, fmt.Errorf("timeout waiting for session completion after %v", sm.maxWait)
		case <-ticker.C:
			status, err := sm.getSessionStatus(ctx)
			if err != nil {
				return nil, fmt.Errorf("failed to get session status: %w", err)
			}

			// Call progress callback if set
			if sm.onProgress != nil {
				sm.onProgress(status)
			}

			// Check if session is done
			if status.IsDone {
				// Call completion callback if set
				if sm.onComplete != nil {
					sm.onComplete(status)
				}
				return status, nil
			}

			// If not continuous polling, just return current status
			if !continuous {
				return status, nil
			}
		}
	}
}

// getSessionStatus retrieves the current session status
func (sm *SessionMonitor) getSessionStatus(ctx context.Context) (*SessionStatus, error) {
	session, err := sm.client.GetSession(ctx, sm.sessionID)
	if err != nil {
		// Check if it's a not found error (handle wrapped errors)
		var apiErr *APIError
		if errors.As(err, &apiErr) && apiErr.IsNotFound() {
			return &SessionStatus{
				State:     SessionStateFailed,
				IsActive:  false,
				IsDone:    true,
				IsSuccess: false,
				Error:     "session not found",
			}, nil
		}
		return nil, err
	}

	status := &SessionStatus{
		Session:   session,
		State:     SessionState(session.State),
		IsActive:  session.State == "PLANNING" || session.State == "IN_PROGRESS",
		IsDone:    session.State == "COMPLETED" || session.State == "FAILED" || session.State == "CANCELLED",
		IsSuccess: session.State == "COMPLETED",
	}

	if session.State == "FAILED" {
		status.Error = "session failed"
	} else if session.State == "CANCELLED" {
		status.Error = "session cancelled"
	}

	return status, nil
}

// WaitForPlan waits for a session to generate a plan
func (sm *SessionMonitor) WaitForPlan(ctx context.Context) (*SessionStatus, error) {
	return sm.pollUntilCondition(ctx, func(status *SessionStatus) bool {
		// Get latest activities to check for plan generation
		activities, err := sm.client.ListActivities(ctx, sm.sessionID, 10)
		if err != nil {
			return false
		}

		// Check if any activity has a plan generated
		for _, activity := range activities {
			if activity.PlanGenerated != nil {
				return true
			}
		}
		return false
	})
}

// pollUntilCondition polls until a custom condition is met
func (sm *SessionMonitor) pollUntilCondition(ctx context.Context, condition func(*SessionStatus) bool) (*SessionStatus, error) {
	ticker := time.NewTicker(sm.interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return nil, fmt.Errorf("context cancelled: %w", ctx.Err())
		case <-time.After(sm.maxWait):
			return nil, fmt.Errorf("timeout waiting for condition after %v", sm.maxWait)
		case <-ticker.C:
			status, err := sm.getSessionStatus(ctx)
			if err != nil {
				return nil, err
			}

			if condition(status) {
				return status, nil
			}
		}
	}
}
