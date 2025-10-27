package logic

import (
	"context"

	"cmdb-rpc/cmpool"
	"cmdb-rpc/internal/model"
	"cmdb-rpc/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetExternalSyncExecutionLogsLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGetExternalSyncExecutionLogsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetExternalSyncExecutionLogsLogic {
	return &GetExternalSyncExecutionLogsLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 获取外部数据同步执行记录列表
func (l *GetExternalSyncExecutionLogsLogic) GetExternalSyncExecutionLogs(in *cmpool.GetExternalSyncExecutionLogsReq) (*cmpool.GetExternalSyncExecutionLogsResp, error) {
	// 1. 设置默认 Limit
	limit := in.Limit
	if limit <= 0 {
		limit = 50
	}

	// 2. 根据 DataSource 和 TaskId 决定查询方式
	var executionLogs []*model.ExternalSyncExecutionLog
	var err error

	// 规范化 DataSource 参数（如果提供）
	dataSource := ""
	if in.DataSource != "" {
		// 支持的数据源: elasticsearch, es, cmsys
		switch in.DataSource {
		case "elasticsearch", "es":
			dataSource = "elasticsearch"
		case "cmsys":
			dataSource = "cmsys"
		default:
			l.Logger.Errorf("不支持的数据源类型: %s", in.DataSource)
			return &cmpool.GetExternalSyncExecutionLogsResp{
				Success: false,
				Message: "不支持的数据源类型，仅支持: elasticsearch, es, cmsys",
			}, nil
		}
	}

	// 3. 根据参数组合选择查询方法
	if in.TaskId > 0 && dataSource != "" {
		// 同时指定了 TaskId 和 DataSource
		l.Logger.Infof("查询指定任务(%d)和数据源(%s)的执行记录", in.TaskId, dataSource)
		executionLogs, err = l.svcCtx.ExternalSyncExecutionLogModel.FindByTaskIdAndDataSource(l.ctx, uint64(in.TaskId), dataSource, limit)
		if err != nil {
			l.Logger.Errorf("FindByTaskIdAndDataSource 查询失败 - TaskId=%d, DataSource=%s, Error=%v", in.TaskId, dataSource, err)
		}
	} else if in.TaskId > 0 {
		// 只指定了 TaskId
		l.Logger.Infof("查询指定任务(%d)的执行记录", in.TaskId)
		executionLogs, err = l.svcCtx.ExternalSyncExecutionLogModel.FindByTaskId(l.ctx, uint64(in.TaskId), limit)
		if err != nil {
			l.Logger.Errorf("FindByTaskId 查询失败 - TaskId=%d, Error=%v", in.TaskId, err)
		}
	} else if dataSource != "" {
		// 只指定了 DataSource
		l.Logger.Infof("查询指定数据源(%s)的执行记录", dataSource)
		executionLogs, err = l.svcCtx.ExternalSyncExecutionLogModel.FindByDataSource(l.ctx, dataSource, limit)
		if err != nil {
			l.Logger.Errorf("FindByDataSource 查询失败 - DataSource=%s, Error=%v", dataSource, err)
		}
	} else {
		// 都没有指定，查询所有
		l.Logger.Infof("查询所有任务的最新执行记录")
		executionLogs, err = l.svcCtx.ExternalSyncExecutionLogModel.FindLatest(l.ctx, limit)
		if err != nil {
			l.Logger.Errorf("FindLatest 查询失败 - Error=%v", err)
		}
	}

	if err != nil {
		l.Logger.Errorf("查询执行记录失败: %v", err)
		return &cmpool.GetExternalSyncExecutionLogsResp{
			Success: false,
			Message: "查询执行记录失败",
		}, nil
	}

	// 4. 转换为 Proto 响应格式
	var logList []*cmpool.ExternalSyncExecutionLog
	for _, log := range executionLogs {
		protoLog := &cmpool.ExternalSyncExecutionLog{
			Id:              int64(log.Id),
			TaskId:          int64(log.TaskId),
			TaskName:        log.TaskName,
			ExecutionTime:   log.ExecutionTime.Format("2006-01-02 15:04:05"),
			ExecutionStatus: log.ExecutionStatus,
			TotalHosts:      int32(log.TotalHosts),
			SuccessCount:    int32(log.SuccessCount),
			FailedCount:     int32(log.FailedCount),
			NotInPoolCount:  int32(log.NotInPoolCount),
			ErrorMessage:    log.ErrorMessage.String,
			DurationMs:      int32(log.DurationMs.Int64),
			QueryTimeRange:  log.QueryTimeRange.String,
			CreatedAt:       log.CreatedAt.Format("2006-01-02 15:04:05"),
		}
		logList = append(logList, protoLog)
	}

	l.Logger.Infof("查询执行记录成功: 共%d条记录", len(logList))

	return &cmpool.GetExternalSyncExecutionLogsResp{
		Success:       true,
		Message:       "查询成功",
		ExecutionLogs: logList,
	}, nil
}
