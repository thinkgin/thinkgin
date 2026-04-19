// 本文件暴露 /metrics 端点。支持可选的 Basic Auth，凭据来自 prometheus.auth 配置。
package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus/promhttp"

	"thinkgin/app"
)

// PrometheusHandler 返回 /metrics 端点的 Gin Handler。
// 监控未启用时返回 404 处理器；启用且配置了 Basic Auth 时，先过认证再交给 promhttp。
func PrometheusHandler() gin.HandlerFunc {
	if !monitoringEnabled() {
		return func(c *gin.Context) {
			c.JSON(http.StatusNotFound, gin.H{"error": "prometheus monitoring is disabled"})
		}
	}

	cfg := app.GetConfig().Prometheus
	metrics := gin.WrapH(promhttp.Handler())

	if !cfg.Auth.Enabled {
		return metrics
	}

	// 启用 Basic Auth：credentials 固定从配置读取，不支持热更新。
	auth := gin.BasicAuth(gin.Accounts{
		cfg.Auth.Username: cfg.Auth.Password,
	})
	return func(c *gin.Context) {
		auth(c)
		if c.IsAborted() {
			return
		}
		metrics(c)
	}
}
