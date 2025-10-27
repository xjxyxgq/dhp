package global

import (
	"sync"
	"cmdb-rpc/internal/scheduler"
)

var (
	taskSchedulerInstance *scheduler.TaskScheduler
	taskSchedulerOnce     sync.Once
	taskSchedulerMutex    sync.RWMutex
)

// SetTaskScheduler 设置全局TaskScheduler实例
func SetTaskScheduler(ts *scheduler.TaskScheduler) {
	taskSchedulerMutex.Lock()
	defer taskSchedulerMutex.Unlock()
	taskSchedulerInstance = ts
}

// GetTaskScheduler 获取全局TaskScheduler实例
func GetTaskScheduler() *scheduler.TaskScheduler {
	taskSchedulerMutex.RLock()
	defer taskSchedulerMutex.RUnlock()
	return taskSchedulerInstance
}