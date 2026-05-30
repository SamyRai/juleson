package static

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/SamyRai/juleson/internal/codeintel"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAnalyzeFile(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "codeintel_static_test")
	require.NoError(t, err)
	defer os.RemoveAll(tmpDir)

	src := `package testpkg

func ComplexFunc(a, b int) {
	// High cyclomatic complexity
	if a > 0 {
		if b > 0 {
			a++
		} else {
			b++
		}
	} else if a < 0 {
		a--
	} else if a == 0 {
		a++
	} else {
		b++
	}
	
	for i := 0; i < 10; i++ {
		if i%2 == 0 {
			a += i
		} else if i%3 == 0 {
			a -= i
		}
	}

	if a == b || a > b && b < 10 || b == 0 {
		a = 0
	}
}

func UnusedVars() {
	var x int = 10 // Unused
	var y int = 20
	println(y)
}
`
	filePath := filepath.Join(tmpDir, "test.go")
	err = os.WriteFile(filePath, []byte(src), 0644)
	require.NoError(t, err)

	runner := NewRunner(nil)
	res, err := runner.AnalyzeFile(filePath, codeintel.IssueSeverityWarning)
	require.NoError(t, err)

	// Check if issues were found
	assert.NotEmpty(t, res.Issues)

	hasComplexityIssue := false

	for _, issue := range res.Issues {
		if issue.Category == codeintel.IssueCategoryComplexity {
			hasComplexityIssue = true
		}
	}

	assert.True(t, hasComplexityIssue, "Should detect high complexity")
}
