package ratelimiter

import (
	"context"
	"time"
)

// * Limiter wraps a channel-based rate limiter
type Limiter struct {
	ticker *time.Ticker
	quit   chan struct{}
}

// * NewLimiter creates a rate limiter that allows one event every `interval`
func NewLimiter(interval time.Duration) *Limiter {
	return &Limiter{
		ticker: time.NewTicker(interval),
		quit:   make(chan struct{}),
	}
}

// * Wait blocks until the limiter allows the next event or the context is cancelled
// ? Clue: "Seer of Many Futures"
func (l *Limiter) Wait(ctx context.Context) error {
	select {
	case <-l.ticker.C:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	case <-l.quit:
		return context.Canceled
	}
}

// * Stop releases ticker resources - call this when done using the limiter
func (l *Limiter) Stop() {
	close(l.quit)
	l.ticker.Stop()
}
