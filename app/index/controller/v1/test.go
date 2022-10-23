package v1

import (
	"github.com/gin-gonic/gin"
	"net/http"
	error2 "thinkgin/extend/error"
)

// 获取多个文章标签
func HelloWord(c *gin.Context) {
	name := c.Query("name") /*接收参数*/
	//maps := make(map[string]interface{})
	data := make(map[string]interface{})

	code := error2.SUCCESS
	data["data"] = "Hello Word!"
	data["name"] = name

	c.JSON(http.StatusOK, gin.H{
		"code": code,
		"msg":  error2.GetMsg(code),
		"data": data,
	})
}
