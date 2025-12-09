package core

import (
	"context"
	"fmt"
	"log/slog"
	"math"
	"strings"
	"time"
)

// Default retry configuration constants
const (
	// DefaultMaxRetries is the default maximum number of retry attempts
	DefaultMaxRetries = 3
	// DefaultInitialDelay is the default initial delay before first retry
	DefaultInitialDelay = 1 * time.Second
	// DefaultMaxDelay is the default maximum delay between retries
	DefaultMaxDelay = 30 * time.Second
	// DefaultBackoffFactor is the default exponential backoff multiplier
	DefaultBackoffFactor = 2.0
	// DefaultJitterFactor is the default jitter factor for randomizing delays
	DefaultJitterFactor = 0.25
	// DefaultTokenCheckInterval is the default interval for checking rate limit tokens
	DefaultTokenCheckInterval = 100 * time.Millisecond
	// SecondsPerMinute is used for rate limit calculations
	SecondsPerMinute = 60.0
)

// RetryStrategy defines how retries should be performed
type RetryStrategy struct {
	MaxRetries      int
	InitialDelay    time.Duration
	MaxDelay        time.Duration
	BackoffFactor   float64
	RetryableErrors []string
	Logger          *slog.Logger
}

// DefaultRetryStrategy returns a sensible retry strategy
func DefaultRetryStrategy() *RetryStrategy {
	return &RetryStrategy{
		MaxRetries:    DefaultMaxRetries,
		InitialDelay:  DefaultInitialDelay,
		MaxDelay:      DefaultMaxDelay,
		BackoffFactor: DefaultBackoffFactor,
		RetryableErrors: []string{
			"timeout",
			"connection refused",
			"temporary failure",
			"rate limit",
			"503",
			"502",
			"504",
		},
		Logger: slog.Default(),
	}
}

// RetryableOperation represents an operation that can be retried
type RetryableOperation func(ctx context.Context, attempt int) error

// Execute runs an operation with retry logic and exponential backoff
func (rs *RetryStrategy) Execute(ctx context.Context, operation RetryableOperation, operationName string) error {
	if operation == nil {
		return fmt.Errorf("cannot execute retry: operation is nil")
	}
	if operationName == "" {
		return fmt.Errorf("cannot execute retry: operationName is empty")
	}

	var lastError error

	for attempt := 0; attempt <= rs.MaxRetries; attempt++ {
		if attempt > 0 {
			delay := rs.calculateBackoff(attempt)
			rs.Logger.Info("retry.waiting",
				"operation", operationName,
				"attempt", attempt,
				"delay", delay)

			select {
			case <-time.After(delay):
				// Continue with retry
			case <-ctx.Done():
				return fmt.Errorf("operation cancelled during retry backoff: %w", ctx.Err())
			}
		}

		rs.Logger.Info("retry.attempt",
			"operation", operationName,
			"attempt", attempt,
			"max_retries", rs.MaxRetries)

		err := operation(ctx, attempt)
		if err == nil {
			if attempt > 0 {
				rs.Logger.Info("retry.success",
					"operation", operationName,
					"attempt", attempt)
			}
			return nil
		}

		lastError = err

		// Check if error is retryable
		if !rs.isRetryable(err) {
			rs.Logger.Warn("retry.non_retryable_error",
				"operation", operationName,
				"error", err,
				"attempt", attempt)
			return fmt.Errorf("non-retryable error: %w", err)
		}

		rs.Logger.Warn("retry.failed_attempt",
			"operation", operationName,
			"error", err,
			"attempt", attempt,
			"will_retry", attempt < rs.MaxRetries)
	}

	return fmt.Errorf("operation failed after %d attempts: %w", rs.MaxRetries+1, lastError)
}

// ExecuteWithResult runs an operation with retry logic and returns a result
func (rs *RetryStrategy) ExecuteWithResult(
	ctx context.Context,
	operation func(ctx context.Context, attempt int) (interface{}, error),
	operationName string,
) (interface{}, error) {
	var lastError error

	for attempt := 0; attempt <= rs.MaxRetries; attempt++ {
		if attempt > 0 {
			delay := rs.calculateBackoff(attempt)

			select {
			case <-time.After(delay):
			case <-ctx.Done():
				return nil, fmt.Errorf("operation cancelled during retry backoff: %w", ctx.Err())
			}
		}

		result, err := operation(ctx, attempt)
		if err == nil {
			return result, nil
		}

		lastError = err

		if !rs.isRetryable(err) {
			return nil, fmt.Errorf("non-retryable error: %w", err)
		}
	}

	return nil, fmt.Errorf("operation failed after %d attempts: %w", rs.MaxRetries+1, lastError)
}

// calculateBackoff calculates the delay for the current attempt using exponential backoff
func (rs *RetryStrategy) calculateBackoff(attempt int) time.Duration {
	// Exponential backoff: delay = initialDelay * (backoffFactor ^ attempt)
	delay := float64(rs.InitialDelay) * math.Pow(rs.BackoffFactor, float64(attempt-1))

	// Add jitter (±25%) to prevent thundering herd
	jitter := delay * DefaultJitterFactor * (2.0*float64(time.Now().UnixNano()%100)/100.0 - 1.0)
	delay += jitter

	// Cap at max delay
	if delay > float64(rs.MaxDelay) {
		delay = float64(rs.MaxDelay)
	}

	return time.Duration(delay)
}

// isRetryable checks if an error should trigger a retry
func (rs *RetryStrategy) isRetryable(err error) bool {
	if err == nil {
		return false
	}

	errMsg := err.Error()
	for _, retryableErr := range rs.RetryableErrors {
		if strings.Contains(errMsg, retryableErr) {
			return true
		}
	}

	return false
}

