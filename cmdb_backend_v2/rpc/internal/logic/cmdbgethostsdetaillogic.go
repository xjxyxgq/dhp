package logic

import (
	"context"

	"cmdb-rpc/cmpool"
	"cmdb-rpc/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
)

type CmdbGetHostsDetailLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewCmdbGetHostsDetailLogic(ctx context.Context, svcCtx *svc.ServiceContext) *CmdbGetHostsDetailLogic {
	return &CmdbGetHostsDetailLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 获取资源池主机信息，包括cmdb信息，应用部署信息等所有相关数据（这里是根据hosts_pool实时检索，数据并不来自于hosts_application表）
func (l *CmdbGetHostsDetailLogic) CmdbGetHostsDetail(in *cmpool.GetHostsDetailReq) (*cmpool.GetHostsDetailResp, error) {
	// todo: add your logic here and delete this line

	return &cmpool.GetHostsDetailResp{}, nil
}
