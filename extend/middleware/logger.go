package middleware

import (
	"crypto/rand"
	"encoding/hex"
	"strings"
	"thinkgin/app"
	"thinkgin/app/ctxkeys"
	"time"

	"github.com/gin-gonic/gin"
)

func RequestID() gin.HandlerFunc {
	return func(c *gin.Context) {
		rid := c.GetHeader("X-Request-ID")
		if strings.TrimSpace(rid) == "" {
			rid = newRequestID()
		}
		ctxkeys.SetRequestID(c, rid)
		c.Header("X-Request-ID", rid)
		c.Next()
	}
}

func AccessLogger() gin.HandlerFunc {
	logger := app.GetLogger()

	return func(c *gin.Context) {
		path := c.Request.URL.Path
		if shouldSkipAccessLog(path) {
			c.Next()
			return
		}

		start := time.Now()
		c.Next()

		latency := time.Since(start)
		statusCode := c.Writer.Status()
		method := c.Request.Method
		routePath := c.FullPath()
		if routePath == "" {
			routePath = path
		}

		fields := map[string]any{
			"request_id":  ctxkeys.GetRequestID(c),
			"status_code": statusCode,
			"latency_ms":  latency.Milliseconds(),
			"client_ip":   c.ClientIP(),
			"method":      method,
			"path":        routePath,
			"user_agent":  c.Request.UserAgent(),
		}

		if statusCode >= 500 {
			logger.WithFields(fields).Error("HTTP Request")
			return
		}
		if statusCode >= 400 {
			logger.WithFields(fields).Warn("HTTP Request")
			return
		}
		logger.WithFields(fields).Info("HTTP Request")
	}
}

func shouldSkipAccessLog(path string) bool {
	for _, p := range getAccessLogSkipPaths() {
		if path == p {
			return true
		}
	}
	return false
}

func getAccessLogSkipPaths() []string {
	paths := []string{"/metrics", "/ping", "/livez", "/readyz"}

	cfg := app.GetConfig()
	if cfg == nil {
		return paths
	}
	root := cfg.Middleware.Config
	if root == nil {
		return paths
	}

	if v, ok := lookupConfig(root, "request_log", "skip_paths"); ok {
		if ss, ok := toStringSlice(v); ok {
			return ss
		}
	}
	if v, ok := lookupConfig(root, "logger", "skip_paths"); ok {
		if ss, ok := toStringSlice(v); ok {
			return ss
		}
	}

	return paths
}

func lookupConfig(root map[string]interface{}, key string, subkey string) (interface{}, bool) {
	child, ok := root[key]
	if !ok {
		return nil, false
	}

	switch m := child.(type) {
	case map[string]interface{}:
		v, ok := m[subkey]
		return v, ok
	case map[interface{}]interface{}:
		v, ok := m[subkey]
		return v, ok
	default:
		return nil, false
	}
}

func toStringSlice(v interface{}) ([]string, bool) {
	switch vv := v.(type) {
	case []string:
		return vv, true
	case []interface{}:
		out := make([]string, 0, len(vv))
		for _, it := range vv {
			s, ok := it.(string)
			if !ok {
				continue
			}
			out = append(out, s)
		}
		if len(out) == 0 {
			return nil, false
		}
		return out, true
	default:
		return nil, false
	}
}

func newRequestID() string {
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		return ""
	}
	return hex.EncodeToString(b)
}

// 业务日志记录器
func BusinessLogger(level string, message string, fields map[string]interface{}) {
	logger := app.GetLogger()

	l := logger.WithFields(fields)

	switch level {
	case "debug":
		l.Debug(message)
	case "info":
		l.Info(message)
	case "warn":
		l.Warn(message)
	case "error":
		l.Error(message)
	case "fatal":
		l.Fatal(message)
	default:
		l.Info(message)
	}
}

