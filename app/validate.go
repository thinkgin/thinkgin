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

	validateJWTSecret(cfg)
}

// knownWeakJWTSecrets 是历史版本中泄漏过或属于明显弱口令的 JWT 密钥黑名单。
// 命中时强提示用户更换，避免拿示例密钥直接上生产。
var knownWeakJWTSecrets = map[string]struct{}{
	"23347$040412": {}, // v3.11.0 之前 config/app.yaml 中的硬编码示例密钥
	"secret":       {},
	"changeme":     {},
	"your-secret":  {},
}

// minJWTSecretLen 是推荐的 JWT 密钥最小长度（HS256 建议 >= 32 字节）。
const minJWTSecretLen = 32

// validateJWTSecret 校验 JWT 密钥的安全性，仅打印告警、不修改配置也不中断启动。
// 设计取舍：JWTIssue/JWTAuth 在密钥为空时本就拒绝工作，这里只做"早发现"的提示。
func validateJWTSecret(cfg *GlobalConfig) {
	secret := strings.TrimSpace(cfg.App.JWT.Secret)
	if secret == "" {
		// 空密钥是安全的（拒绝签发），仅在需要 JWT 时提示如何注入。
		fmt.Println("[config] 提示：app.jwt.secret 为空，JWT 签发/校验已禁用。" +
			"如需启用，请通过环境变量 THINKGIN_APP_JWT_SECRET 注入密钥。")
		return
	}
	if _, weak := knownWeakJWTSecrets[secret]; weak {
		fmt.Println("[config] 警告：检测到弱/示例 JWT 密钥，存在 token 伪造风险！" +
			"请立即更换为高强度随机值，并通过环境变量 THINKGIN_APP_JWT_SECRET 注入。")
		return
	}
	if len(secret) < minJWTSecretLen {
		fmt.Printf("[config] 警告：app.jwt.secret 长度为 %d，建议 >= %d 字节以保证 HS256 安全强度。\n",
			len(secret), minJWTSecretLen)
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
