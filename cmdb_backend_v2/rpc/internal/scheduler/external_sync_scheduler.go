package scheduler

import (
	"context"
	"database/sql"
	"fmt"
	"sync"
	"time"

	"cmdb-rpc/internal/datasource/cmsys"
	"cmdb-rpc/internal/datasource/elasticsearch"
	"cmdb-rpc/internal/model"
	"cmdb-rpc/internal/svc"

	"github.com/robfig/cron/v3"
	"github.com/zeromicro/go-zero/core/logx"
)

// ExternalSyncScheduler 统一外部资源同步任务调度器
// 支持多种数据源：ES、CMSys
type ExternalSyncScheduler struct {
	cron      *cron.Cron
	svcCtx    *svc.ServiceContext
	tasks     map[uint64]cron.EntryID // taskId -> cronEntryId
	taskMutex sync.RWMutex
	logger    logx.Logger
}

// NewExternalSyncScheduler 创建新的统一外部资源同步任务调度器
func NewExternalSyncScheduler(svcCtx *svc.ServiceContext) *ExternalSyncScheduler {
	return &ExternalSyncScheduler{
		cron:   cron.New(cron.WithSeconds()), // 支持秒级精度
		svcCtx: svcCtx,
		tasks:  make(map[uint64]cron.EntryID),
		logger: logx.WithContext(context.Background()),
	}
}

// Start 启动调度器并加载所有已启用的任务
func (s *ExternalSyncScheduler) Start() error {
	s.logger.Info("启动统一外部资源同步调度器...")

	// 从数据库加载所有已启用的任务
	ctx := context.Background()
	tasks, err := s.svcCtx.ExternalSyncTaskConfigModel.FindAll(ctx, true)
	if err != nil {
		s.logger.Errorf("加载外部资源同步任务失败: %v", err)
		return err
	}

	// 注册所有已启用的任务
	for _, task := range tasks {
		if err := s.RegisterTask(task); err != nil {
			s.logger.Errorf("注册任务失败: TaskId=%d, TaskName=%s, DataSource=%s, Error=%v",
				task.Id, task.TaskName, task.DataSource, err)
			continue
		}
		s.logger.Infof("成功注册任务: TaskId=%d, TaskName=%s, DataSource=%s, Cron=%s",
			task.Id, task.TaskName, task.DataSource, task.CronExpression)
	}

	// 启动cron调度器
	s.cron.Start()
	s.logger.Infof("统一外部资源同步调度器已启动，共注册 %d 个任务", len(s.tasks))
	return nil
}

// Stop 停止调度器
func (s *ExternalSyncScheduler) Stop() {
	s.logger.Info("停止统一外部资源同步调度器...")
	ctx := s.cron.Stop()
	<-ctx.Done()
	s.logger.Info("统一外部资源同步调度器已停止")
}

// RegisterTask 注册任务到调度器
func (s *ExternalSyncScheduler) RegisterTask(task *model.ExternalSyncTaskConfig) error {
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
	s.logger.Infof("任务已注册到调度器: TaskId=%d, TaskName=%s, DataSource=%s, EntryId=%d",
		task.Id, task.TaskName, task.DataSource, entryID)

	return nil
}

// UnregisterTask 从调度器注销任务
func (s *ExternalSyncScheduler) UnregisterTask(taskId uint64) {
	s.taskMutex.Lock()
	defer s.taskMutex.Unlock()
	s.unregisterTaskInternal(taskId)
}

