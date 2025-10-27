package logic

import (
	"context"

	"cmdb-rpc/cmpool"
	"cmdb-rpc/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
)

type UserLogoutLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewUserLogoutLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UserLogoutLogic {
	return &UserLogoutLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 用户登出
func (l *UserLogoutLogic) UserLogout(in *cmpool.LogoutReq) (*cmpool.LogoutResp, error) {
	// todo: add your logic here and delete this line

	return &cmpool.LogoutResp{}, nil
}
