// Package middleware 提供 ThinkGin 的横切中间件集合。
// 本文件负责统一 API 响应格式，并提供基于 Recovery 的 panic 兜底。
package middleware

import (
	"errors"
	"net/http"
	"strings"

	"thinkgin/app"
	"thinkgin/app/ctxkeys"
	apperrors "thinkgin/app/errors"

	"github.com/gin-gonic/gin"
)

// APISuccess 输出统一的成功响应。
// 响应体结构：{ code, message, data, request_id }。
func APISuccess(c *gin.Context, data interface{}) {
	c.JSON(http.StatusOK, gin.H{
		"code":       http.StatusOK,
		"message":    "ok",
		"data":       data,
		"request_id": ctxkeys.GetRequestID(c),
	})
	c.Abort()
}

// APIError 输出统一的错误响应。
// httpStatus 控制 HTTP 状态码，code 为业务错误码。
func APIError(c *gin.Context, httpStatus int, code int, message string) {
	c.JSON(httpStatus, gin.H{
		"code":       code,
		"message":    message,
		"request_id": ctxkeys.GetRequestID(c),
	})
	c.Abort()
}

// APIAppError 输出由 *apperrors.AppError 驱动的结构化错误响应。
// 自动提取 HTTPStatus / Code / Message / Data。
func APIAppError(c *gin.Context, err *apperrors.AppError) {
	body := gin.H{
		"code":       err.Code,
		"message":    err.Message,
		"request_id": ctxkeys.GetRequestID(c),
	}
	if err.Data != nil {
		body["data"] = err.Data
	}
	c.JSON(err.HTTPStatus, body)
	c.Abort()
}

// APIErrorHandler 用于捕获 handler 通过 c.Error 抛出的错误，
// 并在未写入响应时统一返回结构化响应。
// 若错误为 *apperrors.AppError 类型，自动提取错误码；否则回退到 500。
func APIErrorHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next()

		if c.Writer.Written() || len(c.Errors) == 0 {
			return
		}

		lastErr := c.Errors.Last().Err
		var appErr *apperrors.AppError
		if errors.As(lastErr, &appErr) {
			APIAppError(c, appErr)
			return
		}

		msg := "internal error"
		cfg := app.GetConfig()
		if cfg != nil && cfg.App.Debug {
			msg = lastErr.Error()
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
