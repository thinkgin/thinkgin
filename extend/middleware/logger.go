package middleware

import (
	"thinkgin/app"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

// 日志记录到文件
func LoggerToFile() gin.HandlerFunc {
	logger := app.GetLogger()

	return func(c *gin.Context) {
		// 开始时间
		startTime := time.Now()

		// 处理请求
		c.Next()

		// 结束时间
		endTime := time.Now()

		// 执行时间
		latencyTime := endTime.Sub(startTime)

		// 请求方式
		reqMethod := c.Request.Method

		// 请求路由
		reqUri := c.Request.RequestURI

		// 状态码
		statusCode := c.Writer.Status()

		// 请求IP
		clientIP := c.ClientIP()

		// 用户代理
		userAgent := c.Request.UserAgent()

		// 日志格式
		logger.WithFields(logrus.Fields{
			"status_code":  statusCode,
			"latency_time": latencyTime,
			"client_ip":    clientIP,
			"req_method":   reqMethod,
			"req_uri":      reqUri,
			"user_agent":   userAgent,
		}).Info("HTTP Request")
	}
}

// 业务日志记录器
func BusinessLogger(level string, message string, fields map[string]interface{}) {
	logger := app.GetLogger()

	logFields := logrus.Fields{}
	for k, v := range fields {
		logFields[k] = v
	}

	switch level {
	case "debug":
		logger.WithFields(logFields).Debug(message)
	case "info":
		logger.WithFields(logFields).Info(message)
	case "warn":
		logger.WithFields(logFields).Warn(message)
	case "error":
		logger.WithFields(logFields).Error(message)
	case "fatal":
		logger.WithFields(logFields).Fatal(message)
	case "panic":
		logger.WithFields(logFields).Panic(message)
	default:
		logger.WithFields(logFields).Info(message)
	}
}

// 日志记录到 MongoDB
func LoggerToMongo() gin.HandlerFunc {
	return func(c *gin.Context) {
		// TODO: 实现MongoDB日志记录
	}
}

// 日志记录到 ES
func LoggerToES() gin.HandlerFunc {
	return func(c *gin.Context) {
		// TODO: 实现Elasticsearch日志记录
	}
}

// 日志记录到 MQ
func LoggerToMQ() gin.HandlerFunc {
	return func(c *gin.Context) {
		// TODO: 实现消息队列日志记录
	}
}
