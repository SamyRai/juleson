package jules

import (
	"net/http"
	"time"
)

// Client represents a Jules API client
type Client struct {
	APIKey        string
	BaseURL       string
	HTTPClient    *http.Client
	RetryAttempts int
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
