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
	CreateTime time.Time
}

// ActivityFilter represents client-side filters for documented activity data.
type ActivityFilter struct {
	Type         string
	Status       string
	CreateTime   time.Time
	Before       time.Time
	After        time.Time
	HasPlan      *bool
	HasArtifacts *bool
}

// ActivitySearchOptions represents client-side search options for activities.
type ActivitySearchOptions struct {
	Query  string
	Filter *ActivityFilter
	Limit  int
}

// ListActivities lists activities within a session.
//
// Deprecated: Use ListActivitiesWithPagination for paginated responses.
func (c *Client) ListActivities(ctx context.Context, sessionID string, pageSize int) ([]Activity, error) {
	response, err := c.ListActivitiesWithPagination(ctx, sessionID, pageSize, "")
	if err != nil {
		return nil, err
	}
	return response.Activities, nil
}

// ListActivitiesWithPagination lists activities within a session with pagination support.
func (c *Client) ListActivitiesWithPagination(ctx context.Context, sessionID string, pageSize int, pageToken string) (*ActivitiesResponse, error) {
	return c.ListActivitiesWithOptions(ctx, sessionID, &ListActivitiesOptions{
		PageSize:  pageSize,
		PageToken: pageToken,
	})
}

// ListActivitiesWithOptions lists activities with pagination. When CreateTime
// is set, activities are filtered client-side because the Jules API currently
// rejects createTime as a list query parameter.
func (c *Client) ListActivitiesWithOptions(ctx context.Context, sessionID string, options *ListActivitiesOptions) (*ActivitiesResponse, error) {
	if sessionID == "" {
		return nil, fmt.Errorf("session ID is required")
	}
	pageSize := 50
	pageToken := ""
	createTime := time.Time{}
	if options != nil {
		pageSize = options.PageSize
		pageToken = options.PageToken
		createTime = options.CreateTime
	}
	if pageSize <= 0 {
		pageSize = 50
	}
	if pageSize > 100 {
		pageSize = 100
	}

	resourcePath, err := sessionPath(sessionID)
	if err != nil {
		return nil, err
	}

	query := url.Values{}
	query.Set("pageSize", fmt.Sprintf("%d", pageSize))
	if pageToken != "" {
		query.Set("pageToken", pageToken)
	}

	requestURL := fmt.Sprintf("%s/%s/activities?%s", c.BaseURL, resourcePath, query.Encode())

	var response ActivitiesResponse
	if err := c.doRequestWithJSON(ctx, "GET", requestURL, nil, &response); err != nil {
		return nil, fmt.Errorf("failed to list activities: %w", err)
	}
	if !createTime.IsZero() {
		response.Activities = activitiesAtOrAfter(response.Activities, createTime)
	}

	return &response, nil
}

// ListAllActivities retrieves every activity by following nextPageToken.
func (c *Client) ListAllActivities(ctx context.Context, sessionID string, pageSize int) ([]Activity, error) {
	var activities []Activity
	pageToken := ""
	for {
		response, err := c.ListActivitiesWithOptions(ctx, sessionID, &ListActivitiesOptions{
			PageSize:  pageSize,
			PageToken: pageToken,
		})
		if err != nil {
			return nil, err
		}
		activities = append(activities, response.Activities...)
		if response.NextPageToken == "" {
			return activities, nil
		}
		pageToken = response.NextPageToken
	}
}

// ListActivitiesSince retrieves activities created at or after the cursor time.
func (c *Client) ListActivitiesSince(ctx context.Context, sessionID string, cursor time.Time, pageSize int) ([]Activity, error) {
	var activities []Activity
	pageToken := ""
	for {
		response, err := c.ListActivitiesWithOptions(ctx, sessionID, &ListActivitiesOptions{
			PageSize:   pageSize,
			PageToken:  pageToken,
			CreateTime: cursor,
		})
		if err != nil {
			return nil, err
		}
		activities = append(activities, response.Activities...)
		if response.NextPageToken == "" {
			return activities, nil
		}
		pageToken = response.NextPageToken
	}
}

func activitiesAtOrAfter(activities []Activity, cursor time.Time) []Activity {
	if cursor.IsZero() {
		return activities
	}
	filtered := make([]Activity, 0, len(activities))
	for _, activity := range activities {
		if activity.CreateTime.IsZero() {
			continue
		}
		if !activity.CreateTime.Before(cursor) {
			filtered = append(filtered, activity)
		}
	}
	return filtered
}

// ActivityCursor returns the latest createTime in the provided activities.
func ActivityCursor(activities []Activity) time.Time {
	var cursor time.Time
	for _, activity := range activities {
		if activity.CreateTime.After(cursor) {
			cursor = activity.CreateTime
		}
	}
	return cursor
}

// GetActivity retrieves a specific activity by ID or resource name.
func (c *Client) GetActivity(ctx context.Context, sessionID, activityID string) (*Activity, error) {
	if activityID == "" {
		return nil, fmt.Errorf("activity ID is required")
	}

	resourcePath, err := activityPath(sessionID, activityID)
	if err != nil {
		return nil, err
	}
	requestURL := fmt.Sprintf("%s/%s", c.BaseURL, resourcePath)

	var activity Activity
	if err := c.doRequestWithJSON(ctx, "GET", requestURL, nil, &activity); err != nil {
		return nil, fmt.Errorf("failed to get activity: %w", err)
	}

	return &activity, nil
}

