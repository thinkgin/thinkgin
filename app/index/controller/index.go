package controller

import (
	"github.com/gin-gonic/gin"
	"net/http"
	"time"
)

// Hello ThinkGin
func HelloWord(c *gin.Context) {
	//name := c.Query("name") /*接收参数*/
	data := make(map[string]interface{})
	now := time.Now()

	data["data"] = "Hello ThinkGin!"
	data["name"] = "who test,who care!"
	data["Time"] = now

	c.JSON(http.StatusOK, gin.H{
		"code": 200,
		"msg":  "SUCCESS",
		"data": data,
	})
}
