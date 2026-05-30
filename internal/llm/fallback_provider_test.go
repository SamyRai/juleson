package llm

import (
	"context"
	"errors"
	"log/slog"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type MockProvider struct {
	GenerateContentFunc       func(ctx context.Context, req Request) (*Response, error)
	GenerateContentStreamFunc func(ctx context.Context, req Request) (<-chan StreamChunk, error)
}

func (m *MockProvider) GenerateContent(ctx context.Context, req Request) (*Response, error) {
	if m.GenerateContentFunc != nil {
		return m.GenerateContentFunc(ctx, req)
	}
	return nil, nil
}

func (m *MockProvider) GenerateContentStream(ctx context.Context, req Request) (<-chan StreamChunk, error) {
	if m.GenerateContentStreamFunc != nil {
		return m.GenerateContentStreamFunc(ctx, req)
	}
	return nil, nil
}

func TestFallbackProvider_GenerateContent(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	ctx := context.Background()
	req := Request{Prompt: "hello"}

	t.Run("primary success", func(t *testing.T) {
		primary := &MockProvider{
			GenerateContentFunc: func(ctx context.Context, req Request) (*Response, error) {
				return &Response{Text: "primary success"}, nil
			},
		}
		secondary := &MockProvider{
			GenerateContentFunc: func(ctx context.Context, req Request) (*Response, error) {
				return &Response{Text: "secondary success"}, nil
			},
		}

		provider := NewFallbackProvider(primary, secondary, logger)
		resp, err := provider.GenerateContent(ctx, req)

		require.NoError(t, err)
		assert.Equal(t, "primary success", resp.Text)
	})

	t.Run("primary fails, secondary success", func(t *testing.T) {
		primary := &MockProvider{
			GenerateContentFunc: func(ctx context.Context, req Request) (*Response, error) {
				return nil, errors.New("primary error")
			},
		}
		secondary := &MockProvider{
			GenerateContentFunc: func(ctx context.Context, req Request) (*Response, error) {
				return &Response{Text: "secondary success"}, nil
			},
		}

		provider := NewFallbackProvider(primary, secondary, logger)
		resp, err := provider.GenerateContent(ctx, req)

		require.NoError(t, err)
		assert.Equal(t, "secondary success", resp.Text)
	})

	t.Run("both fail", func(t *testing.T) {
		primary := &MockProvider{
			GenerateContentFunc: func(ctx context.Context, req Request) (*Response, error) {
				return nil, errors.New("primary error")
			},
		}
		secondary := &MockProvider{
			GenerateContentFunc: func(ctx context.Context, req Request) (*Response, error) {
				return nil, errors.New("secondary error")
			},
		}

		provider := NewFallbackProvider(primary, secondary, logger)
		_, err := provider.GenerateContent(ctx, req)

		require.Error(t, err)
		assert.Contains(t, err.Error(), "primary error")
		assert.Contains(t, err.Error(), "secondary error")
	})
}

func TestFallbackProvider_GenerateContentStream(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	ctx := context.Background()
	req := Request{Prompt: "hello"}

	t.Run("primary success", func(t *testing.T) {
		primary := &MockProvider{
			GenerateContentStreamFunc: func(ctx context.Context, req Request) (<-chan StreamChunk, error) {
				ch := make(chan StreamChunk, 1)
				ch <- StreamChunk{Text: "primary stream"}
				close(ch)
				return ch, nil
			},
		}
		secondary := &MockProvider{
			GenerateContentStreamFunc: func(ctx context.Context, req Request) (<-chan StreamChunk, error) {
				return nil, errors.New("should not be called")
			},
		}

		provider := NewFallbackProvider(primary, secondary, logger)
		ch, err := provider.GenerateContentStream(ctx, req)

		require.NoError(t, err)
		chunk := <-ch
		assert.Equal(t, "primary stream", chunk.Text)
	})

	t.Run("primary fails, secondary success", func(t *testing.T) {
		primary := &MockProvider{
			GenerateContentStreamFunc: func(ctx context.Context, req Request) (<-chan StreamChunk, error) {
				return nil, errors.New("primary stream error")
			},
		}
		secondary := &MockProvider{
			GenerateContentStreamFunc: func(ctx context.Context, req Request) (<-chan StreamChunk, error) {
				ch := make(chan StreamChunk, 1)
				ch <- StreamChunk{Text: "secondary stream"}
				close(ch)
				return ch, nil
			},
		}

		provider := NewFallbackProvider(primary, secondary, logger)
		ch, err := provider.GenerateContentStream(ctx, req)

		require.NoError(t, err)
		chunk := <-ch
		assert.Equal(t, "secondary stream", chunk.Text)
	})

	t.Run("both fail", func(t *testing.T) {
		primary := &MockProvider{
			GenerateContentStreamFunc: func(ctx context.Context, req Request) (<-chan StreamChunk, error) {
				return nil, errors.New("primary stream error")
			},
		}
		secondary := &MockProvider{
			GenerateContentStreamFunc: func(ctx context.Context, req Request) (<-chan StreamChunk, error) {
				return nil, errors.New("secondary stream error")
			},
		}

		provider := NewFallbackProvider(primary, secondary, logger)
		_, err := provider.GenerateContentStream(ctx, req)

		require.Error(t, err)
		assert.Contains(t, err.Error(), "primary stream error")
		assert.Contains(t, err.Error(), "secondary stream error")
	})
}
