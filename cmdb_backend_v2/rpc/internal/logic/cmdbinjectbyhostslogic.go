package logic

import (
	"context"

	"cmdb-rpc/cmpool"
	"cmdb-rpc/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
)

type CmdbInjectByHostsLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewCmdbInjectByHostsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *CmdbInjectByHostsLogic {
	return &CmdbInjectByHostsLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 向 cmdb 中注入维护后的数据库资源池中对象的数据
func (l *CmdbInjectByHostsLogic) CmdbInjectByHosts(in *cmpool.InjectHostsReq) (*cmpool.InjectHostsResp, error) {
	// todo: add your logic here and delete this line

	return &cmpool.InjectHostsResp{}, nil
}
