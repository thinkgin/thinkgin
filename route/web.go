package route

import (
"net/http"

"github.com/gin-gonic/gin"

"thinkgin/app"
"thinkgin/app/index/controller"
)

// RegisterWebRoutes 注册Web页面路由
func RegisterWebRoutes(r *gin.Engine) {
config := app.GetConfig()

// 首页路由
r.GET("/", func(c *gin.Context) {
c.HTML(http.StatusOK, "index.html", gin.H{
"title":   config.App.Name,
"version": config.App.Version,
})
})

// 传统的Web页面路由组 (兼容现有路由)
indexGroup := r.Group("/index")
{
indexGroup.GET("/hello", controller.HelloWord)
// 可以添加更多页面路由:
// indexGroup.GET("/about", controller.About)
// indexGroup.GET("/contact", controller.Contact)
}

// 其他Web页面路由组
// webGroup := r.Group("/web")
// {
//     webGroup.GET("/dashboard", controller.Dashboard)
//     webGroup.GET("/profile", controller.Profile)
// }

// 用户相关页面 (如果需要)
// userWebGroup := r.Group("/user")
// {
//     userWebGroup.GET("/login", controller.LoginPage)
//     userWebGroup.GET("/register", controller.RegisterPage)
//     userWebGroup.GET("/profile", controller.ProfilePage)
// }

// 管理后台页面 (如果需要)
// adminGroup := r.Group("/admin")
// {
//     adminGroup.Use(middleware.AdminAuth()) // 管理员认证中间件
//     adminGroup.GET("/", controller.AdminDashboard)
//     adminGroup.GET("/users", controller.AdminUsers)
// }
}
