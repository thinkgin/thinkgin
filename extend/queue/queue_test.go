package queue

import (
	"context"
	"fmt"
	"sync/atomic"
	"testing"
	"time"
)

func TestMemoryQueue_PushAndConsume(t *testing.T) {
	q := NewMemoryQueue(10, 2)
	var counter int64

	q.Start()

	for i := 0; i < 5; i++ {
		_ = q.Push(Task{
			Name: fmt.Sprintf("task-%d", i),
			Handler: func(ctx context.Context, payload map[string]interface{}) error {
				atomic.AddInt64(&counter, 1)
				return nil
			},
		})
	}

	time.Sleep(100 * time.Millisecond)
	q.Stop()

	if atomic.LoadInt64(&counter) != 5 {
		t.Errorf("counter = %d, want 5", atomic.LoadInt64(&counter))
	}
}

func TestMemoryQueue_BufferFull(t *testing.T) {
	q := NewMemoryQueue(1, 1) // bufSize=1，不启动 worker

	_ = q.Push(Task{Name: "first", Handler: func(ctx context.Context, p map[string]interface{}) error { return nil }})
	err := q.Push(Task{Name: "second", Handler: func(ctx context.Context, p map[string]interface{}) error { return nil }})

	if err == nil {
		t.Error("should return error when buffer is full")
	}

	// 清理
	q.Start()
	q.Stop()
}

func TestMemoryQueue_PanicRecovery(t *testing.T) {
	q := NewMemoryQueue(10, 1)

	var afterPanic int64
	q.Start()

	_ = q.Push(Task{
		Name: "panic-task",
		Handler: func(ctx context.Context, p map[string]interface{}) error {
			panic("boom")
		},
	})
	_ = q.Push(Task{
		Name: "normal-task",
		Handler: func(ctx context.Context, p map[string]interface{}) error {
			atomic.AddInt64(&afterPanic, 1)
			return nil
		},
	})

	time.Sleep(100 * time.Millisecond)
	q.Stop()

	if atomic.LoadInt64(&afterPanic) != 1 {
		t.Error("normal task should execute after panic task")
	}
}

func TestMemoryQueue_HandlerError(t *testing.T) {
	q := NewMemoryQueue(10, 1)
	q.Start()

	_ = q.Push(Task{
		Name: "error-task",
		Handler: func(ctx context.Context, p map[string]interface{}) error {
			return fmt.Errorf("intentional error")
		},
	})

	time.Sleep(50 * time.Millisecond)
	q.Stop()
	// 不应 panic，错误仅记录日志
}

func TestNewMemoryQueue_DefaultValues(t *testing.T) {
	q := NewMemoryQueue(0, 0)
	if cap(q.ch) != 100 {
		t.Errorf("bufSize = %d, want 100 (default)", cap(q.ch))
	}
	if q.workers != 1 {
		t.Errorf("workers = %d, want 1 (default)", q.workers)
	}
	q.Start()
	q.Stop()
}
