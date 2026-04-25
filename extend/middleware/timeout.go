// 本文件提供请求超时控制中间件。
//
// 为每个请求设置 context.WithTimeout，超时后返回 504 Gateway Timeout。
// 下游 handler 应使用 c.Request.Context() 传递超时信号。
//
// 使用方式：
//   在 config/middleware.yaml 中添加 "timeout" 到 global 列表，
//   或手动调用 middleware.Timeout(5 * time.Second) 注册。
package middleware

import (
	"context"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

// defaultTimeout 默认请求超时时间。
const defaultTimeout = 30 * time.Second

// Timeout 返回使用默认超时（30s）的超时控制中间件。
func Timeout() gin.HandlerFunc {
	return TimeoutWithDuration(defaultTimeout)
}

// TimeoutWithDuration 返回使用自定义超时时间的中间件。
func TimeoutWithDuration(timeout time.Duration) gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(c.Request.Context(), timeout)
		defer cancel()

		c.Request = c.Request.WithContext(ctx)

		// 用 channel 等待 handler 完成或超时
		done := make(chan struct{}, 1)
		go func() {
			c.Next()
			done <- struct{}{}
		}()

		select {
		case <-done:
			// handler 正常完成
		case <-ctx.Done():
			// 超时
			if !c.Writer.Written() {
				c.AbortWithStatusJSON(http.StatusGatewayTimeout, gin.H{
					"code":    http.StatusGatewayTimeout,
					"message": "request timeout",
				})
			}
		}
	}
}
