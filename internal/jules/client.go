package jules

import (
	"net/http"
	"time"

	"github.com/SamyRai/juleson/internal/events"
)

// EventEmitter is an interface for emitting events
type EventEmitter interface {
	EmitSessionEvent(ctx interface{}, eventType events.EventType, data events.SessionEventData) error
	EmitActivityEvent(ctx interface{}, eventType events.EventType, data events.ActivityEventData) error
}

// Client represents a Jules API client
type Client struct {
	APIKey        string
	BaseURL       string
	HTTPClient    *http.Client
	RetryAttempts int
	EventEmitter  EventEmitter // Optional event emitter
}

// NewClient creates a new Jules API client
func NewClient(apiKey, baseURL string, timeout time.Duration, retryAttempts int) *Client {
	return &Client{
		APIKey:  apiKey,
		BaseURL: baseURL,
		HTTPClient: &http.Client{
			Timeout: timeout,
		},
		RetryAttempts: retryAttempts,
	}
}

// WithEventEmitter sets the event emitter for the client
func (c *Client) WithEventEmitter(emitter EventEmitter) *Client {
	c.EventEmitter = emitter
	return c
}
