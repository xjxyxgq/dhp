package logic

import (
	"context"

	"cmdb-rpc/cmpool"
	"cmdb-rpc/internal/global"
	"cmdb-rpc/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
)

type DeleteScheduledTaskLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewDeleteScheduledTaskLogic(ctx context.Context, svcCtx *svc.ServiceContext) *DeleteScheduledTaskLogic {
	return &DeleteScheduledTaskLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// DeleteScheduledTask 删除定时任务
func (l *DeleteScheduledTaskLogic) DeleteScheduledTask(in *cmpool.DeleteScheduledTaskReq) (*cmpool.DeleteScheduledTaskResp, error) {
	// 先从调度器中移除任务
	taskScheduler := global.GetTaskScheduler()
	if taskScheduler != nil {
		err := taskScheduler.RemoveTask(in.Id)
		if err != nil {
			l.Errorf("从调度器移除任务失败: %v", err)
			// 继续执行，因为数据库删除更重要
		}
	}

	// 从数据库删除任务
	err := l.svcCtx.ScheduledTaskModel.Delete(l.ctx, in.Id)
	if err != nil {
		l.Errorf("删除定时任务失败: %v", err)
		return &cmpool.DeleteScheduledTaskResp{
			Success: false,
			Message: "删除任务失败",
		}, nil
	}

	return &cmpool.DeleteScheduledTaskResp{
		Success: true,
		Message: "任务删除成功",
	}, nil
}
