package adapters

import (
	"testing"

	"github.com/SamyRai/juleson/internal/llm"
)

func TestParseTasksRejectsMalformedJSON(t *testing.T) {
	_, err := parseTasks(&llm.Response{Text: "not json"})
	if err == nil {
		t.Fatal("parseTasks() error = nil, want malformed JSON error")
	}
}

func TestParseTasksRejectsEmptyTaskList(t *testing.T) {
	_, err := parseTasks(&llm.Response{Text: `{"tasks":[]}`})
	if err == nil {
		t.Fatal("parseTasks() error = nil, want empty task list error")
	}
}

func TestParseDecisionRejectsUnknownDecisionType(t *testing.T) {
	_, err := parseDecision(&llm.Response{Text: `{"decision_type":"teleport","reasoning":"invalid"}`})
	if err == nil {
		t.Fatal("parseDecision() error = nil, want unknown decision type error")
	}
}
