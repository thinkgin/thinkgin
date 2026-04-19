// 本文件是 Prometheus 监控能力的门面，职责：
//   - 统一维护"总开关"判断（monitoringEnabled）
//   - 协调 http / system / business 三类指标的初始化入口
//   - 提供 constLabels 等共享构造工具
//
// HTTP 指标 → prometheus_http.go
// 系统指标 → prometheus_system.go
// 业务指标 → prometheus_business.go
// /metrics → prometheus_handler.go
package middleware

import (
	"sync"

	"github.com/prometheus/client_golang/prometheus"

	"thinkgin/app"
)

// metricsInitOnce 保证 HTTP / System 指标只注册一次，避免 prometheus duplicate register 崩溃。
var metricsInitOnce sync.Once

// InitPrometheusMetrics 根据配置惰性注册 HTTP / System 指标。
// 业务指标按需注册，不在此处理。
//
// 多次调用是安全的：sync.Once 保证注册只执行一次。
func InitPrometheusMetrics() {
	if !monitoringEnabled() {
		return
	}

	metricsInitOnce.Do(func() {
		cfg := app.GetConfig().Prometheus
		labels := constLabels(cfg.Labels)

		if cfg.Metrics.HTTP.Enabled {
			initHTTPMetrics(cfg, labels)
		}
		if cfg.Metrics.System.Enabled {
			initSystemMetrics(cfg, labels)
		}
	})
}

// monitoringEnabled 判断"监控功能"是否启用。
// 要求 app.monitoring.prometheus_enabled 与 prometheus.enabled 同时为 true。
// 任一为 false 即视为关闭，不产生指标、不注册路由。
func monitoringEnabled() bool {
	cfg := app.GetConfig()
	if cfg == nil {
		return false
	}
	return cfg.App.Monitoring.PrometheusEnabled && cfg.Prometheus.Enabled
}

// businessEnabled 在 monitoringEnabled 基础上再叠加 business 子开关。
func businessEnabled() bool {
	if !monitoringEnabled() {
		return false
	}
	return app.GetConfig().Prometheus.Metrics.Business.Enabled
}

// constLabels 把 map[string]string 复制为 prometheus.Labels。
// 复制而不直接返回是为了避免外部修改 config.Labels 影响已注册的指标。
func constLabels(src map[string]string) prometheus.Labels {
	labels := prometheus.Labels{}
	for k, v := range src {
		labels[k] = v
	}
	return labels
}
