package logic

import (
	"context"

	"cmdb-api/cmpool"
	"cmdb-api/internal/svc"
	"cmdb-api/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetExternalSyncTasksLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewGetExternalSyncTasksLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetExternalSyncTasksLogic {
	return &GetExternalSyncTasksLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetExternalSyncTasksLogic) GetExternalSyncTasks(req *types.GetExternalSyncTasksRequest) (resp *types.GetExternalSyncTasksResponse, err error) {
	l.Logger.Infof("统一接口: 获取外部同步任务列表, DataSource=%s, EnabledOnly=%v", req.DataSource, req.EnabledOnly)

	// 如果指定了数据源过滤，验证参数
	var dataSourceFilter string
	if req.DataSource != "" {
		if err := validateDataSource(req.DataSource); err != nil {
			l.Logger.Errorf("参数验证失败: %v", err)
			return &types.GetExternalSyncTasksResponse{
				Success: false,
				Message: err.Error(),
			}, nil
		}
		dataSourceFilter = normalizeDataSource(req.DataSource)
	}

	// 调用统一的外部资源同步任务列表查询RPC（支持 ES 和 CMSys）
	l.Logger.Infof("数据源路由: GetExternalSyncTasks RPC方法, DataSourceFilter=%s", dataSourceFilter)

	rpcReq := &cmpool.GetExternalSyncTasksReq{
		DataSource:  dataSourceFilter,
		EnabledOnly: req.EnabledOnly,
	}

	rpcResp, err := l.svcCtx.CmpoolRpc.GetExternalSyncTasks(l.ctx, rpcReq)
	if err != nil {
		l.Logger.Errorf("调用统一外部同步任务列表查询RPC失败: %v", err)
		return &types.GetExternalSyncTasksResponse{
			Success: false,
			Message: "获取外部同步任务列表失败",
		}, nil
	}

	// 转换RPC响应为API响应
	tasks := make([]types.ExternalSyncTaskInfo, 0, len(rpcResp.Tasks))
	for _, task := range rpcResp.Tasks {
		tasks = append(tasks, types.ExternalSyncTaskInfo{
			Id:                task.Id,
			TaskName:          task.TaskName,
			Description:       task.Description,
			DataSource:        task.DataSource,
			EsEndpoint:        task.EsEndpoint,
			EsIndexPattern:    task.EsIndexPattern,
			CronExpression:    task.CronExpression,
			QueryTimeRange:    task.QueryTimeRange,
			CmsysQuery:        task.CmsysQuery,
			IsEnabled:         task.IsEnabled,
			CreatedBy:         task.CreatedBy,
			CreatedAt:         task.CreatedAt,
			UpdatedAt:         task.UpdatedAt,
			LastExecutionTime: task.LastExecutionTime,
			NextExecutionTime: task.NextExecutionTime,
		})
	}

	l.Logger.Infof("统一接口执行成功: 返回任务数=%d", len(tasks))
	return &types.GetExternalSyncTasksResponse{
		Success: rpcResp.Success,
		Message: rpcResp.Message,
		Tasks:   tasks,
	}, nil
}
