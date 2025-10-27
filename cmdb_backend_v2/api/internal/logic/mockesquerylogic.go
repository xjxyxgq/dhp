package logic

import (
	"context"

	"cmdb-api/internal/svc"
	"github.com/zeromicro/go-zero/core/logx"
)

type MockEsQueryLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewMockEsQueryLogic(ctx context.Context, svcCtx *svc.ServiceContext) *MockEsQueryLogic {
	return &MockEsQueryLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *MockEsQueryLogic) MockEsQuery() error {
	// todo: add your logic here and delete this line

	return nil
}
