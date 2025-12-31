// Package middleware provides HTTP middleware functions.
package middleware

import (
	"context"
	"errors"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis/v8"
	"go.uber.org/zap"
)

// RateLimiter provides rate limiting functionality using Redis.
type RateLimiter struct {
	redis  *redis.Client
	logger *zap.Logger
	limit  int
	window time.Duration
}

// NewRateLimiter creates a new rate limiter.
func NewRateLimiter(redis *redis.Client, logger *zap.Logger, limit int, window time.Duration) *RateLimiter {
	return &RateLimiter{
		redis:  redis,
		logger: logger,
		limit:  limit,
		window: window,
	}
}

// Limit returns a middleware that rate limits requests.
func (r *RateLimiter) Limit() gin.HandlerFunc {
	return func(c *gin.Context) {
		if r.redis == nil {
			c.Next()
			return
		}

		key := "rate_limit:" + c.ClientIP()
		ctx := context.Background()

		// Increment counter
		count, err := r.redis.Incr(ctx, key).Result()
		if err != nil {
			r.logger.Error("rate limiter redis error", zap.Error(err))
			c.Next()
			return
		}

		// Set expiration on first request
		if count == 1 {
			r.redis.Expire(ctx, key, r.window)
		}

		// Check if limit exceeded
		if int(count) > r.limit {
			c.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{
				"error": "Rate limit exceeded",
			})
			return
		}

		// Set rate limit headers
		c.Writer.Header().Set("X-RateLimit-Limit", string(rune(r.limit)))
		c.Writer.Header().Set("X-RateLimit-Remaining", string(rune(r.limit-int(count))))

		c.Next()
	}
}

// LoginRateLimiter provides rate limiting specifically for login attempts.
type LoginRateLimiter struct {
	redis  *redis.Client
	logger *zap.Logger
	limit  int
	window time.Duration
}

// NewLoginRateLimiter creates a new login rate limiter.
func NewLoginRateLimiter(redis *redis.Client, logger *zap.Logger, limit int, window time.Duration) *LoginRateLimiter {
	return &LoginRateLimiter{
		redis:  redis,
		logger: logger,
		limit:  limit,
		window: window,
	}
}

// Limit returns a middleware that rate limits login attempts.
func (r *LoginRateLimiter) Limit() gin.HandlerFunc {
	return func(c *gin.Context) {
		if r.redis == nil {
			c.Next()
			return
		}

		key := "login_limit:" + c.ClientIP()
		ctx := context.Background()

		// Get current count
		count, err := r.redis.Get(ctx, key).Int()
		if err != nil && !errors.Is(err, redis.Nil) {
			r.logger.Error("login rate limiter redis error", zap.Error(err))
			c.Next()
			return
		}

		// Check if limit exceeded
		if count >= r.limit {
			c.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{
				"error": "Too many login attempts. Please try again later.",
			})
			return
		}

		c.Next()

		// Increment counter after request (for failed login tracking)
		// This would be called from the auth handler on failure
	}
}

// RecordFailedLogin records a failed login attempt.
func (r *LoginRateLimiter) RecordFailedLogin(ctx context.Context, ip string) error {
	if r.redis == nil {
		return nil
	}

	key := "login_limit:" + ip
	pipe := r.redis.TxPipeline()
	pipe.Incr(ctx, key)
	pipe.Expire(ctx, key, r.window)
	_, err := pipe.Exec(ctx)
	return err
}

// ClearFailedLogin clears the failed login counter for an IP.
func (r *LoginRateLimiter) ClearFailedLogin(ctx context.Context, ip string) error {
	if r.redis == nil {
		return nil
	}

	key := "login_limit:" + ip
	return r.redis.Del(ctx, key).Err()
}
