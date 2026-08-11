package sessions

import (
	"context"
	"fmt"

	"github.com/SamyRai/juleson/internal/jules/workspace"
)

type PatchRequest struct {
	WorkingDir        string
	ActivityID        string
	ArtifactIndex     int
	DryRun            bool
	Confirm           bool
	AllowDirty        bool
	Force             bool
	CreateBackup      bool
	HasArtifactIndex  bool
	AllowBaseMismatch bool
}

type PatchPreparation struct {
	Options        *workspace.PatchApplicationOptions
	Blocker        string
	WorktreeStatus string
	DryRun         bool
}

func PreparePatchApplication(ctx context.Context, request PatchRequest) (*PatchPreparation, error) {
	dryRun := request.DryRun || !request.Confirm
	preparation := &PatchPreparation{
		DryRun: dryRun,
		Options: &workspace.PatchApplicationOptions{
			WorkingDir:        request.WorkingDir,
			DryRun:            dryRun,
			Force:             request.Force,
			CreateBackup:      request.CreateBackup,
			ActivityID:        request.ActivityID,
			ArtifactIndex:     request.ArtifactIndex,
			HasArtifactIndex:  request.HasArtifactIndex,
			AllowBaseMismatch: request.AllowBaseMismatch,
		},
	}

	if dryRun || request.AllowDirty {
		return preparation, nil
	}

	clean, status, err := workspace.IsGitWorkingTreeClean(ctx, request.WorkingDir)
	if err != nil {
		return preparation, err
	}
	if !clean {
		blocker := "target worktree has local changes; commit/stash them or set allow_dirty=true"
		if status != "" {
			blocker = blocker + ": " + status
		}
		preparation.Blocker = blocker
		preparation.WorktreeStatus = status
		return preparation, fmt.Errorf("%s", blocker)
	}

	return preparation, nil
}

func SessionChangesSummary(changes *workspace.SessionChanges) (string, int, int) {
	totalLinesAdded := 0
	totalLinesRemoved := 0
	for _, file := range changes.Files {
		totalLinesAdded += file.LinesAdded
		totalLinesRemoved += file.LinesRemoved
	}
	return fmt.Sprintf("%d patches affecting %d files (+%d -%d lines)",
		changes.TotalPatches, len(changes.Files), totalLinesAdded, totalLinesRemoved), totalLinesAdded, totalLinesRemoved
}
