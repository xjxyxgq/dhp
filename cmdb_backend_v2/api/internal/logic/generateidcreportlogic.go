package logic

import (
	"context"

	"cmdb-api/internal/svc"
	"cmdb-api/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type GenerateIDCReportLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewGenerateIDCReportLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GenerateIDCReportLogic {
	return &GenerateIDCReportLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GenerateIDCReportLogic) GenerateIDCReport() (resp *types.IDCUsageListResponse, err error) {
	// todo: add your logic here and delete this line

	return
}
