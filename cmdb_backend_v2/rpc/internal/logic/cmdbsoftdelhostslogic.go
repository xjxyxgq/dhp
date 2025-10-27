package logic

import (
	"context"

	"cmdb-rpc/cmpool"
	"cmdb-rpc/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
)

type CmdbSoftDelHostsLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewCmdbSoftDelHostsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *CmdbSoftDelHostsLogic {
	return &CmdbSoftDelHostsLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 软删除资源池主机
func (l *CmdbSoftDelHostsLogic) CmdbSoftDelHosts(in *cmpool.SoftDelHostsReq) (*cmpool.SoftDelHostsResp, error) {
	// todo: add your logic here and delete this line

	return &cmpool.SoftDelHostsResp{}, nil
}
