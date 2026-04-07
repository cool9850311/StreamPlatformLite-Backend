package middleware

import (
	"Go-Service/src/main/infrastructure/config"
	"context"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/ulule/limiter/v3"
)

// RateLimitByIP - IP 维度限流中间件
func RateLimitByIP(lim *limiter.Limiter) gin.HandlerFunc {
	return func(c *gin.Context) {
		if !config.AppConfig.RateLimit.Enabled {
			c.Next()
			return
		}

		// 获取客户端 IP (优先 X-Forwarded-For)
		ip := c.GetHeader("X-Forwarded-For")
		if ip == "" {
			ip = c.ClientIP()
		}

		limiterCtx, err := lim.Get(context.Background(), ip)
		if err != nil {
			// 限流器错误，fail-open (允许请求通过)
			c.Next()
			return
		}

		// 设置响应头
		c.Header("X-RateLimit-Limit", strconv.FormatInt(limiterCtx.Limit, 10))
		c.Header("X-RateLimit-Remaining", strconv.FormatInt(limiterCtx.Remaining, 10))
		c.Header("X-RateLimit-Reset", strconv.FormatInt(limiterCtx.Reset, 10))

		if limiterCtx.Reached {
			// Retry-After should be seconds to wait, not Unix timestamp
		retryAfterSeconds := limiterCtx.Reset - time.Now().Unix()
		if retryAfterSeconds < 0 {
			retryAfterSeconds = 0
		}
		c.Header("Retry-After", strconv.FormatInt(retryAfterSeconds, 10))
			c.JSON(http.StatusTooManyRequests, gin.H{
				"message":     "Rate limit exceeded. Please try again later.",
				"retry_after": retryAfterSeconds,
			})
			c.Abort()
			return
		}

		c.Next()
	}
}

// RateLimitByUserID - 用户ID 维度限流中间件
func RateLimitByUserID(lim *limiter.Limiter) gin.HandlerFunc {
	return func(c *gin.Context) {
		if !config.AppConfig.RateLimit.Enabled {
			c.Next()
			return
		}

		// 从 JWT claims 获取用户 ID
		userID, exists := c.Get("user_id")
		if !exists {
			// 未登录用户，跳过限流
			c.Next()
			return
		}

		key := fmt.Sprintf("user:%v", userID)
		limiterCtx, err := lim.Get(context.Background(), key)
		if err != nil {
			// fail-open
			c.Next()
			return
		}

		// 设置响应头
		c.Header("X-RateLimit-Limit", strconv.FormatInt(limiterCtx.Limit, 10))
		c.Header("X-RateLimit-Remaining", strconv.FormatInt(limiterCtx.Remaining, 10))
		c.Header("X-RateLimit-Reset", strconv.FormatInt(limiterCtx.Reset, 10))

		if limiterCtx.Reached {
			// Retry-After should be seconds to wait, not Unix timestamp
		retryAfterSeconds := limiterCtx.Reset - time.Now().Unix()
		if retryAfterSeconds < 0 {
			retryAfterSeconds = 0
		}
		c.Header("Retry-After", strconv.FormatInt(retryAfterSeconds, 10))
			c.JSON(http.StatusTooManyRequests, gin.H{
				"message":     "Rate limit exceeded. Please try again later.",
				"retry_after": retryAfterSeconds,
			})
			c.Abort()
			return
		}

		c.Next()
	}
}
