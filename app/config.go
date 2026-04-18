package app

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/sirupsen/logrus"
	"gopkg.in/yaml.v3"

	rotatelogs "github.com/lestrrat-go/file-rotatelogs"
	"github.com/rifflock/lfshook"
)

// 全局配置结构
type GlobalConfig struct {
	App        AppConfig        `yaml:"app"`
	Server     ServerConfig     `yaml:"server"`
	Database   DatabaseConfig   `yaml:"database"`
	Cache      CacheConfig      `yaml:"cache"`
	Log        LogConfig        `yaml:"log"`
	Session    SessionConfig    `yaml:"session"`
	Middleware MiddlewareConfig `yaml:"middleware"`
	Route      RouteConfig      `yaml:"route"`
	View       ViewConfig       `yaml:"view"`
	Filesystem FilesystemConfig `yaml:"filesystem"`
	Lang       LangConfig       `yaml:"lang"`
	Trace      TraceConfig      `yaml:"trace"`
	Prometheus PrometheusConfig `yaml:"prometheus"`
}

// 应用配置
type AppConfig struct {
	Name     string `yaml:"name"`
	Version  string `yaml:"version"`
	Debug    bool   `yaml:"debug"`
	Timezone string `yaml:"timezone"`
	Locale   string `yaml:"locale"`
	JWT      struct {
		Secret string `yaml:"secret"`
		Expire int    `yaml:"expire"`
	} `yaml:"jwt"`
	Pagination struct {
		PageSize    int `yaml:"page_size"`
		MaxPageSize int `yaml:"max_page_size"`
	} `yaml:"pagination"`
	Monitoring struct {
		PrometheusEnabled bool `yaml:"prometheus_enabled"`
	} `yaml:"monitoring"`
}

// 服务器配置
type ServerConfig struct {
	HTTP struct {
		Host           string `yaml:"host"`
		Port           int    `yaml:"port"`
		ReadTimeout    int    `yaml:"read_timeout"`
		WriteTimeout   int    `yaml:"write_timeout"`
		IdleTimeout    int    `yaml:"idle_timeout"`
		MaxHeaderBytes int    `yaml:"max_header_bytes"`
	} `yaml:"http"`
	HTTPS struct {
		Enabled  bool   `yaml:"enabled"`
		Port     int    `yaml:"port"`
		CertFile string `yaml:"cert_file"`
		KeyFile  string `yaml:"key_file"`
	} `yaml:"https"`
	Mode   string `yaml:"mode"`
	Static struct {
		Path string `yaml:"path"`
		Root string `yaml:"root"`
	} `yaml:"static"`
	Upload struct {
		MaxSize      int      `yaml:"max_size"`
		AllowedTypes []string `yaml:"allowed_types"`
		Path         string   `yaml:"path"`
	} `yaml:"upload"`
}

// 数据库配置
type DatabaseConfig struct {
	Default     string                 `yaml:"default"`
	Connections map[string]interface{} `yaml:"connections"`
}

// 缓存配置
type CacheConfig struct {
	Default string                 `yaml:"default"`
	Prefix  string                 `yaml:"prefix"`
	Stores  map[string]interface{} `yaml:"stores"`
	Tags    struct {
		Enabled   bool   `yaml:"enabled"`
		Separator string `yaml:"separator"`
	} `yaml:"tags"`
}

// 日志配置
type LogConfig struct {
	Default struct {
		Driver string `yaml:"driver"`
		Level  string `yaml:"level"`
		Format string `yaml:"format"`
	} `yaml:"default"`
	File struct {
		Path         string `yaml:"path"`
		Filename     string `yaml:"filename"`
		MaxAge       int    `yaml:"max_age"`
		RotationTime int    `yaml:"rotation_time"`
		MaxSize      int    `yaml:"max_size"`
		Compress     bool   `yaml:"compress"`
	} `yaml:"file"`
	Console struct {
		Color bool `yaml:"color"`
	} `yaml:"console"`
	Channels map[string]interface{} `yaml:"channels"`
}

