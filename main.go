// ThinkGin 程序入口。
//
// 执行流程：
//   1. 加载配置与日志器（app 包内部已完成初始化）。
//   2. 初始化 OpenTelemetry TracerProvider，main 负责在退出时 Shutdown。
//   3. 监听 SIGINT/SIGTERM，通过 context 向 App 下发停机信号。
//   4. 构造并运行 framework.App，阻塞直到 Run 返回。
package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"thinkgin/app"
	"thinkgin/extend/middleware"
	"thinkgin/framework"
)

// banner 仅在 Debug 模式下打印，避免污染生产日志。
// 版本号不再写死在字符画里，改由 printBanner 从配置动态拼接，避免升级遗漏修改。
const banner = `
 _________  __        _            __        ______   _
|  _   _  |[  |      (_)          [  |  _  .' ___  | (_)
|_/ | | \_| | |--.   __   _ .--.   | | / ]/ .'   \_| __   _ .--.
    | |     | .-. | [  | [ '.-. |  | '' < | |   ____[  | [ '.-. |
   _| |_    | | | |  | |  | | | |  | |'\ \\ '.___]  || |  | | | |
  |_____|  [___]|__][___][___||__][__|  \_]'._____.'[___][___||__]
`

// printBanner 输出 logo + 动态版本号，仅在 Debug 模式调用。
func printBanner(version string) {
	fmt.Print(banner)
	fmt.Printf("  v%s\n\n", version)
}

func main() {
	// 显式再跑一次 Bootstrap：允许通过 THINKGIN_CONFIG_DIR 指定非默认目录，
	// 同时让配置加载错误能被 main 感知（包 init 是兜底，错误只打印）。
	configDir := os.Getenv("THINKGIN_CONFIG_DIR")
	if err := app.Bootstrap(configDir); err != nil {
		fmt.Printf("[main] 配置加载存在缺失: %v\n", err)
	}

	cfg := app.GetConfig()
	logger := app.GetLogger()

	if cfg.App.Debug {
		printBanner(cfg.App.Version)
	}

	// 初始化链路追踪；若未启用 Tracer，返回的 shutdown 是 no-op。
	shutdownTracer := middleware.InitTracer()
	defer func() {
		// 独立超时，避免受 ctx 已取消影响而无法上报最后一批 span。
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()
		_ = shutdownTracer(ctx)
	}()

	// 监听 SIGINT / SIGTERM，用户可通过 Ctrl+C 或容器编排器触发停机。
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	a, err := framework.New(
		framework.WithConfig(cfg),
		framework.WithLogger(logger),
		framework.WithShutdownTimeout(10*time.Second),
		// 仅在 app.debug=true 时生效，生产模式自动跳过。
		framework.WithOpenBrowser(true),
	)
	if err != nil {
		logger.Fatalf("app init failed: %v", err)
	}

	if err := a.Run(ctx); err != nil {
		logger.Fatalf("server exited with error: %v", err)
	}
}
