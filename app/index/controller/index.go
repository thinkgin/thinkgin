package controller

import (
	"github.com/gin-gonic/gin"
	"net/http"
	error2 "thinkgin/extend/error"
	"time"
)

// Hello Word
// TODO 日志跟控制台输出改写  - 2023年1月2日00:48:55
func HelloWord(c *gin.Context) {
	name := c.Query("name") /*接收参数*/
	//maps := make(map[string]interface{})
	data := make(map[string]interface{})
	now := time.Now()

	code := error2.SUCCESS
	data["data"] = "Hello ThinkGin!"
	data["name"] = name
	data["Time"] = now

	c.JSON(http.StatusOK, gin.H{
		"code": code,
		"msg":  error2.GetMsg(code),
		"data": data,
	})
}
