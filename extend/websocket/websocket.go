// Package websocket 提供基于 gorilla/websocket 的 Gin WebSocket 支持封装。
//
// 设计思路：
//   - 提供 Upgrader 封装，简化 Gin Handler 中的 WebSocket 升级流程。
//   - 提供 Hub 广播模型，支持多客户端消息分发。
//   - 提供 JSON 消息辅助读写。
//
// 使用方式：
//
//	r.GET("/ws", websocket.Handler(func(conn *websocket.Conn) {
//	    for {
//	        mt, msg, err := conn.ReadMessage()
//	        if err != nil { break }
//	        conn.WriteMessage(mt, msg) // echo
//	    }
//	}))
package websocket

import (
	"encoding/json"
	"net/http"
	"sync"
	"time"

	"thinkgin/app"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

// 默认的心跳与超时参数。生产环境可通过 KeepAliveConfig 覆盖。
const (
	// defaultWriteWait 单次写操作的超时，防止慢客户端阻塞写协程。
	defaultWriteWait = 10 * time.Second
	// defaultPongWait 读取下一条消息（含 pong）的最大等待时间，超时视为连接已死。
	defaultPongWait = 60 * time.Second
	// defaultPingPeriod 主动发送 ping 的周期，必须小于 pongWait（留出 pong 往返余量）。
	defaultPingPeriod = (defaultPongWait * 9) / 10
)

// Conn 包装 gorilla/websocket.Conn，提供便捷方法。
type Conn struct {
	*websocket.Conn
	mu        sync.Mutex // 保护并发写
	writeWait time.Duration

	keepAliveOnce sync.Once
	keepAliveStop chan struct{}
}

// KeepAliveConfig 配置心跳与超时。零值字段回退到默认常量。
type KeepAliveConfig struct {
	// WriteWait 单次写超时。
	WriteWait time.Duration
	// PongWait 读超时（等待对端任意消息/pong 的最长时间）。
	PongWait time.Duration
	// PingPeriod 主动 ping 周期，应小于 PongWait。
	PingPeriod time.Duration
}

func (k KeepAliveConfig) normalized() KeepAliveConfig {
	if k.WriteWait <= 0 {
		k.WriteWait = defaultWriteWait
	}
	if k.PongWait <= 0 {
		k.PongWait = defaultPongWait
	}
	if k.PingPeriod <= 0 || k.PingPeriod >= k.PongWait {
		// ping 周期必须小于 pong 等待，否则永远来不及收到 pong 就判超时。
		k.PingPeriod = (k.PongWait * 9) / 10
	}
	return k
}

// WriteJSON 线程安全地发送 JSON 消息。
func (c *Conn) WriteJSON(v any) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.applyWriteDeadline()
	return c.Conn.WriteJSON(v)
}

// WriteSafeMessage 线程安全地发送原始消息。
func (c *Conn) WriteSafeMessage(messageType int, data []byte) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.applyWriteDeadline()
	return c.WriteMessage(messageType, data)
}

// applyWriteDeadline 在持锁状态下为下一次写设置超时（writeWait<=0 时不设置）。
func (c *Conn) applyWriteDeadline() {
	if c.writeWait > 0 {
		_ = c.SetWriteDeadline(time.Now().Add(c.writeWait))
	}
}

// StartKeepAlive 启用心跳保活：
//   - 设置读超时为 PongWait，并在收到 pong 时自动续期。
//   - 后台按 PingPeriod 周期发送 ping；写失败或调用方 StopKeepAlive 时退出。
//   - 设置写超时为 WriteWait，避免慢客户端阻塞。
//
// 必须在进入读循环（ReadMessage）之前调用一次。多次调用只有第一次生效。
// 返回的 stop 函数可显式停止心跳协程（通常用 defer 调用）。
func (c *Conn) StartKeepAlive(cfg KeepAliveConfig) func() {
	cfg = cfg.normalized()
	c.keepAliveOnce.Do(func() {
		c.writeWait = cfg.WriteWait
		c.keepAliveStop = make(chan struct{})

		// 读超时 + pong 续期：每次收到 pong 就把读 deadline 往后推。
		_ = c.SetReadDeadline(time.Now().Add(cfg.PongWait))
		c.SetPongHandler(func(string) error {
			return c.SetReadDeadline(time.Now().Add(cfg.PongWait))
		})

		go c.pingLoop(cfg.PingPeriod)
	})
	return c.StopKeepAlive
}