// unregisterTaskInternal 内部注销任务（不加锁）
func (s *ExternalSyncScheduler) unregisterTaskInternal(taskId uint64) {
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
func (s *ExternalSyncScheduler) createJob(task *model.ExternalSyncTaskConfig) func() {
	return func() {
		ctx := context.Background()
		startTime := time.Now()
		s.logger.Infof("开始执行定时同步任务: TaskId=%d, TaskName=%s, DataSource=%s",
			task.Id, task.TaskName, task.DataSource)

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

		s.logger.Infof("定时任务触发: TaskId=%d, TaskName=%s, DataSource=%s, Cron=%s, HostCount=%d",
			latestTask.Id, latestTask.TaskName, latestTask.DataSource, latestTask.CronExpression, len(hostIpList))

		// 根据数据源类型路由到相应的同步方法
		var executionId int64
		var success bool
		var message string

		switch latestTask.DataSource {
		case "elasticsearch":
			executionId, success, message = s.executeSyncFromES(ctx, latestTask, hostIpList, startTime)
		case "cmsys":
			executionId, success, message = s.executeSyncFromCMSys(ctx, latestTask, hostIpList, startTime)
		default:
			s.logger.Errorf("不支持的数据源类型: TaskId=%d, DataSource=%s", task.Id, latestTask.DataSource)
			return
		}

		if success {
			s.logger.Infof("定时任务执行成功: TaskId=%d, DataSource=%s, ExecutionId=%d, Message=%s",
				task.Id, latestTask.DataSource, executionId, message)
		} else {
			s.logger.Errorf("定时任务执行失败: TaskId=%d, DataSource=%s, Message=%s",
				task.Id, latestTask.DataSource, message)
		}
	}
}

// executeSyncFromES 执行从ES数据源同步
func (s *ExternalSyncScheduler) executeSyncFromES(ctx context.Context, task *model.ExternalSyncTaskConfig, hostIpList []string, startTime time.Time) (int64, bool, string) {
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
	successCount, failedCount, notInDatasourceCount := s.syncHostsFromES(ctx, esClient, esIndexPattern, hostIpList, task.QueryTimeRange, executionId)

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

// executeSyncFromCMSys 执行从CMSys数据源同步
func (s *ExternalSyncScheduler) executeSyncFromCMSys(ctx context.Context, task *model.ExternalSyncTaskConfig, hostIpList []string, startTime time.Time) (int64, bool, string) {
	s.logger.Infof("开始执行CMSys同步: TaskId=%d, TaskName=%s, 主机数=%d", task.Id, task.TaskName, len(hostIpList))

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

	// 2. 创建CMSys客户端
	cmsysClient := cmsys.NewCMSysClient(
		s.svcCtx.Config.CMSysDataSource.AuthEndpoint,
		s.svcCtx.Config.CMSysDataSource.DataEndpoint,
		s.svcCtx.Config.CMSysDataSource.AppCode,
		s.svcCtx.Config.CMSysDataSource.AppSecret,
		s.svcCtx.Config.CMSysDataSource.Operator,
		s.svcCtx.Config.CMSysDataSource.TimeoutSeconds,
	)

	// 3. 获取查询参数
	query := ""
	if task.CmsysQuery.Valid {
		query = task.CmsysQuery.String
	}

	// 4. 并发同步数据
	successCount, failedCount, notInDatasourceCount := s.syncHostsFromCMSys(ctx, cmsysClient, query, hostIpList, executionId)

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
	s.logger.Infof("CMSys同步执行完成: TaskId=%d, ExecutionId=%d, %s", task.Id, executionId, message)

	return executionId, true, message
}

// syncHostsFromES 从ES同步主机数据
func (s *ExternalSyncScheduler) syncHostsFromES(ctx context.Context, esClient *elasticsearch.ESClient, indexPattern string, hostIpList []string, timeRange string, executionId int64) (successCount, failedCount, notInDatasourceCount int) {
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
					s.saveExecutionDetail(ctx, executionId, ip, "", "not_in_datasource", "ES数据源中无此主机数据", nil, nil)
					mu.Lock()
					failedCount++
					mu.Unlock()
					return
				}
				s.logger.Errorf("查询主机 %s 的ES数据失败: %v", ip, err)
				s.saveExecutionDetail(ctx, executionId, ip, "", "failed", err.Error(), nil, nil)
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
					s.saveExecutionDetail(ctx, executionId, ip, metrics.HostName, "failed", fmt.Sprintf("插入hosts_pool失败: %v", err), metrics, nil)
					mu.Lock()
					failedCount++
					mu.Unlock()
					return
				}
				hostInfo = &model.HostsPool{Id: uint64(poolId), HostIp: ip, HostName: metrics.HostName}
				s.logger.Infof("成功插入新主机 %s 到hosts_pool, pool_id=%d", ip, poolId)
			} else if err != nil {
				s.logger.Errorf("查询主机 %s 的hosts_pool信息失败: %v", ip, err)
				s.saveExecutionDetail(ctx, executionId, ip, metrics.HostName, "failed", err.Error(), metrics, nil)
				mu.Lock()
				failedCount++
				mu.Unlock()
				return
			}

			// 同步数据到server_resources
			err = s.svcCtx.ServerResourcesModel.UpsertFromES(ctx, uint64(hostInfo.Id), ip,
				metrics.MaxMemory, metrics.MaxDisk, metrics.MaxCPU,
				metrics.MaxCPU, metrics.AvgCPU, 0,
				metrics.MaxMemory, metrics.AvgMemory, 0,
				metrics.MaxDisk, metrics.AvgDisk, 0)
			if err != nil {
				s.logger.Errorf("同步主机 %s 数据到server_resources失败: %v", ip, err)
				s.saveExecutionDetail(ctx, executionId, ip, hostInfo.HostName, "failed", err.Error(), metrics, nil)
				mu.Lock()
				failedCount++
				mu.Unlock()
				return
			}

			// 保存执行详情
			s.saveExecutionDetail(ctx, executionId, ip, hostInfo.HostName, "success", "", metrics, nil)
			mu.Lock()
			successCount++
			mu.Unlock()
		}(hostIP)
	}

	wg.Wait()
	return
}

