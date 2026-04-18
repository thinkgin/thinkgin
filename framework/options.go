package framework

import (
	"time"

	"thinkgin/app"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

type Option func(*App)

func WithConfig(cfg *app.GlobalConfig) Option {
	return func(a *App) {
		if cfg != nil {
			a.Config = cfg
		}
	}
}

func WithLogger(logger *logrus.Logger) Option {
	return func(a *App) {
		if logger != nil {
			a.Logger = logger
		}
	}
}

func WithRouter(router *gin.Engine) Option {
	return func(a *App) {
		if router != nil {
			a.Router = router
		}
	}
}

func WithAddr(addr string) Option {
	return func(a *App) {
		if addr != "" {
			a.Addr = addr
		}
	}
}

func WithOpenBrowser(enabled bool) Option {
	return func(a *App) {
		a.openBrowser = &enabled
	}
}

func WithShutdownTimeout(d time.Duration) Option {
	return func(a *App) {
		if d > 0 {
			a.shutdownTimeout = d
		}
	}
}
