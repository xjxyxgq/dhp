package logic

import (
	"context"

	"cmdb-api/internal/svc"
	"cmdb-api/internal/types"
	"cmdb-api/cmpool"

	"github.com/zeromicro/go-zero/core/logx"
)

type UpdateScheduledTaskLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewUpdateScheduledTaskLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UpdateScheduledTaskLogic {
	return &UpdateScheduledTaskLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *UpdateScheduledTaskLogic) UpdateScheduledTask(req *types.UpdateScheduledTaskRequest) (resp *types.UpdateScheduledTaskResponse, err error) {
	// 调用RPC服务更新定时任务
	rpcReq := &cmpool.UpdateScheduledTaskReq{
		Id:             int64(req.Id),
		TaskName:       req.TaskName,
		Description:    req.Description,
		CronExpression: req.CronExpression,
		HostIpList:     req.HostIpList,
		ResourceType:   req.ResourceType,
		TargetPercent:  int32(req.TargetPercent),
		Duration:       int32(req.Duration),
		ScriptParams:   req.ScriptParams,
		ForceExecution: req.ForceExecution,
	}

	rpcResp, err := l.svcCtx.CmpoolRpc.UpdateScheduledTask(l.ctx, rpcReq)
	if err != nil {
		l.Errorf("调用RPC更新定时任务失败: %v", err)
		return &types.UpdateScheduledTaskResponse{
			Success: false,
			Message: "RPC调用失败",
		}, nil
	}

	return &types.UpdateScheduledTaskResponse{
		Success: rpcResp.Success,
		Message: rpcResp.Message,
	}, nil
}
