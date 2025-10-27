package scheduler

import (
	"context"
	"database/sql"
	"fmt"
	"sync"
	"time"

	"cmdb-rpc/internal/datasource/elasticsearch"
	"cmdb-rpc/internal/model"
	"cmdb-rpc/internal/svc"

	"github.com/robfig/cron/v3"
	"github.com/zeromicro/go-zero/core/logx"
)

// EsSyncScheduler ES数据同步任务调度器
type EsSyncScheduler struct {
	cron      *cron.Cron
	svcCtx    *svc.ServiceContext
	tasks     map[uint64]cron.EntryID // taskId -> cronEntryId
	taskMutex sync.RWMutex
	logger    logx.Logger
}

// NewEsSyncScheduler 创建新的ES同步任务调度器
func NewEsSyncScheduler(svcCtx *svc.ServiceContext) *EsSyncScheduler {
	return &EsSyncScheduler{
		cron:   cron.New(cron.WithSeconds()), // 支持秒级精度
		svcCtx: svcCtx,
		tasks:  make(map[uint64]cron.EntryID),
		logger: logx.WithContext(context.Background()),
	}
}

// Start 启动调度器并加载所有已启用的任务
func (s *EsSyncScheduler) Start() error {
	s.logger.Info("启动ES数据同步调度器...")

	// 从数据库加载所有已启用的任务
	ctx := context.Background()
	tasks, err := s.svcCtx.ExternalSyncTaskConfigModel.FindAll(ctx, true)
	if err != nil {
		s.logger.Errorf("加载ES同步任务失败: %v", err)
		return err
	}

	// 注册所有已启用的任务
	for _, task := range tasks {
		if err := s.RegisterTask(task); err != nil {
			s.logger.Errorf("注册任务失败: TaskId=%d, TaskName=%s, Error=%v",
				task.Id, task.TaskName, err)
			continue
		}
		s.logger.Infof("成功注册任务: TaskId=%d, TaskName=%s, Cron=%s",
			task.Id, task.TaskName, task.CronExpression)
	}

	// 启动cron调度器
	s.cron.Start()
	s.logger.Infof("ES数据同步调度器已启动，共注册 %d 个任务", len(s.tasks))
	return nil
}

// Stop 停止调度器
func (s *EsSyncScheduler) Stop() {
	s.logger.Info("停止ES数据同步调度器...")
	ctx := s.cron.Stop()
	<-ctx.Done()
	s.logger.Info("ES数据同步调度器已停止")
}

// RegisterTask 注册任务到调度器
func (s *EsSyncScheduler) RegisterTask(task *model.ExternalSyncTaskConfig) error {
	s.taskMutex.Lock()
	defer s.taskMutex.Unlock()

	// 检查任务是否已注册
	if _, exists := s.tasks[task.Id]; exists {
		s.logger.Infof("任务已存在，先注销: TaskId=%d", task.Id)
		s.unregisterTaskInternal(task.Id)
	}

	// 创建任务执行函数
	job := s.createJob(task)

	// 添加到cron调度器
	entryID, err := s.cron.AddFunc(task.CronExpression, job)
	if err != nil {
		return fmt.Errorf("添加cron任务失败: %v", err)
	}

	// 记录entryID
	s.tasks[task.Id] = entryID
	s.logger.Infof("任务已注册到调度器: TaskId=%d, TaskName=%s, EntryId=%d",
		task.Id, task.TaskName, entryID)

	return nil
}

// UnregisterTask 从调度器注销任务
func (s *EsSyncScheduler) UnregisterTask(taskId uint64) {
	s.taskMutex.Lock()
	defer s.taskMutex.Unlock()
	s.unregisterTaskInternal(taskId)
}

// unregisterTaskInternal 内部注销任务（不加锁）
func (s *EsSyncScheduler) unregisterTaskInternal(taskId uint64) {
	entryID, exists := s.tasks[taskId]
	if !exists {
		s.logger.Infof("任务未注册，无需注销: TaskId=%d", taskId)
		return
	}

	// 从cron调度器移除
	s.cron.Remove(entryID)
	delete(s.tasks, taskId)
	s.logger.Infof("任务已从调度器注销: TaskId=%d, EntryId=%d", taskId, entryID)
}

