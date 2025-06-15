package ratelimiter

import (
	"sync"
	"time"
)

type HardThrottleLimiter struct {
	mu        sync.Mutex
	limit     int
	interval  time.Duration
	timestamp time.Time
	lastSeen  time.Time
	count     int
}

func NewHardThrottleLimiter(limit int, interval time.Duration) *HardThrottleLimiter {
	return &HardThrottleLimiter{
		limit:     limit,
		interval:  interval,
		timestamp: time.Now(),
		lastSeen:  time.Now(),
		count:     0,
	}
}

// * Allow returns both the permission to proceeed (bool) and remaning quota:
// * if request is within rate limit, false otherwise.
func (htl *HardThrottleLimiter) Allow() (bool, int) {
	htl.mu.Lock()
	defer htl.mu.Unlock()

	// * Update lastSeen on every call. This ensures we can determine when the IP
	// * was last active, regardless of the rate-limiting outcome.
	now := time.Now()
	htl.lastSeen = now
	if now.Sub(htl.timestamp) > htl.interval {
		// * new window
		htl.timestamp = now
		htl.count = 0
	}

	if htl.count < htl.limit {
		htl.count++
		return true, htl.limit - htl.count
	}

	return false, 0 // * limit exceeded
}

// TODO: experiment if the "Seer of Many Futures" pattern can be leveraged
