package tools

// PlanProjectAutomationInput represents input for plan_project_automation tool
type PlanProjectAutomationInput struct {
	ProjectPath   string `json:"project_path" jsonschema:"Path to the project directory"`
	Goals         string `json:"goals" jsonschema:"Project goals and objectives"`
	Constraints   string `json:"constraints,omitempty" jsonschema:"Project constraints and limitations"`
	Timeline      string `json:"timeline,omitempty" jsonschema:"Project timeline and deadlines"`
	PriorityAreas string `json:"priority_areas,omitempty" jsonschema:"Areas to prioritize (e.g., testing, refactoring, documentation)"`
}

// PlanProjectAutomationOutput represents output for plan_project_automation tool
type PlanProjectAutomationOutput struct {
	AutomationPlan   string           `json:"automation_plan"`
	RecommendedSteps []AutomationStep `json:"recommended_steps"`
	EstimatedEffort  string           `json:"estimated_effort"`
	RiskAssessment   string           `json:"risk_assessment"`
	SuccessMetrics   []string         `json:"success_metrics"`
	ResourceNeeds    []string         `json:"resource_needs"`
}

// AutomationStep represents a single step in an automation plan
type AutomationStep struct {
	Name         string   `json:"name"`
	Description  string   `json:"description"`
	Tools        []string `json:"tools"`
	Duration     string   `json:"duration"`
	Dependencies []string `json:"dependencies"`
}

// OrchestrateWorkflowInput represents input for orchestrate_workflow tool
type OrchestrateWorkflowInput struct {
	ProjectPath     string            `json:"project_path" jsonschema:"Path to the project directory"`
	WorkflowSteps   []WorkflowStep    `json:"workflow_steps" jsonschema:"Steps to execute in the workflow"`
	Parameters      map[string]string `json:"parameters,omitempty" jsonschema:"Additional parameters for workflow execution"`
	ContinueOnError bool              `json:"continue_on_error,omitempty" jsonschema:"Whether to continue workflow if a step fails"`
}

// WorkflowStep represents a single step in a workflow
type WorkflowStep struct {
	Name       string            `json:"name"`
	Tool       string            `json:"tool"`
	Parameters map[string]string `json:"parameters"`
	Condition  string            `json:"condition,omitempty"`
}

// OrchestrateWorkflowOutput represents output for orchestrate_workflow tool
type OrchestrateWorkflowOutput struct {
	ExecutionResults []StepResult `json:"execution_results"`
	OverallStatus    string       `json:"overall_status"`
	Summary          string       `json:"summary"`
	NextSteps        []string     `json:"next_steps"`
	Issues           []string     `json:"issues"`
}

// StepResult represents the result of executing a workflow step
type StepResult struct {
	StepName string `json:"step_name"`
	Status   string `json:"status"`
	Output   string `json:"output"`
	Duration string `json:"duration"`
	Error    string `json:"error,omitempty"`
}

// ManageGitHubProjectInput represents input for manage_github_project tool
type ManageGitHubProjectInput struct {
	Action      string            `json:"action" jsonschema:"Action to perform (create_project, create_issue, create_milestone, update_status, etc.)"`
	ProjectName string            `json:"project_name,omitempty" jsonschema:"Name of the GitHub project"`
	Repository  string            `json:"repository,omitempty" jsonschema:"GitHub repository (owner/repo format)"`
	Parameters  map[string]string `json:"parameters,omitempty" jsonschema:"Additional parameters for the action"`
	Description string            `json:"description,omitempty" jsonschema:"Description for created items"`
}

// ManageGitHubProjectOutput represents output for manage_github_project tool
type ManageGitHubProjectOutput struct {
	Action      string      `json:"action"`
	Status      string      `json:"status"`
	Result      interface{} `json:"result"`
	Message     string      `json:"message"`
	NextActions []string    `json:"next_actions,omitempty"`
}

// SynthesizeSessionResultsInput represents input for synthesize_session_results tool
type SynthesizeSessionResultsInput struct {
	SessionID   string `json:"session_id" jsonschema:"Jules session ID to analyze"`
	IncludeLogs bool   `json:"include_logs,omitempty" jsonschema:"Whether to include detailed logs in analysis"`
	FocusAreas  string `json:"focus_areas,omitempty" jsonschema:"Specific areas to focus analysis on"`
}

// SynthesizeSessionResultsOutput represents output for synthesize_session_results tool
type SynthesizeSessionResultsOutput struct {
	SessionSummary   string            `json:"session_summary"`
	KeyInsights      []string          `json:"key_insights"`
	IssuesFound      []IssueSummary    `json:"issues_found"`
	ImprovementsMade []Improvement     `json:"improvements_made"`
	Recommendations  []string          `json:"recommendations"`
	SuccessMetrics   map[string]string `json:"success_metrics"`
	NextSteps        []string          `json:"next_steps"`
}

// IssueSummary represents a summary of an issue found during session
type IssueSummary struct {
	Type        string `json:"type"`
	Severity    string `json:"severity"`
	Description string `json:"description"`
	Location    string `json:"location,omitempty"`
	Suggestion  string `json:"suggestion"`
}

// Improvement represents an improvement made during session
type Improvement struct {
	Category    string `json:"category"`
	Description string `json:"description"`
	Impact      string `json:"impact"`
}
