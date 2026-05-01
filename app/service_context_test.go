package app

import (
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestSvcMiddleware_InjectsServiceContext(t *testing.T) {
	gin.SetMode(gin.TestMode)

	svc := &ServiceContext{
		Config: &GlobalConfig{},
		Log:    NewSlogAdapter(slog.Default()),
	}
	svc.Config.App.Name = "SvcTest"

	r := gin.New()
	r.Use(SvcMiddleware(svc))
	r.GET("/test", func(c *gin.Context) {
		got := SvcFromGin(c)
		if got.Config.App.Name != "SvcTest" {
			t.Errorf("SvcFromGin Config.App.Name = %q, want SvcTest", got.Config.App.Name)
		}
		c.String(http.StatusOK, "ok")
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "/test", nil)
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("status = %d, want 200", w.Code)
	}
}

func TestSvcFromGin_FallsBackToGlobals(t *testing.T) {
	gin.SetMode(gin.TestMode)

	// 不注入 SvcMiddleware，SvcFromGin 应从全局变量兜底
	r := gin.New()
	r.GET("/fallback", func(c *gin.Context) {
		got := SvcFromGin(c)
		if got == nil {
			t.Fatal("SvcFromGin returned nil without middleware")
		}
		if got.Config == nil {
			t.Error("fallback SvcFromGin.Config should not be nil")
		}
		if got.Log == nil {
			t.Error("fallback SvcFromGin.Log should not be nil")
		}
		c.String(http.StatusOK, "ok")
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "/fallback", nil)
	r.ServeHTTP(w, req)
}
