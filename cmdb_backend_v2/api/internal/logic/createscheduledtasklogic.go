package logic

import (
	"context"

	"cmdb-api/internal/svc"
	"cmdb-api/internal/types"
	"cmdb-api/cmpool"

	"github.com/zeromicro/go-zero/core/logx"
)

type CreateScheduledTaskLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewCreateScheduledTaskLogic(ctx context.Context, svcCtx *svc.ServiceContext) *CreateScheduledTaskLogic {
	return &CreateScheduledTaskLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *CreateScheduledTaskLogic) CreateScheduledTask(req *types.CreateScheduledTaskRequest) (resp *types.CreateScheduledTaskResponse, err error) {

	// 调用RPC服务创建定时任务
	rpcReq := &cmpool.CreateScheduledTaskReq{
		TaskName:       req.TaskName,
		Description:    req.Description,
		CronExpression: req.CronExpression,
		HostIpList:     req.HostIpList,
		ResourceType:   req.ResourceType,
		TargetPercent:  int32(req.TargetPercent),
		Duration:       int32(req.Duration),
		ScriptParams:   req.ScriptParams,
		ForceExecution: req.ForceExecution,
		CreatedBy:      req.CreatedBy,
	}

	rpcResp, err := l.svcCtx.CmpoolRpc.CreateScheduledTask(l.ctx, rpcReq)
	if err != nil {
		l.Errorf("调用RPC创建定时任务失败: %v", err)
		return &types.CreateScheduledTaskResponse{
			Success: false,
			Message: "RPC调用失败",
		}, nil
	}

	return &types.CreateScheduledTaskResponse{
		Success: rpcResp.Success,
		Message: rpcResp.Message,
		TaskId:  int(rpcResp.TaskId),
	}, nil
}
