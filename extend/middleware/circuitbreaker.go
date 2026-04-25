// 本文件提供基于滑动窗口的熔断器中间件（Circuit Breaker）。
//
// 状态机：
//   Closed  → 正常转发请求；错误率超过阈值时切换到 Open
//   Open    → 直接拒绝请求（返回 503）；超过冷却时间后切换到 HalfOpen
//   HalfOpen → 放行有限的试探请求；成功则 → Closed，失败则 → Open
//
// 使用方式：
//   在 config/middleware.yaml 中添加 "circuit_breaker" 到 global 列表，
//   或手动调用 middleware.CircuitBreaker() 注册。
package middleware

import (
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
)

// ──────────── 配置 ────────────

// CircuitBreakerConfig 熔断器配置。
type CircuitBreakerConfig struct {
	// WindowSize 滑动窗口大小（最近多少个请求用于计算错误率）。
	WindowSize int
	// ErrorThresholdPercent 错误率阈值（0-100）。超过则熔断。
	ErrorThresholdPercent int
	// CooldownDuration Open 状态的冷却时间，之后进入 HalfOpen。
	CooldownDuration time.Duration
	// HalfOpenMaxRequests HalfOpen 状态允许放行的最大试探请求数。
	HalfOpenMaxRequests int
	// IsServerError 判断 HTTP 状态码是否算作错误，默认 >= 500。
	IsServerError func(statusCode int) bool
}

// DefaultCircuitBreakerConfig 返回默认配置。
func DefaultCircuitBreakerConfig() CircuitBreakerConfig {
	return CircuitBreakerConfig{
		WindowSize:            100,
		ErrorThresholdPercent: 50,
		CooldownDuration:      10 * time.Second,
		HalfOpenMaxRequests:   5,
		IsServerError: func(code int) bool {
			return code >= http.StatusInternalServerError
		},
	}
}

// ──────────── 状态机 ────────────

type cbState int

const (
	cbClosed   cbState = iota
	cbOpen
	cbHalfOpen
)

// circuitBreaker 单个熔断器实例。
type circuitBreaker struct {
	mu sync.Mutex

	cfg   CircuitBreakerConfig
	state cbState

	// 滑动窗口：环形缓冲区，true = error
	window []bool
	pos    int
	count  int // 已填充数

	// Open 状态下的进入时间
	openedAt time.Time

	// HalfOpen 计数
	halfOpenSuccesses int
	halfOpenFailures  int
	halfOpenTotal     int
}

func newCircuitBreaker(cfg CircuitBreakerConfig) *circuitBreaker {
	return &circuitBreaker{
		cfg:    cfg,
		state:  cbClosed,
		window: make([]bool, cfg.WindowSize),
	}
}

// allow 判断是否放行请求。
func (cb *circuitBreaker) allow() bool {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	switch cb.state {
	case cbClosed:
		return true
	case cbOpen:
		if time.Since(cb.openedAt) >= cb.cfg.CooldownDuration {
			cb.toHalfOpen()
			return true
		}
		return false
	case cbHalfOpen:
		return cb.halfOpenTotal < cb.cfg.HalfOpenMaxRequests
	}
	return true
}

// record 记录请求结果。
func (cb *circuitBreaker) record(isError bool) {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	switch cb.state {
	case cbClosed:
		cb.window[cb.pos] = isError
		cb.pos = (cb.pos + 1) % cb.cfg.WindowSize
		if cb.count < cb.cfg.WindowSize {
			cb.count++
		}
		if cb.count >= cb.cfg.WindowSize && cb.errorRate() >= cb.cfg.ErrorThresholdPercent {
			cb.toOpen()
		}
	case cbHalfOpen:
		cb.halfOpenTotal++
		if isError {
			cb.halfOpenFailures++
			cb.toOpen()
		} else {
			cb.halfOpenSuccesses++
			if cb.halfOpenSuccesses >= cb.cfg.HalfOpenMaxRequests {
				cb.toClosed()
			}
		}
	}
}

func (cb *circuitBreaker) errorRate() int {
	if cb.count == 0 {
		return 0
	}
	errors := 0
	for i := 0; i < cb.count; i++ {
		if cb.window[i] {
			errors++
		}
	}
	return errors * 100 / cb.count
}

func (cb *circuitBreaker) toOpen() {
	cb.state = cbOpen
	cb.openedAt = time.Now()
}

func (cb *circuitBreaker) toHalfOpen() {
	cb.state = cbHalfOpen
	cb.halfOpenSuccesses = 0
	cb.halfOpenFailures = 0
	cb.halfOpenTotal = 0
}

func (cb *circuitBreaker) toClosed() {
	cb.state = cbClosed
	cb.count = 0
	cb.pos = 0
	for i := range cb.window {
		cb.window[i] = false
	}
}

// ──────────── Gin 中间件 ────────────

// CircuitBreaker 返回使用默认配置的熔断器中间件。
func CircuitBreaker() gin.HandlerFunc {
	return CircuitBreakerWithConfig(DefaultCircuitBreakerConfig())
}

// CircuitBreakerWithConfig 返回使用自定义配置的熔断器中间件。
func CircuitBreakerWithConfig(cfg CircuitBreakerConfig) gin.HandlerFunc {
	cb := newCircuitBreaker(cfg)

	return func(c *gin.Context) {
		if !cb.allow() {
			c.AbortWithStatusJSON(http.StatusServiceUnavailable, gin.H{
				"code":    http.StatusServiceUnavailable,
				"message": "service temporarily unavailable (circuit breaker open)",
			})
			return
		}

		c.Next()

		isErr := cfg.IsServerError(c.Writer.Status())
		cb.record(isErr)
	}
}
