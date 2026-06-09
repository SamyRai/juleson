package sessions

import (
	"errors"
	"testing"

	"github.com/SamyRai/go-jules"
)

func TestBuildCreateSessionRequest(t *testing.T) {
	tests := []struct {
		assert  func(t *testing.T, req *jules.CreateSessionRequest)
		options CreateSessionRequestOptions
		name    string
	}{
		{
			name: "repoless",
			options: CreateSessionRequestOptions{
				Prompt:              "Draft a migration plan",
				NoSource:            true,
				Title:               "Migration",
				RequirePlanApproval: true,
				AutomationMode:      "AUTO_CREATE_PR",
			},
			assert: func(t *testing.T, req *jules.CreateSessionRequest) {
				if req.SourceContext != nil {
					t.Fatal("SourceContext should be omitted for repoless sessions")
				}
				if req.Prompt != "Draft a migration plan" || req.Title != "Migration" {
					t.Fatalf("unexpected request: %+v", req)
				}
				if !req.RequirePlanApproval || req.AutomationMode != jules.AutomationMode("AUTO_CREATE_PR") {
					t.Fatalf("options not preserved: %+v", req)
				}
			},
		},
		{
			name: "source backed with branch",
			options: CreateSessionRequestOptions{
				Prompt:         "Fix tests",
				Source:         "github/acme/widgets",
				StartingBranch: "main",
			},
			assert: func(t *testing.T, req *jules.CreateSessionRequest) {
				if req.SourceContext == nil {
					t.Fatal("SourceContext should be set")
				}
				if req.SourceContext.Source != "sources/github/acme/widgets" {
					t.Fatalf("source = %q", req.SourceContext.Source)
				}
				if req.SourceContext.GithubRepoContext == nil || req.SourceContext.GithubRepoContext.StartingBranch != "main" {
					t.Fatalf("branch context not preserved: %+v", req.SourceContext)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req, err := BuildCreateSessionRequest(tt.options)
			if err != nil {
				t.Fatalf("BuildCreateSessionRequest returned error: %v", err)
			}
			tt.assert(t, req)
		})
	}
}

func TestBuildCreateSessionRequestRejectsBranchWithoutSource(t *testing.T) {
	_, err := BuildCreateSessionRequest(CreateSessionRequestOptions{
		NoSource:       true,
		StartingBranch: "main",
	})
	if !errors.Is(err, ErrStartingBranchRequiresSource) {
		t.Fatalf("err = %v, want ErrStartingBranchRequiresSource", err)
	}
}

func TestNormalizeSourceID(t *testing.T) {
	tests := map[string]string{
		"github/acme/widgets":         "sources/github/acme/widgets",
		"sources/github/acme/widgets": "sources/github/acme/widgets",
		"  github/acme/widgets  ":     "sources/github/acme/widgets",
	}
	for input, want := range tests {
		if got := NormalizeSourceID(input); got != want {
			t.Fatalf("NormalizeSourceID(%q) = %q, want %q", input, got, want)
		}
	}
}
