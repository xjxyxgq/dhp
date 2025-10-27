package logic

import (
	"context"

	"cmdb-api/cmpool"
	"cmdb-api/internal/svc"
	"cmdb-api/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetExternalSyncTaskDetailLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewGetExternalSyncTaskDetailLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetExternalSyncTaskDetailLogic {
	return &GetExternalSyncTaskDetailLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetExternalSyncTaskDetailLogic) GetExternalSyncTaskDetail(taskId int64) (resp *types.GetExternalSyncTaskDetailResponse, err error) {
	l.Logger.Infof("统一接口: 获取外部同步任务详情, TaskId=%d", taskId)

	// 调用统一的外部资源同步任务详情查询RPC（支持 ES 和 CMSys）
	l.Logger.Infof("数据源路由: GetExternalSyncTaskDetail RPC方法")

	rpcReq := &cmpool.GetExternalSyncTaskDetailReq{
		Id: taskId,
	}

	rpcResp, err := l.svcCtx.CmpoolRpc.GetExternalSyncTaskDetail(l.ctx, rpcReq)
	if err != nil {
		l.Logger.Errorf("调用统一外部同步任务详情查询RPC失败: %v", err)
		return &types.GetExternalSyncTaskDetailResponse{
			Success: false,
			Message: "获取外部同步任务详情失败",
		}, nil
	}

	// 转换RPC响应为API响应
	var task types.ExternalSyncTaskInfo
	if rpcResp.Task != nil {
		task = types.ExternalSyncTaskInfo{
			Id:                rpcResp.Task.Id,
			TaskName:          rpcResp.Task.TaskName,
			Description:       rpcResp.Task.Description,
			DataSource:        rpcResp.Task.DataSource,
			EsEndpoint:        rpcResp.Task.EsEndpoint,
			EsIndexPattern:    rpcResp.Task.EsIndexPattern,
			CronExpression:    rpcResp.Task.CronExpression,
			QueryTimeRange:    rpcResp.Task.QueryTimeRange,
			CmsysQuery:        rpcResp.Task.CmsysQuery,
			IsEnabled:         rpcResp.Task.IsEnabled,
			CreatedBy:         rpcResp.Task.CreatedBy,
			CreatedAt:         rpcResp.Task.CreatedAt,
			UpdatedAt:         rpcResp.Task.UpdatedAt,
			LastExecutionTime: rpcResp.Task.LastExecutionTime,
			NextExecutionTime: rpcResp.Task.NextExecutionTime,
		}
	}

	l.Logger.Infof("统一接口执行成功: TaskId=%d, TaskName=%s, DataSource=%s", taskId, task.TaskName, task.DataSource)
	return &types.GetExternalSyncTaskDetailResponse{
		Success: rpcResp.Success,
		Message: rpcResp.Message,
		Task:    task,
	}, nil
}
