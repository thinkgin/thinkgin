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
├  └─trace.yaml       链路追踪配置
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
└── trace.yaml        # 链路追踪配置
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
      enabled: false # 是否启用SQL日志
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

#### Linux 打包

```azure
go env -w GO111MODULE=on
go env -w GOPROXY=https://goproxy.cn,direct
go build -o thinkgin.sh
```

##### 直接运行即可：

`./thinkgin.sh`

##### 或后台执行：

`nohup ./thinkgin.sh 1>info.log 2>&1 &`
