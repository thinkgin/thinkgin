# Changelog

本项目遵循 [Semantic Versioning](https://semver.org/lang/zh-CN/) 和 [Keep a Changelog](https://keepachangelog.com/zh-CN/1.1.0/) 约定。

## [3.12.0] - 2026-05-31

本版本为 **P1 正确性加固批次**：修复熔断器并发/低流量缺陷、缓存击穿与 goroutine 泄漏，并消除多处文档与代码不一致。

### Added

- **熔断器最小请求数 `MinRequests`**（`extend/middleware/circuitbreaker.go`）：新增配置项，样本数达到 `MinRequests` 且错误率超阈值即可熔断，无需填满整个滑动窗口
  - 默认 10，为 0 时回退为 `WindowSize`（保持旧行为，向后兼容），取值自动 cap 到 `WindowSize`
  - 解决"低流量接口永远凑不满 `WindowSize` 个样本，从而永不熔断"的问题

### Fixed

- **熔断器 HalfOpen 并发准入竞态**：`allow()` 在 HalfOpen 准入时即占用试探配额（`halfOpenTotal++`），不再等到 `record()` 才计数。修复高并发下大量请求同时通过 `halfOpenTotal < max` 判断、导致远超 `HalfOpenMaxRequests` 的请求涌入的问题
- **缓存击穿（cache stampede）**：`MemoryStore.Remember` / `RedisStore.Remember` 接入 `golang.org/x/sync/singleflight`，同一 key 的并发未命中合并为一次回源（含二次检查），避免缓存失效瞬间大量请求穿透到后端
- **内存缓存 goroutine 泄漏**：`cache.NewStore` / `cache.DefaultStore` 的内存实现改为进程级共享单例（`sync.Once`），不再每次调用都新建一个带后台 GC goroutine 的 `MemoryStore`

### Changed

- **secure_headers 文档校准**：包注释不再声称默认启用 `Content-Security-Policy`（实际未启用，CSP 与具体页面强相关），改为说明如何通过 `middleware.config.secure_headers` 按需配置
- **gzip 中间件落实流式跳过**：现在会跳过 SSE（`text/event-stream`）与已设置 `Content-Encoding` 的响应，与注释承诺一致；移除未实现的 `minSize` 死注释
- **validation 字段名使用真实 json tag**：通过 validator 的 `RegisterTagNameFunc`，字段级错误信息中的字段名取自 `json` tag（正确支持 `snake_case`），替换原先"仅首字母小写"的错误实现
- `golang.org/x/sync` 从间接依赖提升为直接依赖

---

## [3.11.1] - 2026-05-31

### Security

本版本是一次**安全加固批次（P0）**，修复 5 项不安全的默认配置与高危操作。**包含若干默认行为变更，升级前请阅读下方迁移说明。**

- **移除明文 JWT 密钥**：`config/app.yaml` 中硬编码的示例密钥（`23347$040412`）清空，改为通过环境变量 `THINKGIN_APP_JWT_SECRET` 注入
  - `app/validate.go` 新增 `validateJWTSecret()` 启动校验：空密钥提示注入方式；命中弱口令黑名单（含历史泄漏密钥）强告警；长度 < 32 字节提示。仅告警不中断启动
- **CORS 安全默认**：`config/middleware.yaml` 的 `allow_credentials` 由 `true` 改为 `false`，消除 `allow_origins=["*"]` + 携带凭证的危险组合
- **Session Cookie 强制 HttpOnly**：`app/session/session.go` 中 Cookie 的 `HttpOnly` 恒为 `true`（即使配置为 false 也纠正），杜绝 XSS 窃取会话 ID
- **RedisStore.Flush 限定前缀**：`app/cache/redis_store.go` 的 `Flush()` 由 `FlushDB`（清空整库）改为按 `prefix+"*"` 的 `SCAN`+`DEL`；`cache.prefix` 为空时拒绝执行，防止误清与 session/限流共用的整个 Redis DB
- **/metrics 鉴权姿态加固**：`config/prometheus.yaml` 移除弱占位密码 `prom/changeme`；release 模式下暴露无鉴权 `/metrics` 时打印启动告警；`auth.enabled=true` 但凭据为空时禁用端点防误暴露；修正 `app/types.go` 中"鉴权未实现"的过时注释

### Migration Notes

从 v3.11.0 升级到 v3.11.1：

1. **JWT**：默认密钥已清空。使用 JWT 的项目须通过环境变量 `THINKGIN_APP_JWT_SECRET=<至少32位随机串>` 注入后才能签发/校验（与原本"空密钥拒签"语义一致，仅移除了不安全的内置默认值）。
2. **CORS**：默认不再携带凭证。若业务依赖跨域 Cookie/Authorization，需把 `config/middleware.yaml` 的 `allow_origins` 改为显式可信域名白名单，再把 `allow_credentials` 设回 `true`。
3. **Session**：`cookie.http_only` 现在恒为 `true`，配置写 `false` 会被强制纠正。
4. **缓存**：`RedisStore.Flush` 现在只清除 `cache.prefix` 前缀下的键；前缀为空时会返回错误而非清空全库，请确保 `config/cache.yaml` 设置了 `prefix`。

---

## [3.11.0] - 2026-05-07

### Added

- **自定义中间件注册表**（P0-2）：新增 `route.RegisterMiddleware(name, factory)` 公共 API，业务可在 `main` 早期注入自定义中间件，并在 `config/middleware.yaml` 的 `global` / `groups` 列表中按名启用，**彻底打破"hardcoded switch 不可扩展"的限制**
  - 新增文件 `route/registry.go`：基于 `sync.Map` 的全局注册表，工厂函数语义保证"每次 Use 独立实例"
  - 公共 API：`RegisterMiddleware` / `UnregisterMiddleware` / `LookupMiddleware` / `HasMiddleware`
  - **解析顺序**：自定义注册表优先 → 未命中回落到内置 switch → 仍未命中静默忽略；允许覆盖内置中间件（例如用业务自有的 access_log 替换框架默认）
  - 新增 8 个测试覆盖注册/查询/覆盖/并发场景，`route` 包测试覆盖率从 41.3% 提升至 47.4%
  - 完整使用示例见 [`README.md`](README.md#自定义中间件注册v3110)

### Why

v3.10.x 之前 `route.resolveMiddleware` 是一个 18 分支 hardcoded `switch`，第三方/业务**唯一的扩展方式是 fork 框架**，违反开闭原则，也是 v3.10.2 内部评估报告中标记为 **P0-2 结构性问题**的项目。本版本以 0.5 天工作量根治该问题，且**100% 向后兼容**——所有 v3.10.x 的配置无需任何改动即可运行。

---

## [3.10.3] - 2026-05-07

### Changed

- **定位声明校准**：明确 ThinkGin 的"单体业务后台脚手架"定位，下掉所有"高性能"宣传词
  - `README.md` 增加"定位声明"段落，诚实标注适合 / 不适合的场景，并指引高性能或微服务需求选用 Fiber/Hertz/Kratos/go-zero
  - `app/index/view/index.html` 默认欢迎页"高性能框架"改为"生产级 Web 框架"
  - 内部约定：性能不再是 ThinkGin 的核心卖点，**工程纪律 · 配置秩序 · 开箱即交付**才是

### Why

性能上限被基础组件（Gin + GORM）锁定，与 Fiber/Hertz 等 fasthttp/netpoll 系框架存在本质代差。继续宣传"高性能"会引发对比尴尬并误导用户。本版本是文档层校准，不涉及代码行为变更。

---

## [3.10.2] - 2026-05-06

### Fixed

- **超时中间件 -race 数据竞争**：v3.9.1 引入的"goroutine handler + select"模式中，超时分支返回前未等待 handler 子协程结束，导致 `go test -race` 在 GitHub Actions CI（macOS/Ubuntu/Windows）报告竞争
  - `case <-ctx.Done()` 分支在写入 504 后增加 `<-done` 等待 handler 协程清理，消除竞争
  - 业务 handler 应当监听 `c.Request.Context().Done()` 提前退出，否则会阻塞此处（属于编写规范要求）

---

## [3.10.1] - 2026-05-01

### Changed

- **全局可变状态标记 Deprecated**（REVIEW #5）：`app.Log` 全局变量添加 `Deprecated` 注释，引导用户迁移至 `ServiceContext.Log`
  - `app.Config` 已在 v3.7.0 标记 Deprecated，本次补齐 `app.Log`
  - 不破坏向后兼容性，仅文档引导

---

## [3.10.0] - 2026-05-01

### Changed

- **默认日志引擎从 logrus 切换为 slog**（REVIEW #11）：框架核心不再直接依赖 logrus
  - `InitLogger()` 和 `GetLogger()` 改用标准库 `log/slog` 作为默认实现
  - `validate.go` 中的 `logrus.ParseLevel()` 替换为自定义 `isValidLogLevel()` 函数
  - `logrus_adapter.go` 保留为可选适配器，用户可通过 `app.Log = app.NewLogrusAdapter(...)` 显式启用
  - 测试文件（`framework/app_test.go`、`service_context_test.go`）统一改用 `NewSlogAdapter`
  - logrus 从框架启动路径中完全移除，仅作为 go.mod 间接依赖保留

---

## [3.9.1] - 2026-05-01

### Fixed

- **超时中间件竞态窗口**（REVIEW #3）：重构为 "goroutine handler + select" 模式
  - `c.Next()` 移入子协程，主协程 `select` 确定性选择"完成"或"超时"分支，消除竞态窗口
  - 新增 `panicCh` 通道捕获 handler panic 并在主协程重新抛出，确保 Recovery 中间件正常工作
  - 新增 `TestTimeout_PanicRecovery` 测试用例

---

## [3.9.0] - 2026-05-01

### Changed

- **数据库/缓存配置强类型化**（REVIEW #10）：`DatabaseConfig.Connections` 从 `map[string]interface{}` 改为 `map[string]ConnectionConfig`，`CacheConfig.Stores` 从 `map[string]interface{}` 改为 `map[string]CacheStoreConfig`
  - 新增 `ConnectionConfig` 结构体：覆盖 MySQL/PostgreSQL/SQLite/Redis 全部字段，编译期类型校验
  - 新增 `CacheStoreConfig` 结构体：driver/connection/max_size/path 强类型
  - 移除 `database/open.go`、`cache/cache.go`、`session/redis_store.go` 中的 `strOr`/`intOr`/`toInt` 运行时类型断言工具函数
  - 所有消费方（database.Init / cache.Init / session.newRedisStore）改为直接读取结构体字段
  - 配套测试全部适配强类型

- **WebSocket 安全默认值**（REVIEW #14）：`DefaultUpgrader.CheckOrigin` 从 `return true` 改为配置驱动
  - 读取 `middleware.config.cors.allow_origins` 白名单
  - 未配置或配置为 `["*"]` 时放行，否则仅允许白名单内 Origin 升级 WebSocket

### Fixed

- **Redis 限流上下文传递**（REVIEW #13）：`ratelimit_redis.go` 的 `context.Background()` 改为 `c.Request.Context()`，支持请求取消和链路追踪
- **超时中间件 CloseNotify**（REVIEW #3）：为 `CloseNotify()` 添加 `Deprecated` 注释和 `nolint:staticcheck` 标记，明确标注为 Gin 接口强制要求

### Renamed

- `extend/middleware/coverage_boost_test.go` → `middleware_extra_test.go`（REVIEW #17）

---

## [3.8.1] - 2026-04-30

### Changed

- **统一版本号管理**：新增 `app/version.go` 作为版本号唯一来源（Single Source of Truth）
  - 支持 `go build -ldflags "-X thinkgin/app.Version=x.y.z"` 编译期注入
  - `defaults.go` 从 `app.Version` 变量读取，不再硬编码
  - `cmd/version.go` 直接使用 `app.Version`，确保版本号全局一致
  - 测试中也引用 `Version` 变量，版本更新无需再改测试文件

---

## [3.8.0] - 2026-04-30

### Changed

- **日志轮转替换为 lumberjack**：移除已归档的 `lestrrat-go/file-rotatelogs` 和 `rifflock/lfshook`
  - 使用 `gopkg.in/natefinch/lumberjack.v2` 替代，按文件大小（100MB）轮转 + 自动压缩旧文件
  - 保留 `MaxAge`（天数）配置项语义不变
  - 日志同时输出到 stderr 和文件（`io.MultiWriter`），便于容器环境采集
  - 跨平台稳定，不再需要 Windows symlink 特殊处理

---

## [3.7.2] - 2026-04-30

### Changed

- **熔断器按路由粒度隔离**：`CircuitBreakerWithConfig` 从全局单例改为按路由维护独立的熔断器实例
  - 使用 `sync.Map` + `LoadOrStore` 懒创建，路由 key 优先取 `c.FullPath()`（Gin 注册模板），避免高基数
  - 一个慢接口的错误率不再影响其他接口的熔断状态
  - 对标 Go-Zero 的按路由熔断设计

---

## [3.7.1] - 2026-04-30

### Fixed

- **LoadHTMLGlob 安全检查**：`registerStaticAndTemplates` 在调用 `LoadHTMLGlob` 前使用 `filepath.Glob` 检查是否有匹配的 `.html` 文件，纯 API 项目（无模板文件）不再 panic

---

## [3.7.0] - 2026-04-30

### Changed

- **配置热更新并发安全**：全局 `Config` 改为 `atomic.Pointer[GlobalConfig]` 原子指针
  - `GetConfig()` 使用原子读取，线程安全
  - 新增 `SetConfig(cfg)` 原子写入函数
  - 热更新 `reload()` 改为"构建新对象→原子替换"模式，消除读写竞态
  - `defaults.go` / `env.go` / `validate.go` 提取参数化版本（`setDefaultsOn` / `applyEnvOverridesOn` / `validateConfigOn`）
  - `loader.go` 新增 `loadNewConfigFromDir()` 构建独立配置对象
  - 旧 `app.Config` 变量保留为 Deprecated 兼容入口

---

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
