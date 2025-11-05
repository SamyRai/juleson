package templates

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	yaml "gopkg.in/yaml.v3"
)

// Template represents a Jules automation template
type Template struct {
	Metadata   TemplateMetadata   `yaml:"metadata"`
	Config     TemplateConfig     `yaml:"config"`
	Context    TemplateContext    `yaml:"context"`
	Tasks      []TemplateTask     `yaml:"tasks"`
	Validation TemplateValidation `yaml:"validation"`
	Output     TemplateOutput     `yaml:"output"`
}

// TemplateMetadata contains template metadata
type TemplateMetadata struct {
	Name        string   `yaml:"name"`
	Version     string   `yaml:"version"`
	Description string   `yaml:"description"`
	Author      string   `yaml:"author"`
	Tags        []string `yaml:"tags"`
	Category    string   `yaml:"category"`
}

// TemplateConfig contains template configuration
type TemplateConfig struct {
	Strategy           string `yaml:"strategy"`
	MaxConcurrentTasks int    `yaml:"max_concurrent_tasks"`
	Timeout            string `yaml:"timeout"`
	RequiresApproval   bool   `yaml:"requires_approval"`
	BackupEnabled      bool   `yaml:"backup_enabled"`
}

// TemplateContext contains context extraction rules
type TemplateContext struct {
	ProjectAnalysis []string             `yaml:"project_analysis"`
	FilePatterns    TemplateFilePatterns `yaml:"file_patterns"`
}

// TemplateFilePatterns contains file inclusion/exclusion patterns
type TemplateFilePatterns struct {
	Include []string `yaml:"include"`
	Exclude []string `yaml:"exclude"`
}

// TemplateTask represents a task within a template
type TemplateTask struct {
	Name             string            `yaml:"name"`
	Type             string            `yaml:"type"`
	Description      string            `yaml:"description"`
	DependsOn        []string          `yaml:"depends_on"`
	RequiresApproval bool              `yaml:"requires_approval"`
	JulesPrompt      string            `yaml:"jules_prompt"`
	ContextVars      map[string]string `yaml:"context_vars"`
}

// TemplateValidation contains validation rules
type TemplateValidation struct {
	PreExecution  []string `yaml:"pre_execution"`
	PostExecution []string `yaml:"post_execution"`
}

// TemplateOutput contains output configuration
type TemplateOutput struct {
	Format  string               `yaml:"format"`
	Include []string             `yaml:"include"`
	Files   []TemplateOutputFile `yaml:"files"`
}

// TemplateOutputFile represents an output file configuration
type TemplateOutputFile struct {
	Path     string `yaml:"path"`
	Template string `yaml:"template"`
}

// Registry represents the template registry
type Registry struct {
	Templates  []RegistryTemplate          `yaml:"templates"`
	Categories map[string]RegistryCategory `yaml:"categories"`
	Registry   RegistryMetadata            `yaml:"registry"`
}

// RegistryTemplate represents a template in the registry
type RegistryTemplate struct {
	Name              string                `yaml:"name"`
	Version           string                `yaml:"version"`
	Category          string                `yaml:"category"`
	Description       string                `yaml:"description"`
	Author            string                `yaml:"author"`
	Tags              []string              `yaml:"tags"`
	File              string                `yaml:"file"`
	Dependencies      []string              `yaml:"dependencies"`
	Compatibility     RegistryCompatibility `yaml:"compatibility"`
	Features          []string              `yaml:"features"`
	Complexity        string                `yaml:"complexity"`
	EstimatedDuration string                `yaml:"estimated_duration"`
}

// RegistryCompatibility contains compatibility information
type RegistryCompatibility struct {
	Languages  []string `yaml:"languages"`
	Frameworks []string `yaml:"frameworks"`
}

// RegistryCategory represents a template category
type RegistryCategory struct {
	Name        string `yaml:"name"`
	Description string `yaml:"description"`
	Icon        string `yaml:"icon"`
}

// RegistryMetadata contains registry metadata
type RegistryMetadata struct {
	Version         string `yaml:"version"`
	LastUpdated     string `yaml:"last_updated"`
	TotalTemplates  int    `yaml:"total_templates"`
	CategoriesCount int    `yaml:"categories_count"`
}

// Manager manages templates and the registry
type Manager struct {
	templatesDir string
	customPath   string
	enableCustom bool
	registry     *Registry
}

