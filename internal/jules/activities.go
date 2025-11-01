package jules

import (
	"context"
	"fmt"
	"time"
)

// ListActivities lists activities within a session
// Deprecated: Use ListActivitiesWithPagination for full pagination support
func (c *Client) ListActivities(ctx context.Context, sessionID string, pageSize int) ([]Activity, error) {
	response, err := c.ListActivitiesWithPagination(ctx, sessionID, pageSize, "")
	if err != nil {
		return nil, err
	}
	return response.Activities, nil
}

// ListActivitiesWithPagination lists activities within a session with full pagination support
func (c *Client) ListActivitiesWithPagination(ctx context.Context, sessionID string, pageSize int, pageToken string) (*ActivitiesResponse, error) {
	if sessionID == "" {
		return nil, fmt.Errorf("session ID is required")
	}
	if pageSize <= 0 {
		pageSize = 10 // default page size
	}

	url := fmt.Sprintf("%s/sessions/%s/activities?pageSize=%d", c.BaseURL, sessionID, pageSize)
	if pageToken != "" {
		url += fmt.Sprintf("&pageToken=%s", pageToken)
	}

	var response ActivitiesResponse
	if err := c.doRequestWithJSON(ctx, "GET", url, nil, &response); err != nil {
		return nil, fmt.Errorf("failed to list activities: %w", err)
	}

	return &response, nil
}

// GetActivity retrieves a specific activity by ID within a session
func (c *Client) GetActivity(ctx context.Context, sessionID, activityID string) (*Activity, error) {
	if sessionID == "" {
		return nil, fmt.Errorf("session ID is required")
	}
	if activityID == "" {
		return nil, fmt.Errorf("activity ID is required")
	}

	url := fmt.Sprintf("%s/sessions/%s/activities/%s", c.BaseURL, sessionID, activityID)

	var activity Activity
	if err := c.doRequestWithJSON(ctx, "GET", url, nil, &activity); err != nil {
		return nil, fmt.Errorf("failed to get activity: %w", err)
	}

	return &activity, nil
}

// ActivityFilter represents filters for activity listing
type ActivityFilter struct {
	Type         string `json:"type,omitempty"`         // Filter by activity type (e.g., "message", "plan", "execution")
	Status       string `json:"status,omitempty"`       // Filter by status (e.g., "pending", "completed", "failed")
	Before       string `json:"before,omitempty"`       // Filter activities before this timestamp (ISO 8601)
	After        string `json:"after,omitempty"`        // Filter activities after this timestamp (ISO 8601)
	HasPlan      *bool  `json:"hasPlan,omitempty"`      // Filter activities that have/don't have plans
	HasArtifacts *bool  `json:"hasArtifacts,omitempty"` // Filter activities that have/don't have artifacts
}

// ActivitySearchOptions represents search options for activities
type ActivitySearchOptions struct {
	Query  string          `json:"query,omitempty"`  // Search query for activity content
	Filter *ActivityFilter `json:"filter,omitempty"` // Additional filters
	Limit  int             `json:"limit,omitempty"`  // Maximum number of results
}

// ListActivitiesFiltered lists activities with advanced filtering
func (c *Client) ListActivitiesFiltered(ctx context.Context, sessionID string, filter *ActivityFilter) ([]Activity, error) {
	if sessionID == "" {
		return nil, fmt.Errorf("session ID is required")
	}

	url := fmt.Sprintf("%s/sessions/%s/activities", c.BaseURL, sessionID)

	// Build query parameters
	params := make(map[string]string)
	if filter != nil {
		if filter.Type != "" {
			params["type"] = filter.Type
		}
		if filter.Status != "" {
			params["status"] = filter.Status
		}
		if filter.Before != "" {
			params["before"] = filter.Before
		}
		if filter.After != "" {
			params["after"] = filter.After
		}
		if filter.HasPlan != nil {
			params["hasPlan"] = fmt.Sprintf("%t", *filter.HasPlan)
		}
		if filter.HasArtifacts != nil {
			params["hasArtifacts"] = fmt.Sprintf("%t", *filter.HasArtifacts)
		}
	}

	// Add query parameters to URL
	if len(params) > 0 {
		url += "?"
		for key, value := range params {
			url += fmt.Sprintf("%s=%s&", key, value)
		}
		url = url[:len(url)-1] // Remove trailing &
	}

	var activities []Activity
	if err := c.doRequestWithJSON(ctx, "GET", url, nil, &activities); err != nil {
		return nil, fmt.Errorf("failed to list filtered activities: %w", err)
	}

	return activities, nil
}

