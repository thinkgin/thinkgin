package cache

import (
	"context"
	"errors"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"thinkgin/app"
)

func TestMemoryStore_SetGet(t *testing.T) {
	ms := NewMemoryStore()
	defer ms.Close()
	ctx := context.Background()

	if err := ms.Set(ctx, "k1", "v1", time.Minute); err != nil {
		t.Fatal(err)
	}
	val, err := ms.Get(ctx, "k1")
	if err != nil {
		t.Fatal(err)
	}
	if val != "v1" {
		t.Errorf("Get=%q, want %q", val, "v1")
	}
}

func TestMemoryStore_GetMiss(t *testing.T) {
	ms := NewMemoryStore()
	defer ms.Close()
	ctx := context.Background()

	_, err := ms.Get(ctx, "nonexist")
	if !errors.Is(err, ErrCacheMiss) {
		t.Errorf("err=%v, want ErrCacheMiss", err)
	}
}

func TestMemoryStore_Expiration(t *testing.T) {
	ms := NewMemoryStore()
	defer ms.Close()
	ctx := context.Background()

	_ = ms.Set(ctx, "expiring", "val", 1*time.Millisecond)
	time.Sleep(5 * time.Millisecond)

	_, err := ms.Get(ctx, "expiring")
	if !errors.Is(err, ErrCacheMiss) {
		t.Error("expired key should return ErrCacheMiss")
	}
}

func TestMemoryStore_Delete(t *testing.T) {
	ms := NewMemoryStore()
	defer ms.Close()
	ctx := context.Background()

	_ = ms.Set(ctx, "k1", "v1", 0)
	_ = ms.Delete(ctx, "k1")

	_, err := ms.Get(ctx, "k1")
	if !errors.Is(err, ErrCacheMiss) {
		t.Error("deleted key should return ErrCacheMiss")
	}
}

func TestMemoryStore_Has(t *testing.T) {
	ms := NewMemoryStore()
	defer ms.Close()
	ctx := context.Background()

	_ = ms.Set(ctx, "k1", "v1", 0)
	ok, _ := ms.Has(ctx, "k1")
	if !ok {
		t.Error("Has should return true for existing key")
	}

	ok, _ = ms.Has(ctx, "missing")
	if ok {
		t.Error("Has should return false for missing key")
	}
}

func TestMemoryStore_Remember(t *testing.T) {
	ms := NewMemoryStore()
	defer ms.Close()
	ctx := context.Background()

	calls := 0
	fn := func() (string, error) {
		calls++
		return "computed", nil
	}

	v1, _ := ms.Remember(ctx, "rem", time.Minute, fn)
	v2, _ := ms.Remember(ctx, "rem", time.Minute, fn)

	if v1 != "computed" || v2 != "computed" {
		t.Errorf("v1=%q, v2=%q, want both %q", v1, v2, "computed")
	}
	if calls != 1 {
		t.Errorf("fn called %d times, want 1 (cached)", calls)
	}
}

func TestMemoryStore_Flush(t *testing.T) {
	ms := NewMemoryStore()
	defer ms.Close()
	ctx := context.Background()

	_ = ms.Set(ctx, "k1", "v1", 0)
	_ = ms.Set(ctx, "k2", "v2", 0)
	_ = ms.Flush(ctx)

	_, err := ms.Get(ctx, "k1")
	if !errors.Is(err, ErrCacheMiss) {
		t.Error("Flush should clear all keys")
	}
}

func TestRedisStore_FlushRefusesEmptyPrefix(t *testing.T) {
	// prefix 为空时 Flush 必须拒绝执行（避免 FlushDB 误清全库），
	// 该分支在触达 Redis 之前就返回，因此无需真实连接即可验证。
	rs := NewRedisStore(nil, "")
	if err := rs.Flush(context.Background()); err == nil {
		t.Fatal("Flush with empty prefix must return an error, not wipe the whole DB")
	}
}

func TestMemoryStore_RememberSingleflight(t *testing.T) {
	ms := NewMemoryStore()
	defer ms.Close()
	ctx := context.Background()

	var calls int64
	start := make(chan struct{})
	fn := func() (string, error) {
		atomic.AddInt64(&calls, 1)
		time.Sleep(20 * time.Millisecond) // 模拟较慢的回源，制造并发窗口
		return "computed", nil
	}

	const goroutines = 50
	var wg sync.WaitGroup
	wg.Add(goroutines)
	for i := 0; i < goroutines; i++ {
		go func() {
			defer wg.Done()
			<-start
			v, err := ms.Remember(ctx, "hot", time.Minute, fn)
			if err != nil || v != "computed" {
				t.Errorf("Remember=%q err=%v", v, err)
			}
		}()
	}
	close(start) // 同时放行，最大化并发未命中
	wg.Wait()

	if got := atomic.LoadInt64(&calls); got != 1 {
		t.Errorf("fn called %d times under concurrent miss, want 1 (singleflight)", got)
	}
}

func TestNewStore_DefaultsToMemory(t *testing.T) {
	old := app.Config
	app.Config = &app.GlobalConfig{}
	defer func() { app.Config = old }()

	store := NewStore("default")
	if _, ok := store.(*MemoryStore); !ok {
		t.Error("NewStore should default to MemoryStore when no config")
	}
	// 不在此处 Close：内存 Store 现在是进程级共享单例，由进程生命周期管理。
}

func TestDefaultStore_ReturnsMemory(t *testing.T) {
	old := app.Config
	app.Config = &app.GlobalConfig{}
	defer func() { app.Config = old }()

	store := DefaultStore()
	if _, ok := store.(*MemoryStore); !ok {
		t.Error("DefaultStore should return MemoryStore when no redis")
	}
}

func TestSharedMemoryStore_IsSingleton(t *testing.T) {
	old := app.Config
	app.Config = &app.GlobalConfig{}
	defer func() { app.Config = old }()

	// 多次获取内存 Store 必须是同一实例，证明不会重复创建 GC goroutine。
	s1 := NewStore("nonexistent-memory")
	s2 := DefaultStore()
	s3 := getSharedMemoryStore()
	if s1 != Store(s3) || s2 != Store(s3) {
		t.Error("memory store must be a process-wide singleton to avoid goroutine leak")
	}
}
