package events

import (
	"context"
	"fmt"
	"log/slog"
	"sync"
	"time"
)

// CircuitBreaker implements the circuit breaker pattern for event processing.
// It prevents cascading failures by temporarily stopping event processing
// when a high failure rate is detected.
type CircuitBreaker struct {
	name           string
	maxFailures    int
	timeout        time.Duration
	resetTimeout   time.Duration
	state          CircuitState
	failures       int
	lastFailTime   time.Time
	stateChangedAt time.Time
	mu             sync.RWMutex
	logger         *slog.Logger
	onStateChange  func(old, new CircuitState)
}

// CircuitState represents the state of a circuit breaker
type CircuitState string

const (
	StateClosed   CircuitState = "CLOSED"    // Normal operation
	StateOpen     CircuitState = "OPEN"      // Blocking requests
	StateHalfOpen CircuitState = "HALF_OPEN" // Testing if service recovered
)

// CircuitBreakerConfig configures a circuit breaker
type CircuitBreakerConfig struct {
	Name          string
	MaxFailures   int
	Timeout       time.Duration
	ResetTimeout  time.Duration
	OnStateChange func(old, new CircuitState)
}

// DefaultCircuitBreakerConfig returns default configuration
func DefaultCircuitBreakerConfig(name string) *CircuitBreakerConfig {
	return &CircuitBreakerConfig{
		Name:         name,
		MaxFailures:  5,
		Timeout:      30 * time.Second,
		ResetTimeout: 60 * time.Second,
	}
}

// NewCircuitBreaker creates a new circuit breaker
func NewCircuitBreaker(config *CircuitBreakerConfig, logger *slog.Logger) *CircuitBreaker {
	if config == nil {
		config = DefaultCircuitBreakerConfig("default")
	}
	if logger == nil {
		logger = slog.Default()
	}

	return &CircuitBreaker{
		name:           config.Name,
		maxFailures:    config.MaxFailures,
		timeout:        config.Timeout,
		resetTimeout:   config.ResetTimeout,
		state:          StateClosed,
		stateChangedAt: time.Now(),
		logger:         logger,
		onStateChange:  config.OnStateChange,
	}
}

// Execute executes a function with circuit breaker protection
func (cb *CircuitBreaker) Execute(ctx context.Context, fn func(ctx context.Context) error) error {
	cb.mu.RLock()
	state := cb.state
	cb.mu.RUnlock()

	switch state {
	case StateOpen:
		// Check if we should transition to half-open
		cb.mu.Lock()
		if time.Since(cb.stateChangedAt) > cb.resetTimeout {
			cb.setState(StateHalfOpen)
			cb.mu.Unlock()
		} else {
			cb.mu.Unlock()
			return fmt.Errorf("circuit breaker %s is open", cb.name)
		}

	case StateHalfOpen:
		// Allow one request through to test if service recovered
	}

	// Execute function
	err := fn(ctx)

	if err != nil {
		cb.recordFailure()
		return err
	}

	cb.recordSuccess()
	return nil
}

// recordFailure records a failed execution
func (cb *CircuitBreaker) recordFailure() {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	cb.failures++
	cb.lastFailTime = time.Now()

	cb.logger.Warn("circuit breaker recorded failure",
		"name", cb.name,
		"failures", cb.failures,
		"max_failures", cb.maxFailures,
		"state", cb.state)

	if cb.state == StateHalfOpen {
		// Failed during half-open, go back to open
		cb.setState(StateOpen)
	} else if cb.failures >= cb.maxFailures {
		// Too many failures, open the circuit
		cb.setState(StateOpen)
	}
}

// recordSuccess records a successful execution
func (cb *CircuitBreaker) recordSuccess() {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	cb.logger.Debug("circuit breaker recorded success",
		"name", cb.name,
		"state", cb.state)

	if cb.state == StateHalfOpen {
		// Success during half-open, close the circuit
		cb.setState(StateClosed)
		cb.failures = 0
	} else if cb.state == StateClosed {
		// Gradually decrease failures on success
		if cb.failures > 0 {
			cb.failures--
		}
	}
}

