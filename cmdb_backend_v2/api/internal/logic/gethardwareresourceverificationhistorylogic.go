package logic

import (
	"context"

	"cmdb-api/cmpool"
	"cmdb-api/internal/svc"
	"cmdb-api/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetHardwareResourceVerificationHistoryLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewGetHardwareResourceVerificationHistoryLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetHardwareResourceVerificationHistoryLogic {
	return &GetHardwareResourceVerificationHistoryLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetHardwareResourceVerificationHistoryLogic) GetHardwareResourceVerificationHistory(req *types.HardwareResourceVerificationHistoryRequest) (resp *types.HardwareResourceVerificationHistoryResponse, err error) {
	// 调用RPC服务
	rpcResp, err := l.svcCtx.CmpoolRpc.GetHardwareResourceVerificationHistory(l.ctx, &cmpool.GetHardwareResourceVerificationHistoryReq{
		HostIp:       req.HostIp,
		ResourceType: req.ResourceType,
		Limit:        int32(req.Limit),
	})
	
	if err != nil {
		return &types.HardwareResourceVerificationHistoryResponse{
			Success: false,
			Message: "调用RPC服务失败: " + err.Error(),
		}, nil
	}
	
	// 转换响应数据
	var records []types.HardwareResourceVerificationHistory
	for _, record := range rpcResp.HistoryRecords {
		history := types.HardwareResourceVerificationHistory{
			Id:              int(record.Id),
			TaskId:          record.TaskId,
			HostIp:          record.HostIp,
			ResourceType:    record.ResourceType,
			TargetPercent:   int(record.TargetPercent),
			Duration:        int(record.Duration),
			ExecutionStatus: record.ExecutionStatus,
			StartTime:       record.StartTime,
			EndTime:         record.EndTime,
			ExitCode:        int(record.ExitCode),
			StdoutLog:       record.StdoutLog,
			StderrLog:       record.StderrLog,
			ResultSummary:   record.ResultSummary,
			CreateTime:      record.CreateTime,
			SSHError:        record.SSHError,
		}
		records = append(records, history)
	}
	
	return &types.HardwareResourceVerificationHistoryResponse{
		Success:        rpcResp.Success,
		Message:        rpcResp.Message,
		HistoryRecords: records,
	}, nil
}
