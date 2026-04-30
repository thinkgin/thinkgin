// Package queue 提供轻量级异步任务队列。
//
// 设计目标：
//   - 内存队列：基于 Go channel，适用于单实例 / 本地开发。
//   - 统一 Queue 接口，后续可扩展 Redis / RabbitMQ 驱动。
//   - Worker 池并发消费，panic 自动恢复。
//   - 支持优雅停机：Stop() 等待所有已入队任务处理完成。
//
// 用法：
//
//	q := queue.NewMemoryQueue(10, 3) // bufSize=10, workers=3
//	q.Start()
//	q.Push(queue.Task{Name: "send-email", Payload: map[string]interface{}{...}, Handler: sendEmail})
//	defer q.Stop()
package queue

import (
	"context"
	"fmt"
	"sync"

	"thinkgin/app"
)

// Task 异步任务。
type Task struct {
	Name    string                 // 任务名称，用于日志
	Payload map[string]interface{} // 任务参数
	Handler func(ctx context.Context, payload map[string]interface{}) error
}

// Queue 异步任务队列接口。
type Queue interface {
	// Push 投递任务。非阻塞；缓冲区满时返回 error。
	Push(task Task) error
	// Start 启动 worker 消费。
	Start()
	// Stop 优雅停止，等待已入队任务处理完毕。
	Stop()
}

// MemoryQueue 基于 channel 的内存队列。
type MemoryQueue struct {
	ch      chan Task
	workers int
	wg      sync.WaitGroup
	once    sync.Once
	done    chan struct{}
}

// NewMemoryQueue 创建内存队列。bufSize 为 channel 缓冲大小，workers 为并发消费者数。
func NewMemoryQueue(bufSize, workers int) *MemoryQueue {
	if bufSize < 1 {
		bufSize = 100
	}
	if workers < 1 {
		workers = 1
	}
	return &MemoryQueue{
		ch:      make(chan Task, bufSize),
		workers: workers,
		done:    make(chan struct{}),
	}
}

// Push 投递任务到队列。
func (mq *MemoryQueue) Push(task Task) error {
	select {
	case mq.ch <- task:
		return nil
	default:
		return fmt.Errorf("queue: buffer full, task %q dropped", task.Name)
	}
}

// Start 启动 worker goroutine。
func (mq *MemoryQueue) Start() {
	for i := 0; i < mq.workers; i++ {
		mq.wg.Add(1)
		go mq.worker(i)
	}

	logger := app.GetLogger()
	if logger != nil {
		logger.Infof("[queue] 已启动 %d 个 worker (buf=%d)", mq.workers, cap(mq.ch))
	}
}

// Stop 关闭队列并等待所有 worker 完成。
func (mq *MemoryQueue) Stop() {
	mq.once.Do(func() {
		close(mq.done)
		close(mq.ch)
	})
	mq.wg.Wait()

	logger := app.GetLogger()
	if logger != nil {
		logger.Info("[queue] 已停止")
	}
}

func (mq *MemoryQueue) worker(id int) {
	defer mq.wg.Done()
	logger := app.GetLogger()

	for task := range mq.ch {
		mq.executeTask(id, task, logger)
	}
}

func (mq *MemoryQueue) executeTask(workerID int, task Task, logger interface{ Errorf(string, ...interface{}); Infof(string, ...interface{}) }) {
	defer func() {
		if r := recover(); r != nil {
			if logger != nil {
				logger.Errorf("[queue] worker-%d panic in task %q: %v", workerID, task.Name, r)
			}
		}
	}()

	ctx := context.Background()
	if err := task.Handler(ctx, task.Payload); err != nil {
		if logger != nil {
			logger.Errorf("[queue] worker-%d task %q failed: %v", workerID, task.Name, err)
		}
	}
}