// createJob 创建任务执行函数
func (s *EsSyncScheduler) createJob(task *model.ExternalSyncTaskConfig) func() {
	return func() {
		ctx := context.Background()
		startTime := time.Now()
		s.logger.Infof("开始执行定时同步任务: TaskId=%d, TaskName=%s", task.Id, task.TaskName)

		// 重新从数据库读取任务配置，确保使用最新配置
		latestTask, err := s.svcCtx.ExternalSyncTaskConfigModel.FindOne(ctx, task.Id)
		if err != nil {
			s.logger.Errorf("读取任务配置失败: TaskId=%d, Error=%v", task.Id, err)
			return
		}

		// 检查任务是否仍然启用
		if latestTask.IsEnabled != 1 {
			s.logger.Infof("任务已禁用，跳过执行: TaskId=%d", task.Id)
			return
		}

		// 从 hosts_pool 获取所有主机IP列表
		hostIpList, err := s.svcCtx.HostsPoolModel.FindAllHostIPs(ctx)
		if err != nil {
			s.logger.Errorf("获取主机列表失败: TaskId=%d, Error=%v", task.Id, err)
			return
		}

		if len(hostIpList) == 0 {
			s.logger.Infof("主机列表为空，跳过执行: TaskId=%d", task.Id)
			return
		}

		s.logger.Infof("定时任务触发: TaskId=%d, TaskName=%s, Cron=%s, HostCount=%d",
			latestTask.Id, latestTask.TaskName, latestTask.CronExpression, len(hostIpList))

		// 执行同步任务
		executionId, success, message := s.executeSync(ctx, latestTask, hostIpList, startTime)

		if success {
			s.logger.Infof("定时任务执行成功: TaskId=%d, ExecutionId=%d, Message=%s",
				task.Id, executionId, message)
		} else {
			s.logger.Errorf("定时任务执行失败: TaskId=%d, Message=%s",
				task.Id, message)
		}
	}
}

// executeSync 执行ES数据同步
func (s *EsSyncScheduler) executeSync(ctx context.Context, task *model.ExternalSyncTaskConfig, hostIpList []string, startTime time.Time) (int64, bool, string) {
	s.logger.Infof("开始执行ES同步: TaskId=%d, TaskName=%s, 主机数=%d", task.Id, task.TaskName, len(hostIpList))

	// 1. 创建执行记录
	executionLog := &model.ExternalSyncExecutionLog{
		TaskId:          task.Id,
		TaskName:        task.TaskName,
		ExecutionTime:   startTime,
		ExecutionStatus: "running",
		TotalHosts:      int64(len(hostIpList)),
		QueryTimeRange:  sql.NullString{String: task.QueryTimeRange, Valid: true},
	}

	result, err := s.svcCtx.ExternalSyncExecutionLogModel.Insert(ctx, executionLog)
	if err != nil {
		s.logger.Errorf("创建执行记录失败: %v", err)
		return 0, false, fmt.Sprintf("创建执行记录失败: %v", err)
	}

	executionId, _ := result.LastInsertId()

	// 2. 获取ES配置
	esEndpoint := s.svcCtx.Config.ESDataSource.DefaultEndpoint
	if task.EsEndpoint.Valid && task.EsEndpoint.String != "" {
		esEndpoint = task.EsEndpoint.String
	}

	esIndexPattern := s.svcCtx.Config.ESDataSource.DefaultIndexPattern
	if task.EsIndexPattern.Valid && task.EsIndexPattern.String != "" {
		esIndexPattern = task.EsIndexPattern.String
	}

	// 3. 创建ES客户端
	esClient := elasticsearch.NewESClient(esEndpoint, s.svcCtx.Config.ESDataSource.TimeoutSeconds)

	// 4. 并发同步数据
	successCount, failedCount, notInDatasourceCount := s.syncHosts(ctx, esClient, esIndexPattern, hostIpList, task.QueryTimeRange, executionId)

	// 5. 更新执行记录
	duration := time.Since(startTime).Milliseconds()
	executionStatus := "success"
	if failedCount > 0 {
		if successCount > 0 {
			executionStatus = "partial"
		} else {
			executionStatus = "failed"
		}
	}

	if err := s.svcCtx.ExternalSyncExecutionLogModel.UpdateExecutionResult(
		ctx, uint64(executionId), executionStatus, successCount, failedCount, notInDatasourceCount, duration); err != nil {
		s.logger.Errorf("更新执行记录状态失败: %v", err)
	}

	message := fmt.Sprintf("同步完成: 成功%d个, 失败%d个, 数据源中不存在%d个", successCount, failedCount, notInDatasourceCount)
	s.logger.Infof("ES同步执行完成: TaskId=%d, ExecutionId=%d, %s", task.Id, executionId, message)

	return executionId, true, message
}

