package logic

import (
	"context"

	"cmdb-rpc/cmpool"
	"cmdb-rpc/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
)

type CmdbExtractAllLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewCmdbExtractAllLogic(ctx context.Context, svcCtx *svc.ServiceContext) *CmdbExtractAllLogic {
	return &CmdbExtractAllLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// <必要数据搜集>，全量抽取CMDB中数据库服务器的数据并刷新到数据库资源池
func (l *CmdbExtractAllLogic) CmdbExtractAll(in *cmpool.ExtractAllReq) (*cmpool.ExtractAllResp, error) {
	// todo: add your logic here and delete this line

	return &cmpool.ExtractAllResp{}, nil
}
