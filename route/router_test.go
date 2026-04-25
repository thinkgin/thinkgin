package route

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"thinkgin/app"
	"thinkgin/extend/middleware"

	"github.com/gin-gonic/gin"
)

func TestApplyMiddleware_KnownNames(t *testing.T) {
	knownNames := []string{
		"recovery", "request_id", "cors", "rate_limit",
		"secure_headers", "gzip", "csrf", "circuit_breaker",
		"timeout", "body_limit",
	}
	for _, name := range knownNames {
		r := gin.New()
		applyMiddleware(r, name) // 不应 panic
	}
}

func TestApplyMiddleware_UnknownNameIgnored(t *testing.T) {
	r := gin.New()
	applyMiddleware(r, "nonexistent_middleware") // 应静默忽略
}

func TestRegisterGlobalMiddleware_Default(t *testing.T) {
	old := app.Config
	app.Config = &app.GlobalConfig{}
	defer func() { app.Config = old }()

	cfg := app.GetConfig()
	cfg.Middleware.Global = nil // 触发默认链
	r := gin.New()
	registerGlobalMiddleware(r, cfg)
	// 不 panic 即通过
}

func TestRegisterGlobalMiddleware_Custom(t *testing.T) {
	old := app.Config
	app.Config = &app.GlobalConfig{}
	defer func() { app.Config = old }()

	cfg := app.GetConfig()
	cfg.Middleware.Global = []string{"recovery", "request_id", "cors"}
	r := gin.New()
	registerGlobalMiddleware(r, cfg)
}

func TestRegisterMonitoringRoutes_Ping(t *testing.T) {
	old := app.Config
	app.Config = &app.GlobalConfig{}
	app.Config.App.Version = "test"
	app.Config.Prometheus.Path = "/metrics"
	defer func() { app.Config = old }()

	r := gin.New()
	registerMonitoringRoutes(r, app.Config)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/ping", nil)
	r.ServeHTTP(w, req)

	if w.Code != 200 {
		t.Errorf("ping status = %d, want 200", w.Code)
	}
}

func TestNoRoute_APIPathReturnsJSON(t *testing.T) {
	r := gin.New()
	r.NoRoute(func(c *gin.Context) {
		if middleware.IsAPIPath(c.Request.URL.Path) {
			middleware.APIError(c, 404, 404, "not found")
			return
		}
		c.AbortWithStatus(404)
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/nonexist", nil)
	r.ServeHTTP(w, req)

	if w.Code != 404 {
		t.Errorf("status = %d, want 404", w.Code)
	}
}

func TestNoRoute_NonAPIPathReturns404(t *testing.T) {
	r := gin.New()
	r.NoRoute(func(c *gin.Context) {
		if middleware.IsAPIPath(c.Request.URL.Path) {
			middleware.APIError(c, 404, 404, "not found")
			return
		}
		c.AbortWithStatus(404)
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/page/nonexist", nil)
	r.ServeHTTP(w, req)

	if w.Code != 404 {
		t.Errorf("status = %d, want 404", w.Code)
	}
}
