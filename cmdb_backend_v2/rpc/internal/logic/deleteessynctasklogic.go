package logic

import (
	"context"

	"cmdb-rpc/cmpool"
	"cmdb-rpc/internal/scheduler"
	"cmdb-rpc/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
)

type DeleteEsSyncTaskLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewDeleteEsSyncTaskLogic(ctx context.Context, svcCtx *svc.ServiceContext) *DeleteEsSyncTaskLogic {
	return &DeleteEsSyncTaskLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 删除ES数据同步任务配置
func (l *DeleteEsSyncTaskLogic) DeleteEsSyncTask(in *cmpool.DeleteESSyncTaskReq) (*cmpool.DeleteESSyncTaskResp, error) {
	// 1. 验证参数
	if in.Id == 0 {
		return &cmpool.DeleteESSyncTaskResp{
			Success: false,
			Message: "任务ID不能为空",
		}, nil
	}

	// 2. 检查任务是否存在
	task, err := l.svcCtx.ExternalSyncTaskConfigModel.FindOne(l.ctx, uint64(in.Id))
	if err != nil {
		l.Logger.Errorf("查询任务失败: %v", err)
		return &cmpool.DeleteESSyncTaskResp{
			Success: false,
			Message: "任务不存在或已被删除",
		}, nil
	}

	// 3. 软删除任务 - 使用 Model 方法
	err = l.svcCtx.ExternalSyncTaskConfigModel.SoftDelete(l.ctx, uint64(in.Id))
	if err != nil {
		l.Logger.Errorf("删除任务失败: %v", err)
		return &cmpool.DeleteESSyncTaskResp{
			Success: false,
			Message: "删除任务失败",
		}, nil
	}

	// 4. 从调度器中注销任务（如果任务已启用）
	if l.svcCtx.EsSyncScheduler != nil && task.IsEnabled == 1 {
		// 类型断言
		if esSyncScheduler, ok := l.svcCtx.EsSyncScheduler.(*scheduler.EsSyncScheduler); ok {
			esSyncScheduler.UnregisterTask(uint64(in.Id))
			l.Logger.Infof("任务已从调度器注销: TaskId=%d", in.Id)
		}
	}

	l.Logger.Infof("删除ES同步任务成功: ID=%d, Name=%s", task.Id, task.TaskName)

	return &cmpool.DeleteESSyncTaskResp{
		Success: true,
		Message: "删除任务成功",
	}, nil
}
