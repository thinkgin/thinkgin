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

// ErrCacheMiss 在 Store.Get 未命中缓存时返回。
var ErrCacheMiss = errors.New("cache: miss")

// NewStore 根据缓存名称创建统一的 Store 接口。
// driver=redis 时复用已建立的连接；driver=memory 或其他则返回内存实现。
func NewStore(name string) Store {
	cfg := app.GetConfig()

	// 尝试从配置中获取 driver 信息
	if cfg != nil {
		if store, ok := cfg.Cache.Stores[name]; ok {
			if store.Driver == "redis" {
				if client, err := Get(name); err == nil {
					return NewRedisStore(client, cfg.Cache.Prefix)
				}
			}
		}
	}

	// 默认回退到内存实现
	return NewMemoryStore()
}

// DefaultStore 返回 config.cache.default 指向的 Store 接口。
func DefaultStore() Store {
	cfg := app.GetConfig()
	if cfg != nil && cfg.Cache.Default != "" {
		return NewStore(cfg.Cache.Default)
	}
	return NewMemoryStore()
}

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
	for name, store := range cfg.Cache.Stores {
		if store.Driver != "redis" {
			// memory / file 暂不实现，不报错也不产生 client。
			continue
		}

		conn, exists := dbConns[store.Connection]
		if !exists {
			errs = append(errs, fmt.Errorf("%s: connection %q not found in database.yaml", name, store.Connection))
			continue
		}

		client, err := openRedis(conn)
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

// openRedis 从 ConnectionConfig 构造 *redis.Client 并 Ping 探活。
// 探活失败时 client 会被关闭并返回错误，避免泄漏。
func openRedis(conn app.ConnectionConfig) (*redis.Client, error) {
	host := conn.Host
	if host == "" {
		host = "127.0.0.1"
	}
	port := conn.Port
	if port == 0 {
		port = 6379
	}
	poolSize := conn.PoolSize
	if poolSize == 0 {
		poolSize = 10
	}

	opt := &redis.Options{
		Addr:     fmt.Sprintf("%s:%d", host, port),
		Password: conn.Password,
		DB:       0,
		PoolSize: poolSize,
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
