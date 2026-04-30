package cmd

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"thinkgin/app"
	"thinkgin/extend/cron"

	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(cronCmd)
}

var cronCmd = &cobra.Command{
	Use:   "cron",
	Short: "启动定时任务调度器",
	Long: `启动后台 Cron 调度器，执行已注册的定时任务。
按 Ctrl+C 优雅停止，等待正在执行的任务完成。

注册任务请在 extend/cron/ 目录下创建 Go 文件，通过 init() 调用 cron.Register()。`,
	Run: func(cmd *cobra.Command, args []string) {
		configDir := os.Getenv("THINKGIN_CONFIG_DIR")
		_ = app.Bootstrap(configDir)

		logger := app.GetLogger()

		jobs := cron.Jobs()
		if len(jobs) == 0 {
			fmt.Println("没有已注册的定时任务。")
			fmt.Println("请在 extend/cron/ 目录下通过 cron.Register() 注册任务。")
			return
		}

		if err := cron.Start(); err != nil {
			logger.Fatalf("[cron] 启动失败: %v", err)
		}

		fmt.Printf("定时任务调度器已启动，共 %d 个任务。按 Ctrl+C 停止。\n", len(jobs))

		quit := make(chan os.Signal, 1)
		signal.Notify(quit, os.Interrupt, syscall.SIGTERM)
		<-quit

		fmt.Println("\n正在停止调度器...")
		cron.Stop()
		fmt.Println("调度器已停止。")
	},
}
