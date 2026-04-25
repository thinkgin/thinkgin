package app

import (
	"fmt"
	"log/slog"
	"os"
)

// SlogAdapter 将 *slog.Logger 包装为 Logger 接口。
// 适用于 Go 1.21+ 项目，用标准库替代第三方日志库。
type SlogAdapter struct {
	L *slog.Logger
}

// NewSlogAdapter 从 *slog.Logger 创建适配器。
func NewSlogAdapter(l *slog.Logger) *SlogAdapter {
	return &SlogAdapter{L: l}
}

func (a *SlogAdapter) Debug(args ...any)                 { a.L.Debug(fmt.Sprint(args...)) }
func (a *SlogAdapter) Debugf(format string, args ...any) { a.L.Debug(fmt.Sprintf(format, args...)) }
func (a *SlogAdapter) Info(args ...any)                  { a.L.Info(fmt.Sprint(args...)) }
func (a *SlogAdapter) Infof(format string, args ...any)  { a.L.Info(fmt.Sprintf(format, args...)) }
func (a *SlogAdapter) Warn(args ...any)                  { a.L.Warn(fmt.Sprint(args...)) }
func (a *SlogAdapter) Warnf(format string, args ...any)  { a.L.Warn(fmt.Sprintf(format, args...)) }
func (a *SlogAdapter) Error(args ...any)                 { a.L.Error(fmt.Sprint(args...)) }
func (a *SlogAdapter) Errorf(format string, args ...any) { a.L.Error(fmt.Sprintf(format, args...)) }

// Fatal 记录 Error 级别后 os.Exit(1)，与 logrus.Fatal 行为一致。
func (a *SlogAdapter) Fatal(args ...any) {
	a.L.Error(fmt.Sprint(args...))
	os.Exit(1)
}

// Fatalf 记录 Error 级别后 os.Exit(1)。
func (a *SlogAdapter) Fatalf(format string, args ...any) {
	a.L.Error(fmt.Sprintf(format, args...))
	os.Exit(1)
}

// WithFields 返回一个携带附加字段的新 Logger 实例。
func (a *SlogAdapter) WithFields(fields map[string]any) Logger {
	attrs := make([]any, 0, len(fields)*2)
	for k, v := range fields {
		attrs = append(attrs, slog.Any(k, v))
	}
	return &SlogAdapter{L: a.L.With(attrs...)}
}
