package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// chdir 切换到临时目录并在测试结束时恢复，避免污染仓库根。
func chdir(t *testing.T) {
	t.Helper()
	oldwd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	tmp := t.TempDir()
	if err := os.Chdir(tmp); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = os.Chdir(oldwd) })
}

func TestValidateModuleName(t *testing.T) {
	tests := []struct {
		name    string
		wantErr bool
	}{
		{"user", false},
		{"order_item", false},
		{"a1", false},
		{"", true},
		{"User", true},      // 大写首字母
		{"order-item", true}, // 连字符
		{"Order", true},
		{"1abc", true}, // 数字开头
	}
	for _, tt := range tests {
		err := validateModuleName(tt.name)
		if (err != nil) != tt.wantErr {
			t.Errorf("validateModuleName(%q) err=%v, wantErr=%v", tt.name, err, tt.wantErr)
		}
	}
}

func TestNewModule_CreatesExpectedFiles(t *testing.T) {
	chdir(t)
	if err := newModule("user"); err != nil {
		t.Fatalf("newModule() error: %v", err)
	}
	for _, want := range []string{
		filepath.Join("app", "user", "controller", "user.go"),
		filepath.Join("app", "user", "model", "user.go"),
		filepath.Join("app", "user", "view", ".gitkeep"),
	} {
		if _, err := os.Stat(want); err != nil {
			t.Errorf("expected %s to exist: %v", want, err)
		}
	}
}

func TestNewModule_RefusesToOverwriteNonEmpty(t *testing.T) {
	chdir(t)
	// 先创建一个已有文件的模块目录。
	path := filepath.Join("app", "user", "controller", "existing.go")
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte("// keep"), 0o644); err != nil {
		t.Fatal(err)
	}

	if err := newModule("user"); err == nil {
		t.Fatal("expected error when module dir is non-empty")
	}
	// 确保原文件未被覆盖
	b, _ := os.ReadFile(path)
	if string(b) != "// keep" {
		t.Errorf("existing file was modified: %q", string(b))
	}
}

func TestNewModule_ControllerContainsValidGo(t *testing.T) {
	chdir(t)
	if err := newModule("foo"); err != nil {
		t.Fatal(err)
	}
	got, err := os.ReadFile(filepath.Join("app", "foo", "controller", "foo.go"))
	if err != nil {
		t.Fatal(err)
	}
	content := string(got)
	for _, must := range []string{
		"package controller",
		"func Foo(c *gin.Context)",
		`"module": "foo"`,
	} {
		if !strings.Contains(content, must) {
			t.Errorf("controller.go missing %q, got:\n%s", must, content)
		}
	}
}

func TestRun_UnknownCommandReturnsError(t *testing.T) {
	if err := run([]string{"banana"}); err == nil {
		t.Error("expected error for unknown command")
	}
}

func TestRun_VersionPrints(t *testing.T) {
	if err := run([]string{"version"}); err != nil {
		t.Errorf("version should not error: %v", err)
	}
}
