package app

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	rotatelogs "github.com/lestrrat-go/file-rotatelogs"
	"github.com/rifflock/lfshook"
	"github.com/sirupsen/logrus"
)

// Logger 是进程级的全局 Logger 单例。
// 通过 InitLogger 或 Bootstrap 填充；调用方统一使用 GetLogger 访问。
var Logger *logrus.Logger

// GetLogger 返回全局 Logger。若尚未初始化，返回一个基础 logrus 实例作为兜底，
// 保证调用方永远不会拿到 nil（便于在早期 init 阶段写日志）。
func GetLogger() *logrus.Logger {
	if Logger == nil {
		return logrus.StandardLogger()
	}
	return Logger
}

// InitLogger 根据已加载的 LogConfig 初始化全局 Logger，包括：
//   - 级别与格式（json/text）
//   - 按天轮转的文件输出（通过 lfshook + file-rotatelogs）
//
// 目录创建或轮转器构造失败时，退化为仅输出到 stderr，不阻塞启动。
func InitLogger() {
	Logger = logrus.New()

	level, err := logrus.ParseLevel(Config.Log.Default.Level)
	if err != nil {
		level = logrus.InfoLevel
	}
	Logger.SetLevel(level)

	Logger.SetFormatter(newFormatter(Config.Log.Default.Format))

	writer, err := newRotator()
	if err != nil {
		fmt.Printf("[logger] 初始化文件轮转器失败: %v\n", err)
		return
	}

	Logger.AddHook(lfshook.NewHook(
		allLevelWriterMap(writer),
		Logger.Formatter,
	))
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

// newRotator 创建按天切分的日志文件 Writer，带软链指向最新文件。
func newRotator() (*rotatelogs.RotateLogs, error) {
	logPath := Config.Log.File.Path
	if err := os.MkdirAll(logPath, 0o755); err != nil {
		return nil, fmt.Errorf("创建日志目录失败: %w", err)
	}

	base := filepath.Join(logPath, Config.Log.File.Filename)
	return rotatelogs.New(
		base+".%Y%m%d.log",
		rotatelogs.WithLinkName(base+".log"),
		rotatelogs.WithMaxAge(time.Duration(Config.Log.File.MaxAge)*24*time.Hour),
		rotatelogs.WithRotationTime(time.Duration(Config.Log.File.RotationTime)*time.Hour),
	)
}

// allLevelWriterMap 把同一个 writer 绑定到 logrus 全部日志级别，
// 避免调用方手工维护级别映射表。
func allLevelWriterMap(w *rotatelogs.RotateLogs) lfshook.WriterMap {
	return lfshook.WriterMap{
		logrus.DebugLevel: w,
		logrus.InfoLevel:  w,
		logrus.WarnLevel:  w,
		logrus.ErrorLevel: w,
		logrus.FatalLevel: w,
		logrus.PanicLevel: w,
	}
}
