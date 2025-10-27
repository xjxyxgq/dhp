package logic

import (
	"context"

	"cmdb-api/cmpool"
	"cmdb-api/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
)

type CmdbExtractHostsByOwnerLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewCmdbExtractHostsByOwnerLogic(ctx context.Context, svcCtx *svc.ServiceContext) *CmdbExtractHostsByOwnerLogic {
	return &CmdbExtractHostsByOwnerLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 数据搜集，按照cmdb中的ownerGroup抽取所有属于具体组的服务器并写入hosts_pool
func (l *CmdbExtractHostsByOwnerLogic) CmdbExtractHostsByOwner(in *cmpool.ExtractByOwnerReq) (*cmpool.ExtractByOwnerResp, error) {
	// todo: add your logic here and delete this line

	return &cmpool.ExtractByOwnerResp{}, nil
}
