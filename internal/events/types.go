package events

import (
	"encoding/json"
	"fmt"
	"time"
)

// Event represents a generic event in the system
type Event struct {
	ID        string                 `json:"id"`
	Type      EventType              `json:"type"`
	Topic     string                 `json:"topic"`
	Source    string                 `json:"source"`
	Timestamp time.Time              `json:"timestamp"`
	Data      interface{}            `json:"data"`
	Metadata  map[string]interface{} `json:"metadata,omitempty"`
	Priority  int                    `json:"priority,omitempty"`
	TTL       time.Duration          `json:"ttl,omitempty"`
	Retries   int                    `json:"retries,omitempty"`
}

// EventType represents the type of event
type EventType string

// System Events
const (
	// Agent Events
	EventAgentStarted      EventType = "agent.started"
	EventAgentStopped      EventType = "agent.stopped"
	EventAgentStateChanged EventType = "agent.state_changed"
	EventAgentError        EventType = "agent.error"
	EventAgentDecision     EventType = "agent.decision"
	EventAgentProgress     EventType = "agent.progress"

	// Task Events
	EventTaskCreated   EventType = "task.created"
	EventTaskStarted   EventType = "task.started"
	EventTaskCompleted EventType = "task.completed"
	EventTaskFailed    EventType = "task.failed"
	EventTaskRetrying  EventType = "task.retrying"

	// Session Events
	EventSessionCreated   EventType = "session.created"
	EventSessionUpdated   EventType = "session.updated"
	EventSessionCompleted EventType = "session.completed"
	EventSessionFailed    EventType = "session.failed"
	EventSessionCancelled EventType = "session.cancelled"

	// Activity Events
	EventActivityReceived  EventType = "activity.received"
	EventActivityProcessed EventType = "activity.processed"
	EventPlanGenerated     EventType = "plan.generated"
	EventPlanApproved      EventType = "plan.approved"

	// Tool Events
	EventToolInvoked   EventType = "tool.invoked"
	EventToolCompleted EventType = "tool.completed"
	EventToolFailed    EventType = "tool.failed"

	// Review Events
	EventReviewStarted   EventType = "review.started"
	EventReviewCompleted EventType = "review.completed"
	EventReviewRejected  EventType = "review.rejected"

	// Change Events
	EventChangeDetected EventType = "change.detected"
	EventChangeApplied  EventType = "change.applied"
	EventChangeReverted EventType = "change.reverted"

	// Orchestration Events
	EventWorkflowStarted   EventType = "workflow.started"
	EventWorkflowCompleted EventType = "workflow.completed"
	EventWorkflowFailed    EventType = "workflow.failed"
	EventPhaseStarted      EventType = "phase.started"
	EventPhaseCompleted    EventType = "phase.completed"

	// GitHub Events
	EventPRCreated EventType = "github.pr.created"
	EventPRMerged  EventType = "github.pr.merged"
	EventPRClosed  EventType = "github.pr.closed"

	// System Events
	EventSystemStarted  EventType = "system.started"
	EventSystemStopping EventType = "system.stopping"
	EventSystemError    EventType = "system.error"
)

// Event Topics for pub/sub
const (
	TopicAgent         = "agent"
	TopicSession       = "session"
	TopicTask          = "task"
	TopicActivity      = "activity"
	TopicTool          = "tool"
	TopicReview        = "review"
	TopicChange        = "change"
	TopicOrchestration = "orchestration"
	TopicGitHub        = "github"
	TopicSystem        = "system"
	TopicAll           = "*" // Subscribe to all events
)

// Event Data Structures

// AgentStateChangedData represents agent state change event data
type AgentStateChangedData struct {
	OldState string `json:"old_state"`
	NewState string `json:"new_state"`
	Reason   string `json:"reason,omitempty"`
	GoalID   string `json:"goal_id,omitempty"`
}

// AgentDecisionData represents agent decision event data
type AgentDecisionData struct {
	DecisionID   string   `json:"decision_id"`
	DecisionType string   `json:"decision_type"`
	Reasoning    string   `json:"reasoning"`
	Confidence   float64  `json:"confidence"`
	Action       string   `json:"action,omitempty"`
	Alternatives []string `json:"alternatives,omitempty"`
}

// AgentProgressData represents agent progress event data
type AgentProgressData struct {
	State          string  `json:"state"`
	CurrentTask    string  `json:"current_task"`
	CompletedTasks int     `json:"completed_tasks"`
	TotalTasks     int     `json:"total_tasks"`
	Progress       float64 `json:"progress"`
	Message        string  `json:"message"`
}

// TaskEventData represents task event data
type TaskEventData struct {
	TaskID   string                 `json:"task_id"`
	TaskName string                 `json:"task_name"`
	Status   string                 `json:"status"`
	Tool     string                 `json:"tool,omitempty"`
	Error    string                 `json:"error,omitempty"`
	Duration time.Duration          `json:"duration,omitempty"`
	Metadata map[string]interface{} `json:"metadata,omitempty"`
}

