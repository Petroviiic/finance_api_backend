package ratelimiter

import (
	"sync"
	"time"
)

type clientStats struct {
	count     int
	expiresAt time.Time
}
type FixedWindowRateLimiter struct {
	sync.RWMutex
	clients map[string]*clientStats
	limit   int
	window  time.Duration
}

func NewFixedWindowLimiter(limit int, window time.Duration) *FixedWindowRateLimiter {
	return &FixedWindowRateLimiter{
		clients: make(map[string]*clientStats),
		limit:   limit,
		window:  window,
	}
}

func (rl *FixedWindowRateLimiter) Allow(v any) (bool, time.Duration) {
	ip, ok := v.(string)
	if !ok {
		return false, 0
	}

	rl.RLock()
	defer rl.RUnlock()

	now := time.Now()
	stats, exists := rl.clients[ip]

	if !exists || now.After(stats.expiresAt) {
		rl.clients[ip] = &clientStats{
			count:     1,
			expiresAt: now.Add(rl.window),
		}
		return true, 0
	}
	if stats.count < rl.limit {
		stats.count++
		return true, 0
	}
	return false, time.Until(stats.expiresAt)
}

func (rl *FixedWindowRateLimiter) Cleanup() {
	ticker := time.NewTicker(rl.window)
	go func() {
		for range ticker.C {
			rl.Lock()
			now := time.Now()
			for ip, stats := range rl.clients {
				if now.After(stats.expiresAt) {
					delete(rl.clients, ip)
				}
			}
			rl.Unlock()
		}
	}()
}
