package logic

import (
	"context"

	"cmdb-api/cmpool"
	"cmdb-api/internal/svc"
	"cmdb-api/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type ExecuteExternalSyncByFileLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewExecuteExternalSyncByFileLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ExecuteExternalSyncByFileLogic {
	return &ExecuteExternalSyncByFileLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *ExecuteExternalSyncByFileLogic) ExecuteExternalSyncByFile(req *types.ExecuteExternalSyncByFileRequest) (resp *types.ExecuteExternalSyncResponse, err error) {
	// 参数验证
	if err := validateDataSource(req.DataSource); err != nil {
		l.Logger.Errorf("参数验证失败: %v", err)
		return &types.ExecuteExternalSyncResponse{
			Success: false,
			Message: err.Error(),
		}, nil
	}

	normalizedSource := normalizeDataSource(req.DataSource)
	l.Logger.Infof("统一接口: 执行外部数据同步(文件), DataSource=%s (标准化后=%s), TaskName=%s",
		req.DataSource, normalizedSource, req.TaskName)

	// 根据数据源路由到不同的RPC方法
	if normalizedSource == "elasticsearch" {
		// 调用ES文件同步RPC
		l.Logger.Infof("数据源路由: %s -> ExecuteEsSyncByFile RPC方法", normalizedSource)

		rpcReq := &cmpool.ExecuteESSyncByFileReq{
			FileContent:    []byte(req.FileContent),
			Filename:       "upload.txt", // 默认文件名
			QueryTimeRange: req.QueryTimeRange,
			EsEndpoint:     req.EsEndpoint,
			TaskName:       req.TaskName,
		}

		rpcResp, err := l.svcCtx.CmpoolRpc.ExecuteEsSyncByFile(l.ctx, rpcReq)
		if err != nil {
			l.Logger.Errorf("调用ES文件同步RPC失败: %v", err)
			return &types.ExecuteExternalSyncResponse{
				Success: false,
				Message: "执行ES文件同步失败",
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
			NotInPoolCount:        rpcResp.NotInPoolCount,
			NotInDatasourceCount:  0,
			SuccessIpList:         rpcResp.SuccessIpList,
			FailedIpList:          rpcResp.FailedIpList,
			NotInPoolIpList:       rpcResp.NotInPoolIpList,
			NotInDatasourceIpList: nil,
		}, nil
	}

	// 调用CMSys文件同步RPC
	l.Logger.Infof("数据源路由: %s -> ExecuteCMSysSyncByFile RPC方法", normalizedSource)

	rpcReq := &cmpool.ExecuteCMSysSyncByFileReq{
		TaskName:    req.TaskName,
		FileContent: []byte(req.FileContent),
		Query:       req.Query,
	}

	rpcResp, err := l.svcCtx.CmpoolRpc.ExecuteCmsysSyncByFile(l.ctx, rpcReq)
	if err != nil {
		l.Logger.Errorf("调用CMSys文件同步RPC失败: %v", err)
		return &types.ExecuteExternalSyncResponse{
			Success: false,
			Message: "执行CMSys文件同步失败",
		}, nil
	}

	// CMSys响应 -> 统一响应（所有字段都存在）
	l.Logger.Infof("统一接口执行成功: ExecutionId=%d, Success=%d, Failed=%d, NotInDatasource=%d",
		rpcResp.ExecutionId, rpcResp.SuccessCount, rpcResp.FailedCount, rpcResp.NotInDatasourceCount)

	return &types.ExecuteExternalSyncResponse{
		Success:               rpcResp.Success,
		Message:               rpcResp.Message,
		DataSource:            "cmsys",
		ExecutionId:           rpcResp.ExecutionId,
		TotalHosts:            rpcResp.TotalHosts,
		SuccessCount:          rpcResp.SuccessCount,
		FailedCount:           rpcResp.FailedCount,
		// CMSys字段
		NotInDatasourceCount:  rpcResp.NotInDatasourceCount,
		NotInDatasourceIpList: rpcResp.NotInDatasourceIpList,
		// ES字段（CMSys时固定为0/空）
		NotInPoolCount:        0,
		NotInPoolIpList:       []string{},
		NewHostsCount:         0,
		NewHostIpList:         []string{},
		UpdatedHostsCount:     0,
		UpdatedHostIpList:     []string{},
		// 通用字段
		SuccessIpList:         rpcResp.SuccessIpList,
		FailedIpList:          rpcResp.FailedIpList,
	}, nil
}
