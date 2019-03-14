package rlimiter

import "github.com/go-redis/redis"

var (
	incrementScript = redis.NewScript(`
		local current
		current = tonumber(redis.call("incr", KEYS[1]))
		if current == 1 then
			redis.call("expire", KEYS[1], ARGV[1])
		end
		return current
  `)
	// Client for connecting to redis database
	Client = redis.NewClient(&redis.Options{
		Addr: "localhost:6379",
		DB:   0, // use default DB
	})
)
