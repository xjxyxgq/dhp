package logic

import (
	"context"

	"cmdb-rpc/cmpool"
	"cmdb-rpc/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetHostsPoolLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGetHostsPoolLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetHostsPoolLogic {
	return &GetHostsPoolLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 查询主机信息，主页面数据展示，返回的数据实际上和detail接口一致，这里没有做区别实现，如果有需要可以添加
func (l *GetHostsPoolLogic) GetHostsPool(in *cmpool.GetHostsPoolReq) (*cmpool.GetHostsDetailResp, error) {
	// todo: add your logic here and delete this line

	return &cmpool.GetHostsDetailResp{}, nil
}
