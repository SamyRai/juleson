package commands

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/SamyRai/juleson/internal/config"
	"github.com/SamyRai/juleson/internal/julesops"
	"github.com/SamyRai/juleson/pkg/jules"
)

// approveSessionPlan approves a plan in a session
func approveSessionPlan(cfg *config.Config, sessionID string) error {
	// Initialize Jules client
	julesClient := jules.NewClient(cfg.Jules.APIKey, jules.WithBaseURL(cfg.Jules.BaseURL), jules.WithTimeout(cfg.Jules.Timeout), jules.WithRetryAttempts(cfg.Jules.RetryAttempts))

	ctx := context.Background()

	fmt.Printf("✅ Approving plan for session: %s\n", sessionID)

	err := julesClient.ApprovePlan(ctx, sessionID)
	if err != nil {
		return fmt.Errorf("failed to approve plan: %w", err)
	}

	fmt.Println("✅ Plan approved successfully!")
	fmt.Printf("💡 Jules will now execute the approved plan. Monitor at: https://jules.google.com/session/%s\n", sessionID)

	return nil
}
func createSession(cfg *config.Config, sourceID string, prompt string, noSource bool) error {
	// Initialize Jules client
	julesClient := jules.NewClient(cfg.Jules.APIKey, jules.WithBaseURL(cfg.Jules.BaseURL), jules.WithTimeout(cfg.Jules.Timeout), jules.WithRetryAttempts(cfg.Jules.RetryAttempts))

	ctx := context.Background()

	fmt.Printf("🚀 Creating new Jules session...\n")
	if noSource {
		fmt.Printf("Source: repoless\n")
	} else {
		fmt.Printf("Source: %s\n", normalizeSourceID(sourceID))
	}
	fmt.Printf("Prompt: %s\n\n", prompt)

	req := &jules.CreateSessionRequest{
		Prompt:              prompt,
		RequirePlanApproval: false, // Default to auto-approval for CLI
	}
	if !noSource {
		req.SourceContext = &jules.SourceContext{
			Source: normalizeSourceID(sourceID),
		}
	}

	session, err := julesClient.CreateSession(ctx, req)
	if err != nil {
		return fmt.Errorf("failed to create session: %w", err)
	}

	fmt.Printf("✅ Session created successfully!\n\n")
	fmt.Printf("📊 Session Details:\n")
	fmt.Printf("ID: %s\n", session.ID)
	fmt.Printf("Title: %s\n", session.Title)
	fmt.Printf("State: %s\n", session.State)
	fmt.Printf("Created: %s\n", session.CreateTime)
	if session.URL != "" {
		fmt.Printf("URL: %s\n", session.URL)
	}

	fmt.Printf("\n💡 Jules is now working on your request. Monitor progress at: %s\n", session.URL)
	fmt.Printf("💡 Use 'juleson sessions get %s' to check status and activities\n", session.ID)

	return nil
}
func deleteSession(cfg *config.Config, sessionID string, force bool) error {
	julesClient := jules.NewClient(cfg.Jules.APIKey, jules.WithBaseURL(cfg.Jules.BaseURL), jules.WithTimeout(cfg.Jules.Timeout), jules.WithRetryAttempts(cfg.Jules.RetryAttempts))

	if !force {
		fmt.Printf("Type the session ID to confirm deletion (%s): ", sessionID)
		scanner := bufio.NewScanner(os.Stdin)
		if !scanner.Scan() {
			if err := scanner.Err(); err != nil {
				return fmt.Errorf("failed to read confirmation: %w", err)
			}
			return fmt.Errorf("session deletion cancelled")
		}
		if strings.TrimSpace(scanner.Text()) != sessionID {
			return fmt.Errorf("session deletion cancelled")
		}
	}

	if err := julesClient.DeleteSession(context.Background(), sessionID); err != nil {
		return fmt.Errorf("failed to delete session: %w", err)
	}

	fmt.Printf("✅ Deleted session: %s\n", sessionID)
	return nil
}
func normalizeSourceID(sourceID string) string {
	return jules.NormalizeSourceName(sourceID)
}
func listSessions(cfg *config.Config) error {
	// Initialize Jules client
	julesClient := jules.NewClient(cfg.Jules.APIKey, jules.WithBaseURL(cfg.Jules.BaseURL), jules.WithTimeout(cfg.Jules.Timeout), jules.WithRetryAttempts(cfg.Jules.RetryAttempts))

	fmt.Println("🔍 Listing Jules sessions...")
	fmt.Println("============================")

	response, err := julesClient.ListSessionsWithPagination(context.Background(), 50, "")
	if err != nil {
		return fmt.Errorf("failed to list sessions: %w", err)
	}

	sessions := response.Sessions
	if len(sessions) == 0 {
		fmt.Println("📭 No sessions found.")
		return nil
	}

	fmt.Printf("📊 Found %d session(s):\n\n", len(sessions))

	for i, session := range sessions {
		fmt.Printf("%d. Session: %s\n", i+1, session.ID)
		fmt.Printf("   Title: %s\n", session.Title)
		fmt.Printf("   State: %s\n", session.State)
		fmt.Printf("   Created: %s\n", session.CreateTime)
		if !session.UpdateTime.IsZero() {
			fmt.Printf("   Updated: %s\n", session.UpdateTime)
		}
		if session.SourceContext != nil && session.SourceContext.Source != "" {
			fmt.Printf("   Source: %s\n", session.SourceContext.Source)
		}
		if session.RequirePlanApproval {
			fmt.Printf("   Plan Approval Required: Yes\n")
		}
		if session.AutomationMode != "" {
			fmt.Printf("   Automation Mode: %s\n", session.AutomationMode)
		}
		if len(session.Outputs) > 0 {
			fmt.Printf("   Outputs: %d\n", len(session.Outputs))
		}

		// Status indicators
		statusText := getSessionStatusText(session.State)
		statusIcon := getSessionStatusIcon(session.State)
		fmt.Printf("   %s %s\n", statusIcon, statusText)
		fmt.Println()
	}

	return nil
}

