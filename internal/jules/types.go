package jules

// ============================================================================
// Session Types
// ============================================================================

// Session represents a Jules coding session
type Session struct {
	Name                string         `json:"name"`
	Title               string         `json:"title"`
	State               string         `json:"state"` // PLANNING, IN_PROGRESS, COMPLETED, FAILED
	CreateTime          string         `json:"createTime"`
	UpdateTime          string         `json:"updateTime"`
	SourceContext       *SourceContext `json:"sourceContext,omitempty"`
	Prompt              string         `json:"prompt"`
	URL                 string         `json:"url"`
	ID                  string         `json:"id"`
	RequirePlanApproval bool           `json:"requirePlanApproval,omitempty"`
	AutomationMode      string         `json:"automationMode,omitempty"` // AUTO_CREATE_PR
	Outputs             []Output       `json:"outputs,omitempty"`
}

// Output represents session outputs (e.g., PRs created)
type Output struct {
	PullRequest *PullRequest `json:"pullRequest,omitempty"`
}

// PullRequest represents a GitHub PR created by Jules
type PullRequest struct {
	URL         string `json:"url"`
	Title       string `json:"title"`
	Description string `json:"description"`
}

// ============================================================================
// Source Types
// ============================================================================

// SourceContext represents the source context
type SourceContext struct {
	Source            string             `json:"source"`
	GithubRepoContext *GithubRepoContext `json:"githubRepoContext,omitempty"`
}

// GithubRepoContext represents GitHub repo context
type GithubRepoContext struct {
	StartingBranch string `json:"startingBranch"`
}

// Source represents a code source (e.g., GitHub repository)
type Source struct {
	Name       string      `json:"name"`
	ID         string      `json:"id"`
	GithubRepo *GithubRepo `json:"githubRepo,omitempty"`
}

// GithubRepo represents GitHub repository information
type GithubRepo struct {
	Owner         string   `json:"owner"`
	Repo          string   `json:"repo"`
	DefaultBranch *Branch  `json:"defaultBranch,omitempty"`
	Branches      []Branch `json:"branches,omitempty"`
}

// Branch represents a Git branch
type Branch struct {
	DisplayName string `json:"displayName"`
}

// ============================================================================
// Activity Types
// ============================================================================

// Activity represents an activity within a session
type Activity struct {
	Name             string            `json:"name"`
	CreateTime       string            `json:"createTime"`
	Originator       string            `json:"originator"` // "agent" or "user"
	PlanGenerated    *PlanGenerated    `json:"planGenerated,omitempty"`
	PlanApproved     *PlanApproved     `json:"planApproved,omitempty"`
	ProgressUpdated  *ProgressUpdated  `json:"progressUpdated,omitempty"`
	SessionCompleted *SessionCompleted `json:"sessionCompleted,omitempty"`
	SessionFailed    *SessionFailed    `json:"sessionFailed,omitempty"`
	Artifacts        []Artifact        `json:"artifacts,omitempty"`
	ID               string            `json:"id"`
}

// PlanGenerated represents a generated plan activity
type PlanGenerated struct {
	Plan Plan `json:"plan"`
}

// Plan represents a coding plan
type Plan struct {
	ID    string `json:"id"`
	Steps []Step `json:"steps"`
}

// Step represents a step in the plan
type Step struct {
	ID          string `json:"id"`
	Title       string `json:"title"`
	Description string `json:"description,omitempty"`
	Index       int    `json:"index,omitempty"`
}

// PlanApproved represents a plan approval activity
type PlanApproved struct {
	PlanID string `json:"planId"`
}

// ProgressUpdated represents a progress update activity
type ProgressUpdated struct {
	Title       string `json:"title"`
	Description string `json:"description,omitempty"`
}

// SessionCompleted represents session completion activity
type SessionCompleted struct{}

// SessionFailed represents session failure activity
type SessionFailed struct {
	Reason string `json:"reason"`
}

// ============================================================================
// Artifact Types
// ============================================================================

// Artifact represents an artifact produced during an activity
type Artifact struct {
	BashOutput *BashOutput `json:"bashOutput,omitempty"`
	ChangeSet  *ChangeSet  `json:"changeSet,omitempty"`
	Media      *Media      `json:"media,omitempty"`
}

// BashOutput represents bash command output
type BashOutput struct {
	Command  string `json:"command"`
	Output   string `json:"output"`
	ExitCode int    `json:"exitCode,omitempty"`
}

// ChangeSet represents code changes
type ChangeSet struct {
	Source   string    `json:"source"`
	GitPatch *GitPatch `json:"gitPatch,omitempty"`
}

// GitPatch represents a git patch
type GitPatch struct {
	UnidiffPatch           string `json:"unidiffPatch"`
	BaseCommitID           string `json:"baseCommitId,omitempty"`
	SuggestedCommitMessage string `json:"suggestedCommitMessage,omitempty"`
}

// Media represents media artifacts (e.g., screenshots)
type Media struct {
	Data     string `json:"data"`
	MimeType string `json:"mimeType"`
}

// ============================================================================
// Request Types
// ============================================================================

// CreateSessionRequest represents a request to create a session
type CreateSessionRequest struct {
	Prompt              string         `json:"prompt"`
	SourceContext       *SourceContext `json:"sourceContext"`
	Title               string         `json:"title,omitempty"`
	AutomationMode      string         `json:"automationMode,omitempty"`
	RequirePlanApproval bool           `json:"requirePlanApproval,omitempty"`
}

// SendMessageRequest represents a request to send a message
type SendMessageRequest struct {
	Prompt string `json:"prompt"`
}

// ============================================================================
// Response Types
// ============================================================================

// SessionsResponse represents the response from listing sessions
type SessionsResponse struct {
	Sessions      []Session `json:"sessions"`
	NextPageToken string    `json:"nextPageToken,omitempty"`
}

// ActivitiesResponse represents the response from listing activities
type ActivitiesResponse struct {
	Activities    []Activity `json:"activities"`
	NextPageToken string     `json:"nextPageToken,omitempty"`
}

// SourcesResponse represents the response from listing sources
type SourcesResponse struct {
	Sources       []Source `json:"sources"`
	NextPageToken string   `json:"nextPageToken,omitempty"`
}
