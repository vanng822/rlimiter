package rlimiter

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func TestAutoBan_NoIncrement(t *testing.T) {
	testIP := "192.168.1.1"
	autobanKey := fmt.Sprintf("%s%s:%s", globalPrefix, "autoban", testIP)
	autobannedKey := fmt.Sprintf("%s%s:%s", globalPrefix, "autobanned", testIP)
	defer cleanRedisKey(autobanKey)
	defer cleanRedisKey(autobannedKey)

	r := gin.New()
	r.Use(AutoBan())
	r.GET("/", func(c *gin.Context) {
		c.Status(http.StatusOK) // 200, not in statuses
	})

	req, _ := http.NewRequest("GET", "/", nil)
	req.RemoteAddr = testIP + ":1234"
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "", getRedis(autobanKey)) // Should not increment
}

func TestAutoBan_Increment(t *testing.T) {
	testIP := "192.168.1.2"
	autobanKey := fmt.Sprintf("%s%s:%s", globalPrefix, "autoban", testIP)
	autobannedKey := fmt.Sprintf("%s%s:%s", globalPrefix, "autobanned", testIP)
	defer cleanRedisKey(autobanKey)
	defer cleanRedisKey(autobannedKey)

	r := gin.New()
	r.Use(AutoBan())
	r.GET("/", func(c *gin.Context) {
		c.Status(http.StatusNotFound) // 404, in statuses
	})

	req, _ := http.NewRequest("GET", "/", nil)
	req.RemoteAddr = testIP + ":1234"
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
	assert.Equal(t, "1", getRedis(autobanKey)) // Should increment
}

func TestAutoBan_Ban(t *testing.T) {
	testIP := "192.168.1.3"
	autobanKey := fmt.Sprintf("%s%s:%s", globalPrefix, "autoban", testIP)
	autobannedKey := fmt.Sprintf("%s%s:%s", globalPrefix, "autobanned", testIP)
	defer cleanRedisKey(autobanKey)
	defer cleanRedisKey(autobannedKey)

	r := gin.New()
	r.Use(AutoBan(AutoBanWithRate(&Rate{Limit: 1, Window: time.Minute})))
	r.GET("/", func(c *gin.Context) {
		c.Status(http.StatusNotFound)
	})

	// First request
	req1, _ := http.NewRequest("GET", "/", nil)
	req1.RemoteAddr = testIP + ":1234"
	w1 := httptest.NewRecorder()
	r.ServeHTTP(w1, req1)
	assert.Equal(t, http.StatusNotFound, w1.Code)
	assert.Equal(t, "1", getRedis(autobanKey))

	// Second request: exceeds limit, bans
	req2, _ := http.NewRequest("GET", "/", nil)
	req2.RemoteAddr = testIP + ":1234"
	w2 := httptest.NewRecorder()
	r.ServeHTTP(w2, req2)
	assert.Equal(t, http.StatusNotFound, w2.Code) // Still processed
	assert.Equal(t, "1", getRedis(autobannedKey)) // Now banned

	// Third request: banned
	req3, _ := http.NewRequest("GET", "/", nil)
	req3.RemoteAddr = testIP + ":1234"
	w3 := httptest.NewRecorder()
	r.ServeHTTP(w3, req3)
	assert.Equal(t, http.StatusTooManyRequests, w3.Code) // Banned
}

func TestAutoBan_BannedRequest(t *testing.T) {
	testIP := "192.168.1.4"
	autobanKey := fmt.Sprintf("%s%s:%s", globalPrefix, "autoban", testIP)
	autobannedKey := fmt.Sprintf("%s%s:%s", globalPrefix, "autobanned", testIP)
	defer cleanRedisKey(autobanKey)
	defer cleanRedisKey(autobannedKey)

	// Manually ban the IP
	GetClient().Set(context.Background(), autobannedKey, "1", time.Minute)

	r := gin.New()
	r.Use(AutoBan())
	r.GET("/", func(c *gin.Context) {
		c.Status(http.StatusOK)
	})

	req, _ := http.NewRequest("GET", "/", nil)
	req.RemoteAddr = testIP + ":1234"
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusTooManyRequests, w.Code)
}

func TestAutoBan_Defaults(t *testing.T) {
	testIP := "192.168.1.5"
	autobanKey := fmt.Sprintf("%s%s:%s", globalPrefix, "autoban", testIP)
	autobannedKey := fmt.Sprintf("%s%s:%s", globalPrefix, "autobanned", testIP)
	defer cleanRedisKey(autobanKey)
	defer cleanRedisKey(autobannedKey)

	r := gin.New()
	r.Use(AutoBan()) // nil rate, default to 100/min, statuses to 404
	r.GET("/", func(c *gin.Context) {
		c.Status(http.StatusNotFound)
	})

	req, _ := http.NewRequest("GET", "/", nil)
	req.RemoteAddr = testIP + ":1234"
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
	assert.Equal(t, "1", getRedis(autobanKey))
}
