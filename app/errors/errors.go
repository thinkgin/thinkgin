// Package errors 提供 ThinkGin 统一的结构化错误码体系。
//
// 设计目标：
//   - 将 HTTP 状态码与业务错误码解耦：同一个 HTTP 400 可对应多个业务场景。
//   - 错误码自带消息模板，支持国际化扩展。
//   - 实现 error 接口，可在 Go 原生 errors.Is / errors.As 中使用。
//   - 通过 Gin 中间件自动捕获并格式化为统一 JSON 响应。
//
// 用法：
//
//	var ErrUserNotFound = errors.New(404, 100001, "用户不存在")
//	return ErrUserNotFound                                // 原始
//	return ErrUserNotFound.WithMsg("ID=42 的用户不存在")   // 覆盖消息
//	return ErrUserNotFound.WithData(gin.H{"id": 42})     // 附加数据
package errors

import (
	"fmt"
	"net/http"
)

// AppError 是 ThinkGin 的结构化错误类型。
// 同时承载 HTTP 状态码、业务错误码和人可读消息。
type AppError struct {
	HTTPStatus int         `json:"-"`                  // HTTP 响应状态码
	Code       int         `json:"code"`               // 业务错误码
	Message    string      `json:"message"`            // 面向用户的错误消息
	Data       interface{} `json:"data,omitempty"`     // 可选的附加数据（调试信息、字段错误等）
	cause      error       // 原始错误，用于 Unwrap
}

// New 创建一个不可变的错误码定义。
// 通常在包级 var 块中声明，运行时通过 WithMsg / WithData 派生新实例。
func New(httpStatus, code int, message string) *AppError {
	return &AppError{
		HTTPStatus: httpStatus,
		Code:       code,
		Message:    message,
	}
}

// Error 实现 error 接口。
func (e *AppError) Error() string {
	if e.cause != nil {
		return fmt.Sprintf("[%d] %s: %v", e.Code, e.Message, e.cause)
	}
	return fmt.Sprintf("[%d] %s", e.Code, e.Message)
}

// Unwrap 支持 errors.Is / errors.As 链式匹配。
func (e *AppError) Unwrap() error {
	return e.cause
}

// Is 支持 errors.Is 比较：只要 Code 相同即视为相同错误。
func (e *AppError) Is(target error) bool {
	t, ok := target.(*AppError)
	if !ok {
		return false
	}
	return e.Code == t.Code
}

// WithMsg 返回新实例，覆盖错误消息，不修改原定义。
func (e *AppError) WithMsg(msg string) *AppError {
	return &AppError{
		HTTPStatus: e.HTTPStatus,
		Code:       e.Code,
		Message:    msg,
		Data:       e.Data,
		cause:      e.cause,
	}
}

// WithData 返回新实例，附加额外数据。
func (e *AppError) WithData(data interface{}) *AppError {
	return &AppError{
		HTTPStatus: e.HTTPStatus,
		Code:       e.Code,
		Message:    e.Message,
		Data:       data,
		cause:      e.cause,
	}
}

// WithCause 返回新实例，包装底层错误。
func (e *AppError) WithCause(err error) *AppError {
	return &AppError{
		HTTPStatus: e.HTTPStatus,
		Code:       e.Code,
		Message:    e.Message,
		Data:       e.Data,
		cause:      err,
	}
}

// ---- 预定义通用错误码（业务模块可在自己的包中扩展） ----

var (
	// 客户端错误 4xx
	ErrBadRequest          = New(http.StatusBadRequest, 400000, "请求参数错误")
	ErrUnauthorized        = New(http.StatusUnauthorized, 401000, "未认证")
	ErrTokenExpired        = New(http.StatusUnauthorized, 401001, "凭证已过期")
	ErrTokenInvalid        = New(http.StatusUnauthorized, 401002, "凭证无效")
	ErrForbidden           = New(http.StatusForbidden, 403000, "无权限")
	ErrNotFound            = New(http.StatusNotFound, 404000, "资源不存在")
	ErrMethodNotAllowed    = New(http.StatusMethodNotAllowed, 405000, "方法不允许")
	ErrConflict            = New(http.StatusConflict, 409000, "资源冲突")
	ErrValidation          = New(http.StatusUnprocessableEntity, 422000, "参数校验失败")
	ErrTooManyRequests     = New(http.StatusTooManyRequests, 429000, "请求过于频繁")

	// 服务端错误 5xx
	ErrInternal            = New(http.StatusInternalServerError, 500000, "服务内部错误")
	ErrServiceUnavailable  = New(http.StatusServiceUnavailable, 503000, "服务暂不可用")
	ErrTimeout             = New(http.StatusGatewayTimeout, 504000, "请求超时")
)
