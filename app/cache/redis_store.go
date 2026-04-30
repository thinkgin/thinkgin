// 本文件实现基于 Redis 的 Store，适用于多实例 / 生产场景。
//
// 直接复用 cache.go 中已建立的 *redis.Client 连接，不重复创建。
package cache

import (
	"context"
	"errors"
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

func (rs *RedisStore) Flush(ctx context.Context) error {
	return rs.client.FlushDB(ctx).Err()
}
