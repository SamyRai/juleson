package jules

import (
	"context"
	"fmt"

	"github.com/SamyRai/juleson/internal/events"
)

// CreateSession creates a new coding session
func (c *Client) CreateSession(ctx context.Context, req *CreateSessionRequest) (*Session, error) {
	if req == nil {
		return nil, fmt.Errorf("request cannot be nil")
	}
	if req.Prompt == "" {
		return nil, fmt.Errorf("prompt is required")
	}

	url := fmt.Sprintf("%s/sessions", c.BaseURL)

	var session Session
	if err := c.doRequestWithJSON(ctx, "POST", url, req, &session); err != nil {
		// Emit session creation failed event if emitter is available
		if c.EventEmitter != nil {
			c.EventEmitter.EmitSessionEvent(ctx, events.EventSessionFailed, events.SessionEventData{
				SessionID: "",
				State:     "failed",
				Error:     err.Error(),
			})
		}
		return nil, fmt.Errorf("failed to create session: %w", err)
	}

	// Emit session created event if emitter is available
	if c.EventEmitter != nil {
		c.EventEmitter.EmitSessionEvent(ctx, events.EventSessionCreated, events.SessionEventData{
			SessionID: session.ID,
			State:     session.State,
			Title:     session.Title,
			URL:       session.URL,
		})
	}

	return &session, nil
}

// GetSession retrieves a specific session by ID
func (c *Client) GetSession(ctx context.Context, sessionID string) (*Session, error) {
	if sessionID == "" {
		return nil, fmt.Errorf("session ID is required")
	}

	url := fmt.Sprintf("%s/sessions/%s", c.BaseURL, sessionID)

	var session Session
	if err := c.doRequestWithJSON(ctx, "GET", url, nil, &session); err != nil {
		return nil, fmt.Errorf("failed to get session: %w", err)
	}

	return &session, nil
}

// ListSessions lists all sessions with pagination support
// Deprecated: Use ListSessionsWithPagination for full pagination support
func (c *Client) ListSessions(ctx context.Context, pageSize int) ([]Session, error) {
	response, err := c.ListSessionsWithPagination(ctx, pageSize, "")
	if err != nil {
		return nil, err
	}
	return response.Sessions, nil
}

// ListSessionsWithPagination lists all sessions with full pagination support
func (c *Client) ListSessionsWithPagination(ctx context.Context, pageSize int, pageToken string) (*SessionsResponse, error) {
	if pageSize <= 0 {
		pageSize = 30 // default page size per API docs
	}
	if pageSize > 100 {
		pageSize = 100 // max page size per API docs
	}

	url := fmt.Sprintf("%s/sessions?pageSize=%d", c.BaseURL, pageSize)
	if pageToken != "" {
		url += fmt.Sprintf("&pageToken=%s", pageToken)
	}

	var response SessionsResponse
	if err := c.doRequestWithJSON(ctx, "GET", url, nil, &response); err != nil {
		return nil, fmt.Errorf("failed to list sessions: %w", err)
	}

	return &response, nil
}

// SendMessage sends a message to Jules within a session
func (c *Client) SendMessage(ctx context.Context, sessionID string, req *SendMessageRequest) error {
	if sessionID == "" {
		return fmt.Errorf("session ID is required")
	}
	if req == nil || req.Prompt == "" {
		return fmt.Errorf("message prompt is required")
	}

	url := fmt.Sprintf("%s/sessions/%s:sendMessage", c.BaseURL, sessionID)

	if err := c.doRequestWithJSON(ctx, "POST", url, req, nil); err != nil {
		// Emit activity event for failed message
		if c.EventEmitter != nil {
			c.EventEmitter.EmitActivityEvent(ctx, events.EventActivityProcessed, events.ActivityEventData{
				SessionID:    sessionID,
				ActivityID:   "",
				ActivityType: "message",
				Originator:   "user",
				Description:  "Failed to send message",
				Artifacts:    0,
				Metadata: map[string]interface{}{
					"error":  err.Error(),
					"prompt": req.Prompt,
				},
			})
		}
		return fmt.Errorf("failed to send message: %w", err)
	}

	// Emit activity event for successful message
	if c.EventEmitter != nil {
		c.EventEmitter.EmitActivityEvent(ctx, events.EventActivityProcessed, events.ActivityEventData{
			SessionID:    sessionID,
			ActivityID:   "",
			ActivityType: "message",
			Originator:   "user",
			Description:  "Message sent successfully",
			Artifacts:    0,
			Metadata: map[string]interface{}{
				"prompt": req.Prompt,
			},
		})
	}

	return nil
}

// ApprovePlan approves a plan in a session
func (c *Client) ApprovePlan(ctx context.Context, sessionID string) error {
	if sessionID == "" {
		return fmt.Errorf("session ID is required")
	}

	url := fmt.Sprintf("%s/sessions/%s:approvePlan", c.BaseURL, sessionID)

	if err := c.doRequestWithJSON(ctx, "POST", url, nil, nil); err != nil {
		// Emit activity event for failed plan approval
		if c.EventEmitter != nil {
			c.EventEmitter.EmitActivityEvent(ctx, events.EventActivityProcessed, events.ActivityEventData{
				SessionID:    sessionID,
				ActivityID:   "",
				ActivityType: "plan_approval",
				Originator:   "user",
				Description:  "Failed to approve plan",
				Artifacts:    0,
				Metadata: map[string]interface{}{
					"error": err.Error(),
				},
			})
		}
		return fmt.Errorf("failed to approve plan: %w", err)
	}

	// Emit activity event for successful plan approval
	if c.EventEmitter != nil {
		c.EventEmitter.EmitActivityEvent(ctx, events.EventPlanApproved, events.ActivityEventData{
			SessionID:    sessionID,
			ActivityID:   "",
			ActivityType: "plan_approval",
			Originator:   "user",
			Description:  "Plan approved successfully",
			Artifacts:    0,
		})
	}

	return nil
}

// NOTE: The Jules API v1alpha does NOT support cancel or delete operations.
// These operations are only available through the Jules web UI.
// See: https://developers.google.com/jules/api/reference/rest/v1alpha/sessions
//
// To manage sessions:
// - Cancel: Use the web UI at the URL returned in session.URL
// - Delete: Use the web UI
// - Monitor state: Use GetSession to check if state is FAILED, COMPLETED, or PAUSED
