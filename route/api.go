package route

import (
	articlecontroller "thinkgin/app/article/controller"
	articlerepo "thinkgin/app/article/repository"
	articleservice "thinkgin/app/article/service"
	"thinkgin/app/database"
	"thinkgin/app/index/controller"
	"thinkgin/extend/middleware"

	"github.com/gin-gonic/gin"
)

// RegisterAPIRoutes 注册 REST API 路由。
// 版本约定：/api/v{n}/<module>/<action>，新版本按需新开 registerXxxAPIV2 等入口。
func RegisterAPIRoutes(r *gin.Engine) {
	api := r.Group("/api")
	api.Use(middleware.APIErrorHandler())
	// 从 middleware.groups.api 配置读取并挂载分组中间件。
	ApplyGroupMiddleware(api, "api")

	v1 := api.Group("/v1")
	registerIndexAPIV1(v1)
	registerArticleAPIV1(v1)
}

// registerIndexAPIV1 示例模块 index 的 v1 路由。
// 扩展新模块时仿造此函数新增一个注册函数即可，保持每个模块自包含。
func registerIndexAPIV1(v1 *gin.RouterGroup) {
	index := v1.Group("/index")
	index.GET("/hello", controller.HelloWord)
}

// registerArticleAPIV1 注册 article 完整业务模块范例的 v1 路由。
//
// 演示要点：
//   - 依赖注入：在装配期构造 repository → service → controller 三层并注入。
//   - 读写分离鉴权：列表/详情公开读取；增删改与发布需 JWT（middleware.JWTAuth）。
//   - 数据库缺省兜底：未配置 DB 时跳过注册，保持"零依赖可启动"。
func registerArticleAPIV1(v1 *gin.RouterGroup) {
	db := database.Default()
	if db == nil {
		// 未配置数据库时不挂载该模块路由，避免运行期空指针。
		return
	}

	ctl := articlecontroller.New(articleservice.New(articlerepo.New(db)))

	articles := v1.Group("/articles")
	{
		// 公开读取
		articles.GET("", ctl.List)
		articles.GET("/:id", ctl.Get)

		// 写操作需 JWT 鉴权
		auth := articles.Group("", middleware.JWTAuth())
		auth.POST("", ctl.Create)
		auth.PUT("/:id", ctl.Update)
		auth.POST("/:id/publish", ctl.Publish)
		auth.DELETE("/:id", ctl.Delete)
	}
}

