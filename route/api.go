package route

import (
	"thinkgin/app/index/controller"
	"thinkgin/extend/middleware"

	"github.com/gin-gonic/gin"
)

// RegisterAPIRoutes 注册 REST API 路由。
// 版本约定：/api/v{n}/<module>/<action>，新版本按需新开 registerXxxAPIV2 等入口。
func RegisterAPIRoutes(r *gin.Engine) {
	api := r.Group("/api")
	api.Use(middleware.APIErrorHandler())

	v1 := api.Group("/v1")
	registerIndexAPIV1(v1)
}

// registerIndexAPIV1 示例模块 index 的 v1 路由。
// 扩展新模块时仿造此函数新增一个注册函数即可，保持每个模块自包含。
func registerIndexAPIV1(v1 *gin.RouterGroup) {
	index := v1.Group("/index")
	index.GET("/hello", controller.HelloWord)
}
