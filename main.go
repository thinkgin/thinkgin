// ThinkGin 程序入口。
//
// 基于 Cobra CLI 框架，支持以下命令：
//   - serve    启动 HTTP 服务器（默认，无参数时自动执行）
//   - version  打印版本号
//   - config   输出当前合并后的配置
//   - cron     启动定时任务调度器
//   - scaffold 生成模块骨架（保留原有脚手架功能）
//
// 向后兼容：直接运行 `go run main.go` 等同于 `go run main.go serve`。
package main

import "thinkgin/cmd"

func main() {
	cmd.Execute()
}
