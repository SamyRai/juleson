package analyzer

import "testing"

func TestCodeQualityAnalyzerUsesDedicatedCollaborators(t *testing.T) {
	analyzer := NewCodeQualityAnalyzer()

	if analyzer.coverage == nil {
		t.Fatal("coverage analyzer is nil")
	}
	if analyzer.complexity == nil {
		t.Fatal("complexity analyzer is nil")
	}
	if analyzer.security == nil {
		t.Fatal("security analyzer is nil")
	}
	if analyzer.smells == nil {
		t.Fatal("smell analyzer is nil")
	}
	if analyzer.scorer == nil {
		t.Fatal("maintainability scorer is nil")
	}
}

func TestMaintainabilityScorerBoundsAndPenalties(t *testing.T) {
	scorer := &maintainabilityScorer{}

	tests := []struct {
		name    string
		metrics *CodeQualityMetrics
		want    float64
	}{
		{
			name:    "perfect score",
			metrics: &CodeQualityMetrics{TestCoverage: 100, CodeComplexity: 5},
			want:    100,
		},
		{
			name: "penalizes complexity security smells and coverage",
			metrics: &CodeQualityMetrics{
				TestCoverage:   60,
				CodeComplexity: 15,
				SecurityIssues: []SecurityIssue{{Severity: "HIGH"}},
				CodeSmells:     []CodeSmell{{Severity: "LOW"}, {Severity: "LOW"}},
			},
			want: 71,
		},
		{
			name: "lower bound",
			metrics: &CodeQualityMetrics{
				TestCoverage:   0,
				CodeComplexity: 100,
				SecurityIssues: make([]SecurityIssue, 10),
				CodeSmells:     make([]CodeSmell, 20),
			},
			want: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := scorer.calculateMaintainabilityIndex(tt.metrics); got != tt.want {
				t.Fatalf("score = %v, want %v", got, tt.want)
			}
		})
	}
}
