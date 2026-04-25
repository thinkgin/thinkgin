// Package framework 封装 ThinkGin 的应用生命周期。
//
// 核心职责：
//   1. 把配置、日志、路由、HTTP Server 组装成一个可运行的 App 实例。
//   2. 提供 Run / Shutdown 契约，配合 context 支持优雅停机。
//   3. 通过函数式 Option 暴露有限的扩展点，避免侵入式继承。
package framework

import (
	"context"
	"crypto/tls"
	"errors"
	"fmt"
	"net/http"
	"time"

	"thinkgin/app"
	"thinkgin/route"

	"github.com/gin-gonic/gin"
)

// defaultShutdownTimeout 优雅停机的默认超时，用于等待 in-flight 请求收尾。
const defaultShutdownTimeout = 10 * time.Second

// App 代表一个 ThinkGin 应用实例。
// 字段一律私有，外部仅通过 Option 与方法交互。
type App struct {
	config          *app.GlobalConfig
	logger          app.Logger
	router          *gin.Engine
	server          *http.Server
	tlsServer       *http.Server // HTTPS server，仅在 config.server.https.enabled=true 时非 nil
	addr            string
	tlsAddr         string
	openBrowser     bool
	shutdownTimeout time.Duration
}

// New 按 Option 顺序组装 App。
// 未通过 Option 指定的依赖会回退到包级默认值（config/logger）。
func New(opts ...Option) (*App, error) {
	a := &App{
		config:          app.GetConfig(),
		logger:          app.GetLogger(),
		shutdownTimeout: defaultShutdownTimeout,
	}

	for _, opt := range opts {
		if opt != nil {
			opt(a)
		}
	}

	if a.config == nil {
		return nil, errors.New("framework: config is nil")
	}
	if a.logger == nil {
		return nil, errors.New("framework: logger is nil")
	}

	// 运行模式必须在路由初始化前设置，否则 Gin 的调试日志会先输出。
	gin.SetMode(a.config.Server.Mode)

	if a.router == nil {
		a.router = route.InitRouter()
	}

	if a.addr == "" {
		a.addr = fmt.Sprintf("%s:%d", a.config.Server.HTTP.Host, a.config.Server.HTTP.Port)
	}

	a.server = &http.Server{
		Addr:           a.addr,
		Handler:        a.router,
		ReadTimeout:    time.Duration(a.config.Server.HTTP.ReadTimeout) * time.Second,
		WriteTimeout:   time.Duration(a.config.Server.HTTP.WriteTimeout) * time.Second,
		IdleTimeout:    time.Duration(a.config.Server.HTTP.IdleTimeout) * time.Second,
		MaxHeaderBytes: a.config.Server.HTTP.MaxHeaderBytes,
	}

	// 若启用 HTTPS，构造 TLS server；证书文件不存在时 Run 阶段才会报错。
	if a.config.Server.HTTPS.Enabled {
		port := a.config.Server.HTTPS.Port
		if port == 0 {
			port = 443
		}
		a.tlsAddr = fmt.Sprintf("%s:%d", a.config.Server.HTTP.Host, port)
		a.tlsServer = &http.Server{
			Addr:           a.tlsAddr,
			Handler:        a.router,
			ReadTimeout:    time.Duration(a.config.Server.HTTP.ReadTimeout) * time.Second,
			WriteTimeout:   time.Duration(a.config.Server.HTTP.WriteTimeout) * time.Second,
			IdleTimeout:    time.Duration(a.config.Server.HTTP.IdleTimeout) * time.Second,
			MaxHeaderBytes: a.config.Server.HTTP.MaxHeaderBytes,
			TLSConfig:      &tls.Config{MinVersion: tls.VersionTLS12},
		}
	}

	return a, nil
}

// Run 启动 HTTP 服务并阻塞，直到 ctx 被取消或监听出错。
// ctx 为 nil 时等价于 context.Background()。
//
//nolint:contextcheck // ctx==nil 兜底到 Background 是显式约定的行为，非遗漏继承。
func (a *App) Run(ctx context.Context) error {
	if ctx == nil {
		ctx = context.Background()
	}

	a.logger.Infof("server listening on http://%s", a.addr)
	a.logger.Infof("mode=%s, app=%s v%s", a.config.Server.Mode, a.config.App.Name, a.config.App.Version)

	errCh := make(chan error, 2)
	go func() {
		errCh <- a.server.ListenAndServe()
	}()

	// 若启用了 HTTPS，并发启动 TLS 监听。
	if a.tlsServer != nil {
		cert := a.config.Server.HTTPS.CertFile
		key := a.config.Server.HTTPS.KeyFile
		a.logger.Infof("server listening on https://%s", a.tlsAddr)
		go func() {
			errCh <- a.tlsServer.ListenAndServeTLS(cert, key)
		}()
	}

	// 延迟异步打开浏览器，等待监听端口就绪；失败静默忽略。
	if a.openBrowser && a.config.App.Debug {
		go func() {
			time.Sleep(300 * time.Millisecond)
			tryOpenBrowser(fmt.Sprintf("http://%s", a.addr))
		}()
	}

	select {
	case <-ctx.Done():
		shutdownCtx, cancel := context.WithTimeout(context.Background(), a.shutdownTimeout)
		defer cancel()
		return a.Shutdown(shutdownCtx)
	case err := <-errCh:
		if errors.Is(err, http.ErrServerClosed) {
			return nil
		}
		return err
	}
}

// Shutdown 触发 HTTP Server 的优雅停机。
// 在 Server 未初始化时返回 nil。
func (a *App) Shutdown(ctx context.Context) error {
	var errs []error
	if a.server != nil {
		if err := a.server.Shutdown(ctx); err != nil {
			errs = append(errs, fmt.Errorf("http shutdown: %w", err))
		}
	}
	if a.tlsServer != nil {
		if err := a.tlsServer.Shutdown(ctx); err != nil {
			errs = append(errs, fmt.Errorf("https shutdown: %w", err))
		}
	}
	return errors.Join(errs...)
}
