package logic

import (
	"context"

	"cmdb-api/cmpool"
	"cmdb-api/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
)

type ValidateTokenLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewValidateTokenLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ValidateTokenLogic {
	return &ValidateTokenLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 验证Token
func (l *ValidateTokenLogic) ValidateToken(in *cmpool.ValidateTokenReq) (*cmpool.ValidateTokenResp, error) {
	// todo: add your logic here and delete this line

	return &cmpool.ValidateTokenResp{}, nil
}
