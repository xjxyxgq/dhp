package logic

import (
	"context"

	"cmdb-api/cmpool"
	"cmdb-api/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
)

type CmdbSyncHostsLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewCmdbSyncHostsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *CmdbSyncHostsLogic {
	return &CmdbSyncHostsLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 数据库主机池数据维护
func (l *CmdbSyncHostsLogic) CmdbSyncHosts(in *cmpool.SyncHostsReq) (*cmpool.SyncHostsResp, error) {
	// todo: add your logic here and delete this line

	return &cmpool.SyncHostsResp{}, nil
}
