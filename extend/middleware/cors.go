package middleware

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"thinkgin/app"

	"github.com/gin-gonic/gin"
)

type corsConfig struct {
	AllowOrigins     []string
	AllowMethods     []string
	AllowHeaders     []string
	ExposeHeaders    []string
	AllowCredentials bool
	MaxAge           int
}

func CORS() gin.HandlerFunc {
	cfg := getCORSConfig()

	// 预计算不变的头值，避免每次请求重复拼接。
	allowMethods := strings.Join(cfg.AllowMethods, ", ")
	allowHeaders := strings.Join(cfg.AllowHeaders, ", ")
	exposeHeaders := strings.Join(cfg.ExposeHeaders, ", ")
	maxAge := ""
	if cfg.MaxAge > 0 {
		maxAge = strconv.Itoa(cfg.MaxAge)
	}

	// 构建 origin 白名单 set，O(1) 查找。
	allowAll := len(cfg.AllowOrigins) == 1 && cfg.AllowOrigins[0] == "*"
	originSet := make(map[string]struct{}, len(cfg.AllowOrigins))
	for _, o := range cfg.AllowOrigins {
		originSet[strings.TrimSpace(o)] = struct{}{}
	}

	return func(c *gin.Context) {
		origin := c.GetHeader("Origin")
		if origin == "" {
			c.Next()
			return
		}

		// Access-Control-Allow-Origin 标准只接受单个 origin 或 "*"。
		// 多 origin 场景必须逐请求匹配后动态回写该 origin。
		// W3C 规范：AllowCredentials=true 时不允许返回 "*"，必须回显具体 origin。
		if allowAll && cfg.AllowCredentials {
			// credentials 模式下回显请求 Origin，同时加 Vary 标记。
			c.Header("Access-Control-Allow-Origin", origin)
			c.Header("Vary", "Origin")
		} else if allowAll {
			c.Header("Access-Control-Allow-Origin", "*")
		} else if _, ok := originSet[origin]; ok {
			c.Header("Access-Control-Allow-Origin", origin)
			c.Header("Vary", "Origin")
		} else {
			// Origin 不在白名单：不写 Allow-Origin，浏览器会拦截。
			if c.Request.Method == http.MethodOptions {
				c.AbortWithStatus(http.StatusNoContent)
			} else {
				c.Next()
			}
			return
		}

		if allowMethods != "" {
			c.Header("Access-Control-Allow-Methods", allowMethods)
		}
		if allowHeaders != "" {
			c.Header("Access-Control-Allow-Headers", allowHeaders)
		}
		if exposeHeaders != "" {
			c.Header("Access-Control-Expose-Headers", exposeHeaders)
		}
		if cfg.AllowCredentials {
			c.Header("Access-Control-Allow-Credentials", "true")
		}
		if maxAge != "" {
			c.Header("Access-Control-Max-Age", maxAge)
		}

		if c.Request.Method == http.MethodOptions {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}

		c.Next()
	}
}

func getCORSConfig() corsConfig {
	out := corsConfig{
		AllowOrigins: []string{"*"},
		AllowMethods: []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders: []string{"Origin", "Content-Type", "Authorization"},
		ExposeHeaders: []string{},
		AllowCredentials: false,
		MaxAge: 0,
	}

	cfg := app.GetConfig()
	if cfg == nil || cfg.Middleware.Config == nil {
		return out
	}
	root := cfg.Middleware.Config

	corsRaw, ok := root["cors"]
	if !ok {
		return out
	}

	m, ok := asStringInterfaceMap(corsRaw)
	if !ok {
		return out
	}

	if v, ok := m["allow_origins"]; ok {
		if ss, ok := toStringSlice(v); ok {
			out.AllowOrigins = ss
		}
	}
	if v, ok := m["allow_methods"]; ok {
		if ss, ok := toStringSlice(v); ok {
			out.AllowMethods = ss
		}
	}
	if v, ok := m["allow_headers"]; ok {
		if ss, ok := toStringSlice(v); ok {
			out.AllowHeaders = ss
		}
	}
	if v, ok := m["expose_headers"]; ok {
		if ss, ok := toStringSlice(v); ok {
			out.ExposeHeaders = ss
		}
	}
	if v, ok := m["allow_credentials"]; ok {
		if b, ok := toBool(v); ok {
			out.AllowCredentials = b
		}
	}
	if v, ok := m["max_age"]; ok {
		if i, ok := toInt(v); ok {
			out.MaxAge = i
		}
	}

	// 安全校验：AllowCredentials + AllowOrigins=* 组合在启动时发出警告。
	if out.AllowCredentials && len(out.AllowOrigins) == 1 && out.AllowOrigins[0] == "*" {
		fmt.Println("[cors] WARNING: allow_credentials=true 与 allow_origins=[*] 同时配置，" +
			"框架将自动回显请求 Origin 而非返回 *，建议显式列出可信域名")
	}

	return out
}

func asStringInterfaceMap(v interface{}) (map[string]interface{}, bool) {
	switch mm := v.(type) {
	case map[string]interface{}:
		return mm, true
	case map[interface{}]interface{}:
		out := map[string]interface{}{}
		for k, v := range mm {
			ks, ok := k.(string)
			if !ok {
				continue
			}
			out[ks] = v
		}
		return out, true
	default:
		return nil, false
	}
}

func toBool(v interface{}) (bool, bool) {
	switch vv := v.(type) {
	case bool:
		return vv, true
	case string:
		b, err := strconv.ParseBool(strings.TrimSpace(vv))
		return b, err == nil
	default:
		return false, false
	}
}

func toInt(v interface{}) (int, bool) {
	switch vv := v.(type) {
	case int:
		return vv, true
	case int64:
		return int(vv), true
	case float64:
		return int(vv), true
	case string:
		i, err := strconv.Atoi(strings.TrimSpace(vv))
		return i, err == nil
	default:
		return 0, false
	}
}
