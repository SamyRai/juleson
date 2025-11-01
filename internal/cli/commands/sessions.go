package commands

import (
	"context"
	"fmt"

	"github.com/SamyRai/juleson/internal/config"
	"github.com/SamyRai/juleson/internal/jules"

	"github.com/spf13/cobra"
)

// getSessionStatusIcon returns the appropriate icon for a session state
func getSessionStatusIcon(state string) string {
	switch state {
	case "IN_PROGRESS", "PLANNING":
		return "⚡"
	case "COMPLETED":
		return "✅"
	case "FAILED":
		return "❌"
	default:
		return "📋"
	}
}

// getSessionStatusText returns the status text for a session state
func getSessionStatusText(state string) string {
	switch state {
	case "IN_PROGRESS", "PLANNING":
		return "ACTIVE"
	case "COMPLETED":
		return "COMPLETED"
	case "FAILED":
		return "FAILED"
	default:
		return state
	}
}

// NewSessionsCommand creates the sessions command
func NewSessionsCommand(cfg *config.Config) *cobra.Command {
	sessionsCmd := &cobra.Command{
		Use:   "sessions",
		Short: "Manage Jules sessions",
		Long:  "List, monitor, and manage Jules AI coding sessions",
	}

	// List sessions
	sessionsCmd.AddCommand(&cobra.Command{
		Use:   "list",
		Short: "List all sessions",
		Long:  "List all Jules sessions with their current status",
		RunE: func(cmd *cobra.Command, args []string) error {
			return listSessions(cfg)
		},
	})

	// Show session status
	sessionsCmd.AddCommand(&cobra.Command{
		Use:   "status",
		Short: "Show session status summary",
		Long:  "Show a summary of current session statuses",
		RunE: func(cmd *cobra.Command, args []string) error {
			return showSessionStatus(cfg)
		},
	})

	return sessionsCmd
}

// listSessions lists all Jules sessions
func listSessions(cfg *config.Config) error {
	// Initialize Jules client
	julesClient := jules.NewClient(
		cfg.Jules.APIKey,
		cfg.Jules.BaseURL,
		cfg.Jules.Timeout,
		cfg.Jules.RetryAttempts,
	)

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
		if session.UpdateTime != "" {
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
	julesClient := jules.NewClient(
		cfg.Jules.APIKey,
		cfg.Jules.BaseURL,
		cfg.Jules.Timeout,
		cfg.Jules.RetryAttempts,
	)

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
		stateCounts[session.State]++
	}

	fmt.Printf("Total Sessions: %d\n\n", totalSessions)

	// Display state breakdown
	fmt.Println("Session States:")
	for state, count := range stateCounts {
		percentage := float64(count) / float64(totalSessions) * 100
		icon := getSessionStatusIcon(state)
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
