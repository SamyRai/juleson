package commands

import (
	"errors"
	"testing"
	"time"

	"github.com/SamyRai/juleson/internal/orchestration/domain"
)

func TestExecutionResultFromDomainMapsPresentationDTO(t *testing.T) {
	start := time.Date(2026, 5, 25, 10, 0, 0, 0, time.UTC)
	end := start.Add(2 * time.Second)
	result := executionResultFromDomain("cleanup", "/repo", &domain.Result{
		Success:   true,
		Duration:  2 * time.Second,
		Learnings: []string{"keep scope narrow"},
		Tasks: []domain.TaskResult{{
			TaskName:  "Refactor",
			TaskType:  "code",
			StartTime: start,
			EndTime:   end,
			Duration:  2 * time.Second,
			Success:   true,
			SessionID: "session-1",
			Output:    "done",
			Metrics:   map[string]any{"dry_run": true},
		}},
	}, []string{"report.md"})

	if result.TemplateName != "cleanup" || result.ProjectPath != "/repo" || !result.Success {
		t.Fatalf("basic result fields not mapped: %+v", result)
	}
	if result.Duration != 2*time.Second || result.StartTime.IsZero() || result.EndTime.IsZero() {
		t.Fatalf("timing fields not mapped: %+v", result)
	}
	if len(result.OutputFiles) != 1 || result.OutputFiles[0] != "report.md" {
		t.Fatalf("output files not mapped: %+v", result.OutputFiles)
	}
	if len(result.Recommendations) != 1 || result.Recommendations[0] != "keep scope narrow" {
		t.Fatalf("recommendations not mapped: %+v", result.Recommendations)
	}
	if len(result.TasksExecuted) != 1 {
		t.Fatalf("tasks length = %d", len(result.TasksExecuted))
	}
	task := result.TasksExecuted[0]
	if task.TaskName != "Refactor" || task.JulesSessionID != "session-1" || task.Output != "done" {
		t.Fatalf("task fields not mapped: %+v", task)
	}
	if task.Metrics["dry_run"] != true {
		t.Fatalf("metrics not mapped: %+v", task.Metrics)
	}
}

func TestExecutionResultFromDomainMapsErrors(t *testing.T) {
	result := executionResultFromDomain("cleanup", "/repo", &domain.Result{
		Error: errors.New("template failed"),
		Tasks: []domain.TaskResult{{
			TaskName: "Refactor",
			Error:    errors.New("task failed"),
		}},
	}, nil)

	if result.Error != "template failed" {
		t.Fatalf("result error = %q", result.Error)
	}
	if len(result.TasksExecuted) != 1 || result.TasksExecuted[0].Error != "task failed" {
		t.Fatalf("task error not mapped: %+v", result.TasksExecuted)
	}
}
