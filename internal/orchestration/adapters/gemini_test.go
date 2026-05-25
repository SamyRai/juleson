package adapters

import "testing"

func TestParseTasksRejectsMalformedJSON(t *testing.T) {
	_, err := parseTasks("not json")
	if err == nil {
		t.Fatal("parseTasks() error = nil, want malformed JSON error")
	}
}

func TestParseTasksRejectsEmptyTaskList(t *testing.T) {
	_, err := parseTasks(`{"tasks":[]}`)
	if err == nil {
		t.Fatal("parseTasks() error = nil, want empty task list error")
	}
}

func TestParseDecisionRejectsUnknownDecisionType(t *testing.T) {
	_, err := parseDecision(`{"decision_type":"teleport","reasoning":"invalid"}`)
	if err == nil {
		t.Fatal("parseDecision() error = nil, want unknown decision type error")
	}
}
