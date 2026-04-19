// 本文件负责 HTTP 四大维度指标：请求数、时延、请求体大小、响应体大小。
// 提供 PrometheusMiddleware 作为 Gin 中间件，以及 detectScope / detectRoute
// 两个标签提取辅助函数（控制 label cardinality，避免指标爆炸）。
package middleware

import (
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"

	"thinkgin/app"
)

// HTTP 四大指标，由 initHTTPMetrics 初始化；未启用时保持 nil 并被中间件跳过。
var (
	httpRequestsTotal   *prometheus.CounterVec
	httpRequestDuration *prometheus.HistogramVec
	httpRequestSize     *prometheus.HistogramVec
	httpResponseSize    *prometheus.HistogramVec
)

// initHTTPMetrics 注册 HTTP 四大指标。labels 维度统一为 method/scope/route(/status)。
// 注意：bucket 选取直接影响直方图内存占用，ExponentialBuckets 参数保守设置。
func initHTTPMetrics(cfg app.PrometheusConfig, labels prometheus.Labels) {
	httpRequestsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Namespace:   cfg.Namespace,
			Name:        cfg.Metrics.HTTP.RequestsTotal,
			Help:        "Total number of HTTP requests",
			ConstLabels: labels,
		},
		[]string{"method", "scope", "route", "status"},
	)

	httpRequestDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Namespace:   cfg.Namespace,
			Name:        cfg.Metrics.HTTP.RequestDuration,
			Help:        "HTTP request duration in seconds",
			ConstLabels: labels,
			Buckets:     prometheus.DefBuckets,
		},
		[]string{"method", "scope", "route"},
	)

	httpRequestSize = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Namespace:   cfg.Namespace,
			Name:        cfg.Metrics.HTTP.RequestSize,
			Help:        "HTTP request size in bytes",
			ConstLabels: labels,
			Buckets:     prometheus.ExponentialBuckets(100, 10, 8),
		},
		[]string{"method", "scope", "route"},
	)

	httpResponseSize = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Namespace:   cfg.Namespace,
			Name:        cfg.Metrics.HTTP.ResponseSize,
			Help:        "HTTP response size in bytes",
			ConstLabels: labels,
			Buckets:     prometheus.ExponentialBuckets(100, 10, 8),
		},
		[]string{"method", "scope", "route"},
	)
}

// PrometheusMiddleware 返回 Gin 中间件，记录每个请求的四维指标。
// 监控总开关关闭时返回 no-op 中间件，不产生任何运行时开销。
func PrometheusMiddleware() gin.HandlerFunc {
	if !monitoringEnabled() {
		return func(c *gin.Context) { c.Next() }
	}

	includePath := app.GetConfig().Prometheus.Metrics.HTTP.IncludePath

	return func(c *gin.Context) {
		start := time.Now()

		method := c.Request.Method
		path := c.Request.URL.Path
		scope := detectScope(path)
		route := detectRoute(c, includePath, scope)

		if httpRequestSize != nil && c.Request.ContentLength > 0 {
			httpRequestSize.WithLabelValues(method, scope, route).Observe(float64(c.Request.ContentLength))
		}

		c.Next()

		status := strconv.Itoa(c.Writer.Status())

		if httpRequestsTotal != nil {
			httpRequestsTotal.WithLabelValues(method, scope, route, status).Inc()
		}
		if httpRequestDuration != nil {
			httpRequestDuration.WithLabelValues(method, scope, route).Observe(time.Since(start).Seconds())
		}
		if httpResponseSize != nil && c.Writer.Size() > 0 {
			httpResponseSize.WithLabelValues(method, scope, route).Observe(float64(c.Writer.Size()))
		}
	}
}

// detectScope 根据请求路径归类到粗粒度 scope 标签。
// 使用 IsAPIPath 判断 API 路径，避免 strings.HasPrefix("/api") 的 /apiary 漏洞。
func detectScope(path string) string {
	p := strings.ToLower(path)
	switch {
	case IsAPIPath(p):
		return "api"
	case strings.HasPrefix(p, "/static/"), strings.HasPrefix(p, "/public/"), strings.HasPrefix(p, "/uploads/"):
		return "static"
	case p == "/metrics":
		return "metrics"
	default:
		return "web"
	}
}

// detectRoute 用于 route 标签的取值：
//   - includePath=false 时退化为 scope，标签基数极低
//   - includePath=true 时使用 Gin 的路由模板（如 /users/:id），避免因路径参数爆炸
//   - 无匹配路由（如 NoRoute）统一标记为 "unmatched"
func detectRoute(c *gin.Context, includePath bool, scope string) string {
	if !includePath {
		return scope
	}
	rp := c.FullPath()
	if rp == "" {
		return "unmatched"
	}
	return rp
}
