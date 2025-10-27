package logic

import (
	"context"

	"cmdb-rpc/cmpool"
	"cmdb-rpc/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetScheduledTasksLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGetScheduledTasksLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetScheduledTasksLogic {
	return &GetScheduledTasksLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// GetScheduledTasks 获取定时任务列表
func (l *GetScheduledTasksLogic) GetScheduledTasks(in *cmpool.GetScheduledTasksReq) (*cmpool.GetScheduledTasksResp, error) {
	// 查询定时任务
	tasks, err := l.svcCtx.ScheduledTaskModel.FindAll(in.ResourceType, in.EnabledOnly)
	if err != nil {
		l.Errorf("查询定时任务列表失败: %v", err)
		return &cmpool.GetScheduledTasksResp{
			Success: false,
			Message: "查询定时任务列表失败",
			Tasks:   nil,
		}, nil
	}

	// 转换为响应格式
	var responseTasks []*cmpool.ScheduledTask
	for _, task := range tasks {
		responseTask := &cmpool.ScheduledTask{
			Id:             task.Id,
			TaskName:       task.TaskName,
			Description:    task.Description.String,
			CronExpression: task.CronExpression,
			HostIpList:     task.HostIpList,
			ResourceType:   task.ResourceType,
			TargetPercent:  int32(task.TargetPercent),
			Duration:       int32(task.Duration),
			ScriptParams:   task.ScriptParams.String,
			ForceExecution: func() bool {
				if task.ForceExecution > 0 {
					return true
				} else {
					return false
				}
			}(),
			IsEnabled: func() bool {
				if task.IsEnabled > 0 {
					return true
				} else {
					return false
				}
			}(),
			CreatedBy: task.CreatedBy.String,
			CreatedAt: task.CreateTime.Format("2006-01-02 15:04:05"),
			UpdatedAt: task.UpdateTime.Format("2006-01-02 15:04:05"),
		}

		if task.LastExecutionTime.Valid {
			responseTask.LastExecutionTime = task.LastExecutionTime.Time.Format("2006-01-02 15:04:05")
		}

		if task.NextExecutionTime.Valid {
			responseTask.NextExecutionTime = task.NextExecutionTime.Time.Format("2006-01-02 15:04:05")
		}

		responseTasks = append(responseTasks, responseTask)
	}

	return &cmpool.GetScheduledTasksResp{
		Success: true,
		Message: "查询成功",
		Tasks:   responseTasks,
	}, nil
}
