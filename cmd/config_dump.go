package cmd

import (
	"encoding/json"
	"fmt"
	"os"

	"thinkgin/app"

	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(configCmd)
}

var configCmd = &cobra.Command{
	Use:   "config",
	Short: "输出当前合并后的配置（JSON 格式）",
	Long:  "加载所有 YAML + 环境变量覆盖后，以 JSON 形式打印完整配置。用于调试配置问题。",
	Run: func(cmd *cobra.Command, args []string) {
		configDir := os.Getenv("THINKGIN_CONFIG_DIR")
		_ = app.Bootstrap(configDir)

		cfg := app.GetConfig()
		data, err := json.MarshalIndent(cfg, "", "  ")
		if err != nil {
			fmt.Fprintf(os.Stderr, "marshal config: %v\n", err)
			os.Exit(1)
		}
		fmt.Println(string(data))
	},
}
