package logic

import (
	"context"
	"fmt"

	"cmdb-api/internal/svc"
	"cmdb-api/internal/types"
	"cmdb-api/cmpool"

	"github.com/zeromicro/go-zero/core/logx"
)

type ModifyHostsApplicationLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewModifyHostsApplicationLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ModifyHostsApplicationLogic {
	return &ModifyHostsApplicationLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *ModifyHostsApplicationLogic) ModifyHostsApplication(req *types.ModifyHostsApplicationRequest) (resp *types.HostApplicationResponse, err error) {
	l.Logger.Info("开始修改主机应用信息")

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

	rpcReq := &cmpool.MdfHostsAppReq{
		Data: rpcData,
	}

	// 调用RPC服务
	rpcResp, err := l.svcCtx.CmpoolRpc.ModifyHostsApplication(l.ctx, rpcReq)
	if err != nil {
		l.Logger.Errorf("调用RPC ModifyHostsApplication失败: %v", err)
		return &types.HostApplicationResponse{
			Success: false,
			Message: fmt.Sprintf("修改应用失败: %v", err),
		}, nil
	}

	l.Logger.Infof("成功修改主机应用信息，返回状态: %t", rpcResp.Success)
	return &types.HostApplicationResponse{
		Success: rpcResp.Success,
		Message: rpcResp.Message,
	}, nil
}
