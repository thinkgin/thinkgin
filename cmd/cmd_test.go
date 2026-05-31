package cmd

import (
	"bytes"
	"io"
	"os"
	"strings"
	"testing"

	"thinkgin/app"

	"github.com/spf13/cobra"
)

// findSubCommand 在 rootCmd 的直接子命令中按 Use 的首单词查找。
func findSubCommand(use string) *cobra.Command {
	for _, c := range rootCmd.Commands() {
		if strings.Fields(c.Use)[0] == use {
			return c
		}
	}
	return nil
}

// TestRootCommand_RegistersAllSubcommands 确认所有内置子命令都已通过各自的 init() 注册到 rootCmd。
func TestRootCommand_RegistersAllSubcommands(t *testing.T) {
	want := []string{"serve", "version", "config", "migrate", "cron", "swagger"}
	for _, name := range want {
		if findSubCommand(name) == nil {
			t.Errorf("子命令 %q 未注册到 rootCmd", name)
		}
	}
}

// TestRootCommand_Metadata 校验根命令的基本元信息。
func TestRootCommand_Metadata(t *testing.T) {
	if rootCmd.Use != "thinkgin" {
		t.Errorf("rootCmd.Use = %q, want %q", rootCmd.Use, "thinkgin")
	}
	if rootCmd.Short == "" {
		t.Error("rootCmd.Short 不应为空")
	}
	if rootCmd.Run == nil {
		t.Error("rootCmd.Run 不应为 nil（无参数时需回退到 serve）")
	}
}

// captureStdout 执行 fn 并返回其写入 os.Stdout 的内容。
func captureStdout(t *testing.T, fn func()) string {
	t.Helper()
	orig := os.Stdout
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf("os.Pipe: %v", err)
	}
	os.Stdout = w

	done := make(chan string, 1)
	go func() {
		var buf bytes.Buffer
		_, _ = io.Copy(&buf, r)
		done <- buf.String()
	}()

	fn()

	_ = w.Close()
	os.Stdout = orig
	return <-done
}

// TestVersionCommand_PrintsVersion 校验 version 命令输出含 app.Version。
func TestVersionCommand_PrintsVersion(t *testing.T) {
	out := captureStdout(t, func() {
		versionCmd.Run(versionCmd, nil)
	})
	if !strings.Contains(out, app.Version) {
		t.Errorf("version 输出 %q 不含版本号 %q", out, app.Version)
	}
	if !strings.Contains(out, "ThinkGin") {
		t.Errorf("version 输出 %q 不含 'ThinkGin'", out)
	}
}

// TestMigrateCommand_HasSubcommands 校验 migrate 命令挂载了 up/down/status 三个子命令。
func TestMigrateCommand_HasSubcommands(t *testing.T) {
	migrate := findSubCommand("migrate")
	if migrate == nil {
		t.Fatal("migrate 命令未注册")
	}
	want := map[string]bool{"up": false, "down": false, "status": false}
	for _, c := range migrate.Commands() {
		name := strings.Fields(c.Use)[0]
		if _, ok := want[name]; ok {
			want[name] = true
		}
	}
	for name, found := range want {
		if !found {
			t.Errorf("migrate 子命令 %q 未注册", name)
		}
	}
}

// TestExecute_VersionDoesNotExit 通过设置参数走 Execute，确认 version 路径不触发非零退出。
// 这里直接驱动 rootCmd 以避免 os.Exit；Execute 内部仅在出错时退出。
func TestExecute_VersionPath(t *testing.T) {
	rootCmd.SetArgs([]string{"version"})
	defer rootCmd.SetArgs(nil)

	out := captureStdout(t, func() {
		if err := rootCmd.Execute(); err != nil {
			t.Errorf("执行 version 命令出错: %v", err)
		}
	})
	if !strings.Contains(out, app.Version) {
		t.Errorf("Execute version 输出 %q 不含版本号", out)
	}
}

// TestRedactConfig_MasksSecrets 校验 config 命令脱敏逻辑覆盖所有敏感字段，
// 且不修改原始配置对象。
func TestRedactConfig_MasksSecrets(t *testing.T) {
	src := &app.GlobalConfig{}
	src.App.JWT.Secret = "super-secret-key"
	src.Prometheus.Auth.Password = "prom-pass"
	src.Database.Connections = map[string]app.ConnectionConfig{
		"mysql": {Driver: "mysql", Password: "db-pass"},
		"sqlite": {Driver: "sqlite"}, // 无密码，应保持空
	}

	got := redactConfig(src)

	if got.App.JWT.Secret != redactedMark {
		t.Errorf("JWT secret 未脱敏: %q", got.App.JWT.Secret)
	}
	if got.Prometheus.Auth.Password != redactedMark {
		t.Errorf("prometheus 密码未脱敏: %q", got.Prometheus.Auth.Password)
	}
	if got.Database.Connections["mysql"].Password != redactedMark {
		t.Errorf("DB 密码未脱敏: %q", got.Database.Connections["mysql"].Password)
	}
	if got.Database.Connections["sqlite"].Password != "" {
		t.Errorf("空密码不应被替换: %q", got.Database.Connections["sqlite"].Password)
	}

	// 原始对象不能被修改（深拷贝保证）。
	if src.App.JWT.Secret != "super-secret-key" {
		t.Error("redactConfig 不应修改原始配置的 JWT secret")
	}
	if src.Database.Connections["mysql"].Password != "db-pass" {
		t.Error("redactConfig 不应修改原始配置的 DB 密码")
	}
}

// TestRedactConfig_Nil 确认 nil 输入安全。
func TestRedactConfig_Nil(t *testing.T) {
	if redactConfig(nil) != nil {
		t.Error("redactConfig(nil) 应返回 nil")
	}
}
