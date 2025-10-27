package logic

import (
	"context"

	"cmdb-api/cmpool"
	"cmdb-api/internal/svc"
	"cmdb-api/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type AutoFetchHostInfoLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewAutoFetchHostInfoLogic(ctx context.Context, svcCtx *svc.ServiceContext) *AutoFetchHostInfoLogic {
	return &AutoFetchHostInfoLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *AutoFetchHostInfoLogic) AutoFetchHostInfo(req *types.AutoFetchHostInfoRequest) (resp *types.AutoFetchHostInfoResponse, err error) {
	// 调用RPC服务
	rpcReq := &cmpool.AutoFetchHostInfoReq{
		HostIp:            req.HostIp,
		FetchFromCmdb:     req.FetchFromCmdb,
		FetchFromClusters: req.FetchFromClusters,
	}

	rpcResp, err := l.svcCtx.CmpoolRpc.AutoFetchHostInfo(l.ctx, rpcReq)
	if err != nil {
		logx.Errorf("调用RPC服务失败: %v", err)
		return &types.AutoFetchHostInfoResponse{
			Success: false,
			Message: "服务调用失败",
		}, nil
	}

	// 转换硬件信息
	hardwareInfo := types.ManualHostHardwareInfo{}
	if rpcResp.HardwareInfo != nil {
		hardwareInfo = types.ManualHostHardwareInfo{
			DiskSize:        int(rpcResp.HardwareInfo.DiskSize),
			Ram:             int(rpcResp.HardwareInfo.Ram),
			Vcpus:           int(rpcResp.HardwareInfo.Vcpus),
			HostType:        rpcResp.HardwareInfo.HostType,
			H3cId:           rpcResp.HardwareInfo.H3CId,
			H3cStatus:       rpcResp.HardwareInfo.H3CStatus,
			IfH3cSync:       rpcResp.HardwareInfo.IfH3CSync,
			H3cImgId:        rpcResp.HardwareInfo.H3CImgId,
			H3cHmName:       rpcResp.HardwareInfo.H3CHmName,
			LeafNumber:      rpcResp.HardwareInfo.LeafNumber,
			RackNumber:      rpcResp.HardwareInfo.RackNumber,
			RackHeight:      int(rpcResp.HardwareInfo.RackHeight),
			RackStartNumber: int(rpcResp.HardwareInfo.RackStartNumber),
			FromFactor:      int(rpcResp.HardwareInfo.FromFactor),
			SerialNumber:    rpcResp.HardwareInfo.SerialNumber,
		}
	}

	// 转换应用信息
	var applications []types.HostApplicationInfo
	for _, app := range rpcResp.Applications {
		applications = append(applications, types.HostApplicationInfo{
			ServerType:     app.ServerType,
			ServerVersion:  app.ServerVersion,
			ServerSubtitle: app.ServerSubtitle,
			ClusterName:    app.ClusterName,
			ServerProtocol: app.ServerProtocol,
			ServerAddr:     app.ServerAddr,
			ServerPort:     int(app.ServerPort),
			ServerRole:     app.ServerRole,
			ServerStatus:   app.ServerStatus,
			DepartmentName: app.DepartmentName,
		})
	}

	// 转换IDC信息
	var idcInfo types.IdcInfo
	if rpcResp.IdcInfo != nil {
		idcInfo = types.IdcInfo{
			ID:             int(rpcResp.IdcInfo.Id),
			IdcName:        rpcResp.IdcInfo.IdcName,
			IdcCode:        rpcResp.IdcInfo.IdcCode,
			IdcLocation:    rpcResp.IdcInfo.IdcLocation,
			IdcDescription: rpcResp.IdcInfo.IdcDescription,
		}
	}

	return &types.AutoFetchHostInfoResponse{
		Success:      rpcResp.Success,
		Message:      rpcResp.Message,
		HardwareInfo: hardwareInfo,
		Applications: applications,
		IdcInfo:      idcInfo,
	}, nil
}
