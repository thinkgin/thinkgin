// 本文件暴露 /metrics 端点。支持可选的 Basic Auth，凭据来自 prometheus.auth 配置。
package middleware

import (
	"fmt"
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
		// release 模式下暴露无鉴权的 /metrics 存在信息泄漏风险，打印启动告警。
		if app.GetConfig().Server.Mode == "release" {
			fmt.Println("[prometheus] 警告：/metrics 端点已暴露且未启用鉴权。" +
				"生产环境建议开启 prometheus.auth 或通过反向代理/网络策略限制访问。")
		}
		return metrics
	}

	// 启用 Basic Auth 但凭据为空属于误配置：拒绝以"看似安全实则放行/锁死"的方式启动。
	if cfg.Auth.Username == "" || cfg.Auth.Password == "" {
		fmt.Println("[prometheus] 警告：prometheus.auth.enabled=true 但 username/password 为空，" +
			"鉴权无法生效，/metrics 端点已被禁用以防误暴露。")
		return func(c *gin.Context) {
			c.JSON(http.StatusServiceUnavailable, gin.H{"error": "metrics auth misconfigured"})
		}
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
