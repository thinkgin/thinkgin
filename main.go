package main

import (
	"fmt"
	"net/http"
	"os/exec"
	"thinkgin/route"
)

func main() {

	router := route.InitRouter()
	s := &http.Server{
		Addr: fmt.Sprintf(":%d", 8000),
		//Addr:           fmt.Sprintf(":%d", setting.HTTPPort),
		Handler:     router,
		ReadTimeout: 60,
		//ReadTimeout:    setting.ReadTimeout,
		WriteTimeout: 60,
		//WriteTimeout:   setting.WriteTimeout,
		MaxHeaderBytes: 1 << 20,
	}

	//后端服务打印输出
	log := "\n _________  __        _            __        ______   _             _____       ____    \n|  _   _  |[  |      (_)          [  |  _  .' ___  | (_)           / ___ `.   .'    '.  \n|_/ | | \\_| | |--.   __   _ .--.   | | / ]/ .'   \\_| __   _ .--.  |_/___) |  |  .--.  | \n    | |     | .-. | [  | [ `.-. |  | '' < | |   ____[  | [ `.-. |  .'____.'  | |    | | \n   _| |_    | | | |  | |  | | | |  | |`\\ \\\\ `.___]  || |  | | | | / /_____  _|  `--'  | \n  |_____|  [___]|__][___][___||__][__|  \\_]`._____.'[___][___||__]|_______|(_)'.____.'  \n                                                                                        \n"
	fmt.Println(log)
	fmt.Println("控制台输入：localhost:8000 进入首页")

	//启动前端页面
	err := exec.Command("cmd", "/c", "start", "http://localhost:8000").Start()
	if err != nil {
		fmt.Println("启动失败")
	}

	s.ListenAndServe()

}