// SearchActivities searches activities within a session
func (c *Client) SearchActivities(ctx context.Context, sessionID string, options *ActivitySearchOptions) ([]Activity, error) {
	if sessionID == "" {
		return nil, fmt.Errorf("session ID is required")
	}

	url := fmt.Sprintf("%s/sessions/%s/activities/search", c.BaseURL, sessionID)

	// Build query parameters
	params := make(map[string]string)
	if options != nil {
		if options.Query != "" {
			params["q"] = options.Query
		}
		if options.Limit > 0 {
			params["limit"] = fmt.Sprintf("%d", options.Limit)
		}
		if options.Filter != nil {
			filter := options.Filter
			if filter.Type != "" {
				params["type"] = filter.Type
			}
			if filter.Status != "" {
				params["status"] = filter.Status
			}
			if filter.Before != "" {
				params["before"] = filter.Before
			}
			if filter.After != "" {
				params["after"] = filter.After
			}
			if filter.HasPlan != nil {
				params["hasPlan"] = fmt.Sprintf("%t", *filter.HasPlan)
			}
			if filter.HasArtifacts != nil {
				params["hasArtifacts"] = fmt.Sprintf("%t", *filter.HasArtifacts)
			}
		}
	}

	// Add query parameters to URL
	if len(params) > 0 {
		url += "?"
		for key, value := range params {
			url += fmt.Sprintf("%s=%s&", key, value)
		}
		url = url[:len(url)-1] // Remove trailing &
	}

	var activities []Activity
	if err := c.doRequestWithJSON(ctx, "GET", url, nil, &activities); err != nil {
		return nil, fmt.Errorf("failed to search activities: %w", err)
	}

	return activities, nil
}

// GetActivitiesByType retrieves activities of a specific type
func (c *Client) GetActivitiesByType(ctx context.Context, sessionID, activityType string) ([]Activity, error) {
	filter := &ActivityFilter{Type: activityType}
	return c.ListActivitiesFiltered(ctx, sessionID, filter)
}

// GetActivitiesWithPlans retrieves activities that have generated plans
func (c *Client) GetActivitiesWithPlans(ctx context.Context, sessionID string) ([]Activity, error) {
	hasPlan := true
	filter := &ActivityFilter{HasPlan: &hasPlan}
	return c.ListActivitiesFiltered(ctx, sessionID, filter)
}

// GetActivitiesWithArtifacts retrieves activities that have artifacts
func (c *Client) GetActivitiesWithArtifacts(ctx context.Context, sessionID string) ([]Activity, error) {
	hasArtifacts := true
	filter := &ActivityFilter{HasArtifacts: &hasArtifacts}
	return c.ListActivitiesFiltered(ctx, sessionID, filter)
}

// GetRecentActivities retrieves activities from the last N hours
func (c *Client) GetRecentActivities(ctx context.Context, sessionID string, hours int) ([]Activity, error) {
	if hours <= 0 {
		return nil, fmt.Errorf("hours must be positive")
	}

	after := time.Now().Add(-time.Duration(hours) * time.Hour).Format(time.RFC3339)
	filter := &ActivityFilter{After: after}
	return c.ListActivitiesFiltered(ctx, sessionID, filter)
}
