// 本文件定义 ThinkGin 的缓存抽象层。
//
// 设计目标：
//   - 统一 Cache 接口，业务代码通过 Get/Set/Delete/Has 操作缓存，不关心底层驱动。
//   - 内置 memory 和 redis 两种实现，通过 driver 配置切换，代码零修改。
//   - Remember 模式：缓存不命中时自动执行回调并写入缓存。
//
// 用法：
//
//	store := cache.NewStore("default")  // 从配置创建
//	store.Set(ctx, "key", "value", 5*time.Minute)
//	val, err := store.Get(ctx, "key")
package cache

import (
	"context"
	"time"
)

// Store 是 ThinkGin 缓存的统一抽象接口。
// 所有业务代码应依赖此接口，而非直接使用 *redis.Client 或 sync.Map。
type Store interface {
	// Get 获取缓存值。key 不存在时返回 ("", ErrCacheMiss)。
	Get(ctx context.Context, key string) (string, error)

	// Set 写入缓存。ttl=0 表示永不过期（memory 驱动下永驻内存）。
	Set(ctx context.Context, key string, value string, ttl time.Duration) error

	// Delete 删除一个或多个 key。
	Delete(ctx context.Context, keys ...string) error

	// Has 判断 key 是否存在。
	Has(ctx context.Context, key string) (bool, error)

	// Remember 缓存不命中时执行 fn 并写入缓存。
	// 命中时直接返回缓存值，不执行 fn。
	Remember(ctx context.Context, key string, ttl time.Duration, fn func() (string, error)) (string, error)

	// Flush 清空该 Store 下的所有缓存。
	Flush(ctx context.Context) error
}
