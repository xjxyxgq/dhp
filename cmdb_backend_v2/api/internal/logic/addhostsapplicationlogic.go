package logic

import (
	"context"
	"fmt"

	"cmdb-api/internal/svc"
	"cmdb-api/internal/types"
	"cmdb-api/cmpool"

	"github.com/zeromicro/go-zero/core/logx"
)

type AddHostsApplicationLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewAddHostsApplicationLogic(ctx context.Context, svcCtx *svc.ServiceContext) *AddHostsApplicationLogic {
	return &AddHostsApplicationLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *AddHostsApplicationLogic) AddHostsApplication(req *types.AddHostsApplicationRequest) (resp *types.HostApplicationResponse, err error) {
	l.Logger.Info("开始添加主机应用信息")

	// 转换API请求到RPC请求
	var rpcData []*cmpool.HostsApp
	for _, apiApp := range req.Data {
		rpcApp := &cmpool.HostsApp{
			HostId:           int32(apiApp.HostId),
			ServerType:       apiApp.ServerType,
			ServerVersion:    apiApp.ServerVersion,
			ServerSubtitle:   apiApp.ServerSubtitle,
			ClusterName:      apiApp.ClusterName,
			ServerProtocol:   apiApp.ServerProtocol,
			ServerAddr:       apiApp.ServerAddr,
			ServerPort:       apiApp.ServerPort,
			ServerRole:       apiApp.ServerRole,
			ServerStatus:     apiApp.ServerStatus,
			DepartmentName:   apiApp.DepartmentName,
		}
		rpcData = append(rpcData, rpcApp)
	}

	rpcReq := &cmpool.AddHostsAppReq{
		Data: rpcData,
	}

	// 调用RPC服务
	rpcResp, err := l.svcCtx.CmpoolRpc.AddHostsApplication(l.ctx, rpcReq)
	if err != nil {
		l.Logger.Errorf("调用RPC AddHostsApplication失败: %v", err)
		return &types.HostApplicationResponse{
			Success: false,
			Message: fmt.Sprintf("添加应用失败: %v", err),
		}, nil
	}

	l.Logger.Infof("成功添加主机应用信息，返回状态: %t", rpcResp.Success)
	return &types.HostApplicationResponse{
		Success: rpcResp.Success,
		Message: rpcResp.Message,
	}, nil
}
