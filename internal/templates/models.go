package templates

// Template represents a Jules automation template.
type Template struct {
	Metadata   TemplateMetadata   `yaml:"metadata"`
	Config     TemplateConfig     `yaml:"config"`
	Context    TemplateContext    `yaml:"context"`
	Tasks      []TemplateTask     `yaml:"tasks"`
	Validation TemplateValidation `yaml:"validation"`
	Output     TemplateOutput     `yaml:"output"`
}

// TemplateMetadata contains template metadata.
type TemplateMetadata struct {
	Name        string   `yaml:"name"`
	Version     string   `yaml:"version"`
	Description string   `yaml:"description"`
	Author      string   `yaml:"author"`
	Category    string   `yaml:"category"`
	Tags        []string `yaml:"tags"`
}

// TemplateConfig contains template configuration.
type TemplateConfig struct {
	Strategy           string `yaml:"strategy"`
	Timeout            string `yaml:"timeout"`
	MaxConcurrentTasks int    `yaml:"max_concurrent_tasks"`
	RequiresApproval   bool   `yaml:"requires_approval"`
	BackupEnabled      bool   `yaml:"backup_enabled"`
}

// TemplateContext contains context extraction rules.
type TemplateContext struct {
	ProjectAnalysis []string             `yaml:"project_analysis"`
	FilePatterns    TemplateFilePatterns `yaml:"file_patterns"`
}

// TemplateFilePatterns contains file inclusion/exclusion patterns.
type TemplateFilePatterns struct {
	Include []string `yaml:"include"`
	Exclude []string `yaml:"exclude"`
}

// TemplateTask represents a task within a template.
type TemplateTask struct {
	ContextVars      map[string]string `yaml:"context_vars"`
	Name             string            `yaml:"name"`
	Type             string            `yaml:"type"`
	Description      string            `yaml:"description"`
	JulesPrompt      string            `yaml:"jules_prompt"`
	DependsOn        []string          `yaml:"depends_on"`
	RequiresApproval bool              `yaml:"requires_approval"`
}

// TemplateValidation contains validation rules.
type TemplateValidation struct {
	PreExecution  []string `yaml:"pre_execution"`
	PostExecution []string `yaml:"post_execution"`
}

// TemplateOutput contains output configuration.
type TemplateOutput struct {
	Format  string               `yaml:"format"`
	Include []string             `yaml:"include"`
	Files   []TemplateOutputFile `yaml:"files"`
}

// TemplateOutputFile represents an output file configuration.
type TemplateOutputFile struct {
	Path     string `yaml:"path"`
	Template string `yaml:"template"`
}

// Registry represents the template registry.
type Registry struct {
	Templates  []RegistryTemplate          `yaml:"templates"`
	Categories map[string]RegistryCategory `yaml:"categories"`
	Registry   RegistryMetadata            `yaml:"registry"`
}

// RegistryTemplate represents a template in the registry.
type RegistryTemplate struct {
	Name              string                `yaml:"name"`
	Version           string                `yaml:"version"`
	Category          string                `yaml:"category"`
	Description       string                `yaml:"description"`
	Author            string                `yaml:"author"`
	File              string                `yaml:"file"`
	Complexity        string                `yaml:"complexity"`
	EstimatedDuration string                `yaml:"estimated_duration"`
	Compatibility     RegistryCompatibility `yaml:"compatibility"`
	Tags              []string              `yaml:"tags"`
	Dependencies      []string              `yaml:"dependencies"`
	Features          []string              `yaml:"features"`
}

// RegistryCompatibility contains compatibility information.
type RegistryCompatibility struct {
	Languages  []string `yaml:"languages"`
	Frameworks []string `yaml:"frameworks"`
}

// RegistryCategory represents a template category.
type RegistryCategory struct {
	Name        string `yaml:"name"`
	Description string `yaml:"description"`
	Icon        string `yaml:"icon"`
}

// RegistryMetadata contains registry metadata.
type RegistryMetadata struct {
	Version         string `yaml:"version"`
	LastUpdated     string `yaml:"last_updated"`
	TotalTemplates  int    `yaml:"total_templates"`
	CategoriesCount int    `yaml:"categories_count"`
}
