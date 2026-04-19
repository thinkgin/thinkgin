// 本文件提供基于 Redis 的 Session 存储，复用 database.yaml 中声明的 redis 连接。
//
// 数据编码：JSON。优点是可读、可跨语言调试；缺点是 float64 会吞掉 int64 精度，
// 业务如需强类型可序列化成结构体再写 Session。
package session

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"

	"thinkgin/app"
)

// redisKeyPrefix 所有 Session key 的统一前缀，便于运维批量扫描。
const redisKeyPrefix = "thinkgin:session:"

type redisStore struct {
	client *redis.Client
	ttl    time.Duration
}

// newRedisStore 按 connection 名称解析 database.yaml 里的 redis 连接并建立 client。
// 这里不共用 app/cache 的实例，原因是 cache 和 session 的生命周期可能不同，
// 也避免引入循环：cache.Init 需要 Redis 启动成功，session 同样需要。
func newRedisStore(connection string, ttl time.Duration) (*redisStore, error) {
	if connection == "" {
		connection = "redis"
	}
	cfg := app.GetConfig()
	if cfg == nil {
		return nil, errors.New("session: global config is nil")
	}
	raw, ok := cfg.Database.Connections[connection]
	if !ok {
		return nil, fmt.Errorf("session: connection %q not found in database.yaml", connection)
	}
	conn, ok := raw.(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("session: connection %q has invalid structure", connection)
	}

	client := redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%d", strOr(conn, "host", "127.0.0.1"), intOr(conn, "port", 6379)),
		Password: strOr(conn, "password", ""),
		DB:       intOr(conn, "database", 0),
		PoolSize: intOr(conn, "pool_size", 10),
	})

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	if err := client.Ping(ctx).Err(); err != nil {
		_ = client.Close()
		return nil, fmt.Errorf("ping %s: %w", client.Options().Addr, err)
	}
	return &redisStore{client: client, ttl: ttl}, nil
}

// Load 反序列化存储值为 map。key 不存在时返回空 map，语义同 memoryStore。
func (rs *redisStore) Load(ctx context.Context, id string) (map[string]any, error) {
	raw, err := rs.client.Get(ctx, redisKeyPrefix+id).Bytes()
	if errors.Is(err, redis.Nil) {
		return map[string]any{}, nil
	}
	if err != nil {
		return nil, err
	}
	out := map[string]any{}
	if err := json.Unmarshal(raw, &out); err != nil {
		return nil, fmt.Errorf("session: unmarshal: %w", err)
	}
	return out, nil
}

// Save 将 data 序列化为 JSON 后写入，并附带 TTL。
func (rs *redisStore) Save(ctx context.Context, id string, data map[string]any, ttl time.Duration) error {
	if ttl <= 0 {
		ttl = rs.ttl
	}
	buf, err := json.Marshal(data)
	if err != nil {
		return fmt.Errorf("session: marshal: %w", err)
	}
	return rs.client.Set(ctx, redisKeyPrefix+id, buf, ttl).Err()
}

// Destroy 删除键。Del 对不存在的 key 返回 0，天然幂等。
func (rs *redisStore) Destroy(ctx context.Context, id string) error {
	return rs.client.Del(ctx, redisKeyPrefix+id).Err()
}

// Close 关闭 Redis client。
func (rs *redisStore) Close() error {
	return rs.client.Close()
}

// 下列工具函数与 app/cache 相同；重复声明避免跨包依赖带来的循环风险。

func strOr(m map[string]interface{}, key, fallback string) string {
	if v, ok := m[key].(string); ok && v != "" {
		return v
	}
	return fallback
}

func intOr(m map[string]interface{}, key string, fallback int) int {
	switch v := m[key].(type) {
	case int:
		return v
	case int32:
		return int(v)
	case int64:
		return int(v)
	case float64:
		return int(v)
	default:
		return fallback
	}
}
