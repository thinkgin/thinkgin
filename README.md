# ThinkGin

基于 Gin 的生产级 Go Web 框架——配置即秩序，中间件即能力，开箱即交付。

[![Go](https://img.shields.io/badge/Go-1.25+-00ADD8?style=flat&logo=go)](https://golang.org/)
[![Gin](https://img.shields.io/badge/Gin-v1.12-blue)](https://github.com/gin-gonic/gin)
[![Version](https://img.shields.io/badge/Version-3.10.3-orange)](CHANGELOG.md)
[![License](https://img.shields.io/badge/License-MIT-green.svg)](LICENSE)

> **当前版本 3.10.3** — 内置 18 个可插拔中间件、ServiceContext 依赖注入、WebSocket、按路由粒度熔断器、atomic 配置热更新、slog 标准日志引擎等。  
> 完整更新日志见 [CHANGELOG.md](CHANGELOG.md)。

---

## 定位声明（先看这个，再决定是否选用）

ThinkGin 的目标是 **「单体 Go Web 业务后台 / 中后台脚手架」**，强调 **工程纪律 · 配置秩序 · 开箱即交付**。

**适合场景** ✅

- 中小型业务后台、SaaS 控制台、内部系统、API 服务
- 团队希望从 Gin 裸跑直接升级到「带 18 中间件、配置热更新、Prometheus、OTel 的生产级骨架」
- 习惯 ThinkPHP / Laravel 风格的 YAML 多文件配置
- 优先 **可读性 / 可维护性 / 工程质量**，性能能跑满 Gin 即可（数千~数万 QPS）

**不适合场景** ❌

- 需要 18 万+ QPS 裸路由的高性能网关 → 请选 [Fiber](https://github.com/gofiber/fiber) / [Hertz](https://github.com/cloudwego/hertz)
- 需要 gRPC / 多协议 / 服务发现 / 配置中心的云原生微服务 → 请选 [Kratos](https://github.com/go-kratos/kratos) / [go-zero](https://github.com/zeromicro/go-zero)
- 追求极致零分配 / 内存优化的底层组件

ThinkGin **不会** 与上述定位竞争。我们专注把单体业务后台的「最后一公里」做好。

---

## 特性一览

### 核心基础设施

- **模块化配置** — 13 个独立 YAML 文件按职责隔离；`THINKGIN_*` 前缀环境变量覆盖（含 JWT Secret）；`fsnotify` 配置热更新（atomic.Pointer 原子替换，并发安全）
- **ServiceContext** — 依赖注入替代全局单例，Config / Logger / DB / Cache 聚合到显式结构体
- **多数据源** — GORM 接入 MySQL / PostgreSQL / SQLite（纯 Go，无 cgo）
- **Redis 缓存** — `go-redis/v9` 多实例管理，启动期 Ping 探活
- **结构化日志** — Logrus + lumberjack 按大小轮转（自动压缩），JSON / Text 格式可选
- **Prometheus 监控** — HTTP 四维 + 运行时 + 业务自定义指标，配置化启停
- **链路追踪** — OpenTelemetry TracerProvider；GORM 插件把 SQL 纳入同一条 trace
- **优雅停机** — 信号监听 + 超时 + DB / Cache / Tracer 依序释放，适配 Kubernetes
- **K8s 探针** — `/livez` 验证进程存活；`/readyz` 真实探测 DB / Redis
- **Swagger / OpenAPI** — 内置 `/swagger/` 端点，配置 `swagger_enabled: true` 一键启用

### 18 个可插拔中间件

通过 `config/middleware.yaml` 按名启用，零代码修改：

| 中间件 | 名称 | 说明 |
|--------|------|------|
| Recovery | `recovery` | panic 兜底，API 路径返回 JSON |
| RequestID | `request_id` | 生成/透传 `X-Request-ID` |
| AccessLog | `access_log` | 结构化访问日志，可配置跳过路径 |
| CORS | `cors` | 动态 Origin 匹配，完整预检处理 |
| RateLimit | `rate_limit` | 内存令牌桶，按 IP 限流 |
| Redis 限流 | `redis_rate_limit` | 分布式限流，基于 Redis 滑动窗口 |
| Session | `session` | memory / redis Store，Cookie 安全标志自动透传 |
| JWT | `jwt` | HS256 签发 / 校验，拒绝 `alg=none` 伪造 |
| CSRF | `csrf` | 双重提交 Cookie 防护 |
| Gzip | `gzip` | 响应压缩 + 请求解压 |
| SecureHeaders | `secure_headers` | HSTS / X-Frame-Options / CSP 等安全头 |
| Trace | `trace` | OpenTelemetry 链路追踪 |
| Prometheus | `prometheus` | 请求指标自动采集 |
| **CircuitBreaker** | `circuit_breaker` | 滑动窗口熔断器，三态状态机，**按路由粒度隔离** |
| **Timeout** | `timeout` | 请求超时控制，`context.WithTimeout` |
| **BodyLimit** | `body_limit` | 请求体大小限制，默认 10MB |

### 更多能力

- **WebSocket** — 基于 gorilla/websocket，`Handler()` 一行升级；Hub 广播模型
- **i18n 国际化** — 运行时多语言切换，Accept-Language 自动协商
- **脚手架 CLI** — `go run ./cmd/scaffold new module user` 一键生成 MVC 骨架
- **数据库迁移** — 推荐 `pressly/goose`，见 [`docs/database-migration.md`](docs/database-migration.md)

## 项目结构

```
thinkgin/
├── main.go                       # 入口：Bootstrap → DB/Cache/Tracer → Run
├── framework/                    # App 生命周期（启动/停机/浏览器打开）
├── app/
│   ├── bootstrap.go              # Bootstrap(configDir) 显式入口
│   ├── loader.go                 # 13 个 YAML 的泛型加载器
│   ├── env.go                    # THINKGIN_* 环境变量覆盖
│   ├── defaults.go               # 零值默认填充
│   ├── service_context.go        # ServiceContext 依赖注入
│   ├── config.go / types.go      # 全局 Config（atomic.Pointer 线程安全）
│   ├── version.go                # 版本号唯一来源，支持 ldflags 注入
│   ├── logger.go                 # Logrus + lumberjack 日志轮转
│   ├── database/                 # GORM 多数据源 + OTel 插件
│   ├── cache/                    # Redis 多实例管理
│   ├── session/                  # memory / redis Session Store
│   ├── ctxkeys/                  # Context 类型安全 key
│   └── index/                    # 示例业务模块 (MVC)
├── config/                       # 13 个 YAML 配置文件
├── route/                        # Web / API 路由 + 健康探针 + Swagger
├── extend/
│   ├── middleware/               # 18 个可插拔中间件
│   ├── websocket/                # WebSocket 连接 + Hub 广播
│   └── i18n/                     # 国际化运行时
├── cmd/scaffold/                 # 脚手架 CLI
├── docs/                         # OpenAPI / FAQ / 迁移指南
├── .github/workflows/ci.yml      # 三平台矩阵 + race + coverage
└── .golangci.yml                 # Lint 规则集
```

## 快速开始

### 环境要求

- **Go 1.25+**（Gin 1.12 通过 quic-go 间接要求）
- Git

### 安装与运行

```bash
git clone https://gitee.com/libaicode/thinkgin.git
cd thinkgin

# 国内用户建议配置代理
go env -w GOPROXY=https://goproxy.cn,direct

go mod tidy
go run main.go
```

启动后访问：

| 地址 | 说明 |
|------|------|
| http://localhost:8000 | 首页 |
| http://localhost:8000/ping | 健康检查（返回版本号） |
| http://localhost:8000/metrics | Prometheus 指标 |
| http://localhost:8000/livez | 存活探针 |
| http://localhost:8000/readyz | 就绪探针 |
| http://localhost:8000/swagger/ | Swagger UI（需 `swagger_enabled: true`） |

### 编译

```bash
go build -o thinkgin main.go

# 通过 ldflags 注入版本号（CI/CD 推荐）
go build -ldflags "-X thinkgin/app.Version=3.10.1" -o thinkgin main.go

# 交叉编译
GOOS=linux GOARCH=amd64 go build -o thinkgin-linux main.go
GOOS=darwin GOARCH=arm64 go build -o thinkgin-darwin main.go
```

## 配置

所有配置位于 `config/` 目录：

| 文件 | 说明 |
|------|------|
| `app.yaml` | 应用名、版本、JWT、分页、Swagger 开关 |
| `server.yaml` | 监听地址、端口、超时、HTTPS、运行模式 |
| `database.yaml` | MySQL / PostgreSQL / SQLite / Redis 连接 |
| `cache.yaml` | Redis / 内存缓存 |
| `log.yaml` | 日志级别、格式、轮转策略 |
| `middleware.yaml` | 全局中间件启用列表 + 中间件参数 |
| `prometheus.yaml` | 监控指标、认证、采集间隔 |
| `trace.yaml` | 链路追踪开关、采样率、导出方式 |
| `session.yaml` | 会话存储（memory / redis） |
| `view.yaml` | 模板引擎（预留） |
| `filesystem.yaml` | 文件存储（预留） |
| `lang.yaml` | 国际化语言包 |
| `route.yaml` | 路由配置 |

### 环境变量覆盖

```bash
THINKGIN_SERVER_MODE=release
THINKGIN_SERVER_HTTP_PORT=9090
THINKGIN_APP_DEBUG=false
THINKGIN_LOG_LEVEL=warn
```

### 配置热更新

基于 `fsnotify` 监听文件变化，修改 YAML 后自动重载，无需重启服务。

## 中间件

### 基础配置

```yaml
# config/middleware.yaml
middleware:
  global:
    - recovery
    - request_id
    - access_log
    - cors
    - rate_limit
    - circuit_breaker
    - timeout
    - body_limit
```

### 熔断器

基于滑动窗口的 Circuit Breaker，三态自动切换：Closed → Open（503）→ HalfOpen → Closed。  
v3.7.2 起支持**按路由粒度隔离**——每条路由拥有独立的滑动窗口和状态机，避免单个慢接口拖垮全局。

```go
// 默认：窗口 100 / 50% 错误率 / 10s 冷却 / 5 次试探
r.Use(middleware.CircuitBreaker())

// 自定义
r.Use(middleware.CircuitBreakerWithConfig(middleware.CircuitBreakerConfig{
    WindowSize:            200,
    ErrorThresholdPercent: 30,
    CooldownDuration:      5 * time.Second,
    HalfOpenMaxRequests:   3,
}))
```

### 超时控制

为每个请求注入 `context.WithTimeout`，超时返回 504 Gateway Timeout。

```go
r.Use(middleware.Timeout())                          // 默认 30s
r.Use(middleware.TimeoutWithDuration(5 * time.Second)) // 自定义
```

Handler 应使用 `c.Request.Context()` 感知超时信号。

### 请求体限制

Content-Length 预检 + `MaxBytesReader` 双重保护，超限返回 413。

```go
r.Use(middleware.BodyLimit())                       // 默认 10MB
r.Use(middleware.BodyLimitWithSize(5 << 20))         // 5MB
h, _ := middleware.BodyLimitFromString("512KB")      // 字符串解析
r.Use(h)
```

### 安全中间件

```yaml
middleware:
  global:
    - csrf            # 双重提交 Cookie
    - secure_headers  # HSTS / X-Frame-Options / CSP
    - gzip            # 响应压缩
```

### 业务监控

```go
if c := middleware.BusinessCounter("api_calls", []string{"endpoint"}); c != nil {
    c.WithLabelValues("/api/users").Inc()
}
if h := middleware.BusinessHistogram("api_duration", []string{"endpoint"}, nil); h != nil {
    h.WithLabelValues("/api/users").Observe(elapsed.Seconds())
}
```

## ServiceContext

v3.3.0 引入 `ServiceContext` 替代全局单例：

```go
svc := app.SvcFromGin(c)
svc.Log.Infof("当前应用: %s", svc.Config.App.Name)
```

保留 `app.GetConfig()` / `app.GetLogger()` 全局函数向后兼容。

## 数据库与缓存

```go
import (
    "thinkgin/app/cache"
    "thinkgin/app/database"
)

db := database.Default()
pg, _ := database.Get("pgsql")
rdb := cache.Default()

// GORM + OTel：传递 context 自动纳入 trace
database.Default().WithContext(c.Request.Context()).Find(&users)
```

## WebSocket

```go
import "thinkgin/extend/websocket"

// Echo
r.GET("/ws", websocket.Handler(func(conn *websocket.Conn) {
    defer conn.Close()
    for {
        mt, msg, err := conn.Conn.ReadMessage()
        if err != nil { break }
        conn.WriteSafeMessage(mt, msg)
    }
}))

// Hub 广播
hub := websocket.NewHub()
hub.Broadcast(ws.TextMessage, msg)
hub.BroadcastJSON(map[string]string{"event": "join"})
```

## Session

```go
import "thinkgin/app/session"

s := session.From(c)
s.Set("user_id", 42)
_ = s.Save(c.Request.Context())
```

| driver | 状态 | 适用场景 |
|--------|------|----------|
| `memory` | ✅ 可用 | 单实例 / 开发调试 |
| `redis` | ✅ 可用 | 多副本部署 |

## JWT 鉴权

```go
// 签发
token, err := middleware.JWTIssue("42", map[string]any{"role": "admin"}, 0)

// 保护路由
api := r.Group("/api/v1", middleware.JWTAuth())
api.GET("/me", func(c *gin.Context) {
    claims := middleware.JWTFromContext(c)
    c.JSON(200, gin.H{"uid": claims.Subject})
})
```

HS256 签名，严格拒绝非 HMAC 算法。

## 脚手架 CLI

```bash
go run ./cmd/scaffold new module user
# → app/user/controller/user.go
# → app/user/model/user.go
# → app/user/view/.gitkeep
```

## 测试 & CI

```bash
go test ./...
CGO_ENABLED=1 go test -race ./...
```

GitHub Actions CI（`.github/workflows/ci.yml`）：

- **三平台矩阵**：Ubuntu / Windows / macOS
- **竞态检测** + **覆盖率报告**
- **golangci-lint** 静态分析

当前覆盖率：middleware **62%**、route **37%**。

## 部署

### Docker

```dockerfile
FROM golang:1.25-alpine AS builder
WORKDIR /app
COPY . .
RUN go mod tidy && go build -o thinkgin

FROM alpine:latest
RUN apk --no-cache add ca-certificates
WORKDIR /app
COPY --from=builder /app/thinkgin .
COPY --from=builder /app/config ./config
EXPOSE 8000
CMD ["./thinkgin"]
```

```bash
docker build -t thinkgin:3.10.1 .
docker run -p 8000:8000 thinkgin:3.10.1
```

### Kubernetes

```yaml
livenessProbe:
  httpGet:
    path: /livez
    port: 8000
readinessProbe:
  httpGet:
    path: /readyz
    port: 8000
```

## 依赖

| 包 | 用途 |
|----|------|
| [gin-gonic/gin](https://github.com/gin-gonic/gin) v1.12 | HTTP 路由 |
| [sirupsen/logrus](https://github.com/sirupsen/logrus) v1.9 | 结构化日志 |
| [prometheus/client_golang](https://github.com/prometheus/client_golang) v1.20 | 监控指标 |
| [go.opentelemetry.io/otel](https://opentelemetry.io/) v1.43 | 链路追踪 |
| [gorm.io/gorm](https://gorm.io/) v1.31 | ORM |
| [redis/go-redis/v9](https://github.com/redis/go-redis) v9.18 | Redis 客户端 |
| [gorilla/websocket](https://github.com/gorilla/websocket) v1.5 | WebSocket |
| [fsnotify/fsnotify](https://github.com/fsnotify/fsnotify) v1.9 | 配置热更新 |
| [golang-jwt/jwt/v5](https://github.com/golang-jwt/jwt) v5.3 | JWT |
| [natefinch/lumberjack](https://github.com/natefinch/lumberjack) v2.2 | 日志轮转（按大小 + 按时间） |
| [glebarez/sqlite](https://github.com/glebarez/sqlite) | SQLite 纯 Go 驱动 |
| [gopkg.in/yaml.v3](https://gopkg.in/yaml.v3) | YAML 解析 |

## 版本历史

| 版本 | 亮点 |
|------|------|
| **3.10.1** | 全局状态 Deprecated 标记 / slog 替代 logrus / 超时中间件竞态修复 |
| **3.9.0** | 强类型配置 / WebSocket 安全 / Redis 限流上下文 / 测试重命名 |
| **3.8.1** | 统一版本号管理（`app/version.go` + ldflags 注入） |
| **3.8.0** | 日志轮转替换为 lumberjack，移除 archived 依赖 |
| **3.7.2** | 熔断器按路由粒度隔离 |
| **3.7.1** | LoadHTMLGlob 安全检查 |
| **3.7.0** | 配置热更新 atomic.Pointer 原子替换 |
| **3.6.3** | JWT Secret/Expire 环境变量注入 |
| **3.6.2** | CORS AllowCredentials + * 安全修复 |
| **3.6.1** | CSRF Token 时序攻击修复 |
| **3.3.4** | middleware 测试覆盖率 62%，route 37% |
| **3.3.0** | ServiceContext / CSRF / Gzip / SecureHeaders / Swagger / 热更新 / Redis 限流 / i18n / WebSocket |
| **3.2.0** | Logger 接口抽象 / HTTPS 双监听 |

详见 [CHANGELOG.md](CHANGELOG.md)。

## 相关文档

- 📖 [官方文档](https://gitee.com/libaicode/home-thinkgin)
- 📋 [从 ThinkPHP 迁移](docs/migration-from-thinkphp.md)
- 🗃️ [数据库迁移指南](docs/database-migration.md)
- ❓ [FAQ / 设计决策](docs/faq.md)
- 📝 [OpenAPI 定义](docs/openapi.yaml)

## 致谢

感谢 [Gin](https://github.com/gin-gonic/gin) 提供高性能的路由引擎。

## 许可证

[MIT](LICENSE)
