package llm

import (
	"context"
	"fmt"

	ollama "github.com/SamyRai/ollama-go/client"
	"github.com/SamyRai/ollama-go/config"
	"github.com/SamyRai/ollama-go/structures"
)

// OllamaProvider implements Provider using the Ollama SDK.
type OllamaProvider struct {
	client *ollama.OllamaClient
	model  string
}

// NewOllamaProvider creates a new Ollama-backed LLM provider.
func NewOllamaProvider(baseURL, defaultModel string) *OllamaProvider {
	cfg := &config.Config{
		BaseURL: baseURL,
	}
	client := ollama.NewClient(cfg)
	return &OllamaProvider{
		client: client,
		model:  defaultModel,
	}
}

func (p *OllamaProvider) buildChatRequest(req Request, stream bool) structures.ChatRequest {
	model := req.Model
	if model == "" {
		model = p.model
	}

	var messages []structures.Message

	if req.System != "" {
		messages = append(messages, structures.Message{
			Role:    "system",
			Content: req.System,
		})
	}

	if len(req.Messages) > 0 {
		for _, msg := range req.Messages {
			role := string(msg.Role)
			if role == "" {
				role = "user"
			}

			// Map tool calls if any
			var toolCalls []structures.ToolCall
			if len(msg.ToolCalls) > 0 {
				for _, tc := range msg.ToolCalls {
					toolCalls = append(toolCalls, structures.ToolCall{
						Function: structures.ToolCallFunction{
							Name:      tc.Name,
							Arguments: tc.Arguments,
						},
					})
				}
			}

			messages = append(messages, structures.Message{
				Role:       role,
				Content:    msg.Content,
				ToolCalls:  toolCalls,
				ToolCallID: msg.ToolCallID,
			})
		}
	} else if req.Prompt != "" {
		messages = append(messages, structures.Message{
			Role:    "user",
			Content: req.Prompt,
		})
	}

	var tools []structures.Tool
	if len(req.Tools) > 0 {
		for _, t := range req.Tools {
			params := mapToOllamaParams(t.Parameters)
			tools = append(tools, structures.Tool{
				Type: "function",
				Function: structures.ToolFunction{
					Name:        t.Name,
					Description: t.Description,
					Parameters:  params,
				},
			})
		}
	}

	options := structures.Options{}
	if req.Temperature != 0 {
		options.Temperature = float64(req.Temperature)
	}

	return structures.ChatRequest{
		Model:    model,
		Messages: messages,
		Tools:    tools,
		Stream:   stream,
		Options:  options,
	}
}

func mapToOllamaParams(params map[string]interface{}) map[string]structures.ToolParam {
	result := make(map[string]structures.ToolParam)
	if params == nil {
		return result
	}

	// Usually JSON schema puts the actual fields in "properties"
	props, ok := params["properties"].(map[string]interface{})
	if !ok {
		return result
	}

	for k, v := range props {
		if vMap, ok := v.(map[string]interface{}); ok {
			p := structures.ToolParam{}
			if t, ok := vMap["type"].(string); ok {
				p.Type = t
			} else {
				p.Type = "string"
			}
			if desc, ok := vMap["description"].(string); ok {
				p.Description = desc
			}

			if enumList, ok := vMap["enum"].([]interface{}); ok {
				var enums []string
				for _, e := range enumList {
					if str, ok := e.(string); ok {
						enums = append(enums, str)
					}
				}
				p.Enum = enums
			} else if enumList, ok := vMap["enum"].([]string); ok {
				p.Enum = enumList
			}
			result[k] = p
		}
	}

	return result
}

// GenerateContent generates content using Ollama.
func (p *OllamaProvider) GenerateContent(ctx context.Context, req Request) (*Response, error) {
	chatReq := p.buildChatRequest(req, false)

	// Ollama currently handles context indirectly or via custom HTTP client
	resp, err := p.client.Chat(chatReq, nil)
	if err != nil {
		return nil, fmt.Errorf("ollama chat error: %w", err)
	}

	var funcCalls []FunctionCall
	for _, tc := range resp.Message.ToolCalls {
		funcCalls = append(funcCalls, FunctionCall{
			Name:      tc.Function.Name,
			Arguments: tc.Function.Arguments,
		})
	}

	// We don't get exact token usage easily from ChatResponse without looking at metadata,
	// but we can leave it empty or try to parse if available.
	tu := TokenUsage{}

	return &Response{
		Text:          resp.Message.Content,
		FunctionCalls: funcCalls,
		TokenUsage:    tu,
		FinishReason:  resp.DoneReason,
	}, nil
}

// GenerateContentStream generates streaming content using Ollama.
func (p *OllamaProvider) GenerateContentStream(ctx context.Context, req Request) (<-chan StreamChunk, error) {
	chatReq := p.buildChatRequest(req, true)
	ch := make(chan StreamChunk)

	go func() {
		defer close(ch)
		// Assuming Chat blocks until completion and calls callback incrementally
		_, err := p.client.Chat(chatReq, func(resp structures.ChatResponse) {
			if resp.Message.Content != "" {
				ch <- StreamChunk{Text: resp.Message.Content}
			}
		})

		if err != nil {
			ch <- StreamChunk{Error: err}
		}
	}()

	return ch, nil
}
