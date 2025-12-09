package github

// Repository represents a GitHub repository with metadata
type Repository struct {
	Owner         string `json:"owner"`
	Name          string `json:"name"`
	FullName      string `json:"full_name"`
	Description   string `json:"description,omitempty"`
	Stars         int    `json:"stars"`
	Forks         int    `json:"forks"`
	OpenIssues    int    `json:"open_issues"`
	DefaultBranch string `json:"default_branch"`
	Private       bool   `json:"private"`
	URL           string `json:"url"`
	UpdatedAt     string `json:"updated_at,omitempty"`
}

// WorkflowRun represents a GitHub Actions workflow run
type WorkflowRun struct {
	ID           int64  `json:"id"`
	Name         string `json:"name"`
	HeadBranch   string `json:"head_branch"`
	Status       string `json:"status"`
	Conclusion   string `json:"conclusion,omitempty"`
	WorkflowID   int64  `json:"workflow_id"`
	URL          string `json:"url"`
	CreatedAt    string `json:"created_at"`
	UpdatedAt    string `json:"updated_at"`
	RunNumber    int    `json:"run_number"`
	Event        string `json:"event"`
	Actor        string `json:"actor"`
	RunAttempt   int    `json:"run_attempt"`
	RunStartedAt string `json:"run_started_at,omitempty"`
}

// Workflow represents a GitHub Actions workflow
type Workflow struct {
	ID        int64  `json:"id"`
	Name      string `json:"name"`
	Path      string `json:"path"`
	State     string `json:"state"`
	CreatedAt string `json:"created_at"`
	UpdatedAt string `json:"updated_at"`
	URL       string `json:"url"`
	BadgeURL  string `json:"badge_url"`
}

// WorkflowJob represents a GitHub Actions workflow job
type WorkflowJob struct {
	ID          int64  `json:"id"`
	RunID       int64  `json:"run_id"`
	Name        string `json:"name"`
	Status      string `json:"status"`
	Conclusion  string `json:"conclusion,omitempty"`
	StartedAt   string `json:"started_at,omitempty"`
	CompletedAt string `json:"completed_at,omitempty"`
	URL         string `json:"url"`
	RunnerName  string `json:"runner_name,omitempty"`
}
