package main

import (
	"fmt"
	"net/http"
	"thinkgin/extend/setting"
	"thinkgin/route"
)

func main() {
	router := route.InitRouter()

	s := &http.Server{
		Addr:           fmt.Sprintf(":%d", setting.HTTPPort),
		Handler:        router,
		ReadTimeout:    setting.ReadTimeout,
		WriteTimeout:   setting.WriteTimeout,
		MaxHeaderBytes: 1 << 20,
	}
	
	s.ListenAndServe()
}
