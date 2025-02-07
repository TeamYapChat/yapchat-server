package middleware

import (
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"golang.org/x/time/rate"
)

var (
	visitors = make(map[string]*rate.Limiter)
	mu       sync.Mutex
)

func RateLimitMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		ip := c.ClientIP()
		limiter := getRateLimiter(ip)

		if !limiter.Allow() {
			c.JSON(
				http.StatusTooManyRequests,
				gin.H{"error": "Too many requests. Please try again later."},
			)
			c.Abort()
			return
		}

		c.Next()
	}
}

func getRateLimiter(ip string) *rate.Limiter {
	mu.Lock()
	defer mu.Unlock()

	if limiter, exists := visitors[ip]; exists {
		return limiter
	}

	limiter := rate.NewLimiter(5, 2)
	visitors[ip] = limiter

	go func() {
		time.Sleep(time.Minute * 5)
		mu.Lock()
		delete(visitors, ip)
		mu.Unlock()
	}()

	return limiter
}
