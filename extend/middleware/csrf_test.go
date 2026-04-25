package middleware

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"thinkgin/app/ctxkeys"

	"github.com/gin-gonic/gin"
)

func newCSRFRouter() *gin.Engine {
	r := gin.New()
	r.Use(CSRF())
	r.GET("/form", func(c *gin.Context) {
		c.String(http.StatusOK, ctxkeys.GetCSRFToken(c))
	})
	r.POST("/submit", func(c *gin.Context) {
		c.String(http.StatusOK, "ok")
	})
	return r
}

func TestCSRF_SetsCookieOnGET(t *testing.T) {
	r := newCSRFRouter()

	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "/form", nil)
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("GET status = %d, want 200", w.Code)
	}

	cookies := w.Result().Cookies()
	found := false
	for _, c := range cookies {
		if c.Name == csrfCookieName {
			found = true
			if c.Value == "" {
				t.Error("csrf cookie value should not be empty")
			}
		}
	}
	if !found {
		t.Error("csrf cookie not set on GET response")
	}
}

func TestCSRF_BlocksPOSTWithoutToken(t *testing.T) {
	r := newCSRFRouter()

	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodPost, "/submit", nil)
	r.ServeHTTP(w, req)

	if w.Code != http.StatusForbidden {
		t.Errorf("POST without token status = %d, want 403", w.Code)
	}
}

func TestCSRF_AllowsPOSTWithMatchingToken(t *testing.T) {
	r := newCSRFRouter()

	// Step 1: GET 获取 csrf cookie
	w1 := httptest.NewRecorder()
	req1, _ := http.NewRequest(http.MethodGet, "/form", nil)
	r.ServeHTTP(w1, req1)

	var token string
	for _, c := range w1.Result().Cookies() {
		if c.Name == csrfCookieName {
			token = c.Value
		}
	}
	if token == "" {
		t.Fatal("could not get csrf token from GET")
	}

	// Step 2: POST 带上匹配的 cookie + header
	w2 := httptest.NewRecorder()
	req2, _ := http.NewRequest(http.MethodPost, "/submit", nil)
	req2.AddCookie(&http.Cookie{Name: csrfCookieName, Value: token})
	req2.Header.Set(csrfHeaderName, token)
	r.ServeHTTP(w2, req2)

	if w2.Code != http.StatusOK {
		t.Errorf("POST with matching token status = %d, want 200", w2.Code)
	}
}

func TestCSRF_AllowsPOSTWithFormField(t *testing.T) {
	r := newCSRFRouter()

	// Step 1: GET 获取 csrf cookie
	w1 := httptest.NewRecorder()
	req1, _ := http.NewRequest(http.MethodGet, "/form", nil)
	r.ServeHTTP(w1, req1)

	var token string
	for _, c := range w1.Result().Cookies() {
		if c.Name == csrfCookieName {
			token = c.Value
		}
	}

	// Step 2: POST 通过表单字段提交 token
	body := strings.NewReader(csrfFormField + "=" + token)
	w2 := httptest.NewRecorder()
	req2, _ := http.NewRequest(http.MethodPost, "/submit", body)
	req2.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req2.AddCookie(&http.Cookie{Name: csrfCookieName, Value: token})
	r.ServeHTTP(w2, req2)

	if w2.Code != http.StatusOK {
		t.Errorf("POST with form token status = %d, want 200", w2.Code)
	}
}

func TestCSRF_BlocksMismatchedToken(t *testing.T) {
	r := newCSRFRouter()

	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodPost, "/submit", nil)
	req.AddCookie(&http.Cookie{Name: csrfCookieName, Value: "real-token"})
	req.Header.Set(csrfHeaderName, "wrong-token")
	r.ServeHTTP(w, req)

	if w.Code != http.StatusForbidden {
		t.Errorf("POST with mismatched token status = %d, want 403", w.Code)
	}
}
