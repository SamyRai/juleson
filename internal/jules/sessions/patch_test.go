package sessions

import (
	"context"
	"os"
	"os/exec"
	"testing"

	"github.com/SamyRai/juleson/internal/jules/workspace"
)

func TestPreparePatchApplicationDryRunSkipsDirtyCheck(t *testing.T) {
	preparation, err := PreparePatchApplication(context.Background(), PatchRequest{
		WorkingDir: "/not/a/git/repo",
		DryRun:     true,
	})
	if err != nil {
		t.Fatalf("PreparePatchApplication returned error: %v", err)
	}
	if !preparation.DryRun || !preparation.Options.DryRun {
		t.Fatalf("dry-run not propagated: %+v", preparation)
	}
}

func TestPreparePatchApplicationBlocksDirtyMutation(t *testing.T) {
	tmpDir := t.TempDir()
	cmd := exec.Command("git", "init")
	cmd.Dir = tmpDir
	if err := cmd.Run(); err != nil {
		t.Fatalf("git init: %v", err)
	}
	if err := os.WriteFile(tmpDir+"/dirty.txt", []byte("dirty"), 0600); err != nil {
		t.Fatalf("write dirty file: %v", err)
	}

	preparation, err := PreparePatchApplication(context.Background(), PatchRequest{
		WorkingDir: tmpDir,
		Confirm:    true,
	})
	if err == nil {
		t.Fatal("expected dirty worktree error")
	}
	if preparation == nil || preparation.Blocker == "" {
		t.Fatalf("expected blocker, got %+v", preparation)
	}
	if preparation.Options == nil || preparation.Options.WorkingDir != tmpDir {
		t.Fatalf("options not preserved: %+v", preparation)
	}
}

func TestPreparePatchApplicationPreservesPatchOptions(t *testing.T) {
	preparation, err := PreparePatchApplication(context.Background(), PatchRequest{
		WorkingDir:        "/repo",
		DryRun:            true,
		Force:             true,
		CreateBackup:      true,
		ActivityID:        "activity-1",
		ArtifactIndex:     2,
		HasArtifactIndex:  true,
		AllowBaseMismatch: true,
	})
	if err != nil {
		t.Fatalf("PreparePatchApplication returned error: %v", err)
	}
	options := preparation.Options
	if options.WorkingDir != "/repo" || !options.Force || !options.CreateBackup || options.ActivityID != "activity-1" {
		t.Fatalf("basic options not preserved: %+v", options)
	}
	if options.ArtifactIndex != 2 || !options.HasArtifactIndex || !options.AllowBaseMismatch {
		t.Fatalf("scoping options not preserved: %+v", options)
	}
}

func TestSessionChangesSummary(t *testing.T) {
	summary, added, removed := SessionChangesSummary(&workspace.SessionChanges{
		TotalPatches: 2,
		Files: []workspace.FileChange{
			{Path: "a.go", LinesAdded: 3, LinesRemoved: 1},
			{Path: "b.go", LinesAdded: 4, LinesRemoved: 2},
		},
	})
	if summary != "2 patches affecting 2 files (+7 -3 lines)" {
		t.Fatalf("summary = %q", summary)
	}
	if added != 7 || removed != 3 {
		t.Fatalf("added/removed = %d/%d", added, removed)
	}
}
