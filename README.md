# ThinkGin2.0

![img.png](img.png)

#### 介绍

实现一个 Go 语言基于 gin 的增删改查基础框架

#### 软件架构

```azure
thinkgin  应用部署目录
├─app                应用目录（可设置）
├  ├─index            默认模块
├  ├ ├─controller      控制器
├  ├ ├─model           模型
├  ├ └─view            视图
├  └─config.go        配置管理系统
├─config             配置目录
├  ├─app.yaml         应用配置
├  ├─server.yaml      服务器配置
├  ├─database.yaml    数据库配置
├  ├─cache.yaml       缓存配置
├  ├─log.yaml         日志配置
├  ├─session.yaml     Session配置
├  ├─middleware.yaml  中间件配置
├  ├─route.yaml       路由配置
├  ├─view.yaml        视图配置
├  ├─filesystem.yaml  文件系统配置
├  ├─lang.yaml        多语言配置
├  ├─trace.yaml       链路追踪配置
├  └─prometheus.yaml  Prometheus监控配置
├─extend             扩展目录
├─public             公共目录
├─route              路由
└─runtime            运行日志
```

#### 快速开始

**环境配置 (Go 1.13+)**

```bash
# 推荐方式
go env -w GO111MODULE=on
go env -w GOPROXY=https://goproxy.cn,direct
```

**macOS/Linux**

```bash
export GO111MODULE=on
export GOPROXY=https://goproxy.cn
```

**Windows**

```cmd
set GO111MODULE=on
set GOPROXY=https://goproxy.cn
```

**运行项目**

```bash
go mod init thinkgin
go mod tidy
go run main.go
```

访问：http://localhost:8000/

#### 📁 配置管理系统

ThinkGin2.0 采用**模块化配置管理**，将各个功能模块的配置分离到独立的 YAML 文件中，便于管理和维护。

##### 🏗️ 配置架构

```
config/
├── app.yaml          # 应用基础配置
├── server.yaml       # 服务器配置
├── database.yaml     # 数据库配置
├── cache.yaml        # 缓存配置
├── log.yaml          # 日志配置
├── session.yaml      # 会话配置
├── middleware.yaml   # 中间件配置
├── route.yaml        # 路由配置
├── view.yaml         # 视图配置
├── filesystem.yaml   # 文件系统配置
├── lang.yaml         # 多语言配置
├── trace.yaml        # 链路追踪配置
└── prometheus.yaml   # Prometheus监控配置
```

##### ⚙️ 使用方法

**1. 获取配置实例**

```go
// 获取全局配置
config := app.GetConfig()

// 获取特定模块配置
appConfig := app.GetAppConfig()
serverConfig := app.GetServerConfig()
dbConfig := app.GetDatabaseConfig()
```

**2. 访问配置项**

```go
// 应用配置
appName := config.App.Name
version := config.App.Version
jwtSecret := config.App.JWT.Secret

// 服务器配置
port := config.Server.HTTP.Port
host := config.Server.HTTP.Host

// 数据库配置
defaultDB := config.Database.Default
```

**3. 配置示例**

`config/app.yaml`:

```yaml
app:
  name: "ThinkGin"
  version: "2.0.1"
  debug: true
  timezone: "Asia/Shanghai"
  jwt:
    secret: "your-secret-key"
    expire: 7200
```

`config/server.yaml`:

```yaml
server:
  http:
    host: "0.0.0.0"
    port: 8000
    read_timeout: 60
    write_timeout: 60
  mode: "debug" # debug, test, release
```

#### 📝 日志管理

ThinkGin2.0 集成了高性能的 **Logrus** 日志管理器，这是目前 GitHub 上 Star 最多的 Go 语言日志库。

##### 日志特性

- 🚀 **高性能**: 基于 Logrus (GitHub 25.3k+ stars) 的高性能日志库
- 📊 **结构化日志**: 支持 JSON 和 Text 两种格式
- 🔄 **自动轮转**: 支持按时间和大小自动切割日志文件
- 📝 **多级别**: 支持 trace, debug, info, warn, error, fatal, panic 7 个级别
- ⚙️ **配置化**: 所有日志设置都可通过配置文件调整
- 🎯 **中间件**: 自动记录所有 HTTP 请求日志
- 💼 **业务日志**: 便捷的业务日志记录接口

##### 配置说明

在 `config/log.yaml` 文件中可以配置日志相关参数：

