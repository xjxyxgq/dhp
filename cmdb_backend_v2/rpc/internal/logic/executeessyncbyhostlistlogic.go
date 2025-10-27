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

type ExecuteEsSyncByHostListLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewExecuteEsSyncByHostListLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ExecuteEsSyncByHostListLogic {
	return &ExecuteEsSyncByHostListLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// ES数据同步执行相关方法
func (l *ExecuteEsSyncByHostListLogic) ExecuteEsSyncByHostList(in *cmpool.ExecuteESSyncByHostListReq) (*cmpool.ExecuteESSyncResp, error) {
	startTime := time.Now()

	l.Logger.Infof("开始执行ES同步: %d个主机", len(in.HostIpList))

	// 1. 验证参数
	if len(in.HostIpList) == 0 {
		return &cmpool.ExecuteESSyncResp{
			Success: false,
			Message: "主机列表不能为空",
		}, nil
	}

	if in.QueryTimeRange == "" {
		in.QueryTimeRange = "30d"
	}

	// 2. 获取ES配置
	esEndpoint := in.EsEndpoint
	if esEndpoint == "" {
		esEndpoint = l.svcCtx.Config.ESDataSource.DefaultEndpoint
	}

	esIndexPattern := l.svcCtx.Config.ESDataSource.DefaultIndexPattern

	// 3. 创建ES客户端
	esClient := elasticsearch.NewESClient(esEndpoint, l.svcCtx.Config.ESDataSource.TimeoutSeconds)

	// 4. 创建执行记录
	executionId, err := l.createExecutionLog(in.TaskName, len(in.HostIpList), in.QueryTimeRange)
	if err != nil {
		l.Logger.Errorf("创建执行记录失败: %v", err)
		return &cmpool.ExecuteESSyncResp{
			Success: false,
			Message: "创建执行记录失败",
		}, nil
	}

	// 5. 并发查询和同步数据
	successIps, failedIps, notInDatasourceIps := l.syncHostsData(esClient, esIndexPattern, in.HostIpList, in.QueryTimeRange, executionId)

	// 6. 统计数量
	successCount := len(successIps)
	failedCount := len(failedIps)
	notInDatasourceCount := len(notInDatasourceIps)

	// 7. 更新执行记录
	duration := time.Since(startTime).Milliseconds()
	executionStatus := "success"
	if failedCount > 0 {
		if successCount > 0 {
			executionStatus = "partial"
		} else {
			executionStatus = "failed"
		}
	}

	err = l.updateExecutionLog(executionId, executionStatus, successCount, failedCount, notInDatasourceCount, duration)
	if err != nil {
		l.Logger.Errorf("更新执行记录失败: %v", err)
	}

	l.Logger.Infof("ES同步完成: 总数=%d, 成功=%d, 失败=%d, 数据源中不存在=%d, 耗时=%dms",
		len(in.HostIpList), successCount, failedCount, notInDatasourceCount, duration)

	return &cmpool.ExecuteESSyncResp{
		Success:         true,
		Message:         fmt.Sprintf("同步完成: 成功%d个, 失败%d个, 数据源中不存在%d个", successCount, failedCount, notInDatasourceCount),
		ExecutionId:     executionId,
		TotalHosts:      int32(len(in.HostIpList)),
		SuccessCount:    int32(successCount),
		FailedCount:     int32(failedCount),
		NotInPoolCount:  int32(notInDatasourceCount),
		SuccessIpList:   successIps,
		FailedIpList:    failedIps,
		NotInPoolIpList: notInDatasourceIps,
	}, nil
}

// createExecutionLog 创建执行记录
func (l *ExecuteEsSyncByHostListLogic) createExecutionLog(taskName string, totalHosts int, queryTimeRange string) (int64, error) {
	// 使用 Model 方法创建执行记录
	executionLog := &model.ExternalSyncExecutionLog{
		TaskId:          0, // 手动执行没有关联任务
		TaskName:        taskName,
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
func (l *ExecuteEsSyncByHostListLogic) updateExecutionLog(executionId int64, status string, successCount, failedCount, notInPoolCount int, durationMs int64) error {
	// 使用 Model 方法更新执行记录
	return l.svcCtx.ExternalSyncExecutionLogModel.UpdateExecutionResult(
		l.ctx,
		uint64(executionId),
		status,
		successCount,
		failedCount,
		notInPoolCount,
		durationMs,
	)
}

// syncHostsData 并发同步主机数据，返回成功、失败、数据源中不存在的IP列表
func (l *ExecuteEsSyncByHostListLogic) syncHostsData(esClient *elasticsearch.ESClient, indexPattern string, hostIpList []string, timeRange string, executionId int64) (successIps, failedIps, notInDatasourceIps []string) {
	// 使用信号量控制并发数
	const maxConcurrency = 10
	sem := make(chan struct{}, maxConcurrency)

	var wg sync.WaitGroup
	var mu sync.Mutex

	// 初始化IP列表切片
	successIps = make([]string, 0)
	failedIps = make([]string, 0)
	notInDatasourceIps = make([]string, 0)

	for _, hostIP := range hostIpList {
		wg.Add(1)
		sem <- struct{}{} // 获取信号量

		go func(ip string) {
			defer wg.Done()
			defer func() { <-sem }() // 释放信号量

			// 查询ES数据
			metrics, err := esClient.QueryHostMetrics(l.ctx, indexPattern, ip, timeRange)
			if err != nil {
				// 区分"ES中无数据"和"查询失败"
				if elasticsearch.IsNoDataError(err) {
					l.Logger.Infof("主机 %s 在ES数据源中没有数据", ip)
					l.saveExecutionDetail(executionId, ip, "", "not_in_datasource", "ES数据源中无此主机数据", nil)
					mu.Lock()
					notInDatasourceIps = append(notInDatasourceIps, ip)
					mu.Unlock()
					return
				}

				l.Logger.Errorf("查询主机 %s 的ES数据失败: %v", ip, err)
				l.saveExecutionDetail(executionId, ip, "", "failed", err.Error(), nil)
				mu.Lock()
				failedIps = append(failedIps, ip)
				mu.Unlock()
				return
			}

			// 检查主机是否在 hosts_pool 中
			hostInfo, err := l.svcCtx.HostsPoolModel.FindByIP(l.ctx, ip)
			if err == sql.ErrNoRows || hostInfo == nil {
				// 主机不在pool中，自动插入
				l.Logger.Infof("主机 %s 不在hosts_pool中，准备插入", ip)
				poolId, err := l.svcCtx.HostsPoolModel.InsertIfNotExists(l.ctx, metrics.HostName, ip, "")
				if err != nil {
					l.Logger.Errorf("插入主机 %s 到hosts_pool失败: %v", ip, err)
					l.saveExecutionDetail(executionId, ip, metrics.HostName, "failed", fmt.Sprintf("插入hosts_pool失败: %v", err), metrics)
					mu.Lock()
					failedIps = append(failedIps, ip)
					mu.Unlock()
					return
				}
				hostInfo = &model.HostsPool{Id: uint64(poolId), HostIp: ip, HostName: metrics.HostName}
				l.Logger.Infof("成功插入新主机 %s 到hosts_pool, pool_id=%d", ip, poolId)
			} else if err != nil {
				l.Logger.Errorf("查询主机 %s 的hosts_pool信息失败: %v", ip, err)
				l.saveExecutionDetail(executionId, ip, metrics.HostName, "failed", err.Error(), metrics)
				mu.Lock()
				failedIps = append(failedIps, ip)
				mu.Unlock()
				return
			}

			// 同步数据到 server_resources
			err = l.syncToServerResources(uint64(hostInfo.Id), ip, metrics)
			if err != nil {
				l.Logger.Errorf("同步主机 %s 数据到server_resources失败: %v", ip, err)
				l.saveExecutionDetail(executionId, ip, hostInfo.HostName, "failed", err.Error(), metrics)
				mu.Lock()
				failedIps = append(failedIps, ip)
				mu.Unlock()
				return
			}

			// 保存执行详情
			l.saveExecutionDetail(executionId, ip, hostInfo.HostName, "success", "", metrics)
			mu.Lock()
			successIps = append(successIps, ip)
			mu.Unlock()

		}(hostIP)
	}

	wg.Wait()
	return
}

// syncToServerResources 同步数据到server_resources表
func (l *ExecuteEsSyncByHostListLogic) syncToServerResources(poolId uint64, hostIP string, metrics *elasticsearch.HostMetrics) error {
// 使用 Model 方法进行 upsert 操作
	// 字段映射:
	// - pool_id: poolId (从 hosts_pool 查询得到)
	// - ip: hostIP
	// - ES返回的数据说明：
	//   * CPU: 百分比值（%）
	//   * Memory: GB绝对值（已用内存）
	//   * Disk: GB绝对值（已用磁盘）
	return l.svcCtx.ServerResourcesModel.UpsertFromES(
		l.ctx,
		poolId,
		hostIP,
		metrics.MaxMemory, metrics.MaxDisk, metrics.MaxCPU, // 绝对值: 已用内存GB, 已用磁盘GB, CPU负载%
		metrics.MaxCPU, metrics.AvgCPU, 0,                   // CPU百分比: max, avg, min
		metrics.MaxMemory, metrics.AvgMemory, 0,             // 内存百分比: max, avg, min（注意：这里是复用绝对值，实际应该转换为%）
		metrics.MaxDisk, metrics.AvgDisk, 0,                 // 磁盘百分比: max, avg, min（注意：这里是复用绝对值，实际应该转换为%）
	)
}

// saveExecutionDetail 保存执行详情
func (l *ExecuteEsSyncByHostListLogic) saveExecutionDetail(executionId int64, hostIP, hostName, status, errorMsg string, metrics *elasticsearch.HostMetrics) {
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

	// 使用 Model 方法保存执行详情
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
