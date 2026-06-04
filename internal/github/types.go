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
