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
	// MinRequests 触发熔断所需的最小样本数。
	// 解决"低流量接口永不熔断"的问题：窗口未填满时，只要样本数达到 MinRequests
	// 且错误率超阈值即可熔断。为 0 时回退为 WindowSize（保持旧行为，向后兼容）。
	// 实际取值会被 cap 到 WindowSize，避免设置过大导致永不触发。
	MinRequests int
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
		MinRequests:           10, // 低流量下也能熔断：累计 10 个样本即可判定
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
	// 归一化 MinRequests：0 时回退为 WindowSize（旧行为），并 cap 到 WindowSize。
	if cfg.MinRequests <= 0 || cfg.MinRequests > cfg.WindowSize {
		cfg.MinRequests = cfg.WindowSize
	}
	return &circuitBreaker{
		cfg:    cfg,
		state:  cbClosed,
		window: make([]bool, cfg.WindowSize),
	}
}

// allow 判断是否放行请求。
//
// 重要：HalfOpen 分支在准入时即占用一个试探配额（halfOpenTotal++），
// 而非等到 record() 才计数。否则高并发下大量请求会在 record 之前同时
// 通过 `halfOpenTotal < max` 判断，导致远超 HalfOpenMaxRequests 的请求涌入。
func (cb *circuitBreaker) allow() bool {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	switch cb.state {
	case cbClosed:
		return true
	case cbOpen:
		if time.Since(cb.openedAt) >= cb.cfg.CooldownDuration {
			cb.toHalfOpen()
			cb.halfOpenTotal++ // 转入 HalfOpen 后立即占用第一个试探名额
			return true
		}
		return false
	case cbHalfOpen:
		if cb.halfOpenTotal < cb.cfg.HalfOpenMaxRequests {
			cb.halfOpenTotal++ // 准入即占用名额，杜绝并发超发
			return true
		}
		return false
	}
	return true
}

// record 记录请求结果。
//
// 注意：HalfOpen 的 halfOpenTotal 已在 allow() 准入时自增，此处不再重复计数，
// 只统计成功/失败并据此切换状态。
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
		// 样本数达到 MinRequests（而非必须填满窗口）且错误率超阈值即熔断，
		// 解决低流量接口永远凑不满窗口、从而永不熔断的问题。
		if cb.count >= cb.cfg.MinRequests && cb.errorRate() >= cb.cfg.ErrorThresholdPercent {
			cb.toOpen()
		}
	case cbHalfOpen:
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

// ──────────── 按路由粒度的熔断器注册表 ────────────

// cbRegistry 为每条路由维护独立的 circuitBreaker，
// 避免一个慢接口（如文件上传）的错误率影响其他接口。
type cbRegistry struct {
	cbs sync.Map // map[string]*circuitBreaker
	cfg CircuitBreakerConfig
}

// getOrCreate 按路由 key 获取或懒创建对应的 circuitBreaker。
func (r *cbRegistry) getOrCreate(key string) *circuitBreaker {
	if v, ok := r.cbs.Load(key); ok {
		return v.(*circuitBreaker)
	}
	cb := newCircuitBreaker(r.cfg)
	actual, _ := r.cbs.LoadOrStore(key, cb)
	return actual.(*circuitBreaker)
}

// routeKey 提取路由粒度的标识：优先用 Gin 注册的路由模板（如 /api/users/:id），
// 未匹配到路由时回退为 METHOD+Path，避免高基数问题。
func routeKey(c *gin.Context) string {
	if tpl := c.FullPath(); tpl != "" {
		return c.Request.Method + " " + tpl
	}
	return c.Request.Method + " " + c.Request.URL.Path
}

// ──────────── Gin 中间件 ────────────

// CircuitBreaker 返回使用默认配置的按路由粒度熔断器中间件。
func CircuitBreaker() gin.HandlerFunc {
	return CircuitBreakerWithConfig(DefaultCircuitBreakerConfig())
}

// CircuitBreakerWithConfig 返回使用自定义配置的按路由粒度熔断器中间件。
// 每条路由拥有独立的滑动窗口和状态机，互不影响。
func CircuitBreakerWithConfig(cfg CircuitBreakerConfig) gin.HandlerFunc {
	reg := &cbRegistry{cfg: cfg}

	return func(c *gin.Context) {
		key := routeKey(c)
		cb := reg.getOrCreate(key)

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
