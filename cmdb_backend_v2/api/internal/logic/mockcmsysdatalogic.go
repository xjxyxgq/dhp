package logic

import (
	"context"

	"cmdb-api/internal/svc"
	"github.com/zeromicro/go-zero/core/logx"
)

type MockCMSysDataLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewMockCMSysDataLogic(ctx context.Context, svcCtx *svc.ServiceContext) *MockCMSysDataLogic {
	return &MockCMSysDataLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *MockCMSysDataLogic) MockCMSysData() error {
	// todo: add your logic here and delete this line

	return nil
}
