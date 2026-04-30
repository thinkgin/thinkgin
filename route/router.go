// Package route 负责 Gin 引擎的装配：全局中间件、静态资源、模板、业务路由及监控端点。
package route

import (
	"path/filepath"

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
// 装配顺序严格：ServiceContext → 全局中间件 → 静态资源/模板 → Web → API → 监控 → NoRoute 兜底。
func InitRouter() *gin.Engine {
	r := gin.New()

	cfg := app.GetConfig()
	gin.SetMode(cfg.Server.Mode)

	// 将 ServiceContext 注入请求链路，后续中间件和 Handler 可通过 app.SvcFromGin(c) 获取。
	svc := &app.ServiceContext{
		Config: cfg,
		Log:    app.GetLogger(),
	}
	r.Use(app.SvcMiddleware(svc))

	registerGlobalMiddleware(r, cfg)
	registerStaticAndTemplates(r)
	RegisterWebRoutes(r)
	RegisterAPIRoutes(r)
	registerMonitoringRoutes(r, cfg)
	registerSwaggerRoutes(r, cfg)

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

// ApplyGroupMiddleware 按组名从 middleware.groups 配置中读取中间件列表，批量挂载到指定路由组。
// 配置中未定义的组名时不做任何操作，方便业务代码无缝调用。
func ApplyGroupMiddleware(group *gin.RouterGroup, groupName string) {
	cfg := app.GetConfig()
	if cfg == nil {
		return
	}
	names, ok := cfg.Middleware.Groups[groupName]
	if !ok || len(names) == 0 {
		return
	}
	for _, name := range names {
		if h := resolveMiddleware(name); h != nil {
			group.Use(h)
		}
	}
}

// applyMiddleware 将单个中间件名映射到具体实现并注册到引擎。
// 未知名称被静默忽略，方便前向兼容。
func applyMiddleware(r *gin.Engine, name string) {
	if h := resolveMiddleware(name); h != nil {
		r.Use(h)
	}
}

// resolveMiddleware 将中间件名解析为具体的 HandlerFunc。
// 未知名称返回 nil，方便前向兼容。
func resolveMiddleware(name string) gin.HandlerFunc {
	switch name {
	case "recovery":
		return middleware.Recovery()
	case "request_id":
		return middleware.RequestID()
	case "trace":
		return middleware.TraceMiddleware()
	case "logger", "access_log":
		return middleware.AccessLogger()
	case "cors":
		return middleware.CORS()
	case "rate_limit":
		return middleware.RateLimit()
	case "session":
		return session.Middleware()
	case "jwt", "jwt_auth", "auth":
		return middleware.JWTAuth()
	case "prometheus":
		middleware.InitPrometheusMetrics()
		return middleware.PrometheusMiddleware()
	case "secure_headers":
		return middleware.SecureHeaders()
	case "gzip":
		return middleware.Gzip()
	case "csrf":
		return middleware.CSRF()
	case "redis_rate_limit":
		return middleware.RedisRateLimit()
	case "circuit_breaker":
		return middleware.CircuitBreaker()
	case "timeout":
		return middleware.Timeout()
	case "body_limit":
		return middleware.BodyLimit()
	case "validation":
		// validation 不是全局中间件，而是在 Handler 内通过 BindAndValidate 调用。
		// 此处返回 no-op，仅为了让 middleware.groups 配置中写 "validation" 不报错。
		return func(c *gin.Context) { c.Next() }
	default:
		return nil
	}
}

// registerStaticAndTemplates 注册静态资源目录与 HTML 模板。
// 模板路径约定：app/<module>/view/<tpl>.html。
func registerStaticAndTemplates(r *gin.Engine) {
	r.Static("/static", "./static")
	r.Static("/public", "./public")
	r.Static("/uploads", "./public/uploads")
	// 仅加载 .html 文件，避免把 .gitkeep、空目录或其他杂项当成模板。
	// 先用 Glob 检查是否有匹配文件，纯 API 项目无模板时跳过，避免 panic。
	const tplGlob = "app/*/view/*.html"
	if matches, _ := filepath.Glob(tplGlob); len(matches) > 0 {
		r.LoadHTMLGlob(tplGlob)
	}
}

// registerMonitoringRoutes 注册 k8s 探针与 Prometheus 端点。
// Prometheus 端点需同时满足总开关与模块开关都为 true 才会暴露。
func registerMonitoringRoutes(r *gin.Engine, cfg *app.GlobalConfig) {
	r.GET("/livez", livezHandler())
	// /readyz 会探测已注册的 DB/Redis；均为空时也返回 200。
	r.GET("/readyz", readyzHandler())
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
