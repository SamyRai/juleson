package tools

import (
	"context"
	"fmt"
	"testing"

	"github.com/SamyRai/juleson/internal/agent"
)

// mockTool is a mock implementation of the Tool interface for testing
type mockTool struct {
	name            string
	canHandleFunc   func(agent.Task) bool
	requireApproval bool
}

func (m *mockTool) Name() string {
	return m.name
}

func (m *mockTool) Description() string {
	return "Mock tool for testing"
}

func (m *mockTool) Parameters() []Parameter {
	return []Parameter{}
}

func (m *mockTool) Execute(ctx context.Context, params map[string]interface{}) (*ToolResult, error) {
	return &ToolResult{Success: true}, nil
}

func (m *mockTool) RequiresApproval() bool {
	return m.requireApproval
}

func (m *mockTool) CanHandle(task agent.Task) bool {
	if m.canHandleFunc != nil {
		return m.canHandleFunc(task)
	}
	return true
}

func TestNewToolRegistry(t *testing.T) {
	registry := NewToolRegistry()
	if registry == nil {
		t.Fatal("expected non-nil registry")
	}

	tools := registry.List()
	if len(tools) != 0 {
		t.Errorf("expected empty registry, got %d tools", len(tools))
	}
}

func TestRegister(t *testing.T) {
	registry := NewToolRegistry()

	tool1 := &mockTool{name: "test-tool"}
	err := registry.Register(tool1)
	if err != nil {
		t.Fatalf("failed to register tool: %v", err)
	}

	// Verify tool was registered
	tools := registry.List()
	if len(tools) != 1 {
		t.Errorf("expected 1 tool, got %d", len(tools))
	}

	// Try to register duplicate
	err = registry.Register(tool1)
	if err == nil {
		t.Error("expected error when registering duplicate tool")
	}
}

func TestRegisterEmptyName(t *testing.T) {
	registry := NewToolRegistry()

	tool := &mockTool{name: ""}
	err := registry.Register(tool)
	if err == nil {
		t.Error("expected error when registering tool with empty name")
	}
}

func TestGet(t *testing.T) {
	registry := NewToolRegistry()

	tool := &mockTool{name: "test-tool"}
	registry.Register(tool)

	// Get existing tool
	retrieved, err := registry.Get("test-tool")
	if err != nil {
		t.Fatalf("failed to get tool: %v", err)
	}
	if retrieved.Name() != "test-tool" {
		t.Errorf("expected tool name 'test-tool', got '%s'", retrieved.Name())
	}

	// Try to get non-existent tool
	_, err = registry.Get("non-existent")
	if err == nil {
		t.Error("expected error when getting non-existent tool")
	}
}

func TestList(t *testing.T) {
	registry := NewToolRegistry()

	tool1 := &mockTool{name: "tool1"}
	tool2 := &mockTool{name: "tool2"}
	tool3 := &mockTool{name: "tool3"}

	registry.Register(tool1)
	registry.Register(tool2)
	registry.Register(tool3)

	tools := registry.List()
	if len(tools) != 3 {
		t.Errorf("expected 3 tools, got %d", len(tools))
	}

	// Verify all tools are present
	names := make(map[string]bool)
	for _, tool := range tools {
		names[tool.Name()] = true
	}

	if !names["tool1"] || !names["tool2"] || !names["tool3"] {
		t.Error("not all registered tools were returned")
	}
}

func TestFindForTask(t *testing.T) {
	registry := NewToolRegistry()

	// Register tools with different capabilities
	tool1 := &mockTool{
		name: "tool1",
		canHandleFunc: func(task agent.Task) bool {
			return task.Name == "task1"
		},
	}
	tool2 := &mockTool{
		name: "tool2",
		canHandleFunc: func(task agent.Task) bool {
			return task.Name == "task2"
		},
	}
	toolAll := &mockTool{
		name: "tool-all",
		canHandleFunc: func(task agent.Task) bool {
			return true
		},
	}

	registry.Register(tool1)
	registry.Register(tool2)
	registry.Register(toolAll)

	// Test finding tools for task1
	task1 := agent.Task{Name: "task1"}
	matches := registry.FindForTask(task1)
	if len(matches) != 2 {
		t.Errorf("expected 2 matching tools for task1, got %d", len(matches))
	}

	// Test finding tools for task2
	task2 := agent.Task{Name: "task2"}
	matches = registry.FindForTask(task2)
	if len(matches) != 2 {
		t.Errorf("expected 2 matching tools for task2, got %d", len(matches))
	}

	// Test with preferred tool
	task3 := agent.Task{Name: "task1", Tool: "tool1"}
	matches = registry.FindForTask(task3)
	if len(matches) != 1 {
		t.Errorf("expected 1 tool when preferred tool specified, got %d", len(matches))
	}
	if matches[0].Name() != "tool1" {
		t.Errorf("expected preferred tool 'tool1', got '%s'", matches[0].Name())
	}
}

func TestConcurrentAccess(t *testing.T) {
	registry := NewToolRegistry()

	// Test concurrent registration
	done := make(chan bool)
	for i := 0; i < 10; i++ {
		go func(n int) {
			tool := &mockTool{name: fmt.Sprintf("tool%d", n)}
			registry.Register(tool)
			done <- true
		}(i)
	}

	for i := 0; i < 10; i++ {
		<-done
	}

	tools := registry.List()
	if len(tools) != 10 {
		t.Errorf("expected 10 tools after concurrent registration, got %d", len(tools))
	}
}
