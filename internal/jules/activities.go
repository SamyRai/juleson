package jules

import (
	"context"
	"fmt"
	"net/url"
	"strings"
	"time"
)

// ListActivitiesOptions controls pagination and official API filters.
type ListActivitiesOptions struct {
	PageSize   int
	PageToken  string
	CreateTime string
}

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
	return c.ListActivitiesWithOptions(ctx, sessionID, &ListActivitiesOptions{
		PageSize:  pageSize,
		PageToken: pageToken,
	})
}

// ListActivitiesWithOptions lists activities with pagination and official createTime filtering.
func (c *Client) ListActivitiesWithOptions(ctx context.Context, sessionID string, options *ListActivitiesOptions) (*ActivitiesResponse, error) {
	if sessionID == "" {
		return nil, fmt.Errorf("session ID is required")
	}
	pageSize := 50
	pageToken := ""
	createTime := ""
	if options != nil {
		pageSize = options.PageSize
		pageToken = options.PageToken
		createTime = options.CreateTime
	}
	if pageSize <= 0 {
		pageSize = 50 // default page size per API docs
	}
	if pageSize > 100 {
		pageSize = 100 // max page size per API docs
	}

	query := url.Values{}
	query.Set("pageSize", fmt.Sprintf("%d", pageSize))
	if pageToken != "" {
		query.Set("pageToken", pageToken)
	}
	if createTime != "" {
		query.Set("createTime", createTime)
	}

	url := fmt.Sprintf("%s/sessions/%s/activities?%s", c.BaseURL, sessionID, query.Encode())

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
	Type         string `json:"type,omitempty"`         // Client-side filter by activity type.
	Status       string `json:"status,omitempty"`       // Client-side filter by status if the response includes one.
	CreateTime   string `json:"createTime,omitempty"`   // Official API createTime filter.
	Before       string `json:"before,omitempty"`       // Filter activities before this timestamp (ISO 8601)
	After        string `json:"after,omitempty"`        // Filter activities after this timestamp (ISO 8601)
	HasPlan      *bool  `json:"hasPlan,omitempty"`      // Client-side filter for plan activities.
	HasArtifacts *bool  `json:"hasArtifacts,omitempty"` // Client-side filter for activities with artifacts.
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

	options := &ListActivitiesOptions{}
	if filter != nil {
		options.CreateTime = filter.CreateTime
		if options.CreateTime == "" {
			options.CreateTime = filter.After
		}
	}

	response, err := c.ListActivitiesWithOptions(ctx, sessionID, options)
	if err != nil {
		return nil, fmt.Errorf("failed to list filtered activities: %w", err)
	}

	return filterActivities(response.Activities, filter), nil
}

// SearchActivities searches activities within a session
func (c *Client) SearchActivities(ctx context.Context, sessionID string, options *ActivitySearchOptions) ([]Activity, error) {
	if sessionID == "" {
		return nil, fmt.Errorf("session ID is required")
	}

	query := url.Values{}
	if options != nil {
		if options.Query != "" {
			query.Set("q", options.Query)
		}
		if options.Limit > 0 {
			query.Set("limit", fmt.Sprintf("%d", options.Limit))
		}
	}

	requestURL := fmt.Sprintf("%s/sessions/%s/activities/search", c.BaseURL, sessionID)
	if encoded := query.Encode(); encoded != "" {
		requestURL += "?" + encoded
	}

	var activities []Activity
	if err := c.doRequestWithJSON(ctx, "GET", requestURL, nil, &activities); err != nil {
		return nil, fmt.Errorf("failed to search activities: %w", err)
	}

	if options != nil {
		return filterActivities(activities, options.Filter), nil
	}
	return activities, nil
}

func filterActivities(activities []Activity, filter *ActivityFilter) []Activity {
	if filter == nil {
		return activities
	}

	filtered := make([]Activity, 0, len(activities))
	for _, activity := range activities {
		if filter.Type != "" && !activityMatchesType(activity, filter.Type) {
			continue
		}
		if filter.Status != "" && !strings.EqualFold(activity.Status, filter.Status) {
			continue
		}
		if filter.Before != "" && activity.CreateTime != "" && activity.CreateTime >= filter.Before {
			continue
		}
		if filter.After != "" && activity.CreateTime != "" && activity.CreateTime < filter.After {
			continue
		}
		if filter.HasPlan != nil && activityHasPlan(activity) != *filter.HasPlan {
			continue
		}
		if filter.HasArtifacts != nil && (len(activity.Artifacts) > 0) != *filter.HasArtifacts {
			continue
		}
		filtered = append(filtered, activity)
	}

	return filtered
}

func activityMatchesType(activity Activity, activityType string) bool {
	normalized := strings.ToLower(activityType)
	switch {
	case strings.Contains(normalized, "plan") && activityHasPlan(activity):
		return true
	case strings.Contains(normalized, "message") && (activity.UserMessaged != nil || activity.AgentMessaged != nil):
		return true
	case strings.Contains(normalized, "user") && activity.UserMessaged != nil:
		return true
	case strings.Contains(normalized, "agent") && activity.AgentMessaged != nil:
		return true
	case strings.Contains(normalized, "progress") && activity.ProgressUpdated != nil:
		return true
	case strings.Contains(normalized, "complete") && activity.SessionCompleted != nil:
		return true
	case strings.Contains(normalized, "fail") && activity.SessionFailed != nil:
		return true
	case strings.Contains(normalized, "artifact") && len(activity.Artifacts) > 0:
		return true
	}

	haystack := strings.ToLower(strings.Join([]string{
		activity.Name,
		activity.Description,
		activity.Originator,
	}, " "))
	return strings.Contains(haystack, normalized)
}

func activityHasPlan(activity Activity) bool {
	return activity.PlanGenerated != nil || activity.PlanApproved != nil
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