// showSessionStatus shows a summary of session statuses
func showSessionStatus(cfg *config.Config) error {
	// Initialize Jules client
	julesClient := jules.NewClient(cfg.Jules.APIKey, jules.WithBaseURL(cfg.Jules.BaseURL), jules.WithTimeout(cfg.Jules.Timeout), jules.WithRetryAttempts(cfg.Jules.RetryAttempts))

	fmt.Println("📊 Jules Session Status")
	fmt.Println("=======================")

	response, err := julesClient.ListSessionsWithPagination(context.Background(), 100, "")
	if err != nil {
		return fmt.Errorf("failed to get session status: %w", err)
	}

	sessions := response.Sessions
	totalSessions := len(sessions)

	if totalSessions == 0 {
		fmt.Println("📭 No sessions found.")
		return nil
	}

	// Count sessions by state
	stateCounts := make(map[string]int)
	for _, session := range sessions {
		stateCounts[string(session.State)]++
	}

	fmt.Printf("Total Sessions: %d\n\n", totalSessions)

	// Display state breakdown
	fmt.Println("Session States:")
	for state, count := range stateCounts {
		percentage := float64(count) / float64(totalSessions) * 100
		icon := getSessionStatusIcon(jules.SessionState(state))
		fmt.Printf("  %s %s: %d (%.1f%%)\n", icon, state, count, percentage)
	}

	// Active sessions summary
	activeCount := stateCounts["IN_PROGRESS"] + stateCounts["PLANNING"]
	if activeCount > 0 {
		fmt.Printf("\n⚠️  %d session(s) are currently active/running\n", activeCount)
	} else {
		fmt.Println("\n✅ No active sessions currently running")
	}

	// Recent sessions (last 5)
	if totalSessions > 0 {
		fmt.Println("\n🕒 Recent Sessions:")
		recentCount := 5
		if totalSessions < recentCount {
			recentCount = totalSessions
		}

		for i := 0; i < recentCount; i++ {
			session := sessions[i]
			statusIcon := getSessionStatusIcon(session.State)
			fmt.Printf("  %s %s - %s (%s)\n", statusIcon, session.ID[:12], session.Title, session.State)
		}
	}

	return nil
}

