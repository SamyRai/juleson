package memory

import (
	"context"
	"testing"
	"time"

	"github.com/SamyRai/juleson/internal/agent"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewMemory(t *testing.T) {
	mem := NewMemory()
	require.NotNil(t, mem)
}

func TestStoreAndRecall(t *testing.T) {
	mem := NewMemory()
	ctx := context.Background()

	learning := agent.Learning{
		Context:    "testing code",
		Pattern:    "if err != nil",
		Lesson:     "always check errors",
		Confidence: 0.8,
	}

	// Test Store
	err := mem.Store(ctx, learning)
	require.NoError(t, err)

	// Test Store invalid learning
	err = mem.Store(ctx, agent.Learning{})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "learning must have either a pattern or lesson")

	// Test Store invalid confidence
	err = mem.Store(ctx, agent.Learning{Pattern: "test", Confidence: 1.5})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "confidence must be between 0 and 1")

	// Test Recall
	results, err := mem.Recall(ctx, "err")
	require.NoError(t, err)
	require.Len(t, results, 1)
	assert.Equal(t, "always check errors", results[0].Lesson)

	// Test Recall empty pattern
	_, err = mem.Recall(ctx, "")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "pattern cannot be empty")
}

func TestRecallSorting(t *testing.T) {
	mem := NewMemory()
	ctx := context.Background()

	_ = mem.Store(ctx, agent.Learning{Pattern: "match", Confidence: 0.3})
	_ = mem.Store(ctx, agent.Learning{Pattern: "match", Confidence: 0.9})
	_ = mem.Store(ctx, agent.Learning{Pattern: "match", Confidence: 0.5})

	results, err := mem.Recall(ctx, "match")
	require.NoError(t, err)
	require.Len(t, results, 3)

	// Should be sorted by confidence descending
	assert.Equal(t, 0.9, results[0].Confidence)
	assert.Equal(t, 0.5, results[1].Confidence)
	assert.Equal(t, 0.3, results[2].Confidence)
}

func TestRecordDecisionAndHistory(t *testing.T) {
	mem := NewMemory()
	ctx := context.Background()

	// Invalid decision (no reasoning)
	err := mem.RecordDecision(ctx, agent.Decision{})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "decision must have reasoning")

	// Store valid decisions
	for i := 1; i <= 5; i++ {
		err := mem.RecordDecision(ctx, agent.Decision{
			Reasoning:  "reasoning",
			Action:     "action",
			Confidence: 0.5,
		})
		require.NoError(t, err)
		time.Sleep(1 * time.Millisecond) // Ensure different timestamps
	}

	// Test history
	history, err := mem.GetDecisionHistory(ctx, 3)
	require.NoError(t, err)
	require.Len(t, history, 3)

	// Should be newest first, but all our reasoning strings are the same.
	// We'll test limit bounds.
	history, err = mem.GetDecisionHistory(ctx, 10)
	require.NoError(t, err)
	require.Len(t, history, 5) // Should cap at total available
}

func TestUpdateLearningConfidence(t *testing.T) {
	mem := NewMemory()
	ctx := context.Background()

	err := mem.Store(ctx, agent.Learning{
		ID:         "learning-1",
		Pattern:    "test pattern",
		Confidence: 0.5,
	})
	require.NoError(t, err)

	// Test successful application
	err = mem.UpdateLearningConfidence(ctx, "learning-1", true)
	require.NoError(t, err)

	results, err := mem.Recall(ctx, "test")
	require.NoError(t, err)
	require.Len(t, results, 1)
	assert.InDelta(t, 0.55, results[0].Confidence, 0.001)
	assert.Equal(t, 1, results[0].Applications)

	// Test unsuccessful application
	err = mem.UpdateLearningConfidence(ctx, "learning-1", false)
	require.NoError(t, err)

	results, _ = mem.Recall(ctx, "test")
	assert.InDelta(t, 0.45, results[0].Confidence, 0.001)

	// Test missing learning
	err = mem.UpdateLearningConfidence(ctx, "non-existent", true)
	require.NoError(t, err) // Should silently ignore
}
