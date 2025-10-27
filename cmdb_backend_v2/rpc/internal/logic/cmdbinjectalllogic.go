package logic

import (
	"context"

	"cmdb-rpc/cmpool"
	"cmdb-rpc/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
)

type CmdbInjectAllLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewCmdbInjectAllLogic(ctx context.Context, svcCtx *svc.ServiceContext) *CmdbInjectAllLogic {
	return &CmdbInjectAllLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 向CMDB回写数据
func (l *CmdbInjectAllLogic) CmdbInjectAll(in *cmpool.InjectAllReq) (*cmpool.InjectAllResp, error) {
	// todo: add your logic here and delete this line

	return &cmpool.InjectAllResp{}, nil
}
