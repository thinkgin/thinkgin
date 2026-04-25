package app

import (
	"bytes"
	"testing"

	"github.com/sirupsen/logrus"
)

func TestLogrusAdapter_ImplementsLogger(t *testing.T) {
	var _ Logger = (*LogrusAdapter)(nil) // 编译期检查
}

func TestLogrusAdapter_NilCreatesDefault(t *testing.T) {
	a := NewLogrusAdapter(nil)
	if a.L == nil {
		t.Fatal("NewLogrusAdapter(nil) should create a default logrus.Logger")
	}
}

func TestLogrusAdapter_WritesOutput(t *testing.T) {
	var buf bytes.Buffer
	l := logrus.New()
	l.SetOutput(&buf)
	l.SetLevel(logrus.DebugLevel)
	l.SetFormatter(&logrus.TextFormatter{DisableTimestamp: true})

	adapter := NewLogrusAdapter(l)

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

func TestLogrusAdapter_WithFields(t *testing.T) {
	var buf bytes.Buffer
	l := logrus.New()
	l.SetOutput(&buf)
	l.SetFormatter(&logrus.JSONFormatter{})

	adapter := NewLogrusAdapter(l)
	child := adapter.WithFields(map[string]any{"key1": "val1"})

	child.Info("with fields")

	out := buf.String()
	if !bytes.Contains([]byte(out), []byte("key1")) {
		t.Errorf("output should contain field key1, got: %s", out)
	}
	if !bytes.Contains([]byte(out), []byte("val1")) {
		t.Errorf("output should contain field val1, got: %s", out)
	}
}

func TestLogrusAdapter_WithFieldsDoesNotMutateParent(t *testing.T) {
	adapter := NewLogrusAdapter(nil)
	_ = adapter.WithFields(map[string]any{"extra": true})

	if len(adapter.fields) != 0 {
		t.Error("WithFields should not mutate the parent adapter")
	}
}
