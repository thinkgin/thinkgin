package middleware

import (
	"bufio"
	"bytes"
	"context"
	"net"
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
)

const (
	defaultTimeout       = 30 * time.Second
	timeoutWriterNoWrite = -1
)

var timeoutResponseBody = []byte(`{"code":504,"message":"request timeout"}`)

// Timeout returns a middleware that applies the default request timeout.
func Timeout() gin.HandlerFunc {
	return TimeoutWithDuration(defaultTimeout)
}

// TimeoutWithDuration returns a middleware that applies a custom request timeout.
//
// 实现采用 "goroutine handler + select" 模式：c.Next() 在子协程中执行，
// 主协程通过 select 确定性地选择"完成"或"超时"分支，消除竞态窗口。
func TimeoutWithDuration(timeout time.Duration) gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(c.Request.Context(), timeout)
		defer cancel()

		originalWriter := c.Writer
		bufferedWriter := newTimeoutWriter(originalWriter)
		c.Writer = bufferedWriter
		c.Request = c.Request.WithContext(ctx)

		done := make(chan struct{})
		panicCh := make(chan interface{}, 1)

		go func() {
			defer func() {
				if p := recover(); p != nil {
					panicCh <- p
				}
				close(done)
			}()
			c.Next()
		}()

		select {
		case <-done:
			// Handler 完成（含 panic）。先检查是否有 panic 需要重新抛出。
			select {
			case p := <-panicCh:
				// 恢复原始 Writer，使上层 Recovery 中间件直接写入真实响应。
				c.Writer = originalWriter
				panic(p)
			default:
			}
			// 正常完成 → 提交缓冲响应到原始 Writer
			if err := bufferedWriter.commit(); err != nil {
				_ = c.Error(err)
			}
		case <-ctx.Done():
			// 超时 → 写 504。handler 协程持有已取消的 context，应尽快退出。
			// 必须等待其退出后再返回：否则 ServeHTTP 已结束、调用方读取响应，
			// 但 handler 协程仍在访问 c/c.Writer，触发 -race 数据竞争。
			// 注意：编写规范要求业务 handler 监听 ctx.Done() 提前退出，否则会阻塞此处。
			bufferedWriter.writeTimeoutResponse()
			<-done
		}
	}
}

type timeoutWriter struct {
	parent gin.ResponseWriter
	header http.Header
	body   bytes.Buffer

	mu                  sync.Mutex
	status              int
	size                int
	committed           bool
	timeoutResponseSent bool
}

func newTimeoutWriter(parent gin.ResponseWriter) *timeoutWriter {
	return &timeoutWriter{
		parent: parent,
		header: parent.Header().Clone(),
		status: http.StatusOK,
		size:   timeoutWriterNoWrite,
	}
}

func (w *timeoutWriter) Header() http.Header {
	return w.header
}

func (w *timeoutWriter) WriteHeader(code int) {
	w.mu.Lock()
	defer w.mu.Unlock()

	if w.committed || w.timeoutResponseSent {
		return
	}
	if code > 0 && w.status != code {
		if w.writtenLocked() {
			return
		}
		w.status = code
	}
}

func (w *timeoutWriter) WriteHeaderNow() {
	w.mu.Lock()
	defer w.mu.Unlock()

	if w.committed || w.timeoutResponseSent {
		return
	}
	if !w.writtenLocked() {
		w.size = 0
	}
}

func (w *timeoutWriter) Write(data []byte) (int, error) {
	w.mu.Lock()
	defer w.mu.Unlock()

	if w.committed || w.timeoutResponseSent {
		return 0, http.ErrHandlerTimeout
	}

	if !w.writtenLocked() {
		w.size = 0
	}
	n, err := w.body.Write(data)
	w.size += n
	return n, err
}

func (w *timeoutWriter) WriteString(s string) (int, error) {
	return w.Write([]byte(s))
}

func (w *timeoutWriter) Status() int {
	w.mu.Lock()
	defer w.mu.Unlock()
	return w.status
}

func (w *timeoutWriter) Size() int {
	w.mu.Lock()
	defer w.mu.Unlock()
	return w.size
}

func (w *timeoutWriter) Written() bool {
	w.mu.Lock()
	defer w.mu.Unlock()
	return w.writtenLocked()
}

func (w *timeoutWriter) Flush() {
	w.WriteHeaderNow()
}

// CloseNotify 实现 gin.ResponseWriter 接口要求。
// Deprecated: Go 1.11 已废弃 http.CloseNotifier，但 Gin 接口仍要求。
func (w *timeoutWriter) CloseNotify() <-chan bool {
	//nolint:staticcheck // SA1019: Gin 接口要求
	return w.parent.CloseNotify()
}

func (w *timeoutWriter) Hijack() (net.Conn, *bufio.ReadWriter, error) {
	return w.parent.Hijack()
}

func (w *timeoutWriter) Pusher() http.Pusher {
	return w.parent.Pusher()
}

func (w *timeoutWriter) commit() error {
	w.mu.Lock()
	defer w.mu.Unlock()

	if w.timeoutResponseSent || w.committed {
		return nil
	}

	w.committed = true
	return writeBufferedResponse(w.parent, w.header, w.status, w.body.Bytes())
}

func (w *timeoutWriter) writeTimeoutResponse() {
	w.mu.Lock()
	defer w.mu.Unlock()

	if w.timeoutResponseSent || w.committed {
		return
	}

	if w.writtenLocked() {
		return
	}

	w.timeoutResponseSent = true
	w.status = http.StatusGatewayTimeout
	w.size = len(timeoutResponseBody)

	header := w.parent.Header()
	resetHeader(header)
	header.Set("Content-Type", "application/json; charset=utf-8")
	header.Del("Content-Length")

	w.parent.WriteHeader(w.status)
	_, _ = w.parent.Write(timeoutResponseBody)
}

func writeBufferedResponse(dst gin.ResponseWriter, header http.Header, status int, body []byte) error {
	resetHeader(dst.Header())
	copyHeader(dst.Header(), header)
	dst.WriteHeader(status)
	if len(body) == 0 {
		dst.WriteHeaderNow()
		return nil
	}

	_, err := dst.Write(body)
	return err
}

func resetHeader(header http.Header) {
	for key := range header {
		header.Del(key)
	}
}

func copyHeader(dst, src http.Header) {
	for key, values := range src {
		for _, value := range values {
			dst.Add(key, value)
		}
	}
}

func (w *timeoutWriter) writtenLocked() bool {
	return w.size != timeoutWriterNoWrite
}
