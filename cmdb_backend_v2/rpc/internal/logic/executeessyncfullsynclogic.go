package logic

import (
	"context"
	"database/sql"
	"fmt"
	"sync"
	"time"

	"cmdb-rpc/cmpool"
	"cmdb-rpc/internal/datasource/elasticsearch"
	"cmdb-rpc/internal/model"
	"cmdb-rpc/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
)

type ExecuteEsSyncFullSyncLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewExecuteEsSyncFullSyncLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ExecuteEsSyncFullSyncLogic {
	return &ExecuteEsSyncFullSyncLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// ExecuteEsSyncFullSync ES全量同步：从ES检索group="DB组"的所有主机数据并同步到数据库
func (l *ExecuteEsSyncFullSyncLogic) ExecuteEsSyncFullSync(in *cmpool.ExecuteESSyncFullSyncReq) (*cmpool.ExecuteESSyncFullSyncResp, error) {
	startTime := time.Now()

	// 1. 设置默认值
	groupName := in.GroupName
	if groupName == "" {
		groupName = "DB组"
	}

	queryTimeRange := in.QueryTimeRange
	if queryTimeRange == "" {
		queryTimeRange = "30d"
	}

	taskName := in.TaskName
	if taskName == "" {
		taskName = "ES全量同步"
	}

	l.Logger.Infof("开始执行ES全量同步: Group=%s, TimeRange=%s", groupName, queryTimeRange)

	// 2. 获取ES配置
	esEndpoint := in.EsEndpoint
	if esEndpoint == "" {
		esEndpoint = l.svcCtx.Config.ESDataSource.DefaultEndpoint
	}
	esIndexPattern := l.svcCtx.Config.ESDataSource.DefaultIndexPattern

	// 3. 创建ES客户端
	esClient := elasticsearch.NewESClient(esEndpoint, l.svcCtx.Config.ESDataSource.TimeoutSeconds)

	// 4. 从ES查询所有group="DB组"的主机数据
	groupHosts, err := esClient.QueryGroupHosts(l.ctx, esIndexPattern, groupName, queryTimeRange)
	if err != nil {
		l.Logger.Errorf("查询ES group=%s 数据失败: %v", groupName, err)
		return &cmpool.ExecuteESSyncFullSyncResp{
			Success: false,
			Message: fmt.Sprintf("查询ES数据失败: %v", err),
		}, nil
	}

	totalHosts := len(groupHosts)
	l.Logger.Infof("从ES中查询到 %d 台主机(group=%s)", totalHosts, groupName)

	// 调试日志：打印ES返回的数据样本
	if totalHosts > 0 {
		l.Logger.Infof("ES数据样本 - 第1台主机: IP=%s, Name=%s, MaxCPU=%.2f, MaxMem=%.2f, MaxDisk=%.2f",
			groupHosts[0].HostIP, groupHosts[0].HostName, groupHosts[0].MaxCPU, groupHosts[0].MaxMemory, groupHosts[0].MaxDisk)
	}

	if totalHosts == 0 {
		return &cmpool.ExecuteESSyncFullSyncResp{
			Success: true,
			Message: fmt.Sprintf("ES中没有group=%s的主机数据", groupName),
		}, nil
	}

	// 5. 创建执行记录
	executionId, err := l.createExecutionLog(taskName, totalHosts, queryTimeRange, groupName)
	if err != nil {
		l.Logger.Errorf("创建执行记录失败: %v", err)
		return &cmpool.ExecuteESSyncFullSyncResp{
			Success: false,
			Message: "创建执行记录失败",
		}, nil
	}

	// 6. 并发同步数据到数据库
	newHostIps, updatedHostIps, failedIps := l.syncGroupHostsData(groupHosts, executionId)

	// 7. 统计结果
	newHostsCount := len(newHostIps)
	updatedHostsCount := len(updatedHostIps)
	failedCount := len(failedIps)

	// 8. 更新执行记录
	duration := time.Since(startTime).Milliseconds()
	executionStatus := "success"
	if failedCount > 0 {
		if updatedHostsCount > 0 || newHostsCount > 0 {
			executionStatus = "partial"
		} else {
			executionStatus = "failed"
		}
	}

	err = l.updateExecutionLog(executionId, executionStatus, newHostsCount, updatedHostsCount, failedCount, duration)
	if err != nil {
		l.Logger.Errorf("更新执行记录失败: %v", err)
	}

	l.Logger.Infof("ES全量同步完成: 总数=%d, 新增=%d, 更新=%d, 失败=%d, 耗时=%dms",
		totalHosts, newHostsCount, updatedHostsCount, failedCount, duration)

	// 根据实际执行状态构建响应消息
	var message string
	var success bool
	switch executionStatus {
	case "success":
		success = true
		message = fmt.Sprintf("同步成功完成: 新增%d个, 更新%d个", newHostsCount, updatedHostsCount)
	case "partial":
		success = true // 部分成功也返回true，但消息中说明有失败
		message = fmt.Sprintf("同步部分完成: 新增%d个, 更新%d个, 失败%d个", newHostsCount, updatedHostsCount, failedCount)
	case "failed":
		success = false
		if failedCount == totalHosts {
			message = fmt.Sprintf("同步全部失败: 共%d个主机全部失败", failedCount)
		} else {
			message = fmt.Sprintf("同步失败: 新增%d个, 更新%d个, 失败%d个", newHostsCount, updatedHostsCount, failedCount)
		}
	default:
		success = false
		message = fmt.Sprintf("同步状态未知: 新增%d个, 更新%d个, 失败%d个", newHostsCount, updatedHostsCount, failedCount)
	}

	return &cmpool.ExecuteESSyncFullSyncResp{
		Success:           success,
		Message:           message,
		ExecutionId:       executionId,
		TotalHosts:        int32(totalHosts),
		NewHostsCount:     int32(newHostsCount),
		UpdatedHostsCount: int32(updatedHostsCount),
		FailedCount:       int32(failedCount),
		NewHostIpList:     newHostIps,
		UpdatedHostIpList: updatedHostIps,
		FailedIpList:      failedIps,
	}, nil
}

// createExecutionLog 创建执行记录
func (l *ExecuteEsSyncFullSyncLogic) createExecutionLog(taskName string, totalHosts int, queryTimeRange, groupName string) (int64, error) {
	executionLog := &model.ExternalSyncExecutionLog{
		TaskId:          0, // 全量同步没有关联任务
		TaskName:        fmt.Sprintf("%s(group=%s)", taskName, groupName),
		DataSource:      "elasticsearch", // 设置数据源为 elasticsearch
		ExecutionTime:   time.Now(),
		ExecutionStatus: "running",
		TotalHosts:      int64(totalHosts),
		QueryTimeRange:  sql.NullString{String: queryTimeRange, Valid: queryTimeRange != ""},
	}

	result, err := l.svcCtx.ExternalSyncExecutionLogModel.Insert(l.ctx, executionLog)
	if err != nil {
		return 0, err
	}

	return result.LastInsertId()
}

// updateExecutionLog 更新执行记录
func (l *ExecuteEsSyncFullSyncLogic) updateExecutionLog(executionId int64, status string, newCount, updatedCount, failedCount int, durationMs int64) error {
	// 注意：这里使用 SuccessCount 存储更新数，NotInPoolCount 存储新增数
	return l.svcCtx.ExternalSyncExecutionLogModel.UpdateExecutionResult(
		l.ctx,
		uint64(executionId),
		status,
		updatedCount, // SuccessCount: 更新到server_resources的主机数
		failedCount,  // FailedCount: 失败数量
		newCount,     // NotInPoolCount: 新增到hosts_pool的主机数
		durationMs,
	)
}

// syncGroupHostsData 并发同步group主机数据
func (l *ExecuteEsSyncFullSyncLogic) syncGroupHostsData(groupHosts []*elasticsearch.GroupHostMetrics, executionId int64) (newHostIps, updatedHostIps, failedIps []string) {
	// 使用信号量控制并发数
	const maxConcurrency = 10
	sem := make(chan struct{}, maxConcurrency)

	var wg sync.WaitGroup
	var mu sync.Mutex

	// 初始化IP列表切片
	newHostIps = make([]string, 0)
	updatedHostIps = make([]string, 0)
	failedIps = make([]string, 0)

	for _, hostMetrics := range groupHosts {
		wg.Add(1)
		sem <- struct{}{} // 获取信号量

		go func(metrics *elasticsearch.GroupHostMetrics) {
			defer wg.Done()
			defer func() { <-sem }() // 释放信号量

			hostIP := metrics.HostIP
			hostName := metrics.HostName

			// 检查主机是否在 hosts_pool 中
			hostInfo, err := l.svcCtx.HostsPoolModel.FindByIP(l.ctx, hostIP)
			isNewHost := false

			if err == sql.ErrNoRows || hostInfo == nil {
				// 主机不在pool中，插入新主机
				l.Logger.Infof("主机 %s(%s) 不在hosts_pool中，准备插入", hostIP, hostName)

				// 插入到 hosts_pool
				poolId, err := l.svcCtx.HostsPoolModel.InsertIfNotExists(l.ctx, hostName, hostIP, "")
				if err != nil {
					l.Logger.Errorf("插入主机 %s 到hosts_pool失败: %v", hostIP, err)
					l.saveExecutionDetail(executionId, hostIP, hostName, "failed", fmt.Sprintf("插入hosts_pool失败: %v", err), metrics)
					mu.Lock()
					failedIps = append(failedIps, hostIP)
					mu.Unlock()
					return
				}

				isNewHost = true
				hostInfo = &model.HostsPool{Id: uint64(poolId), HostIp: hostIP, HostName: hostName}
				l.Logger.Infof("成功插入新主机 %s 到hosts_pool, pool_id=%d", hostIP, poolId)
			} else if err != nil {
				l.Logger.Errorf("查询主机 %s 的hosts_pool信息失败: %v", hostIP, err)
				l.saveExecutionDetail(executionId, hostIP, hostName, "failed", err.Error(), metrics)
				mu.Lock()
				failedIps = append(failedIps, hostIP)
				mu.Unlock()
				return
			}

			// 同步数据到 server_resources
			err = l.syncToServerResources(uint64(hostInfo.Id), hostIP, metrics)
			if err != nil {
				l.Logger.Errorf("同步主机 %s 数据到server_resources失败: %v", hostIP, err)
				l.saveExecutionDetail(executionId, hostIP, hostName, "failed", err.Error(), metrics)
				mu.Lock()
				failedIps = append(failedIps, hostIP)
				mu.Unlock()
				return
			}

			// 保存执行详情
			l.saveExecutionDetail(executionId, hostIP, hostName, "success", "", metrics)

			mu.Lock()
			if isNewHost {
				newHostIps = append(newHostIps, hostIP)
			} else {
				updatedHostIps = append(updatedHostIps, hostIP)
			}
			mu.Unlock()

		}(hostMetrics)
	}

	wg.Wait()
	return
}

// syncToServerResources 同步数据到server_resources表
func (l *ExecuteEsSyncFullSyncLogic) syncToServerResources(poolId uint64, hostIP string, metrics *elasticsearch.GroupHostMetrics) error {
	// 调试日志：打印写入前的数据
	l.Logger.Infof("写入 server_resources: PoolID=%d, IP=%s, UsedMem=%.2fGB, UsedDisk=%.2fGB, CPULoad=%.2f%%, CPU(%.2f/%.2f), Mem(%.2f/%.2f), Disk(%.2f/%.2f)",
		poolId, hostIP, metrics.MaxMemory, metrics.MaxDisk, metrics.MaxCPU,
		metrics.MaxCPU, metrics.AvgCPU, metrics.MaxMemory, metrics.AvgMemory, metrics.MaxDisk, metrics.AvgDisk)

	// ES返回的数据说明：
	// - CPU: 百分比值（%）
	// - Memory: GB绝对值（已用内存）
	// - Disk: GB绝对值（已用磁盘）
	err := l.svcCtx.ServerResourcesModel.UpsertFromES(
		l.ctx,
		poolId,
		hostIP,
		metrics.MaxMemory, metrics.MaxDisk, metrics.MaxCPU, // 绝对值: 已用内存GB, 已用磁盘GB, CPU负载%
		metrics.MaxCPU, metrics.AvgCPU, 0,                   // CPU百分比: max, avg, min
		metrics.MaxMemory, metrics.AvgMemory, 0,             // 内存百分比: max, avg, min（注意：这里是复用绝对值，实际应该转换为%）
		metrics.MaxDisk, metrics.AvgDisk, 0,                 // 磁盘百分比: max, avg, min（注意：这里是复用绝对值，实际应该转换为%）
	)

	if err != nil {
		l.Logger.Errorf("UpsertFromES 失败: PoolID=%d, IP=%s, Error=%v", poolId, hostIP, err)
	} else {
		l.Logger.Infof("UpsertFromES 成功: PoolID=%d, IP=%s", poolId, hostIP)
	}

	return err
}

// saveExecutionDetail 保存执行详情
func (l *ExecuteEsSyncFullSyncLogic) saveExecutionDetail(executionId int64, hostIP, hostName, status, errorMsg string, metrics *elasticsearch.GroupHostMetrics) {
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

	_, err := l.svcCtx.ExternalSyncExecutionDetailModel.Insert(l.ctx, detail)
	if err != nil {
		l.Logger.Errorf("保存执行详情失败 (host=%s): %v", hostIP, err)
	}
}
