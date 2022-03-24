package route

import (
	"github.com/gin-gonic/gin"
	"thinkgin/extend/setting"
)

func InitRouter() *gin.Engine {
	r := gin.New()

	r.Use(gin.Logger(), gin.Recovery())

	gin.SetMode(setting.RunMode)

	r.GET("/test", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"message": "hello,router",
		})
	})

	return r
}
