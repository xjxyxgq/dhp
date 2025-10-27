package logic

import (
	"context"
	"fmt"

	"cmdb-api/internal/svc"
	"cmdb-api/internal/types"
	"cmdb-api/cmpool"

	"github.com/zeromicro/go-zero/core/logx"
)

type DeleteHostsApplicationLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewDeleteHostsApplicationLogic(ctx context.Context, svcCtx *svc.ServiceContext) *DeleteHostsApplicationLogic {
	return &DeleteHostsApplicationLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *DeleteHostsApplicationLogic) DeleteHostsApplication(req *types.DeleteHostsApplicationRequest) (resp *types.HostApplicationResponse, err error) {
	l.Logger.Info("开始删除主机应用信息")

	// 转换API请求到RPC请求
	var rpcAppIds []int64
	for _, appId := range req.AppIds {
		rpcAppIds = append(rpcAppIds, int64(appId))
	}

	rpcReq := &cmpool.DelHostsAppReq{
		AppIds: rpcAppIds,
	}

	// 调用RPC服务
	rpcResp, err := l.svcCtx.CmpoolRpc.DeleteHostsApplication(l.ctx, rpcReq)
	if err != nil {
		l.Logger.Errorf("调用RPC DeleteHostsApplication失败: %v", err)
		return &types.HostApplicationResponse{
			Success: false,
			Message: fmt.Sprintf("删除应用失败: %v", err),
		}, nil
	}

	l.Logger.Infof("成功删除主机应用信息，返回状态: %t", rpcResp.Success)
	return &types.HostApplicationResponse{
		Success: rpcResp.Success,
		Message: rpcResp.Message,
	}, nil
}
