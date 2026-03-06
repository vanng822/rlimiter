package rlimiter

import (
	"context"
	"fmt"
	"net/http"
	"slices"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
)

var autobanPrefix = "autoban"
var autobannedPrefix = "autobanned"

/*
*
AutoBan is a Gin middleware that automatically bans IP addresses that exceed a specified rate limit or return certain HTTP status codes. It uses Redis to track the number of requests and manage banned IPs.

Parameters:
- rate: A pointer to a Rate struct that defines the limit and window for the rate limiter. If nil, a default rate of 100 requests per minute is used.
- responseStatus: The HTTP status code to return when an IP is banned. Default is 429 Too Many Requests.
- statuses: A variadic list of HTTP status codes that, when returned by the server, will count towards the rate limit. If empty, it defaults to 404 Not Found.

The middleware checks if the client's IP address is currently banned before processing the request. If the IP is banned, it responds with the specified responseStatus. After processing the request, it checks if the response status code matches any of the specified statuses. If it does, it increments the usage count for that IP address. If the usage exceeds the defined rate limit, the IP address is banned for a specified duration.
*/
func AutoBan(rate *Rate, responseStatus int, banDuration time.Duration, statuses ...int) gin.HandlerFunc {
	if rate == nil {
		rate = &Rate{
			Limit:  100,
			Window: time.Minute,
		}
	}
	rater := NewRateLimiter(rate, autobanPrefix)

	if len(statuses) == 0 {
		statuses = []int{http.StatusNotFound} // Default to 429 Too Many Requests
	}

	isBanned := func(ip string) (bool, error) {
		key := fmt.Sprintf("%s%s:%s", globalPrefix, autobannedPrefix, ip)
		result, err := GetClient().Get(context.Background(), key).Result()
		if err == nil && result == "1" {
			return true, nil
		}
		if err != nil && err != redis.Nil {
			return false, err
		}
		return false, nil
	}

	banIp := func(ip string) {
		key := fmt.Sprintf("%s%s:%s", globalPrefix, autobannedPrefix, ip)
		// Set the key with an expiration time of 1 minute
		err := GetClient().Set(context.Background(), key, "1", banDuration).Err()
		if err != nil {
			fmt.Printf("Error banning IP %s: %v\n", ip, err)
		}
	}

	return func(c *gin.Context) {
		clientIP := c.ClientIP()
		if banned, err := isBanned(clientIP); err == nil && banned {
			c.AbortWithStatus(responseStatus)
			return
		}

		c.Next()

		status := c.Writer.Status()
		if slices.Contains(statuses, status) {
			ok, err := rater.IncrementUsage(clientIP)
			if err != nil {
				return
			}
			if !ok {
				banIp(clientIP)
				return
			}
		}
	}
}
