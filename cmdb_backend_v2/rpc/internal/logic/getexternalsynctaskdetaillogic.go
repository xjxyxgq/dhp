package logic

import (
	"context"
	"database/sql"

	"cmdb-rpc/cmpool"
	"cmdb-rpc/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetExternalSyncTaskDetailLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGetExternalSyncTaskDetailLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetExternalSyncTaskDetailLogic {
	return &GetExternalSyncTaskDetailLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// GetExternalSyncTaskDetail 获取外部同步任务配置详情
func (l *GetExternalSyncTaskDetailLogic) GetExternalSyncTaskDetail(in *cmpool.GetExternalSyncTaskDetailReq) (*cmpool.GetExternalSyncTaskDetailResp, error) {
	// 1. 验证任务ID
	if in.Id <= 0 {
		return &cmpool.GetExternalSyncTaskDetailResp{
			Success: false,
			Message: "任务ID不能为空",
		}, nil
	}

	// 2. 查询任务详情
	task, err := l.svcCtx.ExternalSyncTaskConfigModel.FindOne(l.ctx, uint64(in.Id))
	if err != nil {
		if err == sql.ErrNoRows {
			return &cmpool.GetExternalSyncTaskDetailResp{
				Success: false,
				Message: "任务不存在",
			}, nil
		}
		l.Logger.Errorf("查询任务详情失败: %v", err)
		return &cmpool.GetExternalSyncTaskDetailResp{
			Success: false,
			Message: "查询任务详情失败",
		}, nil
	}

	// 3. 转换为响应格式
	taskInfo := &cmpool.ExternalSyncTask{
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
	}

	l.Logger.Infof("查询任务详情成功: TaskID=%d, Name=%s, DataSource=%s",
		task.Id, task.TaskName, task.DataSource)

	return &cmpool.GetExternalSyncTaskDetailResp{
		Success: true,
		Message: "查询成功",
		Task:    taskInfo,
	}, nil
}
