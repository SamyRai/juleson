package orchestrator

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCommandAvailable(t *testing.T) {
	// "go" or "ls" should be available
	assert.True(t, CommandAvailable("go"))

	// Non-existent command
	assert.False(t, CommandAvailable("some-random-command-that-doesnt-exist-12345"))
}

func TestShellCommandRunner_CombinedOutput(t *testing.T) {
	runner := shellCommandRunner{}
	ctx := context.Background()

	output, err := runner.CombinedOutput(ctx, "echo", "hello world")
	require.NoError(t, err)
	assert.Contains(t, output, "hello world")

	_, err = runner.CombinedOutput(ctx, "false")
	require.Error(t, err)
}
