package core

import (
	"sync"
	"time"

	"github.com/SamyRai/juleson/internal/agent"
)

// Metrics tracks agent performance metrics
type Metrics struct {
	// Execution metrics
	TotalExecutions      int64
	SuccessfulExecutions int64
	FailedExecutions     int64

	// Decision metrics
	DecisionsMade   map[agent.DecisionType]int64
	DecisionLatency map[agent.DecisionType][]time.Duration

	// Tool metrics
	ToolInvocations    map[string]int64
	ToolSuccessRate    map[string]float64
	ToolAverageLatency map[string]time.Duration

	// State metrics
	TimeInState      map[agent.AgentState]time.Duration
	StateTransitions map[string]int64 // "FROM->TO" format

	// Task metrics
	TasksCreated        int64
	TasksCompleted      int64
	TasksFailed         int64
	AverageTaskDuration time.Duration

	// Review metrics
	ReviewsPerformed   int64
	ReviewsApproved    int64
	ReviewsRejected    int64
	AverageReviewScore float64

	// Learning metrics
	LearningsStored    int64
	LearningsApplied   int64
	LearningConfidence []float64

	mu sync.RWMutex
}

// NewMetrics creates a new metrics tracker
func NewMetrics() *Metrics {
	return &Metrics{
		DecisionsMade:      make(map[agent.DecisionType]int64),
		DecisionLatency:    make(map[agent.DecisionType][]time.Duration),
		ToolInvocations:    make(map[string]int64),
		ToolSuccessRate:    make(map[string]float64),
		ToolAverageLatency: make(map[string]time.Duration),
		TimeInState:        make(map[agent.AgentState]time.Duration),
		StateTransitions:   make(map[string]int64),
		LearningConfidence: make([]float64, 0),
	}
}

// RecordExecution records an execution attempt
func (m *Metrics) RecordExecution(success bool, duration time.Duration) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.TotalExecutions++
	if success {
		m.SuccessfulExecutions++
	} else {
		m.FailedExecutions++
	}
}

// RecordDecision records a decision made by the agent
func (m *Metrics) RecordDecision(decisionType agent.DecisionType, latency time.Duration) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.DecisionsMade[decisionType]++
	m.DecisionLatency[decisionType] = append(m.DecisionLatency[decisionType], latency)
}

// RecordToolInvocation records a tool being used
func (m *Metrics) RecordToolInvocation(toolName string, success bool, latency time.Duration) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.ToolInvocations[toolName]++

	// Update success rate
	currentCount := m.ToolInvocations[toolName]
	currentRate := m.ToolSuccessRate[toolName]
	if success {
		m.ToolSuccessRate[toolName] = (currentRate*float64(currentCount-1) + 1.0) / float64(currentCount)
	} else {
		m.ToolSuccessRate[toolName] = (currentRate * float64(currentCount-1)) / float64(currentCount)
	}

	// Update average latency
	currentLatency := m.ToolAverageLatency[toolName]
	m.ToolAverageLatency[toolName] = (currentLatency*time.Duration(currentCount-1) + latency) / time.Duration(currentCount)
}

// RecordStateTransition records a state change
func (m *Metrics) RecordStateTransition(from, to agent.AgentState, duration time.Duration) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.TimeInState[from] += duration

	transitionKey := string(from) + "->" + string(to)
	m.StateTransitions[transitionKey]++
}

// RecordTask records task metrics
func (m *Metrics) RecordTask(created bool, completed bool, failed bool, duration time.Duration) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if created {
		m.TasksCreated++
	}
	if completed {
		m.TasksCompleted++
		// Update average duration
		totalCompleted := m.TasksCompleted
		m.AverageTaskDuration = (m.AverageTaskDuration*time.Duration(totalCompleted-1) + duration) / time.Duration(totalCompleted)
	}
	if failed {
		m.TasksFailed++
	}
}

