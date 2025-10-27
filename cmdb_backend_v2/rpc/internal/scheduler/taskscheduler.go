package scheduler

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"cmdb-rpc/cmpool"
	"cmdb-rpc/internal/model"
	"cmdb-rpc/internal/service"
	"cmdb-rpc/internal/svc"

	"github.com/robfig/cron/v3"
	"github.com/zeromicro/go-zero/core/logx"
)

// TaskScheduler 定时任务调度器
type TaskScheduler struct {
	ctx                         context.Context
	svcCtx                      *svc.ServiceContext
	hardwareVerificationService service.HardwareVerificationService
	scheduledTaskModel          model.ScheduledHardwareVerificationModel
	scheduledTaskHistoryModel   model.ScheduledTaskExecutionHistoryModel
	hardwareVerificationModel   model.HardwareResourceVerificationModel
	cron                        *cron.Cron
	jobs                        map[int64]cron.EntryID // 任务ID -> cron Entry ID的映射
	mutex                       sync.RWMutex
	logger                      logx.Logger
	stopChan                    chan struct{}
	wg                          sync.WaitGroup
}

// NewTaskScheduler 创建新的任务调度器
func NewTaskScheduler(svcCtx *svc.ServiceContext) *TaskScheduler {
	ctx := context.Background()
	hwService := service.NewHardwareVerificationService(ctx, svcCtx.Config, svcCtx.HardwareResourceVerificationModel)

	return &TaskScheduler{
		ctx:                         ctx,
		svcCtx:                      svcCtx,
		hardwareVerificationService: hwService,
		scheduledTaskModel:          svcCtx.ScheduledTaskModel,
		scheduledTaskHistoryModel:   svcCtx.ScheduledTaskHistoryModel,
		hardwareVerificationModel:   svcCtx.HardwareResourceVerificationModel,
		cron:                        cron.New(), // 使用标准的5字段cron格式
		jobs:                        make(map[int64]cron.EntryID),
		logger:                      logx.WithContext(context.Background()),
		stopChan:                    make(chan struct{}),
	}
}

// Start 启动调度器
func (ts *TaskScheduler) Start() error {
	ts.logger.Info("Starting Task Scheduler...")

	// 加载所有启用的定时任务
	err := ts.loadScheduledTasks()
	if err != nil {
		return fmt.Errorf("failed to load scheduled tasks: %v", err)
	}

	// 启动cron调度器
	ts.cron.Start()

	// 启动定期检查任务状态的协程
	ts.wg.Add(1)
	go ts.periodicCheck()

	ts.logger.Info("Task Scheduler started successfully")
	return nil
}

// Stop 停止调度器
func (ts *TaskScheduler) Stop() error {
	ts.logger.Info("Stopping Task Scheduler...")

	close(ts.stopChan)
	ts.cron.Stop()
	ts.wg.Wait()

	ts.logger.Info("Task Scheduler stopped")
	return nil
}

// loadScheduledTasks 加载所有启用的定时任务
func (ts *TaskScheduler) loadScheduledTasks() error {
	tasks, err := ts.scheduledTaskModel.FindAll("", true) // 只加载启用的任务
	if err != nil {
		return err
	}

	for _, task := range tasks {
		err := ts.addCronJob(task)
		if err != nil {
			ts.logger.Errorf("Failed to add cron job for task %d: %v", task.Id, err)
			continue
		}
		ts.logger.Infof("Loaded scheduled task: %s (ID: %d)", task.TaskName, task.Id)
	}

	return nil
}

// AddTask 添加新的定时任务
func (ts *TaskScheduler) AddTask(task *model.ScheduledHardwareVerification) error {
	ts.mutex.Lock()
	defer ts.mutex.Unlock()

	if task.IsEnabled > 0 {
		return ts.addCronJob(task)
	}

	return nil
}

// RemoveTask 移除定时任务
func (ts *TaskScheduler) RemoveTask(taskId int64) error {
	ts.mutex.Lock()
	defer ts.mutex.Unlock()

	if entryID, exists := ts.jobs[taskId]; exists {
		ts.cron.Remove(entryID)
		delete(ts.jobs, taskId)
		ts.logger.Infof("Removed scheduled task: %d", taskId)
	}

	return nil
}

// UpdateTask 更新定时任务
func (ts *TaskScheduler) UpdateTask(task *model.ScheduledHardwareVerification) error {
	// 先移除旧任务
	ts.RemoveTask(task.Id)

	// 如果任务启用，添加新任务
	if task.IsEnabled > 0 {
		return ts.AddTask(task)
	}

	return nil
}

