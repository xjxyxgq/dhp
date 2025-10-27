package logic

import (
	"context"

	"cmdb-api/cmpool"
	"cmdb-api/internal/svc"
	"cmdb-api/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type ExecuteExternalSyncFullSyncLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewExecuteExternalSyncFullSyncLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ExecuteExternalSyncFullSyncLogic {
	return &ExecuteExternalSyncFullSyncLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *ExecuteExternalSyncFullSyncLogic) ExecuteExternalSyncFullSync(req *types.ExecuteExternalSyncFullSyncRequest) (resp *types.ExecuteExternalSyncFullSyncResponse, err error) {
	// 参数验证
	if err := validateDataSource(req.DataSource); err != nil {
		l.Logger.Errorf("参数验证失败: %v", err)
		return &types.ExecuteExternalSyncFullSyncResponse{
			Success: false,
			Message: err.Error(),
		}, nil
	}

	normalizedSource := normalizeDataSource(req.DataSource)
	l.Logger.Infof("统一接口: 执行外部数据全量同步, DataSource=%s (标准化后=%s), GroupName=%s, TaskName=%s",
		req.DataSource, normalizedSource, req.GroupName, req.TaskName)

	// 根据数据源路由到不同的RPC方法
	if normalizedSource == "elasticsearch" {
		// 调用ES全量同步RPC
		l.Logger.Infof("数据源路由: %s -> ExecuteEsSyncFullSync RPC方法", normalizedSource)

		// 设置默认值
		groupName := req.GroupName
		if groupName == "" {
			groupName = "DB组"
		}
		queryTimeRange := req.QueryTimeRange
		if queryTimeRange == "" {
			queryTimeRange = "30d"
		}

		rpcReq := &cmpool.ExecuteESSyncFullSyncReq{
			GroupName:      groupName,
			QueryTimeRange: queryTimeRange,
			EsEndpoint:     req.EsEndpoint,
			TaskName:       req.TaskName,
		}

		rpcResp, err := l.svcCtx.CmpoolRpc.ExecuteEsSyncFullSync(l.ctx, rpcReq)
		if err != nil {
			l.Logger.Errorf("调用ES全量同步RPC失败: %v", err)
			return &types.ExecuteExternalSyncFullSyncResponse{
				Success: false,
				Message: "执行ES全量同步失败",
			}, nil
		}

		// ES全量同步响应 -> 统一响应
		l.Logger.Infof("统一接口执行成功: ExecutionId=%d, Total=%d, New=%d, Updated=%d, Failed=%d",
			rpcResp.ExecutionId, rpcResp.TotalHosts, rpcResp.NewHostsCount, rpcResp.UpdatedHostsCount, rpcResp.FailedCount)

		return &types.ExecuteExternalSyncFullSyncResponse{
			Success:               rpcResp.Success,
			Message:               rpcResp.Message,
			ExecutionId:           rpcResp.ExecutionId,
			TotalHosts:            rpcResp.TotalHosts,
			NewHostsCount:         rpcResp.NewHostsCount,         // ES全量同步特有
			UpdatedHostsCount:     rpcResp.UpdatedHostsCount,     // ES全量同步特有
			SuccessCount:          0,                              // 按主机列表同步特有，此处为0
			FailedCount:           rpcResp.FailedCount,
			NotInPoolCount:        0,                              // 按主机列表同步特有，此处为0
			NotInDatasourceCount:  0,                              // CMSys特有，此处为0
			NewHostIpList:         rpcResp.NewHostIpList,          // ES全量同步特有
			UpdatedHostIpList:     rpcResp.UpdatedHostIpList,      // ES全量同步特有
			FailedIpList:          rpcResp.FailedIpList,
			SuccessIpList:         nil,                            // 按主机列表同步特有
			NotInPoolIpList:       nil,                            // 按主机列表同步特有
			NotInDatasourceIpList: nil,                            // CMSys特有
		}, nil
	}

	// 调用CMSys全量同步RPC
	l.Logger.Infof("数据源路由: %s -> ExecuteCMSysSyncFullSync RPC方法", normalizedSource)

	rpcReq := &cmpool.ExecuteCMSysSyncFullSyncReq{
		TaskName: req.TaskName,
		Query:    req.Query,
	}

	rpcResp, err := l.svcCtx.CmpoolRpc.ExecuteCmsysSyncFullSync(l.ctx, rpcReq)
	if err != nil {
		l.Logger.Errorf("调用CMSys全量同步RPC失败: %v", err)
		return &types.ExecuteExternalSyncFullSyncResponse{
			Success: false,
			Message: "执行CMSys全量同步失败",
		}, nil
	}

	// CMSys全量同步响应 -> 统一响应（所有字段都存在）
	l.Logger.Infof("统一接口执行成功: ExecutionId=%d, Total=%d, New=%d, Updated=%d, Failed=%d",
		rpcResp.ExecutionId, rpcResp.TotalHosts, rpcResp.NewHostsCount, rpcResp.UpdatedHostsCount, rpcResp.FailedCount)

	return &types.ExecuteExternalSyncFullSyncResponse{
		Success:               rpcResp.Success,
		Message:               rpcResp.Message,
		DataSource:            "cmsys",
		ExecutionId:           rpcResp.ExecutionId,
		TotalHosts:            rpcResp.TotalHosts,
		NewHostsCount:         rpcResp.NewHostsCount,     // 全量同步统计新增
		UpdatedHostsCount:     rpcResp.UpdatedHostsCount, // 全量同步统计更新
		SuccessCount:          rpcResp.SuccessCount,
		FailedCount:           rpcResp.FailedCount,
		// CMSys字段
		NotInDatasourceCount:  rpcResp.NotInDatasourceCount,
		NotInDatasourceIpList: rpcResp.NotInDatasourceIpList,
		// ES字段（CMSys时固定为0/空）
		NotInPoolCount:        0,
		NotInPoolIpList:       []string{},
		// IP列表
		NewHostIpList:         rpcResp.NewHostIpList,
		UpdatedHostIpList:     rpcResp.UpdatedHostIpList,
		SuccessIpList:         rpcResp.SuccessIpList,
		FailedIpList:          rpcResp.FailedIpList,
	}, nil
}
