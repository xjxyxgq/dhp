package logic

import (
	"context"

	"cmdb-api/cmpool"
	"cmdb-api/internal/svc"
	"cmdb-api/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type DeleteExternalSyncTaskLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewDeleteExternalSyncTaskLogic(ctx context.Context, svcCtx *svc.ServiceContext) *DeleteExternalSyncTaskLogic {
	return &DeleteExternalSyncTaskLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *DeleteExternalSyncTaskLogic) DeleteExternalSyncTask(req *types.DeleteExternalSyncTaskRequest) (resp *types.DeleteExternalSyncTaskResponse, err error) {
	l.Logger.Infof("统一接口: 删除外部同步任务, TaskId=%d", req.TaskId)

	// 调用统一的外部资源同步任务删除RPC（支持 ES 和 CMSys）
	l.Logger.Infof("数据源路由: DeleteExternalSyncTask RPC方法")

	rpcReq := &cmpool.DeleteExternalSyncTaskReq{
		Id: req.TaskId,
	}

	rpcResp, err := l.svcCtx.CmpoolRpc.DeleteExternalSyncTask(l.ctx, rpcReq)
	if err != nil {
		l.Logger.Errorf("调用统一外部同步任务删除RPC失败: %v", err)
		return &types.DeleteExternalSyncTaskResponse{
			Success: false,
			Message: "删除外部同步任务失败",
		}, nil
	}

	l.Logger.Infof("统一接口执行成功: TaskId=%d", req.TaskId)
	return &types.DeleteExternalSyncTaskResponse{
		Success: rpcResp.Success,
		Message: rpcResp.Message,
	}, nil
}
