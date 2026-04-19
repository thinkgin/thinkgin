package framework

import (
	"time"

	"thinkgin/app"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

// Option 用于向 New 注入配置，遵循 Dave Cheney 的函数式选项模式。
type Option func(*App)

// WithConfig 注入自定义配置，通常在测试中使用。
// 生产环境直接依赖 app.GetConfig() 即可。
func WithConfig(cfg *app.GlobalConfig) Option {
	return func(a *App) {
		if cfg != nil {
			a.config = cfg
		}
	}
}

// WithLogger 注入自定义 Logger，便于重定向日志输出或适配测试。
func WithLogger(logger *logrus.Logger) Option {
	return func(a *App) {
		if logger != nil {
			a.logger = logger
		}
	}
}

// WithRouter 允许外部传入已构造好的 Gin 引擎。
// 若未提供，New 内部会调用 route.InitRouter 构造默认路由。
func WithRouter(router *gin.Engine) Option {
	return func(a *App) {
		if router != nil {
			a.router = router
		}
	}
}

// WithAddr 覆盖监听地址；空字符串不生效。
func WithAddr(addr string) Option {
	return func(a *App) {
		if addr != "" {
			a.addr = addr
		}
	}
}

// WithOpenBrowser 控制启动时是否尝试自动打开浏览器。
// 仅在 App.Debug=true 且调用方显式启用时才会生效。
func WithOpenBrowser(enabled bool) Option {
	return func(a *App) {
		a.openBrowser = enabled
	}
}

// WithShutdownTimeout 设置优雅停机超时；非正数被忽略。
func WithShutdownTimeout(d time.Duration) Option {
	return func(a *App) {
		if d > 0 {
			a.shutdownTimeout = d
		}
	}
}
