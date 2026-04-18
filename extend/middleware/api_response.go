package middleware

import (
	"net/http"
	"strings"

	"thinkgin/app"

	"github.com/gin-gonic/gin"
)

func APISuccess(c *gin.Context, data interface{}) {
	c.JSON(http.StatusOK, gin.H{
		"code":       http.StatusOK,
		"message":    "ok",
		"data":       data,
		"request_id": c.GetString("request_id"),
	})
	c.Abort()
}

func APIError(c *gin.Context, httpStatus int, code int, message string) {
	c.JSON(httpStatus, gin.H{
		"code":       code,
		"message":    message,
		"request_id": c.GetString("request_id"),
	})
	c.Abort()
}

func APIErrorHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next()

		if c.Writer.Written() {
			return
		}
		if len(c.Errors) == 0 {
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

func Recovery() gin.HandlerFunc {
	logger := app.GetLogger()

	return gin.CustomRecovery(func(c *gin.Context, recovered interface{}) {
		if logger != nil {
			logger.Errorf("panic recovered: %v", recovered)
		}

		path := c.Request.URL.Path
		if strings.HasPrefix(path, "/api") {
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

func IsAPIPath(path string) bool {
	return strings.HasPrefix(path, "/api")
}
