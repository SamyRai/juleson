package review

import (
	"context"
	"fmt"
	"strings"

	"github.com/SamyRai/juleson/internal/agent"
)

// Reviewer performs code review on changes
type Reviewer interface {
	// Review analyzes changes and provides feedback
	Review(ctx context.Context, changes []agent.Change) (*agent.ReviewResult, error)
}

// Config holds reviewer configuration
type Config struct {
	Strictness       Strictness
	MinTestCoverage  float64
	SecurityScan     bool
	PerformanceCheck bool
	StyleCheck       bool
}

// Strictness level for reviews
type Strictness string

const (
	StrictnessLow    Strictness = "low"
	StrictnessMedium Strictness = "medium"
	StrictnessHigh   Strictness = "high"
)

// DefaultConfig returns default review configuration
func DefaultConfig() *Config {
	return &Config{
		Strictness:       StrictnessMedium,
		MinTestCoverage:  0.8,
		SecurityScan:     true,
		PerformanceCheck: true,
		StyleCheck:       true,
	}
}

// basicReviewer implements basic code review
type basicReviewer struct {
	config *Config
}

// NewReviewer creates a new code reviewer
func NewReviewer(config *Config) Reviewer {
	if config == nil {
		config = DefaultConfig()
	}
	return &basicReviewer{
		config: config,
	}
}

// Review analyzes changes and provides structured feedback
func (r *basicReviewer) Review(ctx context.Context, changes []agent.Change) (*agent.ReviewResult, error) {
	if changes == nil {
		return nil, fmt.Errorf("changes cannot be nil")
	}

	if len(changes) == 0 {
		return &agent.ReviewResult{
			Decision: agent.ReviewDecisionApprove,
			Comments: []agent.ReviewComment{},
			Score:    100.0,
			Summary:  "No changes to review",
			Approved: true,
		}, nil
	}

	var comments []agent.ReviewComment
	var reviewErrors []string

	// Analyze each change with error recovery
	for i, change := range changes {
		// Security checks
		if r.config.SecurityScan {
			securityComments := r.checkSecurity(change)
			comments = append(comments, securityComments...)
		}

		// Performance checks
		if r.config.PerformanceCheck {
			perfComments := r.checkPerformance(change)
			comments = append(comments, perfComments...)
		}

		// Style checks
		if r.config.StyleCheck {
			styleComments := r.checkStyle(change)
			comments = append(comments, styleComments...)
		}

		// Best practices
		bestPracticeComments := r.checkBestPractices(change)
		comments = append(comments, bestPracticeComments...)

		// Track any issues with individual changes
		if change.FilePath == "" {
			reviewErrors = append(reviewErrors, fmt.Sprintf("change %d: missing file path", i))
		}
	}

	// Calculate score and make decision
	score := r.calculateScore(comments)
	decision := r.makeDecision(comments, score)

	approved := decision == agent.ReviewDecisionApprove
	changesRequested := decision == agent.ReviewDecisionRequestChanges

	summary := r.generateSummary(comments, score)
	if len(reviewErrors) > 0 {
		summary += fmt.Sprintf(" Note: %d review warning(s) detected.", len(reviewErrors))
	}

	return &agent.ReviewResult{
		Decision:         decision,
		Comments:         comments,
		Score:            score,
		Summary:          summary,
		Approved:         approved,
		ChangesRequested: changesRequested,
	}, nil
}

// checkSecurity looks for common security issues
func (r *basicReviewer) checkSecurity(change agent.Change) []agent.ReviewComment {
	var comments []agent.ReviewComment

	patch := strings.ToLower(change.Patch)
	content := strings.ToLower(change.Description)

	// Check for hardcoded credentials
	if strings.Contains(patch, "password") && strings.Contains(patch, "=") {
		comments = append(comments, agent.ReviewComment{
			Severity:   agent.SeverityCritical,
			Category:   agent.ReviewCategorySecurity,
			Message:    "Potential hardcoded password detected",
			Suggestion: "Use environment variables or secure secret management for credentials",
			Example:    "password := os.Getenv(\"DB_PASSWORD\")",
		})
	}

	// Check for API keys
	if strings.Contains(patch, "api_key") || strings.Contains(patch, "apikey") {
		comments = append(comments, agent.ReviewComment{
			Severity:   agent.SeverityHigh,
			Category:   agent.ReviewCategorySecurity,
			Message:    "API key detected in code",
			Suggestion: "Store API keys in environment variables or secret manager",
			Example:    "apiKey := os.Getenv(\"API_KEY\")",
		})
	}

	// Check for SQL injection risks
	if strings.Contains(patch, "exec(") || strings.Contains(patch, "query(") {
		if !strings.Contains(patch, "prepare") && !strings.Contains(patch, "$1") {
			comments = append(comments, agent.ReviewComment{
				Severity:   agent.SeverityHigh,
				Category:   agent.ReviewCategorySecurity,
				Message:    "Potential SQL injection vulnerability",
				Suggestion: "Use parameterized queries or prepared statements",
				Example:    "db.Query(\"SELECT * FROM users WHERE id = $1\", userID)",
			})
		}
	}

	// Check for missing input validation
	if strings.Contains(content, "user input") || strings.Contains(patch, "request.") {
		if !strings.Contains(patch, "validate") && !strings.Contains(patch, "sanitize") {
			comments = append(comments, agent.ReviewComment{
				Severity:   agent.SeverityMedium,
				Category:   agent.ReviewCategorySecurity,
				Message:    "Consider adding input validation",
				Suggestion: "Validate and sanitize all user inputs",
			})
		}
	}

	return comments
}

