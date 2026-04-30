// Package event 提供 ThinkGin 的进程内事件发布/订阅系统。
//
// 设计目标：
//   - 同步发布（Fire）：所有订阅者按注册顺序依次执行，适用于生命周期钩子。
//   - 异步发布（FireAsync）：每个订阅者在独立 goroutine 执行，适用于通知类场景。
//   - 内置生命周期事件：AppStarting / AppStarted / AppStopping / AppStopped。
//   - 订阅者 panic 自动恢复并记录日志，不影响其他订阅者。
//
// 用法：
//
//	event.On(event.AppStarted, func(payload event.Payload) {
//	    fmt.Println("服务已启动:", payload["addr"])
//	})
//	event.Fire(event.AppStarted, event.Payload{"addr": ":8000"})
package event

import (
	"fmt"
	"sync"

	"thinkgin/app"
)

// Name 事件名称类型。
type Name string

// 内置生命周期事件。
const (
	AppStarting Name = "app.starting" // 服务即将启动
	AppStarted  Name = "app.started"  // 服务已启动
	AppStopping Name = "app.stopping" // 服务即将停止
	AppStopped  Name = "app.stopped"  // 服务已停止
)

// Payload 事件携带的数据。
type Payload map[string]interface{}

// Listener 事件监听器函数签名。
type Listener func(payload Payload)

var (
	mu        sync.RWMutex
	listeners = map[Name][]Listener{}
)

// On 订阅事件。同一事件可注册多个 Listener，按注册顺序执行。
func On(name Name, fn Listener) {
	mu.Lock()
	defer mu.Unlock()
	listeners[name] = append(listeners[name], fn)
}

// Fire 同步发布事件。所有 Listener 依次执行，任一 panic 会被恢复并记录日志。
func Fire(name Name, payload Payload) {
	mu.RLock()
	fns := make([]Listener, len(listeners[name]))
	copy(fns, listeners[name])
	mu.RUnlock()

	for _, fn := range fns {
		safeCall(name, fn, payload)
	}
}

// FireAsync 异步发布事件。每个 Listener 在独立 goroutine 中执行。
func FireAsync(name Name, payload Payload) {
	mu.RLock()
	fns := make([]Listener, len(listeners[name]))
	copy(fns, listeners[name])
	mu.RUnlock()

	for _, fn := range fns {
		go safeCall(name, fn, payload)
	}
}

// Off 移除指定事件的所有 Listener。
func Off(name Name) {
	mu.Lock()
	defer mu.Unlock()
	delete(listeners, name)
}

// Reset 清空所有事件订阅（仅测试用）。
func Reset() {
	mu.Lock()
	defer mu.Unlock()
	listeners = map[Name][]Listener{}
}

// ListenerCount 返回指定事件的订阅者数量。
func ListenerCount(name Name) int {
	mu.RLock()
	defer mu.RUnlock()
	return len(listeners[name])
}

func safeCall(name Name, fn Listener, payload Payload) {
	defer func() {
		if r := recover(); r != nil {
			logger := app.GetLogger()
			if logger != nil {
				logger.Errorf("[event] panic in listener for %q: %v", name, r)
			} else {
				fmt.Printf("[event] panic in listener for %q: %v\n", name, r)
			}
		}
	}()
	fn(payload)
}
