package cmd

import (
	"fmt"

	"thinkgin/app"

	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(versionCmd)
}

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "打印版本号",
	Run: func(cmd *cobra.Command, args []string) {
		// 优先使用 app.Version（支持 ldflags 编译注入）。
		fmt.Printf("ThinkGin v%s\n", app.Version)
	},
}
