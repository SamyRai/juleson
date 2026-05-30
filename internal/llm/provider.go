package llm

import "context"

// Role represents the sender of a message.
type Role string

const (
	RoleSystem    Role = "system"
	RoleUser      Role = "user"
	RoleAssistant Role = "assistant"
	RoleTool      Role = "tool"
)

// Message represents a single chat message.
type Message struct {
	Role    Role
	Content string
	// ToolCalls contains any tool calls made by the assistant
	ToolCalls []FunctionCall
	// ToolCallID is used when Role is RoleTool to identify which call is being answered
	ToolCallID string
}

// Tool represents a function or tool the LLM can invoke.
type Tool struct {
	Name        string
	Description string
	// JSONSchema string or a structured object representing the parameters.
	// For simplicity, we can use an interface{} or raw JSON bytes here.
	Parameters map[string]interface{}
}

// FunctionCall represents the LLM deciding to call a tool.
type FunctionCall struct {
	ID        string
	Name      string
	Arguments map[string]interface{}
}

// Request encapsulates all data needed for an LLM generation call,
// making it easy to add new parameters in the future without breaking the interface.
type Request struct {
	Prompt string
	// Messages can be provided instead of Prompt for multi-turn chats
	Messages    []Message
	System      string
	Model       string
	Temperature float32
	MaxTokens   int
	Tools       []Tool
}

// Response encapsulates the LLM output and associated metadata.
type Response struct {
	Text          string
	FunctionCalls []FunctionCall
	TokenUsage    TokenUsage
	FinishReason  string
}

// TokenUsage tracks token consumption.
type TokenUsage struct {
	PromptTokens     int
	CompletionTokens int
	TotalTokens      int
}

// StreamChunk represents a chunk of streamed response.
type StreamChunk struct {
	Text  string
	Error error
}

// Provider defines the standard interface for all AI backends.
type Provider interface {
	GenerateContent(ctx context.Context, req Request) (*Response, error)
	GenerateContentStream(ctx context.Context, req Request) (<-chan StreamChunk, error)
}