// addCronJob 添加cron任务
func (ts *TaskScheduler) addCronJob(task *model.ScheduledHardwareVerification) error {
	// 解析IP列表
	var hostIpList []string
	if err := json.Unmarshal([]byte(task.HostIpList), &hostIpList); err != nil {
		return fmt.Errorf("failed to parse host IP list: %v", err)
	}

	// 创建任务执行函数
	jobFunc := func() {
		ts.executeScheduledTask(task, hostIpList)
	}

	// 添加到cron调度器
	entryID, err := ts.cron.AddFunc(task.CronExpression, jobFunc)
	if err != nil {
		return fmt.Errorf("failed to add cron job: %v", err)
	}

	ts.jobs[task.Id] = entryID

	// 计算下次执行时间
	nextTime := ts.cron.Entry(entryID).Next
	if !nextTime.IsZero() {
		err = ts.scheduledTaskModel.UpdateExecutionTime(task.Id, nil, &nextTime)
		if err != nil {
			ts.logger.Errorf("Failed to update next execution time for task %d: %v", task.Id, err)
		}
	}

	return nil
}

// executeScheduledTask 执行定时任务
func (ts *TaskScheduler) executeScheduledTask(scheduledTask *model.ScheduledHardwareVerification, hostIpList []string) {
	ts.logger.Infof("Executing scheduled task: %s (ID: %d)", scheduledTask.TaskName, scheduledTask.Id)

	// 记录执行开始时间
	executionTime := time.Now()

	// 创建硬件资源验证请求
	req := &cmpool.HardwareResourceVerificationReq{
		HostIpList:    hostIpList,
		ResourceType:  scheduledTask.ResourceType,
		TargetPercent: int32(scheduledTask.TargetPercent),
		Duration:      int32(scheduledTask.Duration),
		ScriptParams:  scheduledTask.ScriptParams.String,
		ForceExecution: func() bool {
			if scheduledTask.ForceExecution > 0 {
				return true
			} else {
				return false
			}
		}(),
	}

	// 调用硬件资源验证服务
	resp, err := ts.hardwareVerificationService.ExecuteVerification(req)

	// 创建执行历史记录
	history := &model.ScheduledTaskExecutionHistory{
		ScheduledTaskId: scheduledTask.Id,
		ExecutionTaskId: "",
		ExecutionTime:   executionTime,
		ExecutionStatus: "failed",
		TotalHosts:      int64(len(hostIpList)),
		SuccessHosts:    0,
		FailedHosts:     int64(len(hostIpList)),
		ErrorMessage:    sql.NullString{},
	}

	if err != nil {
		history.ErrorMessage = sql.NullString{String: err.Error(), Valid: true}
		ts.logger.Errorf("Failed to execute scheduled task %d: %v", scheduledTask.Id, err)
	} else if resp.Success {
		history.ExecutionTaskId = resp.TaskId
		history.ExecutionStatus = "success"
		ts.logger.Infof("Successfully started scheduled task %d, task ID: %s", scheduledTask.Id, resp.TaskId)

		// 异步统计执行结果
		go ts.updateExecutionStats(scheduledTask.Id, resp.TaskId)
	} else {
		history.ErrorMessage = sql.NullString{String: resp.Message, Valid: true}
		ts.logger.Errorf("Scheduled task %d execution failed: %s", scheduledTask.Id, resp.Message)
	}

	// 保存执行历史
	_, err = ts.scheduledTaskHistoryModel.Insert(ts.ctx, history)
	if err != nil {
		ts.logger.Errorf("Failed to save execution history for task %d: %v", scheduledTask.Id, err)
	}

	// 更新上次执行时间
	err = ts.scheduledTaskModel.UpdateExecutionTime(scheduledTask.Id, &executionTime, nil)
	if err != nil {
		ts.logger.Errorf("Failed to update last execution time for task %d: %v", scheduledTask.Id, err)
	}

	// 计算并更新下次执行时间
	ts.updateNextExecutionTime(scheduledTask.Id)
}

