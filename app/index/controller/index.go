package controller

import (
	"net/http"
	"thinkgin/extend/middleware"
	"time"

	"github.com/gin-gonic/gin"
)

// Hello ThinkGin
func HelloWord(c *gin.Context) {
	start := time.Now()

	//name := c.Query("name") /*接收参数*/
	data := make(map[string]interface{})
	now := time.Now()

	data["data"] = "Hello ThinkGin!"
	data["name"] = "who test,who care!"
	data["Time"] = now

	// 业务监控示例
	// 记录API调用次数
	if counter := middleware.BusinessCounter("api_calls", []string{"endpoint", "status"}); counter != nil {
		counter.WithLabelValues("/index/hello", "success").Inc()
	}

	// 记录API响应时间
	if histogram := middleware.BusinessHistogram("api_duration", []string{"endpoint"}, nil); histogram != nil {
		histogram.WithLabelValues("/index/hello").Observe(time.Since(start).Seconds())
	}

	// 记录业务日志
	middleware.BusinessLogger("info", "Hello API called", map[string]interface{}{
		"endpoint":      "/index/hello",
		"timestamp":     now,
		"response_data": data,
	})

	c.JSON(http.StatusOK, gin.H{
		"code": 200,
		"msg":  "SUCCESS",
		"data": data,
	})
}
