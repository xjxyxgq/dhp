package logic

import (
	"context"
	"fmt"

	"cmdb-rpc/cmpool"
	"cmdb-rpc/internal/global"
	"cmdb-rpc/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
)

type EnableScheduledTaskLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewEnableScheduledTaskLogic(ctx context.Context, svcCtx *svc.ServiceContext) *EnableScheduledTaskLogic {
	return &EnableScheduledTaskLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// EnableScheduledTask 启用/禁用定时任务
func (l *EnableScheduledTaskLogic) EnableScheduledTask(in *cmpool.EnableScheduledTaskReq) (*cmpool.EnableScheduledTaskResp, error) {
	// 更新数据库状态
	err := l.svcCtx.ScheduledTaskModel.EnableTask(in.Id, in.IsEnabled)
	if err != nil {
		l.Errorf("更新定时任务状态失败: %v", err)
		return &cmpool.EnableScheduledTaskResp{
			Success: false,
			Message: "更新任务状态失败",
		}, nil
	}

	// 获取任务详情用于调度器更新
	task, err := l.svcCtx.ScheduledTaskModel.FindOne(l.ctx, in.Id)
	if err != nil {
		l.Errorf("查找定时任务失败: %v", err)
		return &cmpool.EnableScheduledTaskResp{
			Success: false,
			Message: "查找任务失败",
		}, nil
	}

	// 更新任务的启用状态
	task.IsEnabled = func() int64 {
		if in.IsEnabled {
			return 1
		} else {
			return 0
		}
	}()

	// 更新调度器
	taskScheduler := global.GetTaskScheduler()
	if taskScheduler != nil {
		err = taskScheduler.UpdateTask(task)
		if err != nil {
			l.Errorf("更新调度器任务失败: %v", err)
			// 不返回错误，因为数据库已经更新成功
			l.Infof("定时任务状态更新成功但调度器更新失败，任务ID: %d", in.Id)
		}
	}

	status := "禁用"
	if in.IsEnabled {
		status = "启用"
	}

	return &cmpool.EnableScheduledTaskResp{
		Success: true,
		Message: fmt.Sprintf("任务已%s", status),
	}, nil
}
