package jules

import (
	"context"
	"fmt"
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
		return nil, fmt.Errorf("failed to create session: %w", err)
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
		pageSize = 10 // default page size
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
		return fmt.Errorf("failed to send message: %w", err)
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
		return fmt.Errorf("failed to approve plan: %w", err)
	}

	return nil
}

// CancelSession cancels a running session
func (c *Client) CancelSession(ctx context.Context, sessionID string) error {
	if sessionID == "" {
		return fmt.Errorf("session ID is required")
	}

	url := fmt.Sprintf("%s/sessions/%s:cancel", c.BaseURL, sessionID)

	if err := c.doRequestWithJSON(ctx, "POST", url, nil, nil); err != nil {
		return fmt.Errorf("failed to cancel session: %w", err)
	}

	return nil
}

// DeleteSession deletes a session and all its activities
func (c *Client) DeleteSession(ctx context.Context, sessionID string) error {
	if sessionID == "" {
		return fmt.Errorf("session ID is required")
	}

	url := fmt.Sprintf("%s/sessions/%s", c.BaseURL, sessionID)

	if err := c.doRequestWithJSON(ctx, "DELETE", url, nil, nil); err != nil {
		return fmt.Errorf("failed to delete session: %w", err)
	}

	return nil
}
