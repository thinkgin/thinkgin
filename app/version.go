package app

// Version 是框架版本号，作为唯一真实来源（Single Source of Truth）。
//
// 编译时可通过 ldflags 注入自定义版本号：
//
//	go build -ldflags "-X thinkgin/app.Version=3.8.1" -o thinkgin
//
// 未注入时使用此处硬编码的默认值。
var Version = "3.9.1"
