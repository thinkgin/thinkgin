# ThinkGin2 框架化与多平台迭代实施计划

> 目标：在 **保留现有路由模式**（`route/router.go` + `route/web.go` + `route/api.go`）的基础上，逐步演进为可复用框架/库，适配 Windows/macOS/Linux + Docker/k8s，并提升可维护性、可观测性与稳定性。
>
> 范围说明：本计划优先解决“框架级基础设施”问题（配置、日志、指标、生命周期、跨平台），业务模块（如 user/order）仅作为示例。

---

## 0. 当前基线（Baseline）

### 0.1 关键入口

- `main.go`
  - 读取 `app.GetConfig()`
  - 调用 `route.InitRouter()`
  - 创建 `http.Server` 并 `ListenAndServe()`
  - 启动时在 Windows 自动打开浏览器（`cmd /c start`）

### 0.2 路由模式（必须保留）

- `route/router.go`: 初始化 gin engine、全局中间件、静态资源、web/api/monitoring 注册
- `route/web.go`: Web 页面路由（HTML 渲染）
- `route/api.go`: API 路由（版本化 API）

> 约束：后续迭代 **不改变“web/api 分文件注册”的使用方式**；仅对内部实现进行增强（配置化、模块化、可插拔）。

---

## 1. 总体目标与非目标

### 1.1 目标

- 框架可被其它项目通过 `go get` 依赖并快速启动
- 支持多平台与容器化运行：Windows/macOS/Linux、Docker、k8s
- 配置可靠（支持按文件加载、默认值、校验、环境变量覆盖）
- 具备生命周期管理：优雅退出、健康检查（liveness/readiness）
- 可观测性稳定：统一日志、Prometheus 指标线程安全、避免 label 爆炸

### 1.2 非目标（本阶段不强求）

- 不立即引入复杂 DI 容器
- 不立即实现所有配置模块的 runtime 组件（database/cache/session/trace 全落地）
- 不在此计划中重写业务目录结构（`app/index/...` 作为示例保留）

---

## 2. 里程碑与阶段划分

> 推荐顺序：P0（可靠性/跨平台）→ P1（框架化 API / 生命周期）→ P2（可观测/扩展能力）→ P3（稳定化 v1.0）

### P0（必须优先）：可靠性与跨平台基础

#### P0.1 配置加载重构（修复合并覆盖缺陷）

- **问题**：当前 `app/config.go` 通过 `mergeConfig()` 使用“非零值覆盖整块 struct”方式合并，存在字段丢失/配置不生效风险。
- **目标**：配置加载改为“按文件加载到对应子结构 + defaults + validate”，并支持环境变量覆盖（适配 k8s）。

**建议实现方式（推荐）**

- 按文件加载：
  - `config/app.yaml` → `Config.App`
  - `config/server.yaml` → `Config.Server`
  - ...
- 加载后执行：
  - `SetDefaults()`：填充默认值
  - `Validate()`：缺失项/非法值报错（至少返回 error，不能静默失败）
- Env 覆盖策略（最小可用）：
  - 约定变量前缀，例如 `THINKGIN_`
  - 示例：`THINKGIN_SERVER_HTTP_PORT=8000`

**影响文件**

- `app/config.go`
- （新增/拆分可选）`app/config_loader.go` / `internal/config/...`（若后续要拆框架）

**验收标准**

- 修改任意单个 YAML 文件不会导致其它模块配置被覆盖为零值
- 不依赖 `port != 0` 等条件判断即可正确加载 server 的其它字段
- 缺配置时能给出明确错误（或至少 warning + 默认值说明）

#### P0.2 Prometheus 业务指标并发安全

- **问题**：`businessCounters`/`businessHistograms`/`businessGauges` 是普通 map，多协程并发创建指标会 data race。
- **目标**：指标注册线程安全；并提供推荐用法避免运行时重复注册。

**建议实现方式**

- 使用 `sync.Map` 或 `map + sync.RWMutex`
- 或提供 `RegisterBusinessMetric(name, ...)` 在启动期集中注册

**影响文件**

- `extend/middleware/prometheus.go`

**验收标准**

- `go test -race`（或压测场景）不出现数据竞争

#### P0.3 入口跨平台与环境友好

- **问题**：`main.go` 默认执行 Windows 打开浏览器，不适用于 macOS/Linux/Docker/k8s。
- **目标**：移除或配置化该行为，框架默认不做“打开浏览器”。

**建议实现方式**

- 默认关闭，仅在开发环境 + 配置开启时执行
- 多平台实现：按 `runtime.GOOS` 分支，或完全移到示例应用层

**影响文件**

- `main.go`

**验收标准**

- 在 Linux/Docker/k8s 不会尝试打开浏览器
- Windows/macOS 本地可选启用

---

### P1（框架化第一步）：生命周期、健康检查、可复用启动 API

> 该阶段仍保留现有 `route/*.go` 模式。

#### P1.1 优雅退出（Graceful Shutdown）

- **目标**：捕获 `SIGINT/SIGTERM`，在超时内 `Shutdown()` HTTP Server。

