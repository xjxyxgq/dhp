package logic

import (
	"context"

	"cmdb-api/internal/svc"
	"cmdb-api/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetResourceAlertsLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewGetResourceAlertsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetResourceAlertsLogic {
	return &GetResourceAlertsLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetResourceAlertsLogic) GetResourceAlerts() (resp *types.ServerResourceListResponse, err error) {
	// todo: add your logic here and delete this line

	return
}
