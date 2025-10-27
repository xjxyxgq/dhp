package logic

import (
	"context"

	"cmdb-api/internal/svc"
	"cmdb-api/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetClusterResourceUsageLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewGetClusterResourceUsageLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetClusterResourceUsageLogic {
	return &GetClusterResourceUsageLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetClusterResourceUsageLogic) GetClusterResourceUsage() (resp *types.ServerResourceListResponse, err error) {
	// todo: add your logic here and delete this line

	return
}
