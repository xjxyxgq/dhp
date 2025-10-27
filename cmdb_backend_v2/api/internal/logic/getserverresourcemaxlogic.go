package logic

import (
	"context"

	"cmdb-api/cmpool"
	"cmdb-api/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetServerResourceMaxLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGetServerResourceMaxLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetServerResourceMaxLogic {
	return &GetServerResourceMaxLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 查询主机资源最大利用率数据
func (l *GetServerResourceMaxLogic) GetServerResourceMax(in *cmpool.ServerResourceMaxReq) (*cmpool.ServerResourceMaxResp, error) {
	// todo: add your logic here and delete this line

	return &cmpool.ServerResourceMaxResp{}, nil
}
