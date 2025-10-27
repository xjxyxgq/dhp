package logic

import (
	"context"
	"database/sql"
	"encoding/json"

	"cmdb-rpc/cmpool"
	"cmdb-rpc/internal/global"
	"cmdb-rpc/internal/model"
	"cmdb-rpc/internal/svc"

	"github.com/robfig/cron/v3"
	"github.com/zeromicro/go-zero/core/logx"
)

type UpdateScheduledTaskLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewUpdateScheduledTaskLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UpdateScheduledTaskLogic {
	return &UpdateScheduledTaskLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 更新定时任务
func (l *UpdateScheduledTaskLogic) UpdateScheduledTask(in *cmpool.UpdateScheduledTaskReq) (*cmpool.UpdateScheduledTaskResp, error) {
	// 参数校验
	if in.Id <= 0 {
		l.Errorf("任务ID不能为空")
		return &cmpool.UpdateScheduledTaskResp{
			Success: false,
			Message: "参数错误：任务ID不能为空",
		}, nil
	}

	if in.TaskName == "" {
		l.Errorf("任务名称不能为空")
		return &cmpool.UpdateScheduledTaskResp{
			Success: false,
			Message: "参数错误：任务名称不能为空",
		}, nil
	}

	// 验证Cron表达式格式（如果提供了新的表达式）
	if in.CronExpression != "" {
		parser := cron.NewParser(cron.Second | cron.Minute | cron.Hour | cron.Dom | cron.Month | cron.Dow | cron.Descriptor)
		if _, err := parser.Parse(in.CronExpression); err != nil {
			l.Errorf("Cron表达式格式错误: %v", err)
			return &cmpool.UpdateScheduledTaskResp{
				Success: false,
				Message: "Cron表达式格式错误: " + err.Error() + "。正确格式为6字段（秒 分 时 日 月 周），例如: '0 0 2 * * *' 表示每天凌晨2点执行",
			}, nil
		}
	}

	// 检查任务是否存在
	existingTask, err := l.svcCtx.ScheduledTaskModel.FindOne(l.ctx, in.Id)
	if err != nil {
		l.Errorf("查询任务失败: %v", err)
		return &cmpool.UpdateScheduledTaskResp{
			Success: false,
			Message: "任务不存在",
		}, nil
	}

	// 将主机IP列表转换为JSON字符串
	hostIpListBytes, err := json.Marshal(in.HostIpList)
	if err != nil {
		l.Errorf("序列化主机IP列表失败: %v", err)
		return &cmpool.UpdateScheduledTaskResp{
			Success: false,
			Message: "参数错误",
		}, nil
	}

	// 准备更新数据
	updatedTask := &model.ScheduledHardwareVerification{
		Id:             in.Id,
		TaskName:       in.TaskName,
		Description:    sql.NullString{String: in.Description, Valid: in.Description != ""},
		CronExpression: in.CronExpression,
		HostIpList:     string(hostIpListBytes),
		ResourceType:   in.ResourceType,
		TargetPercent:  int64(in.TargetPercent),
		Duration:       int64(in.Duration),
		ScriptParams:   sql.NullString{String: in.ScriptParams, Valid: in.ScriptParams != ""},
		ForceExecution: func() int64 {
			if in.ForceExecution {
				return 1
			} else {
				return 0
			}
		}(),
		IsEnabled:  existingTask.IsEnabled,  // 保持原有的启用状态
		CreatedBy:  existingTask.CreatedBy,  // 保持原有的创建者
		CreateTime: existingTask.CreateTime, // 保持原有的创建时间
	}

	// 更新数据库
	err = l.svcCtx.ScheduledTaskModel.Update(l.ctx, updatedTask)
	if err != nil {
		l.Errorf("更新定时任务失败: %v", err)
		return &cmpool.UpdateScheduledTaskResp{
			Success: false,
			Message: "更新失败",
		}, nil
	}

	// 更新调度器
	taskScheduler := global.GetTaskScheduler()
	if taskScheduler != nil {
		err = taskScheduler.UpdateTask(updatedTask)
		if err != nil {
			l.Errorf("更新调度器任务失败: %v", err)
			// 不返回错误，因为数据库已经更新成功
			l.Infof("定时任务更新成功但调度器更新失败，任务ID: %d", in.Id)
		}
	}

	l.Infof("定时任务更新成功: %s (ID: %d)", in.TaskName, in.Id)
	return &cmpool.UpdateScheduledTaskResp{
		Success: true,
		Message: "任务更新成功",
	}, nil
}
