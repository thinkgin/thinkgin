// Package app 承载全局配置与日志器。
// 本文件只负责声明 YAML 结构体，不包含任何业务逻辑。
package app

// GlobalConfig 是 13 份 YAML 反序列化后的聚合根。
// 每个子配置对应 config/<name>.yaml 文件。
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

// AppConfig 对应 config/app.yaml。
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

// ServerConfig 对应 config/server.yaml。
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

// DatabaseConfig 对应 config/database.yaml。
// NOTE: 当前仅解析，未在 app/ 代码中实际建立连接池。
type DatabaseConfig struct {
	Default     string                 `yaml:"default"`
	Connections map[string]interface{} `yaml:"connections"`
}

// CacheConfig 对应 config/cache.yaml。
// NOTE: 当前仅解析，未建立 Redis/Memcached 客户端。
type CacheConfig struct {
	Default string                 `yaml:"default"`
	Prefix  string                 `yaml:"prefix"`
	Stores  map[string]interface{} `yaml:"stores"`
	Tags    struct {
		Enabled   bool   `yaml:"enabled"`
		Separator string `yaml:"separator"`
	} `yaml:"tags"`
}

// LogConfig 对应 config/log.yaml。
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

// SessionConfig 对应 config/session.yaml。
// 运行时由 app/session 包消费，支持 memory / redis driver；file / database 待实现。
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

// MiddlewareConfig 对应 config/middleware.yaml。
type MiddlewareConfig struct {
	Global []string               `yaml:"global"`
	Groups map[string][]string    `yaml:"groups"`
	Config map[string]interface{} `yaml:"config"`
}

// RouteConfig 对应 config/route.yaml。
// NOTE: URL.Cache/Prefix/Domain 等高级能力暂未消费。
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

// ViewConfig 对应 config/view.yaml。NOTE: 模板加载目前硬编码为 app/*/view/*。
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

// FilesystemConfig 对应 config/filesystem.yaml。NOTE: 暂未接入运行时。
type FilesystemConfig struct {
	Default string                 `yaml:"default"`
	Disks   map[string]interface{} `yaml:"disks"`
	Upload  struct {
		MaxSize      int                 `yaml:"max_size"`
		AllowedTypes map[string][]string `yaml:"allowed_types"`
	} `yaml:"upload"`
}

// LangConfig 对应 config/lang.yaml。NOTE: 暂未接入 i18n 运行时。
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

// TraceConfig 对应 config/trace.yaml。
// NOTE: Jaeger / Zipkin / Otel 字段为前瞻占位，当前 trace.go 仅实现 stdout 导出。
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

// PrometheusConfig 对应 config/prometheus.yaml。
// NOTE: Auth 字段为占位，当前 /metrics 端点未强制鉴权。
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

// PrometheusMetricConfig 控制哪些指标被注册与暴露。
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