// Session配置
type SessionConfig struct {
	Driver        string `yaml:"driver"`
	Name          string `yaml:"name"`
	Lifetime      int    `yaml:"lifetime"`
	ExpireOnClose bool   `yaml:"expire_on_close"`
	Encrypt       bool   `yaml:"encrypt"`
	File          struct {
		Path string `yaml:"path"`
	} `yaml:"file"`
	Redis struct {
		Connection string `yaml:"connection"`
	} `yaml:"redis"`
	Database struct {
		Table      string `yaml:"table"`
		Connection string `yaml:"connection"`
	} `yaml:"database"`
	Cookie struct {
		Path     string `yaml:"path"`
		Domain   string `yaml:"domain"`
		Secure   bool   `yaml:"secure"`
		HttpOnly bool   `yaml:"http_only"`
		SameSite string `yaml:"same_site"`
	} `yaml:"cookie"`
}

// 中间件配置
type MiddlewareConfig struct {
	Global []string               `yaml:"global"`
	Groups map[string][]string    `yaml:"groups"`
	Config map[string]interface{} `yaml:"config"`
}

// 路由配置
type RouteConfig struct {
	URL struct {
		Cache     bool   `yaml:"cache"`
		CacheFile string `yaml:"cache_file"`
	} `yaml:"url"`
	Prefix     map[string]string `yaml:"prefix"`
	Domain     map[string]string `yaml:"domain"`
	Parameters struct {
		Patterns map[string]string `yaml:"patterns"`
	} `yaml:"parameters"`
	Groups map[string][]string `yaml:"groups"`
}

// 视图配置
type ViewConfig struct {
	Engine    string   `yaml:"engine"`
	Paths     []string `yaml:"paths"`
	Extension string   `yaml:"extension"`
	Compile   struct {
		Cache     bool   `yaml:"cache"`
		CachePath string `yaml:"cache_path"`
	} `yaml:"compile"`
	Assets struct {
		Version string `yaml:"version"`
		CssPath string `yaml:"css_path"`
		JsPath  string `yaml:"js_path"`
		ImgPath string `yaml:"img_path"`
	} `yaml:"assets"`
	Globals map[string]interface{} `yaml:"globals"`
	Helpers struct {
		Builtin bool `yaml:"builtin"`
	} `yaml:"helpers"`
}

// 文件系统配置
type FilesystemConfig struct {
	Default string                 `yaml:"default"`
	Disks   map[string]interface{} `yaml:"disks"`
	Upload  struct {
		MaxSize      int                 `yaml:"max_size"`
		AllowedTypes map[string][]string `yaml:"allowed_types"`
	} `yaml:"upload"`
}

// 多语言配置
type LangConfig struct {
	Default   string   `yaml:"default"`
	Supported []string `yaml:"supported"`
	Path      string   `yaml:"path"`
	Detection struct {
		Order    []string `yaml:"order"`
		UrlParam string   `yaml:"url_param"`
		Cookie   struct {
			Name   string `yaml:"name"`
			Expire int    `yaml:"expire"`
		} `yaml:"cookie"`
	} `yaml:"detection"`
	Fallback struct {
		Enabled  bool   `yaml:"enabled"`
		Language string `yaml:"language"`
	} `yaml:"fallback"`
}

// 链路追踪配置
type TraceConfig struct {
	Enabled     bool    `yaml:"enabled"`
	Driver      string  `yaml:"driver"`
	ServiceName string  `yaml:"service_name"`
	SampleRate  float64 `yaml:"sample_rate"`
	Jaeger      struct {
		AgentHost    string `yaml:"agent_host"`
		AgentPort    int    `yaml:"agent_port"`
		CollectorUrl string `yaml:"collector_url"`
		Username     string `yaml:"username"`
		Password     string `yaml:"password"`
	} `yaml:"jaeger"`
	Zipkin struct {
		Endpoint string `yaml:"endpoint"`
	} `yaml:"zipkin"`
	Otel struct {
		Endpoint string `yaml:"endpoint"`
		UseGrpc  bool   `yaml:"use_grpc"`
	} `yaml:"otel"`
	Tags       map[string]string `yaml:"tags"`
	Operations []string          `yaml:"operations"`
}

// Prometheus监控配置
type PrometheusConfig struct {
	Enabled        bool                   `yaml:"enabled"`
	Path           string                 `yaml:"path"`
	Port           int                    `yaml:"port"`
	ServiceName    string                 `yaml:"service_name"`
	Namespace      string                 `yaml:"namespace"`
	Labels         map[string]string      `yaml:"labels"`
	Metrics        PrometheusMetricConfig `yaml:"metrics"`
	ScrapeInterval int                    `yaml:"scrape_interval"`
	Auth           struct {
		Enabled  bool   `yaml:"enabled"`
		Username string `yaml:"username"`
		Password string `yaml:"password"`
	} `yaml:"auth"`
}

