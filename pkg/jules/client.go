package jules

import (
	"log/slog"
	"net/http"
	"strings"
	"time"
)

const defaultBaseURL = "https://jules.googleapis.com/v1alpha"
const defaultUserAgent = "juleson-go-sdk"

// SleepFunc sleeps for the provided duration and may return early with an error.
type SleepFunc func(time.Duration) error

// ClientConfig contains the effective configuration for a Client.
type ClientConfig struct {
	APIKey        string
	BaseURL       string
	HTTPClient    *http.Client
	RetryAttempts int
	RetryBackoff  time.Duration
	UserAgent     string
	Sleep         SleepFunc
	Logger        *slog.Logger
	DebugLog      bool
}

// Client represents a Jules API client
type Client struct {
	APIKey        string
	BaseURL       string
	HTTPClient    *http.Client
	RetryAttempts int
	RetryBackoff  time.Duration
	UserAgent     string
	sleep         SleepFunc
	logger        *slog.Logger
	debugLog      bool
}

// ClientOption configures a Jules API client.
type ClientOption func(*Client)

// NewClient creates a new Jules API client.
func NewClient(apiKey string, options ...ClientOption) *Client {
	client := &Client{
		APIKey:        apiKey,
		BaseURL:       defaultBaseURL,
		HTTPClient:    &http.Client{Timeout: 30 * time.Second},
		RetryAttempts: 3,
		RetryBackoff:  time.Second,
		UserAgent:     defaultUserAgent,
	}

	for _, option := range options {
		if option != nil {
			option(client)
		}
	}

	if client.HTTPClient == nil {
		client.HTTPClient = &http.Client{Timeout: 30 * time.Second}
	}
	if client.BaseURL == "" {
		client.BaseURL = defaultBaseURL
	}
	client.BaseURL = strings.TrimRight(client.BaseURL, "/")
	if client.RetryBackoff <= 0 {
		client.RetryBackoff = time.Second
	}
	if client.UserAgent == "" {
		client.UserAgent = defaultUserAgent
	}
	if client.sleep == nil {
		client.sleep = func(d time.Duration) error {
			time.Sleep(d)
			return nil
		}
	}

	return client
}

// WithBaseURL sets the Jules API base URL.
func WithBaseURL(baseURL string) ClientOption {
	return func(c *Client) {
		c.BaseURL = baseURL
	}
}

// WithTimeout sets the HTTP client timeout.
func WithTimeout(timeout time.Duration) ClientOption {
	return func(c *Client) {
		if c.HTTPClient == nil {
			c.HTTPClient = &http.Client{}
		}
		c.HTTPClient.Timeout = timeout
	}
}

// WithRetryAttempts sets the number of retry attempts for retryable requests.
func WithRetryAttempts(retryAttempts int) ClientOption {
	return func(c *Client) {
		c.RetryAttempts = retryAttempts
	}
}

// WithHTTPClient sets the HTTP client used for requests.
func WithHTTPClient(httpClient *http.Client) ClientOption {
	return func(c *Client) {
		c.HTTPClient = httpClient
	}
}

// WithRetryBackoff sets the base retry backoff duration.
func WithRetryBackoff(backoff time.Duration) ClientOption {
	return func(c *Client) {
		c.RetryBackoff = backoff
	}
}

// WithUserAgent sets the User-Agent header.
func WithUserAgent(userAgent string) ClientOption {
	return func(c *Client) {
		c.UserAgent = userAgent
	}
}

// WithLogger sets the logger for the client.
func WithLogger(logger *slog.Logger) ClientOption {
	return func(c *Client) {
		c.logger = logger
	}
}

// WithDebugLog enables or disables debug logging.
func WithDebugLog(debugLog bool) ClientOption {
	return func(c *Client) {
		c.debugLog = debugLog
	}
}

// WithSleep sets the sleep function used between retries. It is primarily
// intended for tests.
func WithSleep(sleep SleepFunc) ClientOption {
	return func(c *Client) {
		c.sleep = sleep
	}
}

// Config returns the effective client configuration.
func (c *Client) Config() ClientConfig {
	return ClientConfig{
		APIKey:        c.APIKey,
		BaseURL:       c.BaseURL,
		HTTPClient:    c.HTTPClient,
		RetryAttempts: c.RetryAttempts,
		RetryBackoff:  c.RetryBackoff,
		UserAgent:     c.UserAgent,
		Sleep:         c.sleep,
		Logger:        c.logger,
		DebugLog:      c.debugLog,
	}
}
