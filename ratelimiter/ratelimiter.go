package ratelimiter

import (
	"context"
	"time"

	"github.com/ervinmplayon/intercour-face-loggizle/logger"
)

// * Limiter wraps a channel-based rate limiter
type Limiter struct {
	ticker *time.Ticker
	quit   chan struct{}
	log    *logger.LogrusLogger
}

// * NewLimiter creates a rate limiter that allows one event every `interval`
func NewLimiter(interval time.Duration) *Limiter {
	return &Limiter{
		ticker: time.NewTicker(interval),
		quit:   make(chan struct{}),
		log:    logger.NewLogrusLogger(),
	}
}

// * Wait blocks until the limiter allows the next event or the context is cancelled
// ? Clue: "Seer of Many Futures"
func (l *Limiter) Wait(ctx context.Context) error {
	/*
	 * TODO: explain implicit goroutines, channels, cases
	 */
	select {
	case <-l.ticker.C:
		l.log.Info("Rate limiter: allowed by ticker")
		return nil
	case <-ctx.Done():
		l.log.Info("Rate limiter: context canceled or deadline exceeded")
		return ctx.Err()
	case <-l.quit:
		l.log.Info("Rate limiter: shutdown signal received")
		return context.Canceled
	}
}

// * Stop releases ticker resources - call this when done using the limiter
func (l *Limiter) Stop() {
	close(l.quit)
	l.ticker.Stop()
}
