package controller

import (
	"net/http"
	"thinkgin/extend/middleware"
	"time"

	"github.com/gin-gonic/gin"
)

// Hello ThinkGin
func HelloWord(c *gin.Context) {
	//name := c.Query("name") /*接收参数*/
	data := make(map[string]interface{})
	now := time.Now()

	data["data"] = "Hello ThinkGin!"
	data["name"] = "who test,who care!"
	data["Time"] = now

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
