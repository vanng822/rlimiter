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

```go
// Applied to any request method
limiter := rlimiter.GinRateLimiter(
  rlimiter.NewRateLimiter(
    &rlimiter.Rate{Window: 1 * time.Minute, Limit: 10},
    "api.hardwork"),
  []string{})

r := gin.Default()
r.GET("/hardwork", limiter, func(c *gin.Context) {
})
r.POST("/hardwork", limiter, func(c *gin.Context) {
})
r.PUT("/hardwork", limiter, func(c *gin.Context) {
})
```
