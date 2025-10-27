package logic

import (
	"context"

	"cmdb-api/cmpool"
	"cmdb-api/internal/svc"
	"cmdb-api/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetExternalSyncExecutionDetailLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewGetExternalSyncExecutionDetailLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetExternalSyncExecutionDetailLogic {
	return &GetExternalSyncExecutionDetailLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetExternalSyncExecutionDetailLogic) GetExternalSyncExecutionDetail(executionId int64) (resp *types.GetExternalSyncExecutionDetailResponse, err error) {
	l.Logger.Infof("统一接口: 获取外部同步执行详情, ExecutionId=%d", executionId)

	// 查询执行详情（统一使用外部同步接口）
	l.Logger.Infof("调用 GetExternalSyncExecutionDetail RPC方法")

	rpcReq := &cmpool.GetExternalSyncExecutionDetailReq{
		ExecutionId: executionId,
	}

	rpcResp, err := l.svcCtx.CmpoolRpc.GetExternalSyncExecutionDetail(l.ctx, rpcReq)
	if err != nil {
		l.Logger.Errorf("调用执行详情查询RPC失败: %v", err)
		return &types.GetExternalSyncExecutionDetailResponse{
			Success: false,
			Message: "获取同步执行详情失败",
		}, nil
	}

	// 转换RPC响应为API响应
	var executionLog types.ExternalSyncExecutionLogInfo
	var executionDetails []types.ExternalSyncExecutionDetailInfo

	if rpcResp.Data != nil && rpcResp.Data.ExecutionLog != nil {
		log := rpcResp.Data.ExecutionLog
		executionLog = types.ExternalSyncExecutionLogInfo{
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
		}
	}

	if rpcResp.Data != nil && len(rpcResp.Data.Details) > 0 {
		executionDetails = make([]types.ExternalSyncExecutionDetailInfo, 0, len(rpcResp.Data.Details))
		for _, detail := range rpcResp.Data.Details {
			executionDetails = append(executionDetails, types.ExternalSyncExecutionDetailInfo{
				HostIp:         detail.HostIp,
				HostName:       detail.HostName,
				SyncStatus:     detail.SyncStatus,
				ErrorMessage:   detail.ErrorMessage,
				MaxCpu:         detail.MaxCpu,
				AvgCpu:         detail.AvgCpu,
				MaxMemory:      detail.MaxMemory,
				AvgMemory:      detail.AvgMemory,
				MaxDisk:        detail.MaxDisk,
				AvgDisk:        detail.AvgDisk,
				DataPointCount: detail.DataPointCount,
				CreatedAt:      detail.CreatedAt,
			})
		}
	}

	l.Logger.Infof("统一接口执行成功: ExecutionId=%d, DetailCount=%d", executionId, len(executionDetails))
	return &types.GetExternalSyncExecutionDetailResponse{
		Success:          rpcResp.Success,
		Message:          rpcResp.Message,
		ExecutionLog:     executionLog,
		ExecutionDetails: executionDetails,
	}, nil
}
