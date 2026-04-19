// 本文件暴露进程级运行时指标（Alloc 字节数、进程启动时间）。
//
// 关于 CPU 指标：精确的 CPU 使用率需要 gopsutil 等外部库，未引入前不暴露假值。
// 对 Grafana 而言，缺指标比恒零曲线更容易被发现，也不会误导 oncall 判断。
package middleware

import (
	"runtime"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"

	"thinkgin/app"
)

var (
	systemMemoryUsage prometheus.Gauge
	processStartTime  prometheus.Gauge
)

// systemSampleInterval 是采样周期。10 秒足以捕获内存抖动，又不至于造成可观开销。
const systemSampleInterval = 10 * time.Second

// initSystemMetrics 注册内存 / 启动时间指标，并启动后台采样 goroutine。
// 该 goroutine 生命周期与进程绑定，采用无出口循环。
//
// NOTE: 未来若 ThinkGin 支持"嵌入式"使用（非 main 入口），应把该 goroutine
// 的生命周期改为由 App 管理（接收 ctx.Done）。
func initSystemMetrics(cfg app.PrometheusConfig, labels prometheus.Labels) {
	systemMemoryUsage = promauto.NewGauge(
		prometheus.GaugeOpts{
			Namespace:   cfg.Namespace,
			Name:        cfg.Metrics.System.MemoryUsage,
			Help:        "Current Go runtime Alloc in bytes",
			ConstLabels: labels,
		},
	)

	processStartTime = promauto.NewGauge(
		prometheus.GaugeOpts{
			Namespace:   cfg.Namespace,
			Name:        cfg.Metrics.System.ProcessStartTime,
			Help:        "Process start time in unix timestamp",
			ConstLabels: labels,
		},
	)
	processStartTime.SetToCurrentTime()

	go sampleMemoryLoop()
}

// sampleMemoryLoop 周期性将 runtime.MemStats.Alloc 写入 systemMemoryUsage。
// 与进程共存；`go vet` 不会告警的原因是 ticker 被 defer 回收。
func sampleMemoryLoop() {
	ticker := time.NewTicker(systemSampleInterval)
	defer ticker.Stop()

	for range ticker.C {
		var m runtime.MemStats
		runtime.ReadMemStats(&m)
		if systemMemoryUsage != nil {
			systemMemoryUsage.Set(float64(m.Alloc))
		}
	}
}
