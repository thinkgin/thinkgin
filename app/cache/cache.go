// Package cache 提供 Redis 客户端多实例管理。
//
// 本版本仅支持 redis driver，memory / file 驱动为未来工作。
// 命名风格与 app/database 对称：Init / Get / Default / CloseAll。
//
// 关系图：
//
//	config.cache.stores.<name>.driver=redis
//	       │
//	       │ connection 字段 → database.yaml.connections.<conn>
//	       ▼
//	redis.NewClient(opt)
package cache

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/redis/go-redis/v9"

	"thinkgin/app"
)

// ErrNotFound 在按名称获取 store 而不存在时返回。
var ErrNotFound = errors.New("cache: store not found")

// pingTimeout 限定 Init 阶段 Ping 的最大耗时，避免 Redis 不可达时阻塞启动。
const pingTimeout = 2 * time.Second

var (
	mu      sync.RWMutex
	stores  = map[string]*redis.Client{}
	defName string
)

// Init 扫描 config.cache.stores，对每个 driver=redis 的 store 建立连接。
// memory / file 等其他 driver 暂不支持，静默跳过。
//
// 每条连接会做一次 Ping 探活。任一连接失败聚合后返回，但不阻止其他成功的连接注册。
func Init() error {
	cfg := app.GetConfig()
	if cfg == nil {
		return errors.New("cache: global config is nil, did you forget Bootstrap?")
	}

	mu.Lock()
	defer mu.Unlock()
	defName = cfg.Cache.Default

	dbConns := cfg.Database.Connections

	var errs []error
	for name, raw := range cfg.Cache.Stores {
		store, ok := raw.(map[string]interface{})
		if !ok {
			errs = append(errs, fmt.Errorf("%s: invalid store structure", name))
			continue
		}
		driver, _ := store["driver"].(string)
		if driver != "redis" {
			// memory / file 暂不实现，不报错也不产生 client。
			continue
		}

		connRef, _ := store["connection"].(string)
		connRaw, exists := dbConns[connRef]
		if !exists {
			errs = append(errs, fmt.Errorf("%s: connection %q not found in database.yaml", name, connRef))
			continue
		}
		connMap, ok := connRaw.(map[string]interface{})
		if !ok {
			errs = append(errs, fmt.Errorf("%s: connection %q has invalid structure", name, connRef))
			continue
		}

		client, err := openRedis(connMap)
		if err != nil {
			errs = append(errs, fmt.Errorf("%s: %w", name, err))
			continue
		}
		stores[name] = client
	}

	return errors.Join(errs...)
}

// Get 按名称返回 Redis 客户端。未找到时返回 ErrNotFound。
func Get(name string) (*redis.Client, error) {
	mu.RLock()
	defer mu.RUnlock()
	c, ok := stores[name]
	if !ok {
		return nil, fmt.Errorf("%w: %q", ErrNotFound, name)
	}
	return c, nil
}

// Default 返回 config.cache.default 指向的客户端，不存在时返回 nil。
func Default() *redis.Client {
	mu.RLock()
	defer mu.RUnlock()
	return stores[defName]
}

// CloseAll 关闭所有 Redis 客户端。幂等。
func CloseAll() error {
	mu.Lock()
	defer mu.Unlock()

	var errs []error
	for name, c := range stores {
		if err := c.Close(); err != nil {
			errs = append(errs, fmt.Errorf("%s: close: %w", name, err))
		}
	}
	stores = map[string]*redis.Client{}
	return errors.Join(errs...)
}

// openRedis 从配置 map 构造 *redis.Client 并 Ping 探活。
// 探活失败时 client 会被关闭并返回错误，避免泄漏。
func openRedis(conn map[string]interface{}) (*redis.Client, error) {
	opt := &redis.Options{
		Addr:     fmt.Sprintf("%s:%d", strOr(conn, "host", "127.0.0.1"), intOr(conn, "port", 6379)),
		Password: strOr(conn, "password", ""),
		DB:       intOr(conn, "database", 0),
		PoolSize: intOr(conn, "pool_size", 10),
	}
	client := redis.NewClient(opt)

	ctx, cancel := context.WithTimeout(context.Background(), pingTimeout)
	defer cancel()
	if err := client.Ping(ctx).Err(); err != nil {
		_ = client.Close()
		return nil, fmt.Errorf("ping %s: %w", opt.Addr, err)
	}
	return client, nil
}

// 以下工具函数与 app/database 同型，此处重复定义是为了避免包之间的循环依赖。
// 如果将来出现第三个类似需求，可考虑抽到 app/internal/confutil 共用。

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
