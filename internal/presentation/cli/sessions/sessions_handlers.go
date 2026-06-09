package sessions

import (
	"bufio"
	"context"
	"fmt"
	"github.com/SamyRai/juleson/internal/presentation/cli/core"
	"os"
	"strings"

	"github.com/SamyRai/go-jules"
	"github.com/SamyRai/juleson/internal/config"
	julessessions "github.com/SamyRai/juleson/internal/jules/sessions"
	"github.com/SamyRai/juleson/internal/presentation/views"
)

type CreateSessionOptions struct {
	PromptFile          string
	Title               string
	StartingBranch      string
	AutomationMode      string
	NoSource            bool
	RequirePlanApproval bool
	WithIntel           bool
}

type BatchSessionOptions struct {
	Title          string
	BatchID        string
	GroupTitle     string
	StartingBranch string
	AutomationMode string
	Parallel       int
}

type ApplySessionOptions struct {
	ActivityID        string
	ArtifactIndex     int
	Confirm           bool
	AllowDirty        bool
	HasArtifactIndex  bool
	AllowBaseMismatch bool
}

func approveSessionPlan(cfg *config.Config, sessionID string) error {
	julesClient := core.NewJulesClient(cfg)
	ctx := context.Background()

	// Check if session is explicitly waiting for feedback
	session, err := julesClient.Sessions().Get(ctx, sessionID)
	if err == nil && session.State == jules.SessionStateAwaitingUserFeedback {
		fmt.Println("💡 Warning: This session is in AWAITING_USER_FEEDBACK state. The agent requires a direct message response.")
		fmt.Printf("💡 If you meant to reply to a question, use: juleson sessions message %s \"Your reply\"\n\n", sessionID)
	}

	fmt.Printf("✅ Approving plan for session: %s\n", sessionID)

	err = julesClient.Sessions().ApprovePlan(ctx, sessionID)
	if err != nil {
		return fmt.Errorf("failed to approve plan: %w", err)
	}

	fmt.Println("✅ Plan approved successfully!")
	fmt.Printf("💡 Jules will now execute the approved plan. Monitor at: https://jules.google.com/session/%s\n", sessionID)

	return nil
}
func deleteSession(cfg *config.Config, sessionID string, force bool) error {
	julesClient := core.NewJulesClient(cfg)

	if !force {
		fmt.Printf("Type the session ID to confirm deletion (%s): ", sessionID)
		scanner := bufio.NewScanner(os.Stdin)
		if !scanner.Scan() {
			if err := scanner.Err(); err != nil {
				return fmt.Errorf("failed to read confirmation: %w", err)
			}
			return fmt.Errorf("session deletion canceled")
		}
		if strings.TrimSpace(scanner.Text()) != sessionID {
			return fmt.Errorf("session deletion canceled")
		}
	}

	if err := julesClient.Sessions().Delete(context.Background(), sessionID); err != nil {
		return fmt.Errorf("failed to delete session: %w", err)
	}

	fmt.Printf("✅ Deleted session: %s\n", sessionID)
	return nil
}
func listSessions(cfg *config.Config) error {
	julesClient := core.NewJulesClient(cfg)

	fmt.Println("🔍 Listing Jules sessions...")
	fmt.Println("============================")

	response, err := julesClient.Sessions().List(context.Background(), &jules.ListSessionsOptions{PageSize: 50})
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
		statusText := views.SessionStatusText(string(session.State))
		statusIcon := views.SessionStatusIcon(string(session.State))
		fmt.Printf("   %s %s\n", statusIcon, statusText)
		fmt.Println()
	}

	return nil
}

