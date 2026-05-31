// 本文件实现基于 Redis 的 Store，适用于多实例 / 生产场景。
//
// 直接复用 cache.go 中已建立的 *redis.Client 连接，不重复创建。
package cache

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

// RedisStore 基于 Redis 的缓存实现。
type RedisStore struct {
	client *redis.Client
	prefix string // key 前缀，避免多应用共用 Redis 时冲突
}

// NewRedisStore 创建 Redis 缓存。prefix 会自动追加到所有 key 前面。
func NewRedisStore(client *redis.Client, prefix string) *RedisStore {
	return &RedisStore{client: client, prefix: prefix}
}

func (rs *RedisStore) prefixed(key string) string {
	if rs.prefix == "" {
		return key
	}
	return rs.prefix + key
}

func (rs *RedisStore) Get(ctx context.Context, key string) (string, error) {
	val, err := rs.client.Get(ctx, rs.prefixed(key)).Result()
	if errors.Is(err, redis.Nil) {
		return "", ErrCacheMiss
	}
	return val, err
}

func (rs *RedisStore) Set(ctx context.Context, key, value string, ttl time.Duration) error {
	return rs.client.Set(ctx, rs.prefixed(key), value, ttl).Err()
}

func (rs *RedisStore) Delete(ctx context.Context, keys ...string) error {
	if len(keys) == 0 {
		return nil
	}
	prefixed := make([]string, len(keys))
	for i, k := range keys {
		prefixed[i] = rs.prefixed(k)
	}
	return rs.client.Del(ctx, prefixed...).Err()
}

func (rs *RedisStore) Has(ctx context.Context, key string) (bool, error) {
	n, err := rs.client.Exists(ctx, rs.prefixed(key)).Result()
	if err != nil {
		return false, err
	}
	return n > 0, nil
}

func (rs *RedisStore) Remember(ctx context.Context, key string, ttl time.Duration, fn func() (string, error)) (string, error) {
	if v, err := rs.Get(ctx, key); err == nil {
		return v, nil
	}
	v, err := fn()
	if err != nil {
		return "", err
	}
	_ = rs.Set(ctx, key, v, ttl)
	return v, nil
}

// Flush 只清除本 Store 前缀下的键，而非整个 Redis 逻辑库。
//
// 安全考量：FlushDB 会清空当前 DB 的所有数据。由于 cache / session / 限流
// 默认共用同一个 Redis DB（DB 0），调用 FlushDB 会误删会话与限流状态，
// 属于高危的大范围操作。因此改为按 prefix 做 SCAN + 批量 DEL。
//
// 当 prefix 为空时，无法界定清除范围（等同于全库），为防止误清，直接拒绝执行。
func (rs *RedisStore) Flush(ctx context.Context) error {
	if rs.prefix == "" {
		return errors.New("cache: refuse to flush redis without a key prefix (would wipe the whole DB); set cache.prefix")
	}

	match := rs.prefix + "*"
	var cursor uint64
	for {
		keys, next, err := rs.client.Scan(ctx, cursor, match, 256).Result()
		if err != nil {
			return fmt.Errorf("cache: scan %q: %w", match, err)
		}
		if len(keys) > 0 {
			if err := rs.client.Del(ctx, keys...).Err(); err != nil {
				return fmt.Errorf("cache: del during flush: %w", err)
			}
		}
		cursor = next
		if cursor == 0 {
			break
		}
	}
	return nil
}
