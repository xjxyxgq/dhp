package logic

import (
	"context"

	"cmdb-api/internal/svc"
	"cmdb-api/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type AnalyzeResourceUsageLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewAnalyzeResourceUsageLogic(ctx context.Context, svcCtx *svc.ServiceContext) *AnalyzeResourceUsageLogic {
	return &AnalyzeResourceUsageLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *AnalyzeResourceUsageLogic) AnalyzeResourceUsage(req *types.ResourceUsageData) (resp *types.BaseResponse, err error) {
	// todo: add your logic here and delete this line

	return
}
