// 本文件定义 ThinkGin 的日志抽象接口。
//
// 设计目标：
//   - 解耦业务代码与具体日志库（logrus/zap/slog）。
//   - 保持与 logrus 调用风格兼容，降低迁移成本。
//   - WithFields 返回新 Logger 实例，线程安全。
package app

// Logger 是 ThinkGin 全局日志抽象。
// 所有中间件和业务代码应依赖此接口，而非具体实现。
type Logger interface {
	Debug(args ...any)
	Debugf(format string, args ...any)
	Info(args ...any)
	Infof(format string, args ...any)
	Warn(args ...any)
	Warnf(format string, args ...any)
	Error(args ...any)
	Errorf(format string, args ...any)
	Fatal(args ...any)
	Fatalf(format string, args ...any)
	// WithFields 返回携带附加字段的新 Logger，不修改原实例。
	WithFields(fields map[string]any) Logger
}
