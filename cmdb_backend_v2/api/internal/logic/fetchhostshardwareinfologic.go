package logic

import (
	"context"

	"cmdb-api/cmpool"
	"cmdb-api/internal/svc"
	"cmdb-api/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type FetchHostsHardwareInfoLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewFetchHostsHardwareInfoLogic(ctx context.Context, svcCtx *svc.ServiceContext) *FetchHostsHardwareInfoLogic {
	return &FetchHostsHardwareInfoLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *FetchHostsHardwareInfoLogic) FetchHostsHardwareInfo(req *types.FetchHostsHardwareInfoRequest) (resp *types.FetchHostsHardwareInfoResponse, err error) {
	// 调用RPC服务
	rpcReq := &cmpool.FetchHostsHardwareInfoReq{
		HostIpList: req.HostIpList,
	}

	rpcResp, err := l.svcCtx.CmpoolRpc.FetchHostsHardwareInfo(l.ctx, rpcReq)
	if err != nil {
		logx.Errorf("调用RPC服务失败: %v", err)
		return &types.FetchHostsHardwareInfoResponse{
			Success: false,
			Message: "调用RPC服务失败: " + err.Error(),
		}, nil
	}

	// 转换响应
	var hardwareInfoList []types.HostHardwareInfo
	for _, info := range rpcResp.HardwareInfoList {
		hardwareInfoList = append(hardwareInfoList, types.HostHardwareInfo{
			HostIp:   info.HostIp,
			HostName: info.HostName,
			Disk:     info.Disk,
			Ram:      info.Ram,
			Vcpus:    info.Vcpus,
			Message:  info.Message,
			Success:  info.Success,
		})
	}

	return &types.FetchHostsHardwareInfoResponse{
		Success:          rpcResp.Success,
		Message:          rpcResp.Message,
		TotalHosts:       int(rpcResp.TotalHosts),
		UpdatedHosts:     int(rpcResp.UpdatedHosts),
		FailedHosts:      int(rpcResp.FailedHosts),
		HardwareInfoList: hardwareInfoList,
	}, nil
}