// NewManager creates a new template manager
func NewManager(templatesDir string, customPath string, enableCustom bool) (*Manager, error) {
	manager := &Manager{
		templatesDir: templatesDir,
		customPath:   customPath,
		enableCustom: enableCustom,
	}

	// Load registry from embedded files
	if err := manager.loadEmbeddedRegistry(); err != nil {
		return nil, fmt.Errorf("failed to load registry: %w", err)
	}

	// Load custom templates if enabled
	if enableCustom && customPath != "" {
		if err := manager.loadCustomTemplates(); err != nil {
			// Log warning but don't fail - custom templates are optional
			fmt.Printf("Warning: failed to load custom templates: %v\n", err)
		}
	}

	return manager, nil
} // LoadTemplate loads a template by name
func (m *Manager) LoadTemplate(name string) (*Template, error) {
	if name == "" {
		return nil, fmt.Errorf("template name cannot be empty")
	}

	// Find template in registry
	var registryTemplate *RegistryTemplate
	for _, t := range m.registry.Templates {
		if t.Name == name {
			registryTemplate = &t
			break
		}
	}

	if registryTemplate == nil {
		return nil, fmt.Errorf("template '%s' not found in registry", name)
	}

	// Check if it's an embedded template (relative path) or custom template (absolute path)
	var template *Template
	var err error

	if strings.HasPrefix(registryTemplate.File, "builtin/") {
		// Embedded template
		template, err = m.loadEmbeddedTemplate(registryTemplate.File)
	} else {
		// Custom template
		template, err = m.loadCustomTemplate(registryTemplate.File)
	}

	if err != nil {
		return nil, fmt.Errorf("failed to load template '%s': %w", name, err)
	}

	// Validate loaded template
	if err := m.ValidateTemplate(template); err != nil {
		return nil, fmt.Errorf("template '%s' validation failed: %w", name, err)
	}

	return template, nil
}

// ListTemplates returns all available templates
func (m *Manager) ListTemplates() []RegistryTemplate {
	return m.registry.Templates
}

// ListTemplatesByCategory returns templates filtered by category
func (m *Manager) ListTemplatesByCategory(category string) []RegistryTemplate {
	var filtered []RegistryTemplate
	for _, t := range m.registry.Templates {
		if t.Category == category {
			filtered = append(filtered, t)
		}
	}
	return filtered
}

// SearchTemplates searches templates by query
func (m *Manager) SearchTemplates(query string) []RegistryTemplate {
	var results []RegistryTemplate
	queryLower := strings.ToLower(query)

	for _, t := range m.registry.Templates {
		// Search in name, description, tags
		if strings.Contains(strings.ToLower(t.Name), queryLower) ||
			strings.Contains(strings.ToLower(t.Description), queryLower) {
			results = append(results, t)
			continue
		}

		// Search in tags
		for _, tag := range t.Tags {
			if strings.Contains(strings.ToLower(tag), queryLower) {
				results = append(results, t)
				break
			}
		}
	}

	return results
}

// ValidateTemplate validates a template
func (m *Manager) ValidateTemplate(template *Template) error {
	// Validate metadata
	if template.Metadata.Name == "" {
		return fmt.Errorf("template name is required")
	}

	if template.Metadata.Version == "" {
		return fmt.Errorf("template version is required")
	}

	if template.Metadata.Category == "" {
		return fmt.Errorf("template category is required")
	}

	// Validate tasks
	if len(template.Tasks) == 0 {
		return fmt.Errorf("template must have at least one task")
	}

	for i, task := range template.Tasks {
		if task.Name == "" {
			return fmt.Errorf("task %d: name is required", i)
		}

		if task.Type == "" {
			return fmt.Errorf("task %d: type is required", i)
		}

		if task.JulesPrompt == "" {
			return fmt.Errorf("task %d: jules_prompt is required", i)
		}
	}

	return nil
}

// CreateTemplate creates a new template
func (m *Manager) CreateTemplate(name, category, description string) (*Template, error) {
	// Check if template already exists
	if _, err := m.LoadTemplate(name); err == nil {
		return nil, fmt.Errorf("template '%s' already exists", name)
	}

	// Create new template
	template := &Template{
		Metadata: TemplateMetadata{
			Name:        name,
			Version:     "1.0.0",
			Description: description,
			Author:      "Juleson",
			Category:    category,
			Tags:        []string{},
		},
		Config: TemplateConfig{
			Strategy:           "default",
			MaxConcurrentTasks: 3,
			Timeout:            "300s",
			RequiresApproval:   false,
			BackupEnabled:      true,
		},
		Context: TemplateContext{
			ProjectAnalysis: []string{"analyze_project_structure"},
			FilePatterns: TemplateFilePatterns{
				Include: []string{"**/*"},
				Exclude: []string{"**/node_modules/**", "**/vendor/**", "**/.git/**"},
			},
		},
		Tasks: []TemplateTask{
			{
				Name:        "analyze_project",
				Type:        "analysis",
				Description: "Analyze project structure and requirements",
				JulesPrompt: "Analyze the project in {{.ProjectPath}} and provide recommendations for {{.Description}}",
				ContextVars: map[string]string{
					"ProjectPath": "{{.ProjectPath}}",
					"Description": "{{.Description}}",
				},
			},
		},
		Validation: TemplateValidation{
			PreExecution:  []string{"check_git_status"},
			PostExecution: []string{"run_tests"},
		},
		Output: TemplateOutput{
			Format:  "markdown",
			Include: []string{"summary", "recommendations"},
		},
	}

	return template, nil
}

