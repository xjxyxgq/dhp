package logic

import (
	"context"

	"cmdb-api/cmpool"
	"cmdb-api/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
)

type CmdbModifyHostsLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewCmdbModifyHostsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *CmdbModifyHostsLogic {
	return &CmdbModifyHostsLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 手动修改资源池主机信息，如果主机信息最终 isStatic 不为 True，那么这里的修改会被后续的任务以真实数据刷掉
func (l *CmdbModifyHostsLogic) CmdbModifyHosts(in *cmpool.ModifyHostsReq) (*cmpool.ModifyHostsResp, error) {
	// todo: add your logic here and delete this line

	return &cmpool.ModifyHostsResp{}, nil
}
