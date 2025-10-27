package logic

import (
	"context"

	"cmdb-api/internal/svc"
	"cmdb-api/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type GenerateClusterGroupReportLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewGenerateClusterGroupReportLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GenerateClusterGroupReportLogic {
	return &GenerateClusterGroupReportLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GenerateClusterGroupReportLogic) GenerateClusterGroupReport() (resp *types.ClusterGroupListResponse, err error) {
	// todo: add your logic here and delete this line

	return
}
