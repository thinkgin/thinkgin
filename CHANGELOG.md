# Changelog

本项目遵循 [Semantic Versioning](https://semver.org/lang/zh-CN/) 和 [Keep a Changelog](https://keepachangelog.com/zh-CN/1.1.0/) 约定。

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