// syncHosts 同步主机数据
func (s *EsSyncScheduler) syncHosts(ctx context.Context, esClient *elasticsearch.ESClient, indexPattern string, hostIpList []string, timeRange string, executionId int64) (successCount, failedCount, notInDatasourceCount int) {
	const maxConcurrency = 10
	sem := make(chan struct{}, maxConcurrency)
	var wg sync.WaitGroup
	var mu sync.Mutex

	for _, hostIP := range hostIpList {
		wg.Add(1)
		sem <- struct{}{}

		go func(ip string) {
			defer wg.Done()
			defer func() { <-sem }()

			// 查询ES数据
			metrics, err := esClient.QueryHostMetrics(ctx, indexPattern, ip, timeRange)
			if err != nil {
				if elasticsearch.IsNoDataError(err) {
					s.logger.Infof("主机 %s 在ES数据源中没有数据", ip)
					s.saveExecutionDetail(ctx, executionId, ip, "", "not_in_datasource", "ES数据源中无此主机数据", nil)
					mu.Lock()
					failedCount++
					mu.Unlock()
					return
				}
				s.logger.Errorf("查询主机 %s 的ES数据失败: %v", ip, err)
				s.saveExecutionDetail(ctx, executionId, ip, "", "failed", err.Error(), nil)
				mu.Lock()
				failedCount++
				mu.Unlock()
				return
			}

			// 检查主机是否在hosts_pool中
			hostInfo, err := s.svcCtx.HostsPoolModel.FindByIP(ctx, ip)
			if err == sql.ErrNoRows || hostInfo == nil {
				// 主机不在pool中，自动插入
				s.logger.Infof("主机 %s 不在hosts_pool中，准备插入", ip)
				poolId, err := s.svcCtx.HostsPoolModel.InsertIfNotExists(ctx, metrics.HostName, ip, "")
				if err != nil {
					s.logger.Errorf("插入主机 %s 到hosts_pool失败: %v", ip, err)
					s.saveExecutionDetail(ctx, executionId, ip, metrics.HostName, "failed", fmt.Sprintf("插入hosts_pool失败: %v", err), metrics)
					mu.Lock()
					failedCount++
					mu.Unlock()
					return
				}
				hostInfo = &model.HostsPool{Id: uint64(poolId), HostIp: ip, HostName: metrics.HostName}
				s.logger.Infof("成功插入新主机 %s 到hosts_pool, pool_id=%d", ip, poolId)
			} else if err != nil {
				s.logger.Errorf("查询主机 %s 的hosts_pool信息失败: %v", ip, err)
				s.saveExecutionDetail(ctx, executionId, ip, metrics.HostName, "failed", err.Error(), metrics)
				mu.Lock()
				failedCount++
				mu.Unlock()
				return
			}

			// 同步数据到server_resources
			// ES返回的数据说明：
			// - CPU: 百分比值（%）
			// - Memory: GB绝对值（已用内存）
			// - Disk: GB绝对值（已用磁盘）
			err = s.svcCtx.ServerResourcesModel.UpsertFromES(ctx, uint64(hostInfo.Id), ip,
				metrics.MaxMemory, metrics.MaxDisk, metrics.MaxCPU, // 绝对值: 已用内存GB, 已用磁盘GB, CPU负载%
				metrics.MaxCPU, metrics.AvgCPU, 0,                   // CPU百分比: max, avg, min
				metrics.MaxMemory, metrics.AvgMemory, 0,             // 内存百分比: max, avg, min（注意：这里是复用绝对值，实际应该转换为%）
				metrics.MaxDisk, metrics.AvgDisk, 0)                 // 磁盘百分比: max, avg, min（注意：这里是复用绝对值，实际应该转换为%）
			if err != nil {
				s.logger.Errorf("同步主机 %s 数据到server_resources失败: %v", ip, err)
				s.saveExecutionDetail(ctx, executionId, ip, hostInfo.HostName, "failed", err.Error(), metrics)
				mu.Lock()
				failedCount++
				mu.Unlock()
				return
			}

			// 保存执行详情
			s.saveExecutionDetail(ctx, executionId, ip, hostInfo.HostName, "success", "", metrics)
			mu.Lock()
			successCount++
			mu.Unlock()
		}(hostIP)
	}

	wg.Wait()
	return
}

