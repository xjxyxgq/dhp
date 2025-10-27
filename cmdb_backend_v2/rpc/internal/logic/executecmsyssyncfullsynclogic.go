package logic

import (
	"context"
	"database/sql"
	"fmt"
	"sync"
	"time"

	"cmdb-rpc/cmpool"
	"cmdb-rpc/internal/datasource/cmsys"
	"cmdb-rpc/internal/model"
	"cmdb-rpc/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
)

type ExecuteCmsysSyncFullSyncLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewExecuteCmsysSyncFullSyncLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ExecuteCmsysSyncFullSyncLogic {
	return &ExecuteCmsysSyncFullSyncLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// CMSys全量同步
func (l *ExecuteCmsysSyncFullSyncLogic) ExecuteCmsysSyncFullSync(in *cmpool.ExecuteCMSysSyncFullSyncReq) (*cmpool.ExecuteCMSysSyncResp, error) {
	startTime := time.Now()

	l.Logger.Infof("开始执行CMSys全量同步: query=%s", in.Query)

	// 1. 验证任务名称
	taskName := in.TaskName
	if taskName == "" {
		taskName = "CMSys全量同步"
	}

	// 2. 创建CMSys客户端
	cmsysClient := cmsys.NewCMSysClient(
		l.svcCtx.Config.CMSysDataSource.AuthEndpoint,
		l.svcCtx.Config.CMSysDataSource.DataEndpoint,
		l.svcCtx.Config.CMSysDataSource.AppCode,
		l.svcCtx.Config.CMSysDataSource.AppSecret,
		l.svcCtx.Config.CMSysDataSource.Operator,
		l.svcCtx.Config.CMSysDataSource.TimeoutSeconds,
	)

	// 3. 查询CMSys数据
	metrics, err := cmsysClient.QueryHostMetrics(l.ctx, in.Query)
	if err != nil {
		l.Logger.Errorf("查询CMSys数据失败: %v", err)
		return &cmpool.ExecuteCMSysSyncResp{
			Success: false,
			Message: fmt.Sprintf("查询CMSys数据失败: %v", err),
		}, nil
	}

	l.Logger.Infof("从CMSys查询到 %d 条主机数据", len(metrics))

	if len(metrics) == 0 {
		return &cmpool.ExecuteCMSysSyncResp{
			Success: false,
			Message: "CMSys中没有查询到符合条件的数据",
		}, nil
	}

	// 4. 创建执行记录
	executionLogic := NewExecuteCmsysSyncLogic(l.ctx, l.svcCtx)
	executionId, err := executionLogic.CreateExecutionLog(taskName, len(metrics))
	if err != nil {
		l.Logger.Errorf("创建执行记录失败: %v", err)
		return &cmpool.ExecuteCMSysSyncResp{
			Success: false,
			Message: "创建执行记录失败",
		}, nil
	}

	// 5. 并发同步数据（全量同步需要统计新增和更新）
	successIps, failedIps, notInDatasourceIps, newHostIps, updatedHostIps := l.syncHostsDataFull(metrics, executionId)

	// 6. 统计数量
	successCount := len(successIps)
	failedCount := len(failedIps)
	notInDatasourceCount := len(notInDatasourceIps)
	newHostsCount := len(newHostIps)
	updatedHostsCount := len(updatedHostIps)

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

	err = executionLogic.UpdateExecutionLog(executionId, executionStatus, successCount, failedCount, notInDatasourceCount, duration)
	if err != nil {
		l.Logger.Errorf("更新执行记录失败: %v", err)
	}

	l.Logger.Infof("CMSys全量同步完成: 总数=%d, 成功=%d, 失败=%d, 数据源中不存在=%d, 新增=%d, 更新=%d, 耗时=%dms",
		len(metrics), successCount, failedCount, notInDatasourceCount, newHostsCount, updatedHostsCount, duration)

	return &cmpool.ExecuteCMSysSyncResp{
		Success:      true,
		Message:      fmt.Sprintf("全量同步完成: 成功%d个, 失败%d个, 数据源中不存在%d个, 新增%d个, 更新%d个", successCount, failedCount, notInDatasourceCount, newHostsCount, updatedHostsCount),
		ExecutionId:  executionId,
		TotalHosts:   int32(len(metrics)),
		SuccessCount: int32(successCount),
		FailedCount:  int32(failedCount),
		// CMSys特有字段
		NotInDatasourceCount:  int32(notInDatasourceCount),
		NotInDatasourceIpList: notInDatasourceIps,
		// ES字段固定为0/空（统一响应结构要求），但全量同步有new/updated字段
		NotInPoolCount:    0,
		NotInPoolIpList:   []string{},
		NewHostsCount:     int32(newHostsCount),
		NewHostIpList:     newHostIps,
		UpdatedHostsCount: int32(updatedHostsCount),
		UpdatedHostIpList: updatedHostIps,
		// 通用字段
		SuccessIpList: successIps,
		FailedIpList:  failedIps,
	}, nil
}

// syncHostsDataFull 并发同步主机数据（全量同步版本，统计新增和更新）
func (l *ExecuteCmsysSyncFullSyncLogic) syncHostsDataFull(metricsList []*cmsys.HostMetrics, executionId int64) (successIps, failedIps, notInDatasourceIps, newHostIps, updatedHostIps []string) {
	// 使用信号量控制并发数
	const maxConcurrency = 10
	sem := make(chan struct{}, maxConcurrency)

	var wg sync.WaitGroup
	var mu sync.Mutex

	// 初始化IP列表切片
	successIps = make([]string, 0)
	failedIps = make([]string, 0)
	notInDatasourceIps = make([]string, 0)
	newHostIps = make([]string, 0)
	updatedHostIps = make([]string, 0)

	for _, metrics := range metricsList {
		wg.Add(1)
		sem <- struct{}{} // 获取信号量

		go func(m *cmsys.HostMetrics) {
			defer wg.Done()
			defer func() { <-sem }() // 释放信号量

			// 检查是否有有效数据
			if m.MaxCPU == 0 && m.MaxMemory == 0 && m.MaxDisk == 0 {
				l.Logger.Infof("主机 %s 在CMSys数据源中没有有效数据", m.IPAddress)
				l.saveExecutionDetail(executionId, m.IPAddress, "", "not_in_datasource", "CMSys数据源中无有效数据", m)
				mu.Lock()
				notInDatasourceIps = append(notInDatasourceIps, m.IPAddress)
				mu.Unlock()
				return
			}

			// 检查主机是否在 hosts_pool 中
			isNewHost := false
			hostInfo, err := l.svcCtx.HostsPoolModel.FindByIP(l.ctx, m.IPAddress)
			if err == sql.ErrNoRows || hostInfo == nil {
				// 主机不在pool中，标记为新增
				isNewHost = true
				l.Logger.Infof("主机 %s 不在hosts_pool中，准备插入（新增主机）", m.IPAddress)
				poolId, err := l.svcCtx.HostsPoolModel.InsertOrUpdateWithRemark(l.ctx, m.HostName, m.IPAddress, "", m.Remark)
				if err != nil {
					l.Logger.Errorf("插入主机 %s 到hosts_pool失败: %v", m.IPAddress, err)
					l.saveExecutionDetail(executionId, m.IPAddress, m.HostName, "failed", fmt.Sprintf("插入hosts_pool失败: %v", err), m)
					mu.Lock()
					failedIps = append(failedIps, m.IPAddress)
					mu.Unlock()
					return
				}
				hostInfo = &model.HostsPool{Id: uint64(poolId), HostIp: m.IPAddress, HostName: m.HostName}
				l.Logger.Infof("成功插入新主机 %s 到hosts_pool, pool_id=%d", m.IPAddress, poolId)
			} else if err != nil {
				l.Logger.Errorf("查询主机 %s 的hosts_pool信息失败: %v", m.IPAddress, err)
				l.saveExecutionDetail(executionId, m.IPAddress, "", "failed", err.Error(), m)
				mu.Lock()
				failedIps = append(failedIps, m.IPAddress)
				mu.Unlock()
				return
			} else {
				// 主机已存在，标记为更新
				l.Logger.Infof("主机 %s 已存在于hosts_pool中（更新主机）", m.IPAddress)
				// 更新 remark
				if m.Remark != "" {
					_, err = l.svcCtx.HostsPoolModel.InsertOrUpdateWithRemark(l.ctx, hostInfo.HostName, m.IPAddress, "", m.Remark)
					if err != nil {
						l.Logger.Errorf("更新主机 %s 的remark失败: %v", m.IPAddress, err)
					} else {
						l.Logger.Infof("成功更新主机 %s 的remark", m.IPAddress)
					}
				}
			}

			// 同步数据到 server_resources
			executionLogic := NewExecuteCmsysSyncLogic(l.ctx, l.svcCtx)
			err = executionLogic.syncToServerResources(uint64(hostInfo.Id), m.IPAddress, m)
			if err != nil {
				l.Logger.Errorf("同步主机 %s 数据到server_resources失败: %v", m.IPAddress, err)
				l.saveExecutionDetail(executionId, m.IPAddress, hostInfo.HostName, "failed", err.Error(), m)
				mu.Lock()
				failedIps = append(failedIps, m.IPAddress)
				mu.Unlock()
				return
			}

			// 保存执行详情
			l.saveExecutionDetail(executionId, m.IPAddress, hostInfo.HostName, "success", "", m)
			mu.Lock()
			successIps = append(successIps, m.IPAddress)
			if isNewHost {
				newHostIps = append(newHostIps, m.IPAddress)
			} else {
				updatedHostIps = append(updatedHostIps, m.IPAddress)
			}
			mu.Unlock()

		}(metrics)
	}

	wg.Wait()
	return
}

// saveExecutionDetail 保存执行详情 (复用 ES 同步的执行详情表)
func (l *ExecuteCmsysSyncFullSyncLogic) saveExecutionDetail(executionId int64, hostIP, hostName, status, errorMsg string, metrics *cmsys.HostMetrics) {
	var maxCpu, maxMem, maxDisk float64
	var dataPointCount int

	if metrics != nil {
		maxCpu = metrics.MaxCPU
		maxMem = metrics.MaxMemory
		maxDisk = metrics.MaxDisk
		dataPointCount = 1 // CMSys返回的是最大值，不是数据点数量
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
		AvgCpu:         sql.NullFloat64{Float64: maxCpu, Valid: metrics != nil}, // CMSys只有最大值，用最大值填充平均值
		MaxMemory:      sql.NullFloat64{Float64: maxMem, Valid: metrics != nil},
		AvgMemory:      sql.NullFloat64{Float64: maxMem, Valid: metrics != nil}, // CMSys只有最大值，用最大值填充平均值
		MaxDisk:        sql.NullFloat64{Float64: maxDisk, Valid: metrics != nil},
		AvgDisk:        sql.NullFloat64{Float64: maxDisk, Valid: metrics != nil}, // CMSys只有最大值，用最大值填充平均值
		DataPointCount: sql.NullInt64{Int64: int64(dataPointCount), Valid: metrics != nil},
	}

	_, err := l.svcCtx.ExternalSyncExecutionDetailModel.Insert(l.ctx, detail)
	if err != nil {
		l.Logger.Errorf("保存执行详情失败 (host=%s): %v", hostIP, err)
	}
}