// ListActivitiesFiltered lists activities and applies client-side filters over documented fields.
func (c *Client) ListActivitiesFiltered(ctx context.Context, sessionID string, filter *ActivityFilter) ([]Activity, error) {
	options := &ListActivitiesOptions{}
	if filter != nil {
		options.CreateTime = filter.CreateTime
		if options.CreateTime.IsZero() {
			options.CreateTime = filter.After
		}
	}

	response, err := c.ListActivitiesWithOptions(ctx, sessionID, options)
	if err != nil {
		return nil, fmt.Errorf("failed to list filtered activities: %w", err)
	}

	return FilterActivities(response.Activities, filter), nil
}

// SearchActivities searches documented activity payloads client-side. It does
// not call an undocumented search endpoint.
func (c *Client) SearchActivities(ctx context.Context, sessionID string, options *ActivitySearchOptions) ([]Activity, error) {
	activities, err := c.ListAllActivities(ctx, sessionID, 100)
	if err != nil {
		return nil, fmt.Errorf("failed to list activities for search: %w", err)
	}

	if options == nil {
		return activities, nil
	}
	activities = FilterActivities(activities, options.Filter)
	if options.Query != "" {
		activities = searchActivityPayloads(activities, options.Query)
	}
	if options.Limit > 0 && len(activities) > options.Limit {
		activities = activities[:options.Limit]
	}
	return activities, nil
}

// FilterActivities applies client-side filters over documented activity fields.
func FilterActivities(activities []Activity, filter *ActivityFilter) []Activity {
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
		if !filter.Before.IsZero() && !activity.CreateTime.IsZero() && !activity.CreateTime.Before(filter.Before) {
			continue
		}
		if !filter.After.IsZero() && !activity.CreateTime.IsZero() && activity.CreateTime.Before(filter.After) {
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

	return strings.Contains(strings.ToLower(activitySearchText(activity)), normalized)
}

func activityHasPlan(activity Activity) bool {
	return activity.PlanGenerated != nil || activity.PlanApproved != nil
}

func searchActivityPayloads(activities []Activity, query string) []Activity {
	query = strings.ToLower(strings.TrimSpace(query))
	if query == "" {
		return activities
	}
	filtered := make([]Activity, 0, len(activities))
	for _, activity := range activities {
		if strings.Contains(strings.ToLower(activitySearchText(activity)), query) {
			filtered = append(filtered, activity)
		}
	}
	return filtered
}

func activitySearchText(activity Activity) string {
	parts := []string{
		activity.Name,
		activity.Description,
		string(activity.Originator),
		activity.Status,
	}
	if activity.UserMessaged != nil {
		parts = append(parts, activity.UserMessaged.UserMessage)
	}
	if activity.AgentMessaged != nil {
		parts = append(parts, activity.AgentMessaged.AgentMessage)
	}
	if activity.ProgressUpdated != nil {
		parts = append(parts, activity.ProgressUpdated.Title, activity.ProgressUpdated.Description)
	}
	if activity.SessionFailed != nil {
		parts = append(parts, activity.SessionFailed.Reason)
	}
	for _, artifact := range activity.Artifacts {
		if artifact.BashOutput != nil {
			parts = append(parts, artifact.BashOutput.Command, artifact.BashOutput.Output)
		}
		if artifact.ChangeSet != nil {
			parts = append(parts, artifact.ChangeSet.Source)
			if artifact.ChangeSet.GitPatch != nil {
				parts = append(parts, artifact.ChangeSet.GitPatch.SuggestedCommitMessage, artifact.ChangeSet.GitPatch.UnidiffPatch)
			}
		}
		if artifact.Media != nil {
			parts = append(parts, artifact.Media.MimeType)
		}
	}
	return strings.Join(parts, " ")
}

// GetActivitiesByType retrieves activities of a specific type.
func (c *Client) GetActivitiesByType(ctx context.Context, sessionID, activityType string) ([]Activity, error) {
	filter := &ActivityFilter{Type: activityType}
	return c.ListActivitiesFiltered(ctx, sessionID, filter)
}

// GetActivitiesWithPlans retrieves activities that have generated or approved plans.
func (c *Client) GetActivitiesWithPlans(ctx context.Context, sessionID string) ([]Activity, error) {
	hasPlan := true
	filter := &ActivityFilter{HasPlan: &hasPlan}
	return c.ListActivitiesFiltered(ctx, sessionID, filter)
}

// GetActivitiesWithArtifacts retrieves activities that have artifacts.
func (c *Client) GetActivitiesWithArtifacts(ctx context.Context, sessionID string) ([]Activity, error) {
	hasArtifacts := true
	filter := &ActivityFilter{HasArtifacts: &hasArtifacts}
	return c.ListActivitiesFiltered(ctx, sessionID, filter)
}

// GetRecentActivities retrieves activities from the last N hours.
func (c *Client) GetRecentActivities(ctx context.Context, sessionID string, hours int) ([]Activity, error) {
	if hours <= 0 {
		return nil, fmt.Errorf("hours must be positive")
	}

	filter := &ActivityFilter{After: time.Now().Add(-time.Duration(hours) * time.Hour)}
	return c.ListActivitiesFiltered(ctx, sessionID, filter)
}
