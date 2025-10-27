package logic

import (
	"context"
	"database/sql"

	"cmdb-rpc/cmpool"
	"cmdb-rpc/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetScheduledTaskDetailLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGetScheduledTaskDetailLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetScheduledTaskDetailLogic {
	return &GetScheduledTaskDetailLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 获取定时任务详情
func (l *GetScheduledTaskDetailLogic) GetScheduledTaskDetail(in *cmpool.GetScheduledTaskDetailReq) (*cmpool.GetScheduledTaskDetailResp, error) {
	// 参数校验
	if in.Id <= 0 {
		l.Errorf("定时任务ID不能为空或小于等于0")
		return &cmpool.GetScheduledTaskDetailResp{
			Success: false,
			Message: "参数错误：任务ID不能为空",
		}, nil
	}

	// 查询定时任务详情
	task, err := l.svcCtx.ScheduledTaskModel.FindOne(l.ctx, in.Id)
	if err != nil {
		l.Errorf("查询定时任务失败: %v", err)
		return &cmpool.GetScheduledTaskDetailResp{
			Success: false,
			Message: "查询任务失败或任务不存在",
		}, nil
	}

	// 转换为响应格式
	taskDetail := &cmpool.ScheduledTask{
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
		CreatedBy:         task.CreatedBy.String,
		CreatedAt:         task.CreateTime.Format("2006-01-02 15:04:05"),
		UpdatedAt:         task.UpdateTime.Format("2006-01-02 15:04:05"),
		LastExecutionTime: formatNullTime(task.LastExecutionTime),
		NextExecutionTime: formatNullTime(task.NextExecutionTime),
	}

	return &cmpool.GetScheduledTaskDetailResp{
		Success: true,
		Message: "查询成功",
		Task:    taskDetail,
	}, nil
}

// formatNullTime 格式化sql.NullTime为字符串
func formatNullTime(nullTime sql.NullTime) string {
	if nullTime.Valid {
		return nullTime.Time.Format("2006-01-02 15:04:05")
	}
	return ""
}
