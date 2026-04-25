package app

import "github.com/sirupsen/logrus"

// LogrusAdapter 将 *logrus.Logger 包装为 Logger 接口。
// 这是 ThinkGin 的默认日志实现，保持向后兼容。
type LogrusAdapter struct {
	L *logrus.Logger
	// fields 保存 WithFields 累积的字段。
	fields logrus.Fields
}

// NewLogrusAdapter 从 *logrus.Logger 创建适配器。
// l 为 nil 时自动创建默认 logrus 实例，便于测试。
func NewLogrusAdapter(l *logrus.Logger) *LogrusAdapter {
	if l == nil {
		l = logrus.New()
	}
	return &LogrusAdapter{L: l}
}

func (a *LogrusAdapter) entry() *logrus.Entry {
	if len(a.fields) > 0 {
		return a.L.WithFields(a.fields)
	}
	return logrus.NewEntry(a.L)
}

func (a *LogrusAdapter) Debug(args ...any)                 { a.entry().Debug(args...) }
func (a *LogrusAdapter) Debugf(format string, args ...any) { a.entry().Debugf(format, args...) }
func (a *LogrusAdapter) Info(args ...any)                  { a.entry().Info(args...) }
func (a *LogrusAdapter) Infof(format string, args ...any)  { a.entry().Infof(format, args...) }
func (a *LogrusAdapter) Warn(args ...any)                  { a.entry().Warn(args...) }
func (a *LogrusAdapter) Warnf(format string, args ...any)  { a.entry().Warnf(format, args...) }
func (a *LogrusAdapter) Error(args ...any)                 { a.entry().Error(args...) }
func (a *LogrusAdapter) Errorf(format string, args ...any) { a.entry().Errorf(format, args...) }
func (a *LogrusAdapter) Fatal(args ...any)                 { a.entry().Fatal(args...) }
func (a *LogrusAdapter) Fatalf(format string, args ...any) { a.entry().Fatalf(format, args...) }

// WithFields 返回一个携带附加字段的新 Logger 实例，不修改原实例。
func (a *LogrusAdapter) WithFields(fields map[string]any) Logger {
	merged := make(logrus.Fields, len(a.fields)+len(fields))
	for k, v := range a.fields {
		merged[k] = v
	}
	for k, v := range fields {
		merged[k] = v
	}
	return &LogrusAdapter{L: a.L, fields: merged}
}
