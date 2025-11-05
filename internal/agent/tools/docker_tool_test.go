package tools

import (
	"testing"

	"github.com/SamyRai/juleson/internal/agent"
)

func TestDockerTool_Name(t *testing.T) {
	tool := NewDockerTool()
	if tool.Name() != "docker" {
		t.Errorf("Expected name 'docker', got '%s'", tool.Name())
	}
}

func TestDockerTool_Description(t *testing.T) {
	tool := NewDockerTool()
	expected := "Manage Docker containers, images, and Docker operations. Build images, run containers, manage lifecycle, and execute commands."
	if tool.Description() != expected {
		t.Errorf("Description mismatch")
	}
}

func TestDockerTool_RequiresApproval(t *testing.T) {
	tool := NewDockerTool()
	if !tool.RequiresApproval() {
		t.Error("Docker tool should require approval")
	}
}

func TestDockerTool_CanHandle(t *testing.T) {
	tool := NewDockerTool()

	tests := []struct {
		task     agent.Task
		expected bool
	}{
		{
			task: agent.Task{
				Description: "build docker image",
				Prompt:      "Create a Docker image for the project",
			},
			expected: true,
		},
		{
			task: agent.Task{
				Description: "run container",
				Prompt:      "Start a Docker container",
			},
			expected: true,
		},
		{
			task: agent.Task{
				Description: "manage docker containers",
				Prompt:      "List and manage Docker containers",
			},
			expected: true,
		},
		{
			task: agent.Task{
				Description: "write documentation",
				Prompt:      "Create README file",
			},
			expected: false,
		},
		{
			task: agent.Task{
				Description: "fix bug",
				Prompt:      "Debug the application",
			},
			expected: false,
		},
		{
			task: agent.Task{
				Tool: "docker",
			},
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.task.Description, func(t *testing.T) {
			result := tool.CanHandle(tt.task)
			if result != tt.expected {
				t.Errorf("CanHandle() = %v, expected %v for task: %s", result, tt.expected, tt.task.Description)
			}
		})
	}
}

func TestDockerTool_Parameters(t *testing.T) {
	tool := NewDockerTool()
	params := tool.Parameters()

	if len(params) == 0 {
		t.Error("Docker tool should have parameters")
	}

	// Check for required action parameter
	found := false
	for _, param := range params {
		if param.Name == "action" {
			found = true
			if !param.Required {
				t.Error("action parameter should be required")
			}
			break
		}
	}
	if !found {
		t.Error("action parameter not found")
	}
}
