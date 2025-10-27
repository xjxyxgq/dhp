package logic

import (
	"context"

	"cmdb-api/internal/svc"
	"cmdb-api/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetDiskFullPredictionLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewGetDiskFullPredictionLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetDiskFullPredictionLogic {
	return &GetDiskFullPredictionLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetDiskFullPredictionLogic) GetDiskFullPrediction() (resp *types.DiskPredictionResponse, err error) {
	// todo: add your logic here and delete this line

	return
}
