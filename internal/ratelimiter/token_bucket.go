package ratelimiter

import (
	"strconv"
	"sync"
	"time"
)

type client struct {
	tokens     float64
	lastRefill time.Time
}

type TokenBucketRatelimiter struct {
	sync.RWMutex
	clients         map[int64]*client
	limit           float64
	tokensPerSecond float64
}

func NewTokenBuckerRatelimiter(limit, tokensPerMinute float64) *TokenBucketRatelimiter {
	return &TokenBucketRatelimiter{
		clients:         make(map[int64]*client),
		limit:           limit,
		tokensPerSecond: tokensPerMinute / 60,
	}
}

func (rl *TokenBucketRatelimiter) Allow(v any) (bool, time.Duration) {
	vString, ok := v.(string)
	if !ok {
		return false, 0
	}

	id, err := strconv.ParseInt(vString, 10, 64)
	if err != nil {
		return false, 0
	}

	rl.Lock()
	defer rl.Unlock()

	user, ok := rl.clients[id]
	now := time.Now()
	if !ok {
		rl.clients[id] = &client{
			tokens:     rl.limit - 1,
			lastRefill: now,
		}
		return true, 0
	}

	rl.refill(user, now)

	if user.tokens < 1 {
		return false, time.Duration((1-user.tokens)/rl.tokensPerSecond) * time.Second
	}

	user.tokens--
	return true, 0
}

func (rl *TokenBucketRatelimiter) refill(user *client, now time.Time) {
	elapsed := now.Sub(user.lastRefill).Seconds()

	user.tokens += elapsed * rl.tokensPerSecond

	if user.tokens > rl.limit {
		user.tokens = rl.limit
	}
	user.lastRefill = now
}

func (rl *TokenBucketRatelimiter) Cleanup() {
	ticker := time.NewTicker(1 * time.Hour)
	go func() {
		for range ticker.C {
			rl.Lock()
			for id, c := range rl.clients {
				if time.Since(c.lastRefill) > 1*time.Hour {
					delete(rl.clients, id)
				}
			}
			rl.Unlock()
		}
	}()
}
