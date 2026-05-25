package automation

import (
	"testing"

	"google.golang.org/genai"
)

func aiResponse(text string) *genai.GenerateContentResponse {
	return &genai.GenerateContentResponse{
		Candidates: []*genai.Candidate{
			{
				Content: &genai.Content{
					Parts: []*genai.Part{{Text: text}},
				},
			},
		},
	}
}

func TestExtractAnalysisFromResponseParsesStructuredText(t *testing.T) {
	context := ExtractAnalysisFromResponse(aiResponse(`Languages: Go, TypeScript
Architecture: CLI and MCP services
Complexity: Medium
Current State: Stable
Issues:
- Large command files
- Mixed responsibilities
Opportunities:
- Extract presenters
- Add seams`))

	if got := context.Languages; len(got) != 2 || got[0] != "Go" || got[1] != "TypeScript" {
		t.Fatalf("unexpected languages: %#v", got)
	}
	if context.Architecture != "CLI and MCP services" {
		t.Fatalf("unexpected architecture: %q", context.Architecture)
	}
	if context.Complexity != "Medium" {
		t.Fatalf("unexpected complexity: %q", context.Complexity)
	}
	if context.CurrentState != "Stable" {
		t.Fatalf("unexpected current state: %q", context.CurrentState)
	}
	if len(context.Issues) != 2 {
		t.Fatalf("expected two issues, got %#v", context.Issues)
	}
	if len(context.Opportunities) != 2 {
		t.Fatalf("expected two opportunities, got %#v", context.Opportunities)
	}
}

func TestExtractTasksFromResponsePrefersJSONPlan(t *testing.T) {
	tasks := ExtractTasksFromResponse(aiResponse(`{
		"reasoning": "ordered by risk",
		"tasks": [
			{"name": "Extract presenter", "description": "Move formatting", "prompt": "Move formatting only"},
			{"name": "Add seam", "description": "Add interface", "prompt": "Add narrow interface"}
		]
	}`))

	if len(tasks) != 2 {
		t.Fatalf("expected two tasks, got %#v", tasks)
	}
	if tasks[0].Name != "Extract presenter" || tasks[0].Priority != 1 {
		t.Fatalf("unexpected first task: %#v", tasks[0])
	}
	if tasks[1].Rationale != "ordered by risk" {
		t.Fatalf("unexpected rationale: %q", tasks[1].Rationale)
	}
}

func TestExtractDecisionFromResponseClassifiesText(t *testing.T) {
	tests := map[string]string{
		"goal complete after final review": "complete",
		"needs user review first":          "review_needed",
		"continue execution":               "next_task",
	}

	for text, want := range tests {
		t.Run(want, func(t *testing.T) {
			decision := ExtractDecisionFromResponse(aiResponse(text))
			if decision.DecisionType != want {
				t.Fatalf("DecisionType = %q, want %q", decision.DecisionType, want)
			}
			if decision.Confidence != 0.8 {
				t.Fatalf("Confidence = %v, want 0.8", decision.Confidence)
			}
		})
	}
}
