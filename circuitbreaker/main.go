package main

import (
	"errors"
	"fmt"
	"sync"
	"time"
)

type CircuitBreaker struct {
	failures     int
	maxFailures  int
	resetTimeout time.Duration
	lastFailure  time.Time
	state        string
	mutex        sync.Mutex
}

func NewCircuitBreaker(maxFailures int, resetTimeout time.Duration) *CircuitBreaker {
	return &CircuitBreaker{
		maxFailures:  maxFailures,
		resetTimeout: resetTimeout,
		state:        "closed",
	}
}

func (cb *CircuitBreaker) Call(service func() error) error {
	cb.mutex.Lock()
	defer cb.mutex.Unlock()

	if cb.state == "open" {
		if time.Since(cb.lastFailure) > cb.resetTimeout {
			cb.state = "half-open"
		} else {
			return errors.New("circuit breaker is open")
		}
	}

	err := service()
	if err != nil {
		cb.failures++
		cb.lastFailure = time.Now()
		if cb.failures >= cb.maxFailures {
			cb.state = "open"
		}
		return err
	}

	cb.failures = 0
	cb.state = "closed"
	return nil
}

func main() {
	cb := NewCircuitBreaker(3, 5*time.Second)

	service := func() error {
		// Simulate a service call
		return errors.New("service failure")
	}

	for i := 0; i < 10; i++ {
		err := cb.Call(service)
		if err != nil {
			fmt.Println("Call failed:", err)
		} else {
			fmt.Println("Call succeeded")
		}
		time.Sleep(1 * time.Second)
	}
}
