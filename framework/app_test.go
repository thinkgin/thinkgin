package framework

import (
	"context"
	"testing"
	"time"

	"thinkgin/app"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

// newTestConfig 构造一个可直接用于测试的最小配置。
// 使用随机端口（Port=0）避免本地冲突。
func newTestConfig() *app.GlobalConfig {
	cfg := &app.GlobalConfig{}
	cfg.App.Name = "TestApp"
	cfg.App.Version = "3.0.0"
	cfg.Server.HTTP.Host = "127.0.0.1"
	cfg.Server.HTTP.Port = 0
	cfg.Server.Mode = gin.TestMode
	return cfg
}

func TestNew_AssemblesDependencies(t *testing.T) {
	cfg := newTestConfig()
	logger := logrus.New()

	// 注入独立的 Gin 引擎，避免触发默认路由中的模板加载。
	a, err := New(
		WithConfig(cfg),
		WithLogger(logger),
		WithRouter(gin.New()),
		WithShutdownTimeout(3*time.Second),
	)
	if err != nil {
		t.Fatalf("New() error: %v", err)
	}

	if a.config != cfg {
		t.Error("config not injected correctly")
	}
	if a.logger != logger {
		t.Error("logger not injected correctly")
	}
	if a.shutdownTimeout != 3*time.Second {
		t.Errorf("shutdownTimeout = %v, want 3s", a.shutdownTimeout)
	}
	if a.server == nil {
		t.Error("http.Server should be initialized")
	}
}

func TestOptions_Mutators(t *testing.T) {
	a := &App{}

	cfg := &app.GlobalConfig{}
	WithConfig(cfg)(a)
	if a.config != cfg {
		t.Error("WithConfig did not set config")
	}

	logger := logrus.New()
	WithLogger(logger)(a)
	if a.logger != logger {
		t.Error("WithLogger did not set logger")
	}

	WithAddr("0.0.0.0:9999")(a)
	if a.addr != "0.0.0.0:9999" {
		t.Errorf("WithAddr = %s, want 0.0.0.0:9999", a.addr)
	}

	WithShutdownTimeout(5 * time.Second)(a)
	if a.shutdownTimeout != 5*time.Second {
		t.Errorf("WithShutdownTimeout = %v, want 5s", a.shutdownTimeout)
	}

	WithOpenBrowser(true)(a)
	if !a.openBrowser {
		t.Error("WithOpenBrowser did not toggle the flag")
	}

	// 非正数被忽略，保持原值。
	WithShutdownTimeout(-1)(a)
	if a.shutdownTimeout != 5*time.Second {
		t.Error("negative shutdownTimeout should be ignored")
	}
}

func TestShutdown_BeforeRunIsNoop(t *testing.T) {
	a, err := New(
		WithConfig(newTestConfig()),
		WithLogger(logrus.New()),
		WithRouter(gin.New()),
		WithShutdownTimeout(time.Second),
	)
	if err != nil {
		t.Fatalf("New() error: %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	if err := a.Shutdown(ctx); err != nil {
		t.Errorf("Shutdown before Run should not error, got: %v", err)
	}
}
