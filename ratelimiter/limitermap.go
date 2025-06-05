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

func NewLimiterMap(limit int, interval time.Duration) *LimiterMap {
	return &LimiterMap{
		limiters: make(map[string]*HardThrottleLimiter),
		limit:    limit,
		interval: interval,
	}
}

func (lm *LimiterMap) getLimiter(ip string) *HardThrottleLimiter {
	lm.mu.Lock()
	defer lm.mu.Unlock()

	if limiter, exists := lm.limiters[ip]; exists {
		return limiter
	}

	limiter := NewHardThrottleLimiter(lm.limit, lm.interval)
	lm.limiters[ip] = limiter
	return limiter
}

func (lm *LimiterMap) Allow(ip string) bool {
	// * The Allow() in this returned bool is the Allow() function of HardThrottleLimiter
	return lm.getLimiter(ip).Allow()
}

func (lm *LimiterMap) StartCleanup(maxIdle time.Duration, interval time.Duration) {
	go func() {
		ticker := time.NewTicker(interval)
		defer ticker.Stop()

		// * range "iterating " over ticker.C explanation found on README
		for range ticker.C {
			lm.mu.Lock()
			now := time.Now()
			for ip, limiter := range lm.limiters {
				limiter.mu.Lock()
				idle := now.Sub(limiter.lastSeen)
				limiter.mu.Unlock()
				if idle > maxIdle {
					delete(lm.limiters, ip)
				}
			}
			lm.mu.Unlock()
		}
	}()
}
