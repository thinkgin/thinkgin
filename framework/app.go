package framework

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"runtime"
	"strconv"
	"strings"
	"time"

	"thinkgin/app"
	"thinkgin/route"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

type App struct {
	Config          *app.GlobalConfig
	Logger          *logrus.Logger
	Router          *gin.Engine
	Server          *http.Server
	Addr            string
	openBrowser     *bool
	shutdownTimeout time.Duration
}

func New(opts ...Option) (*App, error) {
	a := &App{
		Config:          app.GetConfig(),
		Logger:          app.GetLogger(),
		shutdownTimeout: 10 * time.Second,
	}

	for _, opt := range opts {
		if opt != nil {
			opt(a)
		}
	}

	if a.Config == nil {
		return nil, fmt.Errorf("config is nil")
	}
	if a.Logger == nil {
		return nil, fmt.Errorf("logger is nil")
	}

	gin.SetMode(a.Config.Server.Mode)

	if a.Router == nil {
		a.Router = route.InitRouter()
	}

	serverConfig := a.Config.Server
	if a.Addr == "" {
		a.Addr = fmt.Sprintf("%s:%d", serverConfig.HTTP.Host, serverConfig.HTTP.Port)
	}

	a.Server = &http.Server{
		Addr:           a.Addr,
		Handler:        a.Router,
		ReadTimeout:    time.Duration(serverConfig.HTTP.ReadTimeout) * time.Second,
		WriteTimeout:   time.Duration(serverConfig.HTTP.WriteTimeout) * time.Second,
		IdleTimeout:    time.Duration(serverConfig.HTTP.IdleTimeout) * time.Second,
		MaxHeaderBytes: serverConfig.HTTP.MaxHeaderBytes,
	}

	return a, nil
}

func (a *App) Run(ctx context.Context) error {
	if ctx == nil {
		ctx = context.Background()
	}

	if a.Logger != nil {
		a.Logger.Infof("🚀 服务器启动在: http://%s", a.Addr)
		a.Logger.Infof("📋 运行模式: %s", a.Config.Server.Mode)
		a.Logger.Infof("📱 应用名称: %s v%s", a.Config.App.Name, a.Config.App.Version)
	}

	if shouldOpenBrowser(a) {
		_ = openBrowser(fmt.Sprintf("http://%s", a.Addr))
	}

	errCh := make(chan error, 1)
	go func() {
		errCh <- a.Server.ListenAndServe()
	}()

	select {
	case <-ctx.Done():
		shutdownCtx, cancel := context.WithTimeout(context.Background(), a.shutdownTimeout)
		defer cancel()
		return a.Shutdown(shutdownCtx)
	case err := <-errCh:
		if err == http.ErrServerClosed {
			return nil
		}
		return err
	}
}

func (a *App) Shutdown(ctx context.Context) error {
	if a.Server == nil {
		return nil
	}
	return a.Server.Shutdown(ctx)
}

func shouldOpenBrowser(a *App) bool {
	enabled := false
	if a.openBrowser != nil {
		enabled = *a.openBrowser
	} else {
		v := strings.TrimSpace(os.Getenv("THINKGIN_OPEN_BROWSER"))
		if v == "" {
			enabled = false
		} else {
			b, err := strconv.ParseBool(v)
			enabled = err == nil && b
		}
	}

	if !enabled {
		return false
	}
	if a.Config == nil {
		return false
	}
	if !a.Config.App.Debug {
		return false
	}
	if strings.ToLower(strings.TrimSpace(a.Config.Server.Mode)) != "debug" {
		return false
	}
	return true
}

func openBrowser(url string) error {
	switch runtime.GOOS {
	case "windows":
		return exec.Command("cmd", "/c", "start", url).Start()
	case "darwin":
		return exec.Command("open", url).Start()
	default:
		return exec.Command("xdg-open", url).Start()
	}
}
