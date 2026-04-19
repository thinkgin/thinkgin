# route 包说明

负责把 Gin 引擎装配成一个可运行的路由树，**只做编排，不写业务**。

## 文件分工

| 文件 | 职责 |
|------|------|
| `router.go` | 入口 `InitRouter()`；注册全局中间件、静态资源、模板；分发到 Web / API / 监控路由 |
| `web.go`    | HTML 页面路由，通过模板渲染响应 |
| `api.go`    | REST API 路由，统一包一层 `APIErrorHandler` |

## 全局中间件

由 `config/middleware.yaml` 的 `middleware.global` 列表决定启用顺序。
若配置为空，使用 `router.go` 里的 `defaultMiddlewareChain` 兜底：

```
recovery → request_id → trace → access_log → prometheus
```

受支持的名称：`recovery` / `request_id` / `trace` / `logger` / `access_log` / `cors` / `rate_limit` / `prometheus`。未识别的名字会被忽略，方便前向兼容。

## 新增模块步骤

1. 在 `app/<module>/controller/` 新建 controller。
2. 在 `app/<module>/view/` 放 HTML 模板（若需要）。
3. 在 `api.go` 添加 `registerXxxAPIV1(v1)` 并在 `RegisterAPIRoutes` 中调用。
4. 若是页面路由，直接在 `web.go` 的 `RegisterWebRoutes` 内追加。

## 版本化策略

API 路径约定 `/api/v{n}/<module>/<action>`。发布不兼容变更时新增 `registerXxxAPIV2` 并挂到 `v2 := api.Group("/v2")`，v1 保持不动。

## 监控端点

`registerMonitoringRoutes` 统一暴露：

- `GET /livez` — 进程存活探针
- `GET /readyz` — 依赖就绪探针（当前同 /livez，后续可接入 DB/Redis）
- `GET /ping`  — 返回 `{ status, message, version }`，用于简易健康检查
- `GET /metrics` — 仅当 `app.monitoring.prometheus_enabled` 与 `prometheus.enabled` 同时为 true 时暴露
