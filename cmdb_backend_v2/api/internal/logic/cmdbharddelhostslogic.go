package logic

import (
	"context"

	"cmdb-api/cmpool"
	"cmdb-api/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
)

type CmdbHardDelHostsLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewCmdbHardDelHostsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *CmdbHardDelHostsLogic {
	return &CmdbHardDelHostsLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 硬删除资源池主机
func (l *CmdbHardDelHostsLogic) CmdbHardDelHosts(in *cmpool.HardDelHostsReq) (*cmpool.HardDelHostsResp, error) {
	// todo: add your logic here and delete this line

	return &cmpool.HardDelHostsResp{}, nil
}
