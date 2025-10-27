package logic

import (
	"context"
	"database/sql"
	"errors"

	"cmdb-rpc/cmpool"
	"cmdb-rpc/common/auth"
	"cmdb-rpc/internal/svc"

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
	// 使用JWT服务验证token
	jwtService := auth.NewJWTService(
		l.svcCtx.Config.SSOAuth.JWTSecret,
		l.svcCtx.Config.SSOAuth.TokenExpireHours,
	)

	claims, err := jwtService.ValidateToken(in.Token)
	if err != nil {
		return &cmpool.ValidateTokenResp{
			Valid:   false,
			Message: "Token无效或已过期",
		}, nil
	}

	// 查询用户信息
	user, err := l.svcCtx.UserModel.FindOne(claims.UserID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return &cmpool.ValidateTokenResp{
				Valid:   false,
				Message: "用户不存在",
			}, nil
		}
		l.Logger.Errorf("查询用户失败: %v", err)
		return &cmpool.ValidateTokenResp{
			Valid:   false,
			Message: "验证失败",
		}, nil
	}

	// 检查用户是否还是活跃状态
	if !user.IsActive {
		return &cmpool.ValidateTokenResp{
			Valid:   false,
			Message: "用户账号已被禁用",
		}, nil
	}

	// 检查会话是否存在且有效
	session, err := l.svcCtx.UserSessionModel.FindByToken(in.Token)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return &cmpool.ValidateTokenResp{
				Valid:   false,
				Message: "会话不存在",
			}, nil
		}
		l.Logger.Errorf("查询会话失败: %v", err)
		return &cmpool.ValidateTokenResp{
			Valid:   false,
			Message: "验证失败",
		}, nil
	}

	if !session.IsActive {
		return &cmpool.ValidateTokenResp{
			Valid:   false,
			Message: "会话已失效",
		}, nil
	}

	email := ""
	if user.Email.Valid {
		email = user.Email.String
	}

	displayName := user.Username
	if user.DisplayName.Valid {
		displayName = user.DisplayName.String
	}

	lastLoginAt := ""
	if user.LastLoginAt.Valid {
		lastLoginAt = user.LastLoginAt.Time.Format("2006-01-02 15:04:05")
	}

	return &cmpool.ValidateTokenResp{
		Valid: true,
		UserInfo: &cmpool.UserInfo{
			Id:          user.Id,
			Username:    user.Username,
			Email:       email,
			DisplayName: displayName,
			IsActive:    user.IsActive,
			IsAdmin:     user.IsAdmin,
			LoginSource: user.LoginSource,
			LastLoginAt: lastLoginAt,
		},
		Message: "Token验证成功",
	}, nil
}