// RecordReview records review metrics
func (m *Metrics) RecordReview(decision agent.ReviewDecision, score float64) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.ReviewsPerformed++

	if decision == agent.ReviewDecisionApprove {
		m.ReviewsApproved++
	} else if decision == agent.ReviewDecisionReject {
		m.ReviewsRejected++
	}

	// Update average score
	m.AverageReviewScore = (m.AverageReviewScore*float64(m.ReviewsPerformed-1) + score) / float64(m.ReviewsPerformed)
}

// RecordLearning records learning metrics
func (m *Metrics) RecordLearning(stored bool, applied bool, confidence float64) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if stored {
		m.LearningsStored++
		m.LearningConfidence = append(m.LearningConfidence, confidence)
	}
	if applied {
		m.LearningsApplied++
	}
}

// GetSuccessRate returns the overall success rate
func (m *Metrics) GetSuccessRate() float64 {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if m.TotalExecutions == 0 {
		return 0.0
	}
	return float64(m.SuccessfulExecutions) / float64(m.TotalExecutions)
}

// GetAverageLearningConfidence returns average confidence of learnings
func (m *Metrics) GetAverageLearningConfidence() float64 {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if len(m.LearningConfidence) == 0 {
		return 0.0
	}

	sum := 0.0
	for _, conf := range m.LearningConfidence {
		sum += conf
	}
	return sum / float64(len(m.LearningConfidence))
}

// Summary returns a summary of all metrics
func (m *Metrics) Summary() map[string]interface{} {
	m.mu.RLock()
	defer m.mu.RUnlock()

	return map[string]interface{}{
		"executions": map[string]interface{}{
			"total":        m.TotalExecutions,
			"successful":   m.SuccessfulExecutions,
			"failed":       m.FailedExecutions,
			"success_rate": m.GetSuccessRate(),
		},
		"decisions": map[string]interface{}{
			"by_type": m.DecisionsMade,
			"total":   sumIntMap(m.DecisionsMade),
		},
		"tools": map[string]interface{}{
			"invocations":  m.ToolInvocations,
			"success_rate": m.ToolSuccessRate,
			"avg_latency":  m.ToolAverageLatency,
		},
		"tasks": map[string]interface{}{
			"created":      m.TasksCreated,
			"completed":    m.TasksCompleted,
			"failed":       m.TasksFailed,
			"avg_duration": m.AverageTaskDuration,
		},
		"reviews": map[string]interface{}{
			"performed": m.ReviewsPerformed,
			"approved":  m.ReviewsApproved,
			"rejected":  m.ReviewsRejected,
			"avg_score": m.AverageReviewScore,
		},
		"learning": map[string]interface{}{
			"stored":         m.LearningsStored,
			"applied":        m.LearningsApplied,
			"avg_confidence": m.GetAverageLearningConfidence(),
		},
	}
}

func sumIntMap(m map[agent.DecisionType]int64) int64 {
	sum := int64(0)
	for _, v := range m {
		sum += v
	}
	return sum
}

// TraceSpan represents a traced operation
type TraceSpan struct {
	Name      string
	StartTime time.Time
	EndTime   time.Time
	Duration  time.Duration
	Metadata  map[string]interface{}
	Error     error
}

// Tracer provides distributed tracing capabilities
type Tracer struct {
	spans []TraceSpan
	mu    sync.RWMutex
}

// NewTracer creates a new tracer
func NewTracer() *Tracer {
	return &Tracer{
		spans: make([]TraceSpan, 0),
	}
}

// StartSpan begins a new trace span
func (t *Tracer) StartSpan(name string) *TraceSpan {
	span := &TraceSpan{
		Name:      name,
		StartTime: time.Now(),
		Metadata:  make(map[string]interface{}),
	}
	return span
}

// EndSpan completes a trace span
func (t *Tracer) EndSpan(span *TraceSpan) {
	span.EndTime = time.Now()
	span.Duration = span.EndTime.Sub(span.StartTime)

	t.mu.Lock()
	defer t.mu.Unlock()
	t.spans = append(t.spans, *span)
}

// GetSpans returns all recorded spans
func (t *Tracer) GetSpans() []TraceSpan {
	t.mu.RLock()
	defer t.mu.RUnlock()

	spans := make([]TraceSpan, len(t.spans))
	copy(spans, t.spans)
	return spans
}