// checkPerformance looks for performance issues
func (r *basicReviewer) checkPerformance(change agent.Change) []agent.ReviewComment {
	var comments []agent.ReviewComment

	patch := strings.ToLower(change.Patch)

	// Check for N+1 query patterns
	if strings.Contains(patch, "for") && strings.Contains(patch, "query") {
		comments = append(comments, agent.ReviewComment{
			Severity:   agent.SeverityMedium,
			Category:   agent.ReviewCategoryPerformance,
			Message:    "Potential N+1 query pattern detected",
			Suggestion: "Consider batch loading or eager loading to reduce database calls",
		})
	}

	// Check for missing indexes hint
	if strings.Contains(patch, "where") && !strings.Contains(patch, "index") {
		if strings.Contains(change.FilePath, "migration") || strings.Contains(change.FilePath, "schema") {
			comments = append(comments, agent.ReviewComment{
				Severity:   agent.SeverityLow,
				Category:   agent.ReviewCategoryPerformance,
				Message:    "Consider adding database indexes for queried columns",
				Suggestion: "Add indexes for columns used in WHERE clauses",
			})
		}
	}

	// Check for inefficient loops
	if strings.Contains(patch, "for range") && strings.Contains(patch, "append") {
		comments = append(comments, agent.ReviewComment{
			Severity:   agent.SeverityLow,
			Category:   agent.ReviewCategoryPerformance,
			Message:    "Consider pre-allocating slice capacity",
			Suggestion: "Use make() with capacity to avoid repeated allocations",
			Example:    "result := make([]Type, 0, len(input))",
		})
	}

	return comments
}

// checkStyle looks for style issues
func (r *basicReviewer) checkStyle(change agent.Change) []agent.ReviewComment {
	var comments []agent.ReviewComment

	// Check for long functions (basic heuristic)
	lines := strings.Count(change.Patch, "\n")
	if lines > 100 {
		comments = append(comments, agent.ReviewComment{
			Severity:   agent.SeverityLow,
			Category:   agent.ReviewCategoryStyle,
			Message:    "Function appears to be very long",
			Suggestion: "Consider breaking down into smaller, focused functions",
		})
	}

	// Check for TODO comments
	if strings.Contains(change.Patch, "TODO") || strings.Contains(change.Patch, "FIXME") {
		comments = append(comments, agent.ReviewComment{
			Severity:   agent.SeverityInfo,
			Category:   agent.ReviewCategoryStyle,
			Message:    "TODO/FIXME comment found",
			Suggestion: "Track TODOs in issue tracker and create follow-up tasks",
		})
	}

	return comments
}

