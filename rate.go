package rlimiter

import (
	"context"
	"fmt"
	"time"
)

var globalPrefix = "rlimiter:"

// SetGlobalPrefix sets a global prefix for all rate limiter keys.
// This can be useful to avoid key collisions in Redis when using multiple applications or services.
// The prefix will be automatically suffixed with a colon if it does not already end with one.
// Remove the global prefix if an empty string is provided.
func SetGlobalPrefix(prefix string) {
	if prefix != "" && prefix[len(prefix)-1] != ':' {
		prefix += ":"
	}
	globalPrefix = prefix
}

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

type rateLimiter struct {
	rate   *Rate
	prefix string
}

func (r *rateLimiter) IncrementUsage(key string) (bool, error) {
	name := fmt.Sprintf("%s%s:%s", globalPrefix, r.prefix, key)
	cmd := incrementScript.Run(context.Background(), GetClient(), []string{name}, r.rate.Window.Seconds())
	usage, err := cmd.Int64()
	if err != nil {
		return true, err
	}
	if usage > r.rate.Limit {
		return false, nil
	}
	return true, nil
}
