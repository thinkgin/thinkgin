package middleware

import (
	"compress/gzip"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestGzip_CompressesWhenAccepted(t *testing.T) {
	r := gin.New()
	r.Use(Gzip())
	r.GET("/big", func(c *gin.Context) {
		// 写一个大于 gzipMinSize 的响应
		c.String(http.StatusOK, strings.Repeat("hello world ", 200))
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "/big", nil)
	req.Header.Set("Accept-Encoding", "gzip, deflate")
	r.ServeHTTP(w, req)

	if w.Header().Get("Content-Encoding") != "gzip" {
		t.Errorf("Content-Encoding = %q, want gzip", w.Header().Get("Content-Encoding"))
	}

	// 验证能正常解压
	reader, err := gzip.NewReader(w.Body)
	if err != nil {
		t.Fatalf("failed to create gzip reader: %v", err)
	}
	defer reader.Close()

	body, _ := io.ReadAll(reader)
	if !strings.Contains(string(body), "hello world") {
		t.Errorf("decompressed body should contain 'hello world'")
	}
}

func TestGzip_SkipsWithoutAcceptEncoding(t *testing.T) {
	r := gin.New()
	r.Use(Gzip())
	r.GET("/plain", func(c *gin.Context) {
		c.String(http.StatusOK, "no compress")
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "/plain", nil)
	// 不设置 Accept-Encoding
	r.ServeHTTP(w, req)

	if w.Header().Get("Content-Encoding") == "gzip" {
		t.Error("should not compress when Accept-Encoding is absent")
	}
	if w.Body.String() != "no compress" {
		t.Errorf("body = %q, want 'no compress'", w.Body.String())
	}
}
