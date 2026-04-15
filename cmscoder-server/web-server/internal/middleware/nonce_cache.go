package middleware

import (
	"sync"
	"time"
)

// NonceCache provides thread-safe nonce storage with TTL.
// Used to prevent replay attacks on the model-token endpoint.
type NonceCache struct {
	mu      sync.RWMutex
	nonces  map[string]time.Time
	ttl     time.Duration
	cleaned time.Time
}

// NewNonceCache creates a new nonce cache with the specified TTL.
func NewNonceCache(ttl time.Duration) *NonceCache {
	return &NonceCache{
		nonces:  make(map[string]time.Time),
		ttl:     ttl,
		cleaned: time.Now(),
	}
}

// IsValid checks if a nonce is valid (not used before and within TTL window).
// Returns true if the nonce is new and valid, false if already used or expired.
func (c *NonceCache) IsValid(nonce string, timestamp int64) bool {
	c.mu.Lock()
	defer c.mu.Unlock()

	// Periodic cleanup to prevent memory growth
	if time.Since(c.cleaned) > c.ttl {
		c.cleanup()
	}

	// Check if nonce already exists
	if _, exists := c.nonces[nonce]; exists {
		return false
	}

	// Validate timestamp is within acceptable window (±30 seconds from now)
	now := time.Now().Unix()
	if timestamp < now-30 || timestamp > now+30 {
		return false
	}

	// Store the nonce
	c.nonces[nonce] = time.Now()
	return true
}

// cleanup removes expired entries from the cache.
func (c *NonceCache) cleanup() {
	cutoff := time.Now().Add(-c.ttl)
	for nonce, added := range c.nonces {
		if added.Before(cutoff) {
			delete(c.nonces, nonce)
		}
	}
	c.cleaned = time.Now()
}
