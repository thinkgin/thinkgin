package app

import (
	"fmt"
	"strings"
)

// validateConfig 对全局 Config 做语义校验，非法值回退到安全默认值。
func validateConfig() {
	if Config == nil {
		return
	}
	validateConfigOn(Config)
}

// validateConfigOn 对指定 cfg 做语义校验，非法值回退到安全默认值。
// 热更新路径使用此函数操作新配置对象，避免修改全局变量。
// 为了脚手架的"零配置可跑"目标，校验失败不报错，只打印到标准输出。
func validateConfigOn(cfg *GlobalConfig) {
	// server.mode
	mode := strings.ToLower(strings.TrimSpace(cfg.Server.Mode))
	switch mode {
	case "":
		cfg.Server.Mode = "debug"
	case "debug", "test", "release":
		cfg.Server.Mode = mode
	default:
		fmt.Printf("[config] 无效的 server.mode: %q，已回退为 debug\n", cfg.Server.Mode)
		cfg.Server.Mode = "debug"
	}

	// server.http.host / port
	if cfg.Server.HTTP.Host == "" {
		cfg.Server.HTTP.Host = "0.0.0.0"
	}
	if cfg.Server.HTTP.Port <= 0 || cfg.Server.HTTP.Port > 65535 {
		fmt.Printf("[config] 无效的 server.http.port: %d，已回退为 8000\n", cfg.Server.HTTP.Port)
		cfg.Server.HTTP.Port = 8000
	}

	// log.default.level
	level := strings.ToLower(strings.TrimSpace(cfg.Log.Default.Level))
	if !isValidLogLevel(level) {
		fmt.Printf("[config] 无效的 log.default.level: %q，已回退为 info\n", cfg.Log.Default.Level)
		cfg.Log.Default.Level = "info"
	}

	// log.default.format
	format := strings.ToLower(strings.TrimSpace(cfg.Log.Default.Format))
	switch format {
	case "":
		cfg.Log.Default.Format = "json"
	case "json", "text":
		cfg.Log.Default.Format = format
	default:
		fmt.Printf("[config] 无效的 log.default.format: %q，已回退为 json\n", cfg.Log.Default.Format)
		cfg.Log.Default.Format = "json"
	}

	// log.file.path / filename
	if cfg.Log.File.Path == "" {
		cfg.Log.File.Path = "runtime/log"
	}
	if cfg.Log.File.Filename == "" {
		cfg.Log.File.Filename = "system"
	}
}

// isValidLogLevel 校验日志级别名称是否合法（与 logrus/slog 兼容的通用级别集合）。
func isValidLogLevel(level string) bool {
	switch level {
	case "debug", "info", "warn", "warning", "error", "fatal", "panic", "trace":
		return true
	default:
		return false
	}
}
