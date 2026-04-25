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

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

// Conn 包装 gorilla/websocket.Conn，提供便捷方法。
type Conn struct {
	*websocket.Conn
	mu sync.Mutex // 保护并发写
}

// WriteJSON 线程安全地发送 JSON 消息。
func (c *Conn) WriteJSON(v any) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.Conn.WriteJSON(v)
}

// WriteSafeMessage 线程安全地发送原始消息。
func (c *Conn) WriteSafeMessage(messageType int, data []byte) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.Conn.WriteMessage(messageType, data)
}

// ReadJSON 从连接读取一条 JSON 消息并解码。
func (c *Conn) ReadJSON(v any) error {
	_, msg, err := c.Conn.ReadMessage()
	if err != nil {
		return err
	}
	return json.Unmarshal(msg, v)
}

// ConnHandler 是处理已升级 WebSocket 连接的回调函数类型。
type ConnHandler func(conn *Conn)

// DefaultUpgrader 默认的 WebSocket Upgrader 配置。
// CheckOrigin 默认放通所有 origin，生产环境请替换为严格检查。
var DefaultUpgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
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
		defer func() { _ = ws.Close() }()
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
