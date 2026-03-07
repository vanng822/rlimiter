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

type autoBanOption struct {
	Rate             *Rate
	ResponseStatus   int
	BanDuration      time.Duration
	Statuses         []int
	AutobanPrefix    string
	AutobannedPrefix string
}

type AutoBanOption func(*autoBanOption)

func AutoBanWithRate(rate *Rate) AutoBanOption {
	if rate == nil {
		panic("rate can not be nil")
	}
	return func(o *autoBanOption) {
		o.Rate = rate
	}
}

func AutoBanWithResponseStatus(status int) AutoBanOption {
	return func(o *autoBanOption) {
		o.ResponseStatus = status
	}
}

func AutoBanWithBanDuration(duration time.Duration) AutoBanOption {
	return func(o *autoBanOption) {
		o.BanDuration = duration
	}
}

func AutoBanWithStatuses(statuses ...int) AutoBanOption {
	if len(statuses) == 0 {
		panic("statuses must be set")
	}

	return func(o *autoBanOption) {
		o.Statuses = statuses
	}
}

func AutoBanWithAutobanPrefix(prefix string) AutoBanOption {
	return func(o *autoBanOption) {
		o.AutobanPrefix = prefix
	}
}

func AutoBanWithAutobannedPrefix(prefix string) AutoBanOption {
	return func(o *autoBanOption) {
		o.AutobannedPrefix = prefix
	}
}

func defaultAutoBanOption() *autoBanOption {
	return &autoBanOption{
		Rate:             &Rate{Limit: 20, Window: 2 * time.Minute},
		ResponseStatus:   http.StatusTooManyRequests,
		BanDuration:      30 * time.Minute,
		Statuses:         []int{http.StatusNotFound},
		AutobanPrefix:    "autoban",
		AutobannedPrefix: "autobanned",
	}
}

func AutoBan(opts ...AutoBanOption) gin.HandlerFunc {
	option := defaultAutoBanOption()

	for _, opt := range opts {
		opt(option)
	}

	rater := NewRateLimiter(option.Rate, option.AutobanPrefix)

	if len(option.Statuses) == 0 {
		option.Statuses = []int{http.StatusNotFound} // Default to 404 Not Found
	}

	isBanned := func(ip string) (bool, error) {
		key := fmt.Sprintf("%s%s:%s", globalPrefix, option.AutobannedPrefix, ip)
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
		key := fmt.Sprintf("%s%s:%s", globalPrefix, option.AutobannedPrefix, ip)
		err := GetClient().Set(context.Background(), key, "1", option.BanDuration).Err()
		if err != nil {
			fmt.Printf("Error banning IP %s: %v\n", ip, err)
		}
	}

	return func(c *gin.Context) {
		clientIP := c.ClientIP()
		if banned, err := isBanned(clientIP); err == nil && banned {
			c.AbortWithStatus(option.ResponseStatus)
			return
		}

		c.Next()

		status := c.Writer.Status()
		if slices.Contains(option.Statuses, status) {
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
