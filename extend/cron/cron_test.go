package cron

import (
	"sync/atomic"
	"testing"
	"time"
)

func TestRegisterAndJobs(t *testing.T) {
	Reset()
	Register("task1", "@every 1s", func() {})
	Register("task2", "@every 2s", func() {})

	jobs := Jobs()
	if len(jobs) != 2 {
		t.Fatalf("Jobs() = %d, want 2", len(jobs))
	}
	if jobs[0].Name != "task1" {
		t.Errorf("jobs[0].Name = %q, want %q", jobs[0].Name, "task1")
	}
}

func TestStartAndStop(t *testing.T) {
	Reset()

	var counter int64
	Register("counter", "@every 1s", func() {
		atomic.AddInt64(&counter, 1)
	})

	if err := Start(); err != nil {
		t.Fatalf("Start() error: %v", err)
	}

	// 等待足够时间让任务至少执行一次
	time.Sleep(1500 * time.Millisecond)

	Stop()

	val := atomic.LoadInt64(&counter)
	if val < 1 {
		t.Errorf("counter = %d, want >= 1", val)
	}
}

func TestStartIdempotent(t *testing.T) {
	Reset()
	Register("noop", "@every 10s", func() {})

	_ = Start()
	err := Start() // 二次调用
	if err != nil {
		t.Errorf("second Start() should be no-op, got: %v", err)
	}
	Stop()
}

func TestStartInvalidSchedule(t *testing.T) {
	Reset()
	Register("bad", "not-a-cron-expression", func() {})

	err := Start()
	if err == nil {
		t.Error("Start() should fail with invalid schedule")
		Stop()
	}
}

func TestReset(t *testing.T) {
	Reset()
	Register("x", "@every 1s", func() {})
	_ = Start()
	Reset()

	if len(Jobs()) != 0 {
		t.Error("Reset should clear all jobs")
	}
}
