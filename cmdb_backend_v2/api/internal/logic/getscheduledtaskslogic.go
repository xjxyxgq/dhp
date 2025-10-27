package logic

import (
	"context"

	"cmdb-api/internal/svc"
	"cmdb-api/internal/types"
	"cmdb-api/cmpool"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetScheduledTasksLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewGetScheduledTasksLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetScheduledTasksLogic {
	return &GetScheduledTasksLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetScheduledTasksLogic) GetScheduledTasks(req *types.GetScheduledTasksRequest) (resp *types.GetScheduledTasksResponse, err error) {
	// 调用RPC服务获取定时任务列表
	rpcReq := &cmpool.GetScheduledTasksReq{}

	rpcResp, err := l.svcCtx.CmpoolRpc.GetScheduledTasks(l.ctx, rpcReq)
	if err != nil {
		l.Errorf("调用RPC获取定时任务列表失败: %v", err)
		return &types.GetScheduledTasksResponse{
			Success: false,
			Message: "RPC调用失败",
			Tasks:   []types.ScheduledTaskInfo{},
		}, nil
	}

	// 转换RPC响应为API响应格式
	var tasks []types.ScheduledTaskInfo
	for _, rpcTask := range rpcResp.Tasks {
		task := types.ScheduledTaskInfo{
			Id:                int(rpcTask.Id),
			TaskName:          rpcTask.TaskName,
			Description:       rpcTask.Description,
			CronExpression:    rpcTask.CronExpression,
			HostIpList:        rpcTask.HostIpList,
			ResourceType:      rpcTask.ResourceType,
			TargetPercent:     int(rpcTask.TargetPercent),
			Duration:          int(rpcTask.Duration),
			ScriptParams:      rpcTask.ScriptParams,
			ForceExecution:    rpcTask.ForceExecution,
			IsEnabled:         rpcTask.IsEnabled,
			CreatedBy:         rpcTask.CreatedBy,
			CreatedAt:         rpcTask.CreatedAt,
			UpdatedAt:         rpcTask.UpdatedAt,
			LastExecutionTime: rpcTask.LastExecutionTime,
			NextExecutionTime: rpcTask.NextExecutionTime,
		}
		tasks = append(tasks, task)
	}

	return &types.GetScheduledTasksResponse{
		Success: rpcResp.Success,
		Message: rpcResp.Message,
		Tasks:   tasks,
	}, nil
}
