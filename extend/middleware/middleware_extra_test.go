package middleware

import (
	"bytes"
	"compress/gzip"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"thinkgin/app"

	"github.com/gin-gonic/gin"
)

// ──────────── api_response.go 补充 ────────────

func TestAPIErrorHandler_NoError(t *testing.T) {
	r := gin.New()
	r.Use(APIErrorHandler())
	r.GET("/ok", func(c *gin.Context) {
		c.String(200, "fine")
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/ok", nil)
	r.ServeHTTP(w, req)

	if w.Code != 200 {
		t.Errorf("status = %d, want 200", w.Code)
	}
}

func TestAPIErrorHandler_WithError(t *testing.T) {
	r := gin.New()
	r.Use(APIErrorHandler())
	r.GET("/fail", func(c *gin.Context) {
		_ = c.Error(errors.New("something broke"))
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/fail", nil)
	r.ServeHTTP(w, req)

	if w.Code != 500 {
		t.Errorf("status = %d, want 500", w.Code)
	}
}

func TestAPIErrorHandler_DebugMode(t *testing.T) {
	old := app.Config
	app.Config = &app.GlobalConfig{}
	app.Config.App.Debug = true
	defer func() { app.Config = old }()

	r := gin.New()
	r.Use(APIErrorHandler())
	r.GET("/fail", func(c *gin.Context) {
		_ = c.Error(errors.New("detailed error"))
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/fail", nil)
	r.ServeHTTP(w, req)

	if w.Code != 500 {
		t.Errorf("status = %d, want 500", w.Code)
	}
	if !bytes.Contains(w.Body.Bytes(), []byte("detailed error")) {
		t.Errorf("debug mode should expose error, got %q", w.Body.String())
	}
}

func TestRecovery_APIPanicReturnsJSON(t *testing.T) {
	r := gin.New()
	r.Use(Recovery())
	r.GET("/api/panic", func(c *gin.Context) {
		panic("test panic")
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/panic", nil)
	r.ServeHTTP(w, req)

	if w.Code != 500 {
		t.Errorf("status = %d, want 500", w.Code)
	}
	if !bytes.Contains(w.Body.Bytes(), []byte("internal error")) {
		t.Errorf("body should contain 'internal error', got %q", w.Body.String())
	}
}

func TestRecovery_NonAPIPanicReturns500(t *testing.T) {
	r := gin.New()
	r.Use(Recovery())
	r.GET("/page/panic", func(c *gin.Context) {
		panic("oops")
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/page/panic", nil)
	r.ServeHTTP(w, req)

	if w.Code != 500 {
		t.Errorf("status = %d, want 500", w.Code)
	}
}

// ──────────── gzip.go 补充 ────────────

func TestGzip_NoAcceptEncodingSkips(t *testing.T) {
	r := gin.New()
	r.Use(Gzip())
	r.GET("/test", func(c *gin.Context) {
		c.String(200, "no gzip")
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/test", nil)
	// 不设置 Accept-Encoding
	r.ServeHTTP(w, req)

	if w.Code != 200 {
		t.Errorf("status = %d, want 200", w.Code)
	}
	if w.Body.String() != "no gzip" {
		t.Errorf("body = %q, want 'no gzip'", w.Body.String())
	}
}

func TestGzip_WriteStringCoverage(t *testing.T) {
	r := gin.New()
	r.Use(Gzip())
	r.GET("/test", func(c *gin.Context) {
		// gin 内部可能使用 WriteString，手动触发
		gw, ok := c.Writer.(*gzipWriter)
		if ok {
			_, _ = gw.WriteString("hello from writestring")
		} else {
			c.String(200, "hello")
		}
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/test", nil)
	req.Header.Set("Accept-Encoding", "gzip")
	r.ServeHTTP(w, req)

	if w.Code != 200 {
		t.Errorf("status = %d, want 200", w.Code)
	}
}

func TestGzipDecompressRequest_NormalRequest(t *testing.T) {
	r := gin.New()
	r.Use(GzipDecompressRequest())
	r.POST("/test", func(c *gin.Context) {
		body, _ := io.ReadAll(c.Request.Body)
		c.String(200, string(body))
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/test", bytes.NewReader([]byte("plain")))
	r.ServeHTTP(w, req)

	if w.Code != 200 || w.Body.String() != "plain" {
		t.Errorf("got %d %q", w.Code, w.Body.String())
	}
}

func TestGzipDecompressRequest_ValidGzip(t *testing.T) {
	// 创建 gzip 压缩的 body
	var buf bytes.Buffer
	gw := gzip.NewWriter(&buf)
	_, _ = gw.Write([]byte("compressed data"))
	_ = gw.Close()

	r := gin.New()
	r.Use(GzipDecompressRequest())
	r.POST("/test", func(c *gin.Context) {
		body, _ := io.ReadAll(c.Request.Body)
		c.String(200, string(body))
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/test", bytes.NewReader(buf.Bytes()))
	req.Header.Set("Content-Encoding", "gzip")
	r.ServeHTTP(w, req)

	if w.Code != 200 {
		t.Errorf("status = %d, want 200", w.Code)
	}
	if w.Body.String() != "compressed data" {
		t.Errorf("body = %q, want 'compressed data'", w.Body.String())
	}
}

func TestGzipDecompressRequest_InvalidGzip(t *testing.T) {
	r := gin.New()
	r.Use(GzipDecompressRequest())
	r.POST("/test", func(c *gin.Context) {
		c.String(200, "should not reach")
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/test", bytes.NewReader([]byte("not gzip")))
	req.Header.Set("Content-Encoding", "gzip")
	r.ServeHTTP(w, req)

	if w.Code != 400 {
		t.Errorf("status = %d, want 400", w.Code)
	}
}

// ──────────── logger.go 补充 ────────────

func TestShouldSkipAccessLog(t *testing.T) {
	tests := []struct {
		path string
		want bool
	}{
		{"/metrics", true},
		{"/ping", true},
		{"/livez", true},
		{"/readyz", true},
		{"/api/users", false},
		{"/", false},
	}
	for _, tt := range tests {
		got := shouldSkipAccessLog(tt.path)
		if got != tt.want {
			t.Errorf("shouldSkipAccessLog(%q) = %v, want %v", tt.path, got, tt.want)
		}
	}
}

func TestLookupConfig(t *testing.T) {
	root := map[string]interface{}{
		"request_log": map[string]interface{}{
			"skip_paths": []interface{}{"/a", "/b"},
		},
	}

	v, ok := lookupConfig(root, "request_log", "skip_paths")
	if !ok || v == nil {
		t.Fatal("expected to find skip_paths")
	}

	_, ok = lookupConfig(root, "nonexist", "sub")
	if ok {
		t.Error("should not find nonexist")
	}

	// map[interface{}]interface{} 类型
	root2 := map[string]interface{}{
		"logger": map[interface{}]interface{}{
			"skip_paths": []string{"/x"},
		},
	}
	v2, ok2 := lookupConfig(root2, "logger", "skip_paths")
	if !ok2 || v2 == nil {
		t.Fatal("expected to find skip_paths in logger")
	}

	// 不是 map 类型
	root3 := map[string]interface{}{
		"simple": "value",
	}
	_, ok3 := lookupConfig(root3, "simple", "sub")
	if ok3 {
		t.Error("string child should not match")
	}
}

func TestToStringSlice(t *testing.T) {
	// []string
	ss, ok := toStringSlice([]string{"/a", "/b"})
	if !ok || len(ss) != 2 {
		t.Errorf("[]string failed: %v %v", ss, ok)
	}

	// []interface{} with strings
	si, ok := toStringSlice([]interface{}{"/x", "/y"})
	if !ok || len(si) != 2 {
		t.Errorf("[]interface{} failed: %v %v", si, ok)
	}

	// []interface{} with non-strings
	_, ok = toStringSlice([]interface{}{1, 2})
	if ok {
		t.Error("non-string interface slice should fail")
	}

	// other type
	_, ok = toStringSlice(42)
	if ok {
		t.Error("int should fail")
	}
}

func TestGetAccessLogSkipPaths_DefaultPaths(t *testing.T) {
	paths := getAccessLogSkipPaths()
	if len(paths) == 0 {
		t.Error("expected default skip paths")
	}
}

func TestGetAccessLogSkipPaths_WithConfig(t *testing.T) {
	old := app.Config
	app.Config = &app.GlobalConfig{}
	app.Config.Middleware.Config = map[string]interface{}{
		"request_log": map[string]interface{}{
			"skip_paths": []interface{}{"/custom1", "/custom2"},
		},
	}
	defer func() { app.Config = old }()

	paths := getAccessLogSkipPaths()
	if len(paths) != 2 || paths[0] != "/custom1" {
		t.Errorf("unexpected paths: %v", paths)
	}
}

// ──────────── cors.go helper 补充 ────────────

func TestAsStringInterfaceMap(t *testing.T) {
	// map[string]interface{} 直接返回
	m1 := map[string]interface{}{"a": "1"}
	got, ok := asStringInterfaceMap(m1)
	if !ok || got["a"] != "1" {
		t.Errorf("map[string]interface{} failed")
	}

	// map[interface{}]interface{} 需要转换
	m2 := map[interface{}]interface{}{"b": "2"}
	got2, ok2 := asStringInterfaceMap(m2)
	if !ok2 || got2["b"] != "2" {
		t.Errorf("map[interface{}]interface{} failed")
	}

	// 其他类型
	_, ok3 := asStringInterfaceMap("not a map")
	if ok3 {
		t.Error("string should not convert")
	}
}

func TestToBool(t *testing.T) {
	tests := []struct {
		input  interface{}
		want   bool
		wantOK bool
	}{
		{true, true, true},
		{false, false, true},
		{"true", true, true},
		{"false", false, true},
		{1, false, false},
	}
	for _, tt := range tests {
		got, ok := toBool(tt.input)
		if got != tt.want || ok != tt.wantOK {
			t.Errorf("toBool(%v) = (%v, %v), want (%v, %v)", tt.input, got, ok, tt.want, tt.wantOK)
		}
	}
}

func TestToInt(t *testing.T) {
	tests := []struct {
		input  interface{}
		want   int
		wantOK bool
	}{
		{3600, 3600, true},
		{float64(7200), 7200, true},
		{int64(9000), 9000, true},
		{"100", 100, true},
		{"invalid", 0, false},
		{true, 0, false},
	}
	for _, tt := range tests {
		got, ok := toInt(tt.input)
		if got != tt.want || ok != tt.wantOK {
			t.Errorf("toInt(%v) = (%d, %v), want (%d, %v)", tt.input, got, ok, tt.want, tt.wantOK)
		}
	}
}
