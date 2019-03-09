package rlimiter

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func GinRateLimit(limiter GinRateLimiter, methods []string) gin.HandlerFunc {

	return func(c *gin.Context) {
		if len(methods) == 0 || inStrings(c.Request.Method, methods) {
			key := limiter.Key(c)
			ok, _ := limiter.IncrementUsage(key)
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
