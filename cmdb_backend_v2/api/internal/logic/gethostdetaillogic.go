package logic

import (
	"context"

	"cmdb-api/internal/svc"
	"cmdb-api/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetHostDetailLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewGetHostDetailLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetHostDetailLogic {
	return &GetHostDetailLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetHostDetailLogic) GetHostDetail() (resp *types.HostPool, err error) {
	// todo: add your logic here and delete this line

	return
}
