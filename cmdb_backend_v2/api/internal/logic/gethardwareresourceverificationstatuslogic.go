package logic

import (
	"context"

	"cmdb-api/cmpool"
	"cmdb-api/internal/svc"
	"cmdb-api/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetHardwareResourceVerificationStatusLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewGetHardwareResourceVerificationStatusLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetHardwareResourceVerificationStatusLogic {
	return &GetHardwareResourceVerificationStatusLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetHardwareResourceVerificationStatusLogic) GetHardwareResourceVerificationStatus(req *types.HardwareResourceVerificationStatusRequest) (resp *types.HardwareResourceVerificationStatusResponse, err error) {
	// 调用RPC服务
	rpcResp, err := l.svcCtx.CmpoolRpc.GetHardwareResourceVerificationStatus(l.ctx, &cmpool.GetHardwareResourceVerificationStatusReq{
		HostIpList:   req.HostIpList,
		ResourceType: req.ResourceType,
	})
	
	if err != nil {
		return &types.HardwareResourceVerificationStatusResponse{
			Success: false,
			Message: "调用RPC服务失败: " + err.Error(),
		}, nil
	}
	
	// 转换响应数据
	var records []types.HardwareResourceVerificationStatus
	for _, record := range rpcResp.VerificationRecords {
		status := types.HardwareResourceVerificationStatus{
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
			ResultSummary:   record.ResultSummary,
			CreateTime:      record.CreateTime,
			SSHError:        record.SSHError,
		}
		records = append(records, status)
	}
	
	return &types.HardwareResourceVerificationStatusResponse{
		Success:             rpcResp.Success,
		Message:             rpcResp.Message,
		VerificationRecords: records,
	}, nil
}
