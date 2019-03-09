package rlimiter

import "github.com/gin-gonic/gin"

type RateLimiter interface {
	IncrementUsage(key string) (bool, error)
}

type GinRateLimiter interface {
	RateLimiter
	Key(c *gin.Context) string
}
