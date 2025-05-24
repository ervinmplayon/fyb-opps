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
	 * This select statement is acting on channels without explicit goroutines .
	 * But goroutines are involved under the hood, especially for time.Ticker and context.Context
	 * Regarding the channels - these are receive operations on channels, and are blocking by nature.
	 * This means that code will pause and wait here until one of the channels is ready.
	 * The select block is listening to channels that are driven by goroutines (either by Go stdlib or the app).
	 * I am not spawning goroutines per request. Instead, I rely on external signals from already running goroutines
	 * ***************************************************************************************************************************************
	 * case 1 - Ticker channel - The time.Ticker uses a goroutine internally to send ticks on the channel at regular intervals.
	 * There is a goroutine pushing values into that channel. This waits for the rate limiter to allow the next event.
	 * The core throttling logic.
	 * ***************************************************************************************************************************************
	 * case 2 - Context channel - The context package spawns goroutines when I set timeouts/deadlines or use
	 * context.WithCancel. Those gouroutines will close with the Done() channel when canceled or expired.
	 * This lets me abort early if the context (e.g. HTTP request) is canceled. This is critical for responsive services.
	 * ***************************************************************************************************************************************
	 * case 3 - Quit channel - This one is fully controlled by the app. I can close l.quit in another goroutine that I explicitly write.
	 * This is the shutdown signal, I can cancel all waits when the service is stopping.
	 * ***************************************************************************************************************************************
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
