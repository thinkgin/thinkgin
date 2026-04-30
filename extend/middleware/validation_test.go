package middleware

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

type testBindReq struct {
	Name  string `json:"name"  binding:"required,min=2,max=50"`
	Email string `json:"email" binding:"required,email"`
	Age   int    `json:"age"   binding:"gte=0,lte=150"`
}

func TestBindAndValidate_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.POST("/test", func(c *gin.Context) {
		var req testBindReq
		if err := BindAndValidate(c, &req); err != nil {
			return
		}
		c.JSON(http.StatusOK, gin.H{"name": req.Name})
	})

	body := `{"name":"Alice","email":"alice@example.com","age":25}`
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/test", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("status=%d, want 200, body=%s", w.Code, w.Body.String())
	}
}

func TestBindAndValidate_ValidationError(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.POST("/test", func(c *gin.Context) {
		var req testBindReq
		if err := BindAndValidate(c, &req); err != nil {
			return
		}
		c.JSON(http.StatusOK, gin.H{})
	})

	body := `{"name":"A","email":"not-email","age":-1}`
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/test", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	if w.Code != http.StatusUnprocessableEntity {
		t.Fatalf("status=%d, want 422, body=%s", w.Code, w.Body.String())
	}

	var resp map[string]interface{}
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if resp["code"].(float64) != 422000 {
		t.Errorf("code=%v, want 422000", resp["code"])
	}
	data, ok := resp["data"].([]interface{})
	if !ok || len(data) == 0 {
		t.Error("data should contain field errors")
	}
}

func TestBindAndValidate_MissingRequired(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.POST("/test", func(c *gin.Context) {
		var req testBindReq
		if err := BindAndValidate(c, &req); err != nil {
			return
		}
		c.JSON(http.StatusOK, gin.H{})
	})

	body := `{}`
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/test", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	if w.Code != http.StatusUnprocessableEntity {
		t.Fatalf("status=%d, want 422", w.Code)
	}
}

func TestBindAndValidate_BadJSON(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.POST("/test", func(c *gin.Context) {
		var req testBindReq
		if err := BindAndValidate(c, &req); err != nil {
			return
		}
		c.JSON(http.StatusOK, gin.H{})
	})

	body := `{invalid json`
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/test", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("status=%d, want 400", w.Code)
	}
}