// SessionEventData represents session event data
type SessionEventData struct {
	SessionID string                 `json:"session_id"`
	State     string                 `json:"state"`
	Title     string                 `json:"title,omitempty"`
	SourceID  string                 `json:"source_id,omitempty"`
	Error     string                 `json:"error,omitempty"`
	URL       string                 `json:"url,omitempty"`
	Metadata  map[string]interface{} `json:"metadata,omitempty"`
}

// ActivityEventData represents activity event data
type ActivityEventData struct {
	SessionID    string                 `json:"session_id"`
	ActivityID   string                 `json:"activity_id"`
	ActivityType string                 `json:"activity_type"`
	Originator   string                 `json:"originator"`
	Description  string                 `json:"description,omitempty"`
	Artifacts    int                    `json:"artifacts"`
	Metadata     map[string]interface{} `json:"metadata,omitempty"`
}

// ToolEventData represents tool event data
type ToolEventData struct {
	ToolName   string                 `json:"tool_name"`
	Action     string                 `json:"action"`
	Parameters map[string]interface{} `json:"parameters,omitempty"`
	Success    bool                   `json:"success"`
	Error      string                 `json:"error,omitempty"`
	Duration   time.Duration          `json:"duration,omitempty"`
	Result     interface{}            `json:"result,omitempty"`
}

// ReviewEventData represents review event data
type ReviewEventData struct {
	ReviewID      string  `json:"review_id"`
	Decision      string  `json:"decision"`
	Score         float64 `json:"score"`
	CommentsCount int     `json:"comments_count"`
	Summary       string  `json:"summary"`
	ChangesCount  int     `json:"changes_count"`
}

// ChangeEventData represents change event data
type ChangeEventData struct {
	FilePath    string `json:"file_path"`
	ChangeType  string `json:"change_type"`
	Additions   int    `json:"additions"`
	Deletions   int    `json:"deletions"`
	Description string `json:"description,omitempty"`
}

// WorkflowEventData represents workflow event data
type WorkflowEventData struct {
	WorkflowName string        `json:"workflow_name"`
	Phase        int           `json:"phase,omitempty"`
	TotalPhases  int           `json:"total_phases,omitempty"`
	PhaseName    string        `json:"phase_name,omitempty"`
	Success      bool          `json:"success"`
	Error        string        `json:"error,omitempty"`
	Duration     time.Duration `json:"duration,omitempty"`
}

// GitHubEventData represents GitHub event data
type GitHubEventData struct {
	Repository string `json:"repository"`
	PRNumber   int    `json:"pr_number,omitempty"`
	PRURL      string `json:"pr_url,omitempty"`
	Action     string `json:"action"`
	Error      string `json:"error,omitempty"`
}

// NewEvent creates a new event with default values
func NewEvent(eventType EventType, source string, data interface{}) Event {
	return Event{
		ID:        generateEventID(),
		Type:      eventType,
		Source:    source,
		Timestamp: time.Now(),
		Data:      data,
		Metadata:  make(map[string]interface{}),
		Priority:  0,
	}
}

// WithTopic sets the event topic
func (e Event) WithTopic(topic string) Event {
	e.Topic = topic
	return e
}

// WithPriority sets the event priority
func (e Event) WithPriority(priority int) Event {
	e.Priority = priority
	return e
}

// WithTTL sets the event TTL
func (e Event) WithTTL(ttl time.Duration) Event {
	e.TTL = ttl
	return e
}

// WithMetadata adds metadata to the event
func (e Event) WithMetadata(key string, value interface{}) Event {
	if e.Metadata == nil {
		e.Metadata = make(map[string]interface{})
	}
	e.Metadata[key] = value
	return e
}

// IsExpired checks if the event has expired based on TTL
func (e Event) IsExpired() bool {
	if e.TTL == 0 {
		return false
	}
	return time.Since(e.Timestamp) > e.TTL
}

// Clone creates a deep copy of the event
func (e Event) Clone() Event {
	clone := e

	// Deep copy metadata
	if e.Metadata != nil {
		clone.Metadata = make(map[string]interface{})
		for k, v := range e.Metadata {
			clone.Metadata[k] = v
		}
	}

	// Note: Data is not deep copied as it can be any type
	// Callers should handle data immutability
	return clone
}

// MarshalJSON customizes JSON marshaling
func (e Event) MarshalJSON() ([]byte, error) {
	type Alias Event
	return json.Marshal(&struct {
		Timestamp string `json:"timestamp"`
		*Alias
	}{
		Timestamp: e.Timestamp.Format(time.RFC3339Nano),
		Alias:     (*Alias)(&e),
	})
}

// UnmarshalJSON customizes JSON unmarshaling
func (e *Event) UnmarshalJSON(data []byte) error {
	type Alias Event
	aux := &struct {
		Timestamp string `json:"timestamp"`
		*Alias
	}{
		Alias: (*Alias)(e),
	}

	if err := json.Unmarshal(data, &aux); err != nil {
		return err
	}

	if aux.Timestamp != "" {
		timestamp, err := time.Parse(time.RFC3339Nano, aux.Timestamp)
		if err != nil {
			return err
		}
		e.Timestamp = timestamp
	}

	return nil
}

// generateEventID generates a unique event ID
func generateEventID() string {
	return fmt.Sprintf("evt_%d_%d", time.Now().UnixNano(), time.Now().Nanosecond())
}
