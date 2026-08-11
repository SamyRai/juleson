package sessions

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/SamyRai/go-jules"
	"github.com/SamyRai/juleson/internal/jules/workspace"
)

type ReviewRequest struct {
	SessionID        string
	WorkingDir       string
	ActivityID       string
	ArtifactIndex    int
	HasArtifactIndex bool
}

type ReviewNextAction struct {
	Label   string `json:"label"`
	Command string `json:"command,omitempty"`
	Reason  string `json:"reason,omitempty"`
}

//nolint:govet // Field order follows the public JSON review contract.
type PatchPreviewSummary struct {
	Files                   []workspace.FileChange `json:"files,omitempty"`
	SuggestedCommitMessages []string               `json:"suggested_commit_messages,omitempty"`
	Warnings                []string               `json:"warnings,omitempty"`
	BaseCommitMismatches    []string               `json:"base_commit_mismatches,omitempty"`
	Error                   string                 `json:"error,omitempty"`
	Summary                 string                 `json:"summary"`
	TotalPatches            int                    `json:"total_patches"`
	CanApply                bool                   `json:"can_apply"`
}

type WorktreeReview struct {
	WorkingDir string `json:"working_dir"`
	Status     string `json:"status,omitempty"`
	Error      string `json:"error,omitempty"`
	Clean      bool   `json:"clean"`
}

//nolint:govet // Field order follows the operator review flow and JSON contract.
type SessionReview struct {
	SessionID               string                       `json:"session_id"`
	Session                 jules.Session                `json:"session"`
	Plans                   []PlanSummary                `json:"plans"`
	LatestPlan              *PlanSummary                 `json:"latest_plan,omitempty"`
	Outputs                 []jules.Output               `json:"outputs"`
	ArtifactManifests       []workspace.ArtifactManifest `json:"artifact_manifests"`
	PatchPreview            PatchPreviewSummary          `json:"patch_preview"`
	Worktree                WorktreeReview               `json:"worktree"`
	Warnings                []string                     `json:"warnings,omitempty"`
	Blockers                []string                     `json:"blockers,omitempty"`
	VerificationSuggestions []string                     `json:"verification_suggestions,omitempty"`
	NextActions             []ReviewNextAction           `json:"next_actions"`
}

func BuildSessionReview(ctx context.Context, client *jules.Client, request ReviewRequest) (*SessionReview, error) {
	if request.WorkingDir == "" {
		wd, err := os.Getwd()
		if err != nil {
			return nil, fmt.Errorf("failed to get working directory: %w", err)
		}
		request.WorkingDir = wd
	}
	session, err := client.Sessions().Get(ctx, request.SessionID)
	if err != nil {
		return nil, fmt.Errorf("failed to get session: %w", err)
	}
	activities, err := client.Activities().ListAll(ctx, request.SessionID, 100)
	if err != nil {
		return nil, fmt.Errorf("failed to list activities: %w", err)
	}

	plans := ExtractPlanSummaries(activities)
	review := &SessionReview{
		SessionID:               session.ID,
		Session:                 *session,
		Plans:                   plans,
		LatestPlan:              LatestPlanSummary(plans),
		Outputs:                 append([]jules.Output(nil), session.Outputs...),
		ArtifactManifests:       buildArtifactManifests(activities),
		VerificationSuggestions: verificationSuggestions(request.WorkingDir),
	}

	patchOptions := &workspace.PatchApplicationOptions{
		WorkingDir:        request.WorkingDir,
		ActivityID:        request.ActivityID,
		ArtifactIndex:     request.ArtifactIndex,
		HasArtifactIndex:  request.HasArtifactIndex,
		AllowBaseMismatch: false,
	}
	changes, previewErr := workspace.PreviewSessionPatchesWithOptions(ctx, client, request.SessionID, patchOptions)
	review.PatchPreview = buildPatchPreview(changes, previewErr)
	review.Warnings = append(review.Warnings, review.PatchPreview.Warnings...)
	if len(review.PatchPreview.BaseCommitMismatches) > 0 {
		review.Blockers = append(review.Blockers, "patch base commit does not match target HEAD; inspect before applying")
	}
	if previewErr != nil {
		review.Blockers = append(review.Blockers, "patch dry-run preview failed")
	}

	review.Worktree = inspectWorktree(ctx, request.WorkingDir)
	if review.Worktree.Error != "" {
		review.Blockers = append(review.Blockers, "target worktree status could not be checked")
	} else if !review.Worktree.Clean {
		review.Blockers = append(review.Blockers, "target worktree has local changes; commit or stash them before applying")
	}

	review.NextActions = ReviewNextActions(review, request)
	return review, nil
}

