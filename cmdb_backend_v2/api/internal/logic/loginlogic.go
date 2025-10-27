package logic

import (
	"context"

	"cmdb-api/cmpool"
	"cmdb-api/internal/svc"
	"cmdb-api/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type LoginLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewLoginLogic(ctx context.Context, svcCtx *svc.ServiceContext) *LoginLogic {
	return &LoginLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

// Login 用户登录
func (l *LoginLogic) Login(req *types.LoginRequest) (*types.LoginResponse, error) {
	// 如果启用了CAS，不允许本地登录
	if l.svcCtx.Config.SSOAuth.EnableCAS {
		return &types.LoginResponse{
			Code:    400,
			Message: "请使用CAS单点登录",
		}, nil
	}

	// 调用RPC服务进行登录
	resp, err := l.svcCtx.CmpoolRpc.UserLogin(l.ctx, &cmpool.LoginReq{
		Username:    req.Username,
		Password:    req.Password,
		LoginSource: "local",
	})
	if err != nil {
		l.Errorf("RPC登录调用失败: %v", err)
		return &types.LoginResponse{
			Code:    500,
			Message: "登录服务异常，请稍后重试",
		}, nil
	}

	if !resp.Success {
		return &types.LoginResponse{
			Code:    401,
			Message: resp.Message,
		}, nil
	}

	// 转换用户信息
	userInfo := &types.UserInfo{
		ID:          resp.UserInfo.Id,
		Username:    resp.UserInfo.Username,
		Email:       resp.UserInfo.Email,
		DisplayName: resp.UserInfo.DisplayName,
		IsAdmin:     resp.UserInfo.IsAdmin,
		IsActive:    resp.UserInfo.IsActive,
		LoginSource: resp.UserInfo.LoginSource,
		LastLoginAt: resp.UserInfo.LastLoginAt,
	}

	return &types.LoginResponse{
		Code:    200,
		Message: resp.Message,
		Data: types.LoginResponseData{
			Token:       resp.Token,
			User:        *userInfo,
			LoginSource: resp.UserInfo.LoginSource,
		},
	}, nil
}
