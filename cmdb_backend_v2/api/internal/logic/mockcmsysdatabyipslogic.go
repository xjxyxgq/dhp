package logic

import (
	"context"

	"cmdb-api/internal/svc"
	"github.com/zeromicro/go-zero/core/logx"
)

type MockCMSysDataByIPsLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewMockCMSysDataByIPsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *MockCMSysDataByIPsLogic {
	return &MockCMSysDataByIPsLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *MockCMSysDataByIPsLogic) MockCMSysDataByIPs() error {
	// todo: add your logic here and delete this line

	return nil
}
