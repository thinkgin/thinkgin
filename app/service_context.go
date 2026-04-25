// 本文件定义 ServiceContext —— ThinkGin 的依赖聚合容器。
//
// 设计目标：
//   - 将 Config / Logger / DB / Cache 等核心依赖聚合到一个显式结构体中，
//     取代分散在各包的全局变量，使依赖关系清晰可测。
//   - 通过 gin.Context 在 HTTP 请求链路中传递，中间件与 Handler 均可访问。
//   - 保留全局 GetConfig / GetLogger 等函数向后兼容，逐步迁移即可。
package app

import (
	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
)

// ginContextKey 用于在 gin.Context 中存取 ServiceContext。
const ginContextKey = "thinkgin:svc"

// ServiceContext 聚合应用运行所需的核心依赖。
// 业务代码应通过 SvcFromGin(c) 从请求链路中获取，而非直接访问全局变量。
type ServiceContext struct {
	Config *GlobalConfig
	Log    Logger
	DB     *gorm.DB     // 默认数据库连接，nil 表示未初始化
	Cache  *redis.Client // 默认 Redis 连接，nil 表示未初始化
}

// SvcMiddleware 返回一个 Gin 中间件，将 ServiceContext 注入到每个请求的 gin.Context 中。
func SvcMiddleware(svc *ServiceContext) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Set(ginContextKey, svc)
		c.Next()
	}
}

// SvcFromGin 从 gin.Context 中取出 ServiceContext。
// 未找到时返回一个基于全局变量构建的兜底实例，保证调用方永远不会拿到 nil。
func SvcFromGin(c *gin.Context) *ServiceContext {
	if v, ok := c.Get(ginContextKey); ok {
		if svc, ok := v.(*ServiceContext); ok {
			return svc
		}
	}
	// 兜底：从全局变量构建，保持向后兼容。
	return &ServiceContext{
		Config: GetConfig(),
		Log:    GetLogger(),
	}
}
