package logic

import (
	"context"

	"cmdb-api/cmpool"
	"cmdb-api/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetScheduledTaskDetailLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGetScheduledTaskDetailLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetScheduledTaskDetailLogic {
	return &GetScheduledTaskDetailLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 获取定时任务详情
func (l *GetScheduledTaskDetailLogic) GetScheduledTaskDetail(in *cmpool.GetScheduledTaskDetailReq) (*cmpool.GetScheduledTaskDetailResp, error) {
	// todo: add your logic here and delete this line

	return &cmpool.GetScheduledTaskDetailResp{}, nil
}
