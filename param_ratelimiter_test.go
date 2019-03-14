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
	testParamPath  = "/someparam/testing"
	testParamValue = "testing"
)

func TestGinParamRateLimiter(t *testing.T) {
	key := fmt.Sprintf("%s:%s", testPrefix, testParamValue)
	defer cleanRedisKey(key)
	limiter := NewParamRateLimiter(&Rate{
		Window: 2 * time.Second,
		Limit:  1,
	}, testPrefix, "someparam")
	ginHandleFunc := GinRateLimit(limiter, []string{"GET"})
	c, _ := gin.CreateTestContext(httptest.NewRecorder())
	c.Request, _ = http.NewRequest("GET", testParamPath, nil)
	params := make([]gin.Param, 0)
	params = append(params, gin.Param{Key: "someparam", Value: testParamValue})
	c.Params = params
	assert.Equal(t, "", getRedis(key))
	ginHandleFunc(c)
	assert.False(t, c.IsAborted())
	assert.Equal(t, "1", getRedis(key))
	ginHandleFunc(c)
	assert.True(t, c.IsAborted())
	assert.Equal(t, "2", getRedis(key))
}
