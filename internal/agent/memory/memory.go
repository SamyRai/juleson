package memory

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/SamyRai/juleson/internal/agent"
)

// Memory provides learning and recall capabilities for the agent
type Memory interface {
	// Store saves a learning
	Store(ctx context.Context, learning agent.Learning) error

	// Recall retrieves learnings matching a pattern or context
	Recall(ctx context.Context, pattern string) ([]agent.Learning, error)

	// RecordDecision records a decision and its outcome
	RecordDecision(ctx context.Context, decision agent.Decision) error

	// GetDecisionHistory retrieves decision history
	GetDecisionHistory(ctx context.Context, limit int) ([]agent.Decision, error)

	// UpdateLearningConfidence adjusts confidence based on application success
	UpdateLearningConfidence(ctx context.Context, learningID string, successful bool) error
}

// inMemoryStore implements episodic memory with in-memory storage
// For production, this should be backed by a persistent store like SQLite
type inMemoryStore struct {
	learnings map[string]agent.Learning
	decisions []agent.Decision
	mu        sync.RWMutex
}

// NewMemory creates a new memory system
func NewMemory() Memory {
	return &inMemoryStore{
		learnings: make(map[string]agent.Learning),
		decisions: make([]agent.Decision, 0),
	}
}

// Store saves a learning
func (m *inMemoryStore) Store(ctx context.Context, learning agent.Learning) error {
	if learning.Pattern == "" && learning.Lesson == "" {
		return fmt.Errorf("learning must have either a pattern or lesson")
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	if learning.ID == "" {
		learning.ID = generateID()
	}

	if learning.Timestamp.IsZero() {
		learning.Timestamp = time.Now()
	}

	if learning.Confidence == 0 {
		learning.Confidence = 0.5 // Default moderate confidence
	} else if learning.Confidence < 0 || learning.Confidence > 1 {
		return fmt.Errorf("confidence must be between 0 and 1, got: %f", learning.Confidence)
	}

	m.learnings[learning.ID] = learning
	return nil
}

// Recall retrieves learnings matching a pattern or context
func (m *inMemoryStore) Recall(ctx context.Context, pattern string) ([]agent.Learning, error) {
	if pattern == "" {
		return nil, fmt.Errorf("pattern cannot be empty")
	}

	m.mu.RLock()
	defer m.mu.RUnlock()

	var results []agent.Learning

	for _, learning := range m.learnings {
		// Simple pattern matching - check if pattern appears in context, pattern, or lesson
		if contains(learning.Context, pattern) ||
			contains(learning.Pattern, pattern) ||
			contains(learning.Lesson, pattern) {
			results = append(results, learning)
		}
	}

	// Sort by confidence (highest first) and recency
	sortLearnings(results)

	return results, nil
}

// RecordDecision records a decision and its outcome
func (m *inMemoryStore) RecordDecision(ctx context.Context, decision agent.Decision) error {
	if decision.Reasoning == "" {
		return fmt.Errorf("decision must have reasoning")
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	if decision.ID == "" {
		decision.ID = generateID()
	}

	if decision.Timestamp.IsZero() {
		decision.Timestamp = time.Now()
	}

	if decision.Confidence < 0 || decision.Confidence > 1 {
		return fmt.Errorf("confidence must be between 0 and 1, got: %f", decision.Confidence)
	}

	m.decisions = append(m.decisions, decision)
	return nil
}

// GetDecisionHistory retrieves recent decisions
func (m *inMemoryStore) GetDecisionHistory(ctx context.Context, limit int) ([]agent.Decision, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if limit <= 0 || limit > len(m.decisions) {
		limit = len(m.decisions)
	}

	// Return most recent decisions
	start := len(m.decisions) - limit
	if start < 0 {
		start = 0
	}

	history := make([]agent.Decision, limit)
	copy(history, m.decisions[start:])

	// Reverse to get newest first
	for i := 0; i < len(history)/2; i++ {
		j := len(history) - 1 - i
		history[i], history[j] = history[j], history[i]
	}

	return history, nil
}

// UpdateLearningConfidence adjusts confidence based on whether application was successful
func (m *inMemoryStore) UpdateLearningConfidence(ctx context.Context, learningID string, successful bool) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	learning, exists := m.learnings[learningID]
	if !exists {
		return nil // Silently ignore missing learning
	}

	// Update confidence and application count
	learning.Applications++

	if successful {
		// Increase confidence, but cap at 1.0
		learning.Confidence += 0.05
		if learning.Confidence > 1.0 {
			learning.Confidence = 1.0
		}
	} else {
		// Decrease confidence, but don't go below 0.1
		learning.Confidence -= 0.1
		if learning.Confidence < 0.1 {
			learning.Confidence = 0.1
		}
	}

	m.learnings[learningID] = learning
	return nil
}

// Helper functions

func generateID() string {
	return time.Now().Format("20060102150405") + "-" + randString(8)
}

func randString(n int) string {
	const letters = "abcdefghijklmnopqrstuvwxyz0123456789"
	b := make([]byte, n)
	for i := range b {
		b[i] = letters[time.Now().UnixNano()%int64(len(letters))]
	}
	return string(b)
}

func contains(s, substr string) bool {
	return len(s) > 0 && len(substr) > 0 &&
		(s == substr || len(s) >= len(substr) && findInString(s, substr))
}

func findInString(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

func sortLearnings(learnings []agent.Learning) {
	// Simple bubble sort by confidence (descending)
	n := len(learnings)
	for i := 0; i < n-1; i++ {
		for j := 0; j < n-i-1; j++ {
			if learnings[j].Confidence < learnings[j+1].Confidence {
				learnings[j], learnings[j+1] = learnings[j+1], learnings[j]
			}
		}
	}
}
