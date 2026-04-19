// Package controller 提供 index 模块的示例 Handler。
// 本文件同时演示：业务计数器、业务直方图与结构化业务日志的使用方式。
package controller

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"

	"thinkgin/extend/middleware"
)

// HelloWord 最小示例接口，返回固定数据并演示业务监控埋点。
// 该 Handler 同时供页面路由 /index/hello 与 API /api/v1/index/hello 复用。
func HelloWord(c *gin.Context) {
	start := time.Now()
	now := start

	data := gin.H{
		"data": "Hello ThinkGin!",
		"name": c.Query("name"),
		"time": now,
	}

	// 计数器：按 endpoint + status 维度统计 API 调用量。
	if counter := middleware.BusinessCounter("api_calls", []string{"endpoint", "status"}); counter != nil {
		counter.WithLabelValues("/index/hello", "success").Inc()
	}

	// 直方图：观测该接口的端到端耗时分布。
	if histogram := middleware.BusinessHistogram("api_duration", []string{"endpoint"}, nil); histogram != nil {
		histogram.WithLabelValues("/index/hello").Observe(time.Since(start).Seconds())
	}

	// 结构化业务日志：与 AccessLog 分离，供离线分析使用。
	middleware.BusinessLogger("info", "hello api called", map[string]interface{}{
		"endpoint":  "/index/hello",
		"timestamp": now,
	})

	c.JSON(http.StatusOK, gin.H{
		"code": http.StatusOK,
		"msg":  "ok",
		"data": data,
	})
}
