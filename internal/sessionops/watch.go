package sessionops

import (
	"context"
	"fmt"
	"time"

	"github.com/SamyRai/go-jules"
	"github.com/SamyRai/juleson/internal/julesops"
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

type WatchUpdateType string

const (
	WatchUpdateProgress         WatchUpdateType = "progress"
	WatchUpdateNeedsUserAction  WatchUpdateType = "needs_user_action"
	WatchUpdateTerminalSuccess  WatchUpdateType = "terminal_success"
	WatchUpdateTerminalFailure  WatchUpdateType = "terminal_failure"
	WatchUpdateOutputsAvailable WatchUpdateType = "outputs_available"
	WatchUpdateAgentMessage     WatchUpdateType = "agent_message"
)

type WakePolicy string

const (
	WakePolicyActionable WakePolicy = "actionable"
	WakePolicyAnyStatus  WakePolicy = "any-status"
	WakePolicyTerminal   WakePolicy = "terminal"
	WakePolicyNone       WakePolicy = "none"
)

type WatchDecision struct {
	Kind WatchDecisionKind
	Stop bool
}

type WatchWake struct {
	UpdateType WatchUpdateType
	ShouldWake bool
	WakeReason string
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
	session, err := client.Sessions().Get(ctx, sessionID)
	if err != nil {
		return WatchSnapshot{}, fmt.Errorf("failed to get session: %w", err)
	}

	snapshot := WatchSnapshot{
		Session:    session,
		NextCursor: cursor,
	}

	if options.FetchActivities {
		activities, err := client.Activities().ListSince(ctx, sessionID, cursor, 25)
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

func ParseWakePolicy(value string) (WakePolicy, error) {
	switch WakePolicy(value) {
	case "", WakePolicyActionable:
		return WakePolicyActionable, nil
	case WakePolicyAnyStatus:
		return WakePolicyAnyStatus, nil
	case WakePolicyTerminal:
		return WakePolicyTerminal, nil
	case WakePolicyNone:
		return WakePolicyNone, nil
	default:
		return "", fmt.Errorf("invalid wake policy %q", value)
	}
}

func WatchUpdateTypeForDecision(decision WatchDecision) WatchUpdateType {
	switch decision.Kind {
	case WatchDecisionNeedsUserAction:
		return WatchUpdateNeedsUserAction
	case WatchDecisionFailed:
		return WatchUpdateTerminalFailure
	case WatchDecisionCompletedDeliverableCheckFailed, WatchDecisionCompletedNoDeliverables, WatchDecisionCompletedWithDeliverables:
		return WatchUpdateTerminalSuccess
	case WatchDecisionOutputs:
		return WatchUpdateOutputsAvailable
	default:
		return WatchUpdateProgress
	}
}

func EvaluateWatchWake(policy WakePolicy, updateType WatchUpdateType, stateChanged bool) WatchWake {
	wake := WatchWake{UpdateType: updateType}
	if updateType == WatchUpdateAgentMessage {
		wake.ShouldWake = true
		wake.WakeReason = string(WatchUpdateAgentMessage)
		return wake
	}

	switch policy {
	case WakePolicyAnyStatus:
		if stateChanged {
			wake.ShouldWake = true
			wake.WakeReason = "status_change"
			return wake
		}
		if isActionableUpdate(updateType) {
			wake.ShouldWake = true
			wake.WakeReason = DefaultWatchWakeReasonForUpdate(updateType)
		}
	case WakePolicyTerminal:
		if updateType == WatchUpdateTerminalSuccess || updateType == WatchUpdateTerminalFailure {
			wake.ShouldWake = true
			wake.WakeReason = DefaultWatchWakeReasonForUpdate(updateType)
		}
	case WakePolicyNone:
		return wake
	default:
		if isActionableUpdate(updateType) {
			wake.ShouldWake = true
			wake.WakeReason = DefaultWatchWakeReasonForUpdate(updateType)
		}
	}
	return wake
}

func DefaultWatchWakeReasonForUpdate(updateType WatchUpdateType) string {
	switch updateType {
	case WatchUpdateNeedsUserAction:
		return "user_action"
	case WatchUpdateTerminalSuccess, WatchUpdateTerminalFailure:
		return "terminal_state"
	case WatchUpdateOutputsAvailable:
		return "session_outputs"
	case WatchUpdateAgentMessage:
		return string(WatchUpdateAgentMessage)
	default:
		return ""
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

func isActionableUpdate(updateType WatchUpdateType) bool {
	switch updateType {
	case WatchUpdateNeedsUserAction, WatchUpdateTerminalSuccess, WatchUpdateTerminalFailure, WatchUpdateOutputsAvailable:
		return true
	default:
		return false
	}
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
