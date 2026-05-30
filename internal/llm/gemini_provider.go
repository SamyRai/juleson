package llm

import (
	"context"
	"fmt"

	"github.com/SamyRai/juleson/internal/gemini"
	"google.golang.org/genai"
)

// GeminiProvider implements Provider using the Gemini client.
type GeminiProvider struct {
	client *gemini.Client
	model  string
}

// NewGeminiProvider creates a new Gemini-backed LLM provider.
func NewGeminiProvider(client *gemini.Client, defaultModel string) *GeminiProvider {
	return &GeminiProvider{
		client: client,
		model:  defaultModel,
	}
}

func (p *GeminiProvider) buildGenAIRequest(req Request) (string, []*genai.Content, *genai.GenerateContentConfig) {
	model := req.Model
	if model == "" {
		model = p.model
	}

	config := &genai.GenerateContentConfig{}

	if req.Temperature != 0 {
		temp := float32(req.Temperature)
		config.Temperature = &temp
	}
	if req.MaxTokens != 0 {
		tokens := int32(req.MaxTokens)
		config.MaxOutputTokens = tokens
	}
	if req.System != "" {
		config.SystemInstruction = &genai.Content{
			Parts: []*genai.Part{{Text: req.System}},
		}
	}

	if len(req.Tools) > 0 {
		var funcDecls []*genai.FunctionDeclaration
		for _, t := range req.Tools {
			decl := &genai.FunctionDeclaration{
				Name:        t.Name,
				Description: t.Description,
			}
			if t.Parameters != nil {
				decl.Parameters = mapToGenAISchema(t.Parameters)
			}
			funcDecls = append(funcDecls, decl)
		}
		config.Tools = []*genai.Tool{
			{FunctionDeclarations: funcDecls},
		}
	}

	var contents []*genai.Content
	if len(req.Messages) > 0 {
		for _, msg := range req.Messages {
			role := "user"
			if msg.Role == RoleAssistant {
				role = "model"
			} else if msg.Role == RoleTool {
				role = "function"
			}
			// Map ToolCalls and ToolCallID properly later
			contents = append(contents, &genai.Content{
				Role:  role,
				Parts: []*genai.Part{{Text: msg.Content}},
			})
		}
	} else if req.Prompt != "" {
		contents = append(contents, &genai.Content{
			Role:  "user",
			Parts: []*genai.Part{{Text: req.Prompt}},
		})
	}

	return model, contents, config
}

func mapToGenAISchema(params map[string]interface{}) *genai.Schema {
	if params == nil {
		return nil
	}
	schema := &genai.Schema{}

	if t, ok := params["type"].(string); ok {
		switch t {
		case "string":
			schema.Type = genai.TypeString
		case "number":
			schema.Type = genai.TypeNumber
		case "integer":
			schema.Type = genai.TypeInteger
		case "boolean":
			schema.Type = genai.TypeBoolean
		case "array":
			schema.Type = genai.TypeArray
		case "object":
			schema.Type = genai.TypeObject
		}
	}

	if desc, ok := params["description"].(string); ok {
		schema.Description = desc
	}

	if req, ok := params["required"].([]interface{}); ok {
		for _, r := range req {
			if str, ok := r.(string); ok {
				schema.Required = append(schema.Required, str)
			}
		}
	} else if req, ok := params["required"].([]string); ok {
		schema.Required = req
	}

	if items, ok := params["items"].(map[string]interface{}); ok {
		schema.Items = mapToGenAISchema(items)
	}

	if props, ok := params["properties"].(map[string]interface{}); ok {
		schema.Properties = make(map[string]*genai.Schema)
		for k, v := range props {
			if vMap, ok := v.(map[string]interface{}); ok {
				schema.Properties[k] = mapToGenAISchema(vMap)
			}
		}
	}

	return schema
}

func (p *GeminiProvider) extractResponse(resp *genai.GenerateContentResponse) *Response {
	var text string
	var funcCalls []FunctionCall

	if len(resp.Candidates) > 0 {
		candidate := resp.Candidates[0]
		if candidate.Content != nil {
			for _, part := range candidate.Content.Parts {
				if part.Text != "" {
					text += part.Text
				}
				if part.FunctionCall != nil {
					args := make(map[string]interface{})
					if part.FunctionCall.Args != nil {
						// Simplistic extraction
						for k, v := range part.FunctionCall.Args {
							args[k] = v
						}
					}
					funcCalls = append(funcCalls, FunctionCall{
						Name:      part.FunctionCall.Name,
						Arguments: args,
					})
				}
			}
		}
	}

	// Token usage
	var tu TokenUsage
	if resp.UsageMetadata != nil {
		tu.PromptTokens = int(resp.UsageMetadata.PromptTokenCount)
		tu.CompletionTokens = int(resp.UsageMetadata.CandidatesTokenCount)
		tu.TotalTokens = int(resp.UsageMetadata.TotalTokenCount)
	}

	return &Response{
		Text:          text,
		FunctionCalls: funcCalls,
		TokenUsage:    tu,
	}
}

// GenerateContent generates content using Gemini.
func (p *GeminiProvider) GenerateContent(ctx context.Context, req Request) (*Response, error) {
	model, contents, config := p.buildGenAIRequest(req)

	resp, err := p.client.GenAIClient().Models.GenerateContent(ctx, model, contents, config)
	if err != nil {
		return nil, fmt.Errorf("gemini generate error: %w", err)
	}

	return p.extractResponse(resp), nil
}

// GenerateContentStream generates streaming content.
func (p *GeminiProvider) GenerateContentStream(ctx context.Context, req Request) (<-chan StreamChunk, error) {
	model, contents, config := p.buildGenAIRequest(req)

	iter := p.client.GenAIClient().Models.GenerateContentStream(ctx, model, contents, config)
	ch := make(chan StreamChunk)

	go func() {
		defer close(ch)
		for resp, err := range iter {
			if err != nil {
				ch <- StreamChunk{Error: err}
				return
			}

			if len(resp.Candidates) > 0 && resp.Candidates[0].Content != nil {
				for _, part := range resp.Candidates[0].Content.Parts {
					if part.Text != "" {
						ch <- StreamChunk{Text: part.Text}
					}
				}
			}
		}
	}()

	return ch, nil
}