// updateExecutionStats 更新执行统计信息
func (ts *TaskScheduler) updateExecutionStats(scheduledTaskId int64, executionTaskId string) {
	// 等待一段时间让验证任务开始执行
	time.Sleep(30 * time.Second)

	// 定期检查任务执行状态，直到完成
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	maxWait := 2 * time.Hour // 最长等待2小时
	timeout := time.After(maxWait)

	for {
		select {
		case <-timeout:
			ts.logger.Infof("Timeout waiting for task %s to complete", executionTaskId)
			return
		case <-ticker.C:
			// 查询任务执行状态
			records, err := ts.hardwareVerificationModel.FindByTaskId(ts.ctx, executionTaskId)
			if err != nil {
				ts.logger.Errorf("Failed to query task status for %s: %v", executionTaskId, err)
				continue
			}

			totalHosts := len(records)
			successHosts := 0
			failedHosts := 0
			runningHosts := 0

			for _, record := range records {
				switch record.ExecutionStatus {
				case "completed":
					successHosts++
				case "failed":
					failedHosts++
				case "running", "pending":
					runningHosts++
				}
			}

			// 如果还有正在运行的任务，继续等待
			if runningHosts > 0 {
				continue
			}

			// 所有任务都完成了，更新统计信息
			status := "success"
			if failedHosts > 0 && successHosts == 0 {
				status = "failed"
			} else if failedHosts > 0 {
				status = "partial"
			}

			// 查找对应的历史记录并更新
			histories, err := ts.scheduledTaskHistoryModel.FindByTaskId(scheduledTaskId, 1)
			if err != nil || len(histories) == 0 {
				ts.logger.Errorf("Failed to find execution history for task %d", scheduledTaskId)
				return
			}

			err = ts.scheduledTaskHistoryModel.UpdateExecutionStatus(
				histories[0].Id, status, int32(totalHosts), int32(successHosts), int32(failedHosts), "")
			if err != nil {
				ts.logger.Errorf("Failed to update execution status: %v", err)
			}

			ts.logger.Infof("Updated execution stats for task %d: total=%d, success=%d, failed=%d, status=%s",
				scheduledTaskId, totalHosts, successHosts, failedHosts, status)
			return
		}
	}
}

// updateNextExecutionTime 更新下次执行时间
func (ts *TaskScheduler) updateNextExecutionTime(taskId int64) {
	ts.mutex.RLock()
	entryID, exists := ts.jobs[taskId]
	ts.mutex.RUnlock()

	if !exists {
		return
	}

	entry := ts.cron.Entry(entryID)
	if !entry.Next.IsZero() {
		err := ts.scheduledTaskModel.UpdateExecutionTime(taskId, nil, &entry.Next)
		if err != nil {
			ts.logger.Errorf("Failed to update next execution time for task %d: %v", taskId, err)
		}
	}
}

// periodicCheck 定期检查任务状态
func (ts *TaskScheduler) periodicCheck() {
	defer ts.wg.Done()

	ticker := time.NewTicker(5 * time.Minute) // 每5分钟检查一次
	defer ticker.Stop()

	for {
		select {
		case <-ts.stopChan:
			return
		case <-ticker.C:
			ts.checkTaskStatus()
		}
	}
}

// checkTaskStatus 检查任务状态，重新加载可能有变化的任务
func (ts *TaskScheduler) checkTaskStatus() {
	// 获取所有启用的任务
	tasks, err := ts.scheduledTaskModel.FindAll("", true)
	if err != nil {
		ts.logger.Errorf("Failed to check task status: %v", err)
		return
	}

	ts.mutex.Lock()
	defer ts.mutex.Unlock()

	// 检查是否有新任务需要添加
	for _, task := range tasks {
		if _, exists := ts.jobs[task.Id]; !exists && (task.IsEnabled > 0) {
			err := ts.addCronJob(task)
			if err != nil {
				ts.logger.Errorf("Failed to add new cron job for task %d: %v", task.Id, err)
			} else {
				ts.logger.Infof("Added new scheduled task: %s (ID: %d)", task.TaskName, task.Id)
			}
		}
	}

	// 检查是否有任务需要移除（被禁用或删除）
	taskMap := make(map[int64]*model.ScheduledHardwareVerification)
	for _, task := range tasks {
		taskMap[task.Id] = task
	}

	for taskId, entryID := range ts.jobs {
		if task, exists := taskMap[taskId]; !exists || (task.IsEnabled == 0) {
			ts.cron.Remove(entryID)
			delete(ts.jobs, taskId)
			ts.logger.Infof("Removed scheduled task: %d (disabled or deleted)", taskId)
		}
	}
}

// GetTaskStatus 获取任务状态信息
func (ts *TaskScheduler) GetTaskStatus() map[int64]cron.Entry {
	ts.mutex.RLock()
	defer ts.mutex.RUnlock()

	status := make(map[int64]cron.Entry)
	for taskId, entryID := range ts.jobs {
		entry := ts.cron.Entry(entryID)
		status[taskId] = entry
	}

	return status
}
