package database

import (
	"errors"
	"path/filepath"
	"testing"

	"gorm.io/gorm"

	"thinkgin/app"
)

// resetState 清空全局状态，避免包级单例在用例间串扰。
func resetState() {
	mu.Lock()
	defer mu.Unlock()
	dbs = map[string]*gorm.DB{}
	defName = ""
}

// newTestConfig 构造一个 SQLite 内存连接配置，测试无需外部 DB。
func newTestConfig(t *testing.T) {
	t.Helper()
	app.Config = &app.GlobalConfig{}
	app.Config.Database.Default = "primary"
	app.Config.Database.Connections = map[string]interface{}{
		"primary": map[string]interface{}{
			"driver":   "sqlite",
			"database": filepath.Join(t.TempDir(), "test.db"),
		},
		"secondary": map[string]interface{}{
			"driver":   "sqlite",
			"database": filepath.Join(t.TempDir(), "test2.db"),
		},
		"skipped_redis": map[string]interface{}{
			"driver": "redis", // 应被明确跳过
		},
	}
}

func TestInit_RegistersAllSQLConnections(t *testing.T) {
	resetState()
	newTestConfig(t)

	if err := Init(); err != nil {
		t.Fatalf("Init() returned error: %v", err)
	}
	defer CloseAll()

	if _, err := Get("primary"); err != nil {
		t.Errorf("primary connection should exist: %v", err)
	}
	if _, err := Get("secondary"); err != nil {
		t.Errorf("secondary connection should exist: %v", err)
	}
	// Redis 条目必须被跳过
	if _, err := Get("skipped_redis"); !errors.Is(err, ErrNotFound) {
		t.Errorf("redis entry should be skipped, got err=%v", err)
	}
}

func TestDefault_ReturnsConfiguredConnection(t *testing.T) {
	resetState()
	newTestConfig(t)

	if err := Init(); err != nil {
		t.Fatalf("Init() error: %v", err)
	}
	defer CloseAll()

	if db := Default(); db == nil {
		t.Fatal("Default() returned nil, expected primary connection")
	}
}

func TestGet_UnknownNameReturnsErrNotFound(t *testing.T) {
	resetState()
	newTestConfig(t)

	if err := Init(); err != nil {
		t.Fatalf("Init() error: %v", err)
	}
	defer CloseAll()

	_, err := Get("nonexistent")
	if !errors.Is(err, ErrNotFound) {
		t.Errorf("expected ErrNotFound, got %v", err)
	}
}

func TestInit_InvalidDriverReturnsError(t *testing.T) {
	resetState()
	app.Config = &app.GlobalConfig{}
	app.Config.Database.Default = "bad"
	app.Config.Database.Connections = map[string]interface{}{
		"bad": map[string]interface{}{"driver": "oracle"},
	}
	defer CloseAll()

	if err := Init(); err == nil {
		t.Fatal("expected error for unsupported driver, got nil")
	}
}
