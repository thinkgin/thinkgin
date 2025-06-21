package route

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"thinkgin/app/index/controller"
	"thinkgin/extend/middleware"
)

func InitRouter() *gin.Engine {
	r := gin.New()
	r.Use(gin.Logger(), gin.Recovery())
	r.Use(middleware.LoggerToFile()) // 记录日志

	gin.SetMode(gin.ReleaseMode)

	// 处理静态文件
	r.Static("/static", "./static")

	// 注册路由
	index := r.Group("/index/")
	{
		index.GET("/hello", controller.HelloWord)
	}

	// 加载模板
	r.LoadHTMLGlob("app/index/view/*")

	// 主页
	r.GET("/", func(c *gin.Context) {
		c.HTML(http.StatusOK, "index.html", gin.H{
			"title": "Thinkgin",
		})
	})

	return r
}
