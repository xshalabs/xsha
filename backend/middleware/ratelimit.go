package middleware

import (
	"fmt"
	"net/http"
	"sleep0-backend/i18n"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
)

// RateLimitEntry 限流条目
type RateLimitEntry struct {
	Count     int       // 请求次数
	FirstTime time.Time // 第一次请求时间
	LastTime  time.Time // 最后一次请求时间
}

// RateLimiter 限流器
type RateLimiter struct {
	mu      sync.RWMutex
	entries map[string]*RateLimitEntry
	limit   int           // 时间窗口内最大请求数
	window  time.Duration // 时间窗口
}

// NewRateLimiter 创建新的限流器
func NewRateLimiter(limit int, window time.Duration) *RateLimiter {
	rl := &RateLimiter{
		entries: make(map[string]*RateLimitEntry),
		limit:   limit,
		window:  window,
	}

	// 启动清理过期条目的协程
	go rl.cleanup()

	return rl
}

// IsAllowed 检查是否允许请求
func (rl *RateLimiter) IsAllowed(key string) bool {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	now := time.Now()
	entry, exists := rl.entries[key]

	if !exists {
		// 首次请求
		rl.entries[key] = &RateLimitEntry{
			Count:     1,
			FirstTime: now,
			LastTime:  now,
		}
		return true
	}

	// 检查是否超出时间窗口
	if now.Sub(entry.FirstTime) > rl.window {
		// 重置计数器
		entry.Count = 1
		entry.FirstTime = now
		entry.LastTime = now
		return true
	}

	// 更新最后请求时间
	entry.LastTime = now

	// 检查是否超出限制
	if entry.Count >= rl.limit {
		return false
	}

	// 增加计数
	entry.Count++
	return true
}

// GetRemainingTime 获取剩余等待时间
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

// cleanup 清理过期条目
func (rl *RateLimiter) cleanup() {
	ticker := time.NewTicker(time.Minute * 5) // 每5分钟清理一次
	defer ticker.Stop()

	for range ticker.C {
		rl.mu.Lock()
		now := time.Now()
		for key, entry := range rl.entries {
			// 清理超过时间窗口且很久没有活动的条目
			if now.Sub(entry.LastTime) > rl.window*2 {
				delete(rl.entries, key)
			}
		}
		rl.mu.Unlock()
	}
}

// 全局限流器实例
var (
	loginRateLimiter *RateLimiter
	once             sync.Once
)

// getLoginRateLimiter 获取登录限流器实例
func getLoginRateLimiter() *RateLimiter {
	once.Do(func() {
		// 15分钟内最多允许5次登录尝试
		loginRateLimiter = NewRateLimiter(5, 15*time.Minute)
	})
	return loginRateLimiter
}

// LoginRateLimitMiddleware 登录限流中间件
func LoginRateLimitMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		limiter := getLoginRateLimiter()
		lang := GetLangFromContext(c)

		// 使用客户端IP作为限流key
		clientIP := c.ClientIP()

		if !limiter.IsAllowed(clientIP) {
			remainingTime := limiter.GetRemainingTime(clientIP)

			c.JSON(http.StatusTooManyRequests, gin.H{
				"error":               i18n.T(lang, "login.rate_limit"),
				"retry_after_seconds": int(remainingTime.Seconds()),
				"retry_after":         fmt.Sprintf("%.0f分钟", remainingTime.Minutes()),
			})
			c.Abort()
			return
		}

		c.Next()
	}
}