// SaveTemplate saves a template to disk (for custom templates only)
func (m *Manager) SaveTemplate(template *Template) error {
	if !m.enableCustom || m.customPath == "" {
		return fmt.Errorf("custom templates are not enabled or path not configured")
	}

	// Validate template
	if err := m.ValidateTemplate(template); err != nil {
		return fmt.Errorf("template validation failed: %w", err)
	}

	// Determine file path in custom directory
	categoryDir := filepath.Join(m.customPath, template.Metadata.Category)
	if err := os.MkdirAll(categoryDir, 0755); err != nil {
		return fmt.Errorf("failed to create category directory: %w", err)
	}

	fileName := fmt.Sprintf("%s.yaml", template.Metadata.Name)
	filePath := filepath.Join(categoryDir, fileName)

	// Write template to file
	data, err := yaml.Marshal(template)
	if err != nil {
		return fmt.Errorf("failed to marshal template: %w", err)
	}

	if err := os.WriteFile(filePath, data, 0644); err != nil {
		return fmt.Errorf("failed to write template file: %w", err)
	}

	return nil
}

// loadEmbeddedRegistry loads the template registry from embedded files
func (m *Manager) loadEmbeddedRegistry() error {
	var registryPath string

	if m.templatesDir == "" {
		// If builtin path is empty, use default relative to project root
		registryPath = filepath.Join("templates", "registry", "registry.yaml")
	} else {
		// The registry is at templates/registry/registry.yaml
		// templatesDir is templates/builtin, so we need to go up one level
		registryPath = filepath.Join(filepath.Dir(m.templatesDir), "registry", "registry.yaml")
	}

	data, err := os.ReadFile(registryPath)
	if err != nil {
		return fmt.Errorf("failed to read registry file %s: %w", registryPath, err)
	}

	var registry Registry
	if err := yaml.Unmarshal(data, &registry); err != nil {
		return fmt.Errorf("failed to unmarshal registry: %w", err)
	}

	m.registry = &registry
	return nil
}

// loadEmbeddedTemplate loads a template from embedded files
func (m *Manager) loadEmbeddedTemplate(filePath string) (*Template, error) {
	// filePath is like "builtin/reorganization/modular-restructure.yaml"
	var templatePath string

	if m.templatesDir == "" {
		// If builtin path is empty, use default path
		templatePath = filepath.Join("templates", filePath)
	} else {
		// m.templatesDir is "./templates/builtin"
		// So, trim "builtin/" and join
		relativePath := strings.TrimPrefix(filePath, "builtin/")
		templatePath = filepath.Join(m.templatesDir, relativePath)
	}

	data, err := os.ReadFile(templatePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read template file %s: %w", templatePath, err)
	}

	var template Template
	if err := yaml.Unmarshal(data, &template); err != nil {
		return nil, fmt.Errorf("failed to unmarshal template: %w", err)
	}

	return &template, nil
}

// loadCustomTemplates loads custom templates from the filesystem
func (m *Manager) loadCustomTemplates() error {
	if m.customPath == "" {
		return nil
	}

	// Check if custom path exists
	if _, err := os.Stat(m.customPath); os.IsNotExist(err) {
		return fmt.Errorf("custom templates path does not exist: %s", m.customPath)
	}

	// Find all YAML files in custom directory
	return filepath.WalkDir(m.customPath, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if d.IsDir() || !strings.HasSuffix(strings.ToLower(path), ".yaml") {
			return nil
		}

		// Load custom template
		template, err := m.loadCustomTemplate(path)
		if err != nil {
			return fmt.Errorf("failed to load custom template %s: %w", path, err)
		}

		// Add to registry
		registryTemplate := RegistryTemplate{
			Name:         template.Metadata.Name,
			Version:      template.Metadata.Version,
			Category:     template.Metadata.Category,
			Description:  template.Metadata.Description,
			Author:       template.Metadata.Author,
			Tags:         template.Metadata.Tags,
			File:         path, // Store full path for custom templates
			Dependencies: []string{},
			Compatibility: RegistryCompatibility{
				Languages:  []string{"all"},
				Frameworks: []string{"all"},
			},
			Features:          []string{"custom"},
			Complexity:        "custom",
			EstimatedDuration: "custom",
		}

		m.registry.Templates = append(m.registry.Templates, registryTemplate)
		return nil
	})
}

// loadCustomTemplate loads a custom template from filesystem
func (m *Manager) loadCustomTemplate(filePath string) (*Template, error) {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read custom template file: %w", err)
	}

	var template Template
	if err := yaml.Unmarshal(data, &template); err != nil {
		return nil, fmt.Errorf("failed to unmarshal custom template: %w", err)
	}

	return &template, nil
}