// Prometheus指标配置
type PrometheusMetricConfig struct {
	HTTP struct {
		Enabled         bool   `yaml:"enabled"`
		RequestsTotal   string `yaml:"requests_total"`
		RequestDuration string `yaml:"request_duration"`
		RequestSize     string `yaml:"request_size"`
		ResponseSize    string `yaml:"response_size"`
		IncludePath     bool   `yaml:"include_path"`
	} `yaml:"http"`
	System struct {
		Enabled          bool   `yaml:"enabled"`
		CPUUsage         string `yaml:"cpu_usage"`
		MemoryUsage      string `yaml:"memory_usage"`
		ProcessStartTime string `yaml:"process_start_time"`
	} `yaml:"system"`
	Business struct {
		Enabled         bool   `yaml:"enabled"`
		CounterPrefix   string `yaml:"counter_prefix"`
		HistogramPrefix string `yaml:"histogram_prefix"`
		GaugePrefix     string `yaml:"gauge_prefix"`
	} `yaml:"business"`
}

var Config *GlobalConfig
var Logger *logrus.Logger

// 初始化配置和日志
func init() {
	Config = &GlobalConfig{}
	LoadConfig()
	InitLogger()
}

// 加载所有配置文件
func LoadConfig() {
	configDir := "config"

	loaders := []struct {
		file string
		fn   func(string) error
	}{
		{"app.yaml", func(p string) error { return loadInto("app", p, &Config.App) }},
		{"server.yaml", func(p string) error { return loadInto("server", p, &Config.Server) }},
		{"database.yaml", func(p string) error { return loadInto("database", p, &Config.Database) }},
		{"cache.yaml", func(p string) error { return loadInto("cache", p, &Config.Cache) }},
		{"log.yaml", func(p string) error { return loadInto("log", p, &Config.Log) }},
		{"session.yaml", func(p string) error { return loadInto("session", p, &Config.Session) }},
		{"middleware.yaml", func(p string) error { return loadInto("middleware", p, &Config.Middleware) }},
		{"route.yaml", func(p string) error { return loadInto("route", p, &Config.Route) }},
		{"view.yaml", func(p string) error { return loadInto("view", p, &Config.View) }},
		{"filesystem.yaml", func(p string) error { return loadInto("filesystem", p, &Config.Filesystem) }},
		{"lang.yaml", func(p string) error { return loadInto("lang", p, &Config.Lang) }},
		{"trace.yaml", func(p string) error { return loadInto("trace", p, &Config.Trace) }},
		{"prometheus.yaml", func(p string) error { return loadInto("prometheus", p, &Config.Prometheus) }},
	}

	for _, l := range loaders {
		configPath := filepath.Join(configDir, l.file)
		if err := l.fn(configPath); err != nil {
			fmt.Printf("加载配置文件 %s 失败: %v\n", l.file, err)
		}
	}

	setDefaultConfig()
	applyEnvOverrides()
	validateConfig()
}

func loadInto[T any](rootKey string, configPath string, dst *T) error {
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		return fmt.Errorf("配置文件不存在: %s", configPath)
	}

	data, err := os.ReadFile(configPath)
	if err != nil {
		return fmt.Errorf("读取配置文件失败: %v", err)
	}
	data = bytes.TrimSpace(data)
	if len(data) == 0 {
		return fmt.Errorf("配置文件为空: %s", configPath)
	}

	var m map[string]T
	if err := yaml.Unmarshal(data, &m); err != nil {
		return fmt.Errorf("解析YAML失败: %v", err)
	}

	v, ok := m[rootKey]
	if !ok {
		return fmt.Errorf("缺少根节点 %s", rootKey)
	}

	*dst = v
	return nil
}

