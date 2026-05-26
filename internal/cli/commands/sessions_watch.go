package commands

import (
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/SamyRai/go-jules"
	"github.com/SamyRai/juleson/internal/config"
	"github.com/SamyRai/juleson/internal/presentation"
	"github.com/SamyRai/juleson/internal/sessionops"
)

func watchSession(cfg *config.Config, sessionID, intervalValue, timeoutValue string, followActivities bool, sinceValue, cursorOutput, initialState string, wakeOnStatusChange, wakeOnAgentMessage bool, wakePolicyValue string) error {
	julesClient := newJulesClient(cfg)

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
	wakePolicy, err := cliWakePolicy(wakeOnStatusChange, wakePolicyValue)
	if err != nil {
		return err
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
		stateChanged := false
		if !hasStateBaseline {
			baselineState = update.State
			hasStateBaseline = true
		} else if update.State != baselineState {
			stateChanged = true
		}
		wake := sessionops.EvaluateWatchWake(wakePolicy, update.UpdateType, stateChanged)
		if wakeOnAgentMessage {
			if !hasActivityBaseline {
				hasActivityBaseline = true
			} else if update.HasJulesAgentMessage {
				wake = sessionops.EvaluateWatchWake(wakePolicy, sessionops.WatchUpdateAgentMessage, false)
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
		if wake.ShouldWake {
			switch wake.WakeReason {
			case "status_change":
				fmt.Printf("Wake reason: session state changed from %s to %s.\n", baselineState, update.State)
			case string(sessionops.WatchUpdateAgentMessage):
				fmt.Printf("Wake reason: Jules sent a new message.\n")
			default:
				fmt.Printf("Wake reason: %s.\n", wake.WakeReason)
			}
			if update.NextAction != "" {
				fmt.Println(update.NextAction)
			}
			if !cursor.IsZero() {
				fmt.Printf("Next activity cursor: %s\n", cursor.Format(time.RFC3339Nano))
			}
			return nil
		}
		if stateChanged {
			baselineState = update.State
		}
		select {
		case <-ctx.Done():
			return fmt.Errorf("timeout watching session after %s", timeout)
		case <-ticker.C:
		}
	}
}

type sessionWatchUpdate struct {
	UpdateType           sessionops.WatchUpdateType
	NextAction           string
	NextCursor           time.Time
	State                jules.SessionState
	HasJulesAgentMessage bool
}

func cliWakePolicy(wakeOnStatusChange bool, wakePolicyValue string) (sessionops.WakePolicy, error) {
	if wakeOnStatusChange {
		return sessionops.WakePolicyAnyStatus, nil
	}
	return sessionops.ParseWakePolicy(wakePolicyValue)
}

func printSessionWatchUpdate(ctx context.Context, client *jules.Client, sessionID string, followActivities bool, detectAgentMessage bool, seenActivities map[string]bool, cursor time.Time) (sessionWatchUpdate, error) {
	snapshot, err := sessionops.CurrentWatchSnapshot(ctx, client, sessionID, cursor, sessionops.CurrentWatchOptions{
		FetchActivities: followActivities || detectAgentMessage,
	})
	if err != nil {
		return sessionWatchUpdate{}, err
	}
	session := snapshot.Session
	statusText := presentation.SessionStatusText(string(session.State))

	update := sessionWatchUpdate{
		UpdateType:           sessionops.WatchUpdateTypeForDecision(snapshot.Decision),
		NextAction:           cliNextAction(snapshot, sessionID, statusText),
		NextCursor:           snapshot.NextCursor,
		State:                session.State,
		HasJulesAgentMessage: snapshot.HasJulesAgentMessage,
	}

	statusIcon := presentation.SessionStatusIcon(string(session.State))
	fmt.Printf("%s %s %s [%s]", time.Now().Format(time.RFC3339), statusIcon, session.State, update.UpdateType)
	if session.Title != "" {
		fmt.Printf(" - %s", session.Title)
	}
	fmt.Println()

	if followActivities || detectAgentMessage {
		if snapshot.ActivityError != nil {
			fmt.Printf("⚠️  Could not fetch activities: %v\n", snapshot.ActivityError)
		} else {
			for i := len(snapshot.Activities) - 1; i >= 0; i-- {
				activity := snapshot.Activities[i]
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

	return update, nil
}

func cliNextAction(snapshot sessionops.WatchSnapshot, sessionID, statusText string) string {
	switch snapshot.Decision.Kind {
	case sessionops.WatchDecisionNeedsUserAction:
		return fmt.Sprintf("Next action: %s. Use 'juleson sessions get %s' to inspect, then approve or send feedback.", statusText, sessionID)
	case sessionops.WatchDecisionFailed:
		return fmt.Sprintf("Next action: inspect failure details with 'juleson sessions get %s'.", sessionID)
	case sessionops.WatchDecisionCompletedDeliverableCheckFailed:
		return fmt.Sprintf("⚠️  Could not check deliverables: %v\nNext action: inspect activities with 'juleson sessions artifacts list %s', then preview changes with 'juleson sessions apply %s <project-path>'.", snapshot.DeliverablesError, sessionID, sessionID)
	case sessionops.WatchDecisionCompletedNoDeliverables:
		return fmt.Sprintf("Next action: no retrievable deliverable was produced. Inspect activities with 'juleson sessions artifacts list %s' or create a follow-up session.", sessionID)
	case sessionops.WatchDecisionCompletedWithDeliverables:
		nextAction := fmt.Sprintf("Next action: preview changes with 'juleson sessions apply %s <project-path>'.", sessionID)
		if snapshot.Session != nil && len(snapshot.Session.Outputs) > 0 {
			nextAction += fmt.Sprintf("\nNext output action: inspect outputs with 'juleson sessions outputs %s'.", sessionID)
		}
		return nextAction
	case sessionops.WatchDecisionOutputs:
		return fmt.Sprintf("Next action: inspect outputs with 'juleson sessions outputs %s'.", sessionID)
	default:
		return ""
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
