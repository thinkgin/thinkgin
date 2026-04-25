# ThinkGin Makefile
# 常用开发命令的统一入口，避免记忆冗长的 go 命令行参数。

.PHONY: build run test lint vet fmt clean scaffold help

# 默认目标
help: ## 显示帮助
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | \
		awk 'BEGIN {FS = ":.*?## "}; {printf "  \033[36m%-15s\033[0m %s\n", $$1, $$2}'

build: ## 编译主程序
	go build -v -o bin/thinkgin ./main.go

run: ## 启动开发服务器
	go run ./main.go

test: ## 运行全部测试（含 race 检测）
	go test -race -coverprofile=coverage.out -covermode=atomic ./...
	@echo "coverage report: coverage.out"

cover: test ## 生成并打开覆盖率报告
	go tool cover -html=coverage.out -o coverage.html
	@echo "open coverage.html to view"

lint: ## 运行 golangci-lint
	golangci-lint run --timeout=5m

vet: ## 运行 go vet
	go vet ./...

fmt: ## 格式化代码
	gofmt -s -w .
	goimports -w .

tidy: ## 整理依赖
	go mod tidy
	go mod verify

scaffold: ## 生成模块骨架 (用法: make scaffold MOD=user)
	@if [ -z "$(MOD)" ]; then echo "用法: make scaffold MOD=<module_name>"; exit 1; fi
	go run ./cmd/scaffold new module $(MOD)

clean: ## 清理构建产物
	rm -rf bin/ coverage.out coverage.html
