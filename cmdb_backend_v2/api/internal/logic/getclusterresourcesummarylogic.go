package logic

import (
	"context"

	"cmdb-api/internal/svc"
	"cmdb-api/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetClusterResourceSummaryLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewGetClusterResourceSummaryLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetClusterResourceSummaryLogic {
	return &GetClusterResourceSummaryLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetClusterResourceSummaryLogic) GetClusterResourceSummary() (resp *types.ClusterResourceSummaryListResponse, err error) {
	// todo: add your logic here and delete this line

	return
}
