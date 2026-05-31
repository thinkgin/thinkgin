package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
)

func newCBRouter(cfg CircuitBreakerConfig, handler gin.HandlerFunc) *gin.Engine {
	r := gin.New()
	r.Use(CircuitBreakerWithConfig(cfg))
	r.GET("/test", handler)
	return r
}

func TestCircuitBreaker_ClosedPassesThrough(t *testing.T) {
	cfg := DefaultCircuitBreakerConfig()
	r := newCBRouter(cfg, func(c *gin.Context) {
		c.String(200, "ok")
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/test", nil)
	r.ServeHTTP(w, req)

	if w.Code != 200 {
		t.Errorf("status = %d, want 200", w.Code)
	}
}

func TestCircuitBreaker_OpensOnHighErrorRate(t *testing.T) {
	cfg := CircuitBreakerConfig{
		WindowSize:            10,
		ErrorThresholdPercent: 50,
		CooldownDuration:      5 * time.Second,
		HalfOpenMaxRequests:   2,
		IsServerError:         func(code int) bool { return code >= 500 },
	}

	callCount := 0
	r := newCBRouter(cfg, func(c *gin.Context) {
		callCount++
		c.String(500, "error") // 所有请求都返回 500
	})

	// 先打满窗口
	for i := 0; i < cfg.WindowSize; i++ {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/test", nil)
		r.ServeHTTP(w, req)
	}

	// 窗口已满且 100% 错误率 → 下一个请求应被熔断
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/test", nil)
	r.ServeHTTP(w, req)

	if w.Code != http.StatusServiceUnavailable {
		t.Errorf("status = %d, want 503 (circuit breaker open)", w.Code)
	}
}

func TestCircuitBreaker_HalfOpenRecovery(t *testing.T) {
	cfg := CircuitBreakerConfig{
		WindowSize:            5,
		ErrorThresholdPercent: 50,
		CooldownDuration:      50 * time.Millisecond, // 很短的冷却时间
		HalfOpenMaxRequests:   2,
		IsServerError:         func(code int) bool { return code >= 500 },
	}

	errorMode := true
	r := newCBRouter(cfg, func(c *gin.Context) {
		if errorMode {
			c.String(500, "error")
		} else {
			c.String(200, "ok")
		}
	})

	// 打满窗口触发熔断
	for i := 0; i < cfg.WindowSize; i++ {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/test", nil)
		r.ServeHTTP(w, req)
	}

	// 确认已熔断
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/test", nil)
	r.ServeHTTP(w, req)
	if w.Code != 503 {
		t.Fatalf("should be open, got %d", w.Code)
	}

	// 等待冷却
	time.Sleep(100 * time.Millisecond)

	// 切换为正常模式
	errorMode = false

	// HalfOpen 试探请求应放行并成功
	for i := 0; i < cfg.HalfOpenMaxRequests; i++ {
		w = httptest.NewRecorder()
		req, _ = http.NewRequest("GET", "/test", nil)
		r.ServeHTTP(w, req)
		if w.Code != 200 {
			t.Errorf("half-open request %d: status = %d, want 200", i, w.Code)
		}
	}

	// 熔断器应该恢复到 Closed，后续请求正常
	w = httptest.NewRecorder()
	req, _ = http.NewRequest("GET", "/test", nil)
	r.ServeHTTP(w, req)
	if w.Code != 200 {
		t.Errorf("after recovery: status = %d, want 200", w.Code)
	}
}

func TestCircuitBreaker_LowTrafficTripsViaMinRequests(t *testing.T) {
	// 低流量场景：窗口很大但 MinRequests 很小。
	// 只要累计样本达到 MinRequests 且错误率超阈值，即应熔断，无需填满窗口。
	cfg := CircuitBreakerConfig{
		WindowSize:            100,
		MinRequests:           5,
		ErrorThresholdPercent: 50,
		CooldownDuration:      5 * time.Second,
		HalfOpenMaxRequests:   2,
		IsServerError:         func(code int) bool { return code >= 500 },
	}

	r := newCBRouter(cfg, func(c *gin.Context) {
		c.String(500, "error")
	})

	// 仅发 5 个请求（远小于 WindowSize=100），全部失败。
	for i := 0; i < cfg.MinRequests; i++ {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/test", nil)
		r.ServeHTTP(w, req)
	}

	// 第 6 个请求应被熔断，证明无需填满窗口即可触发。
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/test", nil)
	r.ServeHTTP(w, req)
	if w.Code != http.StatusServiceUnavailable {
		t.Errorf("low-traffic circuit should be open after MinRequests, got %d", w.Code)
	}
}

func TestCircuitBreaker_HalfOpenAdmissionCap(t *testing.T) {
	// 验证 HalfOpen 在准入时即占用配额：单元级别直接驱动 allow()，
	// 在不调用 record() 的情况下，放行次数不得超过 HalfOpenMaxRequests。
	cfg := CircuitBreakerConfig{
		WindowSize:            4,
		MinRequests:           4,
		ErrorThresholdPercent: 50,
		CooldownDuration:      1 * time.Millisecond,
		HalfOpenMaxRequests:   3,
		IsServerError:         func(code int) bool { return code >= 500 },
	}
	cb := newCircuitBreaker(cfg)

	// 打满窗口并全部失败 → Open。
	for i := 0; i < cfg.WindowSize; i++ {
		cb.record(true)
	}
	if cb.state != cbOpen {
		t.Fatalf("breaker should be open, state=%d", cb.state)
	}

	// 冷却后，连续 allow()（不 record），放行次数应恰好等于 HalfOpenMaxRequests。
	time.Sleep(5 * time.Millisecond)
	allowed := 0
	for i := 0; i < 100; i++ {
		if cb.allow() {
			allowed++
		}
	}
	if allowed != cfg.HalfOpenMaxRequests {
		t.Errorf("half-open admitted %d requests, want exactly %d (no over-admission)", allowed, cfg.HalfOpenMaxRequests)
	}
}

func TestCircuitBreaker_DefaultConfig(t *testing.T) {
	cfg := DefaultCircuitBreakerConfig()
	if cfg.WindowSize != 100 {
		t.Errorf("WindowSize = %d, want 100", cfg.WindowSize)
	}
	if cfg.MinRequests != 10 {
		t.Errorf("MinRequests = %d, want 10", cfg.MinRequests)
	}
	if cfg.ErrorThresholdPercent != 50 {
		t.Errorf("ErrorThresholdPercent = %d, want 50", cfg.ErrorThresholdPercent)
	}
	if !cfg.IsServerError(500) {
		t.Error("IsServerError(500) should be true")
	}
	if cfg.IsServerError(499) {
		t.Error("IsServerError(499) should be false")
	}
}

func TestCircuitBreaker_DefaultMiddleware(t *testing.T) {
	r := gin.New()
	r.Use(CircuitBreaker())
	r.GET("/test", func(c *gin.Context) { c.String(200, "ok") })

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/test", nil)
	r.ServeHTTP(w, req)
	if w.Code != 200 {
		t.Errorf("status = %d, want 200", w.Code)
	}
}
