package app

import "sync/atomic"

// configPtr 是进程级的全局配置原子指针。
// 所有读取必须通过 GetConfig()，所有写入必须通过 SetConfig()。
// 这保证了配置热更新期间的并发安全。
var configPtr atomic.Pointer[GlobalConfig]

// Config 是已弃用的全局配置直接访问入口。
// 为保持向后兼容和测试代码兼容，此变量仍可读写，
// 但新代码应使用 GetConfig() / SetConfig()。
//
// Deprecated: 使用 GetConfig() 代替读取，使用 SetConfig() 代替写入。
var Config *GlobalConfig

// GetConfig 返回全局配置（线程安全）。如果尚未初始化，返回 nil。
// 调用方应确保在 main 或 init 阶段完成 Bootstrap 后再使用。
func GetConfig() *GlobalConfig {
	if p := configPtr.Load(); p != nil {
		return p
	}
	// 兼容：如果旧代码直接写了 app.Config = xxx，也能读到。
	return Config
}

// SetConfig 原子地替换全局配置（线程安全）。
// 同时更新 Config 兼容变量，确保旧代码也能读到最新值。
func SetConfig(cfg *GlobalConfig) {
	configPtr.Store(cfg)
	Config = cfg
}
