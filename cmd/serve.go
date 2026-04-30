package cmd

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"thinkgin/app"
	"thinkgin/app/cache"
	"thinkgin/app/database"
	"thinkgin/app/session"
	"thinkgin/extend/middleware"
	"thinkgin/framework"

	"github.com/spf13/cobra"
)

// banner 仅在 Debug 模式下打印。
const banner = `
 _________  __        _            __        ______   _
|  _   _  |[  |      (_)          [  |  _  .' ___  | (_)
|_/ | | \_| | |--.   __   _ .--.   | | / ]/ .'   \_| __   _ .--.
    | |     | .-. | [  | [ '.-. |  | '' < | |   ____[  | [ '.-. |
   _| |_    | | | |  | |  | | | |  | |` + "`" + `\ \\ '.___]  || |  | | | |
  |_____|  [___]|__][___][___||__][__|  \_]'._____.'[___][___||__]
`

func init() {
	rootCmd.AddCommand(serveCmd)
}

var serveCmd = &cobra.Command{
	Use:   "serve",
	Short: "启动 HTTP 服务器",
	Long:  "加载配置、初始化数据库/缓存/Session/Tracer，启动 Gin HTTP(S) 服务。",
	Run:   runServe,
}

func runServe(cmd *cobra.Command, args []string) {
	configDir := os.Getenv("THINKGIN_CONFIG_DIR")
	if err := app.Bootstrap(configDir); err != nil {
		fmt.Printf("[main] 配置加载存在缺失: %v\n", err)
	}

	cfg := app.GetConfig()
	logger := app.GetLogger()

	if cfg.App.Debug {
		fmt.Print(banner)
		fmt.Printf("  v%s\n\n", cfg.App.Version)
	}

	if err := database.Init(); err != nil {
		logger.Warnf("[database] 部分连接初始化失败: %v", err)
	}
	defer func() {
		if err := database.CloseAll(); err != nil {
			logger.Warnf("[database] 关闭连接时发生错误: %v", err)
		}
	}()

	if err := cache.Init(); err != nil {
		logger.Warnf("[cache] 部分缓存初始化失败: %v", err)
	}
	defer func() {
		if err := cache.CloseAll(); err != nil {
			logger.Warnf("[cache] 关闭客户端时发生错误: %v", err)
		}
	}()

	if err := session.Init(); err != nil {
		logger.Warnf("[session] 初始化失败: %v", err)
	}
	defer func() {
		if err := session.Shutdown(); err != nil {
			logger.Warnf("[session] 关闭时发生错误: %v", err)
		}
	}()

	shutdownTracer := middleware.InitTracer()
	defer func() {
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()
		_ = shutdownTracer(ctx)
	}()

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	a, err := framework.New(
		framework.WithConfig(cfg),
		framework.WithLogger(logger),
		framework.WithShutdownTimeout(10*time.Second),
		framework.WithOpenBrowser(true),
	)
	if err != nil {
		logger.Fatalf("app init failed: %v", err)
	}

	if err := a.Run(ctx); err != nil {
		logger.Fatalf("server exited with error: %v", err)
	}
}
