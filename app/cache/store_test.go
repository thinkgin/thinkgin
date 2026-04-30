package cache

import (
	"context"
	"errors"
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

func TestNewStore_DefaultsToMemory(t *testing.T) {
	old := app.Config
	app.Config = &app.GlobalConfig{}
	defer func() { app.Config = old }()

	store := NewStore("default")
	_, ok := store.(*MemoryStore)
	if !ok {
		t.Error("NewStore should default to MemoryStore when no config")
	}
	if ms, ok := store.(*MemoryStore); ok {
		ms.Close()
	}
}

func TestDefaultStore_ReturnsMemory(t *testing.T) {
	old := app.Config
	app.Config = &app.GlobalConfig{}
	defer func() { app.Config = old }()

	store := DefaultStore()
	_, ok := store.(*MemoryStore)
	if !ok {
		t.Error("DefaultStore should return MemoryStore when no redis")
	}
	if ms, ok := store.(*MemoryStore); ok {
		ms.Close()
	}
}
