package app

import (
	"os"
	"strconv"
	"strings"
)

// envPrefix 是所有 ThinkGin 环境变量的公共前缀，避免与其他软件冲突。
const envPrefix = "THINKGIN_"

// applyEnvOverrides 用环境变量覆盖全局 Config，供容器化部署使用。
func applyEnvOverrides() {
	if Config == nil {
		return
	}
	applyEnvOverridesOn(Config)
}

// applyEnvOverridesOn 用环境变量覆盖指定 cfg 的配置。
// 热更新路径使用此函数操作新配置对象，避免修改全局变量。
// 只覆盖显式声明过的字段，未设置的环境变量保留配置文件原值。
func applyEnvOverridesOn(cfg *GlobalConfig) {
	// Server 相关
	if v := os.Getenv(envPrefix + "SERVER_MODE"); v != "" {
		cfg.Server.Mode = v
	}
	if v := os.Getenv(envPrefix + "SERVER_HTTP_HOST"); v != "" {
		cfg.Server.HTTP.Host = v
	}
	if v := getenvInt(envPrefix + "SERVER_HTTP_PORT"); v != nil {
		cfg.Server.HTTP.Port = *v
	}
	if v := getenvInt(envPrefix + "SERVER_HTTP_READ_TIMEOUT"); v != nil {
		cfg.Server.HTTP.ReadTimeout = *v
	}
	if v := getenvInt(envPrefix + "SERVER_HTTP_WRITE_TIMEOUT"); v != nil {
		cfg.Server.HTTP.WriteTimeout = *v
	}
	if v := getenvInt(envPrefix + "SERVER_HTTP_IDLE_TIMEOUT"); v != nil {
		cfg.Server.HTTP.IdleTimeout = *v
	}
	if v := getenvInt(envPrefix + "SERVER_HTTP_MAX_HEADER_BYTES"); v != nil {
		cfg.Server.HTTP.MaxHeaderBytes = *v
	}

	// App / 日志
	if v := getenvBool(envPrefix + "APP_DEBUG"); v != nil {
		cfg.App.Debug = *v
	}
	if v := os.Getenv(envPrefix + "LOG_LEVEL"); v != "" {
		cfg.Log.Default.Level = v
	}
	if v := os.Getenv(envPrefix + "LOG_FORMAT"); v != "" {
		cfg.Log.Default.Format = v
	}
	if v := os.Getenv(envPrefix + "LOG_PATH"); v != "" {
		cfg.Log.File.Path = v
	}
	if v := os.Getenv(envPrefix + "LOG_FILENAME"); v != "" {
		cfg.Log.File.Filename = v
	}

	// JWT（敏感信息，强烈建议通过环境变量注入而非写入配置文件）
	if v := os.Getenv(envPrefix + "APP_JWT_SECRET"); v != "" {
		cfg.App.JWT.Secret = v
	}
	if v := getenvInt(envPrefix + "APP_JWT_EXPIRE"); v != nil {
		cfg.App.JWT.Expire = *v
	}

	// Prometheus 总开关与路径
	if v := getenvBool(envPrefix + "APP_PROMETHEUS_ENABLED"); v != nil {
		cfg.App.Monitoring.PrometheusEnabled = *v
	}
	if v := getenvBool(envPrefix + "PROMETHEUS_ENABLED"); v != nil {
		cfg.Prometheus.Enabled = *v
	}
	if v := os.Getenv(envPrefix + "PROMETHEUS_PATH"); v != "" {
		cfg.Prometheus.Path = v
	}
}

// getenvInt 读取环境变量为 *int。空或非数字时返回 nil。
func getenvInt(key string) *int {
	v := strings.TrimSpace(os.Getenv(key))
	if v == "" {
		return nil
	}
	n, err := strconv.Atoi(v)
	if err != nil {
		return nil
	}
	return &n
}

// getenvBool 读取环境变量为 *bool。空或非法值时返回 nil。
// 识别 strconv.ParseBool 支持的一切写法：1/0/true/false/TRUE/FALSE/T/F…
func getenvBool(key string) *bool {
	v := strings.TrimSpace(os.Getenv(key))
	if v == "" {
		return nil
	}
	b, err := strconv.ParseBool(v)
	if err != nil {
		return nil
	}
	return &b
}
