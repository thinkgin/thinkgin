package middleware

import (
	"net/http"
	"sync"
	"time"

	"thinkgin/app"

	"github.com/gin-gonic/gin"
)

type tokenBucket struct {
	capacity int
	tokens   float64
	rate     float64
	last     time.Time
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