// setState changes the circuit breaker state
func (cb *CircuitBreaker) setState(newState CircuitState) {
	oldState := cb.state
	cb.state = newState
	cb.stateChangedAt = time.Now()

	cb.logger.Info("circuit breaker state changed",
		"name", cb.name,
		"old_state", oldState,
		"new_state", newState)

	if cb.onStateChange != nil {
		cb.onStateChange(oldState, newState)
	}
}

// GetState returns the current state
func (cb *CircuitBreaker) GetState() CircuitState {
	cb.mu.RLock()
	defer cb.mu.RUnlock()
	return cb.state
}

// GetMetrics returns circuit breaker metrics
func (cb *CircuitBreaker) GetMetrics() map[string]interface{} {
	cb.mu.RLock()
	defer cb.mu.RUnlock()

	return map[string]interface{}{
		"name":             cb.name,
		"state":            cb.state,
		"failures":         cb.failures,
		"last_fail_time":   cb.lastFailTime,
		"state_changed_at": cb.stateChangedAt,
	}
}

// Reset manually resets the circuit breaker
func (cb *CircuitBreaker) Reset() {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	cb.setState(StateClosed)
	cb.failures = 0

	cb.logger.Info("circuit breaker manually reset", "name", cb.name)
}

// CircuitBreakerMiddleware creates middleware from a circuit breaker
func CircuitBreakerMiddleware(cb *CircuitBreaker) Middleware {
	return func(next EventHandler) EventHandler {
		return func(ctx context.Context, event Event) error {
			return cb.Execute(ctx, func(ctx context.Context) error {
				return next(ctx, event)
			})
		}
	}
}

// CircuitBreakerPool manages multiple circuit breakers
type CircuitBreakerPool struct {
	breakers map[string]*CircuitBreaker
	mu       sync.RWMutex
	logger   *slog.Logger
}

// NewCircuitBreakerPool creates a new circuit breaker pool
func NewCircuitBreakerPool(logger *slog.Logger) *CircuitBreakerPool {
	if logger == nil {
		logger = slog.Default()
	}

	return &CircuitBreakerPool{
		breakers: make(map[string]*CircuitBreaker),
		logger:   logger,
	}
}

// GetOrCreate gets or creates a circuit breaker
func (cbp *CircuitBreakerPool) GetOrCreate(name string, config *CircuitBreakerConfig) *CircuitBreaker {
	cbp.mu.RLock()
	cb, exists := cbp.breakers[name]
	cbp.mu.RUnlock()

	if exists {
		return cb
	}

	cbp.mu.Lock()
	defer cbp.mu.Unlock()

	// Double-check after acquiring write lock
	cb, exists = cbp.breakers[name]
	if exists {
		return cb
	}

	if config == nil {
		config = DefaultCircuitBreakerConfig(name)
	}

	cb = NewCircuitBreaker(config, cbp.logger)
	cbp.breakers[name] = cb

	return cb
}

// Get retrieves a circuit breaker by name
func (cbp *CircuitBreakerPool) Get(name string) (*CircuitBreaker, bool) {
	cbp.mu.RLock()
	defer cbp.mu.RUnlock()

	cb, exists := cbp.breakers[name]
	return cb, exists
}

// GetAll returns all circuit breakers
func (cbp *CircuitBreakerPool) GetAll() map[string]*CircuitBreaker {
	cbp.mu.RLock()
	defer cbp.mu.RUnlock()

	result := make(map[string]*CircuitBreaker)
	for name, cb := range cbp.breakers {
		result[name] = cb
	}

	return result
}

// ResetAll resets all circuit breakers
func (cbp *CircuitBreakerPool) ResetAll() {
	cbp.mu.RLock()
	breakers := make([]*CircuitBreaker, 0, len(cbp.breakers))
	for _, cb := range cbp.breakers {
		breakers = append(breakers, cb)
	}
	cbp.mu.RUnlock()

	for _, cb := range breakers {
		cb.Reset()
	}

	cbp.logger.Info("all circuit breakers reset")
}
