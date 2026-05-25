package jules

import (
	"fmt"
	"net/http"
	"time"
)

// APIError represents an error response from the Jules API
type APIError struct {
	Method     string
	Path       string
	StatusCode int
	Message    string
	Body       string
	RetryAfter time.Duration
}

// Error implements the error interface
func (e *APIError) Error() string {
	location := e.Path
	if location == "" {
		location = "request"
	}
	method := e.Method
	if method != "" {
		method += " "
	}
	if e.Body != "" {
		return fmt.Sprintf("Jules API error %s%s (HTTP %d): %s - %s", method, location, e.StatusCode, e.Message, e.Body)
	}
	return fmt.Sprintf("Jules API error %s%s (HTTP %d): %s", method, location, e.StatusCode, e.Message)
}

// IsNotFound returns true if the error is a 404 Not Found error
func (e *APIError) IsNotFound() bool {
	return e.StatusCode == http.StatusNotFound
}

// IsBadRequest returns true if the error is a 400 Bad Request error
func (e *APIError) IsBadRequest() bool {
	return e.StatusCode == http.StatusBadRequest
}

// IsUnauthorized returns true if the error is a 401 Unauthorized error
func (e *APIError) IsUnauthorized() bool {
	return e.StatusCode == http.StatusUnauthorized
}

// IsServerError returns true if the error is a 5xx server error
func (e *APIError) IsServerError() bool {
	return e.StatusCode >= 500 && e.StatusCode < 600
}
