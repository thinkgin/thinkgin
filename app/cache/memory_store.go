// 本文件实现基于内存的 Store，适用于单实例 / 测试场景。
//
// 过期策略：惰性删除（读取时检查） + 后台 GC（每 60s 扫描一次）。
// 无持久化，进程重启即丢失。
package cache

import (
	"context"
	"sync"
	"time"

	"golang.org/x/sync/singleflight"
)

// memoryItem 内存缓存条目。
type memoryItem struct {
	value     string
	expiresAt time.Time // zero value 表示永不过期
}

func (m *memoryItem) expired(now time.Time) bool {
	return !m.expiresAt.IsZero() && now.After(m.expiresAt)
}

// MemoryStore 进程内缓存实现。
type MemoryStore struct {
	mu    sync.RWMutex
	items map[string]*memoryItem
	done  chan struct{}
	sf    singleflight.Group // 防缓存击穿：同 key 的并发回源合并为一次
}

// NewMemoryStore 创建内存缓存并启动后台 GC。
func NewMemoryStore() *MemoryStore {
	ms := &MemoryStore{
		items: make(map[string]*memoryItem),
		done:  make(chan struct{}),
	}
	go ms.gc()
	return ms
}

func (ms *MemoryStore) Get(_ context.Context, key string) (string, error) {
	ms.mu.RLock()
	item, ok := ms.items[key]
	ms.mu.RUnlock()

	if !ok || item.expired(time.Now()) {
		if ok {
			ms.mu.Lock()
			delete(ms.items, key)
			ms.mu.Unlock()
		}
		return "", ErrCacheMiss
	}
	return item.value, nil
}

func (ms *MemoryStore) Set(_ context.Context, key, value string, ttl time.Duration) error {
	item := &memoryItem{value: value}
	if ttl > 0 {
		item.expiresAt = time.Now().Add(ttl)
	}
	ms.mu.Lock()
	ms.items[key] = item
	ms.mu.Unlock()
	return nil
}

func (ms *MemoryStore) Delete(_ context.Context, keys ...string) error {
	ms.mu.Lock()
	for _, k := range keys {
		delete(ms.items, k)
	}
	ms.mu.Unlock()
	return nil
}

func (ms *MemoryStore) Has(_ context.Context, key string) (bool, error) {
	ms.mu.RLock()
	item, ok := ms.items[key]
	ms.mu.RUnlock()

	if !ok || item.expired(time.Now()) {
		return false, nil
	}
	return true, nil
}

// Remember 读取缓存，未命中时回源并写回。
//
// 使用 singleflight 防止缓存击穿：同一 key 的并发未命中只会执行一次 fn，
// 其余调用方共享同一结果，避免缓存失效瞬间大量请求同时穿透到后端。
func (ms *MemoryStore) Remember(ctx context.Context, key string, ttl time.Duration, fn func() (string, error)) (string, error) {
	if v, err := ms.Get(ctx, key); err == nil {
		return v, nil
	}

	v, err, _ := ms.sf.Do(key, func() (interface{}, error) {
		// 进入 singleflight 后二次检查：可能在排队期间已被其他请求填充。
		if cached, gerr := ms.Get(ctx, key); gerr == nil {
			return cached, nil
		}
		val, ferr := fn()
		if ferr != nil {
			return "", ferr
		}
		_ = ms.Set(ctx, key, val, ttl)
		return val, nil
	})
	if err != nil {
		return "", err
	}
	return v.(string), nil
}

func (ms *MemoryStore) Flush(_ context.Context) error {
	ms.mu.Lock()
	ms.items = make(map[string]*memoryItem)
	ms.mu.Unlock()
	return nil
}

// Close 停止后台 GC goroutine。
func (ms *MemoryStore) Close() {
	close(ms.done)
}

func (ms *MemoryStore) gc() {
	ticker := time.NewTicker(60 * time.Second)
	defer ticker.Stop()
	for {
		select {
		case <-ms.done:
			return
		case now := <-ticker.C:
			ms.mu.Lock()
			for k, item := range ms.items {
				if item.expired(now) {
					delete(ms.items, k)
				}
			}
			ms.mu.Unlock()
		}
	}
}
