package middleware

import (
	"sync"
	"time"

	"github.com/gogf/gf/v2/net/ghttp"
)

// RateLimiter provides simple in-memory rate limiting.
type RateLimiter struct {
	mu      sync.Mutex
	buckets map[string]*bucket
	rate    int           // max requests per window
	window  time.Duration // time window
}

type bucket struct {
	count     int
	resetAt   time.Time
}

// NewRateLimiter creates a new rate limiter.
func NewRateLimiter(rate int, window time.Duration) *RateLimiter {
	return &RateLimiter{
		buckets: make(map[string]*bucket),
		rate:    rate,
		window:  window,
	}
}

// Middleware returns the rate limiting middleware handler.
func (rl *RateLimiter) Middleware(r *ghttp.Request) {
	key := r.GetClientIp()

	rl.mu.Lock()
	b, ok := rl.buckets[key]
	if !ok || time.Now().After(b.resetAt) {
		b = &bucket{
			count:   1,
			resetAt: time.Now().Add(rl.window),
		}
		rl.buckets[key] = b
		rl.mu.Unlock()
		r.Middleware.Next()
		return
	}

	b.count++
	rl.mu.Unlock()

	if b.count > rl.rate {
		r.Response.WriteStatus(429)
		r.Response.WriteJson(ghttp.DefaultHandlerResponse{
			Code:    429,
			Message: "too many requests, please try again later",
		})
		return
	}

	r.Middleware.Next()
}
