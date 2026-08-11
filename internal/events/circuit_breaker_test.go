package events

import (
	"context"
	"errors"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestCircuitBreaker_BasicTransitions(t *testing.T) {
	cb := NewCircuitBreaker(&CircuitBreakerConfig{
		Name:         "test-cb",
		MaxFailures:  3,
		Timeout:      1 * time.Second,
		ResetTimeout: 50 * time.Millisecond,
	}, nil)

	ctx := context.Background()
	failErr := errors.New("simulated failure")

	// Helper to execute failing function
	failFunc := func(ctx context.Context) error { return failErr }
	// Helper to execute successful function
	successFunc := func(ctx context.Context) error { return nil }

	// 1. Initial State should be Closed
	assert.Equal(t, StateClosed, cb.GetState())

	// 2. Fail 2 times - should still be closed
	_ = cb.Execute(ctx, failFunc)
	_ = cb.Execute(ctx, failFunc)
	assert.Equal(t, StateClosed, cb.GetState())

	// 3. Fail 3rd time - should transition to Open
	_ = cb.Execute(ctx, failFunc)
	assert.Equal(t, StateOpen, cb.GetState())

	// 4. Try to execute while Open - should return circuit breaker error
	err := cb.Execute(ctx, successFunc)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "circuit breaker test-cb is open")

	// 5. Wait for ResetTimeout
	time.Sleep(100 * time.Millisecond)

	// 6. Execute successful function - should transition HalfOpen -> Closed
	err = cb.Execute(ctx, successFunc)
	assert.NoError(t, err)
	assert.Equal(t, StateClosed, cb.GetState())
}

func TestCircuitBreaker_Concurrency(t *testing.T) {
	cb := NewCircuitBreaker(&CircuitBreakerConfig{
		Name:         "concurrent-cb",
		MaxFailures:  50,
		Timeout:      1 * time.Second,
		ResetTimeout: 1 * time.Second,
	}, nil)

	ctx := context.Background()
	var wg sync.WaitGroup
	workers := 100
	iterations := 10

	var successCount int32
	var failCount int32

	// Launch multiple workers hitting the circuit breaker concurrently
	for i := 0; i < workers; i++ {
		wg.Add(1)
		go func(workerID int) {
			defer wg.Done()
			for j := 0; j < iterations; j++ {
				err := cb.Execute(ctx, func(ctx context.Context) error {
					// Make all workers fail so we reliably trip the breaker
					return errors.New("concurrent worker failure")
				})
				if err != nil {
					atomic.AddInt32(&failCount, 1)
				} else {
					atomic.AddInt32(&successCount, 1)
				}
			}
		}(i)
	}

	wg.Wait()

	// 50 workers will succeed 10 times = 500 successes
	// 50 workers will fail 10 times = 500 failures
	// Since MaxFailures is 50, the circuit will trip open early.
	// So we expect total successes to be somewhat low and circuit to be OPEN.
	assert.Equal(t, StateOpen, cb.GetState())
	metrics := cb.GetMetrics()
	assert.GreaterOrEqual(t, metrics["failures"].(int), 50)
}

func TestCircuitBreakerPool_GetOrCreate(t *testing.T) {
	pool := NewCircuitBreakerPool(nil)

	// Create new
	cb1 := pool.GetOrCreate("service-a", nil)
	assert.NotNil(t, cb1)
	assert.Equal(t, "service-a", cb1.name)

	// Get existing
	cb2 := pool.GetOrCreate("service-a", nil)
	assert.Same(t, cb1, cb2)

	// GetAll
	pool.GetOrCreate("service-b", nil)
	all := pool.GetAll()
	assert.Len(t, all, 2)

	// ResetAll
	cb1.setState(StateOpen)
	assert.Equal(t, StateOpen, cb1.GetState())
	pool.ResetAll()
	assert.Equal(t, StateClosed, cb1.GetState())
}
