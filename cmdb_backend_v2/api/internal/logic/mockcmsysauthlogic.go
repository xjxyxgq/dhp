package logic

import (
	"context"

	"cmdb-api/internal/svc"
	"github.com/zeromicro/go-zero/core/logx"
)

type MockCMSysAuthLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewMockCMSysAuthLogic(ctx context.Context, svcCtx *svc.ServiceContext) *MockCMSysAuthLogic {
	return &MockCMSysAuthLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *MockCMSysAuthLogic) MockCMSysAuth() error {
	// todo: add your logic here and delete this line

	return nil
}
