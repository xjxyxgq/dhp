package logic

import (
	"context"

	"cmdb-api/internal/svc"
	"cmdb-api/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type VerifyMonitoringDataOptionsLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewVerifyMonitoringDataOptionsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *VerifyMonitoringDataOptionsLogic {
	return &VerifyMonitoringDataOptionsLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *VerifyMonitoringDataOptionsLogic) VerifyMonitoringDataOptions() (resp *types.BaseResponse, err error) {
	return &types.BaseResponse{
		Success: true,
		Message: "OPTIONS request handled successfully",
	}, nil
}
