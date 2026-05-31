package controller

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/glebarez/sqlite"
	"gorm.io/gorm"

	"thinkgin/app/article/model"
	"thinkgin/app/article/repository"
	"thinkgin/app/article/service"
)

func init() { gin.SetMode(gin.TestMode) }

// newTestRouter 构造一个仅含 article 路由的引擎，使用私有内存 SQLite。
// 测试不挂载 JWT，直接验证 controller 与 service 的协作。
func newTestRouter(t *testing.T) *gin.Engine {
	t.Helper()
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}
	sqlDB, err := db.DB()
	if err != nil {
		t.Fatalf("db handle: %v", err)
	}
	sqlDB.SetMaxOpenConns(1)
	t.Cleanup(func() { _ = sqlDB.Close() })

	if err := db.AutoMigrate(&model.Article{}); err != nil {
		t.Fatalf("automigrate: %v", err)
	}

	ctl := New(service.New(repository.New(db)))
	r := gin.New()
	g := r.Group("/api/v1/articles")
	g.GET("", ctl.List)
	g.GET("/:id", ctl.Get)
	g.POST("", ctl.Create)
	g.PUT("/:id", ctl.Update)
	g.POST("/:id/publish", ctl.Publish)
	g.DELETE("/:id", ctl.Delete)
	return r
}

func doJSON(r *gin.Engine, method, path, body string) *httptest.ResponseRecorder {
	w := httptest.NewRecorder()
	var rdr *bytes.Buffer
	if body != "" {
		rdr = bytes.NewBufferString(body)
	} else {
		rdr = bytes.NewBuffer(nil)
	}
	req, _ := http.NewRequest(method, path, rdr)
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)
	return w
}

func TestController_CreateAndGet(t *testing.T) {
	r := newTestRouter(t)

	w := doJSON(r, "POST", "/api/v1/articles", `{"title":"Hello","content":"body","author":"alice"}`)
	if w.Code != http.StatusOK {
		t.Fatalf("create status = %d, body=%s", w.Code, w.Body.String())
	}

	var resp struct {
		Code int `json:"code"`
		Data struct {
			ID     uint   `json:"ID"`
			Title  string `json:"title"`
			Status string `json:"status"`
		} `json:"data"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if resp.Data.ID == 0 || resp.Data.Title != "Hello" {
		t.Errorf("unexpected create data: %+v", resp.Data)
	}

	// 读取刚创建的文章
	w = doJSON(r, "GET", "/api/v1/articles/1", "")
	if w.Code != http.StatusOK {
		t.Errorf("get status = %d, body=%s", w.Code, w.Body.String())
	}
}

func TestController_CreateValidationError(t *testing.T) {
	r := newTestRouter(t)
	// 缺 title，应被 BindAndValidate 拦截为 422。
	w := doJSON(r, "POST", "/api/v1/articles", `{"content":"no title"}`)
	if w.Code != http.StatusUnprocessableEntity {
		t.Errorf("status = %d, want 422, body=%s", w.Code, w.Body.String())
	}
}

func TestController_GetNotFound(t *testing.T) {
	r := newTestRouter(t)
	w := doJSON(r, "GET", "/api/v1/articles/999", "")
	if w.Code != http.StatusNotFound {
		t.Errorf("status = %d, want 404, body=%s", w.Code, w.Body.String())
	}
}

func TestController_BadID(t *testing.T) {
	r := newTestRouter(t)
	w := doJSON(r, "GET", "/api/v1/articles/abc", "")
	if w.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want 400, body=%s", w.Code, w.Body.String())
	}
}

func TestController_PublishConflict(t *testing.T) {
	r := newTestRouter(t)
	_ = doJSON(r, "POST", "/api/v1/articles", `{"title":"x"}`)

	w := doJSON(r, "POST", "/api/v1/articles/1/publish", "")
	if w.Code != http.StatusOK {
		t.Fatalf("first publish status = %d, body=%s", w.Code, w.Body.String())
	}
	// 二次发布 → 409
	w = doJSON(r, "POST", "/api/v1/articles/1/publish", "")
	if w.Code != http.StatusConflict {
		t.Errorf("second publish status = %d, want 409", w.Code)
	}
}

func TestController_ListPagination(t *testing.T) {
	r := newTestRouter(t)
	for i := 0; i < 3; i++ {
		_ = doJSON(r, "POST", "/api/v1/articles", `{"title":"a"}`)
	}
	w := doJSON(r, "GET", "/api/v1/articles?page=1&page_size=2", "")
	if w.Code != http.StatusOK {
		t.Fatalf("list status = %d", w.Code)
	}
	var resp struct {
		Data struct {
			Total int64 `json:"total"`
			List  []any `json:"list"`
		} `json:"data"`
	}
	_ = json.Unmarshal(w.Body.Bytes(), &resp)
	if resp.Data.Total != 3 {
		t.Errorf("total = %d, want 3", resp.Data.Total)
	}
	if len(resp.Data.List) != 2 {
		t.Errorf("page size = %d, want 2", len(resp.Data.List))
	}
}

func TestController_DeleteFlow(t *testing.T) {
	r := newTestRouter(t)
	_ = doJSON(r, "POST", "/api/v1/articles", `{"title":"doomed"}`)

	w := doJSON(r, "DELETE", "/api/v1/articles/1", "")
	if w.Code != http.StatusOK {
		t.Fatalf("delete status = %d, body=%s", w.Code, w.Body.String())
	}
	// 删除后再读应 404
	w = doJSON(r, "GET", "/api/v1/articles/1", "")
	if w.Code != http.StatusNotFound {
		t.Errorf("after delete status = %d, want 404", w.Code)
	}
}
