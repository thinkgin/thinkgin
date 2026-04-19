// 本文件提供进程内 Session 存储。
//
// 适用场景：单实例部署、开发调试。
// 限制：多副本部署时每个进程各自持有数据，用户被负载均衡分派到不同实例就会"丢失登录"。
package session

import (
	"context"
	"sync"
	"time"
)

// memoryEntry 表示一条带过期时间的 Session 数据。
// data 用 map[string]any 保留 JSON 可序列化的灵活性，与 Redis 后端保持一致语义。
type memoryEntry struct {
	data      map[string]any
	expiresAt time.Time
}

type memoryStore struct {
	mu           sync.RWMutex
	data         map[string]memoryEntry
	ttl          time.Duration
	stopCleanup  chan struct{}
	cleanupOnce  sync.Once
	closedOnce   sync.Once
}

// newMemoryStore 构造进程内存储并启动后台 GC。
// GC 周期固定为 TTL 的 1/4，最少 1 分钟，避免高频扫描。
func newMemoryStore(ttl time.Duration) *memoryStore {
	ms := &memoryStore{
		data:        map[string]memoryEntry{},
		ttl:         ttl,
		stopCleanup: make(chan struct{}),
	}
	ms.cleanupOnce.Do(func() {
		go ms.cleanupLoop()
	})
	return ms
}

// Load 返回 id 对应的 Session 数据。
// 未找到或已过期时返回空 map 并不报错，让中间件把它当作新会话处理。
func (ms *memoryStore) Load(_ context.Context, id string) (map[string]any, error) {
	ms.mu.RLock()
	entry, ok := ms.data[id]
	ms.mu.RUnlock()
	if !ok || time.Now().After(entry.expiresAt) {
		return map[string]any{}, nil
	}
	// 返回浅拷贝，避免调用方意外修改存储内部状态。
	out := make(map[string]any, len(entry.data))
	for k, v := range entry.data {
		out[k] = v
	}
	return out, nil
}

// Save 写入数据并刷新过期时间。ttl=0 时使用 store 初始化时的默认。
func (ms *memoryStore) Save(_ context.Context, id string, data map[string]any, ttl time.Duration) error {
	if ttl <= 0 {
		ttl = ms.ttl
	}
	dup := make(map[string]any, len(data))
	for k, v := range data {
		dup[k] = v
	}
	ms.mu.Lock()
	ms.data[id] = memoryEntry{data: dup, expiresAt: time.Now().Add(ttl)}
	ms.mu.Unlock()
	return nil
}

// Destroy 删除 id 对应的条目。不存在也视为成功，保持幂等。
func (ms *memoryStore) Destroy(_ context.Context, id string) error {
	ms.mu.Lock()
	delete(ms.data, id)
	ms.mu.Unlock()
	return nil
}

// Close 停止后台 GC goroutine。重复调用安全。
func (ms *memoryStore) Close() error {
	ms.closedOnce.Do(func() {
		close(ms.stopCleanup)
	})
	return nil
}

// cleanupLoop 按 TTL/4 周期扫描并清理过期条目；最少 1 分钟扫一次。
func (ms *memoryStore) cleanupLoop() {
	interval := ms.ttl / 4
	if interval < time.Minute {
		interval = time.Minute
	}
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case <-ms.stopCleanup:
			return
		case <-ticker.C:
			ms.gcOnce()
		}
	}
}

// gcOnce 扫一遍内存，删除已过期条目。持有写锁，避免与 Load/Save 竞争。
func (ms *memoryStore) gcOnce() {
	now := time.Now()
	ms.mu.Lock()
	defer ms.mu.Unlock()
	for id, entry := range ms.data {
		if now.After(entry.expiresAt) {
			delete(ms.data, id)
		}
	}
}
