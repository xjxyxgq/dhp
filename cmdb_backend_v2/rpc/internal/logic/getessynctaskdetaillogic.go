package logic

import (
	"context"

	"cmdb-rpc/cmpool"
	"cmdb-rpc/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetEsSyncTaskDetailLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGetEsSyncTaskDetailLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetEsSyncTaskDetailLogic {
	return &GetEsSyncTaskDetailLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 获取ES数据同步任务配置详情
func (l *GetEsSyncTaskDetailLogic) GetEsSyncTaskDetail(in *cmpool.GetESSyncTaskDetailReq) (*cmpool.GetESSyncTaskDetailResp, error) {
	// 1. 验证参数
	if in.Id == 0 {
		return &cmpool.GetESSyncTaskDetailResp{
			Success: false,
			Message: "任务ID不能为空",
		}, nil
	}

	// 2. 调用 Model 方法查询任务详情
	task, err := l.svcCtx.ExternalSyncTaskConfigModel.FindOne(l.ctx, uint64(in.Id))
	if err != nil {
		l.Logger.Errorf("查询任务详情失败: %v", err)
		return &cmpool.GetESSyncTaskDetailResp{
			Success: false,
			Message: "任务不存在或查询失败",
		}, nil
	}

	// 3. 转换为 Proto 响应格式
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
	}

	l.Logger.Infof("查询任务详情成功: ID=%d, Name=%s", task.Id, task.TaskName)

	return &cmpool.GetESSyncTaskDetailResp{
		Success: true,
		Message: "查询成功",
		Task:    protoTask,
	}, nil
}
