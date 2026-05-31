package actions

import "testing"

func TestParseWorkflowDispatchInputs(t *testing.T) {
	got := parseWorkflowDispatchInputs([]string{
		"env=prod",
		"message=hello=world",
		"invalid",
		"empty=",
	})

	if len(got) != 3 {
		t.Fatalf("expected three parsed inputs, got %#v", got)
	}
	if got["env"] != "prod" {
		t.Fatalf("env = %#v, want prod", got["env"])
	}
	if got["message"] != "hello=world" {
		t.Fatalf("message = %#v, want hello=world", got["message"])
	}
	if got["empty"] != "" {
		t.Fatalf("empty = %#v, want empty string", got["empty"])
	}
	if _, ok := got["invalid"]; ok {
		t.Fatalf("invalid input should be ignored: %#v", got)
	}
}
