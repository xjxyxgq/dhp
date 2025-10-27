package logic

import (
	"context"

	"cmdb-api/internal/svc"
	"cmdb-api/internal/types"
	"cmdb-api/cmpool"

	"github.com/zeromicro/go-zero/core/logx"
)

type DeleteScheduledTaskLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewDeleteScheduledTaskLogic(ctx context.Context, svcCtx *svc.ServiceContext) *DeleteScheduledTaskLogic {
	return &DeleteScheduledTaskLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *DeleteScheduledTaskLogic) DeleteScheduledTask(req *types.DeleteScheduledTaskRequest) (resp *types.DeleteScheduledTaskResponse, err error) {
	// 调用RPC服务删除定时任务
	rpcReq := &cmpool.DeleteScheduledTaskReq{
		Id: int64(req.Id),
	}

	rpcResp, err := l.svcCtx.CmpoolRpc.DeleteScheduledTask(l.ctx, rpcReq)
	if err != nil {
		l.Errorf("调用RPC删除定时任务失败: %v", err)
		return &types.DeleteScheduledTaskResponse{
			Success: false,
			Message: "RPC调用失败",
		}, nil
	}

	return &types.DeleteScheduledTaskResponse{
		Success: rpcResp.Success,
		Message: rpcResp.Message,
	}, nil
}
