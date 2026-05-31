// Package session 提供进程内或 Redis 后端的会话存储，
// 消费 config/session.yaml 的所有字段（除 database/file driver 外）。
//
// 设计要点：
//   - Store 接口隔离后端细节（memory / redis），便于扩展与测试。
//   - Session 实例维护 dirty 标志：未修改就不触发持久化，避免无意义 I/O。
//   - Session ID 使用 crypto/rand 生成，32 字节 base64 URL 安全编码。
//
// 暂未实现的 driver：
//   - file：文件存储（易受磁盘故障影响，需要额外考虑加锁）
//   - database：需绑定 GORM 模型，待有业务诉求再做
//
// 业务用法：
//
//	r.Use(session.Middleware())
//	r.GET("/profile", func(c *gin.Context) {
//	    s := session.From(c)
//	    s.Set("user_id", 42)
//	    _ = s.Save(c.Request.Context())
//	})
package session

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"fmt"
	"sync"
	"time"

	"thinkgin/app"
)

// ErrNotInitialized 在调用方未执行 Init 就尝试使用中间件时返回。
var ErrNotInitialized = errors.New("session: manager not initialized; call session.Init() first")

// Store 抽象会话持久化后端。
// ttl 为 0 时表示"随配置默认"，实现方应自行读取 config 或设置合理兜底。
type Store interface {
	Load(ctx context.Context, id string) (map[string]any, error)
	Save(ctx context.Context, id string, data map[string]any, ttl time.Duration) error
	Destroy(ctx context.Context, id string) error
	Close() error
}

// manager 保存全局单例所需的状态，避免跨包暴露可变字段。
type manager struct {
	store    Store
	lifetime time.Duration
	cookie   cookieConfig
	name     string // Cookie name
}

// cookieConfig 是从 config 抽取的 Cookie 相关子集，避免中间件重复读取全局配置。
type cookieConfig struct {
	Path     string
	Domain   string
	Secure   bool
	HttpOnly bool
	SameSite string
}

var (
	managerMu sync.RWMutex
	active    *manager
)

// Init 根据 config/session.yaml 初始化会话管理器。
// 多次调用会替换现有实例并关闭旧 store，便于热更新与测试。
//
// driver=memory 时在进程内保存；driver=redis 时引用 database.yaml 中的 redis 连接。
// 其他 driver 视为不支持，返回错误；调用方可选择继续而不启用 session。
func Init() error {
	cfg := app.GetConfig()
	if cfg == nil {
		return errors.New("session: global config is nil, did you forget Bootstrap?")
	}
	sc := cfg.Session

	lifetime := time.Duration(sc.Lifetime) * time.Second
	if lifetime <= 0 {
		lifetime = 2 * time.Hour // 安全兜底，与 session.yaml 默认 7200s 对齐
	}

	var store Store
	switch sc.Driver {
	case "", "memory":
		store = newMemoryStore(lifetime)
	case "redis":
		s, err := newRedisStore(sc.Redis.Connection, lifetime)
		if err != nil {
			return fmt.Errorf("session: init redis store: %w", err)
		}
		store = s
	case "file", "database":
		return fmt.Errorf("session: driver %q is declared in config but not yet implemented", sc.Driver)
	default:
		return fmt.Errorf("session: unknown driver %q", sc.Driver)
	}

	managerMu.Lock()
	defer managerMu.Unlock()
	if active != nil {
		_ = active.store.Close()
	}
	active = &manager{
		store:    store,
		lifetime: lifetime,
		name:     firstNonEmpty(sc.Name, "THINKGIN_SESSION"),
		cookie: cookieConfig{
			Path:   firstNonEmpty(sc.Cookie.Path, "/"),
			Domain: sc.Cookie.Domain,
			Secure: sc.Cookie.Secure,
			// 安全默认：Session ID 绝不应被 JS 读取，强制 HttpOnly。
			// 即便配置缺省（零值 false），也回正为 true，杜绝 XSS 窃取会话。
			HttpOnly: true,
			SameSite: firstNonEmpty(sc.Cookie.SameSite, "lax"),
		},
	}
	return nil
}

// Shutdown 关闭底层 store 并清空 manager。幂等。
func Shutdown() error {
	managerMu.Lock()
	defer managerMu.Unlock()
	if active == nil {
		return nil
	}
	err := active.store.Close()
	active = nil
	return err
}

// current 以只读方式获取当前 manager；未初始化时返回 nil。
// 中间件通过此函数取实例，避免直接读取 active 变量带来的并发风险。
func current() *manager {
	managerMu.RLock()
	defer managerMu.RUnlock()
	return active
}

// generateID 返回一个 URL 安全的 256 bit 随机 Session ID。
// 失败时返回错误而非 panic，避免占用不必要的进程崩溃预算。
func generateID() (string, error) {
	var buf [32]byte
	if _, err := rand.Read(buf[:]); err != nil {
		return "", fmt.Errorf("session: read random: %w", err)
	}
	return base64.RawURLEncoding.EncodeToString(buf[:]), nil
}

// firstNonEmpty 返回第一个非空字符串；所有都为空时返回最后一个参数作为默认。
func firstNonEmpty(s ...string) string {
	for _, v := range s {
		if v != "" {
			return v
		}
	}
	return ""
}
