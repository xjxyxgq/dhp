package logic

import (
	"context"
	"fmt"

	"cmdb-rpc/cmpool"
	"cmdb-rpc/internal/datasource/cmsys"
	"cmdb-rpc/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
)

type ExecuteExternalSyncByHostListLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewExecuteExternalSyncByHostListLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ExecuteExternalSyncByHostListLogic {
	return &ExecuteExternalSyncByHostListLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// ExecuteExternalSyncByHostList 统一的按主机列表执行同步接口（支持 ES 和 CMSys）
func (l *ExecuteExternalSyncByHostListLogic) ExecuteExternalSyncByHostList(in *cmpool.ExecuteExternalSyncByHostListReq) (*cmpool.ExecuteExternalSyncResp, error) {
	// 1. 验证 data_source 参数
	if in.DataSource == "" {
		return &cmpool.ExecuteExternalSyncResp{
			Success: false,
			Message: "数据源类型不能为空 (elasticsearch/cmsys)",
		}, nil
	}

	// 2. 验证主机列表
	if len(in.HostIpList) == 0 {
		return &cmpool.ExecuteExternalSyncResp{
			Success: false,
			Message: "主机列表不能为空",
		}, nil
	}

	l.Logger.Infof("开始统一执行外部同步: 数据源=%s, 主机数=%d", in.DataSource, len(in.HostIpList))

	// 3. 根据数据源路由到对应实现
	switch in.DataSource {
	case "elasticsearch", "es":
		return l.executeFromES(in)
	case "cmsys":
		return l.executeFromCMSys(in)
	default:
		return &cmpool.ExecuteExternalSyncResp{
			Success: false,
			Message: fmt.Sprintf("不支持的数据源类型: %s (支持: elasticsearch/es, cmsys)", in.DataSource),
		}, nil
	}
}

// executeFromES 调用 ES 同步逻辑
func (l *ExecuteExternalSyncByHostListLogic) executeFromES(in *cmpool.ExecuteExternalSyncByHostListReq) (*cmpool.ExecuteExternalSyncResp, error) {
	// 转换为 ES 特定请求
	esReq := &cmpool.ExecuteESSyncByHostListReq{
		HostIpList:     in.HostIpList,
		QueryTimeRange: in.QueryTimeRange,
		EsEndpoint:     in.EsEndpoint,
		TaskName:       in.TaskName,
	}

	// 调用 ES Logic
	esLogic := NewExecuteEsSyncByHostListLogic(l.ctx, l.svcCtx)
	esResp, err := esLogic.ExecuteEsSyncByHostList(esReq)
	if err != nil {
		return nil, err
	}

	// 转换为统一响应格式
	return &cmpool.ExecuteExternalSyncResp{
		Success:       esResp.Success,
		Message:       esResp.Message,
		ExecutionId:   esResp.ExecutionId,
		TotalHosts:    esResp.TotalHosts,
		SuccessCount:  esResp.SuccessCount,
		FailedCount:   esResp.FailedCount,
		SuccessIpList: esResp.SuccessIpList,
		FailedIpList:  esResp.FailedIpList,
		// ES 特有字段
		NotInPoolCount:  esResp.NotInPoolCount,
		NotInPoolIpList: esResp.NotInPoolIpList,
		// CMSys 字段固定为0/空
		NotInDatasourceCount:  0,
		NotInDatasourceIpList: []string{},
		NewHostsCount:         0,
		NewHostIpList:         []string{},
		UpdatedHostsCount:     0,
		UpdatedHostIpList:     []string{},
	}, nil
}

// executeFromCMSys 调用 CMSys 同步逻辑（按指定IP列表查询）
func (l *ExecuteExternalSyncByHostListLogic) executeFromCMSys(in *cmpool.ExecuteExternalSyncByHostListReq) (*cmpool.ExecuteExternalSyncResp, error) {
	// CMSys 支持按 IP 列表批量查询，使用 QueryHostMetricsByIPs 方法
	// 实现策略：
	// 1. 批量查询指定的 IP 列表
	// 2. 调用同步逻辑处理数据

	l.Logger.Infof("CMSys按主机列表同步: 共 %d 个主机IP", len(in.HostIpList))

	// 创建 CMSys 客户端
	cmsysClient := cmsys.NewCMSysClient(
		l.svcCtx.Config.CMSysDataSource.AuthEndpoint,
		l.svcCtx.Config.CMSysDataSource.DataEndpoint,
		l.svcCtx.Config.CMSysDataSource.AppCode,
		l.svcCtx.Config.CMSysDataSource.AppSecret,
		l.svcCtx.Config.CMSysDataSource.Operator,
		l.svcCtx.Config.CMSysDataSource.TimeoutSeconds,
	)

	// 批量查询 CMSys 数据（使用 POST /platform/cmsys/data-by-ips 接口）
	allMetrics, err := cmsysClient.QueryHostMetricsByIPs(l.ctx, in.HostIpList)
	if err != nil {
		l.Logger.Errorf("批量查询CMSys数据失败: %v", err)
		return &cmpool.ExecuteExternalSyncResp{
			Success: false,
			Message: fmt.Sprintf("批量查询CMSys数据失败: %v", err),
		}, nil
	}

	l.Logger.Infof("CMSys批量查询结果: 找到 %d 个主机数据（请求了 %d 个IP）", len(allMetrics), len(in.HostIpList))

	// 计算哪些IP在数据源中不存在
	metricsIPMap := make(map[string]bool)
	for _, m := range allMetrics {
		metricsIPMap[m.IPAddress] = true
	}

	var notFoundIPs []string
	for _, ip := range in.HostIpList {
		if !metricsIPMap[ip] {
			notFoundIPs = append(notFoundIPs, ip)
		}
	}

	// 如果没有查询到任何数据
	if len(allMetrics) == 0 {
		return &cmpool.ExecuteExternalSyncResp{
			Success:               false,
			Message:               "CMSys中没有找到任何指定主机的数据",
			NotInDatasourceCount:  int32(len(notFoundIPs)),
			NotInDatasourceIpList: notFoundIPs,
		}, nil
	}

	// 调用 CMSys 同步逻辑处理这些数据
	// 这里复用 ExecuteCmsysSyncLogic 的数据处理逻辑，但使用我们查询到的指定主机数据
	cmsysLogic := NewExecuteCmsysSyncLogic(l.ctx, l.svcCtx)

	// 创建执行记录
	taskName := in.TaskName
	if taskName == "" {
		taskName = "CMSys按主机列表同步"
	}

	executionId, err := cmsysLogic.CreateExecutionLog(taskName, len(allMetrics))
	if err != nil {
		l.Logger.Errorf("创建执行记录失败: %v", err)
		return &cmpool.ExecuteExternalSyncResp{
			Success: false,
			Message: "创建执行记录失败",
		}, nil
	}

	// 同步主机数据
	successIps, failedIps, notInDatasourceIps := cmsysLogic.SyncHostsData(allMetrics, executionId)

	// 将查询时不存在的IP加入到 notInDatasourceIps
	notInDatasourceIps = append(notInDatasourceIps, notFoundIPs...)

	// 统计数量
	successCount := len(successIps)
	failedCount := len(failedIps)
	notInDatasourceCount := len(notInDatasourceIps)

	// 更新执行记录
	duration := int64(0) // 这里简化处理
	executionStatus := "success"
	if failedCount > 0 {
		if successCount > 0 {
			executionStatus = "partial"
		} else {
			executionStatus = "failed"
		}
	}

	err = cmsysLogic.UpdateExecutionLog(executionId, executionStatus, successCount, failedCount, notInDatasourceCount, duration)
	if err != nil {
		l.Logger.Errorf("更新执行记录失败: %v", err)
	}

	l.Logger.Infof("CMSys按主机列表同步完成: 成功=%d, 失败=%d, 数据源中不存在=%d",
		successCount, failedCount, notInDatasourceCount)

	// 转换为统一响应格式
	return &cmpool.ExecuteExternalSyncResp{
		Success:       true,
		Message:       fmt.Sprintf("同步完成: 成功%d个, 失败%d个, 数据源中不存在%d个", successCount, failedCount, notInDatasourceCount),
		ExecutionId:   executionId,
		TotalHosts:    int32(len(in.HostIpList)),
		SuccessCount:  int32(successCount),
		FailedCount:   int32(failedCount),
		SuccessIpList: successIps,
		FailedIpList:  failedIps,
		// CMSys 特有字段
		NotInDatasourceCount:  int32(notInDatasourceCount),
		NotInDatasourceIpList: notInDatasourceIps,
		// ES 字段固定为0/空
		NotInPoolCount:    0,
		NotInPoolIpList:   []string{},
		NewHostsCount:     0,
		NewHostIpList:     []string{},
		UpdatedHostsCount: 0,
		UpdatedHostIpList: []string{},
	}, nil
}
