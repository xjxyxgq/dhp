package logic

import (
	"context"

	"cmdb-api/cmpool"
	"cmdb-api/internal/svc"
	"cmdb-api/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetExternalSyncExecutionLogsLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewGetExternalSyncExecutionLogsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetExternalSyncExecutionLogsLogic {
	return &GetExternalSyncExecutionLogsLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetExternalSyncExecutionLogsLogic) GetExternalSyncExecutionLogs(req *types.GetExternalSyncExecutionLogsRequest) (resp *types.GetExternalSyncExecutionLogsResponse, err error) {
	l.Logger.Infof("统一接口: 获取外部同步执行日志列表, TaskId=%d, DataSource=%s, Limit=%d",
		req.TaskId, req.DataSource, req.Limit)

	// 如果指定了数据源过滤，验证参数
	normalizedSource := ""
	if req.DataSource != "" {
		if err := validateDataSource(req.DataSource); err != nil {
			l.Logger.Errorf("参数验证失败: %v", err)
			return &types.GetExternalSyncExecutionLogsResponse{
				Success: false,
				Message: err.Error(),
			}, nil
		}
		normalizedSource = normalizeDataSource(req.DataSource)
		l.Logger.Infof("数据源参数规范化: %s -> %s", req.DataSource, normalizedSource)
	}

	// 调用统一的执行日志查询RPC
	l.Logger.Infof("调用 GetExternalSyncExecutionLogs RPC方法，DataSource=%s", normalizedSource)

	rpcReq := &cmpool.GetExternalSyncExecutionLogsReq{
		TaskId:     req.TaskId,
		Limit:      req.Limit,
		DataSource: normalizedSource, // 传递规范化后的数据源参数
	}

	rpcResp, err := l.svcCtx.CmpoolRpc.GetExternalSyncExecutionLogs(l.ctx, rpcReq)
	if err != nil {
		l.Logger.Errorf("调用执行日志查询RPC失败: %v", err)
		return &types.GetExternalSyncExecutionLogsResponse{
			Success: false,
			Message: "获取同步执行日志失败",
		}, nil
	}

	// 转换RPC响应为API响应
	logs := make([]types.ExternalSyncExecutionLogInfo, 0, len(rpcResp.ExecutionLogs))
	for _, log := range rpcResp.ExecutionLogs {
		logs = append(logs, types.ExternalSyncExecutionLogInfo{
			Id:              log.Id,
			TaskId:          log.TaskId,
			TaskName:        log.TaskName,
			ExecutionTime:   log.ExecutionTime,
			ExecutionStatus: log.ExecutionStatus,
			TotalHosts:      log.TotalHosts,
			SuccessCount:    log.SuccessCount,
			FailedCount:     log.FailedCount,
			NotInPoolCount:  log.NotInPoolCount,
			ErrorMessage:    log.ErrorMessage,
			DurationMs:      log.DurationMs,
			QueryTimeRange:  log.QueryTimeRange,
			CreatedAt:       log.CreatedAt,
		})
	}

	l.Logger.Infof("统一接口执行成功: 返回日志数=%d", len(logs))
	return &types.GetExternalSyncExecutionLogsResponse{
		Success:       rpcResp.Success,
		Message:       rpcResp.Message,
		ExecutionLogs: logs,
	}, nil
}
