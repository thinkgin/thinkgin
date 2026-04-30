package app

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

// defaultConfigDir 是相对工作目录的默认配置根。
const defaultConfigDir = "config"

// LoadConfigFromDir 读取指定目录下的 13 份 YAML 并填充到全局 Config。
// 单个文件加载失败会记录错误但不中断整体流程，使用默认值兜底。
// 返回聚合错误（非空时表示部分文件加载失败，但配置仍可使用）。
func LoadConfigFromDir(dir string) error {
	if Config == nil {
		Config = &GlobalConfig{}
	}

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

	var errs []error
	for _, l := range loaders {
		path := filepath.Join(dir, l.file)
		if err := l.fn(path); err != nil {
			errs = append(errs, fmt.Errorf("%s: %w", l.file, err))
		}
	}

	return errors.Join(errs...)
}

// LoadConfig 为了向后兼容保留，使用默认目录。
func LoadConfig() {
	_ = LoadConfigFromDir(defaultConfigDir)
}

// loadNewConfigFromDir 构建一个全新的 GlobalConfig 对象并加载指定目录的 YAML。
// 与 LoadConfigFromDir 不同，本函数不修改全局变量，适用于热更新的"构建-替换"模式。
func loadNewConfigFromDir(dir string) (*GlobalConfig, error) {
	cfg := &GlobalConfig{}

	loaders := []struct {
		file string
		fn   func(string) error
	}{
		{"app.yaml", func(p string) error { return loadInto("app", p, &cfg.App) }},
		{"server.yaml", func(p string) error { return loadInto("server", p, &cfg.Server) }},
		{"database.yaml", func(p string) error { return loadInto("database", p, &cfg.Database) }},
		{"cache.yaml", func(p string) error { return loadInto("cache", p, &cfg.Cache) }},
		{"log.yaml", func(p string) error { return loadInto("log", p, &cfg.Log) }},
		{"session.yaml", func(p string) error { return loadInto("session", p, &cfg.Session) }},
		{"middleware.yaml", func(p string) error { return loadInto("middleware", p, &cfg.Middleware) }},
		{"route.yaml", func(p string) error { return loadInto("route", p, &cfg.Route) }},
		{"view.yaml", func(p string) error { return loadInto("view", p, &cfg.View) }},
		{"filesystem.yaml", func(p string) error { return loadInto("filesystem", p, &cfg.Filesystem) }},
		{"lang.yaml", func(p string) error { return loadInto("lang", p, &cfg.Lang) }},
		{"trace.yaml", func(p string) error { return loadInto("trace", p, &cfg.Trace) }},
		{"prometheus.yaml", func(p string) error { return loadInto("prometheus", p, &cfg.Prometheus) }},
	}

	var errs []error
	for _, l := range loaders {
		path := filepath.Join(dir, l.file)
		if err := l.fn(path); err != nil {
			errs = append(errs, fmt.Errorf("%s: %w", l.file, err))
		}
	}

	return cfg, errors.Join(errs...)
}

// loadInto 把 YAML 文件反序列化到 *T。
// 约定文件以 rootKey 作为顶层单键（例如 app.yaml 的顶层是 `app:`）。
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
