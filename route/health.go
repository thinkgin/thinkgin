// 本文件定义 /livez 与 /readyz 两个探针的 handler。
//
// 语义严格按 k8s 约定：
//
//	livez  - "进程还活着吗"：只要能执行到 handler 就 200，不检查依赖。
//	         用于驱动 livenessProbe；失败即重启 Pod。
//	readyz - "是否准备好吃流量"：检查所有已注册的 DB / Redis 连接。
//	         任一依赖不可达就返回 503，k8s 会把本 Pod 从 Service 后端摘掉。
//
// 设计取舍：
//   - 每次请求都真正 Ping 依赖，避免缓存状态骗编排器。
//   - 总体 2s 硬超时，防止探针本身阻塞让 kubelet 判超时。
//   - 未配置任何 DB/Redis 时 readyz 也返回 200，与"纯 HTTP 应用"场景兼容。
package route

import (
	"context"
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"

	"thinkgin/app/cache"
	"thinkgin/app/database"
)

// readyzTimeout 是整体探测预算，通常要小于 k8s 探针的 timeoutSeconds。
const readyzTimeout = 2 * time.Second

// componentStatus 表示单个依赖的探测结果，序列化给运维排障用。
type componentStatus struct {
	OK      bool   `json:"ok"`
	Error   string `json:"error,omitempty"`
	Elapsed string `json:"elapsed"`
}

// livezHandler 仅确认 HTTP 处理管道正常。
// 故意不检查任何外部依赖，避免依赖抖动触发 Pod 被 k8s 反复 kill。
func livezHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	}
}

// readyzHandler 并发探测所有已注册的 DB/Redis 连接。
// 任一失败则整体 503，响应体携带每个组件的状态供定位。
func readyzHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(c.Request.Context(), readyzTimeout)
		defer cancel()

		results := map[string]componentStatus{}
		var mu sync.Mutex
		var wg sync.WaitGroup

		// 收集所有待探测目标。Default() 返回 nil 说明未配置，直接跳过。
		type probe struct {
			name string
			run  func(ctx context.Context) error
		}
		var probes []probe

		if db := database.Default(); db != nil {
			probes = append(probes, probe{
				name: "database",
				run: func(ctx context.Context) error {
					sqlDB, err := db.DB()
					if err != nil {
						return err
					}
					return sqlDB.PingContext(ctx)
				},
			})
		}
		if rdb := cache.Default(); rdb != nil {
			probes = append(probes, probe{
				name: "cache",
				run:  func(ctx context.Context) error { return rdb.Ping(ctx).Err() },
			})
		}

		for _, p := range probes {
			wg.Add(1)
			go func(p probe) {
				defer wg.Done()
				start := time.Now()
				err := p.run(ctx)
				status := componentStatus{
					OK:      err == nil,
					Elapsed: time.Since(start).Round(time.Millisecond).String(),
				}
				if err != nil {
					status.Error = err.Error()
				}
				mu.Lock()
				results[p.name] = status
				mu.Unlock()
			}(p)
		}
		wg.Wait()

		// 聚合判定：任一组件不 OK 则整体 503。
		allOK := true
		for _, s := range results {
			if !s.OK {
				allOK = false
				break
			}
		}
		code := http.StatusOK
		status := "ok"
		if !allOK {
			code = http.StatusServiceUnavailable
			status = "degraded"
		}
		c.JSON(code, gin.H{
			"status":     status,
			"components": results,
		})
	}
}
