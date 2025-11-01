package commands

import (
	"fmt"
	"strings"

	"github.com/SamyRai/juleson/internal/analyzer"
	"github.com/SamyRai/juleson/internal/automation"
	"github.com/SamyRai/juleson/internal/templates"
)

// DisplayProjectAnalysis displays the project analysis results
func DisplayProjectAnalysis(context *analyzer.ProjectContext) {
	fmt.Println("📊 Project Analysis Results")
	fmt.Println("==========================")
	fmt.Printf("Project Name: %s\n", context.ProjectName)
	fmt.Printf("Project Type: %s\n", context.ProjectType)
	fmt.Printf("Languages: %s\n", strings.Join(context.Languages, ", "))
	fmt.Printf("Frameworks: %s\n", strings.Join(context.Frameworks, ", "))
	fmt.Printf("Architecture: %s\n", context.Architecture)
	fmt.Printf("Complexity: %s\n", context.Complexity)
	fmt.Printf("Git Status: %s\n", context.GitStatus)
	fmt.Printf("Dependencies: %d\n", len(context.Dependencies))
	fmt.Printf("File Types: %d\n", len(context.FileStructure))
}

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

// DisplayExecutionResult displays execution results
func DisplayExecutionResult(result *automation.ExecutionResult) {
	fmt.Println("🎯 Execution Results")
	fmt.Println("====================")
	fmt.Printf("Template: %s\n", result.TemplateName)
	fmt.Printf("Project: %s\n", result.ProjectPath)
	fmt.Printf("Duration: %v\n", result.Duration)
	fmt.Printf("Success: %t\n", result.Success)

	if result.Error != "" {
		fmt.Printf("Error: %s\n", result.Error)
	}

	fmt.Printf("Tasks Executed: %d\n", len(result.TasksExecuted))
	for _, task := range result.TasksExecuted {
		status := "✅"
		if !task.Success {
			status = "❌"
		}
		fmt.Printf("  %s %s (%s) - %v\n", status, task.TaskName, task.TaskType, task.Duration)
	}

	if len(result.OutputFiles) > 0 {
		fmt.Println("\nOutput Files:")
		for _, file := range result.OutputFiles {
			fmt.Printf("  📄 %s\n", file)
		}
	}
}
