// 本文件实现了一个极简的 GORM → OpenTelemetry 追踪插件，
// 利用 GORM 的 callback 钩子在每次 SQL 执行前后创建 / 结束 span。
//
// 设计取舍：
//   - 不引入 gorm.io/plugin/opentelemetry（它携带 Prometheus 耦合 + 更多间接依赖）。
//   - 复用 otel.Tracer，由 middleware.InitTracer 装配的全局 TracerProvider 自动接管。
//   - span 通过 Statement.Settings 存取，不污染 ctx.Value 命名空间。
//
// 业务代码只需 `db.WithContext(ctx).Find(&users)` 即可继承 HTTP span 形成端到端链路；
// 未传 ctx 的调用仍会产生 span，只是没有父链接——属于最弱保证。
package database

import (
	"context"
	"errors"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
	"gorm.io/gorm"
)

// tracerName 用作 otel.Tracer 的 instrumentation 名，便于在后端按来源过滤。
const (
	tracerName = "thinkgin/app/database"
	spanKey    = "thinkgin:span" // 在 Statement.Settings 中存放 span 的 key
)

// tracingPlugin 是实现 gorm.Plugin 接口的追踪插件。
type tracingPlugin struct{}

// Name 返回插件名，GORM 要求唯一。
func (tracingPlugin) Name() string { return "thinkgin:tracing" }

// Initialize 把 before/after callback 注册到 GORM 六个操作通道。
// GORM 的 processor 类型未导出，这里无法通过抽象复用，只能展开写 6 次调用。
// Register 要求名称唯一，统一使用 "thinkgin:trace:<op>:<phase>" 前缀避免冲突。
func (tracingPlugin) Initialize(db *gorm.DB) error {
	cb := db.Callback()
	var errs []error

	reg := func(err error) {
		if err != nil {
			errs = append(errs, err)
		}
	}

	reg(cb.Create().Before("gorm:create").Register("thinkgin:trace:create:before", traceBefore("create")))
	reg(cb.Create().After("gorm:create").Register("thinkgin:trace:create:after", traceAfter))

	reg(cb.Query().Before("gorm:query").Register("thinkgin:trace:query:before", traceBefore("query")))
	reg(cb.Query().After("gorm:query").Register("thinkgin:trace:query:after", traceAfter))

	reg(cb.Update().Before("gorm:update").Register("thinkgin:trace:update:before", traceBefore("update")))
	reg(cb.Update().After("gorm:update").Register("thinkgin:trace:update:after", traceAfter))

	reg(cb.Delete().Before("gorm:delete").Register("thinkgin:trace:delete:before", traceBefore("delete")))
	reg(cb.Delete().After("gorm:delete").Register("thinkgin:trace:delete:after", traceAfter))

	reg(cb.Row().Before("gorm:row").Register("thinkgin:trace:row:before", traceBefore("row")))
	reg(cb.Row().After("gorm:row").Register("thinkgin:trace:row:after", traceAfter))

	reg(cb.Raw().Before("gorm:raw").Register("thinkgin:trace:raw:before", traceBefore("raw")))
	reg(cb.Raw().After("gorm:raw").Register("thinkgin:trace:raw:after", traceAfter))

	return errors.Join(errs...)
}

// traceBefore 返回执行前回调：开启 span 并放入 Statement.Settings。
// span 名使用 "db.<op>"（例如 db.create），与 OTel 语义约定对齐。
func traceBefore(op string) func(db *gorm.DB) {
	return func(db *gorm.DB) {
		ctx := db.Statement.Context
		if ctx == nil {
			ctx = context.Background()
		}
		_, span := otel.Tracer(tracerName).Start(ctx, "db."+op,
			trace.WithSpanKind(trace.SpanKindClient),
			trace.WithAttributes(attribute.String("db.operation", op)),
		)
		db.Statement.Settings.Store(spanKey, span)
	}
}

// traceAfter 在 SQL 执行后补齐属性并结束 span。
// 若 db.Error 非空则标记 span 为错误状态，便于后端过滤异常调用。
func traceAfter(db *gorm.DB) {
	raw, ok := db.Statement.Settings.LoadAndDelete(spanKey)
	if !ok {
		return
	}
	span, ok := raw.(trace.Span)
	if !ok || span == nil {
		return
	}
	defer span.End()

	if db.Statement.Table != "" {
		span.SetAttributes(attribute.String("db.sql.table", db.Statement.Table))
	}
	if db.Statement.SQL.Len() > 0 {
		span.SetAttributes(attribute.String("db.statement", db.Statement.SQL.String()))
	}
	span.SetAttributes(attribute.Int64("db.rows_affected", db.Statement.RowsAffected))

	if err := db.Error; err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
	}
}