// checkBestPractices looks for best practice violations
func (r *basicReviewer) checkBestPractices(change agent.Change) []agent.ReviewComment {
	var comments []agent.ReviewComment

	patch := strings.ToLower(change.Patch)

	// Check for error handling
	if strings.Contains(patch, "err :=") || strings.Contains(patch, "error") {
		if !strings.Contains(patch, "if err") && !strings.Contains(patch, "return err") {
			comments = append(comments, agent.ReviewComment{
				Severity:   agent.SeverityMedium,
				Category:   agent.ReviewCategoryBestPractice,
				Message:    "Error may not be properly handled",
				Suggestion: "Always check and handle errors appropriately",
				Example:    "if err != nil { return fmt.Errorf(\"operation failed: %w\", err) }",
			})
		}
	}

	// Check for context usage
	if strings.Contains(patch, "func ") && strings.Contains(patch, "(") {
		if !strings.Contains(patch, "context.context") && strings.Contains(change.FilePath, ".go") {
			// Only flag if it's a long-running operation
			if strings.Contains(patch, "http") || strings.Contains(patch, "database") || strings.Contains(patch, "query") {
				comments = append(comments, agent.ReviewComment{
					Severity:   agent.SeverityLow,
					Category:   agent.ReviewCategoryBestPractice,
					Message:    "Consider adding context parameter for cancellation support",
					Suggestion: "Add context.Context as first parameter for long-running operations",
					Example:    "func DoWork(ctx context.Context, params Params) error",
				})
			}
		}
	}

	// Check for test files
	if change.Type == agent.ChangeTypeAdd || change.Type == agent.ChangeTypeModify {
		if strings.HasSuffix(change.FilePath, ".go") && !strings.HasSuffix(change.FilePath, "_test.go") {
			if !strings.Contains(change.Description, "test") {
				comments = append(comments, agent.ReviewComment{
					Severity:   agent.SeverityMedium,
					Category:   agent.ReviewCategoryTesting,
					Message:    "New code should include tests",
					Suggestion: "Add unit tests for new functionality",
				})
			}
		}
	}

	return comments
}

// calculateScore computes an overall score based on comments
func (r *basicReviewer) calculateScore(comments []agent.ReviewComment) float64 {
	if len(comments) == 0 {
		return 100.0
	}

	// Deduct points based on severity
	score := 100.0
	for _, comment := range comments {
		switch comment.Severity {
		case agent.SeverityCritical:
			score -= 20.0
		case agent.SeverityHigh:
			score -= 10.0
		case agent.SeverityMedium:
			score -= 5.0
		case agent.SeverityLow:
			score -= 2.0
		case agent.SeverityInfo:
			score -= 0.5
		}
	}

	if score < 0 {
		score = 0
	}

	return score
}

// makeDecision determines the review decision based on comments and score
func (r *basicReviewer) makeDecision(comments []agent.ReviewComment, score float64) agent.ReviewDecision {
	// Check for critical issues
	for _, comment := range comments {
		if comment.Severity == agent.SeverityCritical {
			return agent.ReviewDecisionReject
		}
	}

	// Count high severity issues
	highSeverityCount := 0
	for _, comment := range comments {
		if comment.Severity == agent.SeverityHigh {
			highSeverityCount++
		}
	}

	// Apply strictness-based thresholds
	var rejectThreshold, changesThreshold float64
	switch r.config.Strictness {
	case StrictnessHigh:
		rejectThreshold = 60.0
		changesThreshold = 80.0
	case StrictnessMedium:
		rejectThreshold = 40.0
		changesThreshold = 70.0
	case StrictnessLow:
		rejectThreshold = 20.0
		changesThreshold = 60.0
	}

	if score < rejectThreshold || highSeverityCount > 3 {
		return agent.ReviewDecisionReject
	}

	if score < changesThreshold || highSeverityCount > 0 {
		return agent.ReviewDecisionRequestChanges
	}

	if len(comments) > 0 {
		return agent.ReviewDecisionComment
	}

	return agent.ReviewDecisionApprove
}

// generateSummary creates a human-readable summary
func (r *basicReviewer) generateSummary(comments []agent.ReviewComment, score float64) string {
	if len(comments) == 0 {
		return "✅ Code review passed with no issues found"
	}

	categoryCounts := make(map[agent.ReviewCategory]int)
	severityCounts := make(map[agent.Severity]int)

	for _, comment := range comments {
		categoryCounts[comment.Category]++
		severityCounts[comment.Severity]++
	}

	summary := fmt.Sprintf("Code review score: %.1f/100. Found %d issue(s): ", score, len(comments))

	if severityCounts[agent.SeverityCritical] > 0 {
		summary += fmt.Sprintf("%d critical, ", severityCounts[agent.SeverityCritical])
	}
	if severityCounts[agent.SeverityHigh] > 0 {
		summary += fmt.Sprintf("%d high, ", severityCounts[agent.SeverityHigh])
	}
	if severityCounts[agent.SeverityMedium] > 0 {
		summary += fmt.Sprintf("%d medium, ", severityCounts[agent.SeverityMedium])
	}
	if severityCounts[agent.SeverityLow] > 0 {
		summary += fmt.Sprintf("%d low, ", severityCounts[agent.SeverityLow])
	}

	summary = strings.TrimSuffix(summary, ", ")
	summary += "."

	return summary
}