```yaml
log:
  # 默认日志配置
  default:
    driver: "file" # file, console, syslog
    level: "info" # trace, debug, info, warn, error, fatal, panic
    format: "json" # json, text

  # 文件日志配置
  file:
    path: "runtime/log"
    filename: "system"
    max_age: 7 # 保存天数
    rotation_time: 24 # 切割时间间隔（小时）
    max_size: 100 # 单个文件最大大小（MB）
    compress: true # 是否压缩旧文件

  # 特定类型日志配置
  channels:
    # 访问日志
    access:
      driver: "file"
      level: "info"
      filename: "access"

    # 错误日志
    error:
      driver: "file"
      level: "error"
      filename: "error"

    # SQL 日志
    sql:
      driver: "file"
      level: "debug"
      filename: "sql"
      enabled: false # 是否启用 SQL 日志
```

##### 使用方法

**1. HTTP 请求自动日志**

框架会自动记录所有 HTTP 请求的详细信息，包括：

- 请求方法和路径
- 状态码和响应时间
- 客户端 IP 和 User-Agent
- 时间戳

**2. 业务日志记录**

在控制器或其他业务逻辑中使用：

```go
import "thinkgin/extend/middleware"

// 记录信息日志
middleware.BusinessLogger("info", "用户登录成功", map[string]interface{}{
    "user_id": 123,
    "username": "john",
    "ip": "192.168.1.1",
})

// 记录错误日志
middleware.BusinessLogger("error", "数据库连接失败", map[string]interface{}{
    "error": err.Error(),
    "database": "mysql",
})

// 记录调试日志
middleware.BusinessLogger("debug", "处理业务逻辑", map[string]interface{}{
    "step": "validation",
    "data": requestData,
})
```

**3. 获取日志实例**

如需更复杂的日志操作，可直接获取 logrus 实例：

```go
import "thinkgin/app"

logger := app.GetLogger()
logger.WithFields(logrus.Fields{
    "user_id": 123,
    "action": "update_profile",
}).Info("用户更新资料")
```

##### 日志文件

- 日志文件位置：`runtime/log/`
- 文件命名：`system.YYYYMMDD.log`
- 当前日志软链：`system.log`
- 自动清理：超过设定天数的旧日志会自动删除

##### 日志级别说明

| 级别  | 说明             | 使用场景         |
| ----- | ---------------- | ---------------- |
| trace | 最详细的跟踪信息 | 调试复杂问题时   |
| debug | 调试信息         | 开发调试         |
| info  | 一般信息         | 业务流程记录     |
| warn  | 警告信息         | 潜在问题提醒     |
| error | 错误信息         | 错误处理         |
| fatal | 致命错误         | 程序无法继续运行 |
| panic | 恐慌级错误       | 触发 panic       |

##### 示例配置

**开发环境配置** (`config/log.yaml`)

```yaml
log:
  default:
    level: "debug"
    format: "text"
  file:
    max_age: 3
    rotation_time: 6
```

**生产环境配置** (`config/log.yaml`)

```yaml
log:
  default:
    level: "info"
    format: "json"
  file:
    max_age: 30
    rotation_time: 24
```

#### 📊 Prometheus 监控

ThinkGin2.0 集成了 **Prometheus** 监控服务，提供完整的应用性能监控和业务指标采集功能。

> 📖 **详细使用指南**: [点击查看完整的 Prometheus 监控使用教程](./PROMETHEUS_GUIDE.md)
>
> 该指南包含：从零开始的配置说明、详细的使用步骤、常见问题解决方案、最佳实践等完整内容。

##### 🎯 监控特性

- 🚀 **HTTP 请求监控**: 自动记录请求数量、响应时间、请求大小、响应大小
- 💾 **系统资源监控**: 实时监控 CPU 使用率、内存使用量、进程信息
- 🔧 **业务指标支持**: 提供计数器、直方图、仪表盘三种业务指标类型
- ⚙️ **配置化管理**: 所有监控参数都可通过配置文件调整
- 🔐 **安全认证**: 支持基础认证保护监控端点
- 🏷️ **自定义标签**: 支持自定义标签和命名空间

##### ⚡ 快速开始

1. **配置总开关**（`config/app.yaml`）

   ```yaml
   app:
     monitoring:
       prometheus_enabled: true # 开启监控
   ```

2. **启动应用**

   ```bash
   go run main.go
   ```

3. **查看监控数据**
   ```bash
   curl http://localhost:8000/metrics
   ```

##### ⚙️ 配置说明

**总开关配置**

首先在 `config/app.yaml` 文件中控制监控功能的总开关：

