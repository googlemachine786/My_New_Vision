// Package circuitbreaker implements the circuit breaker pattern for fault tolerance.
package circuitbreaker

import (
	"sync"
	"time"
)

// State represents the state of a circuit breaker
type State int

const (
	StateClosed State = iota // Normal operation
	StateOpen                // Failing, reject requests
	StateHalfOpen            // Testing if service recovered
)

// String returns the state as a string
func (s State) String() string {
	switch s {
	case StateClosed:
		return "closed"
	case StateOpen:
		return "open"
	case StateHalfOpen:
		return "half-open"
	default:
		return "unknown"
	}
}

// CircuitBreaker implements the circuit breaker pattern
type CircuitBreaker struct {
	mu sync.Mutex

	state             State
	consecutiveFailures int
	consecutiveSuccesses int
	lastFailure         time.Time

	// Configuration
	failureThreshold  int
	recoveryTimeout   time.Duration
	halfOpenMaxRequests int

	// Metrics
	totalRequests  int64
	totalFailures  int64
	totalSuccesses int64
	totalRejected  int64
}

// Option configures a CircuitBreaker
type Option interface {
	apply(*options)
}

type options struct {
	failureThreshold    int
	recoveryTimeout     time.Duration
	halfOpenMaxRequests int
}

type optionFunc func(*options)

func (f optionFunc) apply(o *options) { f(o) }

// WithFailureThreshold sets the number of failures before opening circuit
func WithFailureThreshold(n int) Option {
	return optionFunc(func(o *options) {
		o.failureThreshold = n
	})
}

// WithRecoveryTimeout sets the time before attempting recovery
func WithRecoveryTimeout(d time.Duration) Option {
	return optionFunc(func(o *options) {
		o.recoveryTimeout = d
	})
}

// WithHalfOpenMaxRequests sets max requests in half-open state
func WithHalfOpenMaxRequests(n int) Option {
	return optionFunc(func(o *options) {
		o.halfOpenMaxRequests = n
	})
}

// New creates a new circuit breaker with default settings
func New(opts ...Option) *CircuitBreaker {
	options := options{
		failureThreshold:    5,
		recoveryTimeout:     60 * time.Second,
		halfOpenMaxRequests: 3,
	}

	for _, opt := range opts {
		opt.apply(&options)
	}

	return &CircuitBreaker{
		state:               StateClosed,
		failureThreshold:    options.failureThreshold,
		recoveryTimeout:     options.recoveryTimeout,
		halfOpenMaxRequests: options.halfOpenMaxRequests,
	}
}

// AllowRequest checks if a request should be allowed through
func (cb *CircuitBreaker) AllowRequest() bool {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	cb.totalRequests++

	switch cb.state {
	case StateClosed:
		return true

	case StateOpen:
		if time.Since(cb.lastFailure) > cb.recoveryTimeout {
			cb.state = StateHalfOpen
			cb.consecutiveSuccesses = 0
			return true
		}
		cb.totalRejected++
		return false

	case StateHalfOpen:
		if cb.consecutiveSuccesses < cb.halfOpenMaxRequests {
			return true
		}
		cb.totalRejected++
		return false

	default:
		return false
	}
}

// RecordSuccess records a successful request
func (cb *CircuitBreaker) RecordSuccess() {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	cb.totalSuccesses++

	switch cb.state {
	case StateHalfOpen:
		cb.consecutiveSuccesses++
		if cb.consecutiveSuccesses >= cb.halfOpenMaxRequests {
			cb.state = StateClosed
			cb.consecutiveFailures = 0
		}
	case StateClosed:
		// Reset consecutive failures on success
		cb.consecutiveFailures = 0
	}
}

// RecordFailure records a failed request
func (cb *CircuitBreaker) RecordFailure() {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	cb.totalFailures++
	cb.lastFailure = time.Now()

	switch cb.state {
	case StateClosed:
		cb.consecutiveFailures++
		if cb.consecutiveFailures >= cb.failureThreshold {
			cb.state = StateOpen
		}
	case StateHalfOpen:
		// Any failure in half-open sends back to open
		cb.state = StateOpen
	}
}

// State returns the current state (pure getter, no side effects)
func (cb *CircuitBreaker) State() State {
	cb.mu.Lock()
	defer cb.mu.Unlock()
	return cb.state
}

// Stats returns circuit breaker statistics (pure read, no side effects).
func (cb *CircuitBreaker) Stats() map[string]interface{} {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	return map[string]interface{}{
		"state":                cb.state.String(),
		"consecutive_failures": cb.consecutiveFailures,
		"total_requests":       cb.totalRequests,
		"total_failures":       cb.totalFailures,
		"total_successes":      cb.totalSuccesses,
		"total_rejected":       cb.totalRejected,
		"failure_threshold":    cb.failureThreshold,
		"recovery_timeout":     cb.recoveryTimeout.String(),
	}
}
