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

func watchSession(cfg *config.Config, sessionID, intervalValue, timeoutValue string, followActivities bool, sinceValue, cursorOutput, initialState string, wakeOnStatusChange, wakeOnAgentMessage bool) error {
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
	snapshot, err := sessionops.CurrentWatchSnapshot(ctx, client, sessionID, cursor, sessionops.CurrentWatchOptions{
		FetchActivities: followActivities || detectAgentMessage,
	})
	if err != nil {
		return sessionWatchUpdate{}, err
	}
	session := snapshot.Session

	update := sessionWatchUpdate{
		NextCursor:           snapshot.NextCursor,
		State:                session.State,
		HasJulesAgentMessage: snapshot.HasJulesAgentMessage,
	}

	statusIcon := presentation.SessionStatusIcon(string(session.State))
	statusText := presentation.SessionStatusText(string(session.State))
	fmt.Printf("%s %s %s", time.Now().Format(time.RFC3339), statusIcon, session.State)
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

	switch snapshot.Decision.Kind {
	case sessionops.WatchDecisionNeedsUserAction:
		fmt.Printf("Next action: %s. Use 'juleson sessions get %s' to inspect, then approve or send feedback.\n", statusText, sessionID)
		update.Stop = true
		return update, nil
	case sessionops.WatchDecisionFailed:
		fmt.Printf("Next action: inspect failure details with 'juleson sessions get %s'.\n", sessionID)
		update.Stop = true
		return update, nil
	case sessionops.WatchDecisionCompletedDeliverableCheckFailed:
		fmt.Printf("⚠️  Could not check deliverables: %v\n", snapshot.DeliverablesError)
		fmt.Printf("Next action: inspect activities with 'juleson sessions artifacts list %s', then preview changes with 'juleson sessions apply %s <project-path>'.\n", sessionID, sessionID)
		update.Stop = true
		return update, nil
	case sessionops.WatchDecisionCompletedNoDeliverables:
		fmt.Printf("Next action: no retrievable deliverable was produced. Inspect activities with 'juleson sessions artifacts list %s' or create a follow-up session.\n", sessionID)
		update.Stop = true
		return update, nil
	case sessionops.WatchDecisionCompletedWithDeliverables:
		fmt.Printf("Next action: preview changes with 'juleson sessions apply %s <project-path>'.\n", sessionID)
		if len(session.Outputs) > 0 {
			fmt.Printf("Next output action: inspect outputs with 'juleson sessions outputs %s'.\n", sessionID)
		}
		update.Stop = true
		return update, nil
	case sessionops.WatchDecisionOutputs:
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
