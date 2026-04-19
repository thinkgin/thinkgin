package app

import (
	"os"
	"strconv"
	"strings"
)

// envPrefix 是所有 ThinkGin 环境变量的公共前缀，避免与其他软件冲突。
const envPrefix = "THINKGIN_"

// applyEnvOverrides 用环境变量覆盖已加载的配置，供容器化部署使用。
// 只覆盖显式声明过的字段，未设置的环境变量保留配置文件原值。
func applyEnvOverrides() {
	if Config == nil {
		return
	}

	// Server 相关
	if v := os.Getenv(envPrefix + "SERVER_MODE"); v != "" {
		Config.Server.Mode = v
	}
	if v := os.Getenv(envPrefix + "SERVER_HTTP_HOST"); v != "" {
		Config.Server.HTTP.Host = v
	}
	if v := getenvInt(envPrefix + "SERVER_HTTP_PORT"); v != nil {
		Config.Server.HTTP.Port = *v
	}
	if v := getenvInt(envPrefix + "SERVER_HTTP_READ_TIMEOUT"); v != nil {
		Config.Server.HTTP.ReadTimeout = *v
	}
	if v := getenvInt(envPrefix + "SERVER_HTTP_WRITE_TIMEOUT"); v != nil {
		Config.Server.HTTP.WriteTimeout = *v
	}
	if v := getenvInt(envPrefix + "SERVER_HTTP_IDLE_TIMEOUT"); v != nil {
		Config.Server.HTTP.IdleTimeout = *v
	}
	if v := getenvInt(envPrefix + "SERVER_HTTP_MAX_HEADER_BYTES"); v != nil {
		Config.Server.HTTP.MaxHeaderBytes = *v
	}

	// App / 日志
	if v := getenvBool(envPrefix + "APP_DEBUG"); v != nil {
		Config.App.Debug = *v
	}
	if v := os.Getenv(envPrefix + "LOG_LEVEL"); v != "" {
		Config.Log.Default.Level = v
	}
	if v := os.Getenv(envPrefix + "LOG_FORMAT"); v != "" {
		Config.Log.Default.Format = v
	}
	if v := os.Getenv(envPrefix + "LOG_PATH"); v != "" {
		Config.Log.File.Path = v
	}
	if v := os.Getenv(envPrefix + "LOG_FILENAME"); v != "" {
		Config.Log.File.Filename = v
	}

	// Prometheus 总开关与路径
	if v := getenvBool(envPrefix + "APP_PROMETHEUS_ENABLED"); v != nil {
		Config.App.Monitoring.PrometheusEnabled = *v
	}
	if v := getenvBool(envPrefix + "PROMETHEUS_ENABLED"); v != nil {
		Config.Prometheus.Enabled = *v
	}
	if v := os.Getenv(envPrefix + "PROMETHEUS_PATH"); v != "" {
		Config.Prometheus.Path = v
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
