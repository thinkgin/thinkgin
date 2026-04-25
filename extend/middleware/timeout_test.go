package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
)

func TestTimeout_NormalRequest(t *testing.T) {
	r := gin.New()
	r.Use(TimeoutWithDuration(2 * time.Second))
	r.GET("/test", func(c *gin.Context) {
		c.String(200, "ok")
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/test", nil)
	r.ServeHTTP(w, req)

	if w.Code != 200 {
		t.Errorf("status = %d, want 200", w.Code)
	}
	if w.Body.String() != "ok" {
		t.Errorf("body = %q, want ok", w.Body.String())
	}
}

func TestTimeout_SlowRequestReturns504(t *testing.T) {
	r := gin.New()
	r.Use(TimeoutWithDuration(50 * time.Millisecond))
	r.GET("/slow", func(c *gin.Context) {
		select {
		case <-time.After(500 * time.Millisecond):
			c.String(200, "should not reach")
		case <-c.Request.Context().Done():
			// handler 检测到超时，提前退出
			return
		}
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/slow", nil)
	r.ServeHTTP(w, req)

	if w.Code != http.StatusGatewayTimeout {
		t.Errorf("status = %d, want 504", w.Code)
	}
}

func TestTimeout_ContextPropagation(t *testing.T) {
	r := gin.New()
	r.Use(TimeoutWithDuration(2 * time.Second))
	r.GET("/ctx", func(c *gin.Context) {
		deadline, ok := c.Request.Context().Deadline()
		if !ok {
			t.Error("expected context to have deadline")
			return
		}
		remaining := time.Until(deadline)
		if remaining <= 0 || remaining > 2*time.Second {
			t.Errorf("unexpected remaining time: %v", remaining)
		}
		c.String(200, "ok")
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/ctx", nil)
	r.ServeHTTP(w, req)

	if w.Code != 200 {
		t.Errorf("status = %d, want 200", w.Code)
	}
}

func TestTimeout_DefaultDuration(t *testing.T) {
	r := gin.New()
	r.Use(Timeout()) // 默认 30s
	r.GET("/test", func(c *gin.Context) {
		deadline, ok := c.Request.Context().Deadline()
		if !ok {
			t.Error("expected deadline")
			return
		}
		remaining := time.Until(deadline)
		if remaining < 29*time.Second || remaining > 31*time.Second {
			t.Errorf("expected ~30s remaining, got %v", remaining)
		}
		c.String(200, "ok")
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/test", nil)
	r.ServeHTTP(w, req)

	if w.Code != 200 {
		t.Errorf("status = %d, want 200", w.Code)
	}
}
