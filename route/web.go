package route

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"thinkgin/app"
	"thinkgin/app/index/controller"
)

// RegisterWebRoutes 注册 HTML 页面路由。
// 与 API 路由保持物理分离，便于未来独立服务化或下线。
func RegisterWebRoutes(r *gin.Engine) {
	cfg := app.GetConfig()

	// 首页：渲染 app/index/view/index.html。
	r.GET("/", func(c *gin.Context) {
		c.HTML(http.StatusOK, "index.html", gin.H{
			"title":   cfg.App.Name,
			"version": cfg.App.Version,
		})
	})

	// 兼容旧版 /index/hello 页面跳转到 controller，方便演示请求闭环。
	index := r.Group("/index")
	index.GET("/hello", controller.HelloWord)
}
