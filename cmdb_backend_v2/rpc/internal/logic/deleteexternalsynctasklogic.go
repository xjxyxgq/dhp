package logic

import (
	"context"
	"database/sql"

	"cmdb-rpc/cmpool"
	"cmdb-rpc/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
)

type DeleteExternalSyncTaskLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewDeleteExternalSyncTaskLogic(ctx context.Context, svcCtx *svc.ServiceContext) *DeleteExternalSyncTaskLogic {
	return &DeleteExternalSyncTaskLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// DeleteExternalSyncTask 删除外部同步任务配置（软删除）
func (l *DeleteExternalSyncTaskLogic) DeleteExternalSyncTask(in *cmpool.DeleteExternalSyncTaskReq) (*cmpool.DeleteExternalSyncTaskResp, error) {
	// 1. 验证任务ID
	if in.Id <= 0 {
		return &cmpool.DeleteExternalSyncTaskResp{
			Success: false,
			Message: "任务ID不能为空",
		}, nil
	}

	// 2. 查询任务是否存在
	existingTask, err := l.svcCtx.ExternalSyncTaskConfigModel.FindOne(l.ctx, uint64(in.Id))
	if err != nil {
		if err == sql.ErrNoRows {
			return &cmpool.DeleteExternalSyncTaskResp{
				Success: false,
				Message: "任务不存在",
			}, nil
		}
		l.Logger.Errorf("查询任务失败: %v", err)
		return &cmpool.DeleteExternalSyncTaskResp{
			Success: false,
			Message: "查询任务失败",
		}, nil
	}

	// 3. 软删除任务
	err = l.svcCtx.ExternalSyncTaskConfigModel.SoftDelete(l.ctx, uint64(in.Id))
	if err != nil {
		l.Logger.Errorf("删除任务失败: %v", err)
		return &cmpool.DeleteExternalSyncTaskResp{
			Success: false,
			Message: "删除任务失败",
		}, nil
	}

	// TODO: 4. 从调度器中移除任务
	// if existingTask.DataSource == "elasticsearch" && l.svcCtx.ESSyncScheduler != nil {
	//     l.svcCtx.ESSyncScheduler.RemoveTask(in.Id)
	// } else if existingTask.DataSource == "cmsys" && l.svcCtx.CMSysSyncScheduler != nil {
	//     l.svcCtx.CMSysSyncScheduler.RemoveTask(in.Id)
	// }

	l.Logger.Infof("删除外部同步任务成功: TaskID=%d, Name=%s, DataSource=%s",
		in.Id, existingTask.TaskName, existingTask.DataSource)

	return &cmpool.DeleteExternalSyncTaskResp{
		Success: true,
		Message: "删除任务成功",
	}, nil
}