```yaml
app:
  # 监控配置
  monitoring:
    # 普罗米修斯监控总开关
    prometheus_enabled: true
```

**详细配置**

然后在 `config/prometheus.yaml` 文件中配置具体的监控参数：

```yaml
prometheus:
  # 是否启用 Prometheus 监控
  enabled: true

  # 监控指标暴露路径
  path: "/metrics"

  # 监控端口（0表示使用主服务端口）
  port: 0

  # 监控服务名称
  service_name: "thinkgin"

  # 监控命名空间
  namespace: "app"

  # 自定义标签
  labels:
    environment: "development"
    version: "2.0"

  # 监控指标配置
  metrics:
    # HTTP请求相关指标
    http:
      enabled: true
      requests_total: "http_requests_total"
      request_duration: "http_request_duration_seconds"
      request_size: "http_request_size_bytes"
      response_size: "http_response_size_bytes"
      include_path: false # 是否包含URL路径标签

    # 系统资源指标
    system:
      enabled: true
      cpu_usage: "system_cpu_usage_percent"
      memory_usage: "system_memory_usage_bytes"
      process_start_time: "process_start_time_seconds"

    # 业务指标
    business:
      enabled: true
      counter_prefix: "business_counter"
      histogram_prefix: "business_histogram"
      gauge_prefix: "business_gauge"

  # 监控数据采集间隔（秒）
  scrape_interval: 15

  # 是否启用基础认证
  auth:
    enabled: false
    username: "prometheus"
    password: "secure_password"
```

**配置优先级说明**

监控功能需要同时满足以下条件才会启用：

1. `app.yaml` 中的 `monitoring.prometheus_enabled` 必须为 `true`（总开关）
2. `prometheus.yaml` 中的 `prometheus.enabled` 必须为 `true`（具体功能开关）

这种设计允许您在不删除具体配置的情况下，通过总开关快速禁用整个监控功能。

##### 🔧 基础使用

**1. 访问监控端点**

启动应用后，访问监控指标端点：

```bash
curl http://localhost:8000/metrics
```

**2. 在业务代码中使用监控**

```go
package controller

import (
    "time"
    "thinkgin/extend/middleware"
    "github.com/gin-gonic/gin"
)

func YourHandler(c *gin.Context) {
    start := time.Now()

    // 记录 API 调用次数
    if counter := middleware.BusinessCounter("api_calls", []string{"endpoint", "status"}); counter != nil {
        counter.WithLabelValues("/api/users", "success").Inc()
    }

    // 记录 API 响应时间
    if histogram := middleware.BusinessHistogram("api_duration", []string{"endpoint"}, nil); histogram != nil {
        histogram.WithLabelValues("/api/users").Observe(time.Since(start).Seconds())
    }

    // 记录在线用户数
    if gauge := middleware.BusinessGauge("online_users", []string{"type"}); gauge != nil {
        gauge.WithLabelValues("active").Set(123)
    }

    // ... 业务逻辑
}
```

**3. 可用的业务指标类型**

- **Counter (计数器)**: 只能递增的累计指标，如请求总数、错误总数
- **Histogram (直方图)**: 观察值的分布，如响应时间、请求大小
- **Gauge (仪表盘)**: 可增可减的即时值，如当前连接数、内存使用量

##### 📈 Prometheus 服务器配置

创建 `prometheus.yml` 配置文件：

```yaml
global:
  scrape_interval: 15s

scrape_configs:
  - job_name: "thinkgin"
    static_configs:
      - targets: ["localhost:8000"]
    metrics_path: "/metrics"
    scrape_interval: 15s
```

##### 🚀 启动 Prometheus

```bash
# 下载 Prometheus
wget https://github.com/prometheus/prometheus/releases/download/v2.45.0/prometheus-2.45.0.linux-amd64.tar.gz
tar xvfz prometheus-*.tar.gz
cd prometheus-*

# 启动 Prometheus
./prometheus --config.file=prometheus.yml
```

访问 Prometheus Web UI: http://localhost:9090

##### 📊 常用监控查询

```promql
# HTTP 请求总数
app_http_requests_total

# 平均响应时间
rate(app_http_request_duration_seconds_sum[5m]) / rate(app_http_request_duration_seconds_count[5m])

# 错误率 (5xx 错误)
rate(app_http_requests_total{status=~"5.."}[5m]) / rate(app_http_requests_total[5m])

# 内存使用量
app_system_memory_usage_bytes

# 业务 API 调用频率
rate(app_business_counter_api_calls[5m])
```

##### 🔐 安全配置

