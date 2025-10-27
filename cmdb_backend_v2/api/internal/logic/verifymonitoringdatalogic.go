package logic

import (
	"context"

	"cmdb-api/cmpool"
	"cmdb-api/internal/svc"
	"cmdb-api/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type VerifyMonitoringDataLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewVerifyMonitoringDataLogic(ctx context.Context, svcCtx *svc.ServiceContext) *VerifyMonitoringDataLogic {
	return &VerifyMonitoringDataLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *VerifyMonitoringDataLogic) VerifyMonitoringData(req *types.MonitoringVerificationRequest) (resp *types.MonitoringVerificationResponse, err error) {
	// 调用RPC服务
	rpcResp, err := l.svcCtx.CmpoolRpc.VerifyMonitoringData(l.ctx, &cmpool.VerifyMonitoringDataReq{
		StartTime: req.StartTime,
		EndTime:   req.EndTime,
	})
	if err != nil {
		return nil, err
	}

	// 转换RPC响应为API响应
	var hostsWithoutMonitoringList []types.HostWithoutMonitoring
	for _, host := range rpcResp.HostsWithoutMonitoringList {
		hostsWithoutMonitoringList = append(hostsWithoutMonitoringList, types.HostWithoutMonitoring{
			HostIp:     host.HostIp,
			HostName:   host.HostName,
			PoolName:   host.PoolName,
			CreateTime: host.CreateTime,
		})
	}

	return &types.MonitoringVerificationResponse{
		Success: rpcResp.Success,
		Message: rpcResp.Message,
		Data: types.MonitoringVerificationData{
			TotalHosts:                 int(rpcResp.TotalHosts),
			HostsWithMonitoring:        int(rpcResp.HostsWithMonitoring),
			HostsWithoutMonitoring:     int(rpcResp.HostsWithoutMonitoring),
			MonitoringCoverage:         rpcResp.MonitoringCoverage,
			HostsWithoutMonitoringList: hostsWithoutMonitoringList,
		},
	}, nil
}
