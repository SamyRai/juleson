package commands

import (
	"context"
	"fmt"

	"github.com/SamyRai/juleson/internal/config"
	"github.com/SamyRai/juleson/internal/jules"

	"github.com/spf13/cobra"
)

// NewActivitiesCommand creates the activities command
func NewActivitiesCommand(cfg *config.Config) *cobra.Command {
	activitiesCmd := &cobra.Command{
		Use:   "activities",
		Short: "Manage Jules session activities",
		Long:  "List and manage activities within Jules sessions",
	}

	// List activities
	activitiesCmd.AddCommand(&cobra.Command{
		Use:   "list [session-id]",
		Short: "List all activities in a session",
		Long:  "List all activities for the specified session",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return listActivities(cfg, args[0])
		},
	})

	// Get activity
	activitiesCmd.AddCommand(&cobra.Command{
		Use:   "get [session-id] [activity-id]",
		Short: "Get details for a specific activity",
		Long:  "Get detailed information about a specific activity within a session",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			return getActivity(cfg, args[0], args[1])
		},
	})

	return activitiesCmd
}

// listActivities lists all activities in a session
func listActivities(cfg *config.Config, sessionID string) error {
	// Initialize Jules client
	julesClient := jules.NewClient(
		cfg.Jules.APIKey,
		cfg.Jules.BaseURL,
		cfg.Jules.Timeout,
		cfg.Jules.RetryAttempts,
	)

	ctx := context.Background()

	fmt.Printf("📋 Listing activities for session: %s\n", sessionID)
	fmt.Println("=" + string(make([]byte, 60)))

	activities, err := julesClient.ListActivities(ctx, sessionID, 100)
	if err != nil {
		return fmt.Errorf("failed to list activities: %w", err)
	}

	if len(activities) == 0 {
		fmt.Println("📭 No activities found in this session.")
		return nil
	}

	fmt.Printf("Found %d activities:\n\n", len(activities))

	for i, activity := range activities {
		originator := "❓"
		if activity.Originator == "agent" {
			originator = "🤖"
		} else if activity.Originator == "user" {
			originator = "👤"
		}

		fmt.Printf("%d. %s [%s] - %s\n", i+1, originator, activity.Originator, activity.CreateTime)

		// Show activity type and details
		if activity.PlanGenerated != nil {
			fmt.Printf("   📝 Plan Generated (%d steps)\n", len(activity.PlanGenerated.Plan.Steps))
		}

		if activity.PlanApproved != nil {
			fmt.Printf("   ✅ Plan Approved (Plan ID: %s)\n", activity.PlanApproved.PlanID)
		}

		if activity.ProgressUpdated != nil {
			fmt.Printf("   ⚙️  Progress: %s\n", activity.ProgressUpdated.Title)
		}

		if activity.SessionCompleted != nil {
			fmt.Printf("   ✅ Session Completed\n")
		}

		// Show artifacts summary
		if len(activity.Artifacts) > 0 {
			fmt.Printf("   📦 %d artifact(s)\n", len(activity.Artifacts))
		}

		fmt.Println()
	}

	return nil
}

// getActivity gets details for a specific activity
func getActivity(cfg *config.Config, sessionID string, activityID string) error {
	// Initialize Jules client
	julesClient := jules.NewClient(
		cfg.Jules.APIKey,
		cfg.Jules.BaseURL,
		cfg.Jules.Timeout,
		cfg.Jules.RetryAttempts,
	)

	ctx := context.Background()

	fmt.Printf("🔍 Fetching activity details: %s\n", activityID)
	fmt.Printf("📁 Session: %s\n", sessionID)
	fmt.Println("=" + string(make([]byte, 60)))

	activity, err := julesClient.GetActivity(ctx, sessionID, activityID)
	if err != nil {
		return fmt.Errorf("failed to get activity: %w", err)
	}

	// Display activity information
	fmt.Printf("\n📊 Activity Information\n")
	fmt.Printf("ID: %s\n", activity.ID)
	fmt.Printf("Originator: %s\n", activity.Originator)
	fmt.Printf("Created: %s\n", activity.CreateTime)

	// Show activity type and details
	if activity.PlanGenerated != nil {
		fmt.Printf("\n📝 Plan Generated:\n")
		fmt.Printf("Plan ID: %s\n", activity.PlanGenerated.Plan.ID)
		fmt.Printf("Steps: %d\n", len(activity.PlanGenerated.Plan.Steps))
		for _, step := range activity.PlanGenerated.Plan.Steps {
			fmt.Printf("  %d. %s\n", step.Index, step.Title)
			if step.Description != "" {
				fmt.Printf("     %s\n", step.Description)
			}
		}
	}

	if activity.PlanApproved != nil {
		fmt.Printf("\n✅ Plan Approved:\n")
		fmt.Printf("Plan ID: %s\n", activity.PlanApproved.PlanID)
	}

	if activity.ProgressUpdated != nil {
		fmt.Printf("\n⚙️  Progress Update:\n")
		fmt.Printf("Title: %s\n", activity.ProgressUpdated.Title)
		if activity.ProgressUpdated.Description != "" {
			fmt.Printf("Description: %s\n", activity.ProgressUpdated.Description)
		}
	}

	if activity.SessionCompleted != nil {
		fmt.Printf("\n✅ Session Completed\n")
	}

	if activity.SessionFailed != nil {
		fmt.Printf("\n❌ Session Failed:\n")
		fmt.Printf("Reason: %s\n", activity.SessionFailed.Reason)
	}

	// Show artifacts
	if len(activity.Artifacts) > 0 {
		fmt.Printf("\n📦 Artifacts (%d):\n", len(activity.Artifacts))
		for i, artifact := range activity.Artifacts {
			fmt.Printf("\n  Artifact %d:\n", i+1)

			if artifact.BashOutput != nil {
				fmt.Printf("    🖥️  Bash Output:\n")
				fmt.Printf("    Command: %s\n", artifact.BashOutput.Command)
				fmt.Printf("    Exit Code: %d\n", artifact.BashOutput.ExitCode)
				if len(artifact.BashOutput.Output) > 200 {
					fmt.Printf("    Output: %s... (truncated)\n", artifact.BashOutput.Output[:200])
				} else {
					fmt.Printf("    Output: %s\n", artifact.BashOutput.Output)
				}
			} else if artifact.ChangeSet != nil && artifact.ChangeSet.GitPatch != nil {
				fmt.Printf("    🔀 Git Patch:\n")
				if artifact.ChangeSet.GitPatch.SuggestedCommitMessage != "" {
					fmt.Printf("    Commit Message: %s\n", artifact.ChangeSet.GitPatch.SuggestedCommitMessage)
				}
				fmt.Printf("    Has diff content: %t\n", artifact.ChangeSet.GitPatch.UnidiffPatch != "")
			} else if artifact.Media != nil {
				fmt.Printf("    🖼️  Media:\n")
				fmt.Printf("    Type: %s\n", artifact.Media.MimeType)
				fmt.Printf("    Size: %d bytes\n", len(artifact.Media.Data))
			} else {
				fmt.Printf("    📄 Unknown artifact type\n")
			}
		}
	}

	return nil
}
