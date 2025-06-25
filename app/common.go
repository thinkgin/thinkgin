package app

import (
	"fmt"
	"os"
	"path"
	"strings"
	"time"

	"github.com/sirupsen/logrus"
	"gopkg.in/ini.v1"

	rotatelogs "github.com/lestrrat-go/file-rotatelogs"
	"github.com/rifflock/lfshook"
)

type Config struct {
	LogFilePath     string
	LogFileName     string
	LogLevel        string
	LogFormat       string
	LogMaxAge       int
	LogRotationTime int
}

var AppConfig *Config
var Logger *logrus.Logger

// 初始化配置
func init() {
	AppConfig = &Config{}
	LoadConfig()
	InitLogger()
}

// 加载配置文件
func LoadConfig() {
	cfg, err := ini.Load("config.ini")
	if err != nil {
		fmt.Printf("加载配置文件失败: %v\n", err)
		// 使用默认配置
		setDefaultConfig()
		return
	}

	// 读取日志配置
	logSection := cfg.Section("log")
	AppConfig.LogFilePath = logSection.Key("LogFilePath").MustString("runtime/log")
	AppConfig.LogFileName = logSection.Key("LogFileName").MustString("system")
	AppConfig.LogLevel = logSection.Key("LogLevel").MustString("info")
	AppConfig.LogFormat = logSection.Key("LogFormat").MustString("json")
	AppConfig.LogMaxAge = logSection.Key("LogMaxAge").MustInt(7)
	AppConfig.LogRotationTime = logSection.Key("LogRotationTime").MustInt(24)
}

// 设置默认配置
func setDefaultConfig() {
	AppConfig.LogFilePath = "runtime/log"
	AppConfig.LogFileName = "system"
	AppConfig.LogLevel = "info"
	AppConfig.LogFormat = "json"
	AppConfig.LogMaxAge = 7
	AppConfig.LogRotationTime = 24
}

// 初始化日志记录器
func InitLogger() {
	Logger = logrus.New()

	// 创建日志目录
	if err := os.MkdirAll(AppConfig.LogFilePath, 0755); err != nil {
		fmt.Printf("创建日志目录失败: %v\n", err)
	}

	// 设置日志级别
	level, err := logrus.ParseLevel(AppConfig.LogLevel)
	if err != nil {
		level = logrus.InfoLevel
		fmt.Printf("无效的日志级别，使用默认级别: info\n")
	}
	Logger.SetLevel(level)

	// 设置日志格式
	switch strings.ToLower(AppConfig.LogFormat) {
	case "json":
		Logger.SetFormatter(&logrus.JSONFormatter{
			TimestampFormat: "2006-01-02 15:04:05",
		})
	case "text":
		Logger.SetFormatter(&logrus.TextFormatter{
			TimestampFormat: "2006-01-02 15:04:05",
			FullTimestamp:   true,
		})
	default:
		Logger.SetFormatter(&logrus.JSONFormatter{
			TimestampFormat: "2006-01-02 15:04:05",
		})
	}

	// 设置日志文件轮转
	fileName := path.Join(AppConfig.LogFilePath, AppConfig.LogFileName)
	logWriter, err := rotatelogs.New(
		fileName+".%Y%m%d.log",
		rotatelogs.WithLinkName(fileName+".log"),
		rotatelogs.WithMaxAge(time.Duration(AppConfig.LogMaxAge)*24*time.Hour),
		rotatelogs.WithRotationTime(time.Duration(AppConfig.LogRotationTime)*time.Hour),
	)

	if err != nil {
		fmt.Printf("日志轮转设置失败: %v\n", err)
		return
	}

	// 设置Hook
	writeMap := lfshook.WriterMap{
		logrus.DebugLevel: logWriter,
		logrus.InfoLevel:  logWriter,
		logrus.WarnLevel:  logWriter,
		logrus.ErrorLevel: logWriter,
		logrus.FatalLevel: logWriter,
		logrus.PanicLevel: logWriter,
	}

	lfHook := lfshook.NewHook(writeMap, Logger.Formatter)
	Logger.AddHook(lfHook)

	// 在开发模式下也输出到控制台
	Logger.SetOutput(os.Stdout)
}

// 获取全局Logger实例
func GetLogger() *logrus.Logger {
	return Logger
}
