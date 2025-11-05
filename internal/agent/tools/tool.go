package tools

import (
	"context"

	"github.com/SamyRai/juleson/internal/agent"
)

// Tool represents an action the agent can perform
type Tool interface {
	// Name returns the unique identifier for this tool
	Name() string

	// Description returns what this tool does
	Description() string

	// Parameters returns the parameters this tool accepts
	Parameters() []Parameter

	// Execute runs the tool with given parameters
	Execute(ctx context.Context, params map[string]interface{}) (*ToolResult, error)

	// RequiresApproval returns true if this tool needs human approval
	RequiresApproval() bool

	// CanHandle returns true if this tool can handle the given task
	CanHandle(task agent.Task) bool
}

// Parameter describes a tool parameter
type Parameter struct {
	Name        string
	Description string
	Type        ParameterType
	Required    bool
	Default     interface{}
	Validation  ValidationFunc
}

// ParameterType represents parameter data types
type ParameterType string

const (
	ParameterTypeString ParameterType = "STRING"
	ParameterTypeInt    ParameterType = "INT"
	ParameterTypeFloat  ParameterType = "FLOAT"
	ParameterTypeBool   ParameterType = "BOOL"
	ParameterTypeArray  ParameterType = "ARRAY"
	ParameterTypeObject ParameterType = "OBJECT"
)

// ValidationFunc validates a parameter value
type ValidationFunc func(value interface{}) error

// ToolResult represents the outcome of tool execution
type ToolResult struct {
	Success   bool
	Output    interface{}
	Changes   []agent.Change
	Artifacts []agent.Artifact
	Error     error
	Duration  int64 // milliseconds
	Metadata  map[string]interface{}
}

// ToolRegistry manages available tools
type ToolRegistry interface {
	// Register a tool
	Register(tool Tool) error

	// Get a tool by name
	Get(name string) (Tool, error)

	// List all tools
	List() []Tool

	// Find tools that can handle a task
	FindForTask(task agent.Task) []Tool
}
