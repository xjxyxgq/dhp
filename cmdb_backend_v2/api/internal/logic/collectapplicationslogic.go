package logic

import (
	"context"

	"cmdb-api/cmpool"
	"cmdb-api/internal/svc"
	"cmdb-api/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type CollectApplicationsLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewCollectApplicationsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *CollectApplicationsLogic {
	return &CollectApplicationsLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *CollectApplicationsLogic) CollectApplications() (resp *types.BaseResponse, err error) {
	l.Logger.Info("开始搜集应用数据")

	// 调用RPC服务搜集应用数据
	rpcReq := &cmpool.CollAppsReq{
		// 根据需要设置请求参数
	}
	
	rpcResp, err := l.svcCtx.CmpoolRpc.CollectApplications(l.ctx, rpcReq)
	if err != nil {
		l.Logger.Errorf("调用RPC搜集应用数据失败: %v", err)
		return &types.BaseResponse{
			Success: false,
			Message: "搜集应用数据失败: " + err.Error(),
		}, nil
	}

	l.Logger.Infof("成功搜集应用数据，返回状态: %t", rpcResp.Success)
	return &types.BaseResponse{
		Success: rpcResp.Success,
		Message: rpcResp.Message,
	}, nil
}
