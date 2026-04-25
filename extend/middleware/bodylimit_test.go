package middleware

import (
	"bytes"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestBodyLimit_SmallRequestPasses(t *testing.T) {
	r := gin.New()
	r.Use(BodyLimitWithSize(1024))
	r.POST("/test", func(c *gin.Context) {
		body, _ := io.ReadAll(c.Request.Body)
		c.String(200, string(body))
	})

	payload := bytes.Repeat([]byte("a"), 100)
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/test", bytes.NewReader(payload))
	req.Header.Set("Content-Type", "text/plain")
	r.ServeHTTP(w, req)

	if w.Code != 200 {
		t.Errorf("status = %d, want 200", w.Code)
	}
}

func TestBodyLimit_LargeContentLengthRejected(t *testing.T) {
	r := gin.New()
	r.Use(BodyLimitWithSize(1024))
	r.POST("/test", func(c *gin.Context) {
		c.String(200, "should not reach")
	})

	payload := bytes.Repeat([]byte("a"), 2048)
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/test", bytes.NewReader(payload))
	req.ContentLength = 2048
	r.ServeHTTP(w, req)

	if w.Code != http.StatusRequestEntityTooLarge {
		t.Errorf("status = %d, want 413", w.Code)
	}
}

func TestBodyLimit_NilBodyPasses(t *testing.T) {
	r := gin.New()
	r.Use(BodyLimitWithSize(1024))
	r.GET("/test", func(c *gin.Context) {
		c.String(200, "ok")
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/test", nil)
	r.ServeHTTP(w, req)

	if w.Code != 200 {
		t.Errorf("status = %d, want 200", w.Code)
	}
}

func TestBodyLimit_Default(t *testing.T) {
	r := gin.New()
	r.Use(BodyLimit())
	r.POST("/test", func(c *gin.Context) {
		c.String(200, "ok")
	})

	// 小于 10MB，应通过
	payload := bytes.Repeat([]byte("x"), 1024)
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/test", bytes.NewReader(payload))
	r.ServeHTTP(w, req)

	if w.Code != 200 {
		t.Errorf("status = %d, want 200", w.Code)
	}
}

func TestParseBodyLimit(t *testing.T) {
	tests := []struct {
		input string
		want  int64
		err   bool
	}{
		{"10MB", 10 << 20, false},
		{"512KB", 512 << 10, false},
		{"1GB", 1 << 30, false},
		{"100B", 100, false},
		{"0.5MB", int64(0.5 * float64(1<<20)), false},
		{"", 0, true},
		{"-1MB", 0, true},
		{"abc", 0, true},
	}
	for _, tt := range tests {
		got, err := ParseBodyLimit(tt.input)
		if (err != nil) != tt.err {
			t.Errorf("ParseBodyLimit(%q) err=%v, wantErr=%v", tt.input, err, tt.err)
			continue
		}
		if err == nil && got != tt.want {
			t.Errorf("ParseBodyLimit(%q) = %d, want %d", tt.input, got, tt.want)
		}
	}
}

func TestFormatBytes(t *testing.T) {
	tests := []struct {
		input int64
		want  string
	}{
		{500, "500B"},
		{1024, "1.0KB"},
		{10 << 20, "10.0MB"},
		{1 << 30, "1.0GB"},
	}
	for _, tt := range tests {
		got := formatBytes(tt.input)
		if got != tt.want {
			t.Errorf("formatBytes(%d) = %q, want %q", tt.input, got, tt.want)
		}
	}
}

func TestBodyLimitFromString(t *testing.T) {
	h, err := BodyLimitFromString("1MB")
	if err != nil {
		t.Fatalf("BodyLimitFromString: %v", err)
	}
	if h == nil {
		t.Fatal("expected non-nil handler")
	}

	_, err = BodyLimitFromString("invalid")
	if err == nil {
		t.Fatal("expected error for invalid string")
	}
}
