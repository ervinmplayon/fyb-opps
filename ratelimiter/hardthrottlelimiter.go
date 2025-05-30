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
	count     int
}

func NewRateLimiter(limit int, interval time.Duration) *HardThrottleLimiter {
	return &HardThrottleLimiter{
		limit:     limit,
		interval:  interval,
		timestamp: time.Now(),
		count:     0,
	}
}
