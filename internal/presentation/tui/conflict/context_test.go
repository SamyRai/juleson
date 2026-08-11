package conflict

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestBuildContextPayload(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "conflict_test")
	require.NoError(t, err)
	defer func() { _ = os.RemoveAll(tmpDir) }()

	filePath := "test.txt"
	fullPath := filepath.Join(tmpDir, filePath)
	err = os.WriteFile(fullPath, []byte("local file content"), 0600)
	require.NoError(t, err)

	ctx := context.Background()

	t.Run("all options included", func(t *testing.T) {
		opts := &ResolutionOptions{
			IncludeLocalFile:   true,
			IncludePatchDiff:   true,
			IncludeCompilerOut: true,
			Guidance:           "keep my changes",
		}

		patch := "+ added line\n- removed line"
		payload, err := BuildContextPayload(ctx, tmpDir, filePath, patch, opts)

		require.NoError(t, err)
		assert.Contains(t, payload, "# Merge Conflict Resolution Request")
		assert.Contains(t, payload, "## Developer Instructions")
		assert.Contains(t, payload, "keep my changes")
		assert.Contains(t, payload, "## Current State of `test.txt`")
		assert.Contains(t, payload, "local file content")
		assert.Contains(t, payload, "## Failing Patch Diff")
		assert.Contains(t, payload, "+ added line")
		assert.Contains(t, payload, "## Compiler / Linter Errors")
	})

	t.Run("no options included", func(t *testing.T) {
		opts := &ResolutionOptions{
			IncludeLocalFile:   false,
			IncludePatchDiff:   false,
			IncludeCompilerOut: false,
			Guidance:           "",
		}

		patch := "+ added line\n- removed line"
		payload, err := BuildContextPayload(ctx, tmpDir, filePath, patch, opts)

		require.NoError(t, err)
		assert.Contains(t, payload, "# Merge Conflict Resolution Request")
		assert.NotContains(t, payload, "## Developer Instructions")
		assert.NotContains(t, payload, "## Current State of `test.txt`")
		assert.NotContains(t, payload, "## Failing Patch Diff")
		assert.NotContains(t, payload, "## Compiler / Linter Errors")
	})

	t.Run("file not found", func(t *testing.T) {
		opts := &ResolutionOptions{
			IncludeLocalFile: true,
		}

		payload, err := BuildContextPayload(ctx, tmpDir, "missing.txt", "", opts)

		require.NoError(t, err)
		assert.Contains(t, payload, "## Current State of `missing.txt`")
		assert.Contains(t, payload, "*(File does not exist locally)*")
	})
}
