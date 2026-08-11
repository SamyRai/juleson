package github

// Repository represents a GitHub repository with metadata.
type Repository struct {
	Owner         string `json:"owner"`
	Name          string `json:"name"`
	FullName      string `json:"full_name"`
	Description   string `json:"description,omitempty"`
	DefaultBranch string `json:"default_branch"`
	URL           string `json:"url"`
	UpdatedAt     string `json:"updated_at,omitempty"`
	Stars         int    `json:"stars"`
	Forks         int    `json:"forks"`
	HasIssues     bool   `json:"has_issues"`
	Private       bool   `json:"private"`
}
