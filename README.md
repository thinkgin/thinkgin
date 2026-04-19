# ThinkGin

基于 [Gin](https://github.com/gin-gonic/gin) 的 Go Web 应用框架，提供开箱即用的工程化基础设施。

[![Go](https://img.shields.io/badge/Go-1.22+-00ADD8?style=flat&logo=go)](https://golang.org/)
[![Gin](https://img.shields.io/badge/Gin-v1.10-blue)](https://github.com/gin-gonic/gin)
[![License](https://img.shields.io/badge/License-MIT-green.svg)](LICENSE)

---

## 特性

- **模块化配置** — 13 个独立 YAML 文件，按职责隔离；`THINKGIN_*` 前缀环境变量覆盖
- **多数据源** — GORM 接入 MySQL/PostgreSQL/SQLite（SQLite 纯 Go，无 cgo）
- **Redis 缓存** — `go-redis/v9` 多实例管理，启动期 Ping 探活
- **结构化日志** — Logrus + 按天文件轮转，JSON/Text 格式可选
- **Prometheus 监控** — HTTP 四维 + 运行时 + 业务自定义指标，配置化启停
- **链路追踪** — OpenTelemetry TracerProvider；GORM 插件把 SQL 纳入同一条 trace
- **优雅停机** — 信号监听 + 超时 + DB/Cache/Tracer 依序释放，适配 Kubernetes
- **K8s 探针** — 内置 `/livez`、`/readyz`、`/ping` 端点
- **可插拔中间件** — Recovery、RequestID、AccessLog、CORS、RateLimit、Trace、Prometheus

## 项目结构

```
thinkgin/
├── main.go                       # 入口：Bootstrap → DB/Cache/Tracer → Run
├── framework/                    # App 生命周期
│   ├── app.go
│   ├── options.go
│   └── browser.go
├── app/
│   ├── bootstrap.go              # Bootstrap(configDir) 显式入口
│   ├── loader.go                 # 13 个 YAML 的泛型加载器
│   ├── env.go                    # THINKGIN_* 环境变量覆盖
│   ├── defaults.go               # 零值默认填充
│   ├── validate.go               # 语义校验与纠偏
│   ├── config.go / types.go      # 全局 Config 单例 + 强类型
│   ├── logger.go                 # Logrus 初始化 + 跨平台文件轮转
│   ├── database/                 # GORM 多数据源 + OTel 插件
│   ├── cache/                    # Redis 多实例管理
│   └── index/                    # 示例业务模块 (MVC)
│       ├── controller/
│       ├── model/                # 含 User 示例
│       └── view/
├── config/                       # 13 个 YAML
├── route/                        # 路由按 web / api 拆分
├── extend/middleware/
│   ├── api_response.go           # 统一响应 + Recovery
│   ├── logger.go                 # RequestID + AccessLog
│   ├── cors.go / ratelimit.go
│   ├── trace.go                  # OTel TracerProvider 装配
│   └── prometheus*.go            # 指标按 http/system/business/handler 拆分
├── .github/workflows/ci.yml      # 三平台矩阵 + race + coverage
├── .golangci.yml                 # 精挑 linter
├── public/                       # 静态资源
└── runtime/                      # 运行时产物（日志）
```

## 快速开始

### 环境要求

- Go 1.22+
- Git

### 安装与运行

```bash
# 克隆
git clone https://gitee.com/libaicode/thinkgin.git
cd thinkgin

# 配置 Go 代理（国内用户）
go env -w GOPROXY=https://goproxy.cn,direct

# 下载依赖
go mod tidy

# 运行
go run main.go
```

启动后访问：

| 地址 | 说明 |
|------|------|
| http://localhost:8000 | 首页 |
| http://localhost:8000/index/hello | Hello World API |
| http://localhost:8000/metrics | Prometheus 指标 |
| http://localhost:8000/livez | 存活探针 |

### 编译

```bash
# 本平台
go build -o thinkgin main.go

# 交叉编译
GOOS=linux GOARCH=amd64 go build -o thinkgin-linux main.go
GOOS=darwin GOARCH=arm64 go build -o thinkgin-darwin main.go
```

## 配置

所有配置位于 `config/` 目录，每个文件独立管理一个模块：

| 文件 | 说明 |
|------|------|
| `app.yaml` | 应用名、版本、JWT、分页 |
| `server.yaml` | 监听地址、端口、超时、运行模式 |
| `database.yaml` | MySQL/PostgreSQL/SQLite 连接 |
| `cache.yaml` | Redis/内存缓存 |
| `log.yaml` | 日志级别、格式、轮转策略 |
| `middleware.yaml` | 全局中间件启用列表 |
| `prometheus.yaml` | 监控指标、认证、采集间隔 |
| `trace.yaml` | 链路追踪开关、采样率、导出方式 |
| `session.yaml` | 会话存储 |
| `view.yaml` | 模板引擎 |
| `filesystem.yaml` | 文件存储（本地/OSS） |
| `lang.yaml` | 国际化 |
| `route.yaml` | 路由配置 |

### 环境变量覆盖

所有核心配置支持 `THINKGIN_` 前缀的环境变量覆盖：

```bash
THINKGIN_SERVER_MODE=release
THINKGIN_SERVER_HTTP_PORT=9090
THINKGIN_APP_DEBUG=false
THINKGIN_LOG_LEVEL=warn
```

## 中间件

通过 `config/middleware.yaml` 配置启用：

```yaml
middleware:
  global:
    - recovery
    - request_id
    - trace
    - logger
    - prometheus
    - cors
    - rate_limit
```

### 业务监控示例

```go
// 计数器
if c := middleware.BusinessCounter("api_calls", []string{"endpoint"}); c != nil {
    c.WithLabelValues("/api/users").Inc()
}

// 直方图
if h := middleware.BusinessHistogram("api_duration", []string{"endpoint"}, nil); h != nil {
    h.WithLabelValues("/api/users").Observe(elapsed.Seconds())
}
```

## 数据库与缓存

```go
import (
    "thinkgin/app/cache"
    "thinkgin/app/database"
)

// 默认连接（对应 database.yaml 的 default 字段）
db := database.Default()

// 或按名取
pg, _ := database.Get("pgsql")

// Redis
rdb := cache.Default()
```

GORM 插件会自动为每条 SQL 开启 span，并以当前 HTTP 请求的 trace 为父节点——
前提是业务代码用 `db.WithContext(c.Request.Context())` 传递 context：

```go
func ListUsers(c *gin.Context) {
    var users []model.User
    database.Default().WithContext(c.Request.Context()).Find(&users)
    c.JSON(200, users)
}
```

## 测试 & CI

```bash
go test ./...
# 启用竞态检测（需要 cgo）
CGO_ENABLED=1 go test -race ./...
```

仓库根目录的 `.github/workflows/ci.yml` 会在 push / PR 时跑：

- **Build & Test**：Ubuntu / Windows / macOS 三矩阵，启用 `-race` + coverage
- **Lint**：`golangci-lint`，规则集见 `.golangci.yml`

## 部署

### Docker

```dockerfile
FROM golang:1.22-alpine AS builder
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
docker build -t thinkgin:3.0 .
docker run -p 8000:8000 thinkgin:3.0
```

### Kubernetes

框架内置 `/livez` 和 `/readyz` 探针，可直接配置：

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
| [gin-gonic/gin](https://github.com/gin-gonic/gin) v1.10 | HTTP 路由 |
| [sirupsen/logrus](https://github.com/sirupsen/logrus) v1.9 | 结构化日志 |
| [prometheus/client_golang](https://github.com/prometheus/client_golang) v1.20 | 监控指标 |
| [go.opentelemetry.io/otel](https://opentelemetry.io/) v1.28 | 链路追踪 |
| [gorm.io/gorm](https://gorm.io/) v1.31 | ORM |
| [redis/go-redis/v9](https://github.com/redis/go-redis) v9.18 | Redis 客户端 |
| [glebarez/sqlite](https://github.com/glebarez/sqlite) | SQLite 纯 Go 驱动 |
| [lestrrat-go/file-rotatelogs](https://github.com/lestrrat-go/file-rotatelogs) | 日志轮转 |
| [gopkg.in/yaml.v3](https://gopkg.in/yaml.v3) | YAML 解析 |

## 从 ThinkPHP 迁移？

如果你之前用 ThinkPHP，建议阅读 [`docs/migration-from-thinkphp.md`](docs/migration-from-thinkphp.md)，
里面列出了最常见的 20+ 对照项（Model、Route、Middleware、Config、View 等）。

## 致谢

感谢 [Gin](https://github.com/gin-gonic/gin) 提供高性能的路由引擎。

## 许可证

[MIT](LICENSE)
