// 本文件提供业务自定义指标的注册与复用能力。
//
// 设计要点：
//   - 三种指标（Counter/Histogram/Gauge）按 name 去重，同名返回已注册实例。
//   - 通过泛型 getOrRegister 消除三个函数的重复样板（约 70 行）。
//   - 使用读写锁 RWMutex：命中缓存走快路径（RLock），注册走慢路径（Lock）。
package middleware

import (
	"sync"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"

	"thinkgin/app"
)

var (
	businessMu         sync.RWMutex
	businessCounters   = map[string]*prometheus.CounterVec{}
	businessHistograms = map[string]*prometheus.HistogramVec{}
	businessGauges     = map[string]*prometheus.GaugeVec{}
)

// BusinessCounter 返回一个 Counter 向量，同名指标共用实例。
// 业务未启用或监控总开关关闭时返回 nil，调用方需判空。
func BusinessCounter(name string, labels []string) *prometheus.CounterVec {
	if !businessEnabled() {
		return nil
	}
	cfg := app.GetConfig().Prometheus
	fullName := cfg.Metrics.Business.CounterPrefix + "_" + name

	return getOrRegister(businessCounters, fullName, func() *prometheus.CounterVec {
		return promauto.NewCounterVec(prometheus.CounterOpts{
			Namespace:   cfg.Namespace,
			Name:        fullName,
			Help:        "Business counter for " + name,
			ConstLabels: constLabels(cfg.Labels),
		}, labels)
	})
}

// BusinessHistogram 返回一个 Histogram 向量。buckets 为 nil 时使用 Prometheus 默认桶。
func BusinessHistogram(name string, labels []string, buckets []float64) *prometheus.HistogramVec {
	if !businessEnabled() {
		return nil
	}
	cfg := app.GetConfig().Prometheus
	fullName := cfg.Metrics.Business.HistogramPrefix + "_" + name
	if buckets == nil {
		buckets = prometheus.DefBuckets
	}

	return getOrRegister(businessHistograms, fullName, func() *prometheus.HistogramVec {
		return promauto.NewHistogramVec(prometheus.HistogramOpts{
			Namespace:   cfg.Namespace,
			Name:        fullName,
			Help:        "Business histogram for " + name,
			ConstLabels: constLabels(cfg.Labels),
			Buckets:     buckets,
		}, labels)
	})
}

// BusinessGauge 返回一个 Gauge 向量。
func BusinessGauge(name string, labels []string) *prometheus.GaugeVec {
	if !businessEnabled() {
		return nil
	}
	cfg := app.GetConfig().Prometheus
	fullName := cfg.Metrics.Business.GaugePrefix + "_" + name

	return getOrRegister(businessGauges, fullName, func() *prometheus.GaugeVec {
		return promauto.NewGaugeVec(prometheus.GaugeOpts{
			Namespace:   cfg.Namespace,
			Name:        fullName,
			Help:        "Business gauge for " + name,
			ConstLabels: constLabels(cfg.Labels),
		}, labels)
	})
}

// getOrRegister 用双检锁实现"同名指标只注册一次"。
// 快路径 RLock 命中缓存直接返回；慢路径 Lock 后再次检查，避免并发重复注册。
// register 回调只在未命中时执行，包含对外部注册表（promauto）的调用。
func getOrRegister[T any](cache map[string]T, name string, register func() T) T {
	businessMu.RLock()
	if v, ok := cache[name]; ok {
		businessMu.RUnlock()
		return v
	}
	businessMu.RUnlock()

	businessMu.Lock()
	defer businessMu.Unlock()
	if v, ok := cache[name]; ok {
		return v
	}
	v := register()
	cache[name] = v
	return v
}