**建议实现方式**

- `main.go` 或框架层：
  - `ctx, stop := signal.NotifyContext(...)`
  - goroutine 启动 server
  - `<-ctx.Done()` 后调用 `server.Shutdown(timeoutCtx)`

**影响文件**

- `main.go`

**验收标准**

- k8s 下 `terminationGracePeriodSeconds` 内能优雅退出

#### P1.2 健康检查端点（k8s probes 友好）

- **目标**：提供 `/livez`（进程活着即可）和 `/readyz`（依赖健康才 ready）。
- **现状**：当前有 `/ping`，建议保留兼容，但新增标准端点。

**建议实现方式**

- 在 `route/router.go` 的 monitoring routes 中新增：
  - `/livez`：固定返回 200
  - `/readyz`：检查关键依赖（先可只返回 200，后续接 DB/Redis）

**影响文件**

- `route/router.go`

**验收标准**

- 可直接用于 k8s liveness/readiness probes

#### P1.3 框架化 API（建议但不强制一次到位）

- **目标**：逐步将启动逻辑从 `main.go` 抽到一个可复用包（例如 `thinkgin` 包），使其它项目可复用。

**建议形态（最小）**

- `thinkgin.New()` 返回一个包含 router/server 的 App 对象
- `App.Run(ctx)` / `App.Shutdown(ctx)`
- 支持 Option：`WithConfig(...)`、`WithLogger(...)`

**说明**

- 若暂不拆 repo，可先在当前 module 内实现一个包（例如 `framework/` 或根包 `thinkgin`）作为过渡。

---

### P2（可观测与工程化）：日志统一、指标质量、配置化中间件

#### P2.1 日志统一（避免重复记录）

- **问题**：`gin.Logger()` + `LoggerToFile()` 会重复记录。
- **目标**：提供统一 access log，并可配置跳过路径（`/metrics` `/livez` 等）。

**建议实现方式**

- 在 `route/router.go` 选择：
  - 保留 `gin.Recovery()`
  - access log 使用自定义（logrus/zap 任一）
- 增加 request_id：写入响应头 + log fields
- 避免记录完整 query（或提供脱敏）

**影响文件**

- `route/router.go`
- `extend/middleware/logger.go`

#### P2.2 Prometheus label 策略

- **问题**：当前 `IncludePath=false` 时把 path 强制写成 `api`，会把 web/api/static 混合聚合。
- **目标**：提供更合理的 label 策略，避免 label 爆炸同时保留可观测性。

**建议实现方式**

- 建议 label：`method` + `route`（可选） + `status` + `scope`（web/api/static）
- 对于未命中路由（`c.FullPath()==""`）做兜底值

**影响文件**

- `extend/middleware/prometheus.go`

#### P2.3 中间件配置化（逐步使用 `config/middleware.yaml`）

- **目标**：让中间件启用/禁用、跳过路径等通过配置控制。

**影响文件**

- `config/middleware.yaml`
- `route/router.go`

---

### P3（v1.0 稳定化）：稳定 API、模块化扩展、示例与文档

#### P3.1 模块机制（保留 route 模式的前提下增强）

- **目标**：在不破坏 `route/web.go`/`route/api.go` 结构的情况下，引入“模块自注册”的可选能力。

**建议做法**

- `route/web.go`/`route/api.go` 继续存在，作为聚合入口
- 新增：模块可实现 `RegisterWebRoutes(r *gin.Engine)` / `RegisterAPIRoutes(r *gin.Engine)`
- 聚合层循环加载模块（可由配置控制加载哪些模块）

#### P3.2 拆分 repo/module（推荐）

- `thinkgin`：框架库（稳定 API）
- `thinkgin-examples`：示例应用（保留当前 app/index）

#### P3.3 文档与兼容策略

- 发布版本策略：
  - v0.x：允许 breaking changes
  - v1.0：对外导出 API 尽量保持兼容

---

## 3. 建议的执行顺序（最小可用路径）

- 第 1 周：完成 P0.1 + P0.2 + P0.3
- 第 2 周：完成 P1.1 + P1.2（k8s 可上线）
- 第 3-4 周：完成 P2.1 + P2.2 + P2.3（可观测/运维质量显著提升）
- 后续：推进 P3（模块化、拆分 repo、稳定 API）

---

## 4. 风险与注意事项

- 配置重构是破坏性变更风险最大的部分：建议先写一组“配置加载单测/示例”验证。
- Prometheus 指标一旦发布，指标名/label 变更会影响 Grafana/告警：建议在 v0.x 阶段尽早定型。
- Docker/k8s 环境默认不应写入不可写目录：日志目录、上传目录建议可配置并支持 stdout。

---

## 5. 验收清单（Definition of Done）

- 配置加载：单文件变更不会覆盖其它配置；支持 env 覆盖；有默认值与校验
- 多平台：Windows/macOS/Linux 运行无平台特有副作用（默认不打开浏览器）
- Docker/k8s：支持优雅退出；具备 `/livez` `/readyz`
- 可观测：access log 不重复；Prometheus 线程安全；label 不爆炸

