package jules

import (
	"fmt"
	"net/http"
)

// APIError represents an error response from the Jules API
type APIError struct {
	StatusCode int
	Message    string
	Body       string
}

// Error implements the error interface
func (e *APIError) Error() string {
	if e.Body != "" {
		return fmt.Sprintf("Jules API error (HTTP %d): %s - %s", e.StatusCode, e.Message, e.Body)
	}
	return fmt.Sprintf("Jules API error (HTTP %d): %s", e.StatusCode, e.Message)
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
