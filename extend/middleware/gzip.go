// 本文件提供 Gzip 响应压缩中间件。
//
// 仅在客户端 Accept-Encoding 包含 gzip 时压缩。
// 对 SSE（text/event-stream）等流式响应、以及已设置 Content-Encoding 的响应自动跳过，
// 避免破坏流式语义或二次压缩。
//
// 说明：本实现为流式直压（不缓冲整个响应体），因此不做"小于 minSize 不压缩"的阈值判断。
// 如需基于大小的阈值压缩，需引入响应缓冲，属后续可选优化。
package middleware

import (
	"compress/gzip"
	"io"
	"net/http"
	"strings"
	"sync"

	"github.com/gin-gonic/gin"
)

var gzipPool = sync.Pool{
	New: func() any {
		w, _ := gzip.NewWriterLevel(io.Discard, gzip.DefaultCompression)
		return w
	},
}

// gzipWriter 包装 gin.ResponseWriter，透明压缩写入。
type gzipWriter struct {
	gin.ResponseWriter
	gz *gzip.Writer
}

func (g *gzipWriter) Write(data []byte) (int, error) {
	return g.gz.Write(data)
}

func (g *gzipWriter) WriteString(s string) (int, error) {
	return g.gz.Write([]byte(s))
}

// Gzip 返回 gzip 响应压缩中间件。
func Gzip() gin.HandlerFunc {
	return func(c *gin.Context) {
		if !strings.Contains(c.GetHeader("Accept-Encoding"), "gzip") {
			c.Next()
			return
		}

		// 跳过 SSE 流式响应：压缩会破坏 EventSource 的实时推送语义。
		if strings.Contains(c.GetHeader("Accept"), "text/event-stream") {
			c.Next()
			return
		}

		// 已声明 Content-Encoding 的响应（例如上游已压缩）不再二次压缩。
		if c.Writer.Header().Get("Content-Encoding") != "" {
			c.Next()
			return
		}

		gz := gzipPool.Get().(*gzip.Writer)
		defer gzipPool.Put(gz)
		gz.Reset(c.Writer)

		c.Header("Content-Encoding", "gzip")
		c.Header("Vary", "Accept-Encoding")
		// 删除 Content-Length，因为压缩后长度变化。
		c.Writer.Header().Del("Content-Length")

		c.Writer = &gzipWriter{ResponseWriter: c.Writer, gz: gz}
		defer func() {
			_ = gz.Close()
		}()

		c.Next()
	}
}

// GzipDecompressRequest 解压 Content-Encoding: gzip 的请求体，
// 用于接收客户端发送的压缩请求（如日志收集场景）。
func GzipDecompressRequest() gin.HandlerFunc {
	return func(c *gin.Context) {
		if c.GetHeader("Content-Encoding") != "gzip" {
			c.Next()
			return
		}

		reader, err := gzip.NewReader(c.Request.Body)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
				"code":    http.StatusBadRequest,
				"message": "invalid gzip body",
			})
			return
		}
		defer func() { _ = reader.Close() }()

		c.Request.Body = io.NopCloser(reader)
		c.Request.Header.Del("Content-Encoding")
		c.Next()
	}
}
