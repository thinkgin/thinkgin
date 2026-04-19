package app

import (
	"os"
	"path/filepath"
	"testing"
)

// TestSetDefaultConfig 验证默认值填充逻辑
func TestSetDefaultConfig(t *testing.T) {
	// 清空全局配置
	Config = &GlobalConfig{}

	setDefaultConfig()

	tests := []struct {
		name string
		got  interface{}
		want interface{}
	}{
		{"App.Name", Config.App.Name, "ThinkGin"},
		{"App.Version", Config.App.Version, "3.1.0"},
		{"Server.HTTP.Host", Config.Server.HTTP.Host, "0.0.0.0"},
		{"Server.HTTP.Port", Config.Server.HTTP.Port, 8000},
		{"Server.HTTP.ReadTimeout", Config.Server.HTTP.ReadTimeout, 60},
		{"Log.Default.Level", Config.Log.Default.Level, "info"},
		{"Log.Default.Format", Config.Log.Default.Format, "json"},
		{"Log.File.Path", Config.Log.File.Path, "runtime/log"},
		{"Prometheus.Path", Config.Prometheus.Path, "/metrics"},
		{"Prometheus.Namespace", Config.Prometheus.Namespace, "app"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.got != tt.want {
				t.Errorf("%s = %v, want %v", tt.name, tt.got, tt.want)
			}
		})
	}
}

// TestValidateConfig 验证配置校验和纠正
func TestValidateConfig(t *testing.T) {
	Config = &GlobalConfig{}
	setDefaultConfig()

	// 测试无效 mode 被纠正
	Config.Server.Mode = "invalid_mode"
	validateConfig()
	if Config.Server.Mode != "debug" {
		t.Errorf("invalid mode should fallback to debug, got %s", Config.Server.Mode)
	}

	// 测试无效端口被纠正
	Config.Server.HTTP.Port = -1
	validateConfig()
	if Config.Server.HTTP.Port != 8000 {
		t.Errorf("invalid port should fallback to 8000, got %d", Config.Server.HTTP.Port)
	}

	// 测试无效日志格式被纠正
	Config.Log.Default.Format = "xml"
	validateConfig()
	if Config.Log.Default.Format != "json" {
		t.Errorf("invalid format should fallback to json, got %s", Config.Log.Default.Format)
	}

	// 测试合法 mode 保持不变
	Config.Server.Mode = "release"
	validateConfig()
	if Config.Server.Mode != "release" {
		t.Errorf("valid mode should remain release, got %s", Config.Server.Mode)
	}
}

// TestApplyEnvOverrides 验证环境变量覆盖
func TestApplyEnvOverrides(t *testing.T) {
	Config = &GlobalConfig{}
	setDefaultConfig()

	os.Setenv("THINKGIN_SERVER_MODE", "release")
	os.Setenv("THINKGIN_SERVER_HTTP_PORT", "9090")
	os.Setenv("THINKGIN_APP_DEBUG", "true")
	defer func() {
		os.Unsetenv("THINKGIN_SERVER_MODE")
		os.Unsetenv("THINKGIN_SERVER_HTTP_PORT")
		os.Unsetenv("THINKGIN_APP_DEBUG")
	}()

	applyEnvOverrides()

	if Config.Server.Mode != "release" {
		t.Errorf("env override SERVER_MODE failed, got %s", Config.Server.Mode)
	}
	if Config.Server.HTTP.Port != 9090 {
		t.Errorf("env override SERVER_HTTP_PORT failed, got %d", Config.Server.HTTP.Port)
	}
	if !Config.App.Debug {
		t.Errorf("env override APP_DEBUG failed, got false")
	}
}

// TestLoadInto 验证泛型 YAML 加载
func TestLoadInto(t *testing.T) {
	// 创建临时 YAML 文件
	dir := t.TempDir()
	content := []byte("app:\n  name: TestApp\n  version: \"1.0.0\"\n")
	path := filepath.Join(dir, "app.yaml")
	if err := os.WriteFile(path, content, 0644); err != nil {
		t.Fatal(err)
	}

	var cfg AppConfig
	if err := loadInto("app", path, &cfg); err != nil {
		t.Fatalf("loadInto failed: %v", err)
	}

	if cfg.Name != "TestApp" {
		t.Errorf("Name = %s, want TestApp", cfg.Name)
	}
	if cfg.Version != "1.0.0" {
		t.Errorf("Version = %s, want 1.0.0", cfg.Version)
	}

	// 测试文件不存在
	if err := loadInto("app", filepath.Join(dir, "nonexist.yaml"), &cfg); err == nil {
		t.Error("expected error for nonexistent file")
	}

	// 测试空文件
	emptyPath := filepath.Join(dir, "empty.yaml")
	os.WriteFile(emptyPath, []byte(""), 0644)
	if err := loadInto("app", emptyPath, &cfg); err == nil {
		t.Error("expected error for empty file")
	}

	// 测试缺少根节点
	wrongRoot := filepath.Join(dir, "wrong.yaml")
	os.WriteFile(wrongRoot, []byte("server:\n  mode: debug\n"), 0644)
	if err := loadInto("app", wrongRoot, &cfg); err == nil {
		t.Error("expected error for missing root key")
	}
}
