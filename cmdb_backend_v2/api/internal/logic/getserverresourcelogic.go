package logic

import (
	"context"

	"cmdb-api/cmpool"
	"cmdb-api/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetServerResourceLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGetServerResourceLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetServerResourceLogic {
	return &GetServerResourceLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 查询主机资源使用率数据
func (l *GetServerResourceLogic) GetServerResource(in *cmpool.ServerResourceReq) (*cmpool.ServerResourceResp, error) {
	// todo: add your logic here and delete this line

	return &cmpool.ServerResourceResp{}, nil
}
