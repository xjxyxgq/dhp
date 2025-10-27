package logic

import (
	"context"

	"cmdb-api/internal/svc"
	"cmdb-api/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetClusterUsageLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewGetClusterUsageLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetClusterUsageLogic {
	return &GetClusterUsageLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetClusterUsageLogic) GetClusterUsage() (resp *types.IDCUsageListResponse, err error) {
	// todo: add your logic here and delete this line

	return
}
