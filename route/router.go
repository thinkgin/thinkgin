package route

import (
	"github.com/gin-gonic/gin"
	"net/http"
	"thinkgin/app/index/controller"
	"thinkgin/extend/middleware"
)

/*核心路由*/
func InitRouter() *gin.Engine {
	r := gin.New()
	//r.Use(gin.Logger(), gin.Recovery())
	r.Use(gin.Logger(), gin.Recovery(), middleware.LoggerToFile())
	gin.SetMode("release")
	//gin.SetMode("debug")
	//gin.SetMode(setting.RunMode)
	//——————————————————————————————————————分割线-下面是首页数据数据——————————————————————————————————————————
	index := r.Group("/index/")
	{
		index.GET("/hello", controller.HelloWord) /*测试输出Hello Word*/
	}

	r.LoadHTMLGlob("app/index/view/*")
	r.GET("/", func(c *gin.Context) {
		c.HTML(http.StatusOK, "index.html", gin.H{"title": "Thinkgin"})
	})
	return r
}
