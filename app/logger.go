package app

import (
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/sirupsen/logrus"
	"gopkg.in/natefinch/lumberjack.v2"
)

// Log 是进程级的全局 Logger 单例（Logger 接口）。
// 通过 InitLogger 或 Bootstrap 填充；调用方统一使用 GetLogger 访问。
var Log Logger

// GetLogger 返回全局 Logger 接口。若尚未初始化，返回基于 logrus 标准实例的兜底，
// 保证调用方永远不会拿到 nil（便于在早期 init 阶段写日志）。
func GetLogger() Logger {
	if Log == nil {
		return NewLogrusAdapter(logrus.StandardLogger())
	}
	return Log
}

// InitLogger 根据已加载的 LogConfig 初始化全局 Logger，包括：
//   - 级别与格式（json/text）
//   - 基于 lumberjack 的日志文件按大小轮转与保留策略
//
// 目录创建失败时，退化为仅输出到 stderr，不阻塞启动。
func InitLogger() {
	l := logrus.New()

	level, err := logrus.ParseLevel(Config.Log.Default.Level)
	if err != nil {
		level = logrus.InfoLevel
	}
	l.SetLevel(level)

	l.SetFormatter(newFormatter(Config.Log.Default.Format))

	writer, err := newRotator()
	if err != nil {
		fmt.Printf("[logger] 初始化日志文件失败: %v\n", err)
		Log = NewLogrusAdapter(l)
		return
	}

	// 同时输出到 stderr 和日志文件，便于容器环境下 stdout 采集。
	l.SetOutput(io.MultiWriter(os.Stderr, writer))
	Log = NewLogrusAdapter(l)
}

// newFormatter 按格式名返回对应的 logrus Formatter。
// 非 json 一律视为 text，保持与 YAML 校验一致的语义。
func newFormatter(format string) logrus.Formatter {
	const ts = "2006-01-02 15:04:05"
	if format == "json" {
		return &logrus.JSONFormatter{TimestampFormat: ts}
	}
	return &logrus.TextFormatter{TimestampFormat: ts}
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
