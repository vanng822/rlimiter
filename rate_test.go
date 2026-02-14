package rlimiter

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

var (
	testIP     = "65.121.1.232"
	testPrefix = "test.rlimiter"
)

func cleanRedisKey(key string) {
	GetClient().Del(context.Background(), key)
}

func getRedis(key string) string {
	val, _ := GetClient().Get(context.Background(), key).Result()
	return val
}

func TestInStrings(t *testing.T) {
	assert.True(t, inStrings("POST", []string{"GET", "POST"}))
	assert.False(t, inStrings("GET", []string{"PUT", "POST"}))
}

func TestRateLimiter(t *testing.T) {
	defer cleanRedisKey(fmt.Sprintf("%s%s:%s", globalPrefix, testPrefix, testIP))
	limiter := NewRateLimiter(&Rate{
		Window: 2 * time.Second,
		Limit:  1,
	}, testPrefix)
	ok, err := limiter.IncrementUsage(testIP)
	assert.Nil(t, err)
	assert.True(t, ok)
	ok, err = limiter.IncrementUsage(testIP)
	assert.Nil(t, err)
	assert.False(t, ok)
}

func TestRateLimiterError(t *testing.T) {
	key := fmt.Sprintf("%s%s:%s", globalPrefix, testPrefix, testIP)
	defer cleanRedisKey(key)
	limiter := NewRateLimiter(&Rate{
		Window: 2 * time.Second,
		Limit:  1,
	}, testPrefix)
	GetClient().Set(context.Background(), key, "notAnUint64", 2*time.Second)
	ok, err := limiter.IncrementUsage(testIP)
	assert.NotNil(t, err)
	// We should not block if redis is down
	// or programming error
	assert.True(t, ok)
}
