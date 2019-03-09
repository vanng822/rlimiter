package rlimiter

import (
	"github.com/gin-gonic/gin"
)

func NewPathRateLimiter(rate *Rate, prefix string) GinRateLimiter {
	return &pathRateLimiter{
		rateLimiter{
			rate:   rate,
			prefix: prefix,
		},
	}
}

type pathRateLimiter struct {
	rateLimiter
}

func (r *pathRateLimiter) Key(c *gin.Context) string {
	return c.Request.URL.Path
}
