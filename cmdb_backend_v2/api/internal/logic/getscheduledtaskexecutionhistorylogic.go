package logic

import (
	"context"

	"cmdb-api/cmpool"
	"cmdb-api/internal/svc"
	"cmdb-api/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetScheduledTaskExecutionHistoryLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewGetScheduledTaskExecutionHistoryLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetScheduledTaskExecutionHistoryLogic {
	return &GetScheduledTaskExecutionHistoryLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetScheduledTaskExecutionHistoryLogic) GetScheduledTaskExecutionHistory(req *types.GetScheduledTaskExecutionHistoryRequest) (resp *types.GetScheduledTaskExecutionHistoryResponse, err error) {
	// 调用RPC服务获取执行历史
	rpcReq := &cmpool.GetScheduledTaskExecutionHistoryReq{
		ScheduledTaskId: int64(req.ScheduledTaskId),
		Limit:           int32(req.Limit),
	}

	rpcResp, err := l.svcCtx.CmpoolRpc.GetScheduledTaskExecutionHistory(l.ctx, rpcReq)
	if err != nil {
		l.Errorf("调用RPC获取执行历史失败: %v", err)
		return &types.GetScheduledTaskExecutionHistoryResponse{
			Success:        false,
			Message:        "RPC调用失败",
			HistoryRecords: []types.ScheduledTaskExecutionHistoryInfo{},
		}, nil
	}

	// 转换RPC响应为API响应格式
	var historyRecords []types.ScheduledTaskExecutionHistoryInfo
	for _, rpcHistory := range rpcResp.HistoryRecords {
		// 转换主机详情
		var hostDetails []types.ScheduledTaskExecutionDetail
		for _, rpcDetail := range rpcHistory.HostDetails {
			hostDetail := types.ScheduledTaskExecutionDetail{
				HostIp:          rpcDetail.HostIp,
				ResourceType:    rpcDetail.ResourceType,
				TargetPercent:   int(rpcDetail.TargetPercent),
				Duration:        int(rpcDetail.Duration),
				ExecutionStatus: rpcDetail.ExecutionStatus,
				StartTime:       rpcDetail.StartTime,
				EndTime:         rpcDetail.EndTime,
				ExitCode:        int(rpcDetail.ExitCode),
				StdoutLog:       rpcDetail.StdoutLog,
				StderrLog:       rpcDetail.StderrLog,
				ResultSummary:   rpcDetail.ResultSummary,
				SSHError:        rpcDetail.SSHError,
				CreateTime:      rpcDetail.CreateTime,
			}
			hostDetails = append(hostDetails, hostDetail)
		}

		history := types.ScheduledTaskExecutionHistoryInfo{
			Id:              int(rpcHistory.Id),
			ScheduledTaskId: int(rpcHistory.ScheduledTaskId),
			ExecutionTaskId: rpcHistory.ExecutionTaskId,
			ExecutionTime:   rpcHistory.ExecutionTime,
			ExecutionStatus: rpcHistory.ExecutionStatus,
			TotalHosts:      int(rpcHistory.TotalHosts),
			SuccessHosts:    int(rpcHistory.SuccessHosts),
			FailedHosts:     int(rpcHistory.FailedHosts),
			ErrorMessage:    rpcHistory.ErrorMessage,
		}
		historyRecords = append(historyRecords, history)
	}

	return &types.GetScheduledTaskExecutionHistoryResponse{
		Success:        rpcResp.Success,
		Message:        rpcResp.Message,
		HistoryRecords: historyRecords,
	}, nil
}
