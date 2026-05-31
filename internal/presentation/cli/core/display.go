package core

import (
	"fmt"
	"strings"

	"github.com/SamyRai/juleson/internal/templates"
)

// DisplayTemplates displays a list of templates
func DisplayTemplates(templates []templates.RegistryTemplate) {
	fmt.Println("📋 Available Templates")
	fmt.Println("=====================")

	for _, template := range templates {
		fmt.Printf("• %s (%s) - %s\n", template.Name, template.Category, template.Description)
		fmt.Printf("  Tags: %s\n", strings.Join(template.Tags, ", "))
		fmt.Printf("  Complexity: %s | Duration: %s\n", template.Complexity, template.EstimatedDuration)
		fmt.Println()
	}
}

// DisplayTemplateDetails displays detailed information about a template
func DisplayTemplateDetails(template *templates.Template) {
	fmt.Printf("📄 Template Details: %s\n", template.Metadata.Name)
	fmt.Println("========================")
	fmt.Printf("Version: %s\n", template.Metadata.Version)
	fmt.Printf("Category: %s\n", template.Metadata.Category)
	fmt.Printf("Description: %s\n", template.Metadata.Description)
	fmt.Printf("Author: %s\n", template.Metadata.Author)
	fmt.Printf("Tags: %s\n", strings.Join(template.Metadata.Tags, ", "))
	fmt.Printf("Strategy: %s\n", template.Config.Strategy)
	fmt.Printf("Max Concurrent Tasks: %d\n", template.Config.MaxConcurrentTasks)
	fmt.Printf("Timeout: %s\n", template.Config.Timeout)
	fmt.Printf("Requires Approval: %t\n", template.Config.RequiresApproval)
	fmt.Printf("Backup Enabled: %t\n", template.Config.BackupEnabled)
	fmt.Printf("Tasks: %d\n", len(template.Tasks))

	fmt.Println("\nTasks:")
	for i, task := range template.Tasks {
		fmt.Printf("  %d. %s (%s)\n", i+1, task.Name, task.Type)
		fmt.Printf("     %s\n", task.Description)
		if len(task.DependsOn) > 0 {
			fmt.Printf("     Depends on: %s\n", strings.Join(task.DependsOn, ", "))
		}
	}
}
