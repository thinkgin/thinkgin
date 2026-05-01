package cache

import (
	"errors"
	"testing"

	"github.com/redis/go-redis/v9"

	"thinkgin/app"
)

func resetState() {
	mu.Lock()
	defer mu.Unlock()
	stores = map[string]*redis.Client{}
	defName = ""
}

// setConfig 构造一个可控的全局配置，避免测试用例间相互串扰。
func setConfig(cacheStores map[string]app.CacheStoreConfig, conns map[string]app.ConnectionConfig, def string) {
	app.Config = &app.GlobalConfig{}
	app.Config.Cache.Default = def
	app.Config.Cache.Stores = cacheStores
	app.Config.Database.Connections = conns
}

func TestInit_SkipsNonRedisDrivers(t *testing.T) {
	resetState()
	setConfig(map[string]app.CacheStoreConfig{
		"memory": {Driver: "memory"},
		"file":   {Driver: "file"},
	}, nil, "memory")

	if err := Init(); err != nil {
		t.Fatalf("non-redis drivers should be silently skipped, got err=%v", err)
	}
	if _, err := Get("memory"); !errors.Is(err, ErrNotFound) {
		t.Errorf("memory store should not be registered, got err=%v", err)
	}
}

func TestInit_MissingConnectionReference(t *testing.T) {
	resetState()
	setConfig(map[string]app.CacheStoreConfig{
		"redis": {Driver: "redis", Connection: "nonexistent"},
	}, map[string]app.ConnectionConfig{}, "redis")

	err := Init()
	if err == nil {
		t.Fatal("expected error when connection reference is missing")
	}
}

func TestInit_PingFailureIsReported(t *testing.T) {
	resetState()
	// 使用 RFC 5737 保留的 TEST-NET-1 地址，保证永不连通。
	setConfig(
		map[string]app.CacheStoreConfig{
			"redis": {Driver: "redis", Connection: "redis"},
		},
		map[string]app.ConnectionConfig{
			"redis": {
				Driver: "redis",
				Host:   "192.0.2.1",
				Port:   6379,
			},
		},
		"redis",
	)

	err := Init()
	if err == nil {
		t.Fatal("expected ping failure to be reported")
	}
	if _, getErr := Get("redis"); !errors.Is(getErr, ErrNotFound) {
		t.Errorf("failed client must not be registered, getErr=%v", getErr)
	}
}

func TestDefault_ReturnsNilWhenNoStore(t *testing.T) {
	resetState()
	if got := Default(); got != nil {
		t.Errorf("Default() without Init should be nil, got %v", got)
	}
}
