package rlimiter

import "github.com/gin-gonic/gin"

func NewIPRateLimiter(rate *Rate, prefix string) GinRateLimiter {
	return &ipRateLimiter{
		rateLimiter{
			rate:   rate,
			prefix: prefix,
		},
	}
}

type ipRateLimiter struct {
	rateLimiter
}

func (r *ipRateLimiter) Key(c *gin.Context) string {
	return c.ClientIP()
}
