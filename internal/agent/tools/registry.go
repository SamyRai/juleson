package tools

import (
	"fmt"
	"sync"

	"github.com/SamyRai/juleson/internal/agent"
)

// toolRegistry is a thread-safe registry for managing tools
type toolRegistry struct {
	tools map[string]Tool
	mu    sync.RWMutex
}

// NewToolRegistry creates a new tool registry
func NewToolRegistry() ToolRegistry {
	return &toolRegistry{
		tools: make(map[string]Tool),
	}
}

// Register adds a tool to the registry
func (r *toolRegistry) Register(tool Tool) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	name := tool.Name()
	if name == "" {
		return fmt.Errorf("tool name cannot be empty")
	}

	if _, exists := r.tools[name]; exists {
		return fmt.Errorf("tool %s already registered", name)
	}

	r.tools[name] = tool
	return nil
}

// Get retrieves a tool by name
func (r *toolRegistry) Get(name string) (Tool, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	tool, exists := r.tools[name]
	if !exists {
		return nil, fmt.Errorf("tool %s not found", name)
	}

	return tool, nil
}

// List returns all registered tools
func (r *toolRegistry) List() []Tool {
	r.mu.RLock()
	defer r.mu.RUnlock()

	tools := make([]Tool, 0, len(r.tools))
	for _, tool := range r.tools {
		tools = append(tools, tool)
	}

	return tools
}

// FindForTask finds tools that can handle a specific task
func (r *toolRegistry) FindForTask(task agent.Task) []Tool {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var matchingTools []Tool

	// If task specifies a preferred tool, try that first
	if task.Tool != "" {
		if tool, exists := r.tools[task.Tool]; exists && tool.CanHandle(task) {
			return []Tool{tool}
		}
	}

	// Otherwise, find all tools that can handle this task
	for _, tool := range r.tools {
		if tool.CanHandle(task) {
			matchingTools = append(matchingTools, tool)
		}
	}

	return matchingTools
}
