package middleware

import (
	"runtime"
	"strconv"
	"thinkgin/app"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var (
	// HTTP请求指标
	httpRequestsTotal   *prometheus.CounterVec
	httpRequestDuration *prometheus.HistogramVec
	httpRequestSize     *prometheus.HistogramVec
	httpResponseSize    *prometheus.HistogramVec

	// 系统资源指标
	systemCPUUsage    prometheus.Gauge
	systemMemoryUsage prometheus.Gauge
	processStartTime  prometheus.Gauge

	// 业务指标注册器
	businessCounters   map[string]*prometheus.CounterVec
	businessHistograms map[string]*prometheus.HistogramVec
	businessGauges     map[string]*prometheus.GaugeVec
)

// 初始化Prometheus指标
func InitPrometheusMetrics() {
	appConfig := app.GetConfig().App
	config := app.GetConfig().Prometheus

	// 检查应用配置中的总开关
	if !appConfig.Monitoring.PrometheusEnabled {
		return
	}

	if !config.Enabled {
		return
	}

	// 初始化业务指标映射
	businessCounters = make(map[string]*prometheus.CounterVec)
	businessHistograms = make(map[string]*prometheus.HistogramVec)
	businessGauges = make(map[string]*prometheus.GaugeVec)

	// 构建标签
	constLabels := prometheus.Labels{}
	for k, v := range config.Labels {
		constLabels[k] = v
	}

	// HTTP指标
	if config.Metrics.HTTP.Enabled {
		httpRequestsTotal = promauto.NewCounterVec(
			prometheus.CounterOpts{
				Namespace:   config.Namespace,
				Name:        config.Metrics.HTTP.RequestsTotal,
				Help:        "Total number of HTTP requests",
				ConstLabels: constLabels,
			},
			[]string{"method", "path", "status"},
		)

		httpRequestDuration = promauto.NewHistogramVec(
			prometheus.HistogramOpts{
				Namespace:   config.Namespace,
				Name:        config.Metrics.HTTP.RequestDuration,
				Help:        "HTTP request duration in seconds",
				ConstLabels: constLabels,
				Buckets:     prometheus.DefBuckets,
			},
			[]string{"method", "path"},
		)

		httpRequestSize = promauto.NewHistogramVec(
			prometheus.HistogramOpts{
				Namespace:   config.Namespace,
				Name:        config.Metrics.HTTP.RequestSize,
				Help:        "HTTP request size in bytes",
				ConstLabels: constLabels,
				Buckets:     prometheus.ExponentialBuckets(100, 10, 8),
			},
			[]string{"method", "path"},
		)

		httpResponseSize = promauto.NewHistogramVec(
			prometheus.HistogramOpts{
				Namespace:   config.Namespace,
				Name:        config.Metrics.HTTP.ResponseSize,
				Help:        "HTTP response size in bytes",
				ConstLabels: constLabels,
				Buckets:     prometheus.ExponentialBuckets(100, 10, 8),
			},
			[]string{"method", "path"},
		)
	}

	// 系统资源指标
	if config.Metrics.System.Enabled {
		systemCPUUsage = promauto.NewGauge(
			prometheus.GaugeOpts{
				Namespace:   config.Namespace,
				Name:        config.Metrics.System.CPUUsage,
				Help:        "Current CPU usage percentage",
				ConstLabels: constLabels,
			},
		)

		systemMemoryUsage = promauto.NewGauge(
			prometheus.GaugeOpts{
				Namespace:   config.Namespace,
				Name:        config.Metrics.System.MemoryUsage,
				Help:        "Current memory usage in bytes",
				ConstLabels: constLabels,
			},
		)

		processStartTime = promauto.NewGauge(
			prometheus.GaugeOpts{
				Namespace:   config.Namespace,
				Name:        config.Metrics.System.ProcessStartTime,
				Help:        "Process start time in unix timestamp",
				ConstLabels: constLabels,
			},
		)

		// 设置进程启动时间
		processStartTime.SetToCurrentTime()

		// 启动系统资源监控
		go monitorSystemResources()
	}
}

// Prometheus HTTP中间件
func PrometheusMiddleware() gin.HandlerFunc {
	appConfig := app.GetConfig().App
	config := app.GetConfig().Prometheus

	// 检查应用配置中的总开关
	if !appConfig.Monitoring.PrometheusEnabled {
		return func(c *gin.Context) {
			c.Next()
		}
	}

	if !config.Enabled {
		return func(c *gin.Context) {
			c.Next()
		}
	}

	return func(c *gin.Context) {
		start := time.Now()

		// 获取请求信息
		method := c.Request.Method
		path := c.FullPath()

		// 如果不包含路径标签，使用通用路径
		if !config.Metrics.HTTP.IncludePath {
			path = "api"
		}

		// 记录请求大小
		if httpRequestSize != nil && c.Request.ContentLength > 0 {
			httpRequestSize.WithLabelValues(method, path).Observe(float64(c.Request.ContentLength))
		}

		// 处理请求
		c.Next()

		// 记录指标
		duration := time.Since(start)
		status := strconv.Itoa(c.Writer.Status())

		if httpRequestsTotal != nil {
			httpRequestsTotal.WithLabelValues(method, path, status).Inc()
		}

		if httpRequestDuration != nil {
			httpRequestDuration.WithLabelValues(method, path).Observe(duration.Seconds())
		}

		if httpResponseSize != nil && c.Writer.Size() > 0 {
			httpResponseSize.WithLabelValues(method, path).Observe(float64(c.Writer.Size()))
		}
	}
}

// 系统资源监控协程
func monitorSystemResources() {
	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			// 更新内存使用情况
			var m runtime.MemStats
			runtime.ReadMemStats(&m)

			if systemMemoryUsage != nil {
				systemMemoryUsage.Set(float64(m.Alloc))
			}

			// CPU使用率监控需要更复杂的实现，这里简化处理
			// 可以使用第三方库如 gopsutil 来获取更准确的CPU使用率
		}
	}
}

