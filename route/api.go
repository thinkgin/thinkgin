package route

import (
	"thinkgin/app/index/controller"
	"thinkgin/extend/middleware"

	"github.com/gin-gonic/gin"
)

// RegisterAPIRoutes 注册API路由
func RegisterAPIRoutes(r *gin.Engine) {
	// API根路径
	api := r.Group("/api")
	api.Use(middleware.APIErrorHandler())
	{
		// API v1 版本
		v1 := api.Group("/v1")
		{
			// 注册各个模块的API路由
			registerIndexAPIV1(v1)
			// registerUserAPIV1(v1)
			// registerProductAPIV1(v1)
		}

		// API v2 版本 (未来扩展)
		// v2 := api.Group("/v2")
		// {
		//     registerIndexAPIV2(v2)
		//     registerUserAPIV2(v2)
		// }
	}
}

// registerIndexAPIV1 注册Index模块的API v1路由
func registerIndexAPIV1(v1 *gin.RouterGroup) {
	indexAPI := v1.Group("/index")
	{
		indexAPI.GET("/hello", controller.HelloWord)
		// 可以添加更多RESTful API:
		// indexAPI.GET("/info", controller.GetInfo)
		// indexAPI.POST("/data", controller.CreateData)
		// indexAPI.PUT("/data/:id", controller.UpdateData)
		// indexAPI.DELETE("/data/:id", controller.DeleteData)
	}
}

// registerUserAPIV1 注册用户模块的API v1路由 (示例)
// func registerUserAPIV1(v1 *gin.RouterGroup) {
//     userAPI := v1.Group("/users")
//     {
//         // 用户认证相关
//         userAPI.POST("/register", controller.Register)
//         userAPI.POST("/login", controller.Login)
//         userAPI.POST("/logout", controller.Logout)
//         userAPI.POST("/refresh", controller.RefreshToken)
//
//         // 需要认证的用户接口
//         authRequired := userAPI.Group("/")
//         authRequired.Use(middleware.JWTAuth())
//         {
//             authRequired.GET("/profile", controller.GetProfile)
//             authRequired.PUT("/profile", controller.UpdateProfile)
//             authRequired.POST("/avatar", controller.UploadAvatar)
//             authRequired.GET("/", controller.GetUsers)        // 获取用户列表
//             authRequired.GET("/:id", controller.GetUser)      // 获取单个用户
//             authRequired.PUT("/:id", controller.UpdateUser)   // 更新用户
//             authRequired.DELETE("/:id", controller.DeleteUser) // 删除用户
//         }
//     }
// }

// registerProductAPIV1 注册产品模块的API v1路由 (示例)
// func registerProductAPIV1(v1 *gin.RouterGroup) {
//     productAPI := v1.Group("/products")
//     {
//         // 公开接口
//         productAPI.GET("/", controller.GetProducts)          // 获取产品列表
//         productAPI.GET("/:id", controller.GetProduct)        // 获取单个产品
//         productAPI.GET("/category/:id", controller.GetProductsByCategory) // 按分类获取产品
//
//         // 需要认证的接口
//         authRequired := productAPI.Group("/")
//         authRequired.Use(middleware.JWTAuth())
//         {
//             authRequired.POST("/", controller.CreateProduct)     // 创建产品
//             authRequired.PUT("/:id", controller.UpdateProduct)   // 更新产品
//             authRequired.DELETE("/:id", controller.DeleteProduct) // 删除产品
//         }
//
//         // 需要管理员权限的接口
//         adminRequired := productAPI.Group("/admin")
//         adminRequired.Use(middleware.JWTAuth(), middleware.AdminAuth())
//         {
//             adminRequired.GET("/stats", controller.GetProductStats)
//             adminRequired.POST("/batch", controller.BatchCreateProducts)
//         }
//     }
// }
