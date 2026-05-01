package app

import (
	"fmt"
	"io"
	"log/slog"
	"os"
	"path/filepath"

	"gopkg.in/natefinch/lumberjack.v2"
)

// Log 是进程级的全局 Logger 单例（Logger 接口）。
// 通过 InitLogger 或 Bootstrap 填充；调用方统一使用 GetLogger 访问。
//
// Deprecated: 新代码应通过 ServiceContext.Log 获取 Logger，而非直接访问全局变量。
var Log Logger

// GetLogger 返回全局 Logger 接口。若尚未初始化，返回基于 slog 标准实例的兜底，
// 保证调用方永远不会拿到 nil（便于在早期 init 阶段写日志）。
func GetLogger() Logger {
	if Log == nil {
		return NewSlogAdapter(slog.Default())
	}
	return Log
}

// InitLogger 根据已加载的 LogConfig 初始化全局 Logger，包括：
//   - 级别与格式（json/text）
//   - 基于 lumberjack 的日志文件按大小轮转与保留策略
//
// 默认使用标准库 slog 作为日志引擎，无需第三方依赖。
// 如需 logrus，可在应用启动后手动设置：app.Log = app.NewLogrusAdapter(...)
//
// 目录创建失败时，退化为仅输出到 stderr，不阻塞启动。
func InitLogger() {
	level := parseSlogLevel(Config.Log.Default.Level)

	writer, err := newRotator()
	if err != nil {
		fmt.Printf("[logger] 初始化日志文件失败: %v\n", err)
		Log = NewSlogAdapter(slog.New(newSlogHandler(os.Stderr, level, Config.Log.Default.Format)))
		return
	}

	// 同时输出到 stderr 和日志文件，便于容器环境下 stdout 采集。
	output := io.MultiWriter(os.Stderr, writer)
	Log = NewSlogAdapter(slog.New(newSlogHandler(output, level, Config.Log.Default.Format)))
}

// newSlogHandler 根据格式创建 slog.Handler。
func newSlogHandler(w io.Writer, level slog.Level, format string) slog.Handler {
	opts := &slog.HandlerOptions{Level: level}
	if format == "json" {
		return slog.NewJSONHandler(w, opts)
	}
	return slog.NewTextHandler(w, opts)
}

// parseSlogLevel 将配置字符串映射为 slog.Level。
func parseSlogLevel(s string) slog.Level {
	switch s {
	case "debug", "trace":
		return slog.LevelDebug
	case "warn", "warning":
		return slog.LevelWarn
	case "error", "fatal", "panic":
		return slog.LevelError
	default:
		return slog.LevelInfo
	}
}

// newRotator 创建基于 lumberjack 的日志文件 Writer。
// lumberjack 自动按文件大小轮转、保留指定天数、压缩旧文件，且跨平台稳定。
func newRotator() (*lumberjack.Logger, error) {
	logPath := Config.Log.File.Path
	if err := os.MkdirAll(logPath, 0o755); err != nil {
		return nil, fmt.Errorf("创建日志目录失败: %w", err)
	}

	filename := filepath.Join(logPath, Config.Log.File.Filename+".log")
	return &lumberjack.Logger{
		Filename:   filename,
		MaxSize:    100, // MB，单文件最大 100MB
		MaxAge:     Config.Log.File.MaxAge,
		MaxBackups: 30,
		LocalTime:  true,
		Compress:   true,
	}, nil
}
