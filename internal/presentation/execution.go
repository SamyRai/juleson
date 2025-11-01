package presentation

import (
	"fmt"
	"strings"

	"github.com/SamyRai/juleson/internal/automation"
)

// ExecutionFormatter formats execution results
type ExecutionFormatter struct{}

// NewExecutionFormatter creates a new execution formatter
func NewExecutionFormatter() *ExecutionFormatter {
	return &ExecutionFormatter{}
}

// Format displays execution results
func (f *ExecutionFormatter) Format(result *automation.ExecutionResult) string {
	var sb strings.Builder

	sb.WriteString("🎯 Execution Results\n")
	sb.WriteString("====================\n")
	sb.WriteString(fmt.Sprintf("Template: %s\n", result.TemplateName))
	sb.WriteString(fmt.Sprintf("Project: %s\n", result.ProjectPath))
	sb.WriteString(fmt.Sprintf("Duration: %v\n", result.Duration))
	sb.WriteString(fmt.Sprintf("Success: %t\n", result.Success))

	if result.Error != "" {
		sb.WriteString(fmt.Sprintf("Error: %s\n", result.Error))
	}

	sb.WriteString(fmt.Sprintf("Tasks Executed: %d\n", len(result.TasksExecuted)))
	for _, task := range result.TasksExecuted {
		status := "✅"
		if !task.Success {
			status = "❌"
		}
		sb.WriteString(fmt.Sprintf("  %s %s (%s) - %v\n", status, task.TaskName, task.TaskType, task.Duration))
	}

	if len(result.OutputFiles) > 0 {
		sb.WriteString("\nOutput Files:\n")
		for _, file := range result.OutputFiles {
			sb.WriteString(fmt.Sprintf("  📄 %s\n", file))
		}
	}

	return sb.String()
}
