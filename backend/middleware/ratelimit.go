package middleware

import (
	"fmt"
	"net/http"
	"sync"
	"time"
	"xsha-backend/i18n"

	"github.com/gin-gonic/gin"
)

// RateLimitEntry rate limit entry
type RateLimitEntry struct {
	Count     int       // Request count
	FirstTime time.Time // First request time
	LastTime  time.Time // Last request time
}

// RateLimiter rate limiter
type RateLimiter struct {
	mu      sync.RWMutex
	entries map[string]*RateLimitEntry
	limit   int           // Maximum requests within time window
	window  time.Duration // Time window
}

// NewRateLimiter creates a new rate limiter
func NewRateLimiter(limit int, window time.Duration) *RateLimiter {
	rl := &RateLimiter{
		entries: make(map[string]*RateLimitEntry),
		limit:   limit,
		window:  window,
	}

	// Start goroutine to clean expired entries
	go rl.cleanup()

	return rl
}

// IsAllowed checks if the request is allowed
func (rl *RateLimiter) IsAllowed(key string) bool {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	now := time.Now()
	entry, exists := rl.entries[key]

	if !exists {
		// First request
		rl.entries[key] = &RateLimitEntry{
			Count:     1,
			FirstTime: now,
			LastTime:  now,
		}
		return true
	}

	// Check if beyond time window
	if now.Sub(entry.FirstTime) > rl.window {
		// Reset counter
		entry.Count = 1
		entry.FirstTime = now
		entry.LastTime = now
		return true
	}

	// Update last request time
	entry.LastTime = now

	// Check if limit exceeded
	if entry.Count >= rl.limit {
		return false
	}

	// Increment counter
	entry.Count++
	return true
}

// GetRemainingTime gets remaining wait time
func (rl *RateLimiter) GetRemainingTime(key string) time.Duration {
	rl.mu.RLock()
	defer rl.mu.RUnlock()

	entry, exists := rl.entries[key]
	if !exists {
		return 0
	}

	elapsed := time.Since(entry.FirstTime)
	if elapsed >= rl.window {
		return 0
	}

	return rl.window - elapsed
}

// cleanup cleans expired entries
func (rl *RateLimiter) cleanup() {
	ticker := time.NewTicker(time.Minute * 5) // Clean every 5 minutes
	defer ticker.Stop()

	for range ticker.C {
		rl.mu.Lock()
		now := time.Now()
		for key, entry := range rl.entries {
			// Clean entries that exceed time window and have been inactive for long
			if now.Sub(entry.LastTime) > rl.window*2 {
				delete(rl.entries, key)
			}
		}
		rl.mu.Unlock()
	}
}

// Global rate limiter instances
var (
	loginRateLimiter *RateLimiter
	once             sync.Once
)

// getLoginRateLimiter gets login rate limiter instance
func getLoginRateLimiter() *RateLimiter {
	once.Do(func() {
		// Allow maximum 5 login attempts within 15 minutes
		loginRateLimiter = NewRateLimiter(5, 15*time.Minute)
	})
	return loginRateLimiter
}

// LoginRateLimitMiddleware login rate limit middleware
func LoginRateLimitMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		limiter := getLoginRateLimiter()
		lang := GetLangFromContext(c)

		// Use client IP as rate limit key
		clientIP := c.ClientIP()

		if !limiter.IsAllowed(clientIP) {
			remainingTime := limiter.GetRemainingTime(clientIP)

			c.JSON(http.StatusTooManyRequests, gin.H{
				"error":               i18n.T(lang, "login.rate_limit"),
				"retry_after_seconds": int(remainingTime.Seconds()),
				"retry_after":         fmt.Sprintf("%.0f minutes", remainingTime.Minutes()),
			})
			c.Abort()
			return
		}

		c.Next()
	}
}
