package logic

import (
	"context"

	"cmdb-api/cmpool"
	"cmdb-api/internal/svc"
	"cmdb-api/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type ExecuteExternalSyncLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewExecuteExternalSyncLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ExecuteExternalSyncLogic {
	return &ExecuteExternalSyncLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *ExecuteExternalSyncLogic) ExecuteExternalSync(req *types.ExecuteExternalSyncRequest) (resp *types.ExecuteExternalSyncResponse, err error) {
	// 参数验证
	if err := validateDataSource(req.DataSource); err != nil {
		l.Logger.Errorf("参数验证失败: %v", err)
		return &types.ExecuteExternalSyncResponse{
			Success: false,
			Message: err.Error(),
		}, nil
	}

	normalizedSource := normalizeDataSource(req.DataSource)
	l.Logger.Infof("统一接口: 执行外部数据同步, DataSource=%s (标准化后=%s), TaskName=%s, HostCount=%d",
		req.DataSource, normalizedSource, req.TaskName, len(req.HostIpList))

	// 根据数据源路由到不同的RPC方法
	if normalizedSource == "elasticsearch" {
		// 调用ES同步RPC
		l.Logger.Infof("数据源路由: %s -> ExecuteEsSyncByHostList RPC方法", normalizedSource)

		rpcReq := &cmpool.ExecuteESSyncByHostListReq{
			HostIpList:     req.HostIpList,
			QueryTimeRange: req.QueryTimeRange,
			EsEndpoint:     req.EsEndpoint,
			TaskName:       req.TaskName,
		}

		rpcResp, err := l.svcCtx.CmpoolRpc.ExecuteEsSyncByHostList(l.ctx, rpcReq)
		if err != nil {
			l.Logger.Errorf("调用ES同步RPC失败: %v", err)
			return &types.ExecuteExternalSyncResponse{
				Success: false,
				Message: "执行ES同步失败",
			}, nil
		}

		// ES响应 -> 统一响应
		l.Logger.Infof("统一接口执行成功: ExecutionId=%d, Success=%d, Failed=%d, NotInPool=%d",
			rpcResp.ExecutionId, rpcResp.SuccessCount, rpcResp.FailedCount, rpcResp.NotInPoolCount)

		return &types.ExecuteExternalSyncResponse{
			Success:               rpcResp.Success,
			Message:               rpcResp.Message,
			ExecutionId:           rpcResp.ExecutionId,
			TotalHosts:            rpcResp.TotalHosts,
			SuccessCount:          rpcResp.SuccessCount,
			FailedCount:           rpcResp.FailedCount,
			NotInPoolCount:        rpcResp.NotInPoolCount, // ES特有
			NotInDatasourceCount:  0,                      // CMSys特有，此处为0
			SuccessIpList:         rpcResp.SuccessIpList,
			FailedIpList:          rpcResp.FailedIpList,
			NotInPoolIpList:       rpcResp.NotInPoolIpList, // ES特有
			NotInDatasourceIpList: nil,                     // CMSys特有
		}, nil
	}

	// 调用CMSys同步RPC（使用统一的按主机列表同步接口）
	l.Logger.Infof("数据源路由: %s -> ExecuteExternalSyncByHostList RPC方法 (统一接口)", normalizedSource)

	rpcReq := &cmpool.ExecuteExternalSyncByHostListReq{
		DataSource: normalizedSource,
		HostIpList: req.HostIpList,
		TaskName:   req.TaskName,
		CmsysQuery: req.Query, // CMSys 查询参数（如果需要）
	}

	rpcResp, err := l.svcCtx.CmpoolRpc.ExecuteExternalSyncByHostList(l.ctx, rpcReq)
	if err != nil {
		l.Logger.Errorf("调用CMSys同步RPC失败: %v", err)
		return &types.ExecuteExternalSyncResponse{
			Success: false,
			Message: "执行CMSys同步失败",
		}, nil
	}

	// 统一响应 -> API响应（已经是统一格式，直接使用）
	l.Logger.Infof("统一接口执行成功: ExecutionId=%d, Success=%d, Failed=%d, NotInDatasource=%d",
		rpcResp.ExecutionId, rpcResp.SuccessCount, rpcResp.FailedCount, rpcResp.NotInDatasourceCount)

	return &types.ExecuteExternalSyncResponse{
		Success:               rpcResp.Success,
		Message:               rpcResp.Message,
		ExecutionId:           rpcResp.ExecutionId,
		TotalHosts:            rpcResp.TotalHosts,
		SuccessCount:          rpcResp.SuccessCount,
		FailedCount:           rpcResp.FailedCount,
		NotInPoolCount:        rpcResp.NotInPoolCount,       // ES特有
		NotInDatasourceCount:  rpcResp.NotInDatasourceCount, // CMSys特有
		SuccessIpList:         rpcResp.SuccessIpList,
		FailedIpList:          rpcResp.FailedIpList,
		NotInPoolIpList:       rpcResp.NotInPoolIpList,       // ES特有
		NotInDatasourceIpList: rpcResp.NotInDatasourceIpList, // CMSys特有
	}, nil
}
