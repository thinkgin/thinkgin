// Package middleware 提供 ThinkGin 的横切中间件集合。
// 本文件负责统一 API 响应格式，并提供基于 Recovery 的 panic 兜底。
package middleware

import (
	"net/http"
	"strings"

	"thinkgin/app"

	"github.com/gin-gonic/gin"
)

// APISuccess 输出统一的成功响应。
// 响应体结构：{ code, message, data, request_id }。
func APISuccess(c *gin.Context, data interface{}) {
	c.JSON(http.StatusOK, gin.H{
		"code":       http.StatusOK,
		"message":    "ok",
		"data":       data,
		"request_id": c.GetString("request_id"),
	})
	c.Abort()
}

// APIError 输出统一的错误响应。
// httpStatus 控制 HTTP 状态码，code 为业务错误码。
func APIError(c *gin.Context, httpStatus int, code int, message string) {
	c.JSON(httpStatus, gin.H{
		"code":       code,
		"message":    message,
		"request_id": c.GetString("request_id"),
	})
	c.Abort()
}

// APIErrorHandler 用于捕获 handler 通过 c.Error 抛出的错误，
// 并在未写入响应时统一返回 500 结构化响应。
func APIErrorHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next()

		if c.Writer.Written() || len(c.Errors) == 0 {
			return
		}

		msg := "internal error"
		cfg := app.GetConfig()
		if cfg != nil && cfg.App.Debug {
			msg = c.Errors.Last().Error()
		}
		APIError(c, http.StatusInternalServerError, http.StatusInternalServerError, msg)
	}
}

// Recovery 捕获 panic，API 路径输出 JSON，其他路径返回纯 500 状态码。
func Recovery() gin.HandlerFunc {
	logger := app.GetLogger()

	return gin.CustomRecovery(func(c *gin.Context, recovered interface{}) {
		if logger != nil {
			logger.Errorf("panic recovered: %v", recovered)
		}

		if IsAPIPath(c.Request.URL.Path) {
			msg := "internal error"
			cfg := app.GetConfig()
			if cfg != nil && cfg.App.Debug {
				msg = "panic"
			}
			APIError(c, http.StatusInternalServerError, http.StatusInternalServerError, msg)
			return
		}

		c.AbortWithStatus(http.StatusInternalServerError)
	})
}

// IsAPIPath 判断是否属于 /api 路径族。
// 仅匹配精确的 /api 或 /api/... 前缀，避免误伤 /apiary、/apiXxx 等业务路径。
func IsAPIPath(path string) bool {
	return path == "/api" || strings.HasPrefix(path, "/api/")
}
