// Package ctxkeys 定义所有在 gin.Context 中使用的 key 常量及其
// 类型安全的 getter / setter 函数。
//
// 业务代码应使用 ctxkeys.GetXxx / ctxkeys.SetXxx，而非直接操作
// c.Set / c.Get + 字符串字面量，以避免 typo 和类型断言错误。
package ctxkeys

import "github.com/gin-gonic/gin"

// ────────────────── key 常量 ──────────────────

const (
	// RequestID 请求唯一标识
	RequestID = "thinkgin:request_id"

	// CSRFToken CSRF 令牌
	CSRFToken = "thinkgin:csrf_token"
)

// ────────────────── Request ID ──────────────────

// SetRequestID 将 request_id 写入 gin.Context。
func SetRequestID(c *gin.Context, id string) {
	c.Set(RequestID, id)
}

// GetRequestID 从 gin.Context 读取 request_id。
func GetRequestID(c *gin.Context) string {
	return c.GetString(RequestID)
}

// ────────────────── CSRF Token ──────────────────

// SetCSRFToken 将 csrf_token 写入 gin.Context。
func SetCSRFToken(c *gin.Context, token string) {
	c.Set(CSRFToken, token)
}

// GetCSRFToken 从 gin.Context 读取 csrf_token。
func GetCSRFToken(c *gin.Context) string {
	return c.GetString(CSRFToken)
}
