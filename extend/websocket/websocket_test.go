package websocket

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	ws "github.com/gorilla/websocket"
)

func init() {
	gin.SetMode(gin.TestMode)
}

func closeResponseBody(resp *http.Response) {
	if resp != nil && resp.Body != nil {
		_ = resp.Body.Close()
	}
}

// setupEchoServer 创建一个 echo WebSocket 测试服务器。
func setupEchoServer() *httptest.Server {
	r := gin.New()
	r.GET("/ws", Handler(func(conn *Conn) {
		for {
			mt, msg, err := conn.ReadMessage()
			if err != nil {
				break
			}
			if err := conn.WriteSafeMessage(mt, msg); err != nil {
				break
			}
		}
	}))
	return httptest.NewServer(r)
}

func TestHandler_EchoMessage(t *testing.T) {
	srv := setupEchoServer()
	defer srv.Close()

	url := "ws" + strings.TrimPrefix(srv.URL, "http") + "/ws"
	conn, resp, err := ws.DefaultDialer.Dial(url, nil)
	closeResponseBody(resp)
	if err != nil {
		t.Fatalf("dial: %v", err)
	}
	defer conn.Close()

	// 发送
	want := "hello thinkgin"
	if err := conn.WriteMessage(ws.TextMessage, []byte(want)); err != nil {
		t.Fatalf("write: %v", err)
	}

	// 接收 echo
	conn.SetReadDeadline(time.Now().Add(2 * time.Second))
	_, msg, err := conn.ReadMessage()
	if err != nil {
		t.Fatalf("read: %v", err)
	}
	if string(msg) != want {
		t.Errorf("echo = %q, want %q", string(msg), want)
	}
}

func TestHandlerWithUpgrader_CustomOriginCheck(t *testing.T) {
	r := gin.New()
	strictUpgrader := ws.Upgrader{
		CheckOrigin: func(r *http.Request) bool {
			return r.Header.Get("Origin") == "https://allowed.example.com"
		},
	}
	r.GET("/ws", HandlerWithUpgrader(func(conn *Conn) {
		_ = conn.WriteSafeMessage(ws.TextMessage, []byte("ok"))
	}, strictUpgrader))
	srv := httptest.NewServer(r)
	defer srv.Close()

	url := "ws" + strings.TrimPrefix(srv.URL, "http") + "/ws"

	// 无 Origin → 被拒绝
	_, resp, err := ws.DefaultDialer.Dial(url, nil)
	closeResponseBody(resp)
	if err == nil {
		t.Fatal("expected dial to fail without Origin header")
	}
	if resp != nil && resp.StatusCode != http.StatusForbidden {
		t.Logf("status = %d (rejected as expected)", resp.StatusCode)
	}

	// 带正确 Origin → 成功
	header := http.Header{}
	header.Set("Origin", "https://allowed.example.com")
	conn, resp, err := ws.DefaultDialer.Dial(url, header)
	closeResponseBody(resp)
	if err != nil {
		t.Fatalf("dial with allowed origin: %v", err)
	}
	defer conn.Close()

	conn.SetReadDeadline(time.Now().Add(2 * time.Second))
	_, msg, err := conn.ReadMessage()
	if err != nil {
		t.Fatalf("read: %v", err)
	}
	if string(msg) != "ok" {
		t.Errorf("got %q, want ok", string(msg))
	}
}

func TestConn_WriteJSON(t *testing.T) {
	r := gin.New()
	r.GET("/ws", Handler(func(conn *Conn) {
		type resp struct {
			Status string `json:"status"`
		}
		_ = conn.WriteJSON(resp{Status: "connected"})
	}))
	srv := httptest.NewServer(r)
	defer srv.Close()

	url := "ws" + strings.TrimPrefix(srv.URL, "http") + "/ws"
	conn, resp, err := ws.DefaultDialer.Dial(url, nil)
	closeResponseBody(resp)
	if err != nil {
		t.Fatalf("dial: %v", err)
	}
	defer conn.Close()

	conn.SetReadDeadline(time.Now().Add(2 * time.Second))
	_, msg, err := conn.ReadMessage()
	if err != nil {
		t.Fatalf("read: %v", err)
	}
	if !strings.Contains(string(msg), `"status":"connected"`) {
		t.Errorf("json = %q, want contain status:connected", string(msg))
	}
}

func TestHub_BroadcastAndLen(t *testing.T) {
	hub := NewHub()

	r := gin.New()
	r.GET("/ws", Handler(func(conn *Conn) {
		hub.Register(conn)
		defer hub.Unregister(conn)
		// 阻塞直到客户端断开
		for {
			_, _, err := conn.ReadMessage()
			if err != nil {
				break
			}
		}
	}))
	srv := httptest.NewServer(r)
	defer srv.Close()

	url := "ws" + strings.TrimPrefix(srv.URL, "http") + "/ws"

	// 连接 2 个客户端
	c1, resp, err := ws.DefaultDialer.Dial(url, nil)
	closeResponseBody(resp)
	if err != nil {
		t.Fatalf("dial c1: %v", err)
	}
	defer c1.Close()

	c2, resp, err := ws.DefaultDialer.Dial(url, nil)
	closeResponseBody(resp)
	if err != nil {
		t.Fatalf("dial c2: %v", err)
	}
	defer c2.Close()

	// 等连接注册
	time.Sleep(100 * time.Millisecond)

	if hub.Len() != 2 {
		t.Errorf("hub.Len() = %d, want 2", hub.Len())
	}

	// 广播
	hub.Broadcast(ws.TextMessage, []byte("broadcast"))

	for i, c := range []*ws.Conn{c1, c2} {
		c.SetReadDeadline(time.Now().Add(2 * time.Second))
		_, msg, err := c.ReadMessage()
		if err != nil {
			t.Errorf("client %d read: %v", i, err)
			continue
		}
		if string(msg) != "broadcast" {
			t.Errorf("client %d got %q, want broadcast", i, string(msg))
		}
	}
}

