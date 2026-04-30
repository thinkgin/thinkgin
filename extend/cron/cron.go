// Package cron 提供基于 robfig/cron 的定时任务调度器。
//
// 设计目标：
//   - 统一管理所有定时任务，支持 cron 表达式和 @every 语法。
//   - 任务 panic 自动恢复并记录日志，不影响其他任务。
//   - 支持优雅停机：调用 Stop() 等待正在执行的任务完成。
//   - 通过 Register 注册任务，Start 启动调度，Stop 停止。
//
// 用法：
//
//	cron.Register("cache-cleanup", "0 */5 * * *", func() {
//	    // 每 5 分钟清理过期缓存
//	})
//	cron.Start()
//	defer cron.Stop()
package cron

import (
	"fmt"
	"sync"

	"thinkgin/app"

	robfigcron "github.com/robfig/cron/v3"
)

// Job 描述一个定时任务。
type Job struct {
	Name     string // 任务名称，用于日志标识
	Schedule string // cron 表达式或 @every 1h 等
	Fn       func() // 任务函数
}

var (
	mu       sync.Mutex
	jobs     []Job
	instance *robfigcron.Cron
)

// Register 注册一个定时任务。必须在 Start() 之前调用。
func Register(name, schedule string, fn func()) {
	mu.Lock()
	defer mu.Unlock()
	jobs = append(jobs, Job{Name: name, Schedule: schedule, Fn: fn})
}

// Start 启动调度器。幂等，多次调用只启动一次。
func Start() error {
	mu.Lock()
	defer mu.Unlock()

	if instance != nil {
		return nil
	}

	logger := app.GetLogger()

	instance = robfigcron.New(
		robfigcron.WithSeconds(),
		robfigcron.WithChain(
			robfigcron.Recover(robfigcron.DefaultLogger),
		),
	)

	for _, j := range jobs {
		job := j // 闭包捕获
		_, err := instance.AddFunc(job.Schedule, func() {
			if logger != nil {
				logger.Infof("[cron] 执行任务: %s", job.Name)
			}
			job.Fn()
		})
		if err != nil {
			return fmt.Errorf("cron: 注册任务 %q 失败: %w", job.Name, err)
		}
		if logger != nil {
			logger.Infof("[cron] 已注册任务: %s (%s)", job.Name, job.Schedule)
		}
	}

	instance.Start()
	if logger != nil {
		logger.Infof("[cron] 调度器已启动，共 %d 个任务", len(jobs))
	}
	return nil
}

// Stop 优雅停止调度器，等待正在执行的任务完成。
func Stop() {
	mu.Lock()
	defer mu.Unlock()

	if instance == nil {
		return
	}
	ctx := instance.Stop()
	<-ctx.Done()
	instance = nil

	logger := app.GetLogger()
	if logger != nil {
		logger.Info("[cron] 调度器已停止")
	}
}

// Jobs 返回已注册的任务列表（只读副本）。
func Jobs() []Job {
	mu.Lock()
	defer mu.Unlock()
	out := make([]Job, len(jobs))
	copy(out, jobs)
	return out
}

// Reset 清空所有注册的任务和调度器实例（仅测试用）。
func Reset() {
	mu.Lock()
	defer mu.Unlock()
	if instance != nil {
		ctx := instance.Stop()
		<-ctx.Done()
		instance = nil
	}
	jobs = nil
}