// pingLoop 周期性发送 ping，直到写失败或被停止。
func (c *Conn) pingLoop(period time.Duration) {
	ticker := time.NewTicker(period)
	defer ticker.Stop()
	for {
		select {
		case <-c.keepAliveStop:
			return
		case <-ticker.C:
			if err := c.WriteSafeMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}

// StopKeepAlive 停止心跳协程，幂等。
func (c *Conn) StopKeepAlive() {
	c.mu.Lock()
	stop := c.keepAliveStop
	c.keepAliveStop = nil
	c.mu.Unlock()
	if stop != nil {
		close(stop)
	}
}

// ReadJSON 从连接读取一条 JSON 消息并解码。
func (c *Conn) ReadJSON(v any) error {
	_, msg, err := c.ReadMessage()
	if err != nil {
		return err
	}
	return json.Unmarshal(msg, v)
}

// ConnHandler 是处理已升级 WebSocket 连接的回调函数类型。
type ConnHandler func(conn *Conn)

// DefaultUpgrader 默认的 WebSocket Upgrader 配置。
// CheckOrigin 从 CORS 配置中读取 allow_origins 白名单。
// 若配置为 ["*"] 或未配置 CORS，则允许所有 Origin。
var DefaultUpgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin:     checkOriginFromConfig,
}

// checkOriginFromConfig 根据 middleware.config.cors.allow_origins 校验 WebSocket Origin。
func checkOriginFromConfig(r *http.Request) bool {
	origin := r.Header.Get("Origin")
	if origin == "" {
		return true // 同源请求无 Origin 头
	}

	allowed := getAllowedOrigins()
	if len(allowed) == 0 {
		return true // 未配置时默认放行
	}
	for _, o := range allowed {
		if o == "*" || o == origin {
			return true
		}
	}
	return false
}

// getAllowedOrigins 从全局 CORS 配置中提取 allow_origins 列表。
func getAllowedOrigins() []string {
	cfg := app.GetConfig()
	if cfg == nil || cfg.Middleware.Config == nil {
		return nil
	}
	raw, ok := cfg.Middleware.Config["cors"]
	if !ok {
		return nil
	}
	m, ok := raw.(map[string]interface{})
	if !ok {
		return nil
	}
	rawOrigins, ok := m["allow_origins"]
	if !ok {
		return nil
	}

	switch v := rawOrigins.(type) {
	case []interface{}:
		out := make([]string, 0, len(v))
		for _, item := range v {
			if s, ok := item.(string); ok {
				out = append(out, s)
			}
		}
		return out
	case string:
		return []string{v}
	default:
		return nil
	}
}

// Handler 返回将 HTTP 升级为 WebSocket 的 Gin HandlerFunc。
// 使用 DefaultUpgrader。
func Handler(h ConnHandler) gin.HandlerFunc {
	return HandlerWithUpgrader(h, DefaultUpgrader)
}

// HandlerWithUpgrader 返回使用自定义 Upgrader 的 WebSocket Gin HandlerFunc。
func HandlerWithUpgrader(h ConnHandler, upgrader websocket.Upgrader) gin.HandlerFunc {
	return func(c *gin.Context) {
		ws, err := upgrader.Upgrade(c.Writer, c.Request, nil)
		if err != nil {
			// Upgrade 失败时 gorilla 已经写了 HTTP 错误响应
			return
		}
		conn := &Conn{Conn: ws}
		defer func() {
			conn.StopKeepAlive()
			_ = ws.Close()
		}()
		h(conn)
	}
}

// KeepAliveHandler 返回自动启用心跳保活的 WebSocket Handler。
//
// 与 Handler 的区别：在调用业务回调前自动 StartKeepAlive(cfg)，
// 业务回调只需正常进入 ReadMessage 读循环即可享受 ping/pong 与超时保护。
// 适合"长连接推送"场景，避免每个业务都重复写心跳样板代码。
func KeepAliveHandler(h ConnHandler, cfg KeepAliveConfig) gin.HandlerFunc {
	return func(c *gin.Context) {
		ws, err := DefaultUpgrader.Upgrade(c.Writer, c.Request, nil)
		if err != nil {
			return
		}
		conn := &Conn{Conn: ws}
		stop := conn.StartKeepAlive(cfg)
		defer func() {
			stop()
			_ = ws.Close()
		}()
		h(conn)
	}
}

// ────────────────────────────────────────────────
// Hub：多客户端广播模型
// ────────────────────────────────────────────────

// Hub 管理一组 WebSocket 客户端并广播消息。
type Hub struct {
	mu      sync.RWMutex
	clients map[*Conn]struct{}
}

// NewHub 创建一个空的 Hub。
func NewHub() *Hub {
	return &Hub{
		clients: make(map[*Conn]struct{}),
	}
}

// Register 将连接注册到 Hub。
func (h *Hub) Register(conn *Conn) {
	h.mu.Lock()
	h.clients[conn] = struct{}{}
	h.mu.Unlock()
}

// Unregister 从 Hub 移除连接并关闭。
func (h *Hub) Unregister(conn *Conn) {
	h.mu.Lock()
	delete(h.clients, conn)
	h.mu.Unlock()
	_ = conn.Close()
}

// Broadcast 向所有注册客户端发送消息。
// 发送失败的客户端会被自动移除。
func (h *Hub) Broadcast(messageType int, data []byte) {
	h.mu.RLock()
	clients := make([]*Conn, 0, len(h.clients))
	for c := range h.clients {
		clients = append(clients, c)
	}
	h.mu.RUnlock()

	var failed []*Conn
	for _, c := range clients {
		if err := c.WriteSafeMessage(messageType, data); err != nil {
			failed = append(failed, c)
		}
	}

	// 清理发送失败的连接
	for _, c := range failed {
		h.Unregister(c)
	}
}

// BroadcastJSON 向所有注册客户端发送 JSON 消息。
func (h *Hub) BroadcastJSON(v any) error {
	data, err := json.Marshal(v)
	if err != nil {
		return err
	}
	h.Broadcast(websocket.TextMessage, data)
	return nil
}

// Len 返回当前连接数。
func (h *Hub) Len() int {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return len(h.clients)
}
