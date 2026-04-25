package app

import (
	"bytes"
	"log/slog"
	"testing"
)

func TestSlogAdapter_ImplementsLogger(t *testing.T) {
	var _ Logger = (*SlogAdapter)(nil) // 编译期检查
}

func TestSlogAdapter_WritesOutput(t *testing.T) {
	var buf bytes.Buffer
	l := slog.New(slog.NewTextHandler(&buf, &slog.HandlerOptions{Level: slog.LevelDebug}))
	adapter := NewSlogAdapter(l)

	adapter.Debug("d1")
	adapter.Debugf("d%d", 2)
	adapter.Info("i1")
	adapter.Infof("i%d", 2)
	adapter.Warn("w1")
	adapter.Warnf("w%d", 2)
	adapter.Error("e1")
	adapter.Errorf("e%d", 2)

	out := buf.String()
	for _, want := range []string{"d1", "d2", "i1", "i2", "w1", "w2", "e1", "e2"} {
		if !bytes.Contains([]byte(out), []byte(want)) {
			t.Errorf("output missing %q", want)
		}
	}
}

func TestSlogAdapter_WithFields(t *testing.T) {
	var buf bytes.Buffer
	l := slog.New(slog.NewTextHandler(&buf, &slog.HandlerOptions{Level: slog.LevelInfo}))
	adapter := NewSlogAdapter(l)

	child := adapter.WithFields(map[string]any{"service": "test"})
	child.Info("hello with fields")

	out := buf.String()
	if !bytes.Contains([]byte(out), []byte("service")) {
		t.Errorf("output should contain field 'service', got: %s", out)
	}
}
