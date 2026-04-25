package app

import (
	"os"
	"path/filepath"
	"testing"
)

func TestBootstrap_LoadsConfigFromDir(t *testing.T) {
	// 用临时目录创建最小配置文件测试 Bootstrap 完整流程。
	dir := t.TempDir()

	// 写最小 app.yaml
	appYAML := []byte("app:\n  name: BootstrapTest\n  version: \"9.9.9\"\n  debug: true\n")
	if err := os.WriteFile(filepath.Join(dir, "app.yaml"), appYAML, 0644); err != nil {
		t.Fatal(err)
	}

	// 写最小 server.yaml
	srvYAML := []byte("server:\n  mode: test\n  http:\n    host: 127.0.0.1\n    port: 9999\n")
	if err := os.WriteFile(filepath.Join(dir, "server.yaml"), srvYAML, 0644); err != nil {
		t.Fatal(err)
	}

	// 写最小 log.yaml
	logYAML := []byte("log:\n  default:\n    level: debug\n    format: text\n  file:\n    path: " + filepath.Join(dir, "log") + "\n    filename: test\n")
	if err := os.WriteFile(filepath.Join(dir, "log.yaml"), logYAML, 0644); err != nil {
		t.Fatal(err)
	}

	// 强制重新 Bootstrap（绕过 sync.Once）
	if err := ReBootstrap(dir); err != nil {
		t.Logf("Bootstrap 部分文件缺失属正常: %v", err)
	}

	cfg := GetConfig()
	if cfg == nil {
		t.Fatal("GetConfig() returned nil after Bootstrap")
	}
	if cfg.App.Name != "BootstrapTest" {
		t.Errorf("App.Name = %q, want BootstrapTest", cfg.App.Name)
	}
	if cfg.Server.HTTP.Port != 9999 {
		t.Errorf("Server.HTTP.Port = %d, want 9999", cfg.Server.HTTP.Port)
	}
}

func TestBootstrap_FallsBackToDefaults(t *testing.T) {
	// 空目录 Bootstrap 应使用默认值而不 panic
	dir := t.TempDir()

	// 先清空全局 Config，确保不受前一个测试影响
	Config = nil

	if err := ReBootstrap(dir); err != nil {
		t.Logf("预期错误（无配置文件）: %v", err)
	}

	cfg := GetConfig()
	if cfg == nil {
		t.Fatal("GetConfig() returned nil")
	}
	if cfg.App.Name != "ThinkGin" {
		t.Errorf("App.Name = %q, want ThinkGin (default)", cfg.App.Name)
	}
}
