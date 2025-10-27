package logic

import (
	"context"

	"cmdb-rpc/cmpool"
	"cmdb-rpc/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetExternalSyncTasksLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGetExternalSyncTasksLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetExternalSyncTasksLogic {
	return &GetExternalSyncTasksLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// GetExternalSyncTasks 获取外部同步任务配置列表
func (l *GetExternalSyncTasksLogic) GetExternalSyncTasks(in *cmpool.GetExternalSyncTasksReq) (*cmpool.GetExternalSyncTasksResp, error) {
	var tasks []*cmpool.ExternalSyncTask

	// 根据筛选条件查询任务列表
	if in.DataSource != "" {
		// 按数据源过滤
		enabledOnly := in.EnabledOnly
		taskList, err := l.svcCtx.ExternalSyncTaskConfigModel.FindByDataSource(l.ctx, in.DataSource, enabledOnly)
		if err != nil {
			l.Logger.Errorf("查询任务列表失败 (data_source=%s): %v", in.DataSource, err)
			return &cmpool.GetExternalSyncTasksResp{
				Success: false,
				Message: "查询任务列表失败",
			}, nil
		}

		// 转换为响应格式
		tasks = make([]*cmpool.ExternalSyncTask, 0, len(taskList))
		for _, task := range taskList {
			tasks = append(tasks, &cmpool.ExternalSyncTask{
				Id:             int64(task.Id),
				TaskName:       task.TaskName,
				Description:    task.Description.String,
				DataSource:     task.DataSource,
				CronExpression: task.CronExpression,
				IsEnabled:      task.IsEnabled == 1,
				EsEndpoint:     task.EsEndpoint.String,
				EsIndexPattern: task.EsIndexPattern.String,
				QueryTimeRange: task.QueryTimeRange,
				CmsysQuery:     task.CmsysQuery.String,
				CreatedBy:      task.CreatedBy.String,
				CreatedAt:      task.CreatedAt.Format("2006-01-02 15:04:05"),
				UpdatedAt:      task.UpdatedAt.Format("2006-01-02 15:04:05"),
			})
		}

	} else {
		// 查询所有任务
		enabledOnly := in.EnabledOnly
		taskList, err := l.svcCtx.ExternalSyncTaskConfigModel.FindAll(l.ctx, enabledOnly)
		if err != nil {
			l.Logger.Errorf("查询任务列表失败: %v", err)
			return &cmpool.GetExternalSyncTasksResp{
				Success: false,
				Message: "查询任务列表失败",
			}, nil
		}

		// 转换为响应格式
		tasks = make([]*cmpool.ExternalSyncTask, 0, len(taskList))
		for _, task := range taskList {
			tasks = append(tasks, &cmpool.ExternalSyncTask{
				Id:             int64(task.Id),
				TaskName:       task.TaskName,
				Description:    task.Description.String,
				DataSource:     task.DataSource,
				CronExpression: task.CronExpression,
				IsEnabled:      task.IsEnabled == 1,
				EsEndpoint:     task.EsEndpoint.String,
				EsIndexPattern: task.EsIndexPattern.String,
				QueryTimeRange: task.QueryTimeRange,
				CmsysQuery:     task.CmsysQuery.String,
				CreatedBy:      task.CreatedBy.String,
				CreatedAt:      task.CreatedAt.Format("2006-01-02 15:04:05"),
				UpdatedAt:      task.UpdatedAt.Format("2006-01-02 15:04:05"),
			})
		}
	}

	l.Logger.Infof("查询任务列表成功: 数据源=%s, 仅启用=%v, 结果数=%d",
		in.DataSource, in.EnabledOnly, len(tasks))

	return &cmpool.GetExternalSyncTasksResp{
		Success: true,
		Message: "查询成功",
		Tasks:   tasks,
	}, nil
}
