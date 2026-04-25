// 本文件提供 Gzip 响应压缩中间件。
//
// 仅在客户端 Accept-Encoding 包含 gzip 且响应体大于 minSize 时压缩。
// 对 SSE、WebSocket 等流式响应自动跳过。
package middleware

import (
	"compress/gzip"
	"io"
	"net/http"
	"strings"
	"sync"

	"github.com/gin-gonic/gin"
)

const (
	// gzipMinSize 小于此字节数的响应不压缩，避免小包膨胀。
	gzipMinSize = 1024
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

		gz := gzipPool.Get().(*gzip.Writer)
		defer gzipPool.Put(gz)
		gz.Reset(c.Writer)

		c.Header("Content-Encoding", "gzip")
		c.Header("Vary", "Accept-Encoding")
		// 删除 Content-Length，因为压缩后长度变化。
		c.Writer.Header().Del("Content-Length")

		c.Writer = &gzipWriter{ResponseWriter: c.Writer, gz: gz}
		defer func() {
			gz.Close()
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
		defer reader.Close()

		c.Request.Body = io.NopCloser(reader)
		c.Request.Header.Del("Content-Encoding")
		c.Next()
	}
}
