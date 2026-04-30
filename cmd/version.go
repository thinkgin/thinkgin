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
		cfg := app.GetConfig()
		if cfg != nil && cfg.App.Version != "" {
			fmt.Printf("ThinkGin v%s\n", cfg.App.Version)
		} else {
			fmt.Println("ThinkGin (version unknown)")
		}
	},
}
