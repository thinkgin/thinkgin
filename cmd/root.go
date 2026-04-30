// Package cmd 提供 ThinkGin 的 CLI 命令框架。
//
// 基于 Cobra 构建，所有子命令在各自文件中通过 init() 注册到 rootCmd。
// main.go 调用 cmd.Execute() 即可。
//
// 内置命令：
//   - serve    启动 HTTP 服务器（默认命令）
//   - version  打印版本号
//   - config   输出当前合并后的配置
//   - migrate  数据库迁移
//   - cron     启动定时任务调度器
package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "thinkgin",
	Short: "ThinkGin — 基于 Gin 的 Go Web 框架",
	Long:  "ThinkGin 是一个保持 Gin 简洁性的同时提供企业级能力的 Web 框架。",
	// 无子命令时默认执行 serve，向后兼容 go run main.go 直接启动服务器。
	Run: func(cmd *cobra.Command, args []string) {
		serveCmd, _, _ := cmd.Find([]string{"serve"})
		if serveCmd != nil {
			serveCmd.Run(serveCmd, args)
		}
	},
}

// Execute 是 CLI 的总入口，main.go 调用此函数。
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
