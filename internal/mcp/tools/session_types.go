package tools

import (
	"github.com/SamyRai/juleson/internal/julesops"
	"github.com/SamyRai/juleson/pkg/jules"
)

// ListSessionsInput represents input for list_sessions tool
type ListSessionsInput struct {
	Limit  int    `json:"limit,omitempty" jsonschema:"Maximum number of sessions to return (default: 50)"`
	Cursor string `json:"cursor,omitempty" jsonschema:"Cursor for pagination"`
}

// ListSessionsOutput represents output for list_sessions tool
type ListSessionsOutput struct {
	Sessions   []jules.Session `json:"sessions"`
	NextCursor string          `json:"next_cursor,omitempty"`
	TotalCount int             `json:"total_count"`
}

// GetSessionStatusInput represents input for get_session_status tool
type GetSessionStatusInput struct {
	Limit int `json:"limit,omitempty" jsonschema:"Maximum number of sessions to analyze (default: 100)"`
}

// GetSessionStatusOutput represents output for get_session_status tool
type GetSessionStatusOutput struct {
	TotalSessions  int             `json:"total_sessions"`
	StateBreakdown map[string]int  `json:"state_breakdown"`
	ActiveSessions int             `json:"active_sessions"`
	RecentSessions []jules.Session `json:"recent_sessions"`
	Summary        string          `json:"summary"`
}

// ApproveSessionPlanInput represents input for approve_session_plan tool
type ApproveSessionPlanInput struct {
	SessionID string `json:"session_id" jsonschema:"ID of the session to approve"`
}

// ApproveSessionPlanOutput represents output for approve_session_plan tool
type ApproveSessionPlanOutput struct {
	SessionID string `json:"session_id"`
	Status    string `json:"status"`
	Message   string `json:"message"`
}

// DeleteSessionInput represents input for delete_session tool
type DeleteSessionInput struct {
	SessionID string `json:"session_id" jsonschema:"ID of the session to delete"`
	Confirm   bool   `json:"confirm" jsonschema:"Must be true to confirm destructive deletion"`
}

// DeleteSessionOutput represents output for delete_session tool
type DeleteSessionOutput struct {
	SessionID string `json:"session_id"`
	Status    string `json:"status"`
	Message   string `json:"message"`
}

// ApplySessionPatchesInput represents input for apply_session_patches tool
type ApplySessionPatchesInput struct {
	SessionID    string `json:"session_id" jsonschema:"ID of the session to apply patches from"`
	WorkingDir   string `json:"working_dir,omitempty" jsonschema:"Working directory where patches should be applied (default: current directory)"`
	DryRun       bool   `json:"dry_run,omitempty" jsonschema:"Whether to perform a dry-run without actually applying changes (default: false)"`
	Force        bool   `json:"force,omitempty" jsonschema:"Whether to force application even if some hunks fail (default: false)"`
	CreateBackup bool   `json:"create_backup,omitempty" jsonschema:"Whether to create backup files before applying patches (default: false)"`
}

// ApplySessionPatchesOutput represents output for apply_session_patches tool
type ApplySessionPatchesOutput struct {
	SessionID      string   `json:"session_id"`
	PatchesApplied int      `json:"patches_applied"`
	PatchesFailed  int      `json:"patches_failed"`
	FilesModified  []string `json:"files_modified"`
	Errors         []string `json:"errors,omitempty"`
	DryRun         bool     `json:"dry_run"`
	Message        string   `json:"message"`
}

// PreviewSessionChangesInput represents input for preview_session_changes tool
type PreviewSessionChangesInput struct {
	SessionID  string `json:"session_id" jsonschema:"ID of the session to preview changes for"`
	WorkingDir string `json:"working_dir,omitempty" jsonschema:"Working directory (default: current directory)"`
}

// PreviewSessionChangesOutput represents output for preview_session_changes tool
type PreviewSessionChangesOutput struct {
	SessionID    string                `json:"session_id"`
	TotalPatches int                   `json:"total_patches"`
	Files        []julesops.FileChange `json:"files"`
	CanApply     bool                  `json:"can_apply"`
	Errors       []string              `json:"errors,omitempty"`
	Summary      string                `json:"summary"`
}

// SendSessionMessageInput represents input for send_session_message tool
type SendSessionMessageInput struct {
	SessionID string `json:"session_id" jsonschema:"ID of the session to send message to"`
	Message   string `json:"message" jsonschema:"Message to send to Jules within the session"`
}

// SendSessionMessageOutput represents output for send_session_message tool
type SendSessionMessageOutput struct {
	SessionID string `json:"session_id"`
	Status    string `json:"status"`
	Message   string `json:"message"`
}

// CreateSessionInput represents input for create_session tool
type CreateSessionInput struct {
	Source              string `json:"source,omitempty" jsonschema:"Optional source ID or path (e.g., 'sources/github/owner/repo'); omit for a repoless session"`
	Prompt              string `json:"prompt" jsonschema:"Prompt describing the task for Jules to work on"`
	Title               string `json:"title,omitempty" jsonschema:"Optional title for the session"`
	RequirePlanApproval bool   `json:"require_plan_approval,omitempty" jsonschema:"Whether to require manual approval of plans (default: false)"`
	AutomationMode      string `json:"automation_mode,omitempty" jsonschema:"Automation mode (e.g., 'AUTO_CREATE_PR')"`
	StartingBranch      string `json:"starting_branch,omitempty" jsonschema:"Starting branch for GitHub repos (default: repo's default branch)"`
}

// CreateSessionOutput represents output for create_session tool
type CreateSessionOutput struct {
	SessionID string        `json:"session_id"`
	Session   jules.Session `json:"session"`
	URL       string        `json:"url"`
	Message   string        `json:"message"`
}

// GetSessionInput represents input for get_session tool
type GetSessionInput struct {
	SessionID string `json:"session_id" jsonschema:"ID of the session to retrieve"`
}

// GetSessionOutput represents output for get_session tool
type GetSessionOutput struct {
	SessionID string        `json:"session_id"`
	Session   jules.Session `json:"session"`
	URL       string        `json:"url"`
}
