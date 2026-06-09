package templates

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

// Manager manages templates and the registry.
type Manager struct {
	registry     *Registry
	customPath   string
	stores       []Store
	enableCustom bool
}

// NewManager creates a new template manager.
func NewManager(templatesDir string, customPath string, enableCustom bool) (*Manager, error) {
	manager := &Manager{
		customPath:   customPath,
		enableCustom: enableCustom,
		registry:     &Registry{},
	}

	manager.stores = append(manager.stores, NewEmbeddedStore(templatesDir))
	if enableCustom && customPath != "" {
		manager.stores = append(manager.stores, NewCustomStore(customPath))
	}

	for _, store := range manager.stores {
		reg, err := store.LoadRegistry()
		if err != nil {
			if _, ok := store.(*EmbeddedStore); ok {
				return nil, fmt.Errorf("failed to load registry: %w", err)
			}
			fmt.Printf("Warning: failed to load custom templates: %v\n", err)
			continue
		}
		manager.registry.Templates = append(manager.registry.Templates, reg.Templates...)
	}

	return manager, nil
}

// LoadTemplate loads a template by name.
func (m *Manager) LoadTemplate(name string) (*Template, error) {
	if name == "" {
		return nil, fmt.Errorf("template name cannot be empty")
	}

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

	var template *Template
	var err error

	for _, store := range m.stores {
		if _, ok := store.(*EmbeddedStore); ok && strings.HasPrefix(registryTemplate.File, "builtin/") {
			template, err = store.LoadTemplate(registryTemplate.File)
			break
		}
		if _, ok := store.(*CustomStore); ok && !strings.HasPrefix(registryTemplate.File, "builtin/") {
			template, err = store.LoadTemplate(registryTemplate.File)
			break
		}
	}

	if err != nil {
		return nil, fmt.Errorf("failed to load template '%s': %w", name, err)
	}
	if template == nil {
		return nil, fmt.Errorf("store not found for template file %s", registryTemplate.File)
	}

	if err := m.ValidateTemplate(template); err != nil {
		return nil, fmt.Errorf("template '%s' validation failed: %w", name, err)
	}

	return template, nil
}

// ListTemplates returns all available templates.
func (m *Manager) ListTemplates() []RegistryTemplate {
	return m.registry.Templates
}

// ListTemplatesByCategory returns templates filtered by category.
func (m *Manager) ListTemplatesByCategory(category string) []RegistryTemplate {
	var filtered []RegistryTemplate
	for _, t := range m.registry.Templates {
		if t.Category == category {
			filtered = append(filtered, t)
		}
	}
	return filtered
}

// SearchTemplates searches templates by query.
func (m *Manager) SearchTemplates(query string) []RegistryTemplate {
	var results []RegistryTemplate
	queryLower := strings.ToLower(query)

	for _, t := range m.registry.Templates {
		if strings.Contains(strings.ToLower(t.Name), queryLower) ||
			strings.Contains(strings.ToLower(t.Description), queryLower) {
			results = append(results, t)
			continue
		}
		for _, tag := range t.Tags {
			if strings.Contains(strings.ToLower(tag), queryLower) {
				results = append(results, t)
				break
			}
		}
	}

	return results
}

// ValidateTemplate validates a template.
func (m *Manager) ValidateTemplate(template *Template) error {
	if template.Metadata.Name == "" {
		return fmt.Errorf("template name is required")
	}
	if template.Metadata.Version == "" {
		return fmt.Errorf("template version is required")
	}
	if template.Metadata.Category == "" {
		return fmt.Errorf("template category is required")
	}
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

// CreateTemplate creates a new template.
func (m *Manager) CreateTemplate(name, category, description string) (*Template, error) {
	if _, err := m.LoadTemplate(name); err == nil {
		return nil, fmt.Errorf("template '%s' already exists", name)
	}

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

// SaveTemplate saves a template to disk (for custom templates only).
func (m *Manager) SaveTemplate(template *Template) error {
	if !m.enableCustom || m.customPath == "" {
		return fmt.Errorf("custom templates are not enabled or path not configured")
	}

	if err := m.ValidateTemplate(template); err != nil {
		return fmt.Errorf("template validation failed: %w", err)
	}

	categoryDir := filepath.Join(m.customPath, template.Metadata.Category)
	if err := os.MkdirAll(categoryDir, 0755); err != nil {
		return fmt.Errorf("failed to create category directory: %w", err)
	}

	fileName := fmt.Sprintf("%s.yaml", template.Metadata.Name)
	filePath := filepath.Join(categoryDir, fileName)

	data, err := yaml.Marshal(template)
	if err != nil {
		return fmt.Errorf("failed to marshal template: %w", err)
	}

	if err := os.WriteFile(filePath, data, 0600); err != nil {
		return fmt.Errorf("failed to write template file: %w", err)
	}

	return nil
}
