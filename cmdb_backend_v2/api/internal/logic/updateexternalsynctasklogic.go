package logic

import (
	"context"

	"cmdb-api/cmpool"
	"cmdb-api/internal/svc"
	"cmdb-api/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type UpdateExternalSyncTaskLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewUpdateExternalSyncTaskLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UpdateExternalSyncTaskLogic {
	return &UpdateExternalSyncTaskLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *UpdateExternalSyncTaskLogic) UpdateExternalSyncTask(req *types.UpdateExternalSyncTaskRequest) (resp *types.UpdateExternalSyncTaskResponse, err error) {
	// 参数验证
	if err := validateDataSource(req.DataSource); err != nil {
		l.Logger.Errorf("参数验证失败: %v", err)
		return &types.UpdateExternalSyncTaskResponse{
			Success: false,
			Message: err.Error(),
		}, nil
	}

	normalizedSource := normalizeDataSource(req.DataSource)
	l.Logger.Infof("统一接口: 更新外部同步任务, DataSource=%s (标准化后=%s), TaskId=%d", req.DataSource, normalizedSource, req.TaskId)

	// 调用统一的外部资源同步任务更新RPC（支持 ES 和 CMSys）
	l.Logger.Infof("数据源路由: %s -> UpdateExternalSyncTask RPC方法", normalizedSource)

	rpcReq := &cmpool.UpdateExternalSyncTaskReq{
		Id:             req.TaskId,
		DataSource:     normalizedSource,
		TaskName:       req.TaskName,
		Description:    req.Description,
		EsEndpoint:     req.EsEndpoint,
		EsIndexPattern: req.EsIndexPattern,
		CronExpression: req.CronExpression,
		QueryTimeRange: req.QueryTimeRange,
		CmsysQuery:     req.CmsysQuery,
	}

	rpcResp, err := l.svcCtx.CmpoolRpc.UpdateExternalSyncTask(l.ctx, rpcReq)
	if err != nil {
		l.Logger.Errorf("调用统一外部同步任务更新RPC失败: %v", err)
		return &types.UpdateExternalSyncTaskResponse{
			Success: false,
			Message: "更新外部同步任务失败",
		}, nil
	}

	l.Logger.Infof("统一接口执行成功: TaskId=%d, DataSource=%s", req.TaskId, normalizedSource)
	return &types.UpdateExternalSyncTaskResponse{
		Success: rpcResp.Success,
		Message: rpcResp.Message,
	}, nil
}