func applyEnvOverrides() {
	const prefix = "THINKGIN_"

	if v := os.Getenv(prefix + "SERVER_MODE"); v != "" {
		Config.Server.Mode = v
	}
	if v := os.Getenv(prefix + "SERVER_HTTP_HOST"); v != "" {
		Config.Server.HTTP.Host = v
	}
	if v := getenvInt(prefix + "SERVER_HTTP_PORT"); v != nil {
		Config.Server.HTTP.Port = *v
	}
	if v := getenvInt(prefix + "SERVER_HTTP_READ_TIMEOUT"); v != nil {
		Config.Server.HTTP.ReadTimeout = *v
	}
	if v := getenvInt(prefix + "SERVER_HTTP_WRITE_TIMEOUT"); v != nil {
		Config.Server.HTTP.WriteTimeout = *v
	}
	if v := getenvInt(prefix + "SERVER_HTTP_IDLE_TIMEOUT"); v != nil {
		Config.Server.HTTP.IdleTimeout = *v
	}
	if v := getenvInt(prefix + "SERVER_HTTP_MAX_HEADER_BYTES"); v != nil {
		Config.Server.HTTP.MaxHeaderBytes = *v
	}

	if v := getenvBool(prefix + "APP_DEBUG"); v != nil {
		Config.App.Debug = *v
	}

	if v := os.Getenv(prefix + "LOG_LEVEL"); v != "" {
		Config.Log.Default.Level = v
	}
	if v := os.Getenv(prefix + "LOG_FORMAT"); v != "" {
		Config.Log.Default.Format = v
	}
	if v := os.Getenv(prefix + "LOG_PATH"); v != "" {
		Config.Log.File.Path = v
	}
	if v := os.Getenv(prefix + "LOG_FILENAME"); v != "" {
		Config.Log.File.Filename = v
	}

	if v := getenvBool(prefix + "APP_PROMETHEUS_ENABLED"); v != nil {
		Config.App.Monitoring.PrometheusEnabled = *v
	}
	if v := getenvBool(prefix + "PROMETHEUS_ENABLED"); v != nil {
		Config.Prometheus.Enabled = *v
	}
	if v := os.Getenv(prefix + "PROMETHEUS_PATH"); v != "" {
		Config.Prometheus.Path = v
	}
}

func validateConfig() {
	mode := strings.ToLower(strings.TrimSpace(Config.Server.Mode))
	if mode == "" {
		Config.Server.Mode = "debug"
	} else {
		switch mode {
		case "debug", "test", "release":
			Config.Server.Mode = mode
		default:
			fmt.Printf("无效的 server.mode: %s\n", Config.Server.Mode)
			Config.Server.Mode = "debug"
		}
	}

	if Config.Server.HTTP.Host == "" {
		Config.Server.HTTP.Host = "0.0.0.0"
	}
	if Config.Server.HTTP.Port <= 0 || Config.Server.HTTP.Port > 65535 {
		fmt.Printf("无效的 server.http.port: %d\n", Config.Server.HTTP.Port)
		Config.Server.HTTP.Port = 8000
	}

	if _, err := logrus.ParseLevel(strings.ToLower(strings.TrimSpace(Config.Log.Default.Level))); err != nil {
		fmt.Printf("无效的 log.default.level: %s\n", Config.Log.Default.Level)
		Config.Log.Default.Level = "info"
	}

	format := strings.ToLower(strings.TrimSpace(Config.Log.Default.Format))
	if format == "" {
		Config.Log.Default.Format = "json"
	} else {
		switch format {
		case "json", "text":
			Config.Log.Default.Format = format
		default:
			fmt.Printf("无效的 log.default.format: %s\n", Config.Log.Default.Format)
			Config.Log.Default.Format = "json"
		}
	}

	if Config.Log.File.Path == "" {
		Config.Log.File.Path = "runtime/log"
	}
	if Config.Log.File.Filename == "" {
		Config.Log.File.Filename = "system"
	}
}

func getenvInt(key string) *int {
	v := strings.TrimSpace(os.Getenv(key))
	if v == "" {
		return nil
	}
	n, err := strconv.Atoi(v)
	if err != nil {
		return nil
	}
	return &n
}

func getenvBool(key string) *bool {
	v := strings.TrimSpace(os.Getenv(key))
	if v == "" {
		return nil
	}
	b, err := strconv.ParseBool(v)
	if err != nil {
		return nil
	}
	return &b
}

