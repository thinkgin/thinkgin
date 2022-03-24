package route

import (
	"github.com/gin-gonic/gin"
	v1 "thinkgin/app/index/controller/v1"
	"thinkgin/extend/setting"
)

/*核心路由*/
func InitRouter() *gin.Engine {
	r := gin.New()

	r.Use(gin.Logger(), gin.Recovery())

	gin.SetMode(setting.RunMode)

	apiv1 := r.Group("/api/v1")
	{
		//获取标签列表
		apiv1.GET("/tags", v1.GetTags)
		//新增标签
		apiv1.POST("/tags", v1.AddTag)
		//编辑标签
		apiv1.PUT("/tags/:id", v1.EditTag)
		//删除标签
		apiv1.DELETE("/tags/:id", v1.DelTag)
	}
	//r.GET("/test", func(c *gin.Context) {
	//	c.JSON(200, gin.H{
	//		"message": "hello,router",
	//	})
	//})

	return r
}
