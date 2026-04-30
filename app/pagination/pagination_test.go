package pagination

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"thinkgin/app"

	"github.com/gin-gonic/gin"
)

func init() {
	gin.SetMode(gin.TestMode)
}

func setupConfig(pageSize, maxPageSize int) func() {
	old := app.Config
	app.Config = &app.GlobalConfig{}
	app.Config.App.Pagination.PageSize = pageSize
	app.Config.App.Pagination.MaxPageSize = maxPageSize
	return func() { app.Config = old }
}

func TestFromQuery_Defaults(t *testing.T) {
	restore := setupConfig(20, 100)
	defer restore()

	r := gin.New()
	var p Params
	r.GET("/", func(c *gin.Context) {
		p = FromQuery(c)
		c.Status(200)
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/", nil)
	r.ServeHTTP(w, req)

	if p.Page != 1 {
		t.Errorf("Page=%d, want 1", p.Page)
	}
	if p.PageSize != 20 {
		t.Errorf("PageSize=%d, want 20", p.PageSize)
	}
}

func TestFromQuery_CustomParams(t *testing.T) {
	restore := setupConfig(20, 100)
	defer restore()

	r := gin.New()
	var p Params
	r.GET("/", func(c *gin.Context) {
		p = FromQuery(c)
		c.Status(200)
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/?page=3&page_size=50", nil)
	r.ServeHTTP(w, req)

	if p.Page != 3 {
		t.Errorf("Page=%d, want 3", p.Page)
	}
	if p.PageSize != 50 {
		t.Errorf("PageSize=%d, want 50", p.PageSize)
	}
}

func TestFromQuery_CapsAtMaxPageSize(t *testing.T) {
	restore := setupConfig(20, 100)
	defer restore()

	r := gin.New()
	var p Params
	r.GET("/", func(c *gin.Context) {
		p = FromQuery(c)
		c.Status(200)
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/?page_size=999", nil)
	r.ServeHTTP(w, req)

	if p.PageSize != 100 {
		t.Errorf("PageSize=%d, want 100 (max)", p.PageSize)
	}
}

func TestFromQuery_NegativePageResets(t *testing.T) {
	restore := setupConfig(20, 100)
	defer restore()

	r := gin.New()
	var p Params
	r.GET("/", func(c *gin.Context) {
		p = FromQuery(c)
		c.Status(200)
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/?page=-5", nil)
	r.ServeHTTP(w, req)

	if p.Page != 1 {
		t.Errorf("Page=%d, want 1", p.Page)
	}
}

func TestParams_Offset(t *testing.T) {
	p := Params{Page: 3, PageSize: 20}
	if p.Offset() != 40 {
		t.Errorf("Offset()=%d, want 40", p.Offset())
	}
}

func TestNewResult(t *testing.T) {
	p := Params{Page: 2, PageSize: 10}
	r := NewResult(p, 25, []string{"a", "b"})

	if r.Total != 25 {
		t.Errorf("Total=%d, want 25", r.Total)
	}
	if r.TotalPages != 3 {
		t.Errorf("TotalPages=%d, want 3", r.TotalPages)
	}
	if r.Page != 2 {
		t.Errorf("Page=%d, want 2", r.Page)
	}
}

func TestNewResult_ZeroPageSize(t *testing.T) {
	p := Params{Page: 1, PageSize: 0}
	r := NewResult(p, 10, nil)
	if r.TotalPages != 0 {
		t.Errorf("TotalPages=%d, want 0", r.TotalPages)
	}
}
