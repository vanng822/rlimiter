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
	ginHandleFunc := GinRateLimiter(limiter, []string{"GET"})
	c, _ := gin.CreateTestContext(httptest.NewRecorder())
	c.Request, _ = http.NewRequest("GET", "/blop", nil)
	c.Request.Header.Add("X-Forwarded-For", testIP2)
	assert.Equal(t, "", getRedis(key))
	ginHandleFunc(c)
	assert.False(t, c.IsAborted())
	assert.Equal(t, "1", getRedis(key))
	ginHandleFunc(c)
	assert.True(t, c.IsAborted())
}
