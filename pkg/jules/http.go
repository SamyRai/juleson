package jules

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/url"
	"regexp"
	"strconv"
	"strings"
	"time"
)

var sensitiveLogValuePattern = regexp.MustCompile(`(?i)((?:api[_-]?key|token|auth|secret|credential)[^=\s]*=)[^&\s"']+`)

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
	httpReq.Header.Set("Accept", "application/json")
	if c.UserAgent != "" {
		httpReq.Header.Set("User-Agent", c.UserAgent)
	}

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
			backoff := retryDelay(resp, c.RetryBackoff, attempt)
			if err := sleepWithContext(req.Context(), backoff, c.sleep); err != nil {
				return nil, err
			}
		}

		attemptReq, cloneErr := cloneRequest(req)
		if cloneErr != nil {
			return nil, cloneErr
		}

		start := time.Now()
		resp, err = c.HTTPClient.Do(attemptReq)
		duration := time.Since(start)

		if c.debugLog && c.logger != nil {
			statusCode := 0
			if resp != nil {
				statusCode = resp.StatusCode
			}
			redactedURL := redactURL(attemptReq.URL)

			args := []any{
				slog.String("method", attemptReq.Method),
				slog.String("url", redactedURL),
				slog.Duration("duration", duration),
				slog.Int("attempt", attempt+1),
			}
			if resp != nil {
				args = append(args, slog.Int("status_code", statusCode))
			}
			if err != nil {
				args = append(args, slog.String("error", redactSensitiveLogValue(err.Error())))
			}

			c.logger.DebugContext(req.Context(), "Jules API request", args...)
		}

		// Success case (2xx status codes)
		if err == nil && resp.StatusCode >= 200 && resp.StatusCode < 300 {
			return resp, nil
		}

		// Client errors (4xx) should not be retried except for 429 (rate limit)
		if err == nil && resp.StatusCode >= 400 && resp.StatusCode < 500 {
			if resp.StatusCode != http.StatusTooManyRequests {
				break
			}
			if attempt < c.RetryAttempts {
				resp.Body.Close()
				continue
			}
			break
		}

		// Server errors (5xx) should be retried
		if err == nil && resp.StatusCode >= 500 {
			if attempt < c.RetryAttempts {
				resp.Body.Close()
				continue
			}
			break
		}

		// Network errors should be retried
		if err != nil && shouldRetryTransportError(req.Context(), err) && attempt < c.RetryAttempts {
			continue
		}

		// If we've exhausted retries, break
		if attempt == c.RetryAttempts {
			break
		}
	}

	// Handle final error state
	if err != nil {
		if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) || errors.Is(req.Context().Err(), context.Canceled) || errors.Is(req.Context().Err(), context.DeadlineExceeded) {
			return nil, req.Context().Err()
		}
		return nil, fmt.Errorf("request failed after %d attempts: %w", c.RetryAttempts+1, err)
	}

	// Handle HTTP error response
	if resp != nil && (resp.StatusCode < 200 || resp.StatusCode >= 300) {
		body, _ := io.ReadAll(resp.Body)
		resp.Body.Close()

		return nil, &APIError{
			Method:     req.Method,
			Path:       req.URL.EscapedPath(),
			StatusCode: resp.StatusCode,
			Message:    http.StatusText(resp.StatusCode),
			Body:       string(body),
			RetryAfter: parseRetryAfter(resp.Header.Get("Retry-After")),
		}
	}

	return resp, nil
}

func redactURL(u *url.URL) string {
	if u == nil {
		return ""
	}

	clone := *u
	q := clone.Query()
	for k := range q {
		if isSensitiveLogKey(k) {
			q.Set(k, "REDACTED")
		}
	}
	clone.RawQuery = q.Encode()
	return clone.String()
}

func isSensitiveLogKey(key string) bool {
	lowerKey := strings.ToLower(key)
	return strings.Contains(lowerKey, "key") ||
		strings.Contains(lowerKey, "token") ||
		strings.Contains(lowerKey, "auth") ||
		strings.Contains(lowerKey, "secret") ||
		strings.Contains(lowerKey, "credential")
}

func redactSensitiveLogValue(value string) string {
	return sensitiveLogValuePattern.ReplaceAllString(value, "${1}REDACTED")
}

func cloneRequest(req *http.Request) (*http.Request, error) {
	clone := req.Clone(req.Context())
	if req.Body != nil && req.GetBody != nil {
		body, err := req.GetBody()
		if err != nil {
			return nil, fmt.Errorf("failed to replay request body: %w", err)
		}
		clone.Body = body
	}
	return clone, nil
}

func shouldRetryTransportError(ctx context.Context, err error) bool {
	return err != nil && ctx.Err() == nil
}

func retryDelay(resp *http.Response, base time.Duration, attempt int) time.Duration {
	if resp != nil && resp.StatusCode == http.StatusTooManyRequests {
		if delay := parseRetryAfter(resp.Header.Get("Retry-After")); delay > 0 {
			return delay
		}
	}
	if base <= 0 {
		base = time.Second
	}
	return time.Duration(1<<uint(attempt-1)) * base
}

func sleepWithContext(ctx context.Context, delay time.Duration, sleep SleepFunc) error {
	if delay <= 0 {
		return ctx.Err()
	}
	if sleep == nil {
		sleep = func(d time.Duration) error {
			timer := time.NewTimer(d)
			defer timer.Stop()
			select {
			case <-ctx.Done():
				return ctx.Err()
			case <-timer.C:
				return nil
			}
		}
	}
	done := make(chan error, 1)
	go func() {
		done <- sleep(delay)
	}()
	select {
	case <-ctx.Done():
		return ctx.Err()
	case err := <-done:
		return err
	}
}

func parseRetryAfter(value string) time.Duration {
	if value == "" {
		return 0
	}
	if seconds, err := strconv.Atoi(value); err == nil {
		return time.Duration(seconds) * time.Second
	}
	if retryAt, err := http.ParseTime(value); err == nil {
		delay := time.Until(retryAt)
		if delay > 0 {
			return delay
		}
	}
	return 0
}
