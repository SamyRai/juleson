package presentation

import (
	"fmt"
	"strings"
	"time"
)

type SessionView struct {
	ID                  string
	Title               string
	State               string
	CreateTime          time.Time
	UpdateTime          time.Time
	Source              string
	RequirePlanApproval bool
	AutomationMode      string
	OutputCount         int
}

// SessionFormatter formats session information
type SessionFormatter struct{}

// NewSessionFormatter creates a new session formatter
func NewSessionFormatter() *SessionFormatter {
	return &SessionFormatter{}
}

// FormatList displays a list of sessions
func (f *SessionFormatter) FormatList(sessions []SessionView) string {
	var sb strings.Builder

	sb.WriteString("🔍 Listing Jules sessions...\n")
	sb.WriteString("============================\n")

	if len(sessions) == 0 {
		sb.WriteString("📭 No sessions found.\n")
		return sb.String()
	}

	sb.WriteString(fmt.Sprintf("📊 Found %d session(s):\n\n", len(sessions)))

	for i, session := range sessions {
		sb.WriteString(fmt.Sprintf("%d. Session: %s\n", i+1, session.ID))
		sb.WriteString(fmt.Sprintf("   Title: %s\n", session.Title))
		sb.WriteString(fmt.Sprintf("   State: %s\n", session.State))
		sb.WriteString(fmt.Sprintf("   Created: %s\n", session.CreateTime))
		if !session.UpdateTime.IsZero() {
			sb.WriteString(fmt.Sprintf("   Updated: %s\n", session.UpdateTime))
		}
		if session.Source != "" {
			sb.WriteString(fmt.Sprintf("   Source: %s\n", session.Source))
		}
		if session.RequirePlanApproval {
			sb.WriteString("   Plan Approval Required: Yes\n")
		}
		if session.AutomationMode != "" {
			sb.WriteString(fmt.Sprintf("   Automation Mode: %s\n", session.AutomationMode))
		}
		if session.OutputCount > 0 {
			sb.WriteString(fmt.Sprintf("   Outputs: %d\n", session.OutputCount))
		}

		// Status indicators
		switch SessionStatusText(session.State) {
		case "ACTIVE":
			sb.WriteString("   ⚡ ACTIVE\n")
		case "COMPLETED":
			sb.WriteString("   ✅ COMPLETED\n")
		case "FAILED":
			sb.WriteString("   ❌ FAILED\n")
		}
		sb.WriteString("\n")
	}

	return sb.String()
}

// FormatStatus displays session status summary
func (f *SessionFormatter) FormatStatus(sessions []SessionView) string {
	var sb strings.Builder

	sb.WriteString("📊 Jules Session Status\n")
	sb.WriteString("=======================\n")

	totalSessions := len(sessions)
	if totalSessions == 0 {
		sb.WriteString("📭 No sessions found.\n")
		return sb.String()
	}

	// Count sessions by state
	stateCounts := make(map[string]int)
	for _, session := range sessions {
		stateCounts[session.State]++
	}

	sb.WriteString(fmt.Sprintf("Total Sessions: %d\n\n", totalSessions))

	// Display state breakdown
	sb.WriteString("Session States:\n")
	for state, count := range stateCounts {
		percentage := float64(count) / float64(totalSessions) * 100
		sb.WriteString(fmt.Sprintf("  %s %s: %d (%.1f%%)\n", SessionStatusIcon(state), state, count, percentage))
	}

	// Active sessions summary
	activeCount := stateCounts["IN_PROGRESS"] + stateCounts["PLANNING"] + stateCounts["QUEUED"]
	if activeCount > 0 {
		sb.WriteString(fmt.Sprintf("\n⚠️  %d session(s) are currently active/running\n", activeCount))
	} else {
		sb.WriteString("\n✅ No active sessions currently running\n")
	}

	// Recent sessions (last 5)
	if totalSessions > 0 {
		sb.WriteString("\n🕒 Recent Sessions:\n")
		recentCount := 5
		if totalSessions < recentCount {
			recentCount = totalSessions
		}

		for i := 0; i < recentCount; i++ {
			session := sessions[i]
			sessionIDShort := session.ID
			if len(sessionIDShort) > 12 {
				sessionIDShort = sessionIDShort[:12]
			}
			sb.WriteString(fmt.Sprintf("  %s %s - %s (%s)\n", SessionStatusIcon(session.State), sessionIDShort, session.Title, session.State))
		}
	}

	return sb.String()
}

// SessionStatusIcon returns the presentation icon for a session state.
func SessionStatusIcon(state string) string {
	switch state {
	case "IN_PROGRESS", "PLANNING", "QUEUED":
		return "⚡"
	case "AWAITING_PLAN_APPROVAL", "AWAITING_USER_FEEDBACK":
		return "⏸"
	case "COMPLETED":
		return "✅"
	case "FAILED":
		return "❌"
	default:
		return "📋"
	}
}

// SessionStatusText returns the display status bucket for a session state.
func SessionStatusText(state string) string {
	switch state {
	case "IN_PROGRESS", "PLANNING", "QUEUED":
		return "ACTIVE"
	case "AWAITING_PLAN_APPROVAL", "AWAITING_USER_FEEDBACK":
		return "NEEDS_USER_ACTION"
	case "COMPLETED":
		return "COMPLETED"
	case "FAILED":
		return "FAILED"
	default:
		return state
	}
}
