package gemini

import (
	"context"
	"fmt"
	"time"

	"google.golang.org/genai"
)

// Client represents a Google Gemini AI client
type Client struct {
	client *genai.Client
	config *Config
	ctx    context.Context
}

// Config contains Gemini client configuration
type Config struct {
	APIKey    string
	Backend   string
	Project   string
	Location  string
	Model     string
	Timeout   time.Duration
	MaxTokens int
}

// NewClient creates a new Gemini AI client
func NewClient(config *Config) (*Client, error) {
	ctx := context.Background()

	// Set timeout for context if specified
	if config.Timeout > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, config.Timeout)
		// Note: cancel is not stored as the client is meant to be long-lived
		_ = cancel
	}

	// Create client configuration
	clientConfig := &genai.ClientConfig{}

	// Configure based on backend
	switch config.Backend {
	case "gemini-api":
		if config.APIKey == "" {
			return nil, fmt.Errorf("API key is required for Gemini API backend")
		}
		clientConfig.APIKey = config.APIKey
		clientConfig.Backend = genai.BackendGeminiAPI
	case "vertex-ai":
		if config.Project == "" {
			return nil, fmt.Errorf("project is required for Vertex AI backend")
		}
		if config.Location == "" {
			config.Location = "us-central1" // Default location
		}
		clientConfig.Project = config.Project
		clientConfig.Location = config.Location
		clientConfig.Backend = genai.BackendVertexAI
	default:
		return nil, fmt.Errorf("unsupported backend: %s (supported: gemini-api, vertex-ai)", config.Backend)
	}

	// Create the GenAI client
	genaiClient, err := genai.NewClient(ctx, clientConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create Gemini client: %w", err)
	}

	return &Client{
		client: genaiClient,
		config: config,
		ctx:    ctx,
	}, nil
}

// GenerateContent generates content using the specified model and prompt
func (c *Client) GenerateContent(model, prompt string) (*genai.GenerateContentResponse, error) {
	if model == "" {
		model = c.config.Model
	}

	// Create content parts
	parts := []*genai.Part{
		{Text: prompt},
	}
	content := []*genai.Content{
		{Parts: parts},
	}

	// Generate content with default config
	resp, err := c.client.Models.GenerateContent(c.ctx, model, content, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to generate content: %w", err)
	}

	return resp, nil
}

// GenerateContentWithImages generates content with image input
func (c *Client) GenerateContentWithImages(model, prompt string, imageBytes []byte, mimeType string) (*genai.GenerateContentResponse, error) {
	if model == "" {
		model = c.config.Model
	}

	// Create content parts with text and image
	parts := []*genai.Part{
		{Text: prompt},
		{InlineData: &genai.Blob{Data: imageBytes, MIMEType: mimeType}},
	}
	content := []*genai.Content{
		{Parts: parts},
	}

	// Generate content
	resp, err := c.client.Models.GenerateContent(c.ctx, model, content, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to generate content with images: %w", err)
	}

	return resp, nil
}

// Generate is a convenience method for simple text generation
func (c *Client) Generate(ctx context.Context, prompt string) (string, error) {
	model := c.config.Model
	if model == "" {
		model = "gemini-1.5-flash" // Default model
	}

	resp, err := c.GenerateContent(model, prompt)
	if err != nil {
		return "", err
	}

	// Extract text from response
	if len(resp.Candidates) == 0 {
		return "", fmt.Errorf("no candidates in response")
	}

	candidate := resp.Candidates[0]
	if candidate.Content == nil || len(candidate.Content.Parts) == 0 {
		return "", fmt.Errorf("no content in response")
	}

	// Concatenate all text parts
	var result string
	for _, part := range candidate.Content.Parts {
		if part.Text != "" {
			result += part.Text
		}
	}

	return result, nil
}

// Close cleans up resources
func (c *Client) Close() error {
	// The genai client doesn't have a Close method, just return nil
	return nil
}
