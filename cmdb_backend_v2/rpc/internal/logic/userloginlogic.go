package logic

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"cmdb-rpc/cmpool"
	"cmdb-rpc/common/auth"
	"cmdb-rpc/internal/model"
	"cmdb-rpc/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
)

type UserLoginLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewUserLoginLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UserLoginLogic {
	return &UserLoginLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 用户认证相关方法
func (l *UserLoginLogic) UserLogin(in *cmpool.LoginReq) (*cmpool.LoginResp, error) {
	// 查找用户
	user, err := l.svcCtx.UserModel.FindByUsername(in.Username)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return &cmpool.LoginResp{
				Success: false,
				Message: "用户名或密码错误",
			}, nil
		}
		l.Logger.Errorf("查询用户失败: %v", err)
		return &cmpool.LoginResp{
			Success: false,
			Message: "登录失败，请稍后重试",
		}, nil
	}

	// 检查用户是否活跃
	if !user.IsActive {
		return &cmpool.LoginResp{
			Success: false,
			Message: "用户账号已被禁用",
		}, nil
	}

	// 验证密码
	if in.LoginSource == "local" {
		if !user.PasswordHash.Valid || user.PasswordHash.String == "" {
			return &cmpool.LoginResp{
				Success: false,
				Message: "用户未设置密码，请使用其他登录方式",
			}, nil
		}

		if !auth.VerifyPassword(user.PasswordHash.String, in.Password) {
			return &cmpool.LoginResp{
				Success: false,
				Message: "用户名或密码错误",
			}, nil
		}
	}

	// 生成JWT Token
	jwtService := auth.NewJWTService(
		l.svcCtx.Config.SSOAuth.JWTSecret,
		l.svcCtx.Config.SSOAuth.TokenExpireHours,
	)

	displayName := user.Username
	if user.DisplayName.Valid {
		displayName = user.DisplayName.String
	}

	email := ""
	if user.Email.Valid {
		email = user.Email.String
	}

	token, err := jwtService.GenerateToken(
		user.Id,
		user.Username,
		displayName,
		user.IsAdmin,
		user.LoginSource,
	)
	if err != nil {
		l.Logger.Errorf("生成Token失败: %v", err)
		return &cmpool.LoginResp{
			Success: false,
			Message: "登录失败，请稍后重试",
		}, nil
	}

	// 更新最后登录时间
	err = l.svcCtx.UserModel.UpdateLoginTime(user.Id, time.Now())
	if err != nil {
		l.Logger.Errorf("更新登录时间失败: %v", err)
	}

	// 创建用户会话记录
	expireTime := time.Now().Add(time.Hour * time.Duration(l.svcCtx.Config.SSOAuth.TokenExpireHours))
	session := &model.UserSession{
		UserId:       int64(user.Id),
		SessionToken: token,
		ExpiresAt:    sql.NullTime{Time: expireTime, Valid: true},
		IsActive:     true,
	}
	_, err = l.svcCtx.UserSessionModel.Insert(session)
	if err != nil {
		l.Logger.Errorf("创建用户会话失败: %v", err)
		return &cmpool.LoginResp{
			Success: false,
			Message: "登录失败，会话创建异常",
		}, nil
	}

	lastLoginAt := ""
	if user.LastLoginAt.Valid {
		lastLoginAt = user.LastLoginAt.Time.Format("2006-01-02 15:04:05")
	}

	return &cmpool.LoginResp{
		Success: true,
		Message: "登录成功",
		Token:   token,
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
	}, nil
}
