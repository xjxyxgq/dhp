package logic

import (
	"context"

	"cmdb-api/internal/svc"
	"cmdb-api/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type GenerateAndSendReportLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewGenerateAndSendReportLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GenerateAndSendReportLogic {
	return &GenerateAndSendReportLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GenerateAndSendReportLogic) GenerateAndSendReport(req *types.ReportRequest) (resp *types.BaseResponse, err error) {
	// todo: add your logic here and delete this line

	return
}
