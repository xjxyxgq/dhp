package logic

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"

	"cmdb-rpc/cmpool"
	"cmdb-rpc/internal/global"
	"cmdb-rpc/internal/model"
	"cmdb-rpc/internal/svc"

	"github.com/robfig/cron/v3"
	"github.com/zeromicro/go-zero/core/logx"
)

type CreateScheduledTaskLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewCreateScheduledTaskLogic(ctx context.Context, svcCtx *svc.ServiceContext) *CreateScheduledTaskLogic {
	return &CreateScheduledTaskLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// CreateScheduledTask 创建定时任务
func (l *CreateScheduledTaskLogic) CreateScheduledTask(in *cmpool.CreateScheduledTaskReq) (*cmpool.CreateScheduledTaskResp, error) {
	// 验证参数
	if in.TaskName == "" {
		return &cmpool.CreateScheduledTaskResp{
			Success: false,
			Message: "任务名称不能为空",
		}, nil
	}

	if in.CronExpression == "" {
		return &cmpool.CreateScheduledTaskResp{
			Success: false,
			Message: "Cron表达式不能为空",
		}, nil
	}

	// 验证Cron表达式格式（使用支持秒级精度的解析器，6字段格式）
	parser := cron.NewParser(cron.Second | cron.Minute | cron.Hour | cron.Dom | cron.Month | cron.Dow | cron.Descriptor)
	if _, err := parser.Parse(in.CronExpression); err != nil {
		return &cmpool.CreateScheduledTaskResp{
			Success: false,
			Message: "Cron表达式格式错误: " + err.Error() + "。正确格式为6字段（秒 分 时 日 月 周），例如: '0 0 2 * * *' 表示每天凌晨2点执行",
		}, nil
	}

	if len(in.HostIpList) == 0 {
		return &cmpool.CreateScheduledTaskResp{
			Success: false,
			Message: "主机IP列表不能为空",
		}, nil
	}

	if in.ResourceType != "cpu" && in.ResourceType != "memory" && in.ResourceType != "disk" {
		return &cmpool.CreateScheduledTaskResp{
			Success: false,
			Message: "资源类型必须是 cpu、memory 或 disk",
		}, nil
	}

	if in.TargetPercent <= 0 || in.TargetPercent > 100 {
		return &cmpool.CreateScheduledTaskResp{
			Success: false,
			Message: "目标百分比必须在1-100之间",
		}, nil
	}

	if in.Duration <= 0 {
		return &cmpool.CreateScheduledTaskResp{
			Success: false,
			Message: "执行持续时间必须大于0",
		}, nil
	}

	// 将主机IP列表转换为JSON
	hostIpListJson, err := json.Marshal(in.HostIpList)
	if err != nil {
		return &cmpool.CreateScheduledTaskResp{
			Success: false,
			Message: fmt.Sprintf("主机IP列表格式错误: %v", err),
		}, nil
	}

	// 创建定时任务对象
	task := &model.ScheduledHardwareVerification{
		TaskName:       in.TaskName,
		Description:    sql.NullString{String: in.Description, Valid: in.Description != ""},
		CronExpression: in.CronExpression,
		HostIpList:     string(hostIpListJson),
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
		IsEnabled: 1, // 默认启用
		CreatedBy: sql.NullString{String: in.CreatedBy, Valid: in.CreatedBy != ""},
	}

	// 插入数据库
	result, err := l.svcCtx.ScheduledTaskModel.Insert(l.ctx, task)
	if err != nil {
		l.Errorf("插入定时任务失败: %v", err)
		return &cmpool.CreateScheduledTaskResp{
			Success: false,
			Message: fmt.Sprintf("创建定时任务失败: %v", err),
		}, nil
	}

	taskId, err := result.LastInsertId()
	if err != nil {
		l.Errorf("获取任务ID失败: %v", err)
		return &cmpool.CreateScheduledTaskResp{
			Success: false,
			Message: "创建定时任务失败，无法获取任务ID",
		}, nil
	}

	task.Id = taskId

	// 添加到调度器
	taskScheduler := global.GetTaskScheduler()
	if taskScheduler != nil {
		err = taskScheduler.AddTask(task)
		if err != nil {
			l.Errorf("添加任务到调度器失败: %v", err)
			// 不返回错误，因为任务已经创建成功，只是调度器添加失败
			l.Infof("定时任务创建成功但添加到调度器失败，任务ID: %d", taskId)
		}
	} else {
		l.Errorf("TaskScheduler 未初始化")
	}

	return &cmpool.CreateScheduledTaskResp{
		Success: true,
		Message: "定时任务创建成功",
		TaskId:  taskId,
	}, nil
}
