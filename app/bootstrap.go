package app

import (
	"fmt"
	"os"
	"sync"
)

// bootstrapOnce 保证 Bootstrap 的"副作用部分"（加载 YAML、设默认、校验、初始化 Logger）
// 在进程内最多执行一次。显式调用 Bootstrap(customDir) 会跳过 once，支持测试注入。
var bootstrapOnce sync.Once

// Bootstrap 是显式的初始化入口，执行顺序：
//  1. 加载 configDir 下的所有 YAML
//  2. 应用环境变量覆盖（THINKGIN_* 前缀）
//  3. 填充默认值
//  4. 语义校验与纠偏
//  5. 初始化全局 Logger
//
// configDir 为空时使用 defaultConfigDir（即 "config"）。
// 返回的 error 只反映"部分 YAML 文件加载失败"的聚合错误；
// 校验/默认值阶段为了脚手架的"零配置可跑"目标不会中断流程。
func Bootstrap(configDir string) error {
	if configDir == "" {
		configDir = defaultConfigDir
	}

	loadErr := LoadConfigFromDir(configDir)
	applyEnvOverrides()
	setDefaultConfig()
	validateConfig()
	// 将配置写入原子指针，保证后续 GetConfig() 线程安全读取。
	SetConfig(Config)
	InitLogger()

	return loadErr
}

// ReBootstrap 跳过 sync.Once 重新执行 Bootstrap，仅供测试使用。
func ReBootstrap(configDir string) error {
	return Bootstrap(configDir)
}

// init 作为零配置兜底：当用户直接 import "thinkgin/app" 时，
// 自动从默认路径加载配置，保证 GetConfig / GetLogger 立即可用。
//
// 测试或工作目录不含 config/ 时直接跳过，避免污染 go test 的输出。
// 真正需要从自定义目录加载时，显式调用 Bootstrap(dir) 即可。
func init() {
	bootstrapOnce.Do(func() {
		if _, err := os.Stat(defaultConfigDir); os.IsNotExist(err) {
			// 没有 config/ 目录就什么都不做：依赖 defaults + env 已经够用。
			setDefaultConfig()
			InitLogger()
			return
		}
		if err := Bootstrap(defaultConfigDir); err != nil {
			fmt.Fprintf(os.Stderr, "[bootstrap] 部分配置加载失败，已使用默认值: %v\n", err)
		}
	})
}
