package rlimiter

import (
	"sync/atomic"

	"github.com/redis/go-redis/v9"
)

var (
	incrementScript = redis.NewScript(`
		local current
		current = tonumber(redis.call("incr", KEYS[1]))
		if current == 1 then
			redis.call("expire", KEYS[1], ARGV[1])
		end
		return current
	`)
	// client for connecting to redis database
	client atomic.Pointer[redis.Client]
)

func init() {
	// Set default client
	SetClient(redis.NewClient(&redis.Options{
		Addr: "localhost:6379",
		DB:   0, // use default DB
	}))
}

func SetClient(c *redis.Client) {
	client.Store(c)
}

func GetClient() *redis.Client {
	if c := client.Load(); c != nil {
		return c
	}
	panic("redis client is not initialized")
}
