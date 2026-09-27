package circuitbreaker

import (
	"testing"
	"time"
)

func TestNew(t *testing.T) {
	t.Run("default values", func(t *testing.T) {
		cb := New()

		if cb.State() != StateClosed {
			t.Errorf("New() state = %v, want StateClosed", cb.State())
		}
		if cb.failureThreshold != 5 {
			t.Errorf("New() failureThreshold = %d, want 5", cb.failureThreshold)
		}
		if cb.recoveryTimeout != 60*time.Second {
			t.Errorf("New() recoveryTimeout = %v, want 60s", cb.recoveryTimeout)
		}
		if cb.halfOpenMaxRequests != 3 {
			t.Errorf("New() halfOpenMaxRequests = %d, want 3", cb.halfOpenMaxRequests)
		}
	})

	t.Run("with options", func(t *testing.T) {
		cb := New(
			WithFailureThreshold(10),
			WithRecoveryTimeout(30*time.Second),
			WithHalfOpenMaxRequests(5),
		)

		if cb.failureThreshold != 10 {
			t.Errorf("failureThreshold = %d, want 10", cb.failureThreshold)
		}
		if cb.recoveryTimeout != 30*time.Second {
			t.Errorf("recoveryTimeout = %v, want 30s", cb.recoveryTimeout)
		}
		if cb.halfOpenMaxRequests != 5 {
			t.Errorf("halfOpenMaxRequests = %d, want 5", cb.halfOpenMaxRequests)
		}
	})
}

func TestCircuitBreaker_StateTransitions(t *testing.T) {
	t.Run("closed to open after threshold", func(t *testing.T) {
		cb := New(WithFailureThreshold(3))

		// Should start closed
		if cb.State() != StateClosed {
			t.Fatalf("initial state = %v, want StateClosed", cb.State())
		}

		// Record failures up to threshold
		cb.RecordFailure()
		cb.RecordFailure()
		if cb.State() != StateClosed {
			t.Errorf("state after 2 failures = %v, want StateClosed", cb.State())
		}

		// Third failure should open circuit
		cb.RecordFailure()
		if cb.State() != StateOpen {
			t.Errorf("state after 3 failures = %v, want StateOpen", cb.State())
		}
	})

	t.Run("open to half open after timeout", func(t *testing.T) {
		cb := New(
			WithFailureThreshold(1),
			WithRecoveryTimeout(10*time.Millisecond),
		)

		// Open the circuit
		cb.RecordFailure()
		if cb.State() != StateOpen {
			t.Fatalf("state = %v, want StateOpen", cb.State())
		}

		// Should reject requests immediately
		if cb.AllowRequest() {
			t.Error("AllowRequest() = true, want false (circuit open)")
		}

		// Wait for recovery timeout
		time.Sleep(20 * time.Millisecond)

		// Should transition to half-open and allow request
		if !cb.AllowRequest() {
			t.Error("AllowRequest() = false, want true (recovery timeout elapsed)")
		}
		if cb.State() != StateHalfOpen {
			t.Errorf("state = %v, want StateHalfOpen", cb.State())
		}
	})

	t.Run("half open to closed after successes", func(t *testing.T) {
		cb := New(
			WithFailureThreshold(1),
			WithRecoveryTimeout(10*time.Millisecond),
			WithHalfOpenMaxRequests(2),
		)

		// Open the circuit
		cb.RecordFailure()

		// Wait for recovery timeout
		time.Sleep(20 * time.Millisecond)
		cb.AllowRequest() // transition to half-open

		if cb.State() != StateHalfOpen {
			t.Fatalf("state = %v, want StateHalfOpen", cb.State())
		}

		// Record required successes
		cb.RecordSuccess()
		cb.RecordSuccess()

		// Should be closed again
		if cb.State() != StateClosed {
			t.Errorf("state after successes = %v, want StateClosed", cb.State())
		}
	})

	t.Run("half open to open on failure", func(t *testing.T) {
		cb := New(
			WithFailureThreshold(1),
			WithRecoveryTimeout(10*time.Millisecond),
			WithHalfOpenMaxRequests(3),
		)

		// Open the circuit
		cb.RecordFailure()

		// Wait for recovery timeout
		time.Sleep(20 * time.Millisecond)
		cb.AllowRequest() // transition to half-open

		if cb.State() != StateHalfOpen {
			t.Fatalf("state = %v, want StateHalfOpen", cb.State())
		}

		// Single failure should send back to open
		cb.RecordFailure()
		if cb.State() != StateOpen {
			t.Errorf("state after failure in half-open = %v, want StateOpen", cb.State())
		}
	})
}

func TestCircuitBreaker_AllowRequest(t *testing.T) {
	t.Run("closed allows all", func(t *testing.T) {
		cb := New()

		for i := 0; i < 100; i++ {
			if !cb.AllowRequest() {
				t.Errorf("AllowRequest() #%d = false, want true (closed state)", i+1)
			}
		}
	})

	t.Run("open rejects until timeout", func(t *testing.T) {
		cb := New(
			WithFailureThreshold(1),
			WithRecoveryTimeout(50*time.Millisecond),
		)

		cb.RecordFailure() // opens circuit

		// Should reject multiple requests
		for i := 0; i < 10; i++ {
			if cb.AllowRequest() {
				t.Errorf("AllowRequest() #%d = true, want false (open state)", i+1)
			}
		}

		// Stats should show rejections
		stats := cb.Stats()
		if stats["total_rejected"].(int64) != 10 {
			t.Errorf("total_rejected = %v, want 10", stats["total_rejected"])
		}
	})

	t.Run("half open allows limited requests", func(t *testing.T) {
		cb := New(
			WithFailureThreshold(1),
			WithRecoveryTimeout(10*time.Millisecond),
			WithHalfOpenMaxRequests(2),
		)

		// Open and transition to half-open
		cb.RecordFailure()
		time.Sleep(20 * time.Millisecond)
		cb.AllowRequest() // to half-open

		// Should allow halfOpenMaxRequests
		if !cb.AllowRequest() {
			t.Error("first half-open request should be allowed")
		}
		if !cb.AllowRequest() {
			t.Error("second half-open request should be allowed")
		}

		// Record successes to close the circuit, then verify recovery
		cb.RecordSuccess()
		cb.RecordSuccess()

		// Circuit should now be closed
		if cb.State() != StateClosed {
			t.Errorf("state after successes = %v, want StateClosed", cb.State())
		}
	})
}

