package jules

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestErrorFunctions tests the error handling functions
func TestErrorFunctions(t *testing.T) {
	// Test Error method
	err := &APIError{
		StatusCode: 404,
		Message:    "Not found",
	}
	assert.Contains(t, err.Error(), "404")
	assert.Contains(t, err.Error(), "Not found")

	// Test Error method with body
	errWithBody := &APIError{
		StatusCode: 500,
		Message:    "Internal server error",
		Body:       "Something went wrong",
	}
	assert.Contains(t, errWithBody.Error(), "500")
	assert.Contains(t, errWithBody.Error(), "Internal server error")
	assert.Contains(t, errWithBody.Error(), "Something went wrong")

	// Test IsNotFound
	assert.True(t, err.IsNotFound())
	assert.False(t, (&APIError{StatusCode: 500}).IsNotFound())

	// Test IsBadRequest
	assert.True(t, (&APIError{StatusCode: 400}).IsBadRequest())
	assert.False(t, (&APIError{StatusCode: 500}).IsBadRequest())

	// Test IsUnauthorized
	assert.True(t, (&APIError{StatusCode: 401}).IsUnauthorized())
	assert.False(t, (&APIError{StatusCode: 500}).IsUnauthorized())

	// Test IsServerError
	assert.True(t, (&APIError{StatusCode: 500}).IsServerError())
	assert.True(t, (&APIError{StatusCode: 502}).IsServerError())
	assert.False(t, (&APIError{StatusCode: 400}).IsServerError())
}
