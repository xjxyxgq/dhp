package logic

import (
	"context"

	"cmdb-rpc/cmpool"
	"cmdb-rpc/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
)

type CmdbExtractByHostsLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewCmdbExtractByHostsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *CmdbExtractByHostsLogic {
	return &CmdbExtractByHostsLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 从CMDB抽取主机数据
func (l *CmdbExtractByHostsLogic) CmdbExtractByHosts(in *cmpool.ExtractByHostsReq) (*cmpool.ExtractByHostsResp, error) {
	// todo: add your logic here and delete this line

	return &cmpool.ExtractByHostsResp{}, nil
}
