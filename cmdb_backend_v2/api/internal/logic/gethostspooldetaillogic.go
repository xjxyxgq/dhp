package logic

import (
	"context"

	"cmdb-api/cmpool"
	"cmdb-api/internal/svc"
	"cmdb-api/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetHostsPoolDetailLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewGetHostsPoolDetailLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetHostsPoolDetailLogic {
	return &GetHostsPoolDetailLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetHostsPoolDetailLogic) GetHostsPoolDetail() (resp *types.HostPoolListResponse, err error) {
	l.Logger.Info("开始调用RPC获取主机池详情")

	// 调用RPC服务获取主机池详情
	rpcReq := &cmpool.GetHostsPoolDetailReq{
		IpList: []string{}, // 空列表表示获取所有主机
	}

	rpcResp, err := l.svcCtx.CmpoolRpc.GetHostsPoolDetail(l.ctx, rpcReq)
	if err != nil {
		l.Logger.Errorf("调用RPC获取主机池详情失败: %v", err)
		return nil, err
	}

	if !rpcResp.Success {
		l.Logger.Errorf("RPC返回失败: %s", rpcResp.Message)
		return &types.HostPoolListResponse{
			List: []types.HostPool{},
		}, nil
	}

	// 转换RPC响应为API响应格式
	hostPools := make([]types.HostPool, 0, len(rpcResp.HostsPoolDetail))
	for _, rpcHost := range rpcResp.HostsPoolDetail {
		// 转换应用信息
		applications := make([]types.HostApplication, 0, len(rpcHost.AppList))
		for _, rpcApp := range rpcHost.AppList {
			app := types.HostApplication{
				ID:             int(rpcApp.Aid), // 设置应用ID
				PoolID:         int(rpcHost.Id), // 设置主机池ID
				ServerType:     rpcApp.ServerType,
				ServerVersion:  rpcApp.ServerVersion,
				ServerSubtitle: rpcApp.ServerSubtitle,
				ClusterName:    rpcApp.ClusterName,
				ServerProtocol: rpcApp.ServiceProtocol,
				ServerAddr:     rpcApp.ServiceAddr,
				DepartmentName: rpcApp.DepartmentName,
			}
			applications = append(applications, app)
		}

		// 转换IDC信息
		var idcInfo types.IdcInfo
		if rpcHost.IdcInfo != nil {
			idcInfo = types.IdcInfo{
				ID:             int(rpcHost.IdcInfo.Id),
				IdcName:        rpcHost.IdcInfo.IdcName,
				IdcCode:        rpcHost.IdcInfo.IdcCode,
				IdcLocation:    rpcHost.IdcInfo.IdcLocation,
				IdcDescription: rpcHost.IdcInfo.IdcDescription,
			}
		}

		// 转换主机信息
		hostPool := types.HostPool{
			ID:               int(rpcHost.Id),
			HostName:         rpcHost.Hostname,
			HostIP:           rpcHost.HostIp,
			HostType:         rpcHost.HostType,
			H3cID:            rpcHost.H3CId,
			H3cStatus:        rpcHost.H3CStatus,
			DiskSize:         int(rpcHost.Disk),
			RAM:              int(rpcHost.Ram),
			VCPUs:            int(rpcHost.VCpu),
			IfH3cSync:        rpcHost.IfH3CSync,
			H3cImgID:         rpcHost.H3CImgId,
			H3cHmName:        rpcHost.H3CHmName,
			IsDelete:         rpcHost.IsDelete,
			LeafNumber:       rpcHost.LeafNumber,
			RackNumber:       rpcHost.RackNumber,
			RackHeight:       int(rpcHost.RackHeight),
			RackStartNumber:  int(rpcHost.RackStartNumber),
			FromFactor:       int(rpcHost.FromFactor),
			SerialNumber:     rpcHost.SerialNumber,
			IsDeleted:        rpcHost.IsDeleted,
			IsStatic:         rpcHost.IsStatic,
			IdcInfo:          idcInfo,
			HostApplications: applications,
		}
		hostPools = append(hostPools, hostPool)
	}

	l.Logger.Infof("成功获取%d台主机的详情信息", len(hostPools))
	return &types.HostPoolListResponse{
		List: hostPools,
	}, nil
}
