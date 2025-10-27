package logic

import (
	"context"

	"cmdb-api/cmpool"
	"cmdb-api/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
)

type LoadServerMetricsFromCsvLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewLoadServerMetricsFromCsvLogic(ctx context.Context, svcCtx *svc.ServiceContext) *LoadServerMetricsFromCsvLogic {
	return &LoadServerMetricsFromCsvLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *LoadServerMetricsFromCsvLogic) LoadServerMetricsFromCsv(in *cmpool.LoadServerMetricsCSVReq, stream cmpool.Cmpool_LoadServerMetricsFromCsvServer) error {
	// todo: add your logic here and delete this line

	return nil
}
