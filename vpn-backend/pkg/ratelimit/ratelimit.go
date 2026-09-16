package ratelimit

import (
	"context"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis/v8"
	"github.com/vpn/backend/pkg/redis"
)

type RateLimiter struct {
	client *redis.Client
}

type LimitConfig struct {
	Requests int
	Window   time.Duration
	KeyPrefix string
}

var defaultLimits = map[string]LimitConfig{
	"auth:signup":   {Requests: 5, Window: 15 * time.Minute, KeyPrefix: "rl:signup"},
	"auth:login":    {Requests: 10, Window: 15 * time.Minute, KeyPrefix: "rl:login"},
	"auth:refresh":  {Requests: 20, Window: 15 * time.Minute, KeyPrefix: "rl:refresh"},
	"wireguard:config": {Requests: 10, Window: 1 * time.Hour, KeyPrefix: "rl:wg_config"},
	"servers:list":  {Requests: 30, Window: 1 * time.Minute, KeyPrefix: "rl:servers"},
	"billing:validate": {Requests: 5, Window: 15 * time.Minute, KeyPrefix: "rl:billing"},
}

func NewRateLimiter(client *redis.Client) *RateLimiter {
	return &RateLimiter{client: client}
}

func (rl *RateLimiter) Middleware(endpoint string) gin.HandlerFunc {
	config, exists := defaultLimits[endpoint]
	if !exists {
		config = LimitConfig{Requests: 100, Window: 1 * time.Minute, KeyPrefix: "rl:default"}
	}

	return func(c *gin.Context) {
		identifier := getClientIdentifier(c)
		key := fmt.Sprintf("%s:%s", config.KeyPrefix, identifier)

		allowed, retryAfter, err := rl.checkLimit(c.Request.Context(), key, config)
		if err != nil {
			// On Redis error, allow request (fail open)
			c.Next()
			return
		}

		if !allowed {
			c.Header("X-RateLimit-Limit", strconv.Itoa(config.Requests))
			c.Header("X-RateLimit-Remaining", "0")
			c.Header("Retry-After", strconv.Itoa(int(retryAfter.Seconds())))
			c.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{
				"error":       "rate limit exceeded",
				"retry_after": int(retryAfter.Seconds()),
			})
			return
		}

		remaining := config.Requests - rl.getCurrentCount(c.Request.Context(), key)
		c.Header("X-RateLimit-Limit", strconv.Itoa(config.Requests))
		c.Header("X-RateLimit-Remaining", strconv.Itoa(max(0, remaining)))
		c.Next()
	}
}

func (rl *RateLimiter) checkLimit(ctx context.Context, key string, config LimitConfig) (bool, time.Duration, error) {
	now := time.Now()
	windowStart := now.Truncate(config.Window)
	windowKey := fmt.Sprintf("%s:%d", key, windowStart.Unix())

	pipe := rl.client.TxPipeline()
	incr := pipe.Incr(ctx, windowKey)
	pipe.Expire(ctx, windowKey, config.Window+time.Second)
	_, err := pipe.Exec(ctx)
	if err != nil {
		return false, 0, err
	}

	count := incr.Val()
	if count > int64(config.Requests) {
		retryAfter := config.Window - time.Since(windowStart)
		return false, retryAfter, nil
	}

	return true, 0, nil
}

func (rl *RateLimiter) getCurrentCount(ctx context.Context, key string) int {
	now := time.Now()
	windowStart := now.Truncate(time.Minute)
	windowKey := fmt.Sprintf("%s:%d", key, windowStart.Unix())
	count, err := rl.client.Get(ctx, windowKey).Int()
	if err != nil {
		return 0
	}
	return count
}

func getClientIdentifier(c *gin.Context) string {
	// Try to get user ID from context (for authenticated endpoints)
	if userID, exists := c.Get("user_id"); exists {
		return "user:" + userID.(string)
	}
	// Fall back to IP address
	return "ip:" + c.ClientIP()
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}