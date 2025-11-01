package jules

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

// doRequestWithJSON performs an HTTP request with JSON payload and response handling
func (c *Client) doRequestWithJSON(ctx context.Context, method, url string, body interface{}, result interface{}) error {
	var reqBody io.Reader

	if body != nil {
		jsonData, err := json.Marshal(body)
		if err != nil {
			return fmt.Errorf("failed to marshal request body: %w", err)
		}
		reqBody = bytes.NewBuffer(jsonData)
	}

	httpReq, err := http.NewRequestWithContext(ctx, method, url, reqBody)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	if body != nil {
		httpReq.Header.Set("Content-Type", "application/json")
	}
	httpReq.Header.Set("X-Goog-Api-Key", c.APIKey)

	resp, err := c.doRequest(httpReq)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	// If result is nil, we don't need to decode the response
	if result == nil {
		return nil
	}

	if err := json.NewDecoder(resp.Body).Decode(result); err != nil {
		return fmt.Errorf("failed to decode response: %w", err)
	}

	return nil
}

// doRequest performs an HTTP request with retry logic and error handling
func (c *Client) doRequest(req *http.Request) (*http.Response, error) {
	var resp *http.Response
	var err error

	for attempt := 0; attempt <= c.RetryAttempts; attempt++ {
		if attempt > 0 {
			// Exponential backoff: 1s, 2s, 4s, 8s...
			backoff := time.Duration(1<<uint(attempt-1)) * time.Second
			time.Sleep(backoff)
		}

		resp, err = c.HTTPClient.Do(req)

		// Success case (2xx status codes)
		if err == nil && resp.StatusCode >= 200 && resp.StatusCode < 300 {
			return resp, nil
		}

		// Client errors (4xx) should not be retried except for 429 (rate limit)
		if err == nil && resp.StatusCode >= 400 && resp.StatusCode < 500 {
			if resp.StatusCode != http.StatusTooManyRequests {
				break
			}
			// For 429, we should retry with backoff
			resp.Body.Close()
			continue
		}

		// Server errors (5xx) should be retried
		if err == nil && resp.StatusCode >= 500 {
			resp.Body.Close()
			continue
		}

		// Network errors should be retried
		if err != nil && attempt < c.RetryAttempts {
			continue
		}

		// If we've exhausted retries, break
		if attempt == c.RetryAttempts {
			break
		}
	}

	// Handle final error state
	if err != nil {
		return nil, fmt.Errorf("request failed after %d attempts: %w", c.RetryAttempts+1, err)
	}

	// Handle HTTP error response
	if resp != nil && (resp.StatusCode < 200 || resp.StatusCode >= 300) {
		body, _ := io.ReadAll(resp.Body)
		resp.Body.Close()

		return nil, &APIError{
			StatusCode: resp.StatusCode,
			Message:    http.StatusText(resp.StatusCode),
			Body:       string(body),
		}
	}

	return resp, nil
}
