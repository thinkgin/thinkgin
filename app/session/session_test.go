package session

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"

	"thinkgin/app"
)

func init() { gin.SetMode(gin.TestMode) }

// configureMemory 把全局配置重置为 memory driver，保证测试彼此独立。
func configureMemory(t *testing.T, lifetime int) {
	t.Helper()
	app.Config = &app.GlobalConfig{}
	app.Config.Session.Driver = "memory"
	app.Config.Session.Name = "TG_SESSION"
	app.Config.Session.Lifetime = lifetime
	app.Config.Session.Cookie.Path = "/"
	app.Config.Session.Cookie.HttpOnly = true
	app.Config.Session.Cookie.SameSite = "lax"
	if err := Init(); err != nil {
		t.Fatalf("Init(): %v", err)
	}
	t.Cleanup(func() { _ = Shutdown() })
}

func TestGenerateID_IsStableURLSafe(t *testing.T) {
	id, err := generateID()
	if err != nil {
		t.Fatal(err)
	}
	if len(id) < 40 {
		t.Errorf("id too short: %q", id)
	}
	for _, r := range id {
		if !isSessionIDChar(r) {
			t.Errorf("unexpected char %q in %q", r, id)
		}
	}
}

func TestInit_UnknownDriverReturnsError(t *testing.T) {
	app.Config = &app.GlobalConfig{}
	app.Config.Session.Driver = "oracle"
	if err := Init(); err == nil {
		t.Fatal("expected error for unknown driver")
	}
	_ = Shutdown()
}

func TestInit_FileDriverNotImplemented(t *testing.T) {
	app.Config = &app.GlobalConfig{}
	app.Config.Session.Driver = "file"
	if err := Init(); err == nil {
		t.Fatal("expected file driver to report unimplemented")
	}
	_ = Shutdown()
}

func TestMiddleware_CreatesNewSessionCookie(t *testing.T) {
	configureMemory(t, 3600)

	r := gin.New()
	r.Use(Middleware())
	r.GET("/ping", func(c *gin.Context) {
		s := From(c)
		if s == nil {
			t.Fatal("session should be present")
		}
		c.String(200, s.ID())
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "/ping", nil)
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("status=%d", w.Code)
	}
	cookie := w.Header().Get("Set-Cookie")
	if !strings.Contains(cookie, "TG_SESSION=") {
		t.Errorf("expected Set-Cookie with session name, got %q", cookie)
	}
	if !strings.Contains(cookie, "HttpOnly") {
		t.Errorf("cookie should have HttpOnly flag, got %q", cookie)
	}
}

func TestMiddleware_ReusesSessionFromCookie(t *testing.T) {
	configureMemory(t, 3600)

	r := gin.New()
	r.Use(Middleware())
	r.GET("/get", func(c *gin.Context) {
		if v, ok := From(c).Get("name"); ok {
			c.String(200, v.(string))
			return
		}
		c.String(200, "")
	})
	r.POST("/set", func(c *gin.Context) {
		s := From(c)
		s.Set("name", "alice")
		if err := s.Save(c.Request.Context()); err != nil {
			c.String(500, err.Error())
			return
		}
		c.String(200, s.ID())
	})

	// 第一次 POST：创建并写入
	w1 := httptest.NewRecorder()
	req1, _ := http.NewRequest(http.MethodPost, "/set", nil)
	r.ServeHTTP(w1, req1)
	sessionCookie := extractCookie(w1.Header().Get("Set-Cookie"))
	if sessionCookie == "" {
		t.Fatal("expected session cookie")
	}

	// 第二次 GET：携带 cookie，应拿回 alice
	w2 := httptest.NewRecorder()
	req2, _ := http.NewRequest(http.MethodGet, "/get", nil)
	req2.AddCookie(&http.Cookie{Name: "TG_SESSION", Value: sessionCookie})
	r.ServeHTTP(w2, req2)
	if body := w2.Body.String(); body != "alice" {
		t.Errorf("expected alice, got %q", body)
	}
}

func TestSession_SaveNoopWhenNotDirty(t *testing.T) {
	configureMemory(t, 3600)
	s := &Session{id: "abc", data: map[string]any{"k": "v"}}
	if err := s.Save(context.Background()); err != nil {
		t.Fatalf("non-dirty save should succeed, got %v", err)
	}
	// 直接读 store 应不存在该 id。
	m := current()
	data, _ := m.store.Load(context.Background(), "abc")
	if len(data) != 0 {
		t.Errorf("non-dirty session should not be persisted, got %v", data)
	}
}

func TestSession_DestroyClearsStore(t *testing.T) {
	configureMemory(t, 3600)
	ctx := context.Background()
	s := &Session{id: "abc"}
	s.Set("x", 1)
	if err := s.Save(ctx); err != nil {
		t.Fatal(err)
	}
	if err := s.Destroy(ctx); err != nil {
		t.Fatal(err)
	}
	m := current()
	data, _ := m.store.Load(ctx, "abc")
	if len(data) != 0 {
		t.Errorf("after Destroy, store should be empty, got %v", data)
	}
}

func TestMemoryStore_ExpiredEntryReturnsEmpty(t *testing.T) {
	ms := newMemoryStore(time.Hour)
	defer ms.Close()
	ctx := context.Background()
	// 手动塞一条已过期的条目
	ms.mu.Lock()
	ms.data["old"] = memoryEntry{
		data:      map[string]any{"k": "v"},
		expiresAt: time.Now().Add(-time.Minute),
	}
	ms.mu.Unlock()

	got, err := ms.Load(ctx, "old")
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 0 {
		t.Errorf("expired entry should be treated as empty, got %v", got)
	}
}

func TestParseSameSite(t *testing.T) {
	cases := map[string]http.SameSite{
		"lax":     http.SameSiteLaxMode,
		"strict":  http.SameSiteStrictMode,
		"none":    http.SameSiteNoneMode,
		"":        http.SameSiteLaxMode,
		"unknown": http.SameSiteLaxMode,
	}
	for in, want := range cases {
		if got := parseSameSite(in); got != want {
			t.Errorf("parseSameSite(%q)=%v, want %v", in, got, want)
		}
	}
}

// extractCookie 从 Set-Cookie 头里取出目标 cookie 的 value。
func extractCookie(header string) string {
	// 形如 "TG_SESSION=xxx; Path=/; HttpOnly"
	parts := strings.Split(header, ";")
	if len(parts) == 0 {
		return ""
	}
	kv := strings.SplitN(strings.TrimSpace(parts[0]), "=", 2)
	if len(kv) != 2 || kv[0] != "TG_SESSION" {
		return ""
	}
	return kv[1]
}
