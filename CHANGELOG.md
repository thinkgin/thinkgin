# Changelog

本项目遵循 [Semantic Versioning](https://semver.org/lang/zh-CN/) 和 [Keep a Changelog](https://keepachangelog.com/zh-CN/1.1.0/) 约定。

## [3.6.3] - 2026-04-30

### Added

- **JWT 环境变量注入**：新增 `THINKGIN_APP_JWT_SECRET` 和 `THINKGIN_APP_JWT_EXPIRE` 环境变量覆盖，确保 JWT 密钥可通过环境变量安全注入，无需写入配置文件

---

## [3.6.2] - 2026-04-30

### Fixed

- **CORS AllowCredentials + * 安全修复**：当 `allow_credentials=true` 且 `allow_origins=["*"]` 时，不再返回 `Access-Control-Allow-Origin: *`（W3C 规范禁止），改为回显请求 Origin 并附加 `Vary: Origin`
- 启动时对此非法配置组合打印警告日志，提示显式列出可信域名

---

## [3.6.1] - 2026-04-30

### Fixed

- **CSRF Token 时序攻击修复**：`csrf.go` 中 token 比较从 `!=` 改为 `crypto/subtle.ConstantTimeCompare`，防止时序攻击（Timing Attack）窃取 CSRF token

---

## [3.6.0] - 2026-04-30

### Added

- **CLI 命令框架（Cobra）**：`main.go` 重构为 Cobra 入口
  - `serve` — 启动 HTTP 服务器（默认命令，向后兼容）
  - `version` — 打印版本号
  - `config` — 输出合并后的完整配置（JSON）
  - `cron` — 启动定时任务调度器
  - 无参数运行等同于 `serve`
- **Cron 定时任务调度器**：新增 `extend/cron` 包
  - 基于 robfig/cron v3，支持秒级 cron 表达式和 `@every` 语法
  - `Register(name, schedule, fn)` 注册 + `Start()` 启动 + `Stop()` 优雅停机
  - 任务 panic 自动恢复不影响其他任务
  - `cmd/cron.go` 子命令：独立进程运行调度器
- **事件系统（发布/订阅）**：新增 `extend/event` 包
  - `On(name, listener)` 订阅 + `Fire(name, payload)` 同步发布
  - `FireAsync` 异步发布，每个订阅者独立 goroutine
  - 内置生命周期事件：`AppStarting` / `AppStarted` / `AppStopping` / `AppStopped`
  - `Off` 移除 + `ListenerCount` 查询 + panic 自动恢复

### Changed

- `main.go` 从 116 行精简为 Cobra 入口（18 行），原逻辑迁入 `cmd/serve.go`

---

## [3.5.0] - 2026-04-30

### Added

- **OTel OTLP 实际导出器**：`trace.go` 现支持三种 driver
  - `stdout`（默认）：输出到 `runtime/log/trace.log`
  - `otlp` / `jaeger`：通过 OTLP gRPC(4317) 或 HTTP(4318) 发送到 Jaeger/Tempo/SigNoz
  - 根据 `trace.otel.use_grpc` 自动选择协议
- **Cache 抽象层**：新增 `Store` 接口 + `MemoryStore` / `RedisStore` 双实现
  - `Get` / `Set` / `Delete` / `Has` / `Remember` / `Flush` 统一 API
  - `NewStore(name)` 工厂根据配置 driver 自动选择实现
  - `DefaultStore()` 快捷获取默认缓存
  - `MemoryStore` 含惰性删除 + 后台 GC，done channel 可安全退出
- **Docker Compose 本地开发栈**：一键启动 MySQL + Redis + Jaeger + Prometheus
  - `docker-compose.yml` + `docker/prometheus.yml` 采集配置
  - 含健康检查、数据卷持久化

### Changed

- 升级 OTel 依赖到 v1.43.0（otel/sdk/trace/metric/contrib）
- 升级 go-playground/validator 到 v10.30.2
- 新增 `otlptracegrpc` / `otlptracehttp` / `grpc` 等依赖

---

## [3.4.0] - 2026-04-30

### Added

- **结构化错误码体系**：新增 `app/errors` 包，提供 `AppError` 类型
  - 支持 `errors.Is` / `errors.As`，Code 相同即视为同一错误
  - 预定义 14 个通用错误码（400000~504000）
  - `WithMsg` / `WithData` / `WithCause` 不可变派生
  - `APIErrorHandler` 自动识别 `*AppError` 并格式化响应
  - `APIAppError()` 函数直接输出 AppError 驱动的 JSON
- **请求参数校验层**：新增 `extend/middleware/validation.go`
  - `BindAndValidate(c, &req)` 一行完成绑定+校验
  - 校验失败自动返回 422 + 字段级中文错误消息
  - 覆盖 required/min/max/email/oneof/gt/gte/lt/lte 等常用规则
  - `resolveMiddleware` 新增 `validation` 名称映射
- **统一分页响应**：新增 `app/pagination` 包
  - `FromQuery(c)` 自动解析 page/page_size，尊重 app.yaml 配置
  - `NewResult(params, total, list)` 生成 `{list, total, page, page_size, total_pages}` 响应
  - 自动 cap 到 max_page_size，负数 page 重置为 1

---

## [3.3.5] - 2026-04-30

### Fixed

- RateLimit GC goroutine 无法停止的泄漏：新增 `done` channel，支持 `close(done)` 安全退出
- 清理仓库中残留的覆盖率产物文件（cov_mw, cov_mw.out 等）

### Added

- RateLimit 标准限流响应头：`X-RateLimit-Limit` / `X-RateLimit-Remaining` / `X-RateLimit-Reset` / `Retry-After`
- 分组中间件生效：新增 `ApplyGroupMiddleware(group, name)` 函数，从 `middleware.groups` 配置读取并批量挂载
- `resolveMiddleware()` 统一中间件名到 HandlerFunc 的映射，供全局和分组共用
- API 路由组自动应用 `middleware.groups.api` 配置
- Prometheus `/metrics` 端点 Basic Auth 配置示例（`prometheus.auth`）
- 新增 `TestStartGC_StopsOnDoneClose` / `TestRateLimit_ResponseHeaders` / `TestRateLimit_RetryAfterOnBlock` 测试
- 新增 `TestResolveMiddleware_*` / `TestApplyGroupMiddleware_*` 测试

---

## [3.3.4] - 2026-04-25

### Changed

- middleware 测试覆盖率从 43% 提升到 **62%**
- route 测试覆盖率从 17% 提升到 **37%**
- 新增 Recovery/APIErrorHandler/Gzip 解压/CORS helper/Logger 内部函数的测试
- 新增 route 包 applyMiddleware/registerGlobalMiddleware/registerMonitoringRoutes 测试
- .gitignore 排除覆盖率产物

---

## [3.3.3] - 2026-04-25

### Added

- **请求体大小限制中间件**（`extend/middleware/bodylimit.go`）
  - Content-Length 预检 + `MaxBytesReader` 双重保护
  - 默认 10MB，支持 `BodyLimitWithSize(bytes)` 自定义
  - `ParseBodyLimit("10MB")` 字符串解析（B/KB/MB/GB）
  - `BodyLimitFromString("1GB")` 便捷工厂函数

---

## [3.3.2] - 2026-04-25

### Added

- **请求超时控制中间件**（`extend/middleware/timeout.go`）
  - `context.WithTimeout` 传播超时信号到下游 handler
  - 超时返回 504 Gateway Timeout
  - 默认 30s，支持 `TimeoutWithDuration` 自定义

---

## [3.3.1] - 2026-04-25

### Added

- **熔断器中间件**（`extend/middleware/circuitbreaker.go`）
  - 三态状态机：Closed → Open → HalfOpen → Closed
  - 滑动窗口错误率统计
  - 可配置窗口大小 / 错误阈值 / 冷却时间 / 试探请求数
  - `CircuitBreaker()` 默认配置 + `CircuitBreakerWithConfig()` 自定义

---

## [3.3.0] - 2026-04-25

本版本聚焦**开发体验与代码质量**：WebSocket 实时通信、Context 值类型安全、脚手架子命令扩展。

### Added

- **WebSocket 支持**（`extend/websocket/`）
  - 基于 gorilla/websocket 封装，`Handler()` 一行升级 HTTP → WS
  - `Conn` 包装线程安全写（`WriteJSON` / `WriteSafeMessage`）
  - `Hub` 广播模型：Register / Unregister / Broadcast / BroadcastJSON
  - 支持自定义 Upgrader（如严格 Origin 检查）
- **Context Key 类型安全**（`app/ctxkeys/`）
  - 统一 `thinkgin:` 命名空间常量，消除裸字符串 key
  - 提供 `GetRequestID` / `SetRequestID` / `GetCSRFToken` / `SetCSRFToken` 类型安全函数
- **脚手架扩展**（`cmd/scaffold`）
  - `new middleware <name>` — 生成中间件骨架到 `extend/middleware/`
  - `new migration <name>` — 生成带时间戳的迁移文件到 `app/database/migrations/`

### Fixed

- JWT tampered signature 测试不稳定（翻转整段签名替代单字符翻转）
- golangci-lint 报错：`gzipMinSize` 未使用、`Close()` 返回值未检查

### Internal

- 所有中间件 / API 响应中的裸字符串 context key 迁移为 `ctxkeys` 包
- i18n 内部 key 统一为 `thinkgin:i18n:bundle` 常量
- REVIEW.md 15/15 项全部标注 ✅ 完成

### Migration Notes

从 v3.2.0 升级到 v3.3.0：

1. `c.GetString("request_id")` 建议改为 `ctxkeys.GetRequestID(c)`，旧写法仍可用但 key 已变为 `thinkgin:request_id`。
2. `c.GetString("csrf_token")` 建议改为 `ctxkeys.GetCSRFToken(c)`，key 已变为 `thinkgin:csrf_token`。
3. 新增 `gorilla/websocket` 依赖，运行 `go mod tidy` 更新。

---

## [3.2.0] - 2026-04-25

本版本聚焦**架构解耦、安全加固、运维能力扩展**：Logger 接口化、CORS 安全修复、HTTPS 双监听、ServiceContext 依赖注入、六大新中间件、数据库迁移系统、配置热更新、分布式限流与 i18n 运行时。

### Added

- **Logger 接口抽象**（`app/log_iface.go`）
  - 定义 `Logger` 接口，解耦业务代码与具体日志库
  - `LogrusAdapter`：包装 `*logrus.Logger`（默认实现）
  - `SlogAdapter`：包装 `*slog.Logger`，兼容 Go 1.21+ 标准库
  - `WithFields` 返回新实例，线程安全
- **HTTPS 双监听**（`framework/app.go`）
  - 配置 `server.https.enabled=true` 后自动并发启动 HTTP + HTTPS
  - TLS 最低版本强制 TLS 1.2
  - Shutdown 同时优雅关闭两个 server
- **ServiceContext 依赖容器**（`app/service_context.go`）
  - 聚合 Config / Logger / DB / Cache 核心依赖
  - `SvcMiddleware` 注入 gin.Context，Handler 通过 `SvcFromGin(c)` 获取
  - 未注入时兜底全局变量，完全向后兼容
- **CSRF 防护中间件**（`extend/middleware/csrf.go`）
  - 双重提交 Cookie（Double Submit Cookie）方案
  - 支持 Header `X-CSRF-Token` 和表单字段 `_csrf_token`
- **Gzip 响应压缩**（`extend/middleware/gzip.go`）
  - `sync.Pool` 复用 gzip.Writer，高性能
  - 含请求体解压中间件 `GzipDecompressRequest`
- **安全响应头**（`extend/middleware/secure_headers.go`）
  - 默认启用 X-Content-Type-Options / X-Frame-Options / XSS-Protection / Referrer-Policy
  - 支持配置覆盖或禁用
- **Swagger/OpenAPI 端点**（`route/swagger.go`）
  - `/swagger/` Swagger UI（CDN 加载，零 Go 依赖）
  - `/swagger/doc.yaml` + `/swagger/doc.json` 静态规范文件
  - 仅 debug 或 `swagger_enabled=true` 时暴露
- **数据库迁移系统**（`app/database/migrate.go`）
  - `Migrator` 管理器：Register → Migrate → Rollback → Status
  - `_migrations` 表自动建表，版本字典序排序
  - 迁移以 Go 函数注册，类型安全
- **配置热更新**（`app/hot_reload.go`）
  - `fsnotify` 监听配置目录 YAML 变更
  - 去抖 500ms 合并高频写入
  - 回调函数通知业务层刷新运行时状态
- **Redis 分布式限流**（`extend/middleware/ratelimit_redis.go`）
  - Lua 脚本原子性令牌桶算法
  - Redis 不可用时自动降级为 no-op
- **i18n 运行时**（`extend/i18n/`）
  - YAML 翻译文件按 locale 加载
  - `Accept-Language` + `?lang=` 查询参数自动语言检测
  - `{{.Key}}` 模板变量替换
  - Gin 中间件 + `T(c, key)` 便捷函数
- **WebSocket 支持**（`extend/websocket/`）
  - 基于 gorilla/websocket 封装，`Handler()` 一行升级 HTTP → WS
  - `Conn` 包装线程安全写（`WriteJSON` / `WriteSafeMessage`）
  - `Hub` 广播模型：Register / Unregister / Broadcast / BroadcastJSON
  - 支持自定义 Upgrader（如严格 Origin 检查）
- **Makefile** — 统一 `build` / `test` / `lint` / `fmt` / `scaffold` 等命令

### Changed

- 全局 `app.Logger` 从 `*logrus.Logger` 改为 `Logger` 接口
- `framework.App.logger` / `WithLogger` 改为接口类型
- 所有中间件移除 logrus 直接依赖，改用 `app.GetLogger()` 接口
- `route.InitRouter` 自动注入 `SvcMiddleware`
- `applyMiddleware` 新增 `secure_headers` / `gzip` / `csrf` / `redis_rate_limit` 映射

### Fixed

- **CORS origin 安全 bug**：`Access-Control-Allow-Origin` 不再逗号拼接多个 origin，改为逐请求动态匹配并回写单个 origin
- **Rate Limiter 内存泄漏**：新增 TTL 过期清理机制，定期清理不活跃 IP 桶

### Internal

- 新增 20+ 单元测试文件，覆盖 bootstrap / logger / adapter / service_context / hot_reload / migrate / csrf / gzip / secure_headers / i18n 等模块
- 新增 `ReBootstrap()` 函数供测试绕过 `sync.Once`
- `go test ./...` 全量绿

### Migration Notes

从 v3.1.0 升级到 v3.2.0：

1. `app.GetLogger()` 返回类型从 `*logrus.Logger` 变为 `app.Logger` 接口。如果你的代码直接访问 logrus 特有方法，需改为接口方法或通过 `adapter.L` 访问底层实例。
2. `framework.WithLogger()` 参数类型从 `*logrus.Logger` 改为 `app.Logger`。传入时使用 `app.NewLogrusAdapter(l)` 包装。
3. `middleware.global` 列表可追加 `secure_headers` / `gzip` / `csrf` / `redis_rate_limit` 四个新中间件。
4. 配置热更新需显式调用 `app.WatchConfig(dir, callback...)`，不会自动启用。
5. 新增 `fsnotify` 依赖，运行 `go mod tidy` 更新。

---

## [3.1.0] - 2026-04-20

本版本聚焦**运行时能力的补全与生产就绪度**：Session / JWT 能力落地、K8s 探针串联真实依赖、默认配置瘦身实现零依赖开箱跑。

### Added

- **Session 运行时**（`app/session`）
  - `memory` / `redis` 两种 Store，接口统一
  - Gin 中间件 + `Session.From(c)` / `Save` / `Destroy`
  - Cookie 自动携带 `HttpOnly` / `Secure` / `SameSite`
  - Session ID 采用 256-bit `crypto/rand` base64url
- **JWT 鉴权中间件**（`extend/middleware/jwt.go`）
  - HS256 签发 / 校验，严格拒绝 `alg=none`
  - `JWTIssue` / `JWTAuth` / `JWTFromContext`
  - 密钥从 `config/app.yaml` 的 `app.jwt.secret` 读取，为空即拒绝
- **命令行脚手架**（`cmd/scaffold`）
  - `go run ./cmd/scaffold new module <name>` 生成 controller / model / view 骨架
- **健康探针串联**（`route/health.go`）
  - `/readyz` 并发 Ping 所有已注册的 DB / Redis
  - 2s 硬超时，任一失败返 503 + 详细组件状态
  - 无依赖时返回 200，兼容纯 HTTP 框架场景
- **Prometheus 监控拆分**
  - `extend/middleware/prometheus/` 按 HTTP / 运行时 / 业务维度拆分
  - 总开关 + 模块开关分离控制
- **可插拔中间件映射** 新增 `session` / `jwt` / `jwt_auth`
- 数据库迁移指南：`docs/database-migration.md`（推荐 `pressly/goose`）

### Changed

- **默认配置瘦身**：`database.yaml` / `cache.yaml` / `session.yaml` 不再预置任何真实连接
  - 新人 `go run main.go` 启动**无任何** DB / Redis / Session 告警
  - 需要依赖时从配置文件的注释块按需放开
- `config/app.go` 634 行拆分为 7 个单一职责文件
- `main.go` banner 版本号改由配置动态拼接，避免升级遗漏
- GORM 接入 OpenTelemetry 插件：SQL 查询自动纳入当前 HTTP trace
- `/livez` 与 `/readyz` 语义明确化：`livez` 仅进程存活检查，不查依赖

### Fixed

- `route.yaml` 错写视图内容的历史 bug
- Windows 下模板加载与日志路径兼容问题
- Session file driver 未实现但作为默认值导致启动告警
- 意外提交的运行期 `database/*.db` 文件纳入 `.gitignore`

### Internal

- CI 三矩阵（Ubuntu / Windows / macOS）+ `-race` + coverage
- `golangci-lint` 接入，规则见 `.golangci.yml`
- 新增单测覆盖：session (7)、jwt (7)、scaffold、health 等
- `go test ./...` 全量绿

### Migration Notes

从 v3.0.0 升级到 v3.1.0：

1. 如果你依赖之前 `database.yaml` 预置的 mysql / pgsql / redis 条目，需要从注释块里放出来并改成你的真实地址。
2. `config/session.yaml` 的 `driver` 默认值从 `file` 改为 `memory`。如需多实例共享 Session，改为 `redis`。
3. `middleware.global` 里可追加 `session` / `jwt` 两个新中间件。
4. `/readyz` 行为变化：有 DB/Redis 连接时会真实探活，依赖不可达会返 503 —— 如有 k8s 探针配置请确认超时阈值 ≥ 3s。

---

## [3.0.0] - 2026-01-XX（历史版本）

首个公开版本，定位为 ThinkPHP 生态下的 Go 替代方案。详见 git 历史。
