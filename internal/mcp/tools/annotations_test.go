package tools

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestToolAnnotationHelpers(t *testing.T) {
	readOnly := readOnlyOpenWorldTool("Read")
	require.NotNil(t, readOnly)
	assert.True(t, readOnly.ReadOnlyHint)
	require.NotNil(t, readOnly.OpenWorldHint)
	assert.True(t, *readOnly.OpenWorldHint)

	mutating := mutatingOpenWorldTool("Apply", true, false)
	require.NotNil(t, mutating)
	require.NotNil(t, mutating.DestructiveHint)
	assert.True(t, *mutating.DestructiveHint)
	assert.False(t, mutating.IdempotentHint)
}