func showSessionStatus(cfg *config.Config) error {
	julesClient := core.NewJulesClient(cfg)

	fmt.Println("📊 Jules Session Status")
	fmt.Println("=======================")

	response, err := julesClient.Sessions().List(context.Background(), &jules.ListSessionsOptions{PageSize: 100})
	if err != nil {
		return fmt.Errorf("failed to get session status: %w", err)
	}

	summary := julessessions.SummarizeSessions(response.Sessions, 5)

	if summary.TotalSessions == 0 {
		fmt.Println("📭 No sessions found.")
		return nil
	}

	fmt.Printf("Total Sessions: %d\n\n", summary.TotalSessions)

	fmt.Println("Session States:")
	for state, count := range summary.StateBreakdown {
		percentage := float64(count) / float64(summary.TotalSessions) * 100
		icon := views.SessionStatusIcon(state)
		fmt.Printf("  %s %s: %d (%.1f%%)\n", icon, state, count, percentage)
	}

	if summary.ActiveSessions > 0 {
		fmt.Printf("\n⚠️  %d session(s) are currently active/running\n", summary.ActiveSessions)
	} else {
		fmt.Println("\n✅ No active sessions currently running")
	}
	if summary.UserActionSessions > 0 {
		fmt.Printf("⏸  %d session(s) need user action\n", summary.UserActionSessions)
	}

	if len(summary.RecentSessions) > 0 {
		fmt.Println("\n🕒 Recent Sessions:")
		for _, session := range summary.RecentSessions {
			statusIcon := views.SessionStatusIcon(string(session.State))
			fmt.Printf("  %s %s - %s (%s)\n", statusIcon, shortSessionID(session.ID), session.Title, session.State)
		}
	}

	return nil
}

func getSessionDetails(cfg *config.Config, sessionID string) error {
	julesClient := core.NewJulesClient(cfg)

	ctx := context.Background()

	fmt.Printf("🔍 Fetching session details for: %s\n", sessionID)
	fmt.Println(strings.Repeat("=", 60))

	// Get session details
	session, err := julesClient.Sessions().Get(ctx, sessionID)
	if err != nil {
		return fmt.Errorf("failed to get session: %w", err)
	}

	// Display session information
	fmt.Printf("\n📊 Session Information\n")
	fmt.Printf("ID: %s\n", session.ID)
	fmt.Printf("Title: %s\n", session.Title)
	fmt.Printf("State: %s %s\n", views.SessionStatusIcon(string(session.State)), session.State)
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
	outputs := julessessions.DocumentedOutputs(session)
	if len(outputs) > 0 {
		fmt.Printf("\n📤 Outputs:\n")
		for i, output := range outputs {
			fmt.Printf("  %d. Pull Request:\n", i+1)
			fmt.Printf("     URL: %s\n", output.PullRequest.URL)
			fmt.Printf("     Title: %s\n", output.PullRequest.Title)
			if output.PullRequest.Description != "" {
				fmt.Printf("     Description: %s\n", output.PullRequest.Description)
			}
		}
	}

	// Get activities
	fmt.Printf("\n📋 Activities:\n")
	response, err := julesClient.Activities().List(ctx, sessionID, &jules.ListActivitiesOptions{PageSize: 100})
	if err != nil {
		fmt.Printf("⚠️  Could not fetch activities: %v\n", err)
		return nil
	}
	activities := response.Activities

	if len(activities) == 0 {
		fmt.Println("  No activities yet - session is still initializing")
		return nil
	}

	fmt.Printf("  Found %d activities\n\n", len(activities))

	for i, activity := range activities {
		originator := "❓"
		switch activity.Originator {
		case "agent":
			originator = "🤖"
		case "user":
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

		if activity.AgentMessaged != nil {
			fmt.Printf("   💬 Agent Message: %s\n", activity.AgentMessaged.AgentMessage)
		}

		if activity.UserMessaged != nil {
			fmt.Printf("   💬 User Message: %s\n", activity.UserMessaged.UserMessage)
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

	fmt.Printf("💡 Use `juleson sessions plans %s` to inspect full generated plans.\n", sessionID)
	fmt.Printf("💡 View full session at: %s\n", session.URL)

	return nil
}

func sendSessionMessage(cfg *config.Config, sessionID string, message string) error {
	julesClient := core.NewJulesClient(cfg)

	ctx := context.Background()

	fmt.Printf("📤 Sending message to session: %s\n", sessionID)
	fmt.Printf("Message: %s\n\n", message)

	req := &jules.SendMessageRequest{
		Prompt: message,
	}

	err := julesClient.Sessions().SendMessage(ctx, sessionID, req)
	if err != nil {
		return fmt.Errorf("failed to send message: %w", err)
	}

	fmt.Println("✅ Message sent successfully!")
	fmt.Println("💡 Jules will process your message and respond with activities.")
	fmt.Printf("💡 Monitor at: https://jules.google.com/session/%s\n", sessionID)

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
