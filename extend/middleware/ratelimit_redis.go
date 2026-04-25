// 本文件提供基于 Redis 的分布式令牌桶限流中间件。
//
// 适用场景：
//   - 多实例部署时需要全局统一限流配额。
//   - 通过 Redis Lua 脚本保证原子性，避免竞态条件。
//
// 配置路径：middleware.config.redis_rate_limit
//   - requests_per_minute: int  每分钟请求上限
//   - redis_store: string       使用的 Redis store 名称（对应 cache.yaml 中的 store）
//
// 启用方式：在 middleware.global 列表中加入 "redis_rate_limit"。
package middleware

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"thinkgin/app"
	"thinkgin/app/cache"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
)

// luaTokenBucket 实现原子性令牌桶算法的 Lua 脚本。
// KEYS[1]: 桶 key
// ARGV[1]: capacity (最大令牌数)
// ARGV[2]: rate (每秒新增令牌数)
// ARGV[3]: now (当前 Unix 秒，浮点)
// 返回: 1=允许, 0=拒绝
var luaTokenBucket = redis.NewScript(`
local key = KEYS[1]
local capacity = tonumber(ARGV[1])
local rate = tonumber(ARGV[2])
local now = tonumber(ARGV[3])

local data = redis.call('HMGET', key, 'tokens', 'last')
local tokens = tonumber(data[1])
local last = tonumber(data[2])

if tokens == nil then
    tokens = capacity
    last = now
end

local elapsed = math.max(0, now - last)
tokens = math.min(capacity, tokens + elapsed * rate)

local allowed = 0
if tokens >= 1 then
    tokens = tokens - 1
    allowed = 1
end

redis.call('HMSET', key, 'tokens', tostring(tokens), 'last', tostring(now))
redis.call('EXPIRE', key, 600)

return allowed
`)

type redisRateLimitConfig struct {
	RequestsPerMinute int
	RedisStore        string
}

// RedisRateLimit 返回基于 Redis 的分布式限流中间件。
func RedisRateLimit() gin.HandlerFunc {
	cfg := getRedisRateLimitConfig()
	if cfg.RequestsPerMinute <= 0 {
		return func(c *gin.Context) { c.Next() }
	}

	// 获取 Redis 客户端
	var rdb *redis.Client
	if cfg.RedisStore != "" {
		rdb, _ = cache.Get(cfg.RedisStore)
	}
	if rdb == nil {
		rdb = cache.Default()
	}
	if rdb == nil {
		app.GetLogger().Warnf("[redis_rate_limit] Redis 不可用，降级为 no-op")
		return func(c *gin.Context) { c.Next() }
	}

	capacity := cfg.RequestsPerMinute
	rate := float64(capacity) / 60.0

	return func(c *gin.Context) {
		ip := c.ClientIP()
		key := fmt.Sprintf("thinkgin:ratelimit:%s", ip)
		now := float64(time.Now().UnixMilli()) / 1000.0

		ctx := context.Background()
		result, err := luaTokenBucket.Run(ctx, rdb, []string{key},
			capacity, rate, now,
		).Int()

		if err != nil {
			// Redis 出错时放行，避免限流系统故障导致全站不可用
			app.GetLogger().Warnf("[redis_rate_limit] lua eval error: %v", err)
			c.Next()
			return
		}

		if result == 0 {
			c.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{
				"code":    http.StatusTooManyRequests,
				"message": "too many requests",
			})
			return
		}

		c.Next()
	}
}

func getRedisRateLimitConfig() redisRateLimitConfig {
	cfg := app.GetConfig()
	if cfg == nil || cfg.Middleware.Config == nil {
		return redisRateLimitConfig{}
	}

	raw, ok := cfg.Middleware.Config["redis_rate_limit"]
	if !ok {
		return redisRateLimitConfig{}
	}

	m, ok := asStringInterfaceMap(raw)
	if !ok {
		return redisRateLimitConfig{}
	}

	out := redisRateLimitConfig{}
	if v, ok := m["requests_per_minute"]; ok {
		if n, ok := v.(int); ok {
			out.RequestsPerMinute = n
		}
	}
	if v, ok := m["redis_store"]; ok {
		if s, ok := v.(string); ok {
			out.RedisStore = s
		}
	}
	return out
}
