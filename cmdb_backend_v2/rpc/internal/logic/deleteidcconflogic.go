package logic

import (
	"context"

	"cmdb-rpc/cmpool"
	"cmdb-rpc/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
)

type DeleteIdcConfLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewDeleteIdcConfLogic(ctx context.Context, svcCtx *svc.ServiceContext) *DeleteIdcConfLogic {
	return &DeleteIdcConfLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 删除IDC机房配置
func (l *DeleteIdcConfLogic) DeleteIdcConf(in *cmpool.DeleteIdcConfReq) (*cmpool.DeleteIdcConfResp, error) {
	// todo: add your logic here and delete this line

	return &cmpool.DeleteIdcConfResp{}, nil
}
