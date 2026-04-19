// 命令：thinkgin-scaffold
//
// 当前仅支持一个子命令：
//
//	thinkgin-scaffold new module <name>
//
// 生成位于 app/<name>/ 的 MVC 骨架，包含：
//   - controller/<name>.go
//   - model/<name>.go
//   - view/.gitkeep
//
// 运行示例：
//
//	go run ./cmd/scaffold new module user
//
// 设计取舍：不引入 cobra/cli 等 CLI 框架，保持零外部依赖。
package main

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"unicode"
)

// version 随仓库语义化版本一起发布。
const version = "0.1.0"

func main() {
	if err := run(os.Args[1:]); err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}
}

// run 是真正的入口，便于单测时注入 args。
func run(args []string) error {
	if len(args) == 0 {
		printUsage()
		return nil
	}
	switch args[0] {
	case "new":
		return cmdNew(args[1:])
	case "version", "-v", "--version":
		fmt.Println("thinkgin-scaffold", version)
		return nil
	case "help", "-h", "--help":
		printUsage()
		return nil
	default:
		printUsage()
		return fmt.Errorf("unknown command: %s", args[0])
	}
}

// printUsage 打印一屏简明说明，避免长文案淹没新用户。
func printUsage() {
	fmt.Println(`thinkgin-scaffold - ThinkGin 脚手架

Usage:
  thinkgin-scaffold new module <name>   生成一个 MVC 模块骨架到 app/<name>/
  thinkgin-scaffold version             打印版本号
  thinkgin-scaffold help                显示本帮助

Examples:
  thinkgin-scaffold new module user
  thinkgin-scaffold new module orderItem`)
}

// cmdNew 实现 "new <subcmd>"。当前仅支持 "module"。
func cmdNew(args []string) error {
	if len(args) == 0 {
		return errors.New("usage: thinkgin-scaffold new module <name>")
	}
	switch args[0] {
	case "module":
		if len(args) < 2 {
			return errors.New("usage: thinkgin-scaffold new module <name>")
		}
		return newModule(args[1])
	default:
		return fmt.Errorf("unknown subcommand: new %s", args[0])
	}
}

// newModule 在 app/<name>/ 下创建标准 MVC 骨架。
// 已存在且非空时拒绝覆盖，避免误删用户代码。
func newModule(name string) error {
	if err := validateModuleName(name); err != nil {
		return err
	}

	base := filepath.Join("app", name)
	if fi, err := os.Stat(base); err == nil && fi.IsDir() {
		if notEmpty, err := dirNotEmpty(base); err != nil {
			return err
		} else if notEmpty {
			return fmt.Errorf("%s already exists and is not empty; refusing to overwrite", base)
		}
	}

	files := map[string]string{
		filepath.Join(base, "controller", name+".go"): renderControllerTemplate(name),
		filepath.Join(base, "model", name+".go"):      renderModelTemplate(name),
		filepath.Join(base, "view", ".gitkeep"):       "",
	}

	for path, content := range files {
		if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
			return fmt.Errorf("mkdir %s: %w", filepath.Dir(path), err)
		}
		if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
			return fmt.Errorf("write %s: %w", path, err)
		}
		fmt.Println("created:", path)
	}

	fmt.Printf(`
Module %q created. Next steps:
  1. 在 route/ 中注册路由，导入 "thinkgin/app/%s/controller"
  2. 将 Model 的 AutoMigrate 加入你的启动流程（如有）
`, name, name)
	return nil
}

// validateModuleName 限制模块名为合法的 Go 包名：[a-z][a-z0-9_]*。
// 禁止驼峰（如 orderItem）以避免跨平台的 case-insensitive 目录冲突。
func validateModuleName(name string) error {
	if name == "" {
		return errors.New("module name must not be empty")
	}
	for i, r := range name {
		switch {
		case i == 0 && !unicode.IsLower(r):
			return fmt.Errorf("module name %q must start with a lowercase letter", name)
		case !unicode.IsLower(r) && !unicode.IsDigit(r) && r != '_':
			return fmt.Errorf("module name %q contains invalid char %q; allowed: a-z, 0-9, _", name, r)
		}
	}
	return nil
}

// dirNotEmpty 判断目录是否包含除 .gitkeep 之外的任何条目。
func dirNotEmpty(dir string) (bool, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return false, err
	}
	for _, e := range entries {
		if e.Name() != ".gitkeep" {
			return true, nil
		}
	}
	return false, nil
}

// renderControllerTemplate 生成 controller 文件内容。
// 使用显式字符串拼接而非 text/template，避免额外的模板解析成本与风险。
func renderControllerTemplate(name string) string {
	title := title(name)
	return fmt.Sprintf(`// Package controller 为 %s 模块提供 HTTP 处理函数。
package controller

import "github.com/gin-gonic/gin"

// %s 是占位处理函数，返回 {"module":"%s"}。
// 把它注册到路由：
//
//	r.GET("/%s", controller.%s)
func %s(c *gin.Context) {
	c.JSON(200, gin.H{"module": %q})
}
`, name, title, name, name, title, title, name)
}

// renderModelTemplate 生成一份带 gorm.Model 内嵌的最小 Model。
func renderModelTemplate(name string) string {
	title := title(name)
	return fmt.Sprintf(`// Package model 为 %s 模块提供数据模型。
package model

import "gorm.io/gorm"

// %s 是 %s 模块的示例模型，实际字段请按业务调整。
type %s struct {
	gorm.Model
	Name string %s
}

// TableName 显式指定表名，避免 GORM 默认复数化的歧义。
func (%s) TableName() string { return %q }
`, name, title, name, title, "`gorm:\"size:128;not null\"`", title, name+"s")
}

// title 返回 module name 的首字母大写形式，用作 Go 导出类型名。
// 例如 user → User；order_item → Order_item（模块名已限制为小写+下划线）。
func title(name string) string {
	if name == "" {
		return ""
	}
	return strings.ToUpper(name[:1]) + name[1:]
}
