// Package rlimiter is for rate limit on heavy endpoints
//
// import "github.com/vanng822/rlimiter"
// grbinder.BindVerb(group.Group("/login", rlimiter.GinRateLimit(
//   rlimiter.NewIPRateLimiter(
//     &rlimiter.Rate{Window: 10 * time.Second, Limit: 10},
//     "api.login"),
//   []string{"POST"})),
//   &loginHandler{})
package rlimiter
