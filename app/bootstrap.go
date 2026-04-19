package app

import "fmt"

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
	InitLogger()

	return loadErr
}

// init 作为零配置兜底：当用户直接 import "thinkgin/app" 时，
// 自动从默认路径加载配置，保证 GetConfig / GetLogger 立即可用。
//
// 显式控制场景（测试、嵌入式场景）应调用 Bootstrap 并传入自定义目录。
// 加载错误会被打印但不会 panic，便于本地开发在没有 config/ 目录时也能启动。
func init() {
	if err := Bootstrap(defaultConfigDir); err != nil {
		fmt.Printf("[bootstrap] 部分配置加载失败，已使用默认值: %v\n", err)
	}
}
