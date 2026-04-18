package route

import (
	"github.com/gin-gonic/gin"

	"thinkgin/app"
	"thinkgin/extend/middleware"
)

// InitRouter 初始化路由器 - 主入口点
func InitRouter() *gin.Engine {
	r := gin.New()

	// 配置获取
	config := app.GetConfig()

	// 设置运行模式
	gin.SetMode(config.Server.Mode)

	// 注册全局中间件
	registerGlobalMiddleware(r)

	// 注册静态资源和模板
	registerStaticAndTemplates(r)

	// 注册Web路由 (页面路由)
	registerWebRoutes(r)

	// 注册API路由 (版本化API)
	registerAPIRoutes(r)

	// 注册监控路由
	registerMonitoringRoutes(r)

	// NoRoute
	r.NoRoute(func(c *gin.Context) {
		if middleware.IsAPIPath(c.Request.URL.Path) {
			middleware.APIError(c, 404, 404, "not found")
			return
		}
		c.AbortWithStatus(404)
	})

	return r
}

// registerGlobalMiddleware 注册全局中间件
func registerGlobalMiddleware(r *gin.Engine) {
	config := app.GetConfig()
	global := config.Middleware.Global
	if len(global) == 0 {
		// 默认启用
		r.Use(middleware.Recovery())
		r.Use(middleware.RequestID())
		r.Use(middleware.TraceMiddleware())
		r.Use(middleware.AccessLogger())
		middleware.InitPrometheusMetrics()
		r.Use(middleware.PrometheusMiddleware())
		return
	}

	hasPrometheus := false

	for _, name := range global {
		switch name {
		case "cors":
			r.Use(middleware.CORS())
		case "recovery":
			r.Use(middleware.Recovery())
		case "logger":
			r.Use(middleware.RequestID())
			r.Use(middleware.AccessLogger())
		case "rate_limit":
			r.Use(middleware.RateLimit())
		case "request_id":
			r.Use(middleware.RequestID())
		case "access_log":
			r.Use(middleware.AccessLogger())
		case "trace":
			r.Use(middleware.TraceMiddleware())
		case "prometheus":
			hasPrometheus = true
			middleware.InitPrometheusMetrics()
			r.Use(middleware.PrometheusMiddleware())
		default:
			// 未实现的中间件名称先忽略
		}
	}

	if !hasPrometheus {
		middleware.InitPrometheusMetrics()
		r.Use(middleware.PrometheusMiddleware())
	}

	// 这里可以添加更多全局中间件:
	// r.Use(middleware.CORS())        // 跨域中间件
	// r.Use(middleware.RateLimit())   // 限流中间件
	// r.Use(middleware.JWT())         // JWT认证中间件
}

// registerStaticAndTemplates 注册静态资源和模板
func registerStaticAndTemplates(r *gin.Engine) {
	// 静态文件服务
	r.Static("/static", "./static")
	r.Static("/public", "./public")
	r.Static("/uploads", "./public/uploads")

	// 模板加载 - 支持多模块模板
	r.LoadHTMLGlob("app/*/view/*")

	// 如果配置文件中定义了静态路径，可以使用配置
	// config := app.GetConfig()
	// if config.Server.Static.Path != "" {
	//     r.Static(config.Server.Static.Route, config.Server.Static.Path)
	// }
}

// registerWebRoutes 注册Web页面路由
func registerWebRoutes(r *gin.Engine) {
	RegisterWebRoutes(r)
}

// registerAPIRoutes 注册API路由
func registerAPIRoutes(r *gin.Engine) {
	RegisterAPIRoutes(r)
}

// registerMonitoringRoutes 注册监控相关路由
func registerMonitoringRoutes(r *gin.Engine) {
	config := app.GetConfig()

	// k8s 探针：进程存活检查
	r.GET("/livez", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"status": "ok",
		})
	})

	// k8s 探针：就绪检查（后续可接入 DB/Redis 等依赖检查）
	r.GET("/readyz", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"status": "ok",
		})
	})

	// 健康检查端点
	r.GET("/ping", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"status":  "ok",
			"message": "pong",
			"version": config.App.Version,
		})
	})

	// Prometheus监控端点
	if config.App.Monitoring.PrometheusEnabled && config.Prometheus.Enabled {
		r.GET(config.Prometheus.Path, middleware.PrometheusHandler())
	}

	// 可以添加更多监控端点:
	// r.GET("/health", HealthCheck)      // 详细健康检查
	// r.GET("/metrics/custom", CustomMetrics) // 自定义指标
}
