// 本文件实现基于 IP 的进程内令牌桶限流。
//
// 适用场景：
//   - 单实例部署（令牌桶内存在进程堆中，不跨实例共享）。
//   - 对恶意爬虫 / 简单防刷场景做最后一层兜底。
//
// 多实例部署时，令牌桶在各实例独立计数，整体配额会被放大 N 倍。
// 如需全局限流，请改造为 Redis + Lua 或接入专门的限流网关。
//
// 内存保护：后台 goroutine 每 bucketGCInterval 清理一次超过 bucketTTL
// 未被访问的桶，避免海量 IP 导致内存无限膨胀。
package middleware

import (
	"net/http"
	"sync"
	"time"

	"thinkgin/app"

	"github.com/gin-gonic/gin"
)

const (
	// bucketTTL 超过此时间未被访问的桶将被回收。
	bucketTTL = 10 * time.Minute
	// bucketGCInterval 后台清理周期。
	bucketGCInterval = 1 * time.Minute
)

// tokenBucket 单 IP 的令牌桶状态。
type tokenBucket struct {
	capacity int       // 桶容量（同时也是每分钟上限）
	tokens   float64   // 当前剩余令牌
	rate     float64   // 每秒新增令牌数
	last     time.Time // 上次填充令牌的时间
}

func (b *tokenBucket) allow(now time.Time) bool {
	elapsed := now.Sub(b.last).Seconds()
	if elapsed > 0 {
		b.tokens += elapsed * b.rate
		if b.tokens > float64(b.capacity) {
			b.tokens = float64(b.capacity)
		}
		b.last = now
	}
	if b.tokens >= 1 {
		b.tokens -= 1
		return true
	}
	return false
}

type limiterStore struct {
	mu      sync.Mutex
	buckets map[string]*tokenBucket
	cfg     rateLimitConfig
}

// evictStale 删除所有超过 bucketTTL 未被访问的桶。
func (s *limiterStore) evictStale(now time.Time) {
	s.mu.Lock()
	defer s.mu.Unlock()
	for ip, b := range s.buckets {
		if now.Sub(b.last) > bucketTTL {
			delete(s.buckets, ip)
		}
	}
}

type rateLimitConfig struct {
	RequestsPerMinute int
}

func RateLimit() gin.HandlerFunc {
	cfg := getRateLimitConfig()
	if cfg.RequestsPerMinute <= 0 {
		return func(c *gin.Context) { c.Next() }
	}

	store := &limiterStore{
		buckets: make(map[string]*tokenBucket),
		cfg:     cfg,
	}

	// 后台 goroutine 定期清理不活跃的 IP 桶，防止内存泄漏。
	go func() {
		ticker := time.NewTicker(bucketGCInterval)
		defer ticker.Stop()
		for t := range ticker.C {
			store.evictStale(t)
		}
	}()

	return func(c *gin.Context) {
		ip := c.ClientIP()
		now := time.Now()

		store.mu.Lock()
		b := store.buckets[ip]
		if b == nil {
			b = &tokenBucket{
				capacity: cfg.RequestsPerMinute,
				tokens:   float64(cfg.RequestsPerMinute),
				rate:     float64(cfg.RequestsPerMinute) / 60.0,
				last:     now,
			}
			store.buckets[ip] = b
		}
		allowed := b.allow(now)
		store.mu.Unlock()

		if !allowed {
			c.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{
				"code":    http.StatusTooManyRequests,
				"message": "too many requests",
			})
			return
		}

		c.Next()
	}
}

func getRateLimitConfig() rateLimitConfig {
	out := rateLimitConfig{RequestsPerMinute: 0}

	cfg := app.GetConfig()
	if cfg == nil || cfg.Middleware.Config == nil {
		return out
	}

	root := cfg.Middleware.Config
	v, ok := root["rate_limit"]
	if !ok {
		return out
	}

	m, ok := asStringInterfaceMap(v)
	if !ok {
		return out
	}

	if rpm, ok := m["requests_per_minute"]; ok {
		if i, ok := toInt(rpm); ok {
			out.RequestsPerMinute = i
		}
	}

	return out
}
