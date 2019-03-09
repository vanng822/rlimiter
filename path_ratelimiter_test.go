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
	testPath = "/somepath"
)

func TestGinPathRateLimiter(t *testing.T) {
	key := fmt.Sprintf("%s:%s", testPrefix, testPath)
	defer cleanRedisKey(key)
	limiter := NewPathRateLimiter(&Rate{
		Window: 2 * time.Second,
		Limit:  1,
	}, testPrefix)
	ginHandleFunc := GinRateLimiter(limiter, []string{"GET"})
	c, _ := gin.CreateTestContext(httptest.NewRecorder())
	c.Request, _ = http.NewRequest("GET", testPath, nil)
	c.Request.Header.Add("X-Forwarded-For", testIP)
	assert.Equal(t, "", getRedis(key))
	ginHandleFunc(c)
	assert.False(t, c.IsAborted())
	assert.Equal(t, "1", getRedis(key))
	ginHandleFunc(c)
	assert.True(t, c.IsAborted())
}
