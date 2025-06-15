package ratelimiter

import (
	"sync"
	"time"
)

// * IP-based limiter manager, add a wrapper to manager rate limiters per IP.

type LimiterMap struct {
	limiters sync.Map // * Formerly key: string (IP), value: *HardThrottleLimiter
	limit    int
	interval time.Duration
}

func NewLimiterMap(limit int, interval time.Duration) *LimiterMap {
	return &LimiterMap{
		limit:    limit,
		interval: interval,
	}
}

func (lm *LimiterMap) getLimiter(ip string) *HardThrottleLimiter {
	/*
	 * This is a concurrency-safe pattern is called "doubled-checked" locking with sync.Map, which avoids race conditions.
	 * `sync.Map.Load()` is concurrency-safe, it's fine if multiple goroutines reads from the map at once.
	 */
	if limiter, ok := lm.limiters.Load(ip); ok {
		return limiter.(*HardThrottleLimiter)
	}

	/*
	 * Double-checked locking to avoid races. At this point,
	 * I might not be the only goroutine trying to do this step and thats ok.
	 */
	newLimiter := &HardThrottleLimiter{
		limit:    lm.limit,
		interval: lm.interval,
	}

	/*
	 * This is the safety gate.
	 * `sync.Map.LoadOrStore()` checks again if the IP is already present.
	 * IF it is already present, it returns the existing limiter (`actual`)
	 * IF not, it stores the `newLimiter` safely, and I GET THAT BACK.
	 * This method is atomic. It handles the locking internally so that I don't
	 * need to use `sync.Mutex` externally.
	 */
	actual, _ := lm.limiters.LoadOrStore(ip, newLimiter)
	return actual.(*HardThrottleLimiter)
}

func (lm *LimiterMap) Allow(ip string) (bool, int) {
	// * The Allow() in this returned bool is the Allow() function of HardThrottleLimiter
	return lm.getLimiter(ip).Allow()
}

func (lm *LimiterMap) StartCleanup(interval time.Duration, ttl time.Duration) {
	go func() {
		ticker := time.NewTicker(interval)
		defer ticker.Stop()

		// * range "iterating " over ticker.C explanation found on README
		for range ticker.C {
			now := time.Now()
			lm.limiters.Range(func(key, value any) bool {
				limiter := value.(*HardThrottleLimiter)
				limiter.mu.Lock()
				lastSeen := limiter.lastSeen
				limiter.mu.Unlock()
				if now.Sub(lastSeen) > ttl {
					lm.limiters.Delete(key)
				}
				return true
			})
		}
	}()
}