**启用认证：**

```yaml
prometheus:
  auth:
    enabled: true
    username: "monitor"
    password: "your_secure_password"
```

**访问受保护的端点：**

```bash
curl -u monitor:your_secure_password http://localhost:8000/metrics
```

##### 📚 集成 Grafana

1. 安装 Grafana
2. 添加 Prometheus 数据源：`http://localhost:9090`
3. 导入预置仪表板或创建自定义仪表板
4. 创建告警规则监控关键指标

通过这套完整的监控体系，您可以实时了解应用的运行状态和性能表现。

#### 🗄️ 其他配置模块

- **🗄️ 数据库配置** (`config/database.yaml`): 支持 MySQL、PostgreSQL、SQLite、Redis
- **🚀 缓存配置** (`config/cache.yaml`): 支持 Redis、内存、文件缓存
- **🔐 会话配置** (`config/session.yaml`): 支持文件、Redis、数据库存储
- **🛡️ 中间件配置** (`config/middleware.yaml`): 全局中间件和路由组中间件
- **🎨 视图配置** (`config/view.yaml`): 模板引擎和静态资源配置
- **📁 文件系统配置** (`config/filesystem.yaml`): 本地存储和云存储(阿里云 OSS、腾讯云 COS、七牛云)
- **🌐 多语言配置** (`config/lang.yaml`): 国际化支持
- **🔍 链路追踪配置** (`config/trace.yaml`): Jaeger、Zipkin、OpenTelemetry 支持

通过这个完整的模块化配置系统，你可以轻松管理应用程序的各个方面，实现配置的分离和模块化管理。

#### 常见问题

- goland 导入包爆红的问题的解决方案（成功解决 Nice）：
  `GOPROXY=https://goproxy.cn,direct`
  ![img_1.png](https://gitee.com/goubiwanyi/thinkgin/raw/master/img/img_1.png)

#### 路由规范

###### 路由编写文件：`route/router.go`

###### 访问示例：`http://thinkgin.cn:8000/index/hello`

```azure
{
    "code": 200,
    "data": {
        "Time": "2023-02-01T10:11:41.76127929+08:00",
        "data": "Hello ThinkGin!",
        "name": ""
    },
    "msg": "ok"
}
```

#### 📦 编译与部署

##### Windows 编译

```bash
# 设置环境变量
go env -w GO111MODULE=on
go env -w GOPROXY=https://goproxy.cn,direct

# 编译 Windows 可执行文件
go build -o thinkgin.exe

# 直接运行
./thinkgin.exe
```

##### Linux 编译

```bash
# 设置环境变量
go env -w GO111MODULE=on
go env -w GOPROXY=https://goproxy.cn,direct

# 编译 Linux 可执行文件
go build -o thinkgin

# 直接运行
./thinkgin

# 或后台执行
nohup ./thinkgin 1>info.log 2>&1 &
```

##### 跨平台编译

```bash
# 编译 Linux 版本 (在 Windows 上)
SET GOOS=linux
SET GOARCH=amd64
go build -o thinkgin-linux

# 编译 macOS 版本 (在 Windows 上)
SET GOOS=darwin
SET GOARCH=amd64
go build -o thinkgin-macos

# 编译 Windows 版本 (在 Linux/macOS 上)
GOOS=windows GOARCH=amd64 go build -o thinkgin.exe
```

##### Docker 部署

创建 `Dockerfile`:

```dockerfile
FROM golang:1.19-alpine AS builder

WORKDIR /app
COPY . .
RUN go mod tidy && go build -o thinkgin

FROM alpine:latest
RUN apk --no-cache add ca-certificates
WORKDIR /root/
COPY --from=builder /app/thinkgin .
COPY --from=builder /app/config ./config

CMD ["./thinkgin"]
```

构建和运行：

```bash
# 构建镜像
docker build -t thinkgin:v2.0 .

# 运行容器
docker run -p 8000:8000 thinkgin:v2.0
```

##### 📝 重要说明

> **为什么删除了 thinkgin.exe？**
>
> 1. **版本控制最佳实践**: 不提交编译后的二进制文件到 Git 仓库
> 2. **仓库体积控制**: 避免仓库变得臃肿（单个可执行文件约 20MB）
> 3. **跨平台兼容性**: 不同操作系统需要不同的可执行文件
> 4. **安全考虑**: 避免潜在的安全风险
>
> **如何重新生成可执行文件？**
>
> 只需运行对应平台的编译命令即可重新生成，编译过程通常只需几秒钟。
