package logic

import (
	"context"

	"cmdb-api/internal/svc"
	"cmdb-api/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type InsertServerResourceLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewInsertServerResourceLogic(ctx context.Context, svcCtx *svc.ServiceContext) *InsertServerResourceLogic {
	return &InsertServerResourceLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *InsertServerResourceLogic) InsertServerResource(req *types.ServerResource) (resp *types.BaseResponse, err error) {
	// todo: add your logic here and delete this line

	return
}
