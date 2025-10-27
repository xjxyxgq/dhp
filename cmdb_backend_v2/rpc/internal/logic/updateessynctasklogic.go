package logic

import (
	"context"
	"database/sql"

	"cmdb-rpc/cmpool"
	"cmdb-rpc/internal/scheduler"
	"cmdb-rpc/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
)

type UpdateEsSyncTaskLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewUpdateEsSyncTaskLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UpdateEsSyncTaskLogic {
	return &UpdateEsSyncTaskLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 更新ES数据同步任务配置
func (l *UpdateEsSyncTaskLogic) UpdateEsSyncTask(in *cmpool.UpdateESSyncTaskReq) (*cmpool.UpdateESSyncTaskResp, error) {
	// 1. 验证参数
	if in.Id == 0 {
		return &cmpool.UpdateESSyncTaskResp{
			Success: false,
			Message: "任务ID不能为空",
		}, nil
	}

	if in.TaskName == "" {
		return &cmpool.UpdateESSyncTaskResp{
			Success: false,
			Message: "任务名称不能为空",
		}, nil
	}

	// 2. 检查任务是否存在 - 使用 Model 方法
	existingTask, err := l.svcCtx.ExternalSyncTaskConfigModel.FindOne(l.ctx, uint64(in.Id))
	if err != nil {
		l.Logger.Errorf("查询任务失败: %v", err)
		return &cmpool.UpdateESSyncTaskResp{
			Success: false,
			Message: "任务不存在",
		}, nil
	}

	// 3. 检查任务名称是否与其他任务重复 - 使用 Model 方法
	exists, err := l.svcCtx.ExternalSyncTaskConfigModel.CheckTaskNameExists(l.ctx, in.TaskName, uint64(in.Id))
	if err != nil {
		l.Logger.Errorf("检查任务名称失败: %v", err)
		return &cmpool.UpdateESSyncTaskResp{
			Success: false,
			Message: "检查任务名称失败",
		}, nil
	}
	if exists {
		return &cmpool.UpdateESSyncTaskResp{
			Success: false,
			Message: "任务名称已被其他任务使用",
		}, nil
	}

	// 4. 更新任务配置 - 使用 Model 方法
	existingTask.TaskName = in.TaskName
	existingTask.Description = sql.NullString{String: in.Description, Valid: in.Description != ""}
	existingTask.EsEndpoint = sql.NullString{String: in.EsEndpoint, Valid: in.EsEndpoint != ""}
	existingTask.EsIndexPattern = sql.NullString{String: in.EsIndexPattern, Valid: in.EsIndexPattern != ""}
	existingTask.CronExpression = in.CronExpression
	existingTask.QueryTimeRange = in.QueryTimeRange

	err = l.svcCtx.ExternalSyncTaskConfigModel.Update(l.ctx, existingTask)
	if err != nil {
		l.Logger.Errorf("更新任务配置失败: %v", err)
		return &cmpool.UpdateESSyncTaskResp{
			Success: false,
			Message: "更新任务失败",
		}, nil
	}

	// 5. 如果任务已启用，重新注册到调度器以更新配置
	if l.svcCtx.EsSyncScheduler != nil && existingTask.IsEnabled == 1 {
		// 类型断言
		if esSyncScheduler, ok := l.svcCtx.EsSyncScheduler.(*scheduler.EsSyncScheduler); ok {
			// 先从调度器注销旧配置
			esSyncScheduler.UnregisterTask(uint64(in.Id))
			// 重新注册新配置
			if err := esSyncScheduler.RegisterTask(existingTask); err != nil {
				l.Logger.Errorf("重新注册任务到调度器失败: TaskId=%d, Error=%v", in.Id, err)
				// 注册失败不影响数据库更新，只记录日志
			} else {
				l.Logger.Infof("任务已重新注册到调度器: TaskId=%d", in.Id)
			}
		}
	}

	l.Logger.Infof("更新ES同步任务成功: ID=%d, Name=%s", in.Id, in.TaskName)

	return &cmpool.UpdateESSyncTaskResp{
		Success: true,
		Message: "更新任务成功",
	}, nil
}
