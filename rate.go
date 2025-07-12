package rlimiter

import (
	"context"
	"fmt"
	"time"
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

type rateLimiter struct {
	rate   *Rate
	prefix string
}

func (r *rateLimiter) IncrementUsage(key string) (bool, error) {
	name := fmt.Sprintf("%s:%s", r.prefix, key)
	cmd := incrementScript.Run(context.Background(), Client, []string{name}, r.rate.Window.Seconds())
	usage, err := cmd.Int64()
	if err != nil {
		return true, err
	}
	if usage > r.rate.Limit {
		return false, nil
	}
	return true, nil
}
