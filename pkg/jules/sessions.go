package jules

import (
	"context"
	"fmt"
	"net/url"
)

// CreateSession creates a new coding session
func (c *Client) CreateSession(ctx context.Context, req *CreateSessionRequest) (*Session, error) {
	if req == nil {
		return nil, fmt.Errorf("request cannot be nil")
	}
	if req.Prompt == "" {
		return nil, fmt.Errorf("prompt is required")
	}
	if req.SourceContext != nil && req.SourceContext.Source != "" {
		req = cloneCreateSessionRequest(req)
		req.SourceContext.Source = NormalizeSourceName(req.SourceContext.Source)
	}

	requestURL := fmt.Sprintf("%s/sessions", c.BaseURL)

	var session Session
	if err := c.doRequestWithJSON(ctx, "POST", requestURL, req, &session); err != nil {
		return nil, fmt.Errorf("failed to create session: %w", err)
	}

	return &session, nil
}

// GetSession retrieves a specific session by ID
func (c *Client) GetSession(ctx context.Context, sessionID string) (*Session, error) {
	if sessionID == "" {
		return nil, fmt.Errorf("session ID is required")
	}

	resourcePath, err := sessionPath(sessionID)
	if err != nil {
		return nil, err
	}
	requestURL := fmt.Sprintf("%s/%s", c.BaseURL, resourcePath)

	var session Session
	if err := c.doRequestWithJSON(ctx, "GET", requestURL, nil, &session); err != nil {
		return nil, fmt.Errorf("failed to get session: %w", err)
	}

	return &session, nil
}

// DeleteSession deletes a session by ID.
func (c *Client) DeleteSession(ctx context.Context, sessionID string) error {
	if sessionID == "" {
		return fmt.Errorf("session ID is required")
	}

	resourcePath, err := sessionPath(sessionID)
	if err != nil {
		return err
	}
	requestURL := fmt.Sprintf("%s/%s", c.BaseURL, resourcePath)
	if err := c.doRequestWithJSON(ctx, "DELETE", requestURL, nil, nil); err != nil {
		return fmt.Errorf("failed to delete session: %w", err)
	}

	return nil
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

	query := url.Values{}
	query.Set("pageSize", fmt.Sprintf("%d", pageSize))
	if pageToken != "" {
		query.Set("pageToken", pageToken)
	}
	requestURL := fmt.Sprintf("%s/sessions?%s", c.BaseURL, query.Encode())

	var response SessionsResponse
	if err := c.doRequestWithJSON(ctx, "GET", requestURL, nil, &response); err != nil {
		return nil, fmt.Errorf("failed to list sessions: %w", err)
	}

	return &response, nil
}

// ListAllSessions retrieves every session by following nextPageToken.
func (c *Client) ListAllSessions(ctx context.Context, pageSize int) ([]Session, error) {
	var sessions []Session
	pageToken := ""
	for {
		response, err := c.ListSessionsWithPagination(ctx, pageSize, pageToken)
		if err != nil {
			return nil, err
		}
		sessions = append(sessions, response.Sessions...)
		if response.NextPageToken == "" {
			return sessions, nil
		}
		pageToken = response.NextPageToken
	}
}

// SendMessage sends a message to Jules within a session
func (c *Client) SendMessage(ctx context.Context, sessionID string, req *SendMessageRequest) error {
	if sessionID == "" {
		return fmt.Errorf("session ID is required")
	}
	if req == nil || req.Prompt == "" {
		return fmt.Errorf("message prompt is required")
	}

	resourcePath, err := sessionPath(sessionID)
	if err != nil {
		return err
	}
	requestURL := fmt.Sprintf("%s/%s:sendMessage", c.BaseURL, resourcePath)

	if err := c.doRequestWithJSON(ctx, "POST", requestURL, req, nil); err != nil {
		return fmt.Errorf("failed to send message: %w", err)
	}

	return nil
}

// ApprovePlan approves a plan in a session
func (c *Client) ApprovePlan(ctx context.Context, sessionID string) error {
	if sessionID == "" {
		return fmt.Errorf("session ID is required")
	}

	resourcePath, err := sessionPath(sessionID)
	if err != nil {
		return err
	}
	requestURL := fmt.Sprintf("%s/%s:approvePlan", c.BaseURL, resourcePath)

	if err := c.doRequestWithJSON(ctx, "POST", requestURL, nil, nil); err != nil {
		return fmt.Errorf("failed to approve plan: %w", err)
	}

	return nil
}

func cloneCreateSessionRequest(req *CreateSessionRequest) *CreateSessionRequest {
	clone := *req
	if req.SourceContext != nil {
		sourceContext := *req.SourceContext
		if req.SourceContext.GithubRepoContext != nil {
			githubRepoContext := *req.SourceContext.GithubRepoContext
			sourceContext.GithubRepoContext = &githubRepoContext
		}
		clone.SourceContext = &sourceContext
	}
	return &clone
}
