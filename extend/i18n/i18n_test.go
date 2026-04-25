package i18n

import (
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/gin-gonic/gin"
)

func setupTestBundle(t *testing.T) (*Bundle, string) {
	t.Helper()
	dir := t.TempDir()

	zhYAML := []byte("welcome: \"欢迎, {{.Name}}!\"\nhello: \"你好\"\n")
	enYAML := []byte("welcome: \"Welcome, {{.Name}}!\"\nhello: \"Hello\"\n")

	os.WriteFile(filepath.Join(dir, "zh-CN.yaml"), zhYAML, 0644)
	os.WriteFile(filepath.Join(dir, "en.yaml"), enYAML, 0644)

	b := NewBundle(dir, "zh-CN")
	if err := b.LoadAll(); err != nil {
		t.Fatalf("LoadAll: %v", err)
	}
	return b, dir
}

func TestBundle_TranslateWithParams(t *testing.T) {
	b, _ := setupTestBundle(t)

	got := b.Translate("zh-CN", "welcome", map[string]string{"Name": "张三"})
	want := "欢迎, 张三!"
	if got != want {
		t.Errorf("Translate = %q, want %q", got, want)
	}
}

func TestBundle_FallbackLanguage(t *testing.T) {
	b, _ := setupTestBundle(t)

	// 不存在的 locale 应回退到 fallback
	got := b.Translate("fr", "hello", nil)
	if got != "你好" {
		t.Errorf("fallback = %q, want 你好", got)
	}
}

func TestBundle_MissingKeyReturnsKey(t *testing.T) {
	b, _ := setupTestBundle(t)

	got := b.Translate("zh-CN", "nonexistent", nil)
	if got != "nonexistent" {
		t.Errorf("missing key = %q, want 'nonexistent'", got)
	}
}

func TestMiddleware_DetectsFromQuery(t *testing.T) {
	gin.SetMode(gin.TestMode)
	b, _ := setupTestBundle(t)

	r := gin.New()
	r.Use(Middleware(b))
	r.GET("/test", func(c *gin.Context) {
		msg := T(c, "hello")
		c.String(http.StatusOK, msg)
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "/test?lang=en", nil)
	r.ServeHTTP(w, req)

	if w.Body.String() != "Hello" {
		t.Errorf("body = %q, want Hello", w.Body.String())
	}
}

func TestMiddleware_DetectsFromAcceptLanguage(t *testing.T) {
	gin.SetMode(gin.TestMode)
	b, _ := setupTestBundle(t)

	r := gin.New()
	r.Use(Middleware(b))
	r.GET("/test", func(c *gin.Context) {
		msg := T(c, "hello")
		c.String(http.StatusOK, msg)
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "/test", nil)
	req.Header.Set("Accept-Language", "en,zh-CN;q=0.9")
	r.ServeHTTP(w, req)

	if w.Body.String() != "Hello" {
		t.Errorf("body = %q, want Hello", w.Body.String())
	}
}

func TestParseAcceptLanguage(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"zh-CN,zh;q=0.9,en;q=0.8", "zh-CN"},
		{"en", "en"},
		{"", ""},
		{"fr;q=0.5, en;q=0.9", "fr"},
	}
	for _, tt := range tests {
		got := parseAcceptLanguage(tt.input)
		if got != tt.want {
			t.Errorf("parseAcceptLanguage(%q) = %q, want %q", tt.input, got, tt.want)
		}
	}
}
