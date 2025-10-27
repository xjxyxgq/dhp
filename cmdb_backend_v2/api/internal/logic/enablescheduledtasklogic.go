package logic

import (
	"context"

	"cmdb-api/internal/svc"
	"cmdb-api/internal/types"
	"cmdb-api/cmpool"

	"github.com/zeromicro/go-zero/core/logx"
)

type EnableScheduledTaskLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewEnableScheduledTaskLogic(ctx context.Context, svcCtx *svc.ServiceContext) *EnableScheduledTaskLogic {
	return &EnableScheduledTaskLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *EnableScheduledTaskLogic) EnableScheduledTask(req *types.EnableScheduledTaskRequest) (resp *types.EnableScheduledTaskResponse, err error) {
	// 调用RPC服务启用/禁用定时任务
	rpcReq := &cmpool.EnableScheduledTaskReq{
		Id:        int64(req.Id),
		IsEnabled: req.IsEnabled,
	}

	rpcResp, err := l.svcCtx.CmpoolRpc.EnableScheduledTask(l.ctx, rpcReq)
	if err != nil {
		l.Errorf("调用RPC启用/禁用定时任务失败: %v", err)
		return &types.EnableScheduledTaskResponse{
			Success: false,
			Message: "RPC调用失败",
		}, nil
	}

	return &types.EnableScheduledTaskResponse{
		Success: rpcResp.Success,
		Message: rpcResp.Message,
	}, nil
}
