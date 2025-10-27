package logic

import (
	"context"

	"cmdb-rpc/cmpool"
	"cmdb-rpc/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
)

type UpdateIdcConfLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewUpdateIdcConfLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UpdateIdcConfLogic {
	return &UpdateIdcConfLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 更新IDC机房配置
func (l *UpdateIdcConfLogic) UpdateIdcConf(in *cmpool.UpdateIdcConfReq) (*cmpool.UpdateIdcConfResp, error) {
	// todo: add your logic here and delete this line

	return &cmpool.UpdateIdcConfResp{}, nil
}