func TestKeepAlive_ServerSendsPing(t *testing.T) {
	r := gin.New()
	// 业务侧用很短的 ping 周期，便于测试快速观察到 ping。
	r.GET("/ws", KeepAliveHandler(func(conn *Conn) {
		for {
			if _, _, err := conn.ReadMessage(); err != nil {
				return
			}
		}
	}, KeepAliveConfig{PingPeriod: 50 * time.Millisecond, PongWait: 2 * time.Second, WriteWait: time.Second}))
	srv := httptest.NewServer(r)
	defer srv.Close()

	url := "ws" + strings.TrimPrefix(srv.URL, "http") + "/ws"
	conn, resp, err := ws.DefaultDialer.Dial(url, nil)
	closeResponseBody(resp)
	if err != nil {
		t.Fatalf("dial: %v", err)
	}
	defer conn.Close()

	// 客户端登记 ping handler，收到服务端 ping 即标记。
	gotPing := make(chan struct{}, 1)
	conn.SetPingHandler(func(string) error {
		select {
		case gotPing <- struct{}{}:
		default:
		}
		return nil
	})

	// 触发读循环以处理控制帧（ping）。
	go func() {
		conn.SetReadDeadline(time.Now().Add(2 * time.Second))
		for {
			if _, _, err := conn.ReadMessage(); err != nil {
				return
			}
		}
	}()

	select {
	case <-gotPing:
		// 收到服务端心跳 ping，符合预期。
	case <-time.After(1500 * time.Millisecond):
		t.Error("未在预期时间内收到服务端 ping")
	}
}

func TestStopKeepAlive_Idempotent(t *testing.T) {
	// 未启动心跳时调用 StopKeepAlive 不应 panic。
	c1 := &Conn{}
	c1.StopKeepAlive()

	// 已启动心跳的连接，多次停止也应安全且不重复 close。
	c2 := &Conn{keepAliveStop: make(chan struct{})}
	c2.StopKeepAlive()
	c2.StopKeepAlive()

	// 确认 channel 已被关闭。
	select {
	case <-c2.keepAliveStop:
		// 已关闭，符合预期
	default:
		t.Error("StopKeepAlive 应已关闭 keepAliveStop channel")
	}
}

func TestKeepAliveConfig_Normalized(t *testing.T) {
	// 零值应回退到默认值。
	got := KeepAliveConfig{}.normalized()
	if got.WriteWait != defaultWriteWait || got.PongWait != defaultPongWait {
		t.Errorf("zero config not defaulted: %+v", got)
	}
	if got.PingPeriod >= got.PongWait {
		t.Errorf("PingPeriod %v must be < PongWait %v", got.PingPeriod, got.PongWait)
	}

	// PingPeriod >= PongWait 应被纠正为小于 PongWait。
	bad := KeepAliveConfig{PongWait: time.Second, PingPeriod: 2 * time.Second}.normalized()
	if bad.PingPeriod >= bad.PongWait {
		t.Errorf("PingPeriod not corrected: %+v", bad)
	}
}

func TestHub_BroadcastJSON(t *testing.T) {
	hub := NewHub()

	r := gin.New()
	r.GET("/ws", Handler(func(conn *Conn) {
		hub.Register(conn)
		defer hub.Unregister(conn)
		for {
			_, _, err := conn.ReadMessage()
			if err != nil {
				break
			}
		}
	}))
	srv := httptest.NewServer(r)
	defer srv.Close()

	url := "ws" + strings.TrimPrefix(srv.URL, "http") + "/ws"
	c, resp, err := ws.DefaultDialer.Dial(url, nil)
	closeResponseBody(resp)
	if err != nil {
		t.Fatalf("dial: %v", err)
	}
	defer c.Close()

	time.Sleep(100 * time.Millisecond)

	type event struct {
		Type string `json:"type"`
		Data string `json:"data"`
	}
	if err := hub.BroadcastJSON(event{Type: "msg", Data: "hello"}); err != nil {
		t.Fatalf("BroadcastJSON: %v", err)
	}

	c.SetReadDeadline(time.Now().Add(2 * time.Second))
	_, msg, err := c.ReadMessage()
	if err != nil {
		t.Fatalf("read: %v", err)
	}
	if !strings.Contains(string(msg), `"type":"msg"`) {
		t.Errorf("json = %q, want contain type:msg", string(msg))
	}
}
