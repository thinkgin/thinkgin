// 本文件提供中间件名称到实现的 **运行时可扩展注册表**。
//
// 背景：
//   v3.11 之前 resolveMiddleware 是一个 hardcoded switch，第三方/业务无法
//   把自定义中间件挂到 config/middleware.yaml 的 global / groups 列表里。
//   要扩展唯一办法是 fork 框架，违反开闭原则。
//
// 本注册表的语义：
//   1. 用户在 main 早期（route.InitRouter 之前）调用 RegisterMiddleware。
//   2. 解析 middleware.yaml 时，自定义注册表优先于内置 switch 命中。
//      —— 这意味着用户可以"覆盖"内置中间件（例如自定义 logger 替换 access_log）。
//   3. UnregisterMiddleware 用于测试隔离与运行期热替换。
//   4. 未命中自定义注册表时，自动回落到内置 switch（向后兼容）。
//
// 线程安全：sync.Map 保证并发注册/查询安全；典型场景是 main 启动期一次性注册，
// 运行期只读，几乎无竞争开销。
package route

import (
	"sync"

	"github.com/gin-gonic/gin"
)

// MiddlewareFactory 是中间件工厂函数：返回一个 gin.HandlerFunc 实例。
// 使用工厂而非直接的 HandlerFunc 是为了：
//   - 让中间件在每次 Use 时拿到独立的内部状态（如令牌桶/熔断器实例）。
//   - 给"懒初始化"留出空间（例如 prometheus 的 InitPrometheusMetrics）。
type MiddlewareFactory func() gin.HandlerFunc

// customRegistry 持有用户注册的自定义中间件名 → 工厂映射。
// 使用 sync.Map 而非 mutex+map：注册期写多读少，运行期读多写零，
// sync.Map 的"读不锁"特性在请求路径上零开销。
var customRegistry sync.Map // map[string]MiddlewareFactory

// RegisterMiddleware 注册一个自定义中间件名。
//
// 调用时机：必须在 route.InitRouter() 之前完成（一般在 main 包早期）。
// 重复注册同名中间件会覆盖前一次注册（"最后写入者获胜"）。
// name 为空或 factory 为 nil 时被静默忽略，避免污染注册表。
//
// 用法示例：
//
//	route.RegisterMiddleware("audit", func() gin.HandlerFunc {
//	    return myaudit.New(myaudit.Config{...})
//	})
//	// 然后在 config/middleware.yaml 的 global 列表中写 "audit" 即可生效
func RegisterMiddleware(name string, factory MiddlewareFactory) {
	if name == "" || factory == nil {
		return
	}
	customRegistry.Store(name, factory)
}

// UnregisterMiddleware 移除一个已注册的自定义中间件名。
// 主要用于测试隔离；生产代码不建议在运行期移除。
// 名称未注册时静默返回，调用方无需判断存在性。
func UnregisterMiddleware(name string) {
	customRegistry.Delete(name)
}

// LookupMiddleware 查询自定义注册表，命中返回工厂构造的 HandlerFunc，
// 未命中返回 nil（让上层回落到内置 switch）。
//
// 注意：每次调用都会**新建一个 HandlerFunc 实例**，符合"每次 Use 独立状态"的语义。
func LookupMiddleware(name string) gin.HandlerFunc {
	v, ok := customRegistry.Load(name)
	if !ok {
		return nil
	}
	factory, ok := v.(MiddlewareFactory)
	if !ok || factory == nil {
		return nil
	}
	return factory()
}

// HasMiddleware 报告 name 是否已存在于自定义注册表。
// 与 LookupMiddleware 的区别：本函数不构造 HandlerFunc 实例，零分配，仅做存在性判断。
func HasMiddleware(name string) bool {
	_, ok := customRegistry.Load(name)
	return ok
}

// resetRegistryForTest 仅供包内测试使用：清空整个注册表。
// 不导出避免被业务代码误用。
func resetRegistryForTest() {
	customRegistry.Range(func(k, _ any) bool {
		customRegistry.Delete(k)
		return true
	})
}