func ReviewNextActions(review *SessionReview, request ReviewRequest) []ReviewNextAction {
	sessionID := review.SessionID
	if sessionID == "" {
		sessionID = request.SessionID
	}
	actions := []ReviewNextAction{}
	if review.Session.State == jules.SessionStateAwaitingPlanApproval {
		actions = append(actions, ReviewNextAction{
			Label:   "approve plan",
			Command: fmt.Sprintf("juleson sessions approve %s", sessionID),
			Reason:  "session is awaiting plan approval",
		})
	}
	if review.Session.State == jules.SessionStateAwaitingUserFeedback {
		actions = append(actions, ReviewNextAction{
			Label:   "send feedback",
			Command: fmt.Sprintf("juleson sessions message %s \"<message>\"", sessionID),
			Reason:  "session is awaiting user feedback",
		})
	}
	if review.PatchPreview.TotalPatches > 0 && review.PatchPreview.CanApply && len(review.Blockers) == 0 {
		actions = append(actions, ReviewNextAction{
			Label:   "apply patches",
			Command: applyCommand(sessionID, request),
			Reason:  "dry-run passed and target worktree is clean",
		})
	}
	actions = append(actions, ReviewNextAction{
		Label:   "watch session",
		Command: fmt.Sprintf("juleson sessions watch %s", sessionID),
		Reason:  "continue monitoring for completion or user action",
	})
	return actions
}

func buildPatchPreview(changes *workspace.SessionChanges, err error) PatchPreviewSummary {
	preview := PatchPreviewSummary{
		CanApply: err == nil,
	}
	if changes != nil {
		preview.TotalPatches = changes.TotalPatches
		preview.Files = changes.Files
		preview.SuggestedCommitMessages = changes.SuggestedCommitMessages
		preview.Warnings = changes.Warnings
		preview.BaseCommitMismatches = changes.BaseCommitMismatches
		preview.Summary, _, _ = SessionChangesSummary(changes)
	}
	if err != nil {
		preview.Error = err.Error()
		if preview.Summary == "" {
			preview.Summary = "patch dry-run preview failed"
		}
	}
	if preview.Summary == "" {
		preview.Summary = "0 patches affecting 0 files (+0 -0 lines)"
	}
	return preview
}

func buildArtifactManifests(activities []jules.Activity) []workspace.ArtifactManifest {
	var manifests []workspace.ArtifactManifest
	for activityIndex := range activities {
		activity := &activities[activityIndex]
		for i, artifact := range activity.Artifacts {
			manifests = append(manifests, workspace.BuildArtifactManifest(*activity, i, artifact))
		}
	}
	return manifests
}

func inspectWorktree(ctx context.Context, workingDir string) WorktreeReview {
	clean, status, err := workspace.IsGitWorkingTreeClean(ctx, workingDir)
	review := WorktreeReview{
		WorkingDir: workingDir,
		Clean:      clean,
		Status:     status,
	}
	if err != nil {
		review.Error = err.Error()
	}
	return review
}

func verificationSuggestions(workingDir string) []string {
	if workingDir == "" {
		return []string{"juleson sessions apply <session-id> <project-path>", "go test ./..."}
	}
	return []string{
		fmt.Sprintf("cd %s && go test ./...", shellQuote(workingDir)),
		fmt.Sprintf("cd %s && git diff --check", shellQuote(workingDir)),
	}
}

func applyCommand(sessionID string, request ReviewRequest) string {
	parts := []string{"juleson", "sessions", "apply", sessionID, shellQuote(request.WorkingDir)}
	if request.ActivityID != "" {
		parts = append(parts, "--activity-id", shellQuote(request.ActivityID))
	}
	if request.HasArtifactIndex {
		parts = append(parts, "--artifact-index", fmt.Sprintf("%d", request.ArtifactIndex))
	}
	parts = append(parts, "--confirm")
	return strings.Join(parts, " ")
}

func shellQuote(value string) string {
	if value == "" {
		return "\"\""
	}
	if strings.IndexFunc(value, func(r rune) bool {
		return r == ' ' || r == '\t' || r == '\n' || r == '"' || r == '\'' || r == '\\' || r == '$' || r == '`'
	}) == -1 {
		return value
	}
	return "'" + strings.ReplaceAll(value, "'", "'\\''") + "'"
}
