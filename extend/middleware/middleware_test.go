package middleware

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

func init() {
	gin.SetMode(gin.TestMode)
}

// TestIsAPIPath 验证 API 路径判断
func TestIsAPIPath(t *testing.T) {
	tests := []struct {
		path string
		want bool
	}{
		{"/api/v1/users", true},
		{"/api/v2/orders", true},
		{"/api", true},
		{"/index", false},
		{"/", false},
		{"/static/js/app.js", false},
		{"/apiary", false}, // 前缀相同但不属于 API 路径族，必须返回 false
	}
	for _, tt := range tests {
		got := IsAPIPath(tt.path)
		if got != tt.want {
			t.Errorf("IsAPIPath(%q) = %v, want %v", tt.path, got, tt.want)
		}
	}
}

// TestAPISuccess 验证成功响应格式
func TestAPISuccess(t *testing.T) {
	r := gin.New()
	r.GET("/test", func(c *gin.Context) {
		c.Set("request_id", "test-123")
		APISuccess(c, map[string]string{"hello": "world"})
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/test", nil)
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("status = %d, want %d", w.Code, http.StatusOK)
	}

	var resp map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &resp)

	if resp["code"].(float64) != 200 {
		t.Errorf("code = %v, want 200", resp["code"])
	}
	if resp["message"] != "ok" {
		t.Errorf("message = %v, want ok", resp["message"])
	}
	if resp["request_id"] != "test-123" {
		t.Errorf("request_id = %v, want test-123", resp["request_id"])
	}
}

// TestAPIError 验证错误响应格式
func TestAPIError(t *testing.T) {
	r := gin.New()
	r.GET("/test", func(c *gin.Context) {
		c.Set("request_id", "err-456")
		APIError(c, http.StatusNotFound, 404, "not found")
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/test", nil)
	r.ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("status = %d, want %d", w.Code, http.StatusNotFound)
	}

	var resp map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &resp)

	if resp["code"].(float64) != 404 {
		t.Errorf("code = %v, want 404", resp["code"])
	}
	if resp["message"] != "not found" {
		t.Errorf("message = %v, want 'not found'", resp["message"])
	}
}

// TestRequestIDMiddleware 验证 RequestID 中间件生成 ID
func TestRequestIDMiddleware(t *testing.T) {
	r := gin.New()
	r.Use(RequestID())
	r.GET("/test", func(c *gin.Context) {
		id := c.GetString("request_id")
		c.String(200, id)
	})

	// 不传 X-Request-ID，应自动生成
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/test", nil)
	r.ServeHTTP(w, req)

	body := w.Body.String()
	if body == "" {
		t.Error("expected auto-generated request_id, got empty")
	}
	if w.Header().Get("X-Request-ID") == "" {
		t.Error("expected X-Request-ID in response header")
	}

	// 传入 X-Request-ID，应透传
	w2 := httptest.NewRecorder()
	req2, _ := http.NewRequest("GET", "/test", nil)
	req2.Header.Set("X-Request-ID", "custom-id-789")
	r.ServeHTTP(w2, req2)

	if w2.Body.String() != "custom-id-789" {
		t.Errorf("expected custom-id-789, got %s", w2.Body.String())
	}
}

// TestCORSMiddleware 验证 CORS 中间件处理 OPTIONS 预检
func TestCORSMiddleware(t *testing.T) {
	r := gin.New()
	r.Use(CORS())
	r.GET("/test", func(c *gin.Context) {
		c.String(200, "ok")
	})

	// OPTIONS 预检请求
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("OPTIONS", "/test", nil)
	req.Header.Set("Origin", "http://example.com")
	r.ServeHTTP(w, req)

	if w.Code != http.StatusNoContent && w.Code != http.StatusOK {
		t.Errorf("OPTIONS status = %d, want 204 or 200", w.Code)
	}

	acao := w.Header().Get("Access-Control-Allow-Origin")
	if acao == "" {
		t.Error("expected Access-Control-Allow-Origin header")
	}

	// 正常 GET 请求应通过
	w2 := httptest.NewRecorder()
	req2, _ := http.NewRequest("GET", "/test", nil)
	req2.Header.Set("Origin", "http://example.com")
	r.ServeHTTP(w2, req2)

	if w2.Code != http.StatusOK {
		t.Errorf("GET status = %d, want 200", w2.Code)
	}
}
