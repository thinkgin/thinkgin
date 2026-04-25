package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"thinkgin/app"

	"github.com/gin-gonic/gin"
)

func TestSecureHeaders_SetsDefaultHeaders(t *testing.T) {
	app.Config = &app.GlobalConfig{}

	r := gin.New()
	r.Use(SecureHeaders())
	r.GET("/ok", func(c *gin.Context) { c.String(http.StatusOK, "ok") })

	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "/ok", nil)
	r.ServeHTTP(w, req)

	want := map[string]string{
		"X-Content-Type-Options": "nosniff",
		"X-Frame-Options":        "DENY",
		"X-XSS-Protection":       "1; mode=block",
		"Referrer-Policy":         "strict-origin-when-cross-origin",
	}
	for k, v := range want {
		if got := w.Header().Get(k); got != v {
			t.Errorf("%s = %q, want %q", k, got, v)
		}
	}
}

func TestSecureHeaders_ConfigOverride(t *testing.T) {
	app.Config = &app.GlobalConfig{}
	app.Config.Middleware.Config = map[string]interface{}{
		"secure_headers": map[string]interface{}{
			"X-Frame-Options":           "SAMEORIGIN",
			"Content-Security-Policy":   "default-src 'self'",
			"X-Content-Type-Options":    "", // 禁用此头
		},
	}

	r := gin.New()
	r.Use(SecureHeaders())
	r.GET("/ok", func(c *gin.Context) { c.String(http.StatusOK, "ok") })

	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "/ok", nil)
	r.ServeHTTP(w, req)

	if got := w.Header().Get("X-Frame-Options"); got != "SAMEORIGIN" {
		t.Errorf("X-Frame-Options = %q, want SAMEORIGIN", got)
	}
	if got := w.Header().Get("Content-Security-Policy"); got != "default-src 'self'" {
		t.Errorf("CSP = %q, want default-src 'self'", got)
	}
	// 空值应跳过（不设置头）
	if got := w.Header().Get("X-Content-Type-Options"); got != "" {
		t.Errorf("X-Content-Type-Options = %q, want empty (disabled)", got)
	}
}
