package cmd

import (
	"fmt"
	"os"
	"os/exec"

	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(swaggerCmd)
}

var swaggerCmd = &cobra.Command{
	Use:   "swagger",
	Short: "生成 Swagger/OpenAPI 文档",
	Long: `调用 swag init 生成 API 文档（需要先安装 swag CLI）。

安装 swag:
  go install github.com/swaggo/swag/cmd/swag@latest

生成后的文件位于 docs/ 目录，/swagger/ 端点自动加载。
如果不使用 swag，也可以手写 docs/openapi.yaml。`,
	Run: func(cmd *cobra.Command, args []string) {
		// 检查 swag 是否安装
		swagPath, err := exec.LookPath("swag")
		if err != nil {
			fmt.Println("未找到 swag 命令。请先安装:")
			fmt.Println("  go install github.com/swaggo/swag/cmd/swag@latest")
			fmt.Println("\n或者手写 docs/openapi.yaml 文件。")
			os.Exit(1)
		}

		fmt.Printf("使用 %s 生成文档...\n", swagPath)

		c := exec.Command("swag", "init",
			"--generalInfo", "cmd/serve.go",
			"--output", "docs",
			"--parseDependency",
		)
		c.Stdout = os.Stdout
		c.Stderr = os.Stderr

		if err := c.Run(); err != nil {
			fmt.Fprintf(os.Stderr, "swag init 失败: %v\n", err)
			os.Exit(1)
		}

		fmt.Println("文档已生成到 docs/ 目录。")
		fmt.Println("启动服务后访问 http://localhost:8000/swagger/ 查看。")
	},
}
