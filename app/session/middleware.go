// 本文件提供 Gin 中间件与 Session 实例类型。
// Session 实例保存在 gin.Context 的 Keys 中，业务通过 session.From(c) 取用。
package session

import (
	"context"
	"net/http"
	"strings"
	"sync"

	"github.com/gin-gonic/gin"
)

// ctxKey 在 gin.Context.Keys 中标识 Session 实例；使用未导出常量避免外部覆盖。
const ctxKey = "thinkgin:session"

// Session 是请求内的会话实例。非并发安全：每个请求一个 Session，不应跨 goroutine 共享。
// 若业务确实需要 Copy 到异步任务，应先 Save 再在新 goroutine 里重新 Load。
type Session struct {
	id    string
	data  map[string]any
	dirty bool
	mu    sync.Mutex // 保护并发写（同一请求内也可能被多个 handler 并发调用）
}

// ID 返回当前 Session 的唯一 ID。
func (s *Session) ID() string { return s.id }

// Get 按 key 读取值。不存在时 ok=false。
func (s *Session) Get(key string) (any, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	v, ok := s.data[key]
	return v, ok
}

// Set 写入 key=value，并标记为 dirty。真正持久化发生在 Save 时。
func (s *Session) Set(key string, value any) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.data == nil {
		s.data = map[string]any{}
	}
	s.data[key] = value
	s.dirty = true
}

// Delete 删除 key。无论 key 是否存在都标记 dirty，避免漏写。
func (s *Session) Delete(key string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.data, key)
	s.dirty = true
}

// Save 把脏数据写回存储。未修改时直接返回 nil，避免无意义 I/O。
// 业务可在 handler 末尾手动调用；中间件不强制在请求结束时自动保存，
// 因为自动保存会遮蔽业务错误处理路径。
func (s *Session) Save(ctx context.Context) error {
	s.mu.Lock()
	dirty := s.dirty
	data := make(map[string]any, len(s.data))
	for k, v := range s.data {
		data[k] = v
	}
	s.mu.Unlock()

	if !dirty {
		return nil
	}
	m := current()
	if m == nil {
		return ErrNotInitialized
	}
	if err := m.store.Save(ctx, s.id, data, m.lifetime); err != nil {
		return err
	}
	s.mu.Lock()
	s.dirty = false
	s.mu.Unlock()
	return nil
}

// Destroy 删除存储中的 Session 并清空本地数据。
// 调用方应额外清除客户端 Cookie（见 ClearCookie）。
func (s *Session) Destroy(ctx context.Context) error {
	m := current()
	if m == nil {
		return ErrNotInitialized
	}
	if err := m.store.Destroy(ctx, s.id); err != nil {
		return err
	}
	s.mu.Lock()
	s.data = map[string]any{}
	s.dirty = false
	s.mu.Unlock()
	return nil
}

// Middleware 返回 Gin 中间件：
//   - 从 Cookie 读取 Session ID，缺失或非法时签发新 ID
//   - 从 Store 加载数据并附着到 gin.Context
//   - 把 Session Cookie 写回客户端（每次都写，确保刷新过期时间）
//
// 未初始化时返回 no-op 中间件，避免破坏既有路由。
func Middleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		m := current()
		if m == nil {
			c.Next()
			return
		}

		id := readIDFromCookie(c, m.name)
		isNew := id == ""
		if isNew {
			newID, err := generateID()
			if err != nil {
				// 生成失败属于系统级错误；放行请求但不挂载 Session，避免 500。
				c.Next()
				return
			}
			id = newID
		}

		data, err := m.store.Load(c.Request.Context(), id)
		if err != nil {
			// 加载失败视为新会话：数据退化为空 map，Cookie 继续走写回流程。
			// 此处无需再标记 isNew，因为后续只依赖 id 与 data。
			data = map[string]any{}
		}

		s := &Session{id: id, data: data}
		c.Set(ctxKey, s)

		// 始终写回 Cookie：即使 session 未修改，也能续期客户端的过期时间。
		writeCookie(c, m, id)

		c.Next()
	}
}

// From 从 gin.Context 取出 Session 实例。
// 未挂载（例如 Middleware 未启用）时返回 nil，调用方应判空。
func From(c *gin.Context) *Session {
	v, ok := c.Get(ctxKey)
	if !ok {
		return nil
	}
	s, _ := v.(*Session)
	return s
}

// ClearCookie 用于登出：立即过期客户端的 Session Cookie。
// 本函数不会主动 Destroy 存储端数据，业务应在此前调用 Session.Destroy。
func ClearCookie(c *gin.Context) {
	m := current()
	if m == nil {
		return
	}
	c.SetCookie(m.name, "", -1, m.cookie.Path, m.cookie.Domain, m.cookie.Secure, m.cookie.HttpOnly)
}

// readIDFromCookie 读取并校验 Cookie 中的 Session ID。
// 仅容忍 base64 URL 字符集；任何异常输入都视为"未设置"，让调用方签发新 ID。
func readIDFromCookie(c *gin.Context, name string) string {
	raw, err := c.Cookie(name)
	if err != nil {
		return ""
	}
	raw = strings.TrimSpace(raw)
	if raw == "" || len(raw) > 128 {
		return ""
	}
	for _, r := range raw {
		if !isSessionIDChar(r) {
			return ""
		}
	}
	return raw
}

func isSessionIDChar(r rune) bool {
	return (r >= 'A' && r <= 'Z') || (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') || r == '-' || r == '_'
}

// writeCookie 设置 Session Cookie，包含配置里声明的所有安全标志。
// SameSite 单独通过 http.SetCookie 构造，因为 gin.Context.SetCookie 不支持该字段。
func writeCookie(c *gin.Context, m *manager, id string) {
	cookie := &http.Cookie{
		Name:     m.name,
		Value:    id,
		Path:     m.cookie.Path,
		Domain:   m.cookie.Domain,
		MaxAge:   int(m.lifetime.Seconds()),
		Secure:   m.cookie.Secure,
		HttpOnly: m.cookie.HttpOnly,
		SameSite: parseSameSite(m.cookie.SameSite),
	}
	http.SetCookie(c.Writer, cookie)
}

// parseSameSite 映射配置字符串到 http.SameSite 常量。
// 非法值回退为 Lax，与主流浏览器默认行为对齐。
func parseSameSite(s string) http.SameSite {
	switch strings.ToLower(s) {
	case "strict":
		return http.SameSiteStrictMode
	case "none":
		return http.SameSiteNoneMode
	case "lax", "":
		return http.SameSiteLaxMode
	default:
		return http.SameSiteLaxMode
	}
}
