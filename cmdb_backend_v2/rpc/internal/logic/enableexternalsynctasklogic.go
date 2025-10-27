package logic

import (
	"context"
	"database/sql"

	"cmdb-rpc/cmpool"
	"cmdb-rpc/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
)

type EnableExternalSyncTaskLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewEnableExternalSyncTaskLogic(ctx context.Context, svcCtx *svc.ServiceContext) *EnableExternalSyncTaskLogic {
	return &EnableExternalSyncTaskLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// EnableExternalSyncTask 启用/禁用外部同步任务
func (l *EnableExternalSyncTaskLogic) EnableExternalSyncTask(in *cmpool.EnableExternalSyncTaskReq) (*cmpool.EnableExternalSyncTaskResp, error) {
	// 1. 验证任务ID
	if in.Id <= 0 {
		return &cmpool.EnableExternalSyncTaskResp{
			Success: false,
			Message: "任务ID不能为空",
		}, nil
	}

	// 2. 查询任务是否存在
	existingTask, err := l.svcCtx.ExternalSyncTaskConfigModel.FindOne(l.ctx, uint64(in.Id))
	if err != nil {
		if err == sql.ErrNoRows {
			return &cmpool.EnableExternalSyncTaskResp{
				Success: false,
				Message: "任务不存在",
			}, nil
		}
		l.Logger.Errorf("查询任务失败: %v", err)
		return &cmpool.EnableExternalSyncTaskResp{
			Success: false,
			Message: "查询任务失败",
		}, nil
	}

	// 3. 更新启用状态
	err = l.svcCtx.ExternalSyncTaskConfigModel.UpdateEnabledStatus(l.ctx, uint64(in.Id), in.IsEnabled)
	if err != nil {
		l.Logger.Errorf("更新任务启用状态失败: %v", err)
		return &cmpool.EnableExternalSyncTaskResp{
			Success: false,
			Message: "更新任务状态失败",
		}, nil
	}

	// TODO: 4. 根据状态注册/移除调度任务
	// if in.IsEnabled {
	//     // 启用：注册到调度器
	//     if existingTask.DataSource == "elasticsearch" && l.svcCtx.ESSyncScheduler != nil {
	//         l.svcCtx.ESSyncScheduler.RegisterTask(in.Id, existingTask.CronExpression)
	//     } else if existingTask.DataSource == "cmsys" && l.svcCtx.CMSysSyncScheduler != nil {
	//         l.svcCtx.CMSysSyncScheduler.RegisterTask(in.Id, existingTask.CronExpression)
	//     }
	// } else {
	//     // 禁用：从调度器移除
	//     if existingTask.DataSource == "elasticsearch" && l.svcCtx.ESSyncScheduler != nil {
	//         l.svcCtx.ESSyncScheduler.RemoveTask(in.Id)
	//     } else if existingTask.DataSource == "cmsys" && l.svcCtx.CMSysSyncScheduler != nil {
	//         l.svcCtx.CMSysSyncScheduler.RemoveTask(in.Id)
	//     }
	// }

	statusText := "禁用"
	if in.IsEnabled {
		statusText = "启用"
	}

	l.Logger.Infof("%s外部同步任务成功: TaskID=%d, Name=%s, DataSource=%s",
		statusText, in.Id, existingTask.TaskName, existingTask.DataSource)

	return &cmpool.EnableExternalSyncTaskResp{
		Success: true,
		Message: statusText + "任务成功",
	}, nil
}
