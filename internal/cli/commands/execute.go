package commands

import (
	"fmt"
	"strings"
	"time"

	"github.com/SamyRai/juleson/internal/orchestration"
	"github.com/SamyRai/juleson/internal/orchestration/domain"
	"github.com/SamyRai/juleson/internal/presentation"

	"github.com/spf13/cobra"
)

// NewExecuteCommand creates the execute command
func NewExecuteCommand(initializeRuntime func() (*orchestration.Runtime, error), displayExecutionResult func(*presentation.ExecutionResult)) *cobra.Command {
	executeCmd := &cobra.Command{
		Use:   "execute",
		Short: "Execute automation tasks",
		Long:  "Execute templates and automation tasks on projects",
	}

	// Execute template
	executeCmd.AddCommand(&cobra.Command{
		Use:   "template [template-name] [project-path]",
		Short: "Execute a template on a project",
		Long:  "Execute a specific template on a project",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			templateName := args[0]
			projectPath := args[1]

			runtime, err := initializeRuntime()
			if err != nil {
				return fmt.Errorf("failed to initialize orchestration runtime: %w", err)
			}

			result, outputFiles, err := runtime.TemplateRunner().Run(cmd.Context(), templateName, projectPath, make(map[string]string))
			if err != nil {
				return fmt.Errorf("failed to execute template: %w", err)
			}

			displayExecutionResult(executionResultFromDomain(templateName, projectPath, result, outputFiles))

			return nil
		},
	})

	// Execute with custom parameters
	executeCmd.AddCommand(&cobra.Command{
		Use:   "template-with-params [template-name] [project-path] [key=value]...",
		Short: "Execute template with custom parameters",
		Long:  "Execute a template with custom parameters",
		Args:  cobra.MinimumNArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			templateName := args[0]
			projectPath := args[1]

			// Parse custom parameters
			customParams := make(map[string]string)
			for i := 2; i < len(args); i++ {
				parts := strings.SplitN(args[i], "=", 2)
				if len(parts) == 2 {
					customParams[parts[0]] = parts[1]
				}
			}

			runtime, err := initializeRuntime()
			if err != nil {
				return fmt.Errorf("failed to initialize orchestration runtime: %w", err)
			}

			result, outputFiles, err := runtime.TemplateRunner().Run(cmd.Context(), templateName, projectPath, customParams)
			if err != nil {
				return fmt.Errorf("failed to execute template: %w", err)
			}

			displayExecutionResult(executionResultFromDomain(templateName, projectPath, result, outputFiles))

			return nil
		},
	})

	return executeCmd
}

func executionResultFromDomain(templateName, projectPath string, result *domain.Result, outputFiles []string) *presentation.ExecutionResult {
	converted := &presentation.ExecutionResult{
		TemplateName: templateName,
		ProjectPath:  projectPath,
		OutputFiles:  outputFiles,
		Metrics:      map[string]any{},
	}
	if result == nil {
		converted.Error = "no execution result"
		return converted
	}
	converted.Recommendations = append([]string(nil), result.Learnings...)
	converted.EndTime = time.Now()
	converted.StartTime = converted.EndTime.Add(-result.Duration)
	converted.Duration = result.Duration
	converted.Success = result.Success
	if result.Error != nil {
		converted.Error = result.Error.Error()
	}
	converted.TasksExecuted = make([]presentation.TaskExecutionResult, 0, len(result.Tasks))
	for _, task := range result.Tasks {
		taskResult := presentation.TaskExecutionResult{
			TaskName:       task.TaskName,
			TaskType:       task.TaskType,
			StartTime:      task.StartTime,
			EndTime:        task.EndTime,
			Duration:       task.Duration,
			Success:        task.Success,
			JulesSessionID: task.SessionID,
			Output:         task.Output,
			Metrics:        task.Metrics,
		}
		if task.Error != nil {
			taskResult.Error = task.Error.Error()
		}
		converted.TasksExecuted = append(converted.TasksExecuted, taskResult)
	}
	return converted
}
