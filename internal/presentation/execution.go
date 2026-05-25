package presentation

import (
	"fmt"
	"strings"
	"time"
)

// ExecutionResult is the presentation boundary for template execution output.
type ExecutionResult struct {
	TemplateName    string
	ProjectPath     string
	StartTime       time.Time
	EndTime         time.Time
	Duration        time.Duration
	Success         bool
	TasksExecuted   []TaskExecutionResult
	OutputFiles     []string
	Recommendations []string
	Error           string
	Metrics         map[string]any
}

// TaskExecutionResult is the presentation boundary for a single executed task.
type TaskExecutionResult struct {
	TaskName       string
	TaskType       string
	StartTime      time.Time
	EndTime        time.Time
	Duration       time.Duration
	Success        bool
	JulesSessionID string
	Output         string
	Error          string
	Metrics        map[string]any
}

// ExecutionFormatter formats execution results
type ExecutionFormatter struct{}

// NewExecutionFormatter creates a new execution formatter
func NewExecutionFormatter() *ExecutionFormatter {
	return &ExecutionFormatter{}
}

// Format displays execution results
func (f *ExecutionFormatter) Format(result *ExecutionResult) string {
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
