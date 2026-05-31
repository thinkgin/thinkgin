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

// redactedMark 是脱敏占位符。
const redactedMark = "***REDACTED***"

var configCmd = &cobra.Command{
	Use:   "config",
	Short: "输出当前合并后的配置（JSON 格式，敏感字段已脱敏）",
	Long: "加载所有 YAML + 环境变量覆盖后，以 JSON 形式打印完整配置，用于调试配置问题。\n" +
		"为避免泄漏，JWT 密钥、数据库/Redis 密码、Prometheus 认证密码等敏感字段会被脱敏。\n" +
		"如需查看原始值，请加 --show-secrets（谨慎使用，勿在共享终端/CI 日志中执行）。",
	Run: func(cmd *cobra.Command, args []string) {
		configDir := os.Getenv("THINKGIN_CONFIG_DIR")
		_ = app.Bootstrap(configDir)

		cfg := app.GetConfig()
		showSecrets, _ := cmd.Flags().GetBool("show-secrets")
		if !showSecrets {
			cfg = redactConfig(cfg)
		}

		data, err := json.MarshalIndent(cfg, "", "  ")
		if err != nil {
			fmt.Fprintf(os.Stderr, "marshal config: %v\n", err)
			os.Exit(1)
		}
		fmt.Println(string(data))
	},
}

func init() {
	configCmd.Flags().Bool("show-secrets", false, "输出明文敏感字段（默认脱敏）")
}

// redactConfig 返回一个副本，把敏感字段替换为占位符，避免 dump 时泄漏。
// 仅对"有值"的字段脱敏，空值保持为空，便于区分"未配置"与"已配置但隐藏"。
func redactConfig(src *app.GlobalConfig) *app.GlobalConfig {
	if src == nil {
		return nil
	}
	// 浅拷贝顶层结构（含值类型字段）。
	clone := *src

	// JWT 密钥
	if clone.App.JWT.Secret != "" {
		clone.App.JWT.Secret = redactedMark
	}

	// Prometheus 认证密码
	if clone.Prometheus.Auth.Password != "" {
		clone.Prometheus.Auth.Password = redactedMark
	}

	// 数据库/Redis 连接密码：map 是引用类型，必须深拷贝避免改动全局配置。
	if len(src.Database.Connections) > 0 {
		conns := make(map[string]app.ConnectionConfig, len(src.Database.Connections))
		for name, conn := range src.Database.Connections {
			if conn.Password != "" {
				conn.Password = redactedMark
			}
			conns[name] = conn
		}
		clone.Database.Connections = conns
	}

	return &clone
}
