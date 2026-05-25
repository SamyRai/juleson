package commands

import (
	"strings"
	"testing"

	"github.com/SamyRai/juleson/internal/orchestration"
)

func TestAgentExecuteRejectsInvalidStrictnessBeforeRuntime(t *testing.T) {
	runtimeInitialized := false
	cmd := newAgentExecuteCommand(func() (*orchestration.Runtime, error) {
		runtimeInitialized = true
		return nil, nil
	})
	cmd.SetArgs([]string{"inspect project safety", "--source", "test-source", "--strictness", "extreme"})

	err := cmd.Execute()
	if err == nil {
		t.Fatal("Execute() error = nil, want invalid strictness error")
	}
	if !strings.Contains(err.Error(), "invalid --strictness") {
		t.Fatalf("error = %q, want invalid strictness", err.Error())
	}
	if runtimeInitialized {
		t.Fatal("runtime initialized before strictness validation")
	}
}