// saveExecutionDetail 保存执行详情
func (s *EsSyncScheduler) saveExecutionDetail(ctx context.Context, executionId int64, hostIP, hostName, status, errorMsg string, metrics *elasticsearch.HostMetrics) {
	var maxCpu, avgCpu, maxMem, avgMem, maxDisk, avgDisk float64
	var dataPointCount int

	if metrics != nil {
		maxCpu = metrics.MaxCPU
		avgCpu = metrics.AvgCPU
		maxMem = metrics.MaxMemory
		avgMem = metrics.AvgMemory
		maxDisk = metrics.MaxDisk
		avgDisk = metrics.AvgDisk
		dataPointCount = metrics.DataPointCount
	}

	// 如果 hostName 为空但 metrics 存在，使用 metrics 中的 HostName
	if hostName == "" && metrics != nil {
		hostName = metrics.HostName
	}
	// 如果 hostName 仍然为空，使用 IP 作为主机名
	if hostName == "" {
		hostName = hostIP
	}

	detail := &model.ExternalSyncExecutionDetail{
		ExecutionId:    uint64(executionId),
		HostIp:         hostIP,
		HostName:       sql.NullString{String: hostName, Valid: hostName != ""},
		SyncStatus:     status,
		ErrorMessage:   sql.NullString{String: errorMsg, Valid: errorMsg != ""},
		MaxCpu:         sql.NullFloat64{Float64: maxCpu, Valid: metrics != nil},
		AvgCpu:         sql.NullFloat64{Float64: avgCpu, Valid: metrics != nil},
		MaxMemory:      sql.NullFloat64{Float64: maxMem, Valid: metrics != nil},
		AvgMemory:      sql.NullFloat64{Float64: avgMem, Valid: metrics != nil},
		MaxDisk:        sql.NullFloat64{Float64: maxDisk, Valid: metrics != nil},
		AvgDisk:        sql.NullFloat64{Float64: avgDisk, Valid: metrics != nil},
		DataPointCount: sql.NullInt64{Int64: int64(dataPointCount), Valid: metrics != nil},
	}

	_, err := s.svcCtx.ExternalSyncExecutionDetailModel.Insert(ctx, detail)
	if err != nil {
		s.logger.Errorf("保存执行详情失败 (host=%s): %v", hostIP, err)
	}
}

// GetRegisteredTaskCount 获取已注册的任务数量
func (s *EsSyncScheduler) GetRegisteredTaskCount() int {
	s.taskMutex.RLock()
	defer s.taskMutex.RUnlock()
	return len(s.tasks)
}

// IsTaskRegistered 检查任务是否已注册
func (s *EsSyncScheduler) IsTaskRegistered(taskId uint64) bool {
	s.taskMutex.RLock()
	defer s.taskMutex.RUnlock()
	_, exists := s.tasks[taskId]
	return exists
}
