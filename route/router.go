// Package route 负责 Gin 引擎的装配：全局中间件、静态资源、模板、业务路由及监控端点。
package route

import (
	"github.com/gin-gonic/gin"

	"thinkgin/app"
	"thinkgin/app/session"
	"thinkgin/extend/middleware"
)

// defaultMiddlewareChain 描述 middleware.global 未配置时的兜底启用顺序。
// 注意：Recovery 必须排在最前，用于兜底其他中间件内部的 panic。
var defaultMiddlewareChain = []string{
	"recovery",
	"request_id",
	"trace",
	"access_log",
	"prometheus",
}

// InitRouter 组装完整的 Gin 引擎。
// 装配顺序严格：全局中间件 → 静态资源/模板 → Web → API → 监控 → NoRoute 兜底。
func InitRouter() *gin.Engine {
	r := gin.New()

	cfg := app.GetConfig()
	gin.SetMode(cfg.Server.Mode)

	registerGlobalMiddleware(r, cfg)
	registerStaticAndTemplates(r)
	RegisterWebRoutes(r)
	RegisterAPIRoutes(r)
	registerMonitoringRoutes(r, cfg)

	// NoRoute：API 路径返回结构化 JSON，其余返回裸 404，避免 HTML 污染客户端。
	r.NoRoute(func(c *gin.Context) {
		if middleware.IsAPIPath(c.Request.URL.Path) {
			middleware.APIError(c, 404, 404, "not found")
			return
		}
		c.AbortWithStatus(404)
	})

	return r
}

// registerGlobalMiddleware 按配置列表精确注册全局中间件。
// 若 middleware.global 为空则使用 defaultMiddlewareChain，不做任何隐式追加。
func registerGlobalMiddleware(r *gin.Engine, cfg *app.GlobalConfig) {
	chain := cfg.Middleware.Global
	if len(chain) == 0 {
		chain = defaultMiddlewareChain
	}
	for _, name := range chain {
		applyMiddleware(r, name)
	}
}

// applyMiddleware 将单个中间件名映射到具体实现。
// 未知名称被静默忽略，方便前向兼容。
func applyMiddleware(r *gin.Engine, name string) {
	switch name {
	case "recovery":
		r.Use(middleware.Recovery())
	case "request_id":
		r.Use(middleware.RequestID())
	case "trace":
		r.Use(middleware.TraceMiddleware())
	case "logger", "access_log":
		r.Use(middleware.AccessLogger())
	case "cors":
		r.Use(middleware.CORS())
	case "rate_limit":
		r.Use(middleware.RateLimit())
	case "session":
		r.Use(session.Middleware())
	case "prometheus":
		middleware.InitPrometheusMetrics()
		r.Use(middleware.PrometheusMiddleware())
	}
}

// registerStaticAndTemplates 注册静态资源目录与 HTML 模板。
// 模板路径约定：app/<module>/view/<tpl>.html。
func registerStaticAndTemplates(r *gin.Engine) {
	r.Static("/static", "./static")
	r.Static("/public", "./public")
	r.Static("/uploads", "./public/uploads")
	// 仅加载 .html 文件，避免把 .gitkeep、空目录或其他杂项当成模板。
	r.LoadHTMLGlob("app/*/view/*.html")
}

// registerMonitoringRoutes 注册 k8s 探针与 Prometheus 端点。
// Prometheus 端点需同时满足总开关与模块开关都为 true 才会暴露。
func registerMonitoringRoutes(r *gin.Engine, cfg *app.GlobalConfig) {
	r.GET("/livez", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})
	// /readyz 目前仅返回 ok，未来可串联 DB/Redis 等依赖探活。
	r.GET("/readyz", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})
	r.GET("/ping", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"status":  "ok",
			"message": "pong",
			"version": cfg.App.Version,
		})
	})

	if cfg.App.Monitoring.PrometheusEnabled && cfg.Prometheus.Enabled {
		r.GET(cfg.Prometheus.Path, middleware.PrometheusHandler())
	}
}