// CircuitBreaker prevents cascading failures
type CircuitBreaker struct {
	MaxFailures     int
	ResetTimeout    time.Duration
	HalfOpenTimeout time.Duration

	state         CircuitState
	failures      int
	lastFailTime  time.Time
	lastStateTime time.Time
	logger        *slog.Logger
}

// CircuitState represents the state of a circuit breaker
type CircuitState string

const (
	CircuitStateClosed   CircuitState = "CLOSED"    // Normal operation
	CircuitStateOpen     CircuitState = "OPEN"      // Failing, reject requests
	CircuitStateHalfOpen CircuitState = "HALF_OPEN" // Testing if recovered
)

// NewCircuitBreaker creates a new circuit breaker
func NewCircuitBreaker(maxFailures int, resetTimeout time.Duration, logger *slog.Logger) *CircuitBreaker {
	if maxFailures <= 0 {
		maxFailures = 3
	}
	if resetTimeout <= 0 {
		resetTimeout = 30 * time.Second
	}
	if logger == nil {
		logger = slog.Default()
	}
	return &CircuitBreaker{
		MaxFailures:     maxFailures,
		ResetTimeout:    resetTimeout,
		HalfOpenTimeout: resetTimeout / 2,
		state:           CircuitStateClosed,
		logger:          logger,
	}
}

// Execute runs an operation through the circuit breaker
func (cb *CircuitBreaker) Execute(ctx context.Context, operation RetryableOperation, name string) error {
	if operation == nil {
		return fmt.Errorf("cannot execute circuit breaker: operation is nil")
	}
	if name == "" {
		return fmt.Errorf("cannot execute circuit breaker: name is empty")
	}

	if !cb.AllowRequest() {
		return fmt.Errorf("circuit breaker is OPEN for %s", name)
	}

	err := operation(ctx, 0)

	if err != nil {
		cb.RecordFailure()
		return err
	}

	cb.RecordSuccess()
	return nil
}

// AllowRequest checks if the circuit breaker allows a request
func (cb *CircuitBreaker) AllowRequest() bool {
	now := time.Now()

	switch cb.state {
	case CircuitStateClosed:
		return true

	case CircuitStateOpen:
		// Check if enough time has passed to try again
		if now.Sub(cb.lastFailTime) > cb.ResetTimeout {
			cb.setState(CircuitStateHalfOpen)
			return true
		}
		return false

	case CircuitStateHalfOpen:
		// Allow one request to test
		return true

	default:
		return false
	}
}

// RecordSuccess records a successful operation
func (cb *CircuitBreaker) RecordSuccess() {
	if cb.state == CircuitStateHalfOpen {
		cb.logger.Info("circuit_breaker.recovered")
		cb.setState(CircuitStateClosed)
		cb.failures = 0
	}
}

// RecordFailure records a failed operation
func (cb *CircuitBreaker) RecordFailure() {
	cb.failures++
	cb.lastFailTime = time.Now()

	if cb.state == CircuitStateHalfOpen {
		cb.logger.Warn("circuit_breaker.half_open_failed")
		cb.setState(CircuitStateOpen)
		return
	}

	if cb.failures >= cb.MaxFailures {
		cb.logger.Warn("circuit_breaker.opened",
			"failures", cb.failures,
			"max_failures", cb.MaxFailures)
		cb.setState(CircuitStateOpen)
	}
}

// GetState returns the current circuit state
func (cb *CircuitBreaker) GetState() CircuitState {
	return cb.state
}

// setState changes the circuit breaker state
func (cb *CircuitBreaker) setState(newState CircuitState) {
	oldState := cb.state
	cb.state = newState
	cb.lastStateTime = time.Now()

	if oldState != newState {
		cb.logger.Info("circuit_breaker.state_change",
			"from", oldState,
			"to", newState)
	}
}

// RateLimiter prevents overwhelming external services
type RateLimiter struct {
	RequestsPerMinute int
	BurstSize         int

	tokens    float64
	lastCheck time.Time
	logger    *slog.Logger
}

// NewRateLimiter creates a new rate limiter
func NewRateLimiter(requestsPerMinute int, burstSize int, logger *slog.Logger) *RateLimiter {
	if requestsPerMinute <= 0 {
		requestsPerMinute = 60
	}
	if burstSize <= 0 {
		burstSize = 10
	}
	if logger == nil {
		logger = slog.Default()
	}
	return &RateLimiter{
		RequestsPerMinute: requestsPerMinute,
		BurstSize:         burstSize,
		tokens:            float64(burstSize),
		lastCheck:         time.Now(),
		logger:            logger,
	}
}

// Allow checks if a request is allowed under rate limits
func (rl *RateLimiter) Allow() bool {
	now := time.Now()
	elapsed := now.Sub(rl.lastCheck)
	rl.lastCheck = now

	// Add tokens based on elapsed time
	tokensToAdd := elapsed.Seconds() * float64(rl.RequestsPerMinute) / SecondsPerMinute
	rl.tokens = math.Min(rl.tokens+tokensToAdd, float64(rl.BurstSize))

	if rl.tokens >= 1.0 {
		rl.tokens -= 1.0
		return true
	}

	rl.logger.Warn("rate_limiter.throttled",
		"tokens", rl.tokens,
		"rpm", rl.RequestsPerMinute)
	return false
}

// Wait waits until a request is allowed
func (rl *RateLimiter) Wait(ctx context.Context) error {
	for {
		if rl.Allow() {
			return nil
		}

		// Wait a bit before checking again
		select {
		case <-time.After(DefaultTokenCheckInterval):
			continue
		case <-ctx.Done():
			return ctx.Err()
		}
	}
}
