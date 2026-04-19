package route

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

func init() { gin.SetMode(gin.TestMode) }

func doGet(t *testing.T, h gin.HandlerFunc, path string) *httptest.ResponseRecorder {
	t.Helper()
	r := gin.New()
	r.GET(path, h)
	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, path, nil)
	r.ServeHTTP(w, req)
	return w
}

func TestLivez_AlwaysReturnsOK(t *testing.T) {
	w := doGet(t, livezHandler(), "/livez")
	if w.Code != http.StatusOK {
		t.Fatalf("livez code=%d, want 200", w.Code)
	}
}

func TestReadyz_WithoutDependenciesReturns200(t *testing.T) {
	// 未调用 database.Init / cache.Init 时，Default() 均返回 nil。
	// readyz 应视为"无依赖需要探测"，返回 200。
	w := doGet(t, readyzHandler(), "/readyz")
	if w.Code != http.StatusOK {
		t.Fatalf("readyz code=%d body=%s", w.Code, w.Body.String())
	}
	var payload struct {
		Status     string                 `json:"status"`
		Components map[string]interface{} `json:"components"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &payload); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if payload.Status != "ok" {
		t.Errorf("status=%q, want ok", payload.Status)
	}
	if len(payload.Components) != 0 {
		t.Errorf("expected no components, got %v", payload.Components)
	}
}
