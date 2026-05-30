package llm

import (
	"context"
	"fmt"
	"log/slog"
)

// FallbackProvider implements Provider by attempting requests against a primary provider,
// and automatically falling back to a secondary provider if the primary fails.
type FallbackProvider struct {
	primary   Provider
	secondary Provider
	logger    *slog.Logger
}

// NewFallbackProvider creates a new provider with fallback capabilities.
func NewFallbackProvider(primary, secondary Provider, logger *slog.Logger) *FallbackProvider {
	if logger == nil {
		logger = slog.Default()
	}
	return &FallbackProvider{
		primary:   primary,
		secondary: secondary,
		logger:    logger,
	}
}

// GenerateContent attempts primary, then secondary on error.
func (f *FallbackProvider) GenerateContent(ctx context.Context, req Request) (*Response, error) {
	resp, err := f.primary.GenerateContent(ctx, req)
	if err == nil {
		return resp, nil
	}

	f.logger.Warn("Primary LLM provider failed, falling back to secondary", "error", err)

	resp, err2 := f.secondary.GenerateContent(ctx, req)
	if err2 != nil {
		return nil, fmt.Errorf("both primary and secondary providers failed. primary err: %v, secondary err: %v", err, err2)
	}

	return resp, nil
}

// GenerateContentStream attempts primary, then secondary on error.
func (f *FallbackProvider) GenerateContentStream(ctx context.Context, req Request) (<-chan StreamChunk, error) {
	// For streaming, we can try to start the stream on the primary.
	// If it fails immediately to return a channel, we fallback.
	// If it returns a channel but fails midway, it's harder to fallback cleanly without replaying context,
	// but for simplicity we handle immediate stream initiation failure.
	ch, err := f.primary.GenerateContentStream(ctx, req)
	if err == nil {
		// We could wrap the channel to detect mid-stream errors and fallback, but that's complex
		// as the client already received partial text.
		return ch, nil
	}

	f.logger.Warn("Primary LLM provider failed to start stream, falling back to secondary", "error", err)

	ch, err2 := f.secondary.GenerateContentStream(ctx, req)
	if err2 != nil {
		return nil, fmt.Errorf("both primary and secondary providers failed to start stream. primary err: %v, secondary err: %v", err, err2)
	}

	return ch, nil
}
