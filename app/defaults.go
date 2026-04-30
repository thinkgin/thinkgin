package app

// setDefaultConfig 为全局 Config 的缺省字段填充安全的默认值。
// 仅在"值为零值"时写入，不覆盖 YAML / 环境变量已设置的内容。
func setDefaultConfig() {
	if Config == nil {
		Config = &GlobalConfig{}
	}
	setDefaultsOn(Config)
}

// setDefaultsOn 为指定 cfg 的缺省字段填充安全的默认值。
// 热更新路径使用此函数操作新配置对象，避免修改全局变量。
func setDefaultsOn(cfg *GlobalConfig) {
	// 应用元信息
	if cfg.App.Name == "" {
		cfg.App.Name = "ThinkGin"
	}
	if cfg.App.Version == "" {
		cfg.App.Version = "3.8.0"
	}

	// HTTP 服务
	if cfg.Server.HTTP.Host == "" {
		cfg.Server.HTTP.Host = "0.0.0.0"
	}
	if cfg.Server.HTTP.Port == 0 {
		cfg.Server.HTTP.Port = 8000
	}
	if cfg.Server.HTTP.ReadTimeout == 0 {
		cfg.Server.HTTP.ReadTimeout = 60
	}
	if cfg.Server.HTTP.WriteTimeout == 0 {
		cfg.Server.HTTP.WriteTimeout = 60
	}
	if cfg.Server.HTTP.IdleTimeout == 0 {
		cfg.Server.HTTP.IdleTimeout = 120
	}
	if cfg.Server.HTTP.MaxHeaderBytes == 0 {
		cfg.Server.HTTP.MaxHeaderBytes = 1 << 20 // 1 MiB
	}

	// 日志
	if cfg.Log.Default.Level == "" {
		cfg.Log.Default.Level = "info"
	}
	if cfg.Log.Default.Format == "" {
		cfg.Log.Default.Format = "json"
	}
	if cfg.Log.File.Path == "" {
		cfg.Log.File.Path = "runtime/log"
	}
	if cfg.Log.File.Filename == "" {
		cfg.Log.File.Filename = "system"
	}
	if cfg.Log.File.MaxAge == 0 {
		cfg.Log.File.MaxAge = 30
	}
	if cfg.Log.File.RotationTime == 0 {
		cfg.Log.File.RotationTime = 24
	}

	// Prometheus
	if cfg.Prometheus.ServiceName == "" {
		cfg.Prometheus.ServiceName = "thinkgin"
	}
	if cfg.Prometheus.Path == "" {
		cfg.Prometheus.Path = "/metrics"
	}
	if cfg.Prometheus.Namespace == "" {
		cfg.Prometheus.Namespace = "app"
	}
	if cfg.Prometheus.ScrapeInterval == 0 {
		cfg.Prometheus.ScrapeInterval = 15
	}
}
