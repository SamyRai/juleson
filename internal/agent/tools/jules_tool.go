package tools

import (
	"context"
	"fmt"
	"time"

	"github.com/SamyRai/juleson/internal/agent"
	"github.com/SamyRai/juleson/internal/jules"
)

// JulesTool provides integration with Jules AI
type JulesTool struct {
	client *jules.Client
	config *JulesToolConfig
}

// JulesToolConfig configures the Jules tool
type JulesToolConfig struct {
	RequireApproval    bool
	AutoApprove        bool
	MaxRetries         int
	Timeout            time.Duration
	MinConfidenceScore float64
}

// DefaultJulesToolConfig returns default configuration
func DefaultJulesToolConfig() *JulesToolConfig {
	return &JulesToolConfig{
		RequireApproval:    true,
		AutoApprove:        false,
		MaxRetries:         3,
		Timeout:            10 * time.Minute,
		MinConfidenceScore: 0.7,
	}
}

// NewJulesTool creates a new Jules tool
func NewJulesTool(client *jules.Client, config *JulesToolConfig) *JulesTool {
	if config == nil {
		config = DefaultJulesToolConfig()
	}
	return &JulesTool{
		client: client,
		config: config,
	}
}

// Name returns the tool name
func (j *JulesTool) Name() string {
	return "jules"
}

// Description returns what this tool does
func (j *JulesTool) Description() string {
	return "Execute development tasks using Jules AI. Jules can write code, refactor, add tests, fix bugs, and more."
}

// Parameters returns tool parameters
func (j *JulesTool) Parameters() []Parameter {
	return []Parameter{
		{
			Name:        "action",
			Description: "Action to perform: create_session, send_message, review_plan, apply_patches",
			Type:        ParameterTypeString,
			Required:    true,
		},
		{
			Name:        "session_id",
			Description: "Jules session ID (required for send_message, review_plan, apply_patches)",
			Type:        ParameterTypeString,
			Required:    false,
		},
		{
			Name:        "prompt",
			Description: "Task description for Jules (required for create_session, send_message)",
			Type:        ParameterTypeString,
			Required:    false,
		},
		{
			Name:        "source_id",
			Description: "Source context ID (optional for create_session - sources are managed separately)",
			Type:        ParameterTypeString,
			Required:    false,
		},
		{
			Name:        "approved",
			Description: "Whether to approve the plan (required for review_plan)",
			Type:        ParameterTypeBool,
			Required:    false,
		},
		{
			Name:        "activity_id",
			Description: "Activity ID to apply patches from (required for apply_patches)",
			Type:        ParameterTypeString,
			Required:    false,
		},
	}
}

// Execute runs the Jules tool
func (j *JulesTool) Execute(ctx context.Context, params map[string]interface{}) (*ToolResult, error) {
	start := time.Now()

	action, ok := params["action"].(string)
	if !ok {
		return nil, fmt.Errorf("action parameter is required")
	}

	var result *ToolResult
	var err error

	switch action {
	case "create_session":
		result, err = j.createSession(ctx, params)
	case "send_message":
		result, err = j.sendMessage(ctx, params)
	case "review_plan":
		result, err = j.reviewPlan(ctx, params)
	case "apply_patches":
		result, err = j.applyPatches(ctx, params)
	case "get_activities":
		result, err = j.getActivities(ctx, params)
	default:
		return nil, fmt.Errorf("unknown action: %s", action)
	}

	if err != nil {
		return &ToolResult{
			Success:  false,
			Error:    err,
			Duration: time.Since(start).Milliseconds(),
		}, err
	}

	result.Duration = time.Since(start).Milliseconds()
	return result, nil
}

// RequiresApproval returns whether this tool needs approval
func (j *JulesTool) RequiresApproval() bool {
	return j.config.RequireApproval
}

// CanHandle returns whether this tool can handle a task
func (j *JulesTool) CanHandle(task agent.Task) bool {
	// Jules can handle most development tasks
	return true
}

// createSession creates a new Jules session
func (j *JulesTool) createSession(ctx context.Context, params map[string]interface{}) (*ToolResult, error) {
	prompt, ok := params["prompt"].(string)
	if !ok {
		return nil, fmt.Errorf("prompt parameter is required for create_session")
	}

	sourceID, ok := params["source_id"].(string)
	if !ok {
		return nil, fmt.Errorf("source_id parameter is required for create_session")
	}

	req := &jules.CreateSessionRequest{
		Prompt:         prompt,
		Title:          "CI Configuration Improvement",
		AutomationMode: "AUTO_CREATE_PR",
		SourceContext: &jules.SourceContext{
			Source: sourceID,
			GithubRepoContext: &jules.GithubRepoContext{
				StartingBranch: "main",
			},
		},
	}

	session, err := j.client.CreateSession(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("failed to create Jules session: %w", err)
	}

	return &ToolResult{
		Success: true,
		Output: map[string]interface{}{
			"session_id": session.ID,
			"state":      session.State,
		},
		Metadata: map[string]interface{}{
			"session_id": session.ID,
		},
	}, nil
}

