package main

import (
	"fmt"
	"net/http"
	"thinkgin/route"
)

func main() {

	router := route.InitRouter()
	s := &http.Server{
		Addr: fmt.Sprintf(":%d", 8001),
		//Addr:           fmt.Sprintf(":%d", setting.HTTPPort),
		Handler:     router,
		ReadTimeout: 600,
		//ReadTimeout:    setting.ReadTimeout,
		WriteTimeout: 600,
		//WriteTimeout:   setting.WriteTimeout,
		MaxHeaderBytes: 1 << 20,
	}

	log := "\n _________  __        _            __        ______   _             _____       ____    \n|  _   _  |[  |      (_)          [  |  _  .' ___  | (_)           / ___ `.   .'    '.  \n|_/ | | \\_| | |--.   __   _ .--.   | | / ]/ .'   \\_| __   _ .--.  |_/___) |  |  .--.  | \n    | |     | .-. | [  | [ `.-. |  | '' < | |   ____[  | [ `.-. |  .'____.'  | |    | | \n   _| |_    | | | |  | |  | | | |  | |`\\ \\\\ `.___]  || |  | | | | / /_____  _|  `--'  | \n  |_____|  [___]|__][___][___||__][__|  \\_]`._____.'[___][___||__]|_______|(_)'.____.'  \n                                                                                        \n"
	fmt.Println(log)
	fmt.Println("控制台输入：localhost:8001 进入首页")

	s.ListenAndServe()
}
