package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"

	"thinkgin/app"
)

// setMiddlewareConfig 注入 middleware.config 下的子配置，
// 让 getRateLimitConfig / getCORSConfig 读到测试期望的值。
func setMiddlewareConfig(t *testing.T, section string, cfg map[string]interface{}) {
	t.Helper()
	app.Config = &app.GlobalConfig{}
	app.Config.Middleware.Config = map[string]interface{}{
		section: cfg,
	}
}

// newRouterWithRateLimit 构造一个只挂载 RateLimit + 200 OK handler 的路由器，
// 用于逐次发请求测限流行为。
func newRouterWithRateLimit() *gin.Engine {
	r := gin.New()
	r.Use(RateLimit())
	r.GET("/ping", func(c *gin.Context) { c.String(http.StatusOK, "pong") })
	return r
}

func TestRateLimit_NoopWhenUnconfigured(t *testing.T) {
	app.Config = &app.GlobalConfig{} // middleware.config 为 nil
	r := newRouterWithRateLimit()

	for i := 0; i < 50; i++ {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest(http.MethodGet, "/ping", nil)
		r.ServeHTTP(w, req)
		if w.Code != http.StatusOK {
			t.Fatalf("request #%d: status=%d, want 200 (rate limiter should be no-op)", i, w.Code)
		}
	}
}

func TestRateLimit_BlocksWhenExceeded(t *testing.T) {
	setMiddlewareConfig(t, "rate_limit", map[string]interface{}{
		"requests_per_minute": 3,
	})
	r := newRouterWithRateLimit()

	// 前 3 次应通过，第 4 次必被拦截（同一 IP）。
	for i := 0; i < 3; i++ {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest(http.MethodGet, "/ping", nil)
		req.RemoteAddr = "10.0.0.1:1234"
		r.ServeHTTP(w, req)
		if w.Code != http.StatusOK {
			t.Fatalf("request #%d: status=%d, want 200", i, w.Code)
		}
	}

	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "/ping", nil)
	req.RemoteAddr = "10.0.0.1:1234"
	r.ServeHTTP(w, req)
	if w.Code != http.StatusTooManyRequests {
		t.Fatalf("4th request status=%d, want 429", w.Code)
	}
}

func TestRateLimit_PerIPIndependence(t *testing.T) {
	setMiddlewareConfig(t, "rate_limit", map[string]interface{}{
		"requests_per_minute": 2,
	})
	r := newRouterWithRateLimit()

	// 耗尽 IP A 的配额后，IP B 应仍可通过——验证两个 IP 各自独立计数。
	drain := func(ip string) {
		for i := 0; i < 2; i++ {
			w := httptest.NewRecorder()
			req, _ := http.NewRequest(http.MethodGet, "/ping", nil)
			req.RemoteAddr = ip + ":1000"
			r.ServeHTTP(w, req)
		}
	}
	drain("10.0.0.1")

	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "/ping", nil)
	req.RemoteAddr = "10.0.0.2:1000"
	r.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("IP B first request status=%d, want 200 (should have its own bucket)", w.Code)
	}
}

func TestEvictStale_RemovesExpiredBuckets(t *testing.T) {
	now := time.Now()
	store := &limiterStore{
		buckets: map[string]*tokenBucket{
			"10.0.0.1": {capacity: 60, tokens: 60, rate: 1, last: now.Add(-15 * time.Minute)}, // 过期
			"10.0.0.2": {capacity: 60, tokens: 60, rate: 1, last: now.Add(-1 * time.Minute)},  // 活跃
		},
	}
	store.evictStale(now)

	if _, ok := store.buckets["10.0.0.1"]; ok {
		t.Error("stale bucket 10.0.0.1 should have been evicted")
	}
	if _, ok := store.buckets["10.0.0.2"]; !ok {
		t.Error("active bucket 10.0.0.2 should be retained")
	}
}

func TestTokenBucket_RefillOverTime(t *testing.T) {
	// 白盒测令牌桶刷新：initial=0 token / rate=60/s，1 秒后应回到满桶。
	b := &tokenBucket{
		capacity: 60,
		tokens:   0,
		rate:     60,
		last:     time.Unix(0, 0),
	}
	if b.allow(time.Unix(0, 0)) {
		t.Fatal("empty bucket should reject")
	}
	// 经过 1 秒，应恢复到 capacity
	if !b.allow(time.Unix(1, 0)) {
		t.Fatal("bucket should allow after refill")
	}
	// 连续取用应在 capacity 次以内都成功
	passed := 1
	for i := 0; i < 200; i++ {
		if b.allow(time.Unix(1, 0)) {
			passed++
		}
	}
	if passed > 60 {
		t.Errorf("bucket allowed %d requests in same instant, should cap at 60", passed)
	}
}