// syncHostsFromCMSys 从CMSys同步主机数据
func (s *ExternalSyncScheduler) syncHostsFromCMSys(ctx context.Context, cmsysClient *cmsys.CMSysClient, query string, hostIpList []string, executionId int64) (successCount, failedCount, notInDatasourceCount int) {
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

			// 查询CMSys数据 - 按IP查询单个主机
			metrics, err := cmsysClient.QueryHostMetricsByIP(ctx, ip)
			if err != nil {
				if cmsys.IsNoDataError(err) {
					s.logger.Infof("主机 %s 在CMSys数据源中没有数据", ip)
					s.saveExecutionDetail(ctx, executionId, ip, "", "not_in_datasource", "CMSys数据源中无此主机数据", nil, nil)
					mu.Lock()
					notInDatasourceCount++
					mu.Unlock()
					return
				}
				s.logger.Errorf("查询主机 %s 的CMSys数据失败: %v", ip, err)
				s.saveExecutionDetail(ctx, executionId, ip, "", "failed", err.Error(), nil, nil)
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
					s.saveExecutionDetail(ctx, executionId, ip, metrics.HostName, "failed", fmt.Sprintf("插入hosts_pool失败: %v", err), nil, metrics)
					mu.Lock()
					failedCount++
					mu.Unlock()
					return
				}
				hostInfo = &model.HostsPool{Id: uint64(poolId), HostIp: ip, HostName: metrics.HostName}
				s.logger.Infof("成功插入新主机 %s 到hosts_pool, pool_id=%d", ip, poolId)
			} else if err != nil {
				s.logger.Errorf("查询主机 %s 的hosts_pool信息失败: %v", ip, err)
				s.saveExecutionDetail(ctx, executionId, ip, metrics.HostName, "failed", err.Error(), nil, metrics)
				mu.Lock()
				failedCount++
				mu.Unlock()
				return
			}

			// 同步数据到server_resources
			err = s.svcCtx.ServerResourcesModel.UpsertFromCMSys(ctx, uint64(hostInfo.Id), ip,
				metrics.CPUUsedPercent, metrics.MemoryUsedPercent, metrics.DiskUsedPercent)
			if err != nil {
				s.logger.Errorf("同步主机 %s 数据到server_resources失败: %v", ip, err)
				s.saveExecutionDetail(ctx, executionId, ip, hostInfo.HostName, "failed", err.Error(), nil, metrics)
				mu.Lock()
				failedCount++
				mu.Unlock()
				return
			}

			// 保存执行详情
			s.saveExecutionDetail(ctx, executionId, ip, hostInfo.HostName, "success", "", nil, metrics)
			mu.Lock()
			successCount++
			mu.Unlock()
		}(hostIP)
	}

	wg.Wait()
	return
}

