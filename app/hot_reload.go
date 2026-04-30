// 本文件提供配置热更新能力。
//
// 设计思路：
//   - 使用 fsnotify 监听配置目录下 .yaml 文件的写入/创建事件。
//   - 收到事件后去抖（debounce 500ms），重新加载全部配置。
//   - 重新加载后调用注册的回调函数，业务层可据此刷新运行时状态。
//   - 不影响 Logger（日志通常需要持续写入，不适合热换）。
//
// 使用方式：
//
//	watcher, _ := app.WatchConfig("config", func(cfg *app.GlobalConfig) {
//	    log.Infof("config reloaded: app.name=%s", cfg.App.Name)
//	})
//	defer watcher.Close()
package app

import (
	"fmt"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/fsnotify/fsnotify"
)

// ConfigReloadCallback 在配置热更新后调用的回调。
type ConfigReloadCallback func(cfg *GlobalConfig)

// ConfigWatcher 封装 fsnotify.Watcher 并提供去抖重加载逻辑。
type ConfigWatcher struct {
	watcher   *fsnotify.Watcher
	configDir string
	callbacks []ConfigReloadCallback
	mu        sync.Mutex
	done      chan struct{}
}

// WatchConfig 开始监听 configDir 中 YAML 文件变更。
// 返回的 ConfigWatcher 应在程序退出时调用 Close。
func WatchConfig(configDir string, callbacks ...ConfigReloadCallback) (*ConfigWatcher, error) {
	w, err := fsnotify.NewWatcher()
	if err != nil {
		return nil, fmt.Errorf("hot_reload: create watcher: %w", err)
	}

	if err := w.Add(configDir); err != nil {
		_ = w.Close()
		return nil, fmt.Errorf("hot_reload: watch %s: %w", configDir, err)
	}

	cw := &ConfigWatcher{
		watcher:   w,
		configDir: configDir,
		callbacks: callbacks,
		done:      make(chan struct{}),
	}

	go cw.loop()
	return cw, nil
}

// Close 停止监听。
func (cw *ConfigWatcher) Close() error {
	close(cw.done)
	return cw.watcher.Close()
}

func (cw *ConfigWatcher) loop() {
	// 去抖定时器：收到事件后等 500ms 再真正重载，合并高频连续写入。
	var debounce *time.Timer

	for {
		select {
		case <-cw.done:
			return
		case event, ok := <-cw.watcher.Events:
			if !ok {
				return
			}
			// 只关心 .yaml 文件的写入和创建
			if !isYAMLFile(event.Name) {
				continue
			}
			if !event.Has(fsnotify.Write) && !event.Has(fsnotify.Create) {
				continue
			}

			if debounce != nil {
				debounce.Stop()
			}
			debounce = time.AfterFunc(500*time.Millisecond, func() {
				cw.reload()
			})

		case _, ok := <-cw.watcher.Errors:
			if !ok {
				return
			}
		}
	}
}

func (cw *ConfigWatcher) reload() {
	cw.mu.Lock()
	defer cw.mu.Unlock()

	// 构建全新的 GlobalConfig 对象，在新对象上完成所有变更，
	// 最后原子替换全局指针，保证读取方永远看到一致的配置快照。
	cfg, _ := loadNewConfigFromDir(cw.configDir)
	applyEnvOverridesOn(cfg)
	setDefaultsOn(cfg)
	validateConfigOn(cfg)
	SetConfig(cfg)

	for _, cb := range cw.callbacks {
		cb(cfg)
	}
}

func isYAMLFile(name string) bool {
	ext := strings.ToLower(filepath.Ext(name))
	return ext == ".yaml" || ext == ".yml"
}
