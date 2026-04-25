package app

import (
	"testing"
)

func TestGetLogger_NeverReturnsNil(t *testing.T) {
	// 即使 Log 未初始化，GetLogger 也应返回可用的 Logger。
	old := Log
	Log = nil
	defer func() { Log = old }()

	logger := GetLogger()
	if logger == nil {
		t.Fatal("GetLogger() returned nil when Log is nil")
	}
	// 不应 panic
	logger.Info("test message from fallback logger")
}

func TestGetLogger_ReturnsGlobalLogAfterInit(t *testing.T) {
	Config = &GlobalConfig{}
	setDefaultConfig()
	InitLogger()

	logger := GetLogger()
	if logger == nil {
		t.Fatal("GetLogger() returned nil after InitLogger()")
	}
	if Log == nil {
		t.Fatal("global Log should not be nil after InitLogger()")
	}
	if logger != Log {
		t.Error("GetLogger() should return the global Log instance")
	}
}

func TestInitLogger_SetsGlobalLog(t *testing.T) {
	old := Log
	defer func() { Log = old }()

	Log = nil
	Config = &GlobalConfig{}
	Config.Log.Default.Level = "debug"
	Config.Log.Default.Format = "text"
	// 使用当前进程可写的固定目录，避免 t.TempDir() 在 Windows 上
	// 因日志文件被轮转器持有而清理失败。
	Config.Log.File.Path = "runtime/log"
	Config.Log.File.Filename = "test_init"

	InitLogger()

	if Log == nil {
		t.Fatal("InitLogger() did not set global Log")
	}
	// 确保可以正常写日志
	Log.Debugf("test debug from InitLogger")
	Log.Infof("test info from InitLogger")
}
