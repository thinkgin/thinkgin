package framework

import (
	"os/exec"
	"runtime"
)

// tryOpenBrowser 尽力而为地在默认浏览器打开 URL，失败时静默忽略。
// 该能力仅用于开发期，生产部署请勿依赖。
func tryOpenBrowser(url string) {
	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "windows":
		cmd = exec.Command("cmd", "/c", "start", url)
	case "darwin":
		cmd = exec.Command("open", url)
	default:
		cmd = exec.Command("xdg-open", url)
	}
	_ = cmd.Start()
}