// getSessionDetails gets detailed information about a specific session
func getSessionDetails(cfg *config.Config, sessionID string) error {
	// Initialize Jules client
	julesClient := jules.NewClient(cfg.Jules.APIKey, jules.WithBaseURL(cfg.Jules.BaseURL), jules.WithTimeout(cfg.Jules.Timeout), jules.WithRetryAttempts(cfg.Jules.RetryAttempts))

	ctx := context.Background()

	fmt.Printf("🔍 Fetching session details for: %s\n", sessionID)
	fmt.Println("=" + string(make([]byte, 60)))

	// Get session details
	session, err := julesClient.GetSession(ctx, sessionID)
	if err != nil {
		return fmt.Errorf("failed to get session: %w", err)
	}

	// Display session information
	fmt.Printf("\n📊 Session Information\n")
	fmt.Printf("ID: %s\n", session.ID)
	fmt.Printf("Title: %s\n", session.Title)
	fmt.Printf("State: %s %s\n", getSessionStatusIcon(session.State), session.State)
	fmt.Printf("Created: %s\n", session.CreateTime)
	if !session.UpdateTime.IsZero() {
		fmt.Printf("Updated: %s\n", session.UpdateTime)
	}
	if session.URL != "" {
		fmt.Printf("URL: %s\n", session.URL)
	}
	if session.SourceContext != nil && session.SourceContext.Source != "" {
		fmt.Printf("Source: %s\n", session.SourceContext.Source)
		if session.SourceContext.GithubRepoContext != nil {
			fmt.Printf("Branch: %s\n", session.SourceContext.GithubRepoContext.StartingBranch)
		}
	}
	fmt.Printf("Automation Mode: %s\n", session.AutomationMode)
	fmt.Printf("Requires Approval: %t\n", session.RequirePlanApproval)

	// Display outputs if any
	if len(session.Outputs) > 0 {
		fmt.Printf("\n📤 Outputs:\n")
		for i, output := range session.Outputs {
			if output.PullRequest != nil {
				fmt.Printf("  %d. Pull Request:\n", i+1)
				fmt.Printf("     URL: %s\n", output.PullRequest.URL)
				fmt.Printf("     Title: %s\n", output.PullRequest.Title)
				if output.PullRequest.Description != "" {
					fmt.Printf("     Description: %s\n", output.PullRequest.Description)
				}
			}
		}
	}

	// Get activities
	fmt.Printf("\n📋 Activities:\n")
	activities, err := julesClient.ListActivities(ctx, sessionID, 100)
	if err != nil {
		fmt.Printf("⚠️  Could not fetch activities: %v\n", err)
		return nil
	}

	if len(activities) == 0 {
		fmt.Println("  No activities yet - session is still initializing")
		return nil
	}

	fmt.Printf("  Found %d activities\n\n", len(activities))

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
			for j, step := range activity.PlanGenerated.Plan.Steps {
				if j < 5 { // Show first 5 steps
					fmt.Printf("      %d. %s\n", step.Index, step.Title)
				}
			}
			if len(activity.PlanGenerated.Plan.Steps) > 5 {
				fmt.Printf("      ... and %d more steps\n", len(activity.PlanGenerated.Plan.Steps)-5)
			}
		}

		if activity.PlanApproved != nil {
			fmt.Printf("   ✅ Plan Approved (Plan ID: %s)\n", activity.PlanApproved.PlanID)
		}

		if activity.ProgressUpdated != nil {
			fmt.Printf("   ⚙️  Progress: %s\n", activity.ProgressUpdated.Title)
			if activity.ProgressUpdated.Description != "" {
				desc := activity.ProgressUpdated.Description
				if len(desc) > 100 {
					desc = desc[:100] + "..."
				}
				fmt.Printf("      %s\n", desc)
			}
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

	fmt.Printf("💡 View full session at: %s\n", session.URL)

	return nil
}

// sendSessionMessage sends a message to a session
func sendSessionMessage(cfg *config.Config, sessionID string, message string) error {
	// Initialize Jules client
	julesClient := jules.NewClient(cfg.Jules.APIKey, jules.WithBaseURL(cfg.Jules.BaseURL), jules.WithTimeout(cfg.Jules.Timeout), jules.WithRetryAttempts(cfg.Jules.RetryAttempts))

	ctx := context.Background()

	fmt.Printf("📤 Sending message to session: %s\n", sessionID)
	fmt.Printf("Message: %s\n\n", message)

	req := &jules.SendMessageRequest{
		Prompt: message,
	}

	err := julesClient.SendMessage(ctx, sessionID, req)
	if err != nil {
		return fmt.Errorf("failed to send message: %w", err)
	}

	fmt.Println("✅ Message sent successfully!")
	fmt.Println("💡 Jules will process your message and respond with activities.")
	fmt.Printf("💡 Monitor at: https://jules.google.com/session/%s\n", sessionID)

	return nil
}

// downloadSessionArtifacts downloads all artifacts from all activities in a session
func downloadSessionArtifacts(cfg *config.Config, sessionID string, outputDir string) error {
	// Initialize Jules client
	julesClient := jules.NewClient(cfg.Jules.APIKey, jules.WithBaseURL(cfg.Jules.BaseURL), jules.WithTimeout(cfg.Jules.Timeout), jules.WithRetryAttempts(cfg.Jules.RetryAttempts))

	ctx := context.Background()

	fmt.Printf("📥 Downloading artifacts from session: %s\n", sessionID)
	fmt.Printf("📁 Output directory: %s\n", outputDir)
	fmt.Println("=" + string(make([]byte, 60)))

	// Create output directory if it doesn't exist
	options := &julesops.ArtifactDownloadOptions{
		DestinationDir: outputDir,
		CreateDir:      true,
		Overwrite:      false,
	}

	// Download all artifacts from the session
	downloadedFiles, err := julesops.DownloadAllSessionArtifacts(ctx, julesClient, sessionID, options)
	if err != nil {
		return fmt.Errorf("failed to download session artifacts: %w", err)
	}

	if len(downloadedFiles) == 0 {
		fmt.Println("📭 No artifacts found in this session.")
		return nil
	}

	fmt.Printf("✅ Successfully downloaded %d artifact(s):\n", len(downloadedFiles))
	for i, filename := range downloadedFiles {
		fmt.Printf("  %d. %s\n", i+1, filename)
	}

	fmt.Printf("\n💡 Artifacts saved to: %s\n", outputDir)
	return nil
}

