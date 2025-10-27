package logic

import (
	"context"

	"cmdb-api/cmpool"
	"cmdb-api/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
)

type CreateUserLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewCreateUserLogic(ctx context.Context, svcCtx *svc.ServiceContext) *CreateUserLogic {
	return &CreateUserLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 创建用户（CAS用户自动创建）
func (l *CreateUserLogic) CreateUser(in *cmpool.CreateUserReq) (*cmpool.CreateUserResp, error) {
	// todo: add your logic here and delete this line

	return &cmpool.CreateUserResp{}, nil
}
