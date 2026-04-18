# ThinkGin

基于 [Gin](https://github.com/gin-gonic/gin) 的 Go Web 应用框架，提供开箱即用的工程化基础设施。

[![Go](https://img.shields.io/badge/Go-1.22+-00ADD8?style=flat&logo=go)](https://golang.org/)
[![Gin](https://img.shields.io/badge/Gin-v1.10-blue)](https://github.com/gin-gonic/gin)
[![License](https://img.shields.io/badge/License-MIT-green.svg)](LICENSE)

---

## 特性

- **模块化配置** — 13 个独立 YAML 文件，按职责隔离，支持环境变量覆盖
- **结构化日志** — Logrus + 文件轮转，JSON/Text 格式可选，按天切割
- **Prometheus 监控** — HTTP 四维指标 + 系统资源 + 业务自定义指标，配置化启停
- **链路追踪** — 基于 OpenTelemetry，开箱支持 stdout 导出，可扩展 OTLP/Jaeger
- **优雅停机** — 信号监听 + 超时控制，适配 Kubernetes 生命周期
- **K8s 探针** — 内置 `/livez`、`/readyz`、`/ping` 端点
- **可插拔中间件** — Recovery、RequestID、AccessLog、CORS、RateLimit、Trace、Prometheus

## 项目结构

```
thinkgin/
├── main.go                  # 入口
├── framework/
│   ├── app.go               # App 生命周期管理
│   └── options.go           # 函数式选项
├── app/
│   ├── config.go            # 配置加载（泛型 YAML 解析）
│   └── index/               # 示例业务模块 (MVC)
│       ├── controller/
│       ├── model/
│       └── view/
├── config/                  # 配置文件（13 个 YAML）
├── route/
│   ├── router.go            # 路由总入口
│   ├── web.go               # 页面路由
│   └── api.go               # API 路由（版本化）
├── extend/middleware/        # 中间件
│   ├── api_response.go      # 统一响应 + Recovery
│   ├── logger.go            # RequestID + AccessLog
│   ├── cors.go              # CORS
│   ├── ratelimit.go         # IP 令牌桶限流
│   ├── prometheus.go        # Prometheus 指标
│   └── trace.go             # OpenTelemetry 链路追踪
├── public/                  # 静态资源
└── runtime/                 # 运行时（日志输出）
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

## 测试

```bash
go test ./...
```

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
| [lestrrat-go/file-rotatelogs](https://github.com/lestrrat-go/file-rotatelogs) | 日志轮转 |
| [gopkg.in/yaml.v3](https://gopkg.in/yaml.v3) | YAML 解析 |

## 致谢

感谢 [Gin](https://github.com/gin-gonic/gin) 提供高性能的路由引擎。

## 许可证

[MIT](LICENSE)
