package rlimiter

func NewIPRateLimiter(rate *Rate, prefix string) RateLimiter {
	return &rateLimiter{
		rate:   rate,
		prefix: prefix,
	}
}