func TestCircuitBreaker_RecordSuccess(t *testing.T) {
	t.Run("resets consecutive failures in closed", func(t *testing.T) {
		cb := New(WithFailureThreshold(5))

		cb.RecordFailure()
		cb.RecordFailure()
		cb.RecordFailure()

		// Success should reset consecutive failures
		cb.RecordSuccess()

		// Should need 5 more failures to open
		for i := 0; i < 4; i++ {
			cb.RecordFailure()
		}
		if cb.State() != StateClosed {
			t.Errorf("state = %v, want StateClosed (success should have reset counter)", cb.State())
		}

		cb.RecordFailure()
		if cb.State() != StateOpen {
			t.Errorf("state = %v, want StateOpen", cb.State())
		}
	})
}

func TestCircuitBreaker_Stats(t *testing.T) {
	cb := New(
		WithFailureThreshold(3),
		WithRecoveryTimeout(30*time.Second),
	)

	// Generate some activity
	cb.AllowRequest()
	cb.AllowRequest()
	cb.RecordSuccess()
	cb.RecordSuccess()
	cb.RecordFailure()

	stats := cb.Stats()

	if stats["state"] != "closed" {
		t.Errorf("stats[state] = %v, want closed", stats["state"])
	}
	if stats["total_requests"].(int64) != 2 {
		t.Errorf("stats[total_requests] = %v, want 2", stats["total_requests"])
	}
	if stats["total_successes"].(int64) != 2 {
		t.Errorf("stats[total_successes] = %v, want 2", stats["total_successes"])
	}
	if stats["total_failures"].(int64) != 1 {
		t.Errorf("stats[total_failures] = %v, want 1", stats["total_failures"])
	}
	if stats["failure_threshold"].(int) != 3 {
		t.Errorf("stats[failure_threshold] = %v, want 3", stats["failure_threshold"])
	}
	if stats["recovery_timeout"] != "30s" {
		t.Errorf("stats[recovery_timeout] = %v, want 30s", stats["recovery_timeout"])
	}
}

func TestState_String(t *testing.T) {
	tests := []struct {
		state State
		want  string
	}{
		{StateClosed, "closed"},
		{StateOpen, "open"},
		{StateHalfOpen, "half-open"},
		{State(999), "unknown"},
	}

	for _, tt := range tests {
		t.Run(tt.want, func(t *testing.T) {
			got := tt.state.String()
			if got != tt.want {
				t.Errorf("State(%d).String() = %q, want %q", tt.state, got, tt.want)
			}
		})
	}
}

func TestCircuitBreaker_Concurrency(t *testing.T) {
	cb := New()

	done := make(chan bool)

	// Spawn multiple goroutines to test thread safety
	for i := 0; i < 10; i++ {
		go func() {
			for j := 0; j < 100; j++ {
				cb.AllowRequest()
				if j%2 == 0 {
					cb.RecordSuccess()
				} else {
					cb.RecordFailure()
				}
			}
			done <- true
		}()
	}

	// Wait for all goroutines
	for i := 0; i < 10; i++ {
		<-done
	}

	// Stats should be consistent
	stats := cb.Stats()
	if stats["total_requests"].(int64) != 1000 {
		t.Errorf("total_requests = %v, want 1000", stats["total_requests"])
	}
}

func TestCircuitBreaker_EndToEnd(t *testing.T) {
	cb := New(
		WithFailureThreshold(2),
		WithRecoveryTimeout(10*time.Millisecond),
		WithHalfOpenMaxRequests(2),
	)

	// Phase 1: Normal operation
	if cb.State() != StateClosed {
		t.Fatal("initial state not closed")
	}
	if !cb.AllowRequest() {
		t.Fatal("closed circuit should allow requests")
	}

	// Phase 2: Failures open the circuit
	cb.RecordFailure()
	cb.RecordFailure()
	if cb.State() != StateOpen {
		t.Fatalf("state = %v, want open", cb.State())
	}
	if cb.AllowRequest() {
		t.Error("open circuit should reject requests")
	}

	// Phase 3: Recovery timeout transitions to half-open
	time.Sleep(20 * time.Millisecond)
	if !cb.AllowRequest() {
		t.Fatal("half-open circuit should allow requests after timeout")
	}
	if cb.State() != StateHalfOpen {
		t.Fatalf("state = %v, want half-open", cb.State())
	}

	// Phase 4: Successes close the circuit
	cb.RecordSuccess()
	cb.RecordSuccess()
	if cb.State() != StateClosed {
		t.Fatalf("state = %v, want closed", cb.State())
	}
	if !cb.AllowRequest() {
		t.Fatal("closed circuit should allow requests")
	}
}