// downloadActivityArtifacts downloads all artifacts from a specific activity
func downloadActivityArtifacts(cfg *config.Config, sessionID string, activityID string, outputDir string) error {
	// Initialize Jules client
	julesClient := jules.NewClient(cfg.Jules.APIKey, jules.WithBaseURL(cfg.Jules.BaseURL), jules.WithTimeout(cfg.Jules.Timeout), jules.WithRetryAttempts(cfg.Jules.RetryAttempts))

	ctx := context.Background()

	fmt.Printf("📥 Downloading artifacts from activity: %s\n", activityID)
	fmt.Printf("📁 Session: %s\n", sessionID)
	fmt.Printf("📁 Output directory: %s\n", outputDir)
	fmt.Println("=" + string(make([]byte, 60)))

	// Create output directory if it doesn't exist
	options := &julesops.ArtifactDownloadOptions{
		DestinationDir: outputDir,
		CreateDir:      true,
		Overwrite:      false,
	}

	// Download artifacts from the specific activity
	downloadedFiles, err := julesops.DownloadArtifactFromActivity(ctx, julesClient, sessionID, activityID, options)
	if err != nil {
		return fmt.Errorf("failed to download activity artifacts: %w", err)
	}

	if len(downloadedFiles) == 0 {
		fmt.Println("📭 No artifacts found in this activity.")
		return nil
	}

	fmt.Printf("✅ Successfully downloaded %d artifact(s):\n", len(downloadedFiles))
	for i, filename := range downloadedFiles {
		fmt.Printf("  %d. %s\n", i+1, filename)
	}

	fmt.Printf("\n💡 Artifacts saved to: %s\n", outputDir)
	return nil
}

// previewSessionArtifacts previews all artifacts from all activities in a session
func previewSessionArtifacts(cfg *config.Config, sessionID string) error {
	// Initialize Jules client
	julesClient := jules.NewClient(cfg.Jules.APIKey, jules.WithBaseURL(cfg.Jules.BaseURL), jules.WithTimeout(cfg.Jules.Timeout), jules.WithRetryAttempts(cfg.Jules.RetryAttempts))

	ctx := context.Background()

	fmt.Printf("👁️  Previewing artifacts from session: %s\n", sessionID)
	fmt.Println("=" + string(make([]byte, 60)))

	// Get all activities for the session
	activities, err := julesClient.ListActivities(ctx, sessionID, 100)
	if err != nil {
		return fmt.Errorf("failed to list activities: %w", err)
	}

	if len(activities) == 0 {
		fmt.Println("📭 No activities found in this session.")
		return nil
	}

	totalArtifacts := 0
	for i, activity := range activities {
		if len(activity.Artifacts) > 0 {
			fmt.Printf("\n📋 Activity %d: %s\n", i+1, activity.ID)
			err := previewActivityArtifactsContent(activity.Artifacts)
			if err != nil {
				fmt.Printf("⚠️  Failed to preview activity %s: %v\n", activity.ID, err)
			} else {
				totalArtifacts += len(activity.Artifacts)
			}
		}
	}

	if totalArtifacts == 0 {
		fmt.Println("📭 No artifacts found in this session.")
	} else {
		fmt.Printf("\n✅ Previewed %d artifact(s) total\n", totalArtifacts)
	}

	return nil
}

// previewActivityArtifacts previews all artifacts from a specific activity
func previewActivityArtifacts(cfg *config.Config, sessionID string, activityID string) error {
	// Initialize Jules client
	julesClient := jules.NewClient(cfg.Jules.APIKey, jules.WithBaseURL(cfg.Jules.BaseURL), jules.WithTimeout(cfg.Jules.Timeout), jules.WithRetryAttempts(cfg.Jules.RetryAttempts))

	ctx := context.Background()

	fmt.Printf("👁️  Previewing artifacts from activity: %s\n", activityID)
	fmt.Printf("📁 Session: %s\n", sessionID)
	fmt.Println("=" + string(make([]byte, 60)))

	// Get the activity to access its artifacts
	activity, err := julesClient.GetActivity(ctx, sessionID, activityID)
	if err != nil {
		return fmt.Errorf("failed to get activity: %w", err)
	}

	if len(activity.Artifacts) == 0 {
		fmt.Println("📭 No artifacts found in this activity.")
		return nil
	}

	err = previewActivityArtifactsContent(activity.Artifacts)
	if err != nil {
		return err
	}

	fmt.Printf("\n✅ Previewed %d artifact(s)\n", len(activity.Artifacts))
	return nil
}
