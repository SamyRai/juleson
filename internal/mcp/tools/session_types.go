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
	TotalSessions      int             `json:"total_sessions"`
	StateBreakdown     map[string]int  `json:"state_breakdown"`
	ActiveSessions     int             `json:"active_sessions"`
	UserActionSessions int             `json:"user_action_sessions"`
	RecentSessions     []jules.Session `json:"recent_sessions"`
	Summary            string          `json:"summary"`
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
	SessionID         string `json:"session_id" jsonschema:"ID of the session to apply patches from"`
	WorkingDir        string `json:"working_dir,omitempty" jsonschema:"Working directory where patches should be applied (default: current directory)"`
	DryRun            bool   `json:"dry_run,omitempty" jsonschema:"Whether to perform a dry-run without actually applying changes"`
	ConfirmApply      bool   `json:"confirm_apply,omitempty" jsonschema:"Must be true to mutate the working directory; otherwise the tool dry-runs"`
	AllowDirty        bool   `json:"allow_dirty,omitempty" jsonschema:"Allow applying patches when the target worktree has local changes"`
	Force             bool   `json:"force,omitempty" jsonschema:"Whether to use git apply --3way if some hunks need merging (default: false)"`
	CreateBackup      bool   `json:"create_backup,omitempty" jsonschema:"Whether to create backup files before applying patches (default: false)"`
	ActivityID        string `json:"activity_id,omitempty" jsonschema:"Optional activity ID or resource name to apply from"`
	ArtifactIndex     *int   `json:"artifact_index,omitempty" jsonschema:"Optional artifact index to apply within the selected scope"`
	AllowBaseMismatch bool   `json:"allow_base_mismatch,omitempty" jsonschema:"Allow applying when gitPatch.baseCommitId differs from target HEAD"`
}

// ApplySessionPatchesOutput represents output for apply_session_patches tool
type ApplySessionPatchesOutput struct {
	SessionID               string   `json:"session_id"`
	PatchesApplied          int      `json:"patches_applied"`
	PatchesFailed           int      `json:"patches_failed"`
	FilesModified           []string `json:"files_modified"`
	SuggestedCommitMessages []string `json:"suggested_commit_messages,omitempty"`
	Warnings                []string `json:"warnings,omitempty"`
	BaseCommitMismatches    []string `json:"base_commit_mismatches,omitempty"`
	Blockers                []string `json:"blockers,omitempty"`
	Errors                  []string `json:"errors,omitempty"`
	DryRun                  bool     `json:"dry_run"`
	Message                 string   `json:"message"`
}

// PreviewSessionChangesInput represents input for preview_session_changes tool
type PreviewSessionChangesInput struct {
	SessionID     string `json:"session_id" jsonschema:"ID of the session to preview changes for"`
	WorkingDir    string `json:"working_dir,omitempty" jsonschema:"Working directory (default: current directory)"`
	ActivityID    string `json:"activity_id,omitempty" jsonschema:"Optional activity ID or resource name to preview"`
	ArtifactIndex *int   `json:"artifact_index,omitempty" jsonschema:"Optional artifact index to preview within the selected scope"`
}

// PreviewSessionChangesOutput represents output for preview_session_changes tool
type PreviewSessionChangesOutput struct {
	SessionID               string                `json:"session_id"`
	TotalPatches            int                   `json:"total_patches"`
	Files                   []julesops.FileChange `json:"files"`
	SuggestedCommitMessages []string              `json:"suggested_commit_messages,omitempty"`
	Warnings                []string              `json:"warnings,omitempty"`
	BaseCommitMismatches    []string              `json:"base_commit_mismatches,omitempty"`
	CanApply                bool                  `json:"can_apply"`
	Errors                  []string              `json:"errors,omitempty"`
	Summary                 string                `json:"summary"`
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
	Prompt              string `json:"prompt,omitempty" jsonschema:"Prompt describing the task for Jules to work on"`
	PromptFile          string `json:"prompt_file,omitempty" jsonschema:"Local file path to read the prompt from"`
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

// WatchSessionInput represents input for watch_session tool.
type WatchSessionInput struct {
	SessionID                 string `json:"session_id" jsonschema:"ID of the session to watch"`
	IntervalSeconds           int    `json:"interval_seconds,omitempty" jsonschema:"Polling interval in seconds (default: 30)"`
	TimeoutSeconds            int    `json:"timeout_seconds,omitempty" jsonschema:"Maximum watch duration in seconds (default: 1800)"`
	Since                     string `json:"since,omitempty" jsonschema:"Optional RFC3339 createTime activity cursor"`
	InitialState              string `json:"initial_state,omitempty" jsonschema:"Known current session state; used with return_on_status_change"`
	ReturnOnStatusChange      bool   `json:"return_on_status_change,omitempty" jsonschema:"Return when session state changes from initial_state or the first observed state"`
	ReturnOnJulesAgentMessage bool   `json:"return_on_jules_agent_message,omitempty" jsonschema:"Return when a new Jules-authored message activity appears after since or after the initial poll"`
}

// WatchSessionOutput represents output for watch_session tool.
type WatchSessionOutput struct {
	SessionID          string           `json:"session_id"`
	State              string           `json:"state"`
	NeedsUserAction    bool             `json:"needs_user_action"`
	IsTerminal         bool             `json:"is_terminal"`
	NextAction         string           `json:"next_action"`
	WakeReason         string           `json:"wake_reason,omitempty"`
	NextActivityCursor string           `json:"next_activity_cursor,omitempty"`
	Session            *jules.Session   `json:"session,omitempty"`
	RecentActivities   []jules.Activity `json:"recent_activities,omitempty"`
}

// VerifySessionChangesInput represents input for verify_session_changes tool.
type VerifySessionChangesInput struct {
	WorkingDir string   `json:"working_dir,omitempty" jsonschema:"Working directory to verify (default: current directory)"`
	Packages   []string `json:"packages,omitempty" jsonschema:"Go packages to test (default: ./...)"`
	Short      bool     `json:"short,omitempty" jsonschema:"Run short tests only"`
	Command    string   `json:"command,omitempty" jsonschema:"Explicit verification command to run instead of auto-detection"`
}

// VerifySessionChangesOutput represents output for verify_session_changes tool.
type VerifySessionChangesOutput struct {
	WorkingDir string `json:"working_dir"`
	Success    bool   `json:"success"`
	Command    string `json:"command"`
	Output     string `json:"output,omitempty"`
	Summary    string `json:"summary"`
}

type ListSessionArtifactsInput struct {
	SessionID string `json:"session_id" jsonschema:"ID of the session to inspect"`
}

type ListSessionArtifactsOutput struct {
	SessionID  string                      `json:"session_id"`
	Artifacts  []julesops.ArtifactManifest `json:"artifacts"`
	TotalCount int                         `json:"total_count"`
}

type GetSessionOutputsInput struct {
	SessionID string `json:"session_id" jsonschema:"ID of the session to inspect"`
}

type GetSessionOutputsOutput struct {
	SessionID  string         `json:"session_id"`
	Outputs    []jules.Output `json:"outputs"`
	TotalCount int            `json:"total_count"`
}
