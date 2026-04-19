package middleware

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"

	"thinkgin/app"
)

// newCORSRouter 创建一个 CORS + 200 OK handler 的路由器，便于断言响应头。
func newCORSRouter() *gin.Engine {
	r := gin.New()
	r.Use(CORS())
	r.GET("/ok", func(c *gin.Context) { c.String(http.StatusOK, "ok") })
	return r
}

func TestCORS_AppliesAllHeadersFromConfig(t *testing.T) {
	app.Config = &app.GlobalConfig{}
	app.Config.Middleware.Config = map[string]interface{}{
		"cors": map[string]interface{}{
			"allow_origins":     []interface{}{"https://a.com", "https://b.com"},
			"allow_methods":     []interface{}{"GET", "POST"},
			"allow_headers":     []interface{}{"X-Test"},
			"expose_headers":    []interface{}{"X-Total-Count"},
			"allow_credentials": true,
			"max_age":           3600,
		},
	}

	r := newCORSRouter()

	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "/ok", nil)
	req.Header.Set("Origin", "https://a.com")
	r.ServeHTTP(w, req)

	want := map[string]string{
		"Access-Control-Allow-Origin":      "https://a.com, https://b.com",
		"Access-Control-Allow-Methods":     "GET, POST",
		"Access-Control-Allow-Headers":     "X-Test",
		"Access-Control-Expose-Headers":    "X-Total-Count",
		"Access-Control-Allow-Credentials": "true",
		"Access-Control-Max-Age":           "3600",
		"Vary":                             "Origin",
	}
	for k, v := range want {
		if got := w.Header().Get(k); got != v {
			t.Errorf("%s = %q, want %q", k, got, v)
		}
	}
}

func TestCORS_OptionsPreflightReturns204(t *testing.T) {
	app.Config = &app.GlobalConfig{}
	r := newCORSRouter()

	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodOptions, "/ok", nil)
	req.Header.Set("Origin", "https://example.com")
	r.ServeHTTP(w, req)

	if w.Code != http.StatusNoContent {
		t.Errorf("OPTIONS status=%d, want 204", w.Code)
	}
}

func TestCORS_OmitsHeadersWhenOriginMissing(t *testing.T) {
	// 无 Origin 头的请求不应被打上 Allow-Origin，避免给同源请求加噪。
	app.Config = &app.GlobalConfig{}
	r := newCORSRouter()

	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "/ok", nil)
	r.ServeHTTP(w, req)

	if got := w.Header().Get("Access-Control-Allow-Origin"); got != "" {
		t.Errorf("Allow-Origin = %q, want empty when Origin header absent", got)
	}
}

func TestCORS_DefaultsWhenUnconfigured(t *testing.T) {
	// middleware.config 为空时应使用默认值：Allow-Origin=*、常见方法集。
	app.Config = &app.GlobalConfig{}
	r := newCORSRouter()

	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "/ok", nil)
	req.Header.Set("Origin", "https://example.com")
	r.ServeHTTP(w, req)

	if got := w.Header().Get("Access-Control-Allow-Origin"); got != "*" {
		t.Errorf("default Allow-Origin = %q, want *", got)
	}
	methods := w.Header().Get("Access-Control-Allow-Methods")
	for _, m := range []string{"GET", "POST", "OPTIONS"} {
		if !strings.Contains(methods, m) {
			t.Errorf("default Allow-Methods=%q should contain %s", methods, m)
		}
	}
}
