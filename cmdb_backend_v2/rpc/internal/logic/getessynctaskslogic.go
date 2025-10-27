package logic

import (
	"context"

	"cmdb-rpc/cmpool"
	"cmdb-rpc/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetEsSyncTasksLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGetEsSyncTasksLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetEsSyncTasksLogic {
	return &GetEsSyncTasksLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 获取ES数据同步任务配置列表
func (l *GetEsSyncTasksLogic) GetEsSyncTasks(in *cmpool.GetESSyncTasksReq) (*cmpool.GetESSyncTasksResp, error) {
	// 1. 调用 Model 方法查询任务列表
	tasks, err := l.svcCtx.ExternalSyncTaskConfigModel.FindAll(l.ctx, in.EnabledOnly)
	if err != nil {
		l.Logger.Errorf("查询任务列表失败: %v", err)
		return &cmpool.GetESSyncTasksResp{
			Success: false,
			Message: "查询任务列表失败",
		}, nil
	}

	// 2. 转换为 Proto 响应格式
	var taskList []*cmpool.ESSyncTask
	for _, task := range tasks {
		protoTask := &cmpool.ESSyncTask{
			Id:             int64(task.Id),
			TaskName:       task.TaskName,
			Description:    task.Description.String,
			EsEndpoint:     task.EsEndpoint.String,
			EsIndexPattern: task.EsIndexPattern.String,
			CronExpression: task.CronExpression,
			QueryTimeRange: task.QueryTimeRange,
			IsEnabled:      task.IsEnabled == 1,
			CreatedBy:      task.CreatedBy.String,
			CreatedAt:      task.CreatedAt.Format("2006-01-02 15:04:05"),
			UpdatedAt:      task.UpdatedAt.Format("2006-01-02 15:04:05"),
			// TODO: LastExecutionTime 和 NextExecutionTime 需要从其他地方获取
			// 可以从 es_sync_execution_log 表查询最后执行时间
			// NextExecutionTime 需要从调度器获取
		}
		taskList = append(taskList, protoTask)
	}

	l.Logger.Infof("查询任务列表成功: 共%d个任务", len(taskList))

	return &cmpool.GetESSyncTasksResp{
		Success: true,
		Message: "查询成功",
		Tasks:   taskList,
	}, nil
}
