package presentation

import (
	"fmt"
	"strings"

	"github.com/SamyRai/juleson/internal/templates"
)

// TemplateFormatter formats template information
type TemplateFormatter struct{}

// NewTemplateFormatter creates a new template formatter
func NewTemplateFormatter() *TemplateFormatter {
	return &TemplateFormatter{}
}

// FormatList displays a list of templates
func (f *TemplateFormatter) FormatList(templates []templates.RegistryTemplate) string {
	var sb strings.Builder

	sb.WriteString("📋 Available Templates\n")
	sb.WriteString("=====================\n\n")

	for _, template := range templates {
		sb.WriteString(fmt.Sprintf("• %s (%s) - %s\n", template.Name, template.Category, template.Description))
		sb.WriteString(fmt.Sprintf("  Tags: %s\n", strings.Join(template.Tags, ", ")))
		sb.WriteString(fmt.Sprintf("  Complexity: %s | Duration: %s\n", template.Complexity, template.EstimatedDuration))
		sb.WriteString("\n")
	}

	return sb.String()
}

// FormatDetails displays detailed template information
func (f *TemplateFormatter) FormatDetails(template *templates.Template) string {
	var sb strings.Builder

	sb.WriteString(fmt.Sprintf("📄 Template Details: %s\n", template.Metadata.Name))
	sb.WriteString("========================\n")
	sb.WriteString(fmt.Sprintf("Version: %s\n", template.Metadata.Version))
	sb.WriteString(fmt.Sprintf("Category: %s\n", template.Metadata.Category))
	sb.WriteString(fmt.Sprintf("Description: %s\n", template.Metadata.Description))
	sb.WriteString(fmt.Sprintf("Author: %s\n", template.Metadata.Author))
	sb.WriteString(fmt.Sprintf("Tags: %s\n", strings.Join(template.Metadata.Tags, ", ")))
	sb.WriteString(fmt.Sprintf("Strategy: %s\n", template.Config.Strategy))
	sb.WriteString(fmt.Sprintf("Max Concurrent Tasks: %d\n", template.Config.MaxConcurrentTasks))
	sb.WriteString(fmt.Sprintf("Timeout: %s\n", template.Config.Timeout))
	sb.WriteString(fmt.Sprintf("Requires Approval: %t\n", template.Config.RequiresApproval))
	sb.WriteString(fmt.Sprintf("Backup Enabled: %t\n", template.Config.BackupEnabled))
	sb.WriteString(fmt.Sprintf("Tasks: %d\n", len(template.Tasks)))

	sb.WriteString("\nTasks:\n")
	for i, task := range template.Tasks {
		sb.WriteString(fmt.Sprintf("  %d. %s (%s)\n", i+1, task.Name, task.Type))
		sb.WriteString(fmt.Sprintf("     %s\n", task.Description))
		if len(task.DependsOn) > 0 {
			sb.WriteString(fmt.Sprintf("     Depends on: %s\n", strings.Join(task.DependsOn, ", ")))
		}
	}

	return sb.String()
}
