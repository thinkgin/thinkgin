package app

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
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

	// 配置文件列表
	configFiles := []string{
		"app.yaml",
		"server.yaml",
		"database.yaml",
		"cache.yaml",
		"log.yaml",
		"session.yaml",
		"middleware.yaml",
		"route.yaml",
		"view.yaml",
		"filesystem.yaml",
		"lang.yaml",
		"trace.yaml",
	}

	// 逐个加载配置文件
	for _, configFile := range configFiles {
		configPath := filepath.Join(configDir, configFile)
		if err := loadConfigFile(configPath); err != nil {
			fmt.Printf("加载配置文件 %s 失败: %v\n", configFile, err)
			// 继续加载其他配置文件
		}
	}

	// 设置默认配置
	setDefaultConfig()
}

// 加载单个配置文件
func loadConfigFile(configPath string) error {
	// 检查文件是否存在
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		return fmt.Errorf("配置文件不存在: %s", configPath)
	}

	// 读取文件内容
	data, err := ioutil.ReadFile(configPath)
	if err != nil {
		return fmt.Errorf("读取配置文件失败: %v", err)
	}

	// 解析YAML并合并到全局配置
	var tempConfig GlobalConfig
	if err := yaml.Unmarshal(data, &tempConfig); err != nil {
		return fmt.Errorf("解析YAML失败: %v", err)
	}

	// 合并配置（这里可以根据需要实现更复杂的合并逻辑）
	mergeConfig(&tempConfig)

	return nil
}

// 合并配置
func mergeConfig(tempConfig *GlobalConfig) {
	// 使用反射或手动合并配置
	// 这里简化处理，直接覆盖非零值
	if tempConfig.App.Name != "" {
		Config.App = tempConfig.App
	}
	if tempConfig.Server.HTTP.Port != 0 {
		Config.Server = tempConfig.Server
	}
	if tempConfig.Database.Default != "" {
		Config.Database = tempConfig.Database
	}
	if tempConfig.Cache.Default != "" {
		Config.Cache = tempConfig.Cache
	}
	if tempConfig.Log.Default.Driver != "" {
		Config.Log = tempConfig.Log
	}
	if tempConfig.Session.Driver != "" {
		Config.Session = tempConfig.Session
	}
	if len(tempConfig.Middleware.Global) > 0 {
		Config.Middleware = tempConfig.Middleware
	}
	if tempConfig.Route.URL.CacheFile != "" {
		Config.Route = tempConfig.Route
	}
	if tempConfig.View.Engine != "" {
		Config.View = tempConfig.View
	}
	if tempConfig.Filesystem.Default != "" {
		Config.Filesystem = tempConfig.Filesystem
	}
	if tempConfig.Lang.Default != "" {
		Config.Lang = tempConfig.Lang
	}
	if tempConfig.Trace.ServiceName != "" {
		Config.Trace = tempConfig.Trace
	}
}

// 设置默认配置
func setDefaultConfig() {
	// 设置默认值
	if Config.App.Name == "" {
		Config.App.Name = "ThinkGin"
	}
	if Config.App.Version == "" {
		Config.App.Version = "2.0.1"
	}
	if Config.Server.HTTP.Port == 0 {
		Config.Server.HTTP.Port = 8000
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
