package rlimiter

import (
	"github.com/gin-gonic/gin"
)

func NewParamRateLimiter(rate *Rate, prefix, paramName string) GinRateLimiter {
	return &paramRateLimiter{
		limiter:   NewRateLimiter(rate, prefix),
		paramName: paramName,
	}
}

type paramRateLimiter struct {
	limiter   RateLimiter
	paramName string
}

func (r *paramRateLimiter) Key(c *gin.Context) string {
	return c.Param(r.paramName)
}

func (r *paramRateLimiter) IncrementUsage(key string) (bool, error) {
	return r.limiter.IncrementUsage(key)
}
