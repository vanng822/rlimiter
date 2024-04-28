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
	testIP2 = "65.121.1.233"
)

func TestGinIpRateLimiter(t *testing.T) {
	key := fmt.Sprintf("%s:%s", testPrefix, testIP2)
	defer cleanRedisKey(key)
	limiter := NewIPRateLimiter(&Rate{
		Window: 2 * time.Second,
		Limit:  1,
	}, testPrefix)
	ginHandleFunc := GinRateLimit(limiter, []string{"GET"})
	c, _ := gin.CreateTestContext(httptest.NewRecorder())
	c.Request, _ = http.NewRequest("GET", "/blop", nil)
	c.Request.RemoteAddr = testIP2 + ":1234"
	assert.Equal(t, "", getRedis(key))
	ginHandleFunc(c)
	assert.False(t, c.IsAborted())
	assert.Equal(t, "1", getRedis(key))
	ginHandleFunc(c)
	assert.True(t, c.IsAborted())
	assert.Equal(t, "2", getRedis(key))
}

func TestGinRateLimiter(t *testing.T) {
	defer cleanRedisKey(fmt.Sprintf("%s:%s", testPrefix, testIP))
	limiter := NewIPRateLimiter(&Rate{
		Window: 2 * time.Second,
		Limit:  1,
	}, testPrefix)
	ginHandleFunc := GinRateLimit(limiter, []string{"GET"})
	c, _ := gin.CreateTestContext(httptest.NewRecorder())
	c.Request, _ = http.NewRequest("GET", "/", nil)
	c.Request.RemoteAddr = testIP + ":1234"
	ginHandleFunc(c)
	assert.False(t, c.IsAborted())
	ginHandleFunc(c)
	assert.True(t, c.IsAborted())
}

func TestGinIPRateLimiterNotSameMethod(t *testing.T) {
	defer cleanRedisKey(fmt.Sprintf("%s:%s", testPrefix, testIP))
	limiter := NewIPRateLimiter(&Rate{
		Window: 2 * time.Second,
		Limit:  1,
	}, testPrefix)

	ginHandleFunc := GinRateLimit(limiter, []string{"POST"})
	c, _ := gin.CreateTestContext(httptest.NewRecorder())

	c.Request, _ = http.NewRequest("GET", "/", nil)
	c.Request.RemoteAddr = testIP + ":1234"

	assert.Equal(t, testIP, c.ClientIP())
	ginHandleFunc(c)
	assert.False(t, c.IsAborted())
	ginHandleFunc(c)
	assert.False(t, c.IsAborted())
}

func TestGinIPRateLimiterAllMethods(t *testing.T) {
	defer cleanRedisKey(fmt.Sprintf("%s:%s", testPrefix, testIP))
	limiter := NewIPRateLimiter(&Rate{
		Window: 2 * time.Second,
		Limit:  1,
	}, testPrefix)

	ginHandleFunc := GinRateLimit(limiter, []string{})
	c, _ := gin.CreateTestContext(httptest.NewRecorder())

	c.Request, _ = http.NewRequest("GET", "/", nil)
	c.Request.RemoteAddr = testIP + ":1234"

	assert.Equal(t, testIP, c.ClientIP())
	ginHandleFunc(c)
	assert.False(t, c.IsAborted())
	ginHandleFunc(c)
	assert.True(t, c.IsAborted())
}
