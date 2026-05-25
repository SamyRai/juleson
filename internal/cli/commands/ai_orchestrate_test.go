package commands

import (
	"testing"

	"github.com/SamyRai/juleson/internal/orchestration/domain"
)

func TestAIWorkflowOptionsFromFlags(t *testing.T) {
	options := aiWorkflowOptionsFromFlags(7, true)

	if options.MaxIterations != 7 {
		t.Fatalf("max iterations = %d, want 7", options.MaxIterations)
	}
	if !options.ApprovalPolicy.AutoApprove {
		t.Fatal("auto-approve policy was not enabled")
	}
	if options.ApprovalPolicy.RequirePlanApproval {
		t.Fatal("plan approval should not be required when auto-approve is enabled")
	}
}

func TestAIWorkflowOptionsFromFlagsDefaultsToPlanApproval(t *testing.T) {
	options := aiWorkflowOptionsFromFlags(0, false)

	if options.ApprovalPolicy != (domain.ApprovalPolicy{RequirePlanApproval: true}) {
		t.Fatalf("approval policy = %+v, want plan approval required", options.ApprovalPolicy)
	}
}
