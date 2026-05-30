package review

import (
	"context"
	"testing"

	"github.com/SamyRai/juleson/internal/agent"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewReviewer(t *testing.T) {
	// Test with nil config
	r := NewReviewer(nil)
	require.NotNil(t, r)

	// Test with custom config
	cfg := &Config{
		Strictness:   StrictnessHigh,
		SecurityScan: false,
	}
	r = NewReviewer(cfg)
	require.NotNil(t, r)
}

func TestReviewEmptyOrNil(t *testing.T) {
	r := NewReviewer(nil)
	ctx := context.Background()

	// Test nil changes
	_, err := r.Review(ctx, nil)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "changes cannot be nil")

	// Test empty changes
	res, err := r.Review(ctx, []agent.Change{})
	require.NoError(t, err)
	assert.True(t, res.Approved)
	assert.Equal(t, 100.0, res.Score)
	assert.Empty(t, res.Comments)
}

func TestReviewSecurityChecks(t *testing.T) {
	r := NewReviewer(&Config{SecurityScan: true})
	ctx := context.Background()

	cases := []struct {
		name          string
		patch         string
		desc          string
		expectedIssue string
	}{
		{
			name:          "hardcoded password",
			patch:         `db_password = "mysecretpassword"`,
			desc:          "adding db password",
			expectedIssue: "Potential hardcoded password detected",
		},
		{
			name:          "api key",
			patch:         `var apiKey = "AIzaSy..."`,
			desc:          "adding api key",
			expectedIssue: "API key detected in code",
		},
		{
			name:          "sql injection",
			patch:         `db.Exec("SELECT * FROM users WHERE id = " + id)`,
			desc:          "querying user",
			expectedIssue: "Potential SQL injection vulnerability",
		},
		{
			name:          "missing input validation",
			patch:         `user := request.Body`,
			desc:          "handling user input",
			expectedIssue: "Consider adding input validation",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			res, err := r.Review(ctx, []agent.Change{
				{FilePath: "main.go", Patch: tc.patch, Description: tc.desc},
			})
			require.NoError(t, err)

			found := false
			for _, c := range res.Comments {
				if c.Message == tc.expectedIssue {
					found = true
					break
				}
			}
			assert.True(t, found, "Expected to find issue: %s", tc.expectedIssue)
		})
	}
}

func TestReviewPerformanceChecks(t *testing.T) {
	r := NewReviewer(&Config{PerformanceCheck: true})
	ctx := context.Background()

	cases := []struct {
		name          string
		filepath      string
		patch         string
		expectedIssue string
	}{
		{
			name:          "N+1 query",
			filepath:      "main.go",
			patch:         `for _, u := range users { db.Query("...") }`,
			expectedIssue: "Potential N+1 query pattern detected",
		},
		{
			name:          "missing index hint",
			filepath:      "schema.sql",
			patch:         `SELECT * FROM items WHERE status = 'active'`,
			expectedIssue: "Consider adding database indexes for queried columns",
		},
		{
			name:          "inefficient append",
			filepath:      "main.go",
			patch:         `var res []int; for range items { res = append(res, 1) }`,
			expectedIssue: "Consider pre-allocating slice capacity",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			res, err := r.Review(ctx, []agent.Change{
				{FilePath: tc.filepath, Patch: tc.patch},
			})
			require.NoError(t, err)

			found := false
			for _, c := range res.Comments {
				if c.Message == tc.expectedIssue {
					found = true
					break
				}
			}
			assert.True(t, found, "Expected to find issue: %s", tc.expectedIssue)
		})
	}
}

func TestReviewStyleChecks(t *testing.T) {
	r := NewReviewer(&Config{StyleCheck: true})
	ctx := context.Background()

	longPatch := ""
	for i := 0; i < 105; i++ {
		longPatch += "fmt.Println()\n"
	}

	cases := []struct {
		name          string
		patch         string
		expectedIssue string
	}{
		{
			name:          "long function",
			patch:         longPatch,
			expectedIssue: "Function appears to be very long",
		},
		{
			name:          "todo comment",
			patch:         `// TODO: fix this later`,
			expectedIssue: "TODO/FIXME comment found",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			res, err := r.Review(ctx, []agent.Change{
				{FilePath: "main.go", Patch: tc.patch},
			})
			require.NoError(t, err)

			found := false
			for _, c := range res.Comments {
				if c.Message == tc.expectedIssue {
					found = true
					break
				}
			}
			assert.True(t, found, "Expected to find issue: %s", tc.expectedIssue)
		})
	}
}

func TestReviewBestPractices(t *testing.T) {
	r := NewReviewer(&Config{
		SecurityScan: false, PerformanceCheck: false, StyleCheck: false, // Isolate
	})
	ctx := context.Background()

	cases := []struct {
		name          string
		change        agent.Change
		expectedIssue string
	}{
		{
			name: "unhandled error",
			change: agent.Change{
				FilePath: "main.go",
				Patch:    `err := doSomething()`, // no if err != nil
			},
			expectedIssue: "Error may not be properly handled",
		},
		{
			name: "missing context",
			change: agent.Change{
				FilePath: "main.go",
				Patch:    `func makeRequest(url string) { http.Get(url) }`,
			},
			expectedIssue: "Consider adding context parameter for cancellation support",
		},
		{
			name: "missing test",
			change: agent.Change{
				Type:        agent.ChangeTypeAdd,
				FilePath:    "logic.go",
				Patch:       `func Calc() {}`,
				Description: "new logic", // no test in description
			},
			expectedIssue: "New code should include tests",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			res, err := r.Review(ctx, []agent.Change{tc.change})
			require.NoError(t, err)

			found := false
			for _, c := range res.Comments {
				if c.Message == tc.expectedIssue {
					found = true
					break
				}
			}
			assert.True(t, found, "Expected to find issue: %s", tc.expectedIssue)
		})
	}
}

func TestMakeDecision(t *testing.T) {
	ctx := context.Background()

	// High strictness config
	rHigh := NewReviewer(&Config{Strictness: StrictnessHigh, SecurityScan: true, StyleCheck: true})

	// Create a critical security issue which should cause rejection
	res, err := rHigh.Review(ctx, []agent.Change{
		{FilePath: "main.go", Patch: `password = "secret"`}, // SeverityCritical
	})
	require.NoError(t, err)
	assert.Equal(t, agent.ReviewDecisionReject, res.Decision)
	assert.False(t, res.Approved)

	// Low strictness, minor issue
	rLow := NewReviewer(&Config{Strictness: StrictnessLow, SecurityScan: false, StyleCheck: true, PerformanceCheck: false})
	res, err = rLow.Review(ctx, []agent.Change{
		{FilePath: "main.go", Patch: `// TODO: fix`}, // SeverityInfo
	})
	require.NoError(t, err)
	assert.Equal(t, agent.ReviewDecisionComment, res.Decision)
}

func TestGenerateSummary(t *testing.T) {
	r := NewReviewer(DefaultConfig())
	ctx := context.Background()

	res, err := r.Review(ctx, []agent.Change{
		{FilePath: "main.go", Patch: `password = "secret" // TODO: fix`},
	})
	require.NoError(t, err)

	assert.Contains(t, res.Summary, "critical")
}