// sendMessage sends a message to an existing Jules session
func (j *JulesTool) sendMessage(ctx context.Context, params map[string]interface{}) (*ToolResult, error) {
	sessionID, ok := params["session_id"].(string)
	if !ok {
		return nil, fmt.Errorf("session_id parameter is required for send_message")
	}

	prompt, ok := params["prompt"].(string)
	if !ok {
		return nil, fmt.Errorf("prompt parameter is required for send_message")
	}

	req := &jules.SendMessageRequest{
		Prompt: prompt,
	}

	err := j.client.SendMessage(ctx, sessionID, req)
	if err != nil {
		return nil, fmt.Errorf("failed to send message to Jules: %w", err)
	}

	// Wait a bit for Jules to process
	time.Sleep(2 * time.Second)

	return &ToolResult{
		Success: true,
		Output: map[string]interface{}{
			"session_id": sessionID,
			"message":    "Message sent successfully",
		},
	}, nil
}

// reviewPlan reviews and approves/rejects a Jules plan
// Note: This is a placeholder until plan approval API is available
func (j *JulesTool) reviewPlan(ctx context.Context, params map[string]interface{}) (*ToolResult, error) {
	sessionID, ok := params["session_id"].(string)
	if !ok {
		return nil, fmt.Errorf("session_id parameter is required for review_plan")
	}

	approved, ok := params["approved"].(bool)
	if !ok {
		return nil, fmt.Errorf("approved parameter is required for review_plan")
	}

	// Get session status
	session, err := j.client.GetSession(ctx, sessionID)
	if err != nil {
		return nil, fmt.Errorf("failed to get session: %w", err)
	}

	// For now, just return the approval decision
	// TODO: Implement actual plan approval when API supports it
	return &ToolResult{
		Success: true,
		Output: map[string]interface{}{
			"session_id": sessionID,
			"state":      session.State,
			"approved":   approved,
		},
	}, nil
}

// applyPatches applies patches from a Jules activity
func (j *JulesTool) applyPatches(ctx context.Context, params map[string]interface{}) (*ToolResult, error) {
	sessionID, ok := params["session_id"].(string)
	if !ok {
		return nil, fmt.Errorf("session_id parameter is required for apply_patches")
	}

	activityID, ok := params["activity_id"].(string)
	if !ok {
		return nil, fmt.Errorf("activity_id parameter is required for apply_patches")
	}

	// Get activity to extract patches
	activities, err := j.client.ListActivities(ctx, sessionID, 100)
	if err != nil {
		return nil, fmt.Errorf("failed to list activities: %w", err)
	}

	var targetActivity *jules.Activity
	for _, activity := range activities {
		if activity.ID == activityID {
			targetActivity = &activity
			break
		}
	}

	if targetActivity == nil {
		return nil, fmt.Errorf("activity %s not found", activityID)
	}

	// Extract changes from artifacts
	var changes []agent.Change
	for _, artifact := range targetActivity.Artifacts {
		if artifact.ChangeSet != nil && artifact.ChangeSet.GitPatch != nil {
			changes = append(changes, agent.Change{
				FilePath:    artifact.ChangeSet.Source, // Use changeset source as file indicator
				Type:        agent.ChangeTypeModify,
				Patch:       artifact.ChangeSet.GitPatch.UnidiffPatch,
				Description: fmt.Sprintf("Changes from Jules activity %s: %s", activityID, artifact.ChangeSet.GitPatch.SuggestedCommitMessage),
			})
		}
	}

	return &ToolResult{
		Success: true,
		Output: map[string]interface{}{
			"session_id":  sessionID,
			"activity_id": activityID,
			"changes":     len(changes),
		},
		Changes: changes,
	}, nil
}

// getActivities retrieves activities from a Jules session
func (j *JulesTool) getActivities(ctx context.Context, params map[string]interface{}) (*ToolResult, error) {
	sessionID, ok := params["session_id"].(string)
	if !ok {
		return nil, fmt.Errorf("session_id parameter is required for get_activities")
	}

	limit := 10
	if l, ok := params["limit"].(int); ok {
		limit = l
	}

	activities, err := j.client.ListActivities(ctx, sessionID, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to list activities: %w", err)
	}

	return &ToolResult{
		Success: true,
		Output: map[string]interface{}{
			"session_id": sessionID,
			"activities": activities,
			"count":      len(activities),
		},
	}, nil
}
