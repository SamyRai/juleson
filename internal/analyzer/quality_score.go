package analyzer

// calculateMaintainabilityIndex calculates the maintainability index.
func (c *maintainabilityScorer) calculateMaintainabilityIndex(metrics *CodeQualityMetrics) float64 {
	baseScore := 100.0

	if metrics.CodeComplexity > 10 {
		baseScore -= (metrics.CodeComplexity - 10) * 2
	}
	baseScore -= float64(len(metrics.SecurityIssues)) * 5
	baseScore -= float64(len(metrics.CodeSmells)) * 2
	if metrics.TestCoverage < 80 {
		baseScore -= (80 - metrics.TestCoverage) * 0.5
	}

	if baseScore < 0 {
		return 0
	}
	if baseScore > 100 {
		return 100
	}
	return baseScore
}
