package sessionops

import (
	"fmt"

	"github.com/SamyRai/go-jules"
)

type StatusSummary struct {
	TotalSessions      int
	StateBreakdown     map[string]int
	ActiveSessions     int
	UserActionSessions int
	RecentSessions     []jules.Session
	Summary            string
}

func SummarizeSessions(sessions []jules.Session, recentLimit int) StatusSummary {
	if sessions == nil {
		sessions = []jules.Session{}
	}
	stateBreakdown := make(map[string]int)
	for _, session := range sessions {
		stateBreakdown[string(session.State)]++
	}

	if recentLimit < 0 {
		recentLimit = 0
	}
	if len(sessions) < recentLimit {
		recentLimit = len(sessions)
	}

	activeSessions := stateBreakdown[string(jules.SessionStateInProgress)] +
		stateBreakdown[string(jules.SessionStatePlanning)] +
		stateBreakdown[string(jules.SessionStateQueued)]
	userActionSessions := stateBreakdown[string(jules.SessionStateAwaitingPlanApproval)] +
		stateBreakdown[string(jules.SessionStateAwaitingUserFeedback)]

	return StatusSummary{
		TotalSessions:      len(sessions),
		StateBreakdown:     stateBreakdown,
		ActiveSessions:     activeSessions,
		UserActionSessions: userActionSessions,
		RecentSessions:     append([]jules.Session(nil), sessions[:recentLimit]...),
		Summary:            fmt.Sprintf("Found %d total sessions with %d currently active and %d needing user action", len(sessions), activeSessions, userActionSessions),
	}
}

func DocumentedOutputs(session *jules.Session) []jules.Output {
	if session == nil {
		return []jules.Output{}
	}
	outputs := make([]jules.Output, 0, len(session.Outputs))
	for _, output := range session.Outputs {
		if output.PullRequest != nil {
			outputs = append(outputs, output)
		}
	}
	return outputs
}
