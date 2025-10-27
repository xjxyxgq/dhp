package logic

import (
	"context"

	"cmdb-api/internal/svc"
	"cmdb-api/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type TriggerReportLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewTriggerReportLogic(ctx context.Context, svcCtx *svc.ServiceContext) *TriggerReportLogic {
	return &TriggerReportLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *TriggerReportLogic) TriggerReport(req *types.ReportRequest) (resp *types.BaseResponse, err error) {
	// todo: add your logic here and delete this line

	return
}
