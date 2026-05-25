package automation

import (
	"context"
	"fmt"

	"github.com/SamyRai/juleson/internal/templates"
)

// executeTasks executes template tasks in dependency order.
func (e *Engine) executeTasks(ctx context.Context, tasks []templates.TemplateTask) ([]TaskExecutionResult, error) {
	var results []TaskExecutionResult
	executed := make(map[string]bool)

	for len(executed) < len(tasks) {
		progress := false

		for _, task := range tasks {
			if executed[task.Name] {
				continue
			}

			depsSatisfied := true
			for _, dep := range task.DependsOn {
				if !executed[dep] {
					depsSatisfied = false
					break
				}
			}

			if !depsSatisfied {
				continue
			}

			result, err := e.executeTask(ctx, task)
			if err != nil {
				return results, fmt.Errorf("task '%s' failed: %w", task.Name, err)
			}

			results = append(results, *result)
			executed[task.Name] = true
			progress = true
		}

		if !progress {
			return results, fmt.Errorf("circular dependency detected in tasks")
		}
	}

	return results, nil
}
