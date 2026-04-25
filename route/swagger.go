// 本文件提供 OpenAPI/Swagger 文档端点。
//
// 设计思路（零依赖方式）：
//   - /swagger/doc.json 或 /swagger/doc.yaml: 直接挂载 docs/ 下的 OpenAPI 规范文件。
//   - /swagger/: 返回 Swagger UI HTML，通过 CDN 加载 swagger-ui。
//   - 不引入 swaggo 等代码生成工具，保持框架轻量；
//     用户可选择手写 openapi.yaml 或后期集成 swag init。
//
// 启用条件：config.app.debug = true 或 config.app.swagger_enabled = true。
package route

import (
	"net/http"
	"os"
	"path/filepath"

	"thinkgin/app"

	"github.com/gin-gonic/gin"
)

const swaggerUIHTML = `<!DOCTYPE html>
<html lang="en">
<head>
  <meta charset="UTF-8">
  <title>ThinkGin API Docs</title>
  <link rel="stylesheet" href="https://unpkg.com/swagger-ui-dist@5/swagger-ui.css">
</head>
<body>
  <div id="swagger-ui"></div>
  <script src="https://unpkg.com/swagger-ui-dist@5/swagger-ui-bundle.js"></script>
  <script>
    SwaggerUIBundle({
      url: "./doc.yaml",
      dom_id: '#swagger-ui',
      presets: [SwaggerUIBundle.presets.apis, SwaggerUIBundle.SwaggerUIStandalonePreset],
      layout: "BaseLayout"
    })
  </script>
</body>
</html>`

// registerSwaggerRoutes 注册 Swagger/OpenAPI 相关端点。
// 仅在 Debug 模式或显式开启时注册，生产环境不暴露。
func registerSwaggerRoutes(r *gin.Engine, cfg *app.GlobalConfig) {
	if !cfg.App.Debug && !cfg.App.SwaggerEnabled {
		return
	}

	// 查找 docs/ 目录下的 openapi 规范文件（yaml 优先，json 备选）。
	docsDir := "docs"
	yamlPath := filepath.Join(docsDir, "openapi.yaml")
	jsonPath := filepath.Join(docsDir, "openapi.json")

	r.GET("/swagger/", func(c *gin.Context) {
		c.Data(http.StatusOK, "text/html; charset=utf-8", []byte(swaggerUIHTML))
	})

	r.GET("/swagger/doc.yaml", func(c *gin.Context) {
		if _, err := os.Stat(yamlPath); err != nil {
			c.JSON(http.StatusNotFound, gin.H{"message": "docs/openapi.yaml not found"})
			return
		}
		c.File(yamlPath)
	})

	r.GET("/swagger/doc.json", func(c *gin.Context) {
		if _, err := os.Stat(jsonPath); err != nil {
			c.JSON(http.StatusNotFound, gin.H{"message": "docs/openapi.json not found"})
			return
		}
		c.File(jsonPath)
	})
}
