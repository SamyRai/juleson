package sessions

import (
	"context"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/SamyRai/go-jules"
	"github.com/jarcoal/httpmock"
)

func TestHandleResolutionResponse(t *testing.T) {
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	client := jules.NewClient("test-api-key", jules.WithBaseURL("https://jules.googleapis.com/v1alpha"), jules.WithRetryAttempts(0))
	ctx := context.Background()

	httpmock.RegisterResponder("GET", "https://jules.googleapis.com/v1alpha/sessions/session-1/activities?pageSize=10",
		func(req *http.Request) (*http.Response, error) {
			return httpmock.NewJsonResponse(200, jules.ActivitiesResponse{
				Activities: []jules.Activity{
					{
						ID:         "act-1",
						CreateTime: time.Now(),
						Artifacts: []jules.Artifact{
							{
								Media: &jules.Media{
									MimeType: "text/markdown",
									Data:     "# Resolved Patch\nFixed.",
								},
							},
							{
								ChangeSet: &jules.ChangeSet{
									GitPatch: &jules.GitPatch{
										UnidiffPatch: "+ new code",
									},
								},
							},
						},
					},
				},
			})
		})

	output := captureStdout(t, func() {
		err := handleResolutionResponse(ctx, client, "session-1")
		if err != nil {
			t.Fatalf("handleResolutionResponse failed: %v", err)
		}
	})

	if !strings.Contains(output, "--- Resolution Report ---") {
		t.Errorf("Missing resolution report in output: %s", output)
	}
	if !strings.Contains(output, "A new patch was created!") {
		t.Errorf("Missing patch message in output: %s", output)
	}
}