// 设置默认配置
func setDefaultConfig() {
	// 设置默认值
	if Config.App.Name == "" {
		Config.App.Name = "ThinkGin"
	}
	if Config.App.Version == "" {
		Config.App.Version = "3.0.0"
	}
	if Config.Server.HTTP.Host == "" {
		Config.Server.HTTP.Host = "0.0.0.0"
	}
	if Config.Server.HTTP.Port == 0 {
		Config.Server.HTTP.Port = 8000
	}
	if Config.Server.HTTP.ReadTimeout == 0 {
		Config.Server.HTTP.ReadTimeout = 60
	}
	if Config.Server.HTTP.WriteTimeout == 0 {
		Config.Server.HTTP.WriteTimeout = 60
	}
	if Config.Server.HTTP.IdleTimeout == 0 {
		Config.Server.HTTP.IdleTimeout = 120
	}
	if Config.Server.HTTP.MaxHeaderBytes == 0 {
		Config.Server.HTTP.MaxHeaderBytes = 1048576
	}
	if Config.Log.Default.Level == "" {
		Config.Log.Default.Level = "info"
	}
	if Config.Log.Default.Format == "" {
		Config.Log.Default.Format = "json"
	}
	if Config.Log.File.Path == "" {
		Config.Log.File.Path = "runtime/log"
	}
	if Config.Log.File.Filename == "" {
		Config.Log.File.Filename = "system"
	}
	if Config.Log.File.MaxAge == 0 {
		Config.Log.File.MaxAge = 30
	}
	if Config.Log.File.RotationTime == 0 {
		Config.Log.File.RotationTime = 24
	}

	// 设置Prometheus默认配置
	if Config.Prometheus.ServiceName == "" {
		Config.Prometheus.ServiceName = "thinkgin"
	}
	if Config.Prometheus.Path == "" {
		Config.Prometheus.Path = "/metrics"
	}
	if Config.Prometheus.Namespace == "" {
		Config.Prometheus.Namespace = "app"
	}
	if Config.Prometheus.ScrapeInterval == 0 {
		Config.Prometheus.ScrapeInterval = 15
	}
}

// 初始化日志系统
func InitLogger() {
	Logger = logrus.New()

	// 设置日志级别
	level, err := logrus.ParseLevel(Config.Log.Default.Level)
	if err != nil {
		level = logrus.InfoLevel
	}
	Logger.SetLevel(level)

	// 设置日志格式
	if Config.Log.Default.Format == "json" {
		Logger.SetFormatter(&logrus.JSONFormatter{
			TimestampFormat: "2006-01-02 15:04:05",
		})
	} else {
		Logger.SetFormatter(&logrus.TextFormatter{
			TimestampFormat: "2006-01-02 15:04:05",
		})
	}

	// 创建日志目录
	logPath := Config.Log.File.Path
	if err := os.MkdirAll(logPath, 0755); err != nil {
		fmt.Printf("创建日志目录失败: %v\n", err)
		return
	}

	// 配置文件轮转
	logFileName := filepath.Join(logPath, Config.Log.File.Filename)
	writer, err := rotatelogs.New(
		logFileName+".%Y%m%d.log",
		rotatelogs.WithLinkName(logFileName+".log"),
		rotatelogs.WithMaxAge(time.Duration(Config.Log.File.MaxAge)*24*time.Hour),
		rotatelogs.WithRotationTime(time.Duration(Config.Log.File.RotationTime)*time.Hour),
	)

	if err != nil {
		fmt.Printf("创建日志轮转器失败: %v\n", err)
		return
	}

	// 添加文件 Hook
	Logger.AddHook(lfshook.NewHook(
		lfshook.WriterMap{
			logrus.DebugLevel: writer,
			logrus.InfoLevel:  writer,
			logrus.WarnLevel:  writer,
			logrus.ErrorLevel: writer,
			logrus.FatalLevel: writer,
			logrus.PanicLevel: writer,
		},
		Logger.Formatter,
	))
}

// 获取日志实例
func GetLogger() *logrus.Logger {
	return Logger
}

// 获取配置值的便捷方法
func GetConfig() *GlobalConfig {
	return Config
}

// 获取应用配置
func GetAppConfig() *AppConfig {
	return &Config.App
}

// 获取服务器配置
func GetServerConfig() *ServerConfig {
	return &Config.Server
}

// 获取数据库配置
func GetDatabaseConfig() *DatabaseConfig {
	return &Config.Database
}

// 获取日志配置
func GetLogConfig() *LogConfig {
	return &Config.Log
}

// 获取Prometheus配置
func GetPrometheusConfig() *PrometheusConfig {
	return &Config.Prometheus
}