// Prometheus指标暴露端点
func PrometheusHandler() gin.HandlerFunc {
	appConfig := app.GetConfig().App
	config := app.GetConfig().Prometheus

	// 检查应用配置中的总开关
	if !appConfig.Monitoring.PrometheusEnabled {
		return func(c *gin.Context) {
			c.JSON(404, gin.H{"error": "Prometheus monitoring is disabled"})
		}
	}

	// 如果启用了认证
	if config.Auth.Enabled {
		// 创建基础认证中间件
		authMiddleware := gin.BasicAuth(gin.Accounts{
			config.Auth.Username: config.Auth.Password,
		})

		// 返回认证后的处理函数
		return func(c *gin.Context) {
			authMiddleware(c)
			if c.IsAborted() {
				return
			}
			gin.WrapH(promhttp.Handler())(c)
		}
	}

	return gin.WrapH(promhttp.Handler())
}

// 业务计数器
func BusinessCounter(name string, labels []string) *prometheus.CounterVec {
	appConfig := app.GetConfig().App
	config := app.GetConfig().Prometheus

	// 检查应用配置中的总开关
	if !appConfig.Monitoring.PrometheusEnabled {
		return nil
	}

	if !config.Enabled || !config.Metrics.Business.Enabled {
		return nil
	}

	fullName := config.Metrics.Business.CounterPrefix + "_" + name

	if counter, exists := businessCounters[fullName]; exists {
		return counter
	}

	constLabels := prometheus.Labels{}
	for k, v := range config.Labels {
		constLabels[k] = v
	}

	counter := promauto.NewCounterVec(
		prometheus.CounterOpts{
			Namespace:   config.Namespace,
			Name:        fullName,
			Help:        "Business counter for " + name,
			ConstLabels: constLabels,
		},
		labels,
	)

	businessCounters[fullName] = counter
	return counter
}

// 业务直方图
func BusinessHistogram(name string, labels []string, buckets []float64) *prometheus.HistogramVec {
	appConfig := app.GetConfig().App
	config := app.GetConfig().Prometheus

	// 检查应用配置中的总开关
	if !appConfig.Monitoring.PrometheusEnabled {
		return nil
	}

	if !config.Enabled || !config.Metrics.Business.Enabled {
		return nil
	}

	fullName := config.Metrics.Business.HistogramPrefix + "_" + name

	if histogram, exists := businessHistograms[fullName]; exists {
		return histogram
	}

	if buckets == nil {
		buckets = prometheus.DefBuckets
	}

	constLabels := prometheus.Labels{}
	for k, v := range config.Labels {
		constLabels[k] = v
	}

	histogram := promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Namespace:   config.Namespace,
			Name:        fullName,
			Help:        "Business histogram for " + name,
			ConstLabels: constLabels,
			Buckets:     buckets,
		},
		labels,
	)

	businessHistograms[fullName] = histogram
	return histogram
}

// 业务仪表盘
func BusinessGauge(name string, labels []string) *prometheus.GaugeVec {
	appConfig := app.GetConfig().App
	config := app.GetConfig().Prometheus

	// 检查应用配置中的总开关
	if !appConfig.Monitoring.PrometheusEnabled {
		return nil
	}

	if !config.Enabled || !config.Metrics.Business.Enabled {
		return nil
	}

	fullName := config.Metrics.Business.GaugePrefix + "_" + name

	if gauge, exists := businessGauges[fullName]; exists {
		return gauge
	}

	constLabels := prometheus.Labels{}
	for k, v := range config.Labels {
		constLabels[k] = v
	}

	gauge := promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace:   config.Namespace,
			Name:        fullName,
			Help:        "Business gauge for " + name,
			ConstLabels: constLabels,
		},
		labels,
	)

	businessGauges[fullName] = gauge
	return gauge
}
