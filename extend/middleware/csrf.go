// 本文件提供基于双重提交 Cookie (Double Submit Cookie) 的 CSRF 防护中间件。
//
// 原理：
//  1. 首次请求（无 CSRF cookie）时生成随机 token 写入 cookie。
//  2. 后续写操作（POST/PUT/PATCH/DELETE）要求请求同时携带：
//     - Cookie 中的 csrf_token
//     - Header X-CSRF-Token 或表单字段 _csrf_token
//  3. 两者一致才放行，否则返回 403。
//
// 安全保证：SameSite=Strict + HttpOnly=false（前端需读取） + Secure（HTTPS）。
package middleware

import (
	"crypto/rand"
	"crypto/subtle"
	"encoding/hex"
	"net/http"

	"thinkgin/app/ctxkeys"

	"github.com/gin-gonic/gin"
)

const (
	csrfCookieName = "csrf_token"
	csrfHeaderName = "X-CSRF-Token"
	csrfFormField  = "_csrf_token"
	csrfTokenLen   = 32 // 32 bytes = 64 hex chars
)

// CSRF 返回 CSRF 防护中间件。
// 安全方法（GET/HEAD/OPTIONS）只写 cookie 不校验；
// 非安全方法必须携带匹配的 token。
func CSRF() gin.HandlerFunc {
	return func(c *gin.Context) {
		cookieToken, err := c.Cookie(csrfCookieName)
		if err != nil || cookieToken == "" {
			cookieToken = generateCSRFToken()
			c.SetCookie(csrfCookieName, cookieToken, 0, "/", "", c.Request.TLS != nil, false)
		}

		// 将 token 存入 context，模板渲染可用 {{ .csrf_token }}。
		ctxkeys.SetCSRFToken(c, cookieToken)

		// 安全方法不校验 token。
		method := c.Request.Method
		if method == http.MethodGet || method == http.MethodHead || method == http.MethodOptions {
			c.Next()
			return
		}

		// 非安全方法：从 header 或 form 取 token 并与 cookie 比对。
		requestToken := c.GetHeader(csrfHeaderName)
		if requestToken == "" {
			requestToken = c.PostForm(csrfFormField)
		}

		if requestToken == "" || subtle.ConstantTimeCompare([]byte(requestToken), []byte(cookieToken)) != 1 {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{
				"code":    http.StatusForbidden,
				"message": "CSRF token mismatch",
			})
			return
		}

		c.Next()
	}
}

func generateCSRFToken() string {
	b := make([]byte, csrfTokenLen)
	if _, err := rand.Read(b); err != nil {
		return ""
	}
	return hex.EncodeToString(b)
}
