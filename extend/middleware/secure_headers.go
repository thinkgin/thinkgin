// 本文件提供 HTTP 安全响应头中间件。
//
// 默认启用的头部：
//   - X-Content-Type-Options: nosniff
//   - X-Frame-Options: DENY
//   - X-XSS-Protection: 1; mode=block
//   - Referrer-Policy: strict-origin-when-cross-origin
//   - Content-Security-Policy: default-src 'self'
//
// 各头部可通过 middleware.config.secure_headers 按需覆盖或禁用。
package middleware

import (
	"thinkgin/app"

	"github.com/gin-gonic/gin"
)

// defaultSecureHeaders 是未配置时使用的安全头。
var defaultSecureHeaders = map[string]string{
	"X-Content-Type-Options": "nosniff",
	"X-Frame-Options":        "DENY",
	"X-XSS-Protection":       "1; mode=block",
	"Referrer-Policy":         "strict-origin-when-cross-origin",
}

// SecureHeaders 返回安全响应头中间件。
// 配置路径：middleware.config.secure_headers（map[string]string）。
// 值为空字符串时跳过该头，便于精确控制。
func SecureHeaders() gin.HandlerFunc {
	headers := getSecureHeadersConfig()

	return func(c *gin.Context) {
		for k, v := range headers {
			if v != "" {
				c.Header(k, v)
			}
		}
		c.Next()
	}
}

func getSecureHeadersConfig() map[string]string {
	out := make(map[string]string, len(defaultSecureHeaders))
	for k, v := range defaultSecureHeaders {
		out[k] = v
	}

	cfg := app.GetConfig()
	if cfg == nil || cfg.Middleware.Config == nil {
		return out
	}

	raw, ok := cfg.Middleware.Config["secure_headers"]
	if !ok {
		return out
	}

	m, ok := asStringInterfaceMap(raw)
	if !ok {
		return out
	}

	for k, v := range m {
		if s, ok := v.(string); ok {
			out[k] = s
		}
	}
	return out
}
