package controller

import (
	"github.com/gin-gonic/gin"
	"net/http"
	"time"
)

// Hello Word
// TODO 日志跟控制台输出改写  - 2023年1月2日00:48:55 -输出的时候输出一段ThinkGin的图画-MarkDown这一个版本也要更新一下，然后就是写一个文档
func HelloWord(c *gin.Context) {
	name := c.Query("name") /*接收参数*/
	//maps := make(map[string]interface{})
	data := make(map[string]interface{})
	now := time.Now()

	data["data"] = "Hello ThinkGin!"
	data["name"] = name
	data["Time"] = now

	c.JSON(http.StatusOK, gin.H{
		"code": 200,
		"msg":  "SUCCESS",
		"data": data,
	})
}
