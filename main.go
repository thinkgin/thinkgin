package main

import (
	"fmt"
	"net/http"
	"os/exec"
	"thinkgin/app"
	"thinkgin/route"
	"time"

	"github.com/gin-gonic/gin"
)

func main() {
	// 获取配置
	config := app.GetConfig()
	logger := app.GetLogger()

	// 设置Gin运行模式
	gin.SetMode(config.Server.Mode)

	router := route.InitRouter()

	// 从配置文件读取服务器设置
	serverConfig := config.Server
	addr := fmt.Sprintf("%s:%d", serverConfig.HTTP.Host, serverConfig.HTTP.Port)

	s := &http.Server{
		Addr:           addr,
		Handler:        router,
		ReadTimeout:    time.Duration(serverConfig.HTTP.ReadTimeout) * time.Second,
		WriteTimeout:   time.Duration(serverConfig.HTTP.WriteTimeout) * time.Second,
		IdleTimeout:    time.Duration(serverConfig.HTTP.IdleTimeout) * time.Second,
		MaxHeaderBytes: serverConfig.HTTP.MaxHeaderBytes,
	}

	//后端服务打印输出
	log := "\n _________  __        _            __        ______   _             _____       ____    \n|  _   _  |[  |      (_)          [  |  _  .' ___  | (_)           / ___ `.   .'    '.  \n|_/ | | \\_| | |--.   __   _ .--.   | | / ]/ .'   \\_| __   _ .--.  |_/___) |  |  .--.  | \n    | |     | .-. | [  | [ `.-. |  | '' < | |   ____[  | [ `.-. |  .'____.'  | |    | | \n   _| |_    | | | |  | |  | | | |  | |`\\ \\\\ `.___]  || |  | | | | / /_____  _|  `--'  | \n  |_____|  [___]|__][___][___||__][__|  \\_]`._____.'[___][___||__]|_______|(_)'.____.'  \n                                                                                        \n"
	fmt.Println(log)

	logger.Infof("🚀 服务器启动在: http://%s", addr)
	logger.Infof("📋 运行模式: %s", serverConfig.Mode)
	logger.Infof("📱 应用名称: %s v%s", config.App.Name, config.App.Version)

	fmt.Printf("控制台输入：%s 进入首页\n", addr)

	//启动前端页面
	err := exec.Command("cmd", "/c", "start", fmt.Sprintf("http://%s", addr)).Start()
	if err != nil {
		logger.Warnf("自动打开浏览器失败: %v", err)
	}

	if err := s.ListenAndServe(); err != nil {
		logger.Fatalf("❌ 服务器启动失败: %v", err)
	}
}
