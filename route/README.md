# ThinkGin 路由架构设计文档

![路由架构](https://img.shields.io/badge/Route-Architecture-blue?style=flat-square)
![版本](https://img.shields.io/badge/Version-2.0-green?style=flat-square)
![状态](https://img.shields.io/badge/Status-Production%20Ready-brightgreen?style=flat-square)

## ��� 目录

- [���️ 架构概览](#️-架构概览)
- [��� 文件结构](#-文件结构)
- [��� 设计原则](#-设计原则)
- [���️ 路由规范](#️-路由规范)
- [���️ 中间件系统](#️-中间件系统)
- [��� 开发指南](#-开发指南)
- [��� 配置说明](#-配置说明)
- [��� 监控集成](#-监控集成)
- [��� 扩展示例](#-扩展示例)
- [❓ 常见问题](#-常见问题)

---

## ���️ 架构概览

ThinkGin 2.0 采用**分层模块化**的路由架构，将不同类型的路由分离到独立文件中，实现高内聚、低耦合的设计。

### �� 核心优势

| 优势 | 说明 | 价值 |
|------|------|------|
| **职责分离** | 不同类型路由分文件管理 | 提高代码可读性和维护性 |
| **版本控制** | API支持多版本并存 | 保证向后兼容，平滑升级 |
| **模块化** | 业务模块独立注册 | 便于团队协作开发 |
| **配置化** | 路由行为可配置 | 适应不同环境需求 |
| **监控友好** | 内置性能监控 | 便于运维和故障排查 |

---

## ��� 文件结构

```
route/
├── ��� router.go     # ��� 主路由入口 - 中间件、静态资源、路由分发
├── ��� web.go        # ��� Web页面路由 - HTML页面渲染
├── ��� api.go        # ��� API接口路由 - JSON数据接口  
└── ��� README.md     # ��� 架构文档 - 本文档
```

### ��� 文件职责详解

#### ��� router.go - 主入口控制器
```
核心职责:
- ��� 全局中间件注册 (日志、恢复、监控等)
- ��� 静态资源配置 (CSS、JS、图片等)
- ��� 模板引擎加载 (HTML模板文件)
- ��� 路由分发调度 (调用各模块路由注册)
- ��� 监控端点管理 (健康检查、Prometheus等)
- ⚙️ 配置驱动初始化 (根据配置文件设置行为)
```

#### ��� web.go - 页面路由处理器
```
专门处理:
- ��� 网站首页 (返回HTML页面)
- ��� HTML页面渲染 (用户界面)
- ���️ 静态页面展示 (关于、联系等)
- ��� 用户界面路由 (登录、注册、个人中心)
- ���️ 管理后台页面 (仪表板、用户管理)
- ��� 移动端页面适配 (响应式设计)
```

#### ��� api.go - API接口处理器  
```
专门处理:
- ��� RESTful API接口 (标准HTTP方法)
- ��� JSON数据交换 (API响应格式)
- ��� 认证授权接口 (登录、权限验证)
- ��� 数据查询接口 (获取业务数据)
- ��� 数据操作接口 (增删改操作)
- ��� 第三方API集成 (外部服务对接)
```

---

## ��� 设计原则

### 1. ��� **单一职责原则 (SRP)**
每个文件只负责一种类型的路由，避免功能混杂：

```go
// ❌ 不推荐：混合在一个文件
func InitRouter() {
    // 页面路由
    r.GET("/", homePage)
    // API路由  
    r.GET("/api/users", getUsers)
    // 监控路由
    r.GET("/metrics", metrics)
}

// ✅ 推荐：职责分离
func InitRouter() {
    registerWebRoutes(r)     // 页面路由
    registerAPIRoutes(r)     // API路由
    registerMonitorRoutes(r) // 监控路由
}
```

### 2. ��� **开闭原则 (OCP)**
对扩展开放，对修改封闭：

```go
// ✅ 添加新模块时，只需要新增注册函数
func RegisterAPIRoutes(r *gin.Engine) {
    v1 := r.Group("/api/v1")
    {
        registerUserAPIV1(v1)     // 现有模块
        registerProductAPIV1(v1)  // 新增模块，无需修改现有代码
    }
}
```

### 3. ��� **接口隔离原则 (ISP)**
不同的客户端使用不同的接口：

```go
// Web客户端路由
r.GET("/login", loginPage)        // 返回HTML
r.GET("/dashboard", dashboardPage) // 返回HTML

// API客户端路由  
api.POST("/auth/login", apiLogin)     // 返回JSON
api.GET("/user/profile", apiProfile)  // 返回JSON
```

---

## ���️ 路由规范

### ��� Web页面路由规范

#### **命名规范**
```go
// ✅ 推荐的命名方式
/                    # 首页
/about               # 关于页面
/contact             # 联系页面
/user/profile        # 用户资料页面
/admin/dashboard     # 管理后台首页

// ❌ 避免的命名方式
/index.html          # 避免文件扩展名
/getUserProfile      # 避免驼峰命名
/user_profile        # 避免下划线
```

#### **参数传递**
```go
// 路径参数
r.GET("/user/:id", userProfile)           // /user/123
r.GET("/post/:id/:action", postAction)    // /post/456/edit

// 查询参数
r.GET("/search", searchPage)              // /search?q=golang&page=1

// 表单参数
r.POST("/user/update", updateProfile)     // form data
```

#### **完整示例**
```go
func RegisterWebRoutes(r *gin.Engine) {
    // 首页和基础页面
    r.GET("/", homePage)
    r.GET("/about", aboutPage)
    r.GET("/contact", contactPage)
    
    // 用户相关页面
    userGroup := r.Group("/user")
    {
        userGroup.GET("/login", loginPage)
        userGroup.GET("/register", registerPage)
        userGroup.GET("/profile", profilePage)
        userGroup.GET("/settings", settingsPage)
    }
    
    // 管理后台页面
    adminGroup := r.Group("/admin")
    adminGroup.Use(middleware.AdminAuth()) // 管理员权限中间件
    {
        adminGroup.GET("/", adminDashboard)
        adminGroup.GET("/users", adminUsers)
        adminGroup.GET("/settings", adminSettings)
    }
}
```

### ��� API路由规范

#### **RESTful API 设计规范**

| HTTP方法 | 路径 | 描述 | 示例 |
|----------|------|------|------|
| GET | `/api/v1/users` | 获取用户列表 | `GET /api/v1/users?page=1&limit=10` |
| GET | `/api/v1/users/:id` | 获取单个用户 | `GET /api/v1/users/123` |
| POST | `/api/v1/users` | 创建新用户 | `POST /api/v1/users` |
| PUT | `/api/v1/users/:id` | 完整更新用户 | `PUT /api/v1/users/123` |
| PATCH | `/api/v1/users/:id` | 部分更新用户 | `PATCH /api/v1/users/123` |
| DELETE | `/api/v1/users/:id` | 删除用户 | `DELETE /api/v1/users/123` |

#### **版本控制策略**

```go
// ✅ 推荐：URL版本控制
/api/v1/users        # 版本1
/api/v2/users        # 版本2

// ✅ 备选：请求头版本控制  
GET /api/users
Accept: application/vnd.thinkgin.v1+json

// ❌ 不推荐：参数版本控制
/api/users?version=1
```

#### **状态码规范**

| 状态码 | 含义 | 使用场景 |
|--------|------|----------|
| 200 | OK | 成功获取资源 |
| 201 | Created | 成功创建资源 |
| 204 | No Content | 成功删除资源 |
| 400 | Bad Request | 请求参数错误 |
| 401 | Unauthorized | 未认证 |
| 403 | Forbidden | 无权限 |
| 404 | Not Found | 资源不存在 |
| 409 | Conflict | 资源冲突 |
| 422 | Unprocessable Entity | 验证失败 |
| 500 | Internal Server Error | 服务器错误 |

#### **响应格式标准**

```json
// 成功响应格式
{
    "code": 200,
    "message": "success",
    "data": {
        "id": 123,
        "name": "张三",
        "email": "zhangsan@example.com"
    },
    "timestamp": "2023-12-01T10:30:00Z"
}

// 错误响应格式
{
    "code": 400,
    "message": "参数验证失败",
    "errors": [
        {
            "field": "email",
            "message": "邮箱格式不正确"
        }
    ],
    "timestamp": "2023-12-01T10:30:00Z"
}

// 列表响应格式
{
    "code": 200,
    "message": "success", 
    "data": [
        {"id": 1, "name": "用户1"},
        {"id": 2, "name": "用户2"}
    ],
    "pagination": {
        "page": 1,
        "limit": 10,
        "total": 100,
        "total_pages": 10
    },
    "timestamp": "2023-12-01T10:30:00Z"
}
```

---

## ���️ 中间件系统

### ��� 全局中间件

在 `router.go` 中注册的全局中间件，应用于所有路由：

```go
func registerGlobalMiddleware(r *gin.Engine) {
    // 基础中间件
    r.Use(gin.Logger())                    // Gin默认日志
    r.Use(gin.Recovery())                  // 恐慌恢复
    
    // 自定义中间件
    r.Use(middleware.LoggerToFile())       // 文件日志记录
    r.Use(middleware.PrometheusMiddleware()) // 性能监控
    r.Use(middleware.CORS())               // 跨域处理
    r.Use(middleware.RequestID())          // 请求ID追踪
    r.Use(middleware.RateLimit())          // 访问限流
}
```

### ��� 路由组中间件

针对特定路由组的中间件：

```go
// API认证中间件
apiAuth := r.Group("/api/v1")
apiAuth.Use(middleware.JWTAuth())
{
    apiAuth.GET("/profile", controller.GetProfile)
    apiAuth.PUT("/profile", controller.UpdateProfile)
}

// 管理员权限中间件
adminGroup := r.Group("/admin")
adminGroup.Use(middleware.AdminAuth())
{
    adminGroup.GET("/users", controller.AdminUsers)
    adminGroup.DELETE("/users/:id", controller.DeleteUser)
}

// API限流中间件
apiLimited := r.Group("/api/v1/public")
apiLimited.Use(middleware.APIRateLimit(100)) // 每分钟100次
{
    apiLimited.GET("/posts", controller.GetPosts)
    apiLimited.GET("/categories", controller.GetCategories)
}
```

### ��� 自定义中间件开发

#### **认证中间件示例**

```go
// extend/middleware/auth.go
func JWTAuth() gin.HandlerFunc {
    return func(c *gin.Context) {
        token := c.GetHeader("Authorization")
        if token == "" {
            c.JSON(401, gin.H{"error": "未提供认证令牌"})
            c.Abort()
            return
        }
        
        // 验证JWT令牌
        claims, err := validateJWT(token)
        if err != nil {
            c.JSON(401, gin.H{"error": "无效的认证令牌"})
            c.Abort()
            return
        }
        
        // 将用户信息存储到上下文
        c.Set("user_id", claims.UserID)
        c.Set("username", claims.Username)
        c.Next()
    }
}
```

#### **权限中间件示例**

```go
// extend/middleware/permission.go
func RequirePermission(permission string) gin.HandlerFunc {
    return func(c *gin.Context) {
        userID := c.GetInt("user_id")
        if userID == 0 {
            c.JSON(403, gin.H{"error": "未认证用户"})
            c.Abort()
            return
        }
        
        // 检查用户权限
        hasPermission := checkUserPermission(userID, permission)
        if !hasPermission {
            c.JSON(403, gin.H{"error": "权限不足"})
            c.Abort()
            return
        }
        
        c.Next()
    }
}

// 使用示例
adminGroup.Use(middleware.RequirePermission("admin.users.manage"))
```

#### **限流中间件示例**

```go
// extend/middleware/ratelimit.go
func RateLimit(maxRequests int) gin.HandlerFunc {
    limiter := rate.NewLimiter(rate.Limit(maxRequests), maxRequests)
    
    return func(c *gin.Context) {
        if !limiter.Allow() {
            c.JSON(429, gin.H{
                "error": "请求过于频繁，请稍后再试",
                "retry_after": "60s",
            })
            c.Abort()
            return
        }
        c.Next()
    }
}
```

---

## ��� 开发指南

### ��� 添加新的Web页面路由

#### **步骤1：创建控制器**

```go
// app/user/controller/profile.go
func ProfilePage(c *gin.Context) {
    userID := c.Param("id")
    
    // 获取用户数据
    user, err := getUserByID(userID)
    if err != nil {
        c.HTML(404, "error.html", gin.H{
            "error": "用户不存在",
        })
        return
    }
    
    // 渲染页面
    c.HTML(200, "profile.html", gin.H{
        "user": user,
        "title": "用户资料",
    })
}
```

#### **步骤2：添加路由**

在 `route/web.go` 中添加：

```go
func RegisterWebRoutes(r *gin.Engine) {
    // 现有路由...
    
    // 用户相关页面
    userGroup := r.Group("/user")
    {
        userGroup.GET("/:id/profile", controller.ProfilePage)  // 新增
        userGroup.GET("/:id/settings", controller.SettingsPage)
    }
}
```

#### **步骤3：创建模板**

```html
<!-- app/user/view/profile.html -->
<!DOCTYPE html>
<html>
<head>
    <title>{{.title}}</title>
</head>
<body>
    <h1>{{.user.Name}}的资料</h1>
    <p>邮箱：{{.user.Email}}</p>
    <p>注册时间：{{.user.CreatedAt}}</p>
</body>
</html>
```

### ��� 添加新的API接口

#### **步骤1：创建API控制器**

```go
// app/user/controller/api.go
func GetUserProfile(c *gin.Context) {
    userID := c.GetInt("user_id") // 从JWT中间件获取
    
    // 获取用户数据
    user, err := getUserByID(userID)
    if err != nil {
        c.JSON(404, gin.H{
            "code": 404,
            "message": "用户不存在",
        })
        return
    }
    
    // 返回JSON响应
    c.JSON(200, gin.H{
        "code": 200,
        "message": "success",
        "data": user,
    })
}

func UpdateUserProfile(c *gin.Context) {
    userID := c.GetInt("user_id")
    
    // 绑定请求数据
    var req struct {
        Name  string `json:"name" binding:"required"`
        Email string `json:"email" binding:"required,email"`
    }
    
    if err := c.ShouldBindJSON(&req); err != nil {
        c.JSON(400, gin.H{
            "code": 400,
            "message": "参数验证失败",
            "errors": err.Error(),
        })
        return
    }
    
    // 更新用户数据
    err := updateUser(userID, req.Name, req.Email)
    if err != nil {
        c.JSON(500, gin.H{
            "code": 500,
            "message": "更新失败",
        })
        return
    }
    
    c.JSON(200, gin.H{
        "code": 200,
        "message": "更新成功",
    })
}
```

#### **步骤2：创建路由注册函数**

在 `route/api.go` 中添加：

```go
// registerUserAPIV1 注册用户模块的API v1路由
func registerUserAPIV1(v1 *gin.RouterGroup) {
    userAPI := v1.Group("/users")
    {
        // 公开接口
        userAPI.POST("/register", controller.Register)
        userAPI.POST("/login", controller.Login)
        
        // 需要认证的接口
        authRequired := userAPI.Group("/")
        authRequired.Use(middleware.JWTAuth())
        {
            authRequired.GET("/profile", controller.GetUserProfile)     // 新增
            authRequired.PUT("/profile", controller.UpdateUserProfile)  // 新增
            authRequired.POST("/avatar", controller.UploadAvatar)
        }
    }
}
```

#### **步骤3：注册到主路由**

在 `RegisterAPIRoutes` 函数中调用：

```go
func RegisterAPIRoutes(r *gin.Engine) {
    api := r.Group("/api")
    {
        v1 := api.Group("/v1")
        {
            registerIndexAPIV1(v1)
            registerUserAPIV1(v1)  // 新增这行
        }
    }
}
```

### ��� 添加业务模块

#### **完整的产品模块示例**

```go
// registerProductAPIV1 注册产品模块的API v1路由
func registerProductAPIV1(v1 *gin.RouterGroup) {
    productAPI := v1.Group("/products")
    {
        // 公开接口 - 无需认证
        productAPI.GET("/", controller.GetProducts)           // 获取产品列表
        productAPI.GET("/:id", controller.GetProduct)         // 获取单个产品
        productAPI.GET("/category/:id", controller.GetProductsByCategory)
        productAPI.GET("/search", controller.SearchProducts)
        
        // 需要认证的接口
        authRequired := productAPI.Group("/")
        authRequired.Use(middleware.JWTAuth())
        {
            authRequired.POST("/", controller.CreateProduct)     // 创建产品
            authRequired.PUT("/:id", controller.UpdateProduct)   // 更新产品
            authRequired.DELETE("/:id", controller.DeleteProduct) // 删除产品
            authRequired.POST("/:id/favorite", controller.FavoriteProduct) // 收藏
        }
        
        // 需要管理员权限的接口
        adminRequired := productAPI.Group("/admin")
        adminRequired.Use(middleware.JWTAuth(), middleware.RequirePermission("admin.products"))
        {
            adminRequired.GET("/stats", controller.GetProductStats)
            adminRequired.POST("/batch", controller.BatchCreateProducts)
            adminRequired.PUT("/:id/approve", controller.ApproveProduct)
            adminRequired.PUT("/:id/reject", controller.RejectProduct)
        }
        
        // API限流的公开接口
        publicLimited := productAPI.Group("/public")
        publicLimited.Use(middleware.RateLimit(60)) // 每分钟60次
        {
            publicLimited.GET("/trending", controller.GetTrendingProducts)
            publicLimited.GET("/recommendations", controller.GetRecommendations)
        }
    }
}
```

---

## ��� 配置说明

### ⚙️ 路由相关配置

#### **服务器配置 (config/server.yaml)**

```yaml
server:
  http:
    host: "0.0.0.0"
    port: 8000
    read_timeout: 60
    write_timeout: 60
    max_header_bytes: 1048576
  
  # 运行模式：debug, test, release
  mode: "debug"
  
  # 静态资源配置
  static:
    path: "./public"
    route: "/public"
    max_age: "24h"
  
  # 上传文件配置
  upload:
    path: "./public/uploads"
    max_size: "10MB"
    allowed_types: ["jpg", "png", "gif", "pdf"]
```

#### **中间件配置 (config/middleware.yaml)**

```yaml
middleware:
  # CORS跨域配置
  cors:
    enabled: true
    allow_origins: ["*"]
    allow_methods: ["GET", "POST", "PUT", "DELETE", "OPTIONS"]
    allow_headers: ["*"]
    expose_headers: ["Content-Length"]
    max_age: 12
  
  # JWT认证配置
  jwt:
    secret: "your-secret-key"
    expire: 3600  # 1小时
    refresh_expire: 86400  # 24小时
  
  # 限流配置
  rate_limit:
    enabled: true
    requests_per_minute: 100
    burst: 20
  
  # 请求日志配置
  request_log:
    enabled: true
    skip_paths: ["/ping", "/metrics"]
    log_request_body: false
    log_response_body: false
```

### �� 动态配置路由

```go
func registerConfigurableRoutes(r *gin.Engine) {
    config := app.GetConfig()
    
    // 根据配置启用/禁用功能
    if config.Features.UserManagement {
        registerUserRoutes(r)
    }
    
    if config.Features.ProductCatalog {
        registerProductRoutes(r)
    }
    
    if config.Features.OrderSystem {
        registerOrderRoutes(r)
    }
    
    // 根据环境配置不同行为
    if config.Environment == "development" {
        // 开发环境专用路由
        debugGroup := r.Group("/debug")
        {
            debugGroup.GET("/routes", showAllRoutes)
            debugGroup.GET("/config", showConfig)
        }
    }
}
```

---

## ��� 监控集成

### ��� 健康检查

```go
// 基础健康检查
r.GET("/ping", func(c *gin.Context) {
    c.JSON(200, gin.H{
        "status": "ok",
        "message": "pong",
        "timestamp": time.Now().Unix(),
    })
})

// 详细健康检查
r.GET("/health", func(c *gin.Context) {
    health := checkSystemHealth()
    
    status := 200
    if !health.Healthy {
        status = 503
    }
    
    c.JSON(status, gin.H{
        "healthy": health.Healthy,
        "version": config.App.Version,
        "uptime": health.Uptime,
        "database": health.Database,
        "redis": health.Redis,
        "disk_space": health.DiskSpace,
        "memory_usage": health.MemoryUsage,
    })
})
```

### �� Prometheus监控

#### **自动监控指标**

```go
// HTTP请求指标自动收集
- http_requests_total        # 请求总数
- http_request_duration_seconds  # 请求耗时
- http_request_size_bytes    # 请求大小
- http_response_size_bytes   # 响应大小
```

#### **业务指标收集**

```go
// 在控制器中添加业务指标
func CreateUser(c *gin.Context) {
    start := time.Now()
    
    // 业务逻辑...
    user, err := userService.CreateUser(req)
    
    // 记录业务指标
    if counter := middleware.BusinessCounter("user_created_total", []string{"status"}); counter != nil {
        if err != nil {
            counter.WithLabelValues("error").Inc()
        } else {
            counter.WithLabelValues("success").Inc()
        }
    }
    
    // 记录处理时间
    if histogram := middleware.BusinessHistogram("user_creation_duration", []string{"status"}, nil); histogram != nil {
        status := "success"
        if err != nil {
            status = "error"
        }
        histogram.WithLabelValues(status).Observe(time.Since(start).Seconds())
    }
    
    // 返回响应...
}
```

### ��� 监控仪表板

```promql
# 常用Prometheus查询

# 请求QPS
rate(http_requests_total[5m])

# 平均响应时间
rate(http_request_duration_seconds_sum[5m]) / rate(http_request_duration_seconds_count[5m])

# 错误率
rate(http_requests_total{status=~"5.."}[5m]) / rate(http_requests_total[5m])

# P95响应时间
histogram_quantile(0.95, rate(http_request_duration_seconds_bucket[5m]))

# 按路径分组的请求量
sum(rate(http_requests_total[5m])) by (path)
```

---

## ��� 扩展示例

### ��� WebSocket路由

```go
// route/websocket.go
func RegisterWebSocketRoutes(r *gin.Engine) {
    ws := r.Group("/ws")
    {
        ws.GET("/chat", handleChatWebSocket)
        ws.GET("/notifications", handleNotificationWebSocket)
    }
}

func handleChatWebSocket(c *gin.Context) {
    conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
    if err != nil {
        return
    }
    defer conn.Close()
    
    // WebSocket处理逻辑
    for {
        messageType, message, err := conn.ReadMessage()
        if err != nil {
            break
        }
        
        // 处理消息
        response := processChatMessage(message)
        
        err = conn.WriteMessage(messageType, response)
        if err != nil {
            break
        }
    }
}
```

### ��� 文件上传路由

```go
// route/upload.go
func RegisterUploadRoutes(r *gin.Engine) {
    upload := r.Group("/upload")
    upload.Use(middleware.JWTAuth()) // 需要认证
    {
        upload.POST("/image", handleImageUpload)
        upload.POST("/document", handleDocumentUpload)
        upload.POST("/avatar", handleAvatarUpload)
    }
}

func handleImageUpload(c *gin.Context) {
    file, header, err := c.Request.FormFile("image")
    if err != nil {
        c.JSON(400, gin.H{"error": "上传文件失败"})
        return
    }
    defer file.Close()
    
    // 验证文件类型
    if !isValidImageType(header.Filename) {
        c.JSON(400, gin.H{"error": "不支持的文件类型"})
        return
    }
    
    // 保存文件
    filename := generateUniqueFilename(header.Filename)
    filePath := filepath.Join("public/uploads/images", filename)
    
    out, err := os.Create(filePath)
    if err != nil {
        c.JSON(500, gin.H{"error": "保存文件失败"})
        return
    }
    defer out.Close()
    
    _, err = io.Copy(out, file)
    if err != nil {
        c.JSON(500, gin.H{"error": "保存文件失败"})
        return
    }
    
    c.JSON(200, gin.H{
        "message": "上传成功",
        "url": "/public/uploads/images/" + filename,
    })
}
```

### ��� GraphQL路由

```go
// route/graphql.go
func RegisterGraphQLRoutes(r *gin.Engine) {
    schema := buildGraphQLSchema()
    
    graphql := r.Group("/graphql")
    {
        // GraphQL查询端点
        graphql.POST("/", handleGraphQLQuery(schema))
        
        // GraphQL Playground (开发环境)
        if gin.Mode() == gin.DebugMode {
            graphql.GET("/playground", handleGraphQLPlayground)
        }
    }
}
```

---

## ❓ 常见问题

### ��� 问题排查

#### **Q: 路由不生效，返回404**

```go
// 检查项目：
1. 确认路由注册函数被调用
2. 检查路由路径是否正确
3. 确认中间件没有提前终止请求
4. 检查路由组前缀是否正确

// 调试方法：
func showAllRoutes(c *gin.Context) {
    routes := c.Engine.Routes()
    c.JSON(200, routes)
}
```

#### **Q: 中间件不生效**

```go
// 检查项目：
1. 中间件注册顺序
2. 中间件是否调用了 c.Next()
3. 路由组中间件作用域
4. 全局中间件与局部中间件冲突

// 调试中间件：
func debugMiddleware() gin.HandlerFunc {
    return func(c *gin.Context) {
        fmt.Printf("Before: %s %s\n", c.Request.Method, c.Request.URL.Path)
        c.Next()
        fmt.Printf("After: %d\n", c.Writer.Status())
    }
}
```

#### **Q: CORS跨域问题**

```go
// 解决方案：
func setupCORS() gin.HandlerFunc {
    return cors.New(cors.Config{
        AllowOrigins:     []string{"*"},
        AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
        AllowHeaders:     []string{"*"},
        ExposeHeaders:    []string{"Content-Length"},
        AllowCredentials: true,
        MaxAge:           12 * time.Hour,
    })
}

// 注册：
r.Use(setupCORS())
```

#### **Q: 性能优化建议**

```go
// 1. 路由优化
- 使用路由组减少路径匹配开销
- 避免过深的路由嵌套
- 合理使用通配符路由

// 2. 中间件优化
- 只在需要的路由组使用中间件
- 避免重复的中间件逻辑
- 使用缓存减少重复计算

// 3. 响应优化
- 使用合适的响应格式 (JSON vs XML)
- 启用 Gzip 压缩
- 设置合理的缓存头
```

### ��� 最佳实践总结

1. **��� 职责明确**: 每个路由文件只处理特定类型的请求
2. **��� 命名规范**: 使用清晰、一致的路由命名规则
3. **��� 安全考虑**: 合理使用认证和权限中间件
4. **��� 监控完备**: 为关键路由添加监控指标
5. **��� 配置驱动**: 通过配置文件控制路由行为
6. **��� 文档维护**: 及时更新API文档和路由说明
7. **��� 测试覆盖**: 为每个路由编写单元测试
8. **⚡ 性能优化**: 关注路由性能和响应时间

---

## ��� 参考文档

- [Gin框架官方文档](https://gin-gonic.com/)
- [RESTful API设计指南](https://restfulapi.net/)
- [HTTP状态码参考](https://developer.mozilla.org/en-US/docs/Web/HTTP/Status)
- [JWT认证标准](https://jwt.io/)
- [CORS规范](https://developer.mozilla.org/en-US/docs/Web/HTTP/CORS)
- [Prometheus监控](https://prometheus.io/docs/)

---

**��� 恭喜！您已经掌握了 ThinkGin 2.0 的路由架构设计。现在可以开始构建您的应用程序了！**
