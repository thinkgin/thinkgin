package cmd

import (
	"fmt"
	"os"
	"strconv"

	"thinkgin/app"
	"thinkgin/app/database"
	"thinkgin/app/database/migrations"

	"github.com/spf13/cobra"
)

func init() {
	migrateCmd.AddCommand(migrateUpCmd)
	migrateCmd.AddCommand(migrateDownCmd)
	migrateCmd.AddCommand(migrateStatusCmd)
	rootCmd.AddCommand(migrateCmd)
}

var migrateCmd = &cobra.Command{
	Use:   "migrate",
	Short: "数据库迁移管理",
	Long:  "管理数据库结构迁移：执行、回退和查看状态。",
}

var migrateUpCmd = &cobra.Command{
	Use:   "up",
	Short: "执行所有未应用的迁移",
	Run: func(cmd *cobra.Command, args []string) {
		m := getMigrator()
		if m == nil {
			return
		}

		if err := m.Migrate(); err != nil {
			fmt.Fprintf(os.Stderr, "迁移失败: %v\n", err)
			os.Exit(1)
		}
		fmt.Println("迁移完成。")
	},
}

var migrateDownCmd = &cobra.Command{
	Use:   "down [n]",
	Short: "回退最后 n 个迁移（默认 1）",
	Run: func(cmd *cobra.Command, args []string) {
		n := 1
		if len(args) > 0 {
			if v, err := strconv.Atoi(args[0]); err == nil && v > 0 {
				n = v
			}
		}

		m := getMigrator()
		if m == nil {
			return
		}

		if err := m.Rollback(n); err != nil {
			fmt.Fprintf(os.Stderr, "回退失败: %v\n", err)
			os.Exit(1)
		}
		fmt.Printf("已回退 %d 个迁移。\n", n)
	},
}

var migrateStatusCmd = &cobra.Command{
	Use:   "status",
	Short: "查看迁移状态",
	Run: func(cmd *cobra.Command, args []string) {
		m := getMigrator()
		if m == nil {
			return
		}

		statuses, err := m.Status()
		if err != nil {
			fmt.Fprintf(os.Stderr, "查询状态失败: %v\n", err)
			os.Exit(1)
		}

		if len(statuses) == 0 {
			fmt.Println("没有已注册的迁移。")
			return
		}

		fmt.Printf("%-20s %s\n", "VERSION", "STATUS")
		fmt.Printf("%-20s %s\n", "-------", "------")
		for _, s := range statuses {
			status := "pending"
			if s.Applied {
				status = "applied"
			}
			fmt.Printf("%-20s %s\n", s.Version, status)
		}
	},
}

func getMigrator() *database.Migrator {
	configDir := os.Getenv("THINKGIN_CONFIG_DIR")
	_ = app.Bootstrap(configDir)

	if err := database.Init(); err != nil {
		fmt.Fprintf(os.Stderr, "数据库初始化失败: %v\n", err)
		return nil
	}

	db := database.Default()
	if db == nil {
		fmt.Fprintln(os.Stderr, "未配置默认数据库连接。请检查 config/database.yaml 的 default 字段。")
		return nil
	}

	m := database.NewMigrator(db)
	m.Register(migrations.All()...)
	return m
}
