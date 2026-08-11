package views

import (
	"fmt"
	"strings"
	"time"

	"github.com/SamyRai/juleson/internal/presentation/views/theme"
	"github.com/charmbracelet/lipgloss"
)

type SessionView struct {
	CreateTime          time.Time
	UpdateTime          time.Time
	ID                  string
	Title               string
	State               string
	Source              string
	AutomationMode      string
	OutputCount         int
	RequirePlanApproval bool
}

type SessionFormatter struct{}

func NewSessionFormatter() *SessionFormatter {
	return &SessionFormatter{}
}

func (f *SessionFormatter) FormatList(sessions []SessionView) string {
	var sb strings.Builder

	sb.WriteString(theme.HeaderStyle.Render("🔍 Listing Jules sessions...") + "\n\n")

	if len(sessions) == 0 {
		sb.WriteString(theme.MutedStyle.Render("📭 No sessions found.") + "\n")
		return sb.String()
	}

	f.formatCount(&sb, len(sessions))

	for i, session := range sessions {
		f.formatSessionItem(&sb, i, session)
	}

	return sb.String()
}

func (f *SessionFormatter) formatCount(sb *strings.Builder, count int) {
	sb.WriteString(theme.InfoStyle.Render(fmt.Sprintf("📊 Found %d session(s):", count)) + "\n\n")
}

func (f *SessionFormatter) formatSessionItem(sb *strings.Builder, i int, session SessionView) {
	statusColor := theme.MutedColor
	statusIcon := "📋"
	switch SessionStatusText(session.State) {
	case "ACTIVE":
		statusColor = theme.SuccessColor
		statusIcon = "⚡"
	case "COMPLETED":
		statusColor = theme.InfoColor
		statusIcon = "✅"
	case "FAILED":
		statusColor = theme.ErrorColor
		statusIcon = "❌"
	}

	style := lipgloss.NewStyle().Foreground(statusColor).Bold(true)

	fmt.Fprintf(sb, "%s %s\n", style.Render(fmt.Sprintf("%d. Session:", i+1)), session.ID)
	fmt.Fprintf(sb, "   Title: %s\n", session.Title)
	fmt.Fprintf(sb, "   State: %s %s\n", statusIcon, session.State)
	fmt.Fprintf(sb, "   Created: %s\n", session.CreateTime)
	if !session.UpdateTime.IsZero() {
		fmt.Fprintf(sb, "   Updated: %s\n", session.UpdateTime)
	}
	if session.Source != "" {
		fmt.Fprintf(sb, "   Source: %s\n", session.Source)
	}
	if session.RequirePlanApproval {
		sb.WriteString(theme.WarnStyle.Render("   Plan Approval Required: Yes") + "\n")
	}
	if session.AutomationMode != "" {
		fmt.Fprintf(sb, "   Automation Mode: %s\n", session.AutomationMode)
	}
	if session.OutputCount > 0 {
		fmt.Fprintf(sb, "   Outputs: %d\n", session.OutputCount)
	}
	sb.WriteString("\n")
}

func (f *SessionFormatter) FormatStatus(sessions []SessionView) string {
	var sb strings.Builder

	sb.WriteString(theme.HeaderStyle.Render("📊 Jules Session Status") + "\n\n")

	totalSessions := len(sessions)
	if totalSessions == 0 {
		sb.WriteString(theme.MutedStyle.Render("📭 No sessions found.") + "\n")
		return sb.String()
	}

	stateCounts := make(map[string]int)
	for _, session := range sessions {
		stateCounts[session.State]++
	}

	fmt.Fprintf(&sb, "Total Sessions: %s\n\n", theme.InfoStyle.Render(fmt.Sprintf("%d", totalSessions)))
	sb.WriteString("Session States:\n")

	for state, count := range stateCounts {
		percentage := float64(count) / float64(totalSessions) * 100
		fmt.Fprintf(&sb, "  %s %s: %d (%.1f%%)\n", SessionStatusIcon(state), state, count, percentage)
	}

	activeCount := stateCounts["IN_PROGRESS"] + stateCounts["PLANNING"] + stateCounts["QUEUED"]
	if activeCount > 0 {
		fmt.Fprintf(&sb, "\n%s\n", theme.WarnStyle.Render(fmt.Sprintf("⚠️  %d session(s) are currently active/running", activeCount)))
	} else {
		fmt.Fprintf(&sb, "\n%s\n", theme.SuccessStyle.Render("✅ No active sessions currently running"))
	}

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
			fmt.Fprintf(&sb, "  %s %s - %s (%s)\n", SessionStatusIcon(session.State), sessionIDShort, session.Title, session.State)
		}
	}

	return sb.String()
}

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
