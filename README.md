# rlimiter
Simple redis rate limit

# example

```go
// applied only to endpoint login and POST
grbinder.BindVerb(group.Group("/login", rlimiter.GinRateLimiter(
  rlimiter.NewRateLimiter(
    &rlimiter.Rate{Window: 10 * time.Second, Limit: 10},
    "api.login"),
  []string{"POST"})),
  &loginHandler{})

```