// saveExecutionDetail 保存执行详情
func (s *ExternalSyncScheduler) saveExecutionDetail(ctx context.Context, executionId int64, hostIP, hostName, status, errorMsg string, esMetrics *elasticsearch.HostMetrics, cmsysMetrics *cmsys.HostMetrics) {
	var maxCpu, avgCpu, maxMem, avgMem, maxDisk, avgDisk float64
	var dataPointCount int

	if esMetrics != nil {
		maxCpu = esMetrics.MaxCPU
		avgCpu = esMetrics.AvgCPU
		maxMem = esMetrics.MaxMemory
		avgMem = esMetrics.AvgMemory
		maxDisk = esMetrics.MaxDisk
		avgDisk = esMetrics.AvgDisk
		dataPointCount = esMetrics.DataPointCount
		if hostName == "" {
			hostName = esMetrics.HostName
		}
	} else if cmsysMetrics != nil {
		maxCpu = cmsysMetrics.CPUUsedPercent
		avgCpu = cmsysMetrics.CPUUsedPercent
		maxMem = cmsysMetrics.MemoryUsedPercent
		avgMem = cmsysMetrics.MemoryUsedPercent
		maxDisk = cmsysMetrics.DiskUsedPercent
		avgDisk = cmsysMetrics.DiskUsedPercent
		dataPointCount = 1
		if hostName == "" {
			hostName = cmsysMetrics.HostName
		}
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
		MaxCpu:         sql.NullFloat64{Float64: maxCpu, Valid: esMetrics != nil || cmsysMetrics != nil},
		AvgCpu:         sql.NullFloat64{Float64: avgCpu, Valid: esMetrics != nil || cmsysMetrics != nil},
		MaxMemory:      sql.NullFloat64{Float64: maxMem, Valid: esMetrics != nil || cmsysMetrics != nil},
		AvgMemory:      sql.NullFloat64{Float64: avgMem, Valid: esMetrics != nil || cmsysMetrics != nil},
		MaxDisk:        sql.NullFloat64{Float64: maxDisk, Valid: esMetrics != nil || cmsysMetrics != nil},
		AvgDisk:        sql.NullFloat64{Float64: avgDisk, Valid: esMetrics != nil || cmsysMetrics != nil},
		DataPointCount: sql.NullInt64{Int64: int64(dataPointCount), Valid: esMetrics != nil || cmsysMetrics != nil},
	}

	_, err := s.svcCtx.ExternalSyncExecutionDetailModel.Insert(ctx, detail)
	if err != nil {
		s.logger.Errorf("保存执行详情失败 (host=%s): %v", hostIP, err)
	}
}

// GetRegisteredTaskCount 获取已注册的任务数量
func (s *ExternalSyncScheduler) GetRegisteredTaskCount() int {
	s.taskMutex.RLock()
	defer s.taskMutex.RUnlock()
	return len(s.tasks)
}

// IsTaskRegistered 检查任务是否已注册
func (s *ExternalSyncScheduler) IsTaskRegistered(taskId uint64) bool {
	s.taskMutex.RLock()
	defer s.taskMutex.RUnlock()
	_, exists := s.tasks[taskId]
	return exists
}
