package rlimiter

import (
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

// Rate define window as describe in https://redis.io/commands/incr#pattern-rate-limiter-2
type Rate struct {
	Window time.Duration
	Limit  int64
}

func NewRateLimiter(rate *Rate, prefix string) RateLimiter {
	return &rateLimiter{
		rate:   rate,
		prefix: prefix,
	}
}

type RateLimiter interface {
	IncrementUsage(key string) (bool, error)
}

type rateLimiter struct {
	rate   *Rate
	prefix string
}

func (r *rateLimiter) IncrementUsage(key string) (bool, error) {
	name := fmt.Sprintf("%s:%s", r.prefix, key)
	cmd := incrementScript.Run(Client, []string{name}, r.rate.Window.Seconds())
	usage, err := cmd.Int64()
	if err != nil {
		return true, err
	}
	if usage > r.rate.Limit {
		return false, nil
	}
	return true, nil
}

func GinRateLimiter(limiter RateLimiter, methods []string) gin.HandlerFunc {

	return func(c *gin.Context) {
		if len(methods) == 0 || inStrings(c.Request.Method, methods) {
			ip := c.ClientIP()
			ok, _ := limiter.IncrementUsage(ip)
			if !ok {
				c.AbortWithStatusJSON(
					http.StatusTooManyRequests,
					gin.H{
						"status": "error",
						"error":  http.StatusText(http.StatusTooManyRequests)})
				return
			}
		}
		c.Next()
	}
}

func inStrings(needle string, haystack []string) bool {
	for _, m := range haystack {
		if needle == m {
			return true
		}
	}
	return false
}
