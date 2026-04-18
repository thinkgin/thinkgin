package framework

import (
	"context"
	"testing"
	"time"

	"github.com/sirupsen/logrus"
	"thinkgin/app"
)

// TestNewApp 验证 App 创建
func TestNewApp(t *testing.T) {
	cfg := &app.GlobalConfig{}
	cfg.App.Name = "TestApp"
	cfg.App.Version = "3.0.0"
	cfg.Server.HTTP.Host = "127.0.0.1"
	cfg.Server.HTTP.Port = 0 // 随机端口
	cfg.Server.Mode = "test"

	logger := logrus.New()

	appInstance, err := New(
		WithConfig(cfg),
		WithLogger(logger),
		WithShutdownTimeout(3*time.Second),
	)

	if err != nil {
		t.Fatalf("New() error: %v", err)
	}
	if appInstance == nil {
		t.Fatal("New() returned nil")
	}
	if appInstance.config != cfg {
		t.Error("config not injected correctly")
	}
	if appInstance.logger != logger {
		t.Error("logger not injected correctly")
	}
	if appInstance.shutdownTimeout != 3*time.Second {
		t.Errorf("shutdownTimeout = %v, want 3s", appInstance.shutdownTimeout)
	}
}

// TestOptions 验证所有 Option 函数
func TestOptions(t *testing.T) {
	a := &App{}

	cfg := &app.GlobalConfig{}
	WithConfig(cfg)(a)
	if a.config != cfg {
		t.Error("WithConfig failed")
	}

	logger := logrus.New()
	WithLogger(logger)(a)
	if a.logger != logger {
		t.Error("WithLogger failed")
	}

	WithAddress("0.0.0.0:9999")(a)
	if a.addr != "0.0.0.0:9999" {
		t.Errorf("WithAddress = %s, want 0.0.0.0:9999", a.addr)
	}

	WithShutdownTimeout(5 * time.Second)(a)
	if a.shutdownTimeout != 5*time.Second {
		t.Errorf("WithShutdownTimeout = %v, want 5s", a.shutdownTimeout)
	}

	WithOpenBrowser(true)(a)
	if !a.openBrowser {
		t.Error("WithOpenBrowser failed")
	}
}

// TestAppShutdown 验证优雅停机
func TestAppShutdown(t *testing.T) {
	cfg := &app.GlobalConfig{}
	cfg.Server.HTTP.Host = "127.0.0.1"
	cfg.Server.HTTP.Port = 0
	cfg.Server.Mode = "test"

	logger := logrus.New()

	appInstance, err := New(
		WithConfig(cfg),
		WithLogger(logger),
		WithShutdownTimeout(1*time.Second),
	)
	if err != nil {
		t.Fatalf("New() error: %v", err)
	}

	// Shutdown 不应 panic（即使 server 未启动）
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()
	appInstance.Shutdown(ctx)
}
