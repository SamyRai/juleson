package commands

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/SamyRai/juleson/internal/config"
	"github.com/SamyRai/juleson/internal/julesops"
	"github.com/SamyRai/juleson/pkg/jules"
)

type CreateSessionOptions struct {
	NoSource            bool
	PromptFile          string
	Title               string
	StartingBranch      string
	RequirePlanApproval bool
	AutomationMode      string
}

type BatchSessionOptions struct {
	Parallel       int
	Title          string
	BatchID        string
	GroupTitle     string
	StartingBranch string
	AutomationMode string
}

type ApplySessionOptions struct {
	Confirm           bool
	AllowDirty        bool
	ActivityID        string
	ArtifactIndex     int
	HasArtifactIndex  bool
	AllowBaseMismatch bool
}

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
func createSession(cfg *config.Config, sourceID string, prompt string, options CreateSessionOptions) error {
	// Initialize Jules client
	julesClient := jules.NewClient(cfg.Jules.APIKey, jules.WithBaseURL(cfg.Jules.BaseURL), jules.WithTimeout(cfg.Jules.Timeout), jules.WithRetryAttempts(cfg.Jules.RetryAttempts))

	ctx := context.Background()
	if options.PromptFile != "" {
		loadedPrompt, err := loadPromptFile(options.PromptFile)
		if err != nil {
			return err
		}
		prompt = loadedPrompt
	}
	sourceName := normalizeSourceID(sourceID)
	if !options.NoSource && sourceID == "." {
		source, err := julesops.InferSourceFromGitRemote(ctx, julesClient, ".")
		if err != nil {
			return err
		}
		sourceName = source.Name
	}

	fmt.Printf("🚀 Creating new Jules session...\n")
	if options.NoSource {
		fmt.Printf("Source: repoless\n")
	} else {
		fmt.Printf("Source: %s\n", sourceName)
	}
	fmt.Printf("Prompt: %s\n\n", prompt)

	req := &jules.CreateSessionRequest{
		Prompt:              prompt,
		Title:               options.Title,
		RequirePlanApproval: options.RequirePlanApproval,
		AutomationMode:      jules.AutomationMode(options.AutomationMode),
	}
	if !options.NoSource {
		req.SourceContext = &jules.SourceContext{
			Source: sourceName,
		}
		if options.StartingBranch != "" {
			req.SourceContext.GithubRepoContext = &jules.GithubRepoContext{StartingBranch: options.StartingBranch}
		}
	} else if options.StartingBranch != "" {
		return fmt.Errorf("--starting-branch requires a source-backed session")
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
	activeCount := stateCounts[string(jules.SessionStateInProgress)] + stateCounts[string(jules.SessionStatePlanning)] + stateCounts[string(jules.SessionStateQueued)]
	userActionCount := stateCounts[string(jules.SessionStateAwaitingPlanApproval)] + stateCounts[string(jules.SessionStateAwaitingUserFeedback)]
	if activeCount > 0 {
		fmt.Printf("\n⚠️  %d session(s) are currently active/running\n", activeCount)
	} else {
		fmt.Println("\n✅ No active sessions currently running")
	}
	if userActionCount > 0 {
		fmt.Printf("⏸  %d session(s) need user action\n", userActionCount)
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
			fmt.Printf("  %s %s - %s (%s)\n", statusIcon, shortSessionID(session.ID), session.Title, session.State)
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
	fmt.Println(strings.Repeat("=", 60))

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

func watchSession(cfg *config.Config, sessionID, intervalValue, timeoutValue string, followActivities bool, sinceValue, cursorOutput, initialState string, wakeOnStatusChange, wakeOnAgentMessage bool) error {
	julesClient := jules.NewClient(cfg.Jules.APIKey, jules.WithBaseURL(cfg.Jules.BaseURL), jules.WithTimeout(cfg.Jules.Timeout), jules.WithRetryAttempts(cfg.Jules.RetryAttempts))

	interval, err := time.ParseDuration(intervalValue)
	if err != nil {
		return fmt.Errorf("invalid --interval: %w", err)
	}
	timeout, err := time.ParseDuration(timeoutValue)
	if err != nil {
		return fmt.Errorf("invalid --timeout: %w", err)
	}
	if interval <= 0 {
		return fmt.Errorf("--interval must be greater than zero")
	}
	if timeout <= 0 {
		return fmt.Errorf("--timeout must be greater than zero")
	}
	var cursor time.Time
	baselineState := jules.SessionState(strings.TrimSpace(initialState))
	hasStateBaseline := baselineState != ""
	hasActivityBaseline := false
	if sinceValue != "" {
		parsed, err := time.Parse(time.RFC3339Nano, sinceValue)
		if err != nil {
			return fmt.Errorf("invalid --since: %w", err)
		}
		cursor = parsed
		hasActivityBaseline = true
	}

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	fmt.Printf("👁️  Watching session: %s\n", sessionID)
	fmt.Printf("Polling every %s for up to %s\n", interval, timeout)
	fmt.Println(strings.Repeat("=", 60))

	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	seenActivities := map[string]bool{}
	for {
		update, err := printSessionWatchUpdate(ctx, julesClient, sessionID, followActivities, wakeOnAgentMessage, seenActivities, cursor)
		if err != nil {
			return err
		}
		if wakeOnStatusChange {
			if !hasStateBaseline {
				baselineState = update.State
				hasStateBaseline = true
			} else if update.State != baselineState {
				fmt.Printf("Wake reason: session state changed from %s to %s.\n", baselineState, update.State)
				return nil
			}
		}
		if wakeOnAgentMessage {
			if !hasActivityBaseline {
				hasActivityBaseline = true
			} else if update.HasJulesAgentMessage {
				fmt.Printf("Wake reason: Jules sent a new message.\n")
				return nil
			}
		}
		if update.NextCursor.After(cursor) {
			cursor = update.NextCursor
			if cursorOutput != "" {
				if err := os.WriteFile(cursorOutput, []byte(cursor.Format(time.RFC3339Nano)+"\n"), 0644); err != nil {
					return fmt.Errorf("failed to write cursor output: %w", err)
				}
			}
		}
		if update.Stop {
			if !cursor.IsZero() {
				fmt.Printf("Next activity cursor: %s\n", cursor.Format(time.RFC3339Nano))
			}
			return nil
		}

		select {
		case <-ctx.Done():
			return fmt.Errorf("timeout watching session after %s", timeout)
		case <-ticker.C:
		}
	}
}

type sessionWatchUpdate struct {
	Stop                 bool
	NextCursor           time.Time
	State                jules.SessionState
	HasJulesAgentMessage bool
}

func printSessionWatchUpdate(ctx context.Context, client *jules.Client, sessionID string, followActivities bool, detectAgentMessage bool, seenActivities map[string]bool, cursor time.Time) (sessionWatchUpdate, error) {
	session, err := client.GetSession(ctx, sessionID)
	if err != nil {
		return sessionWatchUpdate{}, fmt.Errorf("failed to get session: %w", err)
	}

	update := sessionWatchUpdate{
		NextCursor: cursor,
		State:      session.State,
	}

	statusIcon := getSessionStatusIcon(session.State)
	statusText := getSessionStatusText(session.State)
	fmt.Printf("%s %s %s", time.Now().Format(time.RFC3339), statusIcon, session.State)
	if session.Title != "" {
		fmt.Printf(" - %s", session.Title)
	}
	fmt.Println()

	if followActivities || detectAgentMessage {
		activities, err := client.ListActivitiesSince(ctx, sessionID, cursor, 25)
		if err != nil {
			fmt.Printf("⚠️  Could not fetch activities: %v\n", err)
		} else {
			nextCursor := jules.ActivityCursor(activities)
			if nextCursor.After(cursor) {
				update.NextCursor = nextCursor
			}
			for i := len(activities) - 1; i >= 0; i-- {
				activity := activities[i]
				if activity.AgentMessaged != nil && (cursor.IsZero() || activity.CreateTime.After(cursor)) {
					update.HasJulesAgentMessage = true
				}
				if !followActivities {
					continue
				}
				key := activityResourceKey(activity)
				if key == "" || seenActivities[key] {
					continue
				}
				seenActivities[key] = true
				fmt.Printf("  • %s %s\n", activity.CreateTime.Format(time.RFC3339), describeActivity(activity))
			}
		}
	}

	switch {
	case session.State.NeedsUserAction():
		fmt.Printf("Next action: %s. Use 'juleson sessions get %s' to inspect, then approve or send feedback.\n", statusText, sessionID)
		update.Stop = true
		return update, nil
	case session.State == jules.SessionStateFailed:
		fmt.Printf("Next action: inspect failure details with 'juleson sessions get %s'.\n", sessionID)
		update.Stop = true
		return update, nil
	case session.State == jules.SessionStateCompleted:
		fmt.Printf("Next action: preview changes with 'juleson sessions apply %s <project-path>'.\n", sessionID)
		if len(session.Outputs) > 0 {
			fmt.Printf("Next output action: inspect outputs with 'juleson sessions outputs %s'.\n", sessionID)
		}
		update.Stop = true
		return update, nil
	case len(session.Outputs) > 0:
		fmt.Printf("Next action: inspect outputs with 'juleson sessions outputs %s'.\n", sessionID)
		update.Stop = true
		return update, nil
	default:
		return update, nil
	}
}

func activityResourceKey(activity jules.Activity) string {
	if activity.Name != "" {
		return activity.Name
	}
	return activity.ID
}

func describeActivity(activity jules.Activity) string {
	switch {
	case activity.PlanGenerated != nil:
		return fmt.Sprintf("plan generated (%d steps)", len(activity.PlanGenerated.Plan.Steps))
	case activity.PlanApproved != nil:
		return "plan approved"
	case activity.ProgressUpdated != nil:
		if activity.ProgressUpdated.Description != "" {
			return fmt.Sprintf("%s: %s", activity.ProgressUpdated.Title, truncate(activity.ProgressUpdated.Description, 120))
		}
		return activity.ProgressUpdated.Title
	case activity.SessionCompleted != nil:
		return "session completed"
	case activity.SessionFailed != nil:
		return fmt.Sprintf("session failed: %s", activity.SessionFailed.Reason)
	case activity.UserMessaged != nil:
		return "user message sent"
	case activity.AgentMessaged != nil:
		return fmt.Sprintf("agent message: %s", truncate(activity.AgentMessaged.AgentMessage, 120))
	default:
		return fmt.Sprintf("%s activity", activity.Originator)
	}
}

func applySessionChanges(cfg *config.Config, sessionID, projectPath string, options ApplySessionOptions) error {
	julesClient := jules.NewClient(cfg.Jules.APIKey, jules.WithBaseURL(cfg.Jules.BaseURL), jules.WithTimeout(cfg.Jules.Timeout), jules.WithRetryAttempts(cfg.Jules.RetryAttempts))
	ctx := context.Background()

	dryRun := !options.Confirm
	if !dryRun && !options.AllowDirty {
		clean, status, err := julesops.IsGitWorkingTreeClean(ctx, projectPath)
		if err != nil {
			return err
		}
		if !clean {
			return fmt.Errorf("target worktree has local changes; commit/stash them or pass --allow-dirty\n%s", status)
		}
	}

	patchOptions := &julesops.PatchApplicationOptions{
		WorkingDir:        projectPath,
		ActivityID:        options.ActivityID,
		ArtifactIndex:     options.ArtifactIndex,
		HasArtifactIndex:  options.HasArtifactIndex,
		AllowBaseMismatch: options.AllowBaseMismatch,
	}
	changes, previewErr := julesops.PreviewSessionPatchesWithOptions(ctx, julesClient, sessionID, patchOptions)
	if changes != nil {
		printSessionChangesSummary(changes)
	}
	if dryRun {
		if previewErr != nil {
			return previewErr
		}
		fmt.Printf("\nDry-run only. Re-run with --confirm to apply patches.\n")
		return nil
	}
	if previewErr != nil {
		return fmt.Errorf("refusing to apply because preview failed: %w", previewErr)
	}

	patchOptions.DryRun = false
	result, err := julesops.ApplySessionPatches(ctx, julesClient, sessionID, patchOptions)
	if err != nil {
		return fmt.Errorf("failed to apply session patches: %w", err)
	}
	for _, warning := range result.Warnings {
		fmt.Printf("⚠️  %s\n", warning)
	}
	if len(result.Errors) > 0 {
		return fmt.Errorf("some patches failed: %s", strings.Join(result.Errors, "; "))
	}

	fmt.Printf("\n✅ Applied %d patch(es) touching %d file(s).\n", result.PatchesApplied, len(result.FilesModified))
	return nil
}

func printSessionChangesSummary(changes *julesops.SessionChanges) {
	totalAdded := 0
	totalRemoved := 0
	for _, file := range changes.Files {
		totalAdded += file.LinesAdded
		totalRemoved += file.LinesRemoved
	}
	fmt.Printf("Patch summary: %d patch(es), %d file(s), +%d -%d\n", changes.TotalPatches, len(changes.Files), totalAdded, totalRemoved)
	for _, file := range changes.Files {
		fmt.Printf("  %s (+%d -%d)\n", file.Path, file.LinesAdded, file.LinesRemoved)
	}
	for _, message := range changes.SuggestedCommitMessages {
		fmt.Printf("Suggested commit message: %s\n", message)
	}
	for _, warning := range changes.Warnings {
		fmt.Printf("Warning: %s\n", warning)
	}
}

func listSessionArtifacts(cfg *config.Config, sessionID string) error {
	julesClient := jules.NewClient(cfg.Jules.APIKey, jules.WithBaseURL(cfg.Jules.BaseURL), jules.WithTimeout(cfg.Jules.Timeout), jules.WithRetryAttempts(cfg.Jules.RetryAttempts))
	manifests, err := julesops.ListSessionArtifactManifests(context.Background(), julesClient, sessionID)
	if err != nil {
		return fmt.Errorf("failed to list session artifacts: %w", err)
	}
	if len(manifests) == 0 {
		fmt.Println("No artifacts found.")
		return nil
	}
	fmt.Printf("Artifacts for session %s\n", sessionID)
	fmt.Println(strings.Repeat("=", 60))
	for _, manifest := range manifests {
		fmt.Printf("Activity: %s  Index: %d  Type: %s\n", manifest.ActivityID, manifest.Index, manifest.Type)
		if !manifest.ActivityCreateTime.IsZero() {
			fmt.Printf("  Created: %s\n", manifest.ActivityCreateTime.Format(time.RFC3339))
		}
		if manifest.FileCount > 0 {
			fmt.Printf("  Files: %d\n", manifest.FileCount)
			for _, file := range manifest.Files {
				fmt.Printf("    %s (+%d -%d)\n", file.Path, file.LinesAdded, file.LinesRemoved)
			}
		} else if manifest.Empty {
			fmt.Printf("  Empty changeset: no diff content\n")
		}
		if manifest.BaseCommitID != "" {
			fmt.Printf("  Base commit: %s\n", manifest.BaseCommitID)
		}
		if manifest.SuggestedCommitMessage != "" {
			fmt.Printf("  Suggested commit: %s\n", manifest.SuggestedCommitMessage)
		}
		if manifest.MediaMIMEType != "" {
			fmt.Printf("  Media MIME: %s\n", manifest.MediaMIMEType)
		}
		if manifest.BashCommand != "" {
			fmt.Printf("  Bash command: %s\n", manifest.BashCommand)
		}
		if manifest.BashExitCode != nil {
			fmt.Printf("  Bash exit code: %d\n", *manifest.BashExitCode)
		}
		fmt.Println()
	}
	return nil
}

func showSessionOutputs(cfg *config.Config, sessionID string) error {
	julesClient := jules.NewClient(cfg.Jules.APIKey, jules.WithBaseURL(cfg.Jules.BaseURL), jules.WithTimeout(cfg.Jules.Timeout), jules.WithRetryAttempts(cfg.Jules.RetryAttempts))
	session, err := julesClient.GetSession(context.Background(), sessionID)
	if err != nil {
		return fmt.Errorf("failed to get session: %w", err)
	}
	if len(session.Outputs) == 0 {
		fmt.Println("No outputs found.")
		return nil
	}
	fmt.Printf("Outputs for session %s\n", sessionID)
	fmt.Println(strings.Repeat("=", 60))
	supported := 0
	for _, output := range session.Outputs {
		if output.PullRequest != nil {
			supported++
			fmt.Printf("%d. ", supported)
			fmt.Println("Pull Request")
			fmt.Printf("   URL: %s\n", output.PullRequest.URL)
			fmt.Printf("   Title: %s\n", output.PullRequest.Title)
			if output.PullRequest.Description != "" {
				fmt.Printf("   Description: %s\n", output.PullRequest.Description)
			}
		}
	}
	if supported == 0 {
		fmt.Println("No supported documented output payloads found.")
	}
	return nil
}

func batchCreateSessions(cfg *config.Config, sourceID, taskFileOrPrompt string, options BatchSessionOptions) error {
	if options.Parallel < 1 || options.Parallel > 5 {
		return fmt.Errorf("--parallel must be between 1 and 5")
	}

	prompt, err := loadPromptArgument(taskFileOrPrompt)
	if err != nil {
		return err
	}

	julesClient := jules.NewClient(cfg.Jules.APIKey, jules.WithBaseURL(cfg.Jules.BaseURL), jules.WithTimeout(cfg.Jules.Timeout), jules.WithRetryAttempts(cfg.Jules.RetryAttempts))
	ctx := context.Background()
	sourceName := normalizeSourceID(sourceID)
	if options.BatchID == "" {
		options.BatchID = "batch-" + time.Now().UTC().Format("20060102150405")
	}
	if options.GroupTitle == "" {
		options.GroupTitle = options.Title
	}

	fmt.Printf("🚀 Creating %d parallel Jules session(s)\n", options.Parallel)
	fmt.Printf("Batch ID: %s\n", options.BatchID)
	if options.GroupTitle != "" {
		fmt.Printf("Group title: %s\n", options.GroupTitle)
	}
	fmt.Printf("Source: %s\n", sourceName)
	fmt.Printf("Plan approval: required\n")
	fmt.Println(strings.Repeat("=", 60))

	for i := 1; i <= options.Parallel; i++ {
		title := options.Title
		if title != "" && options.Parallel > 1 {
			title = fmt.Sprintf("%s (%d/%d)", title, i, options.Parallel)
		}
		batchPrompt := fmt.Sprintf("Batch ID: %s\n", options.BatchID)
		if options.GroupTitle != "" {
			batchPrompt += fmt.Sprintf("Group title: %s\n", options.GroupTitle)
		}
		batchPrompt += fmt.Sprintf("Parallel run: %d/%d\n\n%s", i, options.Parallel, prompt)
		req := &jules.CreateSessionRequest{
			Prompt:              batchPrompt,
			Title:               title,
			RequirePlanApproval: true,
			AutomationMode:      jules.AutomationMode(options.AutomationMode),
			SourceContext: &jules.SourceContext{
				Source: sourceName,
			},
		}
		if options.StartingBranch != "" {
			req.SourceContext.GithubRepoContext = &jules.GithubRepoContext{StartingBranch: options.StartingBranch}
		}

		session, err := julesClient.CreateSession(ctx, req)
		if err != nil {
			return fmt.Errorf("created %d/%d sessions before failure: %w", i-1, options.Parallel, err)
		}
		fmt.Printf("%d. %s", i, session.ID)
		if session.URL != "" {
			fmt.Printf(" - %s", session.URL)
		}
		fmt.Println()
	}

	return nil
}

func loadPromptFile(path string) (string, error) {
	info, err := os.Stat(path)
	if err != nil {
		return "", fmt.Errorf("failed to inspect prompt file: %w", err)
	}
	if info.IsDir() {
		return "", fmt.Errorf("prompt file is a directory: %s", path)
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return "", fmt.Errorf("failed to read prompt file: %w", err)
	}
	prompt := strings.TrimSpace(string(data))
	if prompt == "" {
		return "", fmt.Errorf("prompt file is empty: %s", path)
	}
	return prompt, nil
}

func loadPromptArgument(value string) (string, error) {
	info, err := os.Stat(value)
	if err == nil {
		if info.IsDir() {
			return "", fmt.Errorf("task file is a directory: %s", value)
		}
		data, err := os.ReadFile(value)
		if err != nil {
			return "", fmt.Errorf("failed to read task file: %w", err)
		}
		prompt := strings.TrimSpace(string(data))
		if prompt == "" {
			return "", fmt.Errorf("task file is empty: %s", value)
		}
		return prompt, nil
	}
	if !os.IsNotExist(err) {
		return "", fmt.Errorf("failed to inspect task file: %w", err)
	}
	if strings.TrimSpace(value) == "" {
		return "", fmt.Errorf("prompt cannot be empty")
	}
	return value, nil
}

func shortSessionID(sessionID string) string {
	if len(sessionID) <= 12 {
		return sessionID
	}
	return sessionID[:12]
}

func truncate(value string, limit int) string {
	value = strings.TrimSpace(value)
	if len(value) <= limit {
		return value
	}
	return value[:limit] + "..."
}

// downloadSessionArtifacts downloads all artifacts from all activities in a session
func downloadSessionArtifacts(cfg *config.Config, sessionID string, outputDir string) error {
	// Initialize Jules client
	julesClient := jules.NewClient(cfg.Jules.APIKey, jules.WithBaseURL(cfg.Jules.BaseURL), jules.WithTimeout(cfg.Jules.Timeout), jules.WithRetryAttempts(cfg.Jules.RetryAttempts))

	ctx := context.Background()

	fmt.Printf("📥 Downloading artifacts from session: %s\n", sessionID)
	fmt.Printf("📁 Output directory: %s\n", outputDir)
	fmt.Println(strings.Repeat("=", 60))

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
	fmt.Println(strings.Repeat("=", 60))

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
	fmt.Println(strings.Repeat("=", 60))

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
	fmt.Println(strings.Repeat("=", 60))

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
