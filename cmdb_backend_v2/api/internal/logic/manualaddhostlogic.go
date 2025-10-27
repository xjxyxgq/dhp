package logic

import (
	"context"

	"cmdb-api/cmpool"
	"cmdb-api/internal/svc"
	"cmdb-api/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type ManualAddHostLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewManualAddHostLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ManualAddHostLogic {
	return &ManualAddHostLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *ManualAddHostLogic) ManualAddHost(req *types.ManualAddHostRequest) (resp *types.ManualAddHostResponse, err error) {
	// 构造硬件信息
	var hardwareInfo *cmpool.ManualHostHardwareInfo
	if req.HardwareInfo.DiskSize != 0 || req.HardwareInfo.Ram != 0 || req.HardwareInfo.Vcpus != 0 || req.HardwareInfo.HostType != "" {
		hardwareInfo = &cmpool.ManualHostHardwareInfo{
			DiskSize:        int32(req.HardwareInfo.DiskSize),
			Ram:             int32(req.HardwareInfo.Ram),
			Vcpus:           int32(req.HardwareInfo.Vcpus),
			HostType:        req.HardwareInfo.HostType,
			H3CId:           req.HardwareInfo.H3cId,
			H3CStatus:       req.HardwareInfo.H3cStatus,
			IfH3CSync:       req.HardwareInfo.IfH3cSync,
			H3CImgId:        req.HardwareInfo.H3cImgId,
			H3CHmName:       req.HardwareInfo.H3cHmName,
			LeafNumber:      req.HardwareInfo.LeafNumber,
			RackNumber:      req.HardwareInfo.RackNumber,
			RackHeight:      int32(req.HardwareInfo.RackHeight),
			RackStartNumber: int32(req.HardwareInfo.RackStartNumber),
			FromFactor:      int32(req.HardwareInfo.FromFactor),
			SerialNumber:    req.HardwareInfo.SerialNumber,
		}
	}

	// 构造应用信息
	var applications []*cmpool.HostApplicationInfo
	for _, app := range req.Applications {
		applications = append(applications, &cmpool.HostApplicationInfo{
			ServerType:     app.ServerType,
			ServerVersion:  app.ServerVersion,
			ServerSubtitle: app.ServerSubtitle,
			ClusterName:    app.ClusterName,
			ServerProtocol: app.ServerProtocol,
			ServerAddr:     app.ServerAddr,
			ServerPort:     int32(app.ServerPort),
			ServerRole:     app.ServerRole,
			ServerStatus:   app.ServerStatus,
			DepartmentName: app.DepartmentName,
		})
	}

	// 调用RPC服务
	rpcReq := &cmpool.ManualAddHostReq{
		HostIp:                req.HostIp,
		HostName:              req.HostName,
		HardwareInfo:          hardwareInfo,
		Applications:          applications,
		AutoFetchFromCmdb:     req.AutoFetchFromCmdb,
		AutoFetchFromClusters: req.AutoFetchFromClusters,
		IdcId:                 int64(req.IdcId),
	}

	rpcResp, err := l.svcCtx.CmpoolRpc.ManualAddHost(l.ctx, rpcReq)
	if err != nil {
		logx.Errorf("调用RPC服务失败: %v", err)
		return &types.ManualAddHostResponse{
			Success: false,
			Message: "服务调用失败",
		}, nil
	}

	// 转换自动获取结果
	var autoFetchResult types.AutoFetchResult
	if rpcResp.AutoFetchResult != nil {
		hardwareInfo := types.ManualHostHardwareInfo{}
		if rpcResp.AutoFetchResult.HardwareInfo != nil {
			hardwareInfo = types.ManualHostHardwareInfo{
				DiskSize:        int(rpcResp.AutoFetchResult.HardwareInfo.DiskSize),
				Ram:             int(rpcResp.AutoFetchResult.HardwareInfo.Ram),
				Vcpus:           int(rpcResp.AutoFetchResult.HardwareInfo.Vcpus),
				HostType:        rpcResp.AutoFetchResult.HardwareInfo.HostType,
				H3cId:           rpcResp.AutoFetchResult.HardwareInfo.H3CId,
				H3cStatus:       rpcResp.AutoFetchResult.HardwareInfo.H3CStatus,
				IfH3cSync:       rpcResp.AutoFetchResult.HardwareInfo.IfH3CSync,
				H3cImgId:        rpcResp.AutoFetchResult.HardwareInfo.H3CImgId,
				H3cHmName:       rpcResp.AutoFetchResult.HardwareInfo.H3CHmName,
				LeafNumber:      rpcResp.AutoFetchResult.HardwareInfo.LeafNumber,
				RackNumber:      rpcResp.AutoFetchResult.HardwareInfo.RackNumber,
				RackHeight:      int(rpcResp.AutoFetchResult.HardwareInfo.RackHeight),
				RackStartNumber: int(rpcResp.AutoFetchResult.HardwareInfo.RackStartNumber),
				FromFactor:      int(rpcResp.AutoFetchResult.HardwareInfo.FromFactor),
				SerialNumber:    rpcResp.AutoFetchResult.HardwareInfo.SerialNumber,
			}
		}

		var applications []types.HostApplicationInfo
		for _, app := range rpcResp.AutoFetchResult.Applications {
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

		autoFetchResult = types.AutoFetchResult{
			Success:      rpcResp.AutoFetchResult.Success,
			Message:      rpcResp.AutoFetchResult.Message,
			HardwareInfo: hardwareInfo,
			Applications: applications,
		}
	}

	return &types.ManualAddHostResponse{
		Success:         rpcResp.Success,
		Message:         rpcResp.Message,
		AutoFetchResult: autoFetchResult,
	}, nil
}
