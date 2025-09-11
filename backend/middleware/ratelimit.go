package middleware

import (
	"fmt"
	"net/http"
	"sync"
	"time"
	"xsha-backend/i18n"
	"xsha-backend/utils"

	"github.com/gin-gonic/gin"
)

type RateLimitEntry struct {
	Count     int
	FirstTime time.Time
	LastTime  time.Time
}

type RateLimiter struct {
	mu      sync.RWMutex
	entries map[string]*RateLimitEntry
	limit   int
	window  time.Duration
}

func NewRateLimiter(limit int, window time.Duration) *RateLimiter {
	rl := &RateLimiter{
		entries: make(map[string]*RateLimitEntry),
		limit:   limit,
		window:  window,
	}

	go rl.cleanup()

	return rl
}

func (rl *RateLimiter) IsAllowed(key string) bool {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	now := utils.Now()
	entry, exists := rl.entries[key]

	if !exists {
		rl.entries[key] = &RateLimitEntry{
			Count:     1,
			FirstTime: now,
			LastTime:  now,
		}
		return true
	}

	if now.Sub(entry.FirstTime) > rl.window {
		entry.Count = 1
		entry.FirstTime = now
		entry.LastTime = now
		return true
	}

	entry.LastTime = now

	if entry.Count >= rl.limit {
		return false
	}

	entry.Count++
	return true
}

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

func (rl *RateLimiter) cleanup() {
	ticker := time.NewTicker(time.Minute * 5)
	defer ticker.Stop()

	for range ticker.C {
		rl.mu.Lock()
		now := utils.Now()
		for key, entry := range rl.entries {
			if now.Sub(entry.LastTime) > rl.window*2 {
				delete(rl.entries, key)
			}
		}
		rl.mu.Unlock()
	}
}

var (
	loginRateLimiter *RateLimiter
	once             sync.Once
)

func getLoginRateLimiter() *RateLimiter {
	once.Do(func() {
		loginRateLimiter = NewRateLimiter(60, 60*time.Minute)
	})
	return loginRateLimiter
}

func LoginRateLimitMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		limiter := getLoginRateLimiter()
		lang := GetLangFromContext(c)

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
