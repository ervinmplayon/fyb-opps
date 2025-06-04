package ratelimiter

import (
	"sync"
	"time"
)

// * IP-based limiter manager, add a wrapper to manager rate limiters per IP.

type LimiterMap struct {
	limiters map[string]*HardThrottleLimiter
	mu       sync.Mutex
	limit    int
	interval time.Duration
}
