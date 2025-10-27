package logic

import (
	"context"

	"cmdb-api/cmpool"
	"cmdb-api/internal/svc"
	"cmdb-api/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type HardwareResourceVerificationLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewHardwareResourceVerificationLogic(ctx context.Context, svcCtx *svc.ServiceContext) *HardwareResourceVerificationLogic {
	return &HardwareResourceVerificationLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *HardwareResourceVerificationLogic) HardwareResourceVerification(req *types.HardwareResourceVerificationRequest) (resp *types.HardwareResourceVerificationResponse, err error) {
	// 调用RPC服务
	rpcResp, err := l.svcCtx.CmpoolRpc.HardwareResourceVerification(l.ctx, &cmpool.HardwareResourceVerificationReq{
		HostIpList:      req.HostIpList,
		ResourceType:    req.ResourceType,
		TargetPercent:   int32(req.TargetPercent),
		Duration:        int32(req.Duration),
		ForceExecution:  req.ForceExecution,
		ScriptParams:    req.ScriptParams,
	})
	
	if err != nil {
		return &types.HardwareResourceVerificationResponse{
			Success: false,
			Message: "调用RPC服务失败: " + err.Error(),
		}, nil
	}
	
	return &types.HardwareResourceVerificationResponse{
		Success: rpcResp.Success,
		Message: rpcResp.Message,
		TaskId:  rpcResp.TaskId,
	}, nil
}
