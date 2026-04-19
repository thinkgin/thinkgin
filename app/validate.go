package app

import (
	"fmt"
	"strings"

	"github.com/sirupsen/logrus"
)

// validateConfig 对已加载的配置做语义校验，非法值回退到安全默认值。
// 为了脚手架的"零配置可跑"目标，校验失败不报错，只打印到标准输出。
func validateConfig() {
	if Config == nil {
		return
	}

	validateServerMode()
	validateHTTPHostPort()
	validateLogLevel()
	validateLogFormat()
	validateLogPath()
}

// validateServerMode 将 mode 归一化为 gin 支持的三种值，无效值退化为 debug。
func validateServerMode() {
	mode := strings.ToLower(strings.TrimSpace(Config.Server.Mode))
	switch mode {
	case "":
		Config.Server.Mode = "debug"
	case "debug", "test", "release":
		Config.Server.Mode = mode
	default:
		fmt.Printf("[config] 无效的 server.mode: %q，已回退为 debug\n", Config.Server.Mode)
		Config.Server.Mode = "debug"
	}
}

// validateHTTPHostPort 校验 HTTP 监听地址与端口的合法性。
func validateHTTPHostPort() {
	if Config.Server.HTTP.Host == "" {
		Config.Server.HTTP.Host = "0.0.0.0"
	}
	if Config.Server.HTTP.Port <= 0 || Config.Server.HTTP.Port > 65535 {
		fmt.Printf("[config] 无效的 server.http.port: %d，已回退为 8000\n", Config.Server.HTTP.Port)
		Config.Server.HTTP.Port = 8000
	}
}

// validateLogLevel 确保日志级别能被 logrus 识别。
func validateLogLevel() {
	level := strings.ToLower(strings.TrimSpace(Config.Log.Default.Level))
	if _, err := logrus.ParseLevel(level); err != nil {
		fmt.Printf("[config] 无效的 log.default.level: %q，已回退为 info\n", Config.Log.Default.Level)
		Config.Log.Default.Level = "info"
	}
}

// validateLogFormat 只接受 json / text 两种格式。
func validateLogFormat() {
	format := strings.ToLower(strings.TrimSpace(Config.Log.Default.Format))
	switch format {
	case "":
		Config.Log.Default.Format = "json"
	case "json", "text":
		Config.Log.Default.Format = format
	default:
		fmt.Printf("[config] 无效的 log.default.format: %q，已回退为 json\n", Config.Log.Default.Format)
		Config.Log.Default.Format = "json"
	}
}

// validateLogPath 保证落盘路径与文件名不为空。
func validateLogPath() {
	if Config.Log.File.Path == "" {
		Config.Log.File.Path = "runtime/log"
	}
	if Config.Log.File.Filename == "" {
		Config.Log.File.Filename = "system"
	}
}
