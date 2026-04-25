// 本文件提供请求体大小限制中间件。
//
// 拒绝 Content-Length 超出阈值的请求（返回 413 Payload Too Large），
// 同时对 Body 做 io.LimitReader 包装，防止客户端发送超长数据流。
//
// 使用方式：
//   在 config/middleware.yaml 中添加 "body_limit" 到 global 列表，
//   或手动调用 middleware.BodyLimit("10MB") 注册。
package middleware

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
)

// defaultBodyLimit 默认请求体上限 10MB。
const defaultBodyLimit int64 = 10 << 20

// BodyLimit 返回使用默认上限（10MB）的请求体大小限制中间件。
func BodyLimit() gin.HandlerFunc {
	return BodyLimitWithSize(defaultBodyLimit)
}

// BodyLimitWithSize 返回使用自定义上限（字节数）的中间件。
func BodyLimitWithSize(maxBytes int64) gin.HandlerFunc {
	return func(c *gin.Context) {
		if c.Request.Body == nil {
			c.Next()
			return
		}

		// 检查 Content-Length 头（如果有）
		if c.Request.ContentLength > maxBytes {
			c.AbortWithStatusJSON(http.StatusRequestEntityTooLarge, gin.H{
				"code":    http.StatusRequestEntityTooLarge,
				"message": fmt.Sprintf("request body too large (limit: %s)", formatBytes(maxBytes)),
			})
			return
		}

		// 用 LimitReader 包装，防止没有 Content-Length 的流式请求
		c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, maxBytes)
		c.Next()
	}
}

// ParseBodyLimit 解析人类可读的大小字符串，如 "10MB"、"512KB"、"1GB"。
// 支持 B / KB / MB / GB 单位（大小写不敏感）。
func ParseBodyLimit(s string) (int64, error) {
	s = strings.TrimSpace(strings.ToUpper(s))
	if s == "" {
		return 0, fmt.Errorf("empty body limit string")
	}

	var multiplier int64 = 1
	unit := ""
	numStr := s

	for _, suffix := range []struct {
		name string
		mult int64
	}{
		{"GB", 1 << 30},
		{"MB", 1 << 20},
		{"KB", 1 << 10},
		{"B", 1},
	} {
		if strings.HasSuffix(s, suffix.name) {
			multiplier = suffix.mult
			unit = suffix.name
			numStr = strings.TrimSuffix(s, suffix.name)
			break
		}
	}
	_ = unit

	n, err := strconv.ParseFloat(strings.TrimSpace(numStr), 64)
	if err != nil {
		return 0, fmt.Errorf("invalid body limit %q: %w", s, err)
	}
	if n < 0 {
		return 0, fmt.Errorf("body limit must be positive: %s", s)
	}

	return int64(n * float64(multiplier)), nil
}

// formatBytes 格式化字节数为人类可读字符串。
func formatBytes(b int64) string {
	switch {
	case b >= 1<<30:
		return fmt.Sprintf("%.1fGB", float64(b)/float64(1<<30))
	case b >= 1<<20:
		return fmt.Sprintf("%.1fMB", float64(b)/float64(1<<20))
	case b >= 1<<10:
		return fmt.Sprintf("%.1fKB", float64(b)/float64(1<<10))
	default:
		return fmt.Sprintf("%dB", b)
	}
}

// BodyLimitFromString 便捷函数，从字符串解析大小并返回中间件。
func BodyLimitFromString(limit string) (gin.HandlerFunc, error) {
	maxBytes, err := ParseBodyLimit(limit)
	if err != nil {
		return nil, err
	}
	return BodyLimitWithSize(maxBytes), nil
}
