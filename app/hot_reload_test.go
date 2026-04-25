package app

import (
	"os"
	"path/filepath"
	"sync/atomic"
	"testing"
	"time"
)

func TestWatchConfig_ReloadsOnFileChange(t *testing.T) {
	dir := t.TempDir()

	// 写初始配置
	appYAML := filepath.Join(dir, "app.yaml")
	if err := os.WriteFile(appYAML, []byte("app:\n  name: Before\n"), 0644); err != nil {
		t.Fatal(err)
	}

	// Bootstrap 使用该目录
	Config = nil
	_ = Bootstrap(dir)

	if GetConfig().App.Name != "Before" {
		t.Fatalf("initial App.Name = %q, want Before", GetConfig().App.Name)
	}

	var reloaded atomic.Int32
	watcher, err := WatchConfig(dir, func(cfg *GlobalConfig) {
		reloaded.Add(1)
	})
	if err != nil {
		t.Fatalf("WatchConfig: %v", err)
	}
	defer watcher.Close()

	// 修改文件触发热更新
	time.Sleep(100 * time.Millisecond) // 等 watcher 就绪
	if err := os.WriteFile(appYAML, []byte("app:\n  name: After\n"), 0644); err != nil {
		t.Fatal(err)
	}

	// 等去抖 + 重载完成（debounce 500ms + margin）
	time.Sleep(1500 * time.Millisecond)

	if reloaded.Load() == 0 {
		t.Error("callback was not invoked after file change")
	}
	if GetConfig().App.Name != "After" {
		t.Errorf("App.Name = %q, want After", GetConfig().App.Name)
	}
}

func TestWatchConfig_IgnoresNonYAML(t *testing.T) {
	dir := t.TempDir()

	// 需要至少一个 yaml 让 Bootstrap 不报错
	os.WriteFile(filepath.Join(dir, "app.yaml"), []byte("app:\n  name: Test\n"), 0644)
	Config = nil
	_ = Bootstrap(dir)

	var reloaded atomic.Int32
	watcher, err := WatchConfig(dir, func(cfg *GlobalConfig) {
		reloaded.Add(1)
	})
	if err != nil {
		t.Fatalf("WatchConfig: %v", err)
	}
	defer watcher.Close()

	time.Sleep(100 * time.Millisecond)

	// 写一个 .txt 文件，不应触发重载
	os.WriteFile(filepath.Join(dir, "notes.txt"), []byte("ignore me"), 0644)

	time.Sleep(1000 * time.Millisecond)

	if reloaded.Load() != 0 {
		t.Error("callback should not be invoked for non-YAML files")
	}
}
