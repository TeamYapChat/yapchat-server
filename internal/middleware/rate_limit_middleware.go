package middleware

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/charmbracelet/log"
	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
)

type RateLimiter struct {
	client   *redis.Client
	limiters map[string]RateLimitConfig
}

type RateLimitConfig struct {
	Limit  int
	Window time.Duration
}

func NewRateLimiter(redisClient *redis.Client) *RateLimiter {
	return &RateLimiter{
		client:   redisClient,
		limiters: make(map[string]RateLimitConfig),
	}
}

func (rl *RateLimiter) AddLimiter(name string, config RateLimitConfig) {
	rl.limiters[name] = config
}

func (rl *RateLimiter) Middleware(name string) gin.HandlerFunc {
	config, exists := rl.limiters[name]
	if !exists {
		log.Fatal("Rate limiter not found", "name", name)
	}

	return func(c *gin.Context) {
		ctx := context.Background()
		key := c.ClientIP()
		redisKey := fmt.Sprintf("ratelimit:%s:%s", name, key)

		// Increment the request count
		count, err := rl.client.Incr(ctx, redisKey).Result()
		if err != nil {
			c.AbortWithStatusJSON(
				http.StatusInternalServerError,
				gin.H{"error": "Internal server error"},
			)
			return
		}

		// Set expiration on the first request
		if count == 1 {
			rl.client.Expire(ctx, redisKey, config.Window)
		}

		if count > int64(config.Limit) {
			c.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{"error": "Rate limit exceeded"})
			return
		}

		c.Next()
	}
}
