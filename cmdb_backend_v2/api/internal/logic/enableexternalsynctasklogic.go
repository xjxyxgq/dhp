package logic

import (
	"context"

	"cmdb-api/cmpool"
	"cmdb-api/internal/svc"
	"cmdb-api/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type EnableExternalSyncTaskLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewEnableExternalSyncTaskLogic(ctx context.Context, svcCtx *svc.ServiceContext) *EnableExternalSyncTaskLogic {
	return &EnableExternalSyncTaskLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *EnableExternalSyncTaskLogic) EnableExternalSyncTask(req *types.EnableExternalSyncTaskRequest) (resp *types.EnableExternalSyncTaskResponse, err error) {
	l.Logger.Infof("统一接口: 启用/禁用外部同步任务, TaskId=%d, IsEnabled=%v", req.Id, req.IsEnabled)

	// 调用统一的外部资源同步任务启用/禁用RPC（支持 ES 和 CMSys）
	l.Logger.Infof("数据源路由: EnableExternalSyncTask RPC方法")

	rpcReq := &cmpool.EnableExternalSyncTaskReq{
		Id:        req.Id,
		IsEnabled: req.IsEnabled,
	}

	rpcResp, err := l.svcCtx.CmpoolRpc.EnableExternalSyncTask(l.ctx, rpcReq)
	if err != nil {
		l.Logger.Errorf("调用统一外部同步任务启用/禁用RPC失败: %v", err)
		return &types.EnableExternalSyncTaskResponse{
			Success: false,
			Message: "启用/禁用外部同步任务失败",
		}, nil
	}

	l.Logger.Infof("统一接口执行成功: TaskId=%d, IsEnabled=%v", req.Id, req.IsEnabled)
	return &types.EnableExternalSyncTaskResponse{
		Success: rpcResp.Success,
		Message: rpcResp.Message,
	}, nil
}
