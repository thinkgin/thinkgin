package app

// setDefaultConfig 为缺省字段填充安全的默认值。
// 仅在"值为零值"时写入，不覆盖 YAML / 环境变量已设置的内容。
func setDefaultConfig() {
	if Config == nil {
		Config = &GlobalConfig{}
	}

	// 应用元信息
	if Config.App.Name == "" {
		Config.App.Name = "ThinkGin"
	}
	if Config.App.Version == "" {
		Config.App.Version = "3.6.0"
	}

	// HTTP 服务
	if Config.Server.HTTP.Host == "" {
		Config.Server.HTTP.Host = "0.0.0.0"
	}
	if Config.Server.HTTP.Port == 0 {
		Config.Server.HTTP.Port = 8000
	}
	if Config.Server.HTTP.ReadTimeout == 0 {
		Config.Server.HTTP.ReadTimeout = 60
	}
	if Config.Server.HTTP.WriteTimeout == 0 {
		Config.Server.HTTP.WriteTimeout = 60
	}
	if Config.Server.HTTP.IdleTimeout == 0 {
		Config.Server.HTTP.IdleTimeout = 120
	}
	if Config.Server.HTTP.MaxHeaderBytes == 0 {
		Config.Server.HTTP.MaxHeaderBytes = 1 << 20 // 1 MiB
	}

	// 日志
	if Config.Log.Default.Level == "" {
		Config.Log.Default.Level = "info"
	}
	if Config.Log.Default.Format == "" {
		Config.Log.Default.Format = "json"
	}
	if Config.Log.File.Path == "" {
		Config.Log.File.Path = "runtime/log"
	}
	if Config.Log.File.Filename == "" {
		Config.Log.File.Filename = "system"
	}
	if Config.Log.File.MaxAge == 0 {
		Config.Log.File.MaxAge = 30
	}
	if Config.Log.File.RotationTime == 0 {
		Config.Log.File.RotationTime = 24
	}

	// Prometheus
	if Config.Prometheus.ServiceName == "" {
		Config.Prometheus.ServiceName = "thinkgin"
	}
	if Config.Prometheus.Path == "" {
		Config.Prometheus.Path = "/metrics"
	}
	if Config.Prometheus.Namespace == "" {
		Config.Prometheus.Namespace = "app"
	}
	if Config.Prometheus.ScrapeInterval == 0 {
		Config.Prometheus.ScrapeInterval = 15
	}
}
