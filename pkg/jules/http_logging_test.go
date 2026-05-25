package jules

import (
	"bytes"
	"context"
	"errors"
	"log/slog"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/jarcoal/httpmock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestClientLogging_DisabledByDefault(t *testing.T) {
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	httpmock.RegisterResponder("GET", "https://jules.googleapis.com/v1alpha/sessions/123",
		httpmock.NewStringResponder(200, `{}`))

	var logBuf bytes.Buffer
	logger := slog.New(slog.NewTextHandler(&logBuf, &slog.HandlerOptions{Level: slog.LevelDebug}))

	client := NewClient("test-key", WithLogger(logger)) // Note: DebugLog is false by default
	// Ensure the HTTP client uses httpmock
	client.HTTPClient = &http.Client{}

	req, _ := http.NewRequestWithContext(context.Background(), "GET", "https://jules.googleapis.com/v1alpha/sessions/123", nil)
	_, err := client.doRequest(req)
	require.NoError(t, err)

	assert.Empty(t, logBuf.String(), "Expected no logs to be written when debugLog is false")
}

func TestClientLogging_EnabledLogsRequestData(t *testing.T) {
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	httpmock.RegisterResponder("GET", "https://jules.googleapis.com/v1alpha/sessions/123",
		httpmock.NewStringResponder(200, `{}`))

	var logBuf bytes.Buffer
	logger := slog.New(slog.NewTextHandler(&logBuf, &slog.HandlerOptions{Level: slog.LevelDebug}))

	client := NewClient("test-key", WithLogger(logger), WithDebugLog(true))
	client.HTTPClient = &http.Client{}

	req, _ := http.NewRequestWithContext(context.Background(), "GET", "https://jules.googleapis.com/v1alpha/sessions/123", nil)
	_, err := client.doRequest(req)
	require.NoError(t, err)

	logOutput := logBuf.String()
	assert.NotEmpty(t, logOutput, "Expected logs to be written when debugLog is true")
	assert.Contains(t, logOutput, "method=GET")
	assert.Contains(t, logOutput, "url=https://jules.googleapis.com/v1alpha/sessions/123")
	assert.Contains(t, logOutput, "status_code=200")
	assert.Contains(t, logOutput, "attempt=1")
	assert.Contains(t, logOutput, "duration=")
}

func TestClientLogging_RedactsSensitiveQueryParams(t *testing.T) {
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	httpmock.RegisterResponder("GET", "=~^https://jules.googleapis.com/v1alpha/sessions.*",
		httpmock.NewStringResponder(200, `{}`))

	var logBuf bytes.Buffer
	logger := slog.New(slog.NewTextHandler(&logBuf, &slog.HandlerOptions{Level: slog.LevelDebug}))

	client := NewClient("test-key", WithLogger(logger), WithDebugLog(true))
	client.HTTPClient = &http.Client{}

	req, _ := http.NewRequestWithContext(context.Background(), "GET", "https://jules.googleapis.com/v1alpha/sessions?api_key=secret1&token=secret2&auth_token=secret3&credential=secret4&safe_param=hello", nil)
	_, err := client.doRequest(req)
	require.NoError(t, err)

	logOutput := logBuf.String()
	assert.NotContains(t, logOutput, "secret1", "Expected api_key to be redacted")
	assert.NotContains(t, logOutput, "secret2", "Expected token to be redacted")
	assert.NotContains(t, logOutput, "secret3", "Expected auth_token to be redacted")
	assert.NotContains(t, logOutput, "secret4", "Expected credential to be redacted")

	assert.Contains(t, logOutput, "api_key=REDACTED")
	assert.Contains(t, logOutput, "token=REDACTED")
	assert.Contains(t, logOutput, "auth_token=REDACTED")
	assert.Contains(t, logOutput, "credential=REDACTED")
	assert.Contains(t, logOutput, "safe_param=hello", "Expected non-sensitive params to remain unredacted")
}

func TestClientLogging_LogsRetriesAndErrors(t *testing.T) {
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	callCount := 0
	httpmock.RegisterResponder("GET", "https://jules.googleapis.com/v1alpha/sessions/123",
		func(req *http.Request) (*http.Response, error) {
			callCount++
			if callCount < 2 {
				return httpmock.NewStringResponse(500, "Internal Server Error"), nil
			}
			return httpmock.NewStringResponse(200, `{}`), nil
		})

	var logBuf bytes.Buffer
	logger := slog.New(slog.NewTextHandler(&logBuf, &slog.HandlerOptions{Level: slog.LevelDebug}))

	client := NewClient("test-key",
		WithLogger(logger),
		WithDebugLog(true),
		WithRetryAttempts(2),
		WithRetryBackoff(1*time.Millisecond))
	client.HTTPClient = &http.Client{}

	req, _ := http.NewRequestWithContext(context.Background(), "GET", "https://jules.googleapis.com/v1alpha/sessions/123", nil)
	_, err := client.doRequest(req)
	require.NoError(t, err)

	logOutput := logBuf.String()

	// Should see attempt 1 failing
	assert.Contains(t, logOutput, "attempt=1")
	assert.Contains(t, logOutput, "status_code=500")

	// Should see attempt 2 succeeding
	assert.Contains(t, logOutput, "attempt=2")
	assert.Contains(t, logOutput, "status_code=200")

	// Split lines to verify order
	lines := strings.Split(strings.TrimSpace(logOutput), "\n")
	assert.Len(t, lines, 2)
}

func TestClientLogging_RedactsSensitiveErrorDetails(t *testing.T) {
	var logBuf bytes.Buffer
	logger := slog.New(slog.NewTextHandler(&logBuf, &slog.HandlerOptions{Level: slog.LevelDebug}))

	client := NewClient("test-key",
		WithHTTPClient(&http.Client{
			Transport: roundTripFunc(func(*http.Request) (*http.Response, error) {
				return nil, errors.New(`Get "https://jules.googleapis.com/v1alpha/sessions?api_key=secret1&token=secret2": failed`)
			}),
		}),
		WithLogger(logger),
		WithDebugLog(true),
		WithRetryAttempts(0))

	req, _ := http.NewRequestWithContext(context.Background(), "GET", "https://jules.googleapis.com/v1alpha/sessions?api_key=secret1&token=secret2", nil)
	_, err := client.doRequest(req)
	require.Error(t, err)

	logOutput := logBuf.String()
	assert.NotContains(t, logOutput, "secret1")
	assert.NotContains(t, logOutput, "secret2")
	assert.Contains(t, logOutput, "api_key=REDACTED")
	assert.Contains(t, logOutput, "token=REDACTED")
}

type roundTripFunc func(*http.Request) (*http.Response, error)

func (f roundTripFunc) RoundTrip(req *http.Request) (*http.Response, error) {
	return f(req)
}
