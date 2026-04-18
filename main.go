package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"thinkgin/app"
	"thinkgin/extend/middleware"
	"thinkgin/framework"
	"time"
)

func main() {
	// 获取配置
	config := app.GetConfig()
	logger := app.GetLogger()

	//后端服务打印输出
	log := "\n _________  __        _            __        ______   _             _____       ____    \n|  _   _  |[  |      (_)          [  |  _  .' ___  | (_)           / ___ `.   .'    '.  \n|_/ | | \\_| | |--.   __   _ .--.   | | / ]/ .'   \\_| __   _ .--.  |_/___) |  |  .--.  | \n    | |     | .-. | [  | [ `.-. |  | '' < | |   ____[  | [ `.-. |  .'____.'  | |    | | \n   _| |_    | | | |  | |  | | | |  | |`\\ \\\\ `.___]  || |  | | | | / /_____  _|  `--'  | \n  |_____|  [___]|__][___][___||__][__|  \\_]`._____.'[___][___||__]|_______|(_)'.____.'  \n                                                                                        \n"
	fmt.Println(log)

	// 初始化链路追踪
	shutdownTracer := middleware.InitTracer()
	defer shutdownTracer(context.Background())

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	appInstance, err := framework.New(
		framework.WithConfig(config),
		framework.WithLogger(logger),
		framework.WithShutdownTimeout(10*time.Second),
	)
	if err != nil {
		logger.Fatalf("❌ 应用初始化失败: %v", err)
	}

	if err := appInstance.Run(ctx); err != nil {
		logger.Fatalf("❌ 服务器运行失败: %v", err)
	}
}
