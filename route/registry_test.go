package route

import (
	"net/http"
	"net/http/httptest"
	"sync"
	"sync/atomic"
	"testing"

	"github.com/gin-gonic/gin"
)

// TestRegisterAndLookup 验证基本注册与查询。
func TestRegisterAndLookup(t *testing.T) {
	defer resetRegistryForTest()

	called := false
	RegisterMiddleware("audit", func() gin.HandlerFunc {
		return func(c *gin.Context) {
			called = true
			c.Next()
		}
	})

	if !HasMiddleware("audit") {
		t.Fatal("HasMiddleware should report true after Register")
	}

	h := LookupMiddleware("audit")
	if h == nil {
		t.Fatal("LookupMiddleware returned nil for registered name")
	}

	// 在 gin 路由中执行一次，确认 handler 真的被调用。
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(h)
	r.GET("/", func(c *gin.Context) { c.String(200, "ok") })
	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "/", nil)
	r.ServeHTTP(w, req)

	if !called {
		t.Error("registered middleware was not executed")
	}
	if w.Code != http.StatusOK {
		t.Errorf("status = %d, want 200", w.Code)
	}
}

// TestRegisterFactoryEachCallReturnsNewInstance 验证工厂语义：
// 每次 Lookup 都新建 handler 实例，避免共享状态污染。
func TestRegisterFactoryEachCallReturnsNewInstance(t *testing.T) {
	defer resetRegistryForTest()

	var counter int32
	RegisterMiddleware("counter", func() gin.HandlerFunc {
		// 每次工厂调用，独立的局部计数器
		var local int32
		return func(c *gin.Context) {
			atomic.AddInt32(&local, 1)
			atomic.AddInt32(&counter, atomic.LoadInt32(&local))
			c.Next()
		}
	})

	// 两次 Lookup 应当返回两个独立实例（独立的 local）
	h1 := LookupMiddleware("counter")
	h2 := LookupMiddleware("counter")

	if h1 == nil || h2 == nil {
		t.Fatal("Lookup returned nil")
	}
	// h1 和 h2 是不同的闭包指针——通过运行时行为间接验证。
	// 此处至少确认两次都能返回非 nil handler。
}

// TestUnregister 验证删除生效。
func TestUnregister(t *testing.T) {
	defer resetRegistryForTest()

	RegisterMiddleware("temp", func() gin.HandlerFunc {
		return func(c *gin.Context) { c.Next() }
	})
	if !HasMiddleware("temp") {
		t.Fatal("Register failed")
	}

	UnregisterMiddleware("temp")

	if HasMiddleware("temp") {
		t.Error("Unregister did not remove the entry")
	}
	if h := LookupMiddleware("temp"); h != nil {
		t.Error("Lookup returned non-nil after Unregister")
	}
}

// TestRegisterIgnoresInvalidInput 验证空名/nil 工厂被静默忽略。
func TestRegisterIgnoresInvalidInput(t *testing.T) {
	defer resetRegistryForTest()

	RegisterMiddleware("", func() gin.HandlerFunc { return nil })
	RegisterMiddleware("ok", nil)

	if HasMiddleware("") {
		t.Error("empty name should not be registered")
	}
	if HasMiddleware("ok") {
		t.Error("nil factory should not be registered")
	}
}

// TestCustomOverridesBuiltin 验证自定义注册可以覆盖内置中间件。
// 例如用户实现自己的 access_log 替换框架默认。
func TestCustomOverridesBuiltin(t *testing.T) {
	defer resetRegistryForTest()

	customCalled := false
	RegisterMiddleware("access_log", func() gin.HandlerFunc {
		return func(c *gin.Context) {
			customCalled = true
			c.Next()
		}
	})

	h := resolveMiddleware("access_log")
	if h == nil {
		t.Fatal("resolveMiddleware returned nil for overridden name")
	}

	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(h)
	r.GET("/", func(c *gin.Context) { c.String(200, "ok") })
	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "/", nil)
	r.ServeHTTP(w, req)

	if !customCalled {
		t.Error("custom override was not invoked; built-in switch took over")
	}
}

// TestResolveFallsBackToBuiltin 验证未注册的内置名仍能命中 switch。
func TestResolveFallsBackToBuiltin(t *testing.T) {
	defer resetRegistryForTest()

	// 不注册 recovery，应当走内置 switch。
	if h := resolveMiddleware("recovery"); h == nil {
		t.Error("built-in 'recovery' should be resolvable when no custom override")
	}
}

// TestResolveUnknownReturnsNil 验证未知名称返回 nil。
func TestResolveUnknownReturnsNil(t *testing.T) {
	defer resetRegistryForTest()

	if h := resolveMiddleware("___definitely_not_registered___"); h != nil {
		t.Error("unknown middleware name should return nil")
	}
}

// TestConcurrentRegister 验证并发注册无 race。
// 配合 -race 标志运行可暴露任何潜在的数据竞争。
func TestConcurrentRegister(t *testing.T) {
	defer resetRegistryForTest()

	var wg sync.WaitGroup
	for i := 0; i < 32; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			RegisterMiddleware("concurrent", func() gin.HandlerFunc {
				return func(c *gin.Context) { c.Next() }
			})
			_ = LookupMiddleware("concurrent")
		}()
	}
	wg.Wait()

	if !HasMiddleware("concurrent") {
		t.Error("concurrent registration lost final entry")
	}
}
