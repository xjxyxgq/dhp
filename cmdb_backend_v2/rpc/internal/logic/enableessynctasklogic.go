package logic

import (
	"context"
	"fmt"

	"cmdb-rpc/cmpool"
	"cmdb-rpc/internal/scheduler"
	"cmdb-rpc/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
)

type EnableEsSyncTaskLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewEnableEsSyncTaskLogic(ctx context.Context, svcCtx *svc.ServiceContext) *EnableEsSyncTaskLogic {
	return &EnableEsSyncTaskLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 启用/禁用ES数据同步任务
func (l *EnableEsSyncTaskLogic) EnableEsSyncTask(in *cmpool.EnableESSyncTaskReq) (*cmpool.EnableESSyncTaskResp, error) {
	// 1. 验证参数
	if in.Id == 0 {
		return &cmpool.EnableESSyncTaskResp{
			Success: false,
			Message: "任务ID不能为空",
		}, nil
	}

	// 2. 检查任务是否存在
	task, err := l.svcCtx.ExternalSyncTaskConfigModel.FindOne(l.ctx, uint64(in.Id))
	if err != nil {
		l.Logger.Errorf("查询任务失败: %v", err)
		return &cmpool.EnableESSyncTaskResp{
			Success: false,
			Message: "任务不存在或已被删除",
		}, nil
	}

	// 3. 更新启用状态 - 使用 Model 方法
	err = l.svcCtx.ExternalSyncTaskConfigModel.UpdateEnabledStatus(l.ctx, uint64(in.Id), in.IsEnabled)
	if err != nil {
		l.Logger.Errorf("更新任务启用状态失败: %v", err)
		return &cmpool.EnableESSyncTaskResp{
			Success: false,
			Message: "更新任务状态失败",
		}, nil
	}

	// 4. 根据状态注册/注销调度器任务
	if l.svcCtx.EsSyncScheduler != nil {
		// 类型断言
		if esSyncScheduler, ok := l.svcCtx.EsSyncScheduler.(*scheduler.EsSyncScheduler); ok {
			if in.IsEnabled {
				// 启用任务 - 注册到调度器
				if err := esSyncScheduler.RegisterTask(task); err != nil {
					l.Logger.Errorf("注册任务到调度器失败: TaskId=%d, Error=%v", in.Id, err)
					// 注册失败不影响数据库更新，继续执行
				} else {
					l.Logger.Infof("任务已注册到调度器: TaskId=%d", in.Id)
				}
			} else {
				// 禁用任务 - 从调度器注销
				esSyncScheduler.UnregisterTask(uint64(in.Id))
				l.Logger.Infof("任务已从调度器注销: TaskId=%d", in.Id)
			}
		}
	}

	statusText := "禁用"
	if in.IsEnabled {
		statusText = "启用"
	}

	l.Logger.Infof("%sES同步任务成功: ID=%d, Name=%s", statusText, task.Id, task.TaskName)

	return &cmpool.EnableESSyncTaskResp{
		Success: true,
		Message: fmt.Sprintf("%s任务成功", statusText),
	}, nil
}
