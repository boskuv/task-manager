package circuitbreaker

import (
	"errors"
	"sync"
	"time"
)

var ErrOpen = errors.New("circuit breaker open")

// State describes the breaker lifecycle.
type State int

const (
	StateClosed State = iota
	StateOpen
	StateHalfOpen
)

// Config controls when the breaker opens and closes.
type Config struct {
	FailureThreshold uint32
	SuccessThreshold uint32
	OpenTimeout      time.Duration
}

// Breaker protects downstream calls from repeated failures.
type Breaker struct {
	cfg Config
	now func() time.Time

	mu         sync.Mutex
	state      State
	failures   uint32
	successes  uint32
	openedAt   time.Time
}

// New creates a breaker with defaults for unset config fields.
func New(cfg Config) *Breaker {
	if cfg.FailureThreshold == 0 {
		cfg.FailureThreshold = 5
	}
	if cfg.SuccessThreshold == 0 {
		cfg.SuccessThreshold = 1
	}
	if cfg.OpenTimeout == 0 {
		cfg.OpenTimeout = 30 * time.Second
	}

	return &Breaker{
		cfg: cfg,
		now: time.Now,
	}
}

// State returns the current breaker state.
func (b *Breaker) State() State {
	b.mu.Lock()
	defer b.mu.Unlock()
	return b.state
}

// Execute runs fn when the breaker allows the call.
func (b *Breaker) Execute(fn func() error) error {
	if err := b.beforeCall(); err != nil {
		return err
	}

	err := fn()
	b.afterCall(err)
	return err
}

func (b *Breaker) beforeCall() error {
	b.mu.Lock()
	defer b.mu.Unlock()

	switch b.state {
	case StateClosed:
		return nil
	case StateOpen:
		if b.now().Sub(b.openedAt) >= b.cfg.OpenTimeout {
			b.state = StateHalfOpen
			b.successes = 0
			return nil
		}
		return ErrOpen
	case StateHalfOpen:
		return nil
	default:
		return nil
	}
}

func (b *Breaker) afterCall(err error) {
	b.mu.Lock()
	defer b.mu.Unlock()

	if err != nil {
		b.onFailure()
		return
	}
	b.onSuccess()
}

func (b *Breaker) onFailure() {
	switch b.state {
	case StateClosed:
		b.failures++
		if b.failures >= b.cfg.FailureThreshold {
			b.open()
		}
	case StateHalfOpen:
		b.open()
	}
}

func (b *Breaker) onSuccess() {
	switch b.state {
	case StateClosed:
		b.failures = 0
	case StateHalfOpen:
		b.successes++
		if b.successes >= b.cfg.SuccessThreshold {
			b.state = StateClosed
			b.failures = 0
			b.successes = 0
		}
	}
}

func (b *Breaker) open() {
	b.state = StateOpen
	b.openedAt = b.now()
	b.failures = 0
	b.successes = 0
}
