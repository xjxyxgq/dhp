package logic

import (
	"context"

	"cmdb-rpc/cmpool"
	"cmdb-rpc/internal/service"
	"cmdb-rpc/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
)

type HardwareResourceVerificationLogic struct {
	ctx                         context.Context
	svcCtx                      *svc.ServiceContext
	hardwareVerificationService service.HardwareVerificationService
	logx.Logger
}

func NewHardwareResourceVerificationLogic(ctx context.Context, svcCtx *svc.ServiceContext) *HardwareResourceVerificationLogic {
	hwService := service.NewHardwareVerificationService(ctx, svcCtx.Config, svcCtx.HardwareResourceVerificationModel)

	return &HardwareResourceVerificationLogic{
		ctx:                         ctx,
		svcCtx:                      svcCtx,
		hardwareVerificationService: hwService,
		Logger:                      logx.WithContext(ctx),
	}
}

// 硬件资源验证
func (l *HardwareResourceVerificationLogic) HardwareResourceVerification(in *cmpool.HardwareResourceVerificationReq) (*cmpool.HardwareResourceVerificationResp, error) {
	return l.hardwareVerificationService.ExecuteVerification(in)
}
