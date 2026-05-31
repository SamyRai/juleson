package sessions

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/SamyRai/juleson/internal/presentation/cli/core"
	"os"
	"strings"
	"time"

	"github.com/SamyRai/juleson/internal/config"
	julessessions "github.com/SamyRai/juleson/internal/jules/sessions"
)

type ReviewSessionOptions struct {
	ActivityID       string
	ArtifactIndex    int
	HasArtifactIndex bool
	JSON             bool
}

func showSessionPlans(cfg *config.Config, sessionID string, latestOnly, jsonOutput bool) error {
	julesClient := core.NewJulesClient(cfg)
	activities, err := julesClient.Activities().ListAll(context.Background(), sessionID, 100)
	if err != nil {
		return fmt.Errorf("failed to list activities: %w", err)
	}
	plans := julessessions.ExtractPlanSummaries(activities)
	if latestOnly {
		if latest := julessessions.LatestPlanSummary(plans); latest != nil {
			plans = []julessessions.PlanSummary{*latest}
		} else {
			plans = []julessessions.PlanSummary{}
		}
	}
	if jsonOutput {
		return printJSON(map[string]any{
			"session_id":  sessionID,
			"plans":       plans,
			"total_count": len(plans),
		})
	}
	printPlanSummaries(sessionID, plans)
	return nil
}

func reviewSession(cfg *config.Config, sessionID, projectPath string, options ReviewSessionOptions) error {
	julesClient := core.NewJulesClient(cfg)
	review, err := julessessions.BuildSessionReview(context.Background(), julesClient, julessessions.ReviewRequest{
		SessionID:        sessionID,
		WorkingDir:       projectPath,
		ActivityID:       options.ActivityID,
		ArtifactIndex:    options.ArtifactIndex,
		HasArtifactIndex: options.HasArtifactIndex,
	})
	if err != nil {
		return err
	}
	if options.JSON {
		return printJSON(review)
	}
	printSessionReview(review)
	return nil
}

func printPlanSummaries(sessionID string, plans []julessessions.PlanSummary) {
	fmt.Printf("Plans for session %s\n", sessionID)
	fmt.Println(strings.Repeat("=", 60))
	if len(plans) == 0 {
		fmt.Println("No generated plans found.")
		fmt.Println()
		fmt.Println("Next commands:")
		fmt.Printf("  juleson sessions watch %s\n", sessionID)
		return
	}
	for i := range plans {
		plan := &plans[i]
		fmt.Printf("%d. Activity ID: %s\n", i+1, plan.ActivityID)
		if plan.ActivityName != "" {
			fmt.Printf("   Activity Name: %s\n", plan.ActivityName)
		}
		fmt.Printf("   Plan ID: %s\n", plan.PlanID)
		if !plan.PlanCreateTime.IsZero() {
			fmt.Printf("   Created: %s\n", formatTime(plan.PlanCreateTime))
		} else if !plan.ActivityCreateTime.IsZero() {
			fmt.Printf("   Created: %s\n", formatTime(plan.ActivityCreateTime))
		}
		fmt.Printf("   Approved: %t\n", plan.Approved)
		if plan.ApprovalActivityID != "" {
			fmt.Printf("   Approval Activity ID: %s\n", plan.ApprovalActivityID)
		}
		fmt.Printf("   Steps (%d):\n", len(plan.Steps))
		for stepIndex, step := range plan.Steps {
			fmt.Printf("     %d. %s\n", stepIndex+1, step.Title)
			if step.Description != "" {
				fmt.Printf("        %s\n", step.Description)
			}
		}
		fmt.Println()
	}
	fmt.Println("Next commands:")
	fmt.Printf("  juleson sessions approve %s\n", sessionID)
	fmt.Printf("  juleson sessions message %s \"<message>\"\n", sessionID)
	fmt.Printf("  juleson sessions review %s <project-path>\n", sessionID)
	fmt.Printf("  juleson sessions watch %s\n", sessionID)
}

