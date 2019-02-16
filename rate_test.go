package rlimiter

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

var (
	testIP     = "65.121.1.232"
	testPrefix = "test.rlimiter"
)

func cleanRedisKey(key string) {
	Client.Del(key)
}

func TestInStrings(t *testing.T) {
	assert.True(t, inStrings("POST", []string{"GET", "POST"}))
	assert.False(t, inStrings("GET", []string{"PUT", "POST"}))
}

func TestRateLimiter(t *testing.T) {
	defer cleanRedisKey(fmt.Sprintf("%s:%s", testPrefix, testIP))
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
	key := fmt.Sprintf("%s:%s", testPrefix, testIP)
	defer cleanRedisKey(key)
	limiter := NewRateLimiter(&Rate{
		Window: 2 * time.Second,
		Limit:  1,
	}, testPrefix)
	Client.Set(key, "notAnUint64", 2*time.Second)
	ok, err := limiter.IncrementUsage(testIP)
	assert.NotNil(t, err)
	// We should not block if redis is down
	// or programming error
	assert.True(t, ok)
}

func TestGinRateLimiter(t *testing.T) {
	defer cleanRedisKey(fmt.Sprintf("%s:%s", testPrefix, testIP))
	limiter := NewRateLimiter(&Rate{
		Window: 2 * time.Second,
		Limit:  1,
	}, testPrefix)
	ginHandleFunc := GinRateLimiter(limiter, []string{"GET"})
	c, _ := gin.CreateTestContext(httptest.NewRecorder())
	c.Request, _ = http.NewRequest("GET", "/", nil)
	c.Request.Header.Add("X-Forwarded-For", testIP)
	ginHandleFunc(c)
	assert.False(t, c.IsAborted())
	ginHandleFunc(c)
	assert.True(t, c.IsAborted())
}

func TestGinRateLimiterNotSameMethod(t *testing.T) {
	defer cleanRedisKey(fmt.Sprintf("%s:%s", testPrefix, testIP))
	limiter := NewRateLimiter(&Rate{
		Window: 2 * time.Second,
		Limit:  1,
	}, testPrefix)

	ginHandleFunc := GinRateLimiter(limiter, []string{"POST"})
	c, _ := gin.CreateTestContext(httptest.NewRecorder())
	c.Request, _ = http.NewRequest("GET", "/", nil)
	c.Request.Header.Add("X-Forwarded-For", testIP)
	assert.Equal(t, testIP, c.ClientIP())
	ginHandleFunc(c)
	assert.False(t, c.IsAborted())
	ginHandleFunc(c)
	assert.False(t, c.IsAborted())
}

func TestGinRateLimiterAllMethods(t *testing.T) {
	defer cleanRedisKey(fmt.Sprintf("%s:%s", testPrefix, testIP))
	limiter := NewRateLimiter(&Rate{
		Window: 2 * time.Second,
		Limit:  1,
	}, testPrefix)

	ginHandleFunc := GinRateLimiter(limiter, []string{})
	c, _ := gin.CreateTestContext(httptest.NewRecorder())
	c.Request, _ = http.NewRequest("GET", "/", nil)
	c.Request.Header.Add("X-Forwarded-For", testIP)
	assert.Equal(t, testIP, c.ClientIP())
	ginHandleFunc(c)
	assert.False(t, c.IsAborted())
	ginHandleFunc(c)
	assert.True(t, c.IsAborted())
}
