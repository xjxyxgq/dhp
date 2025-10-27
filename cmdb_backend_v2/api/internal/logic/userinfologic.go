package logic

import (
	"context"

	"cmdb-api/internal/middleware"
	"cmdb-api/internal/svc"
	"cmdb-api/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type UserInfoLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewUserInfoLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UserInfoLogic {
	return &UserInfoLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

// UserInfo 获取用户信息
func (l *UserInfoLogic) UserInfo() (*types.UserInfoResponse, error) {
	// 从上下文中获取用户信息
	user, ok := middleware.GetUserFromContext(l.ctx)
	if !ok {
		return &types.UserInfoResponse{
			Code:    401,
			Message: "用户信息获取失败",
		}, nil
	}

	userInfo := &types.UserInfo{
		ID:          user.UserID,
		Username:    user.Username,
		DisplayName: user.DisplayName,
		IsAdmin:     user.IsAdmin,
		LoginSource: user.LoginSource,
		IsActive:    true,
	}

	return &types.UserInfoResponse{
		Code:    200,
		Message: "获取用户信息成功",
		Data:    *userInfo,
	}, nil
}
