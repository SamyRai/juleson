package sessionops

import (
	"context"
	"fmt"
	"time"

	"github.com/SamyRai/juleson/internal/julesops"
	"github.com/SamyRai/juleson/pkg/jules"
)

type WatchDecisionKind string

const (
	WatchDecisionContinue                        WatchDecisionKind = "continue"
	WatchDecisionNeedsUserAction                 WatchDecisionKind = "needs_user_action"
	WatchDecisionFailed                          WatchDecisionKind = "failed"
	WatchDecisionCompletedDeliverableCheckFailed WatchDecisionKind = "completed_deliverable_check_failed"
	WatchDecisionCompletedNoDeliverables         WatchDecisionKind = "completed_no_deliverables"
	WatchDecisionCompletedWithDeliverables       WatchDecisionKind = "completed_with_deliverables"
	WatchDecisionOutputs                         WatchDecisionKind = "outputs"
)

type WatchDecision struct {
	Kind WatchDecisionKind
	Stop bool
}

type CurrentWatchOptions struct {
	FetchActivities bool
}

type WatchSnapshot struct {
	Session              *jules.Session
	Activities           []jules.Activity
	ActivityError        error
	DeliverablesError    error
	NextCursor           time.Time
	HasJulesAgentMessage bool
	Decision             WatchDecision
}

func CurrentWatchSnapshot(ctx context.Context, client *jules.Client, sessionID string, cursor time.Time, options CurrentWatchOptions) (WatchSnapshot, error) {
	session, err := client.GetSession(ctx, sessionID)
	if err != nil {
		return WatchSnapshot{}, fmt.Errorf("failed to get session: %w", err)
	}

	snapshot := WatchSnapshot{
		Session:    session,
		NextCursor: cursor,
	}

	if options.FetchActivities {
		activities, err := client.ListActivitiesSince(ctx, sessionID, cursor, 25)
		if err != nil {
			snapshot.ActivityError = err
		} else {
			snapshot.Activities = activities
			nextCursor := jules.ActivityCursor(activities)
			if nextCursor.After(cursor) {
				snapshot.NextCursor = nextCursor
			}
			snapshot.HasJulesAgentMessage = HasJulesAgentMessageAfter(activities, cursor)
		}
	}

	var hasDeliverables *bool
	var deliverablesErr error
	if session.State == jules.SessionStateCompleted {
		value, err := julesops.SessionHasDeliverables(ctx, client, session)
		if err != nil {
			deliverablesErr = err
			snapshot.DeliverablesError = err
		} else {
			hasDeliverables = &value
		}
	}
	snapshot.Decision = EvaluateWatchDecision(session, hasDeliverables, deliverablesErr)

	return snapshot, nil
}

func EvaluateWatchDecision(session *jules.Session, hasDeliverables *bool, deliverablesErr error) WatchDecision {
	if session == nil {
		return WatchDecision{Kind: WatchDecisionContinue}
	}
	switch {
	case session.State.NeedsUserAction():
		return WatchDecision{Kind: WatchDecisionNeedsUserAction, Stop: true}
	case session.State == jules.SessionStateFailed:
		return WatchDecision{Kind: WatchDecisionFailed, Stop: true}
	case session.State == jules.SessionStateCompleted:
		if deliverablesErr != nil {
			return WatchDecision{Kind: WatchDecisionCompletedDeliverableCheckFailed, Stop: true}
		}
		if hasDeliverables != nil && !*hasDeliverables {
			return WatchDecision{Kind: WatchDecisionCompletedNoDeliverables, Stop: true}
		}
		return WatchDecision{Kind: WatchDecisionCompletedWithDeliverables, Stop: true}
	case len(session.Outputs) > 0:
		return WatchDecision{Kind: WatchDecisionOutputs, Stop: true}
	default:
		return WatchDecision{Kind: WatchDecisionContinue}
	}
}

func HasJulesAgentMessageAfter(activities []jules.Activity, cursor time.Time) bool {
	for _, activity := range activities {
		if activity.AgentMessaged == nil {
			continue
		}
		if cursor.IsZero() || activity.CreateTime.After(cursor) {
			return true
		}
	}
	return false
}

func DefaultWatchWakeReason(decision WatchDecision) string {
	switch decision.Kind {
	case WatchDecisionNeedsUserAction:
		return "user_action"
	case WatchDecisionFailed, WatchDecisionCompletedDeliverableCheckFailed, WatchDecisionCompletedNoDeliverables, WatchDecisionCompletedWithDeliverables:
		return "terminal_state"
	case WatchDecisionOutputs:
		return "session_outputs"
	default:
		return ""
	}
}

func MCPNextAction(snapshot WatchSnapshot) string {
	session := snapshot.Session
	if session == nil {
		return "session is still active; keep watching"
	}
	switch snapshot.Decision.Kind {
	case WatchDecisionNeedsUserAction:
		if session.State == jules.SessionStateAwaitingPlanApproval {
			return "inspect get_session_plans, then call approve_session_plan or send_session_message"
		}
		return "inspect recent activities, then call send_session_message"
	case WatchDecisionFailed:
		return "inspect get_session and list_session_activities for failure details"
	case WatchDecisionCompletedDeliverableCheckFailed:
		return "inspect list_session_activities and list_session_artifacts; deliverable check failed"
	case WatchDecisionCompletedNoDeliverables:
		return "no retrievable deliverable was produced; inspect list_session_artifacts or create a follow-up session"
	case WatchDecisionCompletedWithDeliverables:
		return "call preview_session_changes, then apply_session_patches with confirm_apply=true if acceptable"
	case WatchDecisionOutputs:
		return "call get_session_outputs to inspect created pull requests or other outputs"
	default:
		return "session is still active; keep watching"
	}
}