func printSessionReview(review *julessessions.SessionReview) {
	fmt.Printf("Session review for %s\n", review.SessionID)
	fmt.Println(strings.Repeat("=", 60))
	fmt.Printf("State: %s\n", review.Session.State)
	fmt.Printf("Title: %s\n", review.Session.Title)
	if review.Session.URL != "" {
		fmt.Printf("URL: %s\n", review.Session.URL)
	}
	printReviewLatestPlan(review)
	printReviewOutputs(review)
	printReviewArtifacts(review)
	printReviewPatchPreview(review)
	printReviewWorktree(review)

	printStringList("Warnings", review.Warnings)
	printStringList("Blockers", review.Blockers)
	printStringList("Verification suggestions", review.VerificationSuggestions)

	printReviewNextActions(review)
}

func printReviewLatestPlan(review *julessessions.SessionReview) {
	if review.LatestPlan != nil {
		fmt.Printf("\nLatest plan: %s (%d step(s), approved: %t)\n", review.LatestPlan.PlanID, len(review.LatestPlan.Steps), review.LatestPlan.Approved)
		for stepIndex, step := range review.LatestPlan.Steps {
			fmt.Printf("  %d. %s\n", stepIndex+1, step.Title)
			if step.Description != "" {
				fmt.Printf("     %s\n", step.Description)
			}
		}
	} else {
		fmt.Println("\nLatest plan: none")
	}
}

func printReviewOutputs(review *julessessions.SessionReview) {
	fmt.Printf("\nOutputs: %d\n", len(review.Outputs))
	for _, output := range review.Outputs {
		if output.PullRequest != nil {
			fmt.Printf("  Pull Request: %s\n", output.PullRequest.URL)
			fmt.Printf("    Title: %s\n", output.PullRequest.Title)
		} else if output.ChangeSet != nil {
			fmt.Println("  ChangeSet output")
		}
	}
}

func printReviewArtifacts(review *julessessions.SessionReview) {
	fmt.Printf("\nArtifacts: %d\n", len(review.ArtifactManifests))
	for i := range review.ArtifactManifests {
		manifest := &review.ArtifactManifests[i]
		fmt.Printf("  Activity %s artifact %d: %s", manifest.ActivityID, manifest.Index, manifest.Type)
		if manifest.FileCount > 0 {
			fmt.Printf(" (%d file(s))", manifest.FileCount)
		}
		fmt.Println()
	}
}

func printReviewPatchPreview(review *julessessions.SessionReview) {
	fmt.Printf("\nPatch preview: %s\n", review.PatchPreview.Summary)
	for _, file := range review.PatchPreview.Files {
		fmt.Printf("  %s (+%d -%d)\n", file.Path, file.LinesAdded, file.LinesRemoved)
	}
	for _, message := range review.PatchPreview.SuggestedCommitMessages {
		fmt.Printf("  Suggested commit: %s\n", message)
	}
	if review.PatchPreview.Error != "" {
		fmt.Printf("  Preview error: %s\n", review.PatchPreview.Error)
	}
}

func printReviewWorktree(review *julessessions.SessionReview) {
	fmt.Printf("\nWorktree: %s\n", review.Worktree.WorkingDir)
	switch {
	case review.Worktree.Error != "":
		fmt.Printf("  Error: %s\n", review.Worktree.Error)
	case review.Worktree.Clean:
		fmt.Println("  Clean: true")
	default:
		fmt.Println("  Clean: false")
		fmt.Printf("  Status:\n%s\n", review.Worktree.Status)
	}
}

func printReviewNextActions(review *julessessions.SessionReview) {
	fmt.Println("\nNext actions:")
	for _, action := range review.NextActions {
		if action.Command != "" {
			fmt.Printf("  %s: %s\n", action.Label, action.Command)
		} else {
			fmt.Printf("  %s\n", action.Label)
		}
		if action.Reason != "" {
			fmt.Printf("    %s\n", action.Reason)
		}
	}
}

func printStringList(title string, values []string) {
	if len(values) == 0 {
		return
	}
	fmt.Printf("\n%s:\n", title)
	for _, value := range values {
		fmt.Printf("  %s\n", value)
	}
}

func printJSON(value any) error {
	encoder := json.NewEncoder(os.Stdout)
	encoder.SetIndent("", "  ")
	return encoder.Encode(value)
}

func formatTime(value time.Time) string {
	return value.Format(time.RFC3339)
}
