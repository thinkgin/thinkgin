package event

import (
	"sync/atomic"
	"testing"
	"time"
)

func TestOnAndFire(t *testing.T) {
	Reset()

	var called int
	On("test.event", func(p Payload) {
		called++
		if p["key"] != "value" {
			t.Errorf("payload[key] = %v, want %q", p["key"], "value")
		}
	})

	Fire("test.event", Payload{"key": "value"})

	if called != 1 {
		t.Errorf("called = %d, want 1", called)
	}
}

func TestMultipleListeners_OrderPreserved(t *testing.T) {
	Reset()

	var order []int
	On("test.order", func(p Payload) { order = append(order, 1) })
	On("test.order", func(p Payload) { order = append(order, 2) })
	On("test.order", func(p Payload) { order = append(order, 3) })

	Fire("test.order", nil)

	if len(order) != 3 || order[0] != 1 || order[1] != 2 || order[2] != 3 {
		t.Errorf("order = %v, want [1 2 3]", order)
	}
}

func TestFire_PanicRecovery(t *testing.T) {
	Reset()

	var afterPanic bool
	On("test.panic", func(p Payload) { panic("boom") })
	On("test.panic", func(p Payload) { afterPanic = true })

	Fire("test.panic", nil) // 不应 panic

	if !afterPanic {
		t.Error("second listener should still execute after first panics")
	}
}

func TestFireAsync(t *testing.T) {
	Reset()

	var counter int64
	On("test.async", func(p Payload) { atomic.AddInt64(&counter, 1) })
	On("test.async", func(p Payload) { atomic.AddInt64(&counter, 1) })

	FireAsync("test.async", nil)
	time.Sleep(50 * time.Millisecond)

	if atomic.LoadInt64(&counter) != 2 {
		t.Errorf("counter = %d, want 2", atomic.LoadInt64(&counter))
	}
}

func TestOff(t *testing.T) {
	Reset()

	var called bool
	On("test.off", func(p Payload) { called = true })
	Off("test.off")
	Fire("test.off", nil)

	if called {
		t.Error("listener should not be called after Off")
	}
}

func TestListenerCount(t *testing.T) {
	Reset()

	On("test.count", func(p Payload) {})
	On("test.count", func(p Payload) {})

	if ListenerCount("test.count") != 2 {
		t.Errorf("ListenerCount = %d, want 2", ListenerCount("test.count"))
	}
	if ListenerCount("test.none") != 0 {
		t.Errorf("ListenerCount for unregistered = %d, want 0", ListenerCount("test.none"))
	}
}

func TestLifecycleEvents(t *testing.T) {
	Reset()

	events := []Name{AppStarting, AppStarted, AppStopping, AppStopped}
	var fired []Name
	for _, e := range events {
		e := e
		On(e, func(p Payload) { fired = append(fired, e) })
	}

	for _, e := range events {
		Fire(e, nil)
	}

	if len(fired) != 4 {
		t.Fatalf("fired = %d events, want 4", len(fired))
	}
	for i, e := range events {
		if fired[i] != e {
			t.Errorf("fired[%d] = %q, want %q", i, fired[i], e)
		}
	}
}

func TestReset(t *testing.T) {
	Reset()
	On("test.reset", func(p Payload) {})
	Reset()

	if ListenerCount("test.reset") != 0 {
		t.Error("Reset should clear all listeners")
	}
}
