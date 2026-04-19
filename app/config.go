package app

// Config 是进程级的全局配置单例。
// 通过 init 或 Bootstrap 填充；业务代码只应通过 GetConfig/GetXxxConfig 访问。
var Config *GlobalConfig

// GetConfig 返回全局配置。如果尚未初始化，返回 nil。
// 调用方应确保在 main 或 init 阶段完成 Bootstrap 后再使用。
func GetConfig() *GlobalConfig {
	return Config
}

// GetAppConfig 返回 app 段快捷引用。
func GetAppConfig() *AppConfig {
	if Config == nil {
		return nil
	}
	return &Config.App
}

// GetServerConfig 返回 server 段快捷引用。
func GetServerConfig() *ServerConfig {
	if Config == nil {
		return nil
	}
	return &Config.Server
}

// GetDatabaseConfig 返回 database 段快捷引用。
func GetDatabaseConfig() *DatabaseConfig {
	if Config == nil {
		return nil
	}
	return &Config.Database
}

// GetLogConfig 返回 log 段快捷引用。
func GetLogConfig() *LogConfig {
	if Config == nil {
		return nil
	}
	return &Config.Log
}

// GetPrometheusConfig 返回 prometheus 段快捷引用。
func GetPrometheusConfig() *PrometheusConfig {
	if Config == nil {
		return nil
	}
	return &Config.Prometheus
}
