package logic

import (
	"context"

	"cmdb-rpc/cmpool"
	"cmdb-rpc/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
)

type CollectApplicationsLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewCollectApplicationsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *CollectApplicationsLogic {
	return &CollectApplicationsLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 内部数据加工和检索
func (l *CollectApplicationsLogic) CollectApplications(in *cmpool.CollAppsReq) (*cmpool.CollAppsResp, error) {
	// todo: add your logic here and delete this line

	return &cmpool.CollAppsResp{}, nil
}
