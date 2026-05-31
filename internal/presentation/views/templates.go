package views

import (
	"fmt"
	"strings"

	"github.com/SamyRai/juleson/internal/presentation/views/theme"
	"github.com/SamyRai/juleson/internal/templates"
	"github.com/charmbracelet/lipgloss"
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

	sb.WriteString(theme.HeaderStyle.Render("📋 Available Templates") + "\n\n")

	if len(templates) == 0 {
		sb.WriteString(theme.MutedStyle.Render("📭 No templates found.") + "\n")
		return sb.String()
	}

	for _, template := range templates {
		nameStyle := lipgloss.NewStyle().Foreground(theme.InfoColor).Bold(true)
		categoryStyle := theme.MutedStyle

		fmt.Fprintf(&sb, "• %s %s - %s\n", nameStyle.Render(template.Name), categoryStyle.Render("("+template.Category+")"), template.Description)
		if len(template.Tags) > 0 {
			fmt.Fprintf(&sb, "  %s %s\n", theme.MutedStyle.Render("Tags:"), strings.Join(template.Tags, ", "))
		}

		complexityColor := theme.SuccessColor
		if template.Complexity == "High" {
			complexityColor = theme.ErrorColor
		} else if template.Complexity == "Medium" {
			complexityColor = theme.WarnColor
		}
		complexityStyle := lipgloss.NewStyle().Foreground(complexityColor)

		fmt.Fprintf(&sb, "  %s %s | %s %s\n",
			theme.MutedStyle.Render("Complexity:"),
			complexityStyle.Render(template.Complexity),
			theme.MutedStyle.Render("Duration:"),
			template.EstimatedDuration)
		sb.WriteString("\n")
	}

	return sb.String()
}

// FormatDetails displays detailed template information
func (f *TemplateFormatter) FormatDetails(template *templates.Template) string {
	var sb strings.Builder

	title := fmt.Sprintf("📄 Template Details: %s", template.Metadata.Name)
	sb.WriteString(theme.HeaderStyle.Render(title) + "\n\n")

	labelStyle := theme.MutedStyle
	valStyle := theme.InfoStyle

	fmt.Fprintf(&sb, "%s %s\n", labelStyle.Render("Version:"), valStyle.Render(template.Metadata.Version))
	fmt.Fprintf(&sb, "%s %s\n", labelStyle.Render("Category:"), valStyle.Render(template.Metadata.Category))
	fmt.Fprintf(&sb, "%s %s\n", labelStyle.Render("Description:"), template.Metadata.Description)
	fmt.Fprintf(&sb, "%s %s\n", labelStyle.Render("Author:"), template.Metadata.Author)
	if len(template.Metadata.Tags) > 0 {
		fmt.Fprintf(&sb, "%s %s\n", labelStyle.Render("Tags:"), strings.Join(template.Metadata.Tags, ", "))
	}

	sb.WriteString("\n" + theme.StepStyle.Render("Configuration") + "\n")
	fmt.Fprintf(&sb, "%s %s\n", labelStyle.Render("Strategy:"), valStyle.Render(template.Config.Strategy))
	fmt.Fprintf(&sb, "%s %s\n", labelStyle.Render("Max Concurrent Tasks:"), valStyle.Render(fmt.Sprintf("%d", template.Config.MaxConcurrentTasks)))
	fmt.Fprintf(&sb, "%s %s\n", labelStyle.Render("Timeout:"), valStyle.Render(template.Config.Timeout))

	reqApproval := "No"
	if template.Config.RequiresApproval {
		reqApproval = "Yes"
	}
	fmt.Fprintf(&sb, "%s %s\n", labelStyle.Render("Requires Approval:"), valStyle.Render(reqApproval))

	backupEnabled := "No"
	if template.Config.BackupEnabled {
		backupEnabled = "Yes"
	}
	fmt.Fprintf(&sb, "%s %s\n", labelStyle.Render("Backup Enabled:"), valStyle.Render(backupEnabled))

	sb.WriteString("\n" + theme.StepStyle.Render(fmt.Sprintf("Tasks (%d)", len(template.Tasks))) + "\n")
	for i, task := range template.Tasks {
		taskNameStyle := lipgloss.NewStyle().Foreground(theme.SuccessColor).Bold(true)
		fmt.Fprintf(&sb, "  %d. %s %s\n", i+1, taskNameStyle.Render(task.Name), labelStyle.Render("("+task.Type+")"))
		fmt.Fprintf(&sb, "     %s\n", task.Description)
		if len(task.DependsOn) > 0 {
			fmt.Fprintf(&sb, "     %s %s\n", labelStyle.Render("Depends on:"), strings.Join(task.DependsOn, ", "))
		}
	}

	return sb.String()
}
