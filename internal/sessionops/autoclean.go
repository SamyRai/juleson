package sessionops

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/SamyRai/go-jules"
)

// VerifySessionMerged checks if a session's patch is already merged into its repository
// by cloning the repository to a tmpfs and running git apply --check --reverse.
func VerifySessionMerged(ctx context.Context, client *jules.Client, sessionID string, sourceContext *jules.SourceContext) (bool, error) {
	if sourceContext == nil || !strings.HasPrefix(sourceContext.Source, "sources/github/") {
		return false, fmt.Errorf("unsupported source format")
	}

	ownerRepo := strings.TrimPrefix(sourceContext.Source, "sources/github/")
	gitURL := fmt.Sprintf("git@github.com:%s.git", ownerRepo)

	// Create tmpfs for shallow clone
	tmpDir, err := os.MkdirTemp("", fmt.Sprintf("jules-autoclean-%s-*", sessionID))
	if err != nil {
		return false, fmt.Errorf("failed to create tmpfs: %w", err)
	}
	defer os.RemoveAll(tmpDir)

	cloneCmd := exec.CommandContext(ctx, "git", "clone", "--depth=1", "--filter=blob:none", gitURL, ".")
	cloneCmd.Dir = tmpDir
	if out, err := cloneCmd.CombinedOutput(); err != nil {
		return false, fmt.Errorf("git clone failed: %v\n%s", err, out)
	}

	// Fetch the raw patch
	var rawPatch string
	activitiesResp, err := client.Activities().List(ctx, sessionID, &jules.ListActivitiesOptions{PageSize: 100})
	if err == nil {
		for i := len(activitiesResp.Activities) - 1; i >= 0; i-- {
			for _, artifact := range activitiesResp.Activities[i].Artifacts {
				if artifact.ChangeSet != nil && artifact.ChangeSet.GitPatch != nil {
					rawPatch = artifact.ChangeSet.GitPatch.UnidiffPatch
					break
				}
			}
			if rawPatch != "" {
				break
			}
		}
	}

	if rawPatch == "" {
		return false, fmt.Errorf("no patch artifact found for session")
	}

	// Write patch to tmpfs
	patchFile := filepath.Join(tmpDir, "session.patch")
	if err := os.WriteFile(patchFile, []byte(rawPatch), 0644); err != nil {
		return false, fmt.Errorf("failed to write patch to tmpfs: %w", err)
	}

	// Verify if patch is applied exactly using git apply --check --reverse
	applyCmd := exec.CommandContext(ctx, "git", "apply", "--check", "--reverse", "session.patch")
	applyCmd.Dir = tmpDir

	if err := applyCmd.Run(); err == nil {
		return true, nil // It perfectly reverses, meaning it is applied
	}

	return false, nil
}
