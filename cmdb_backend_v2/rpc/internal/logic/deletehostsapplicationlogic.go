package logic

import (
	"context"
	"fmt"

	"cmdb-rpc/cmpool"
	"cmdb-rpc/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
)

type DeleteHostsApplicationLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewDeleteHostsApplicationLogic(ctx context.Context, svcCtx *svc.ServiceContext) *DeleteHostsApplicationLogic {
	return &DeleteHostsApplicationLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 删除资源池主机应用信息
func (l *DeleteHostsApplicationLogic) DeleteHostsApplication(in *cmpool.DelHostsAppReq) (*cmpool.DelHostsAppResp, error) {
	if len(in.AppIds) == 0 {
		return &cmpool.DelHostsAppResp{
			Success: false,
			Message: "应用ID列表不能为空",
		}, nil
	}

	// 执行删除操作
	for _, appId := range in.AppIds {
		err := l.svcCtx.HostsApplicationsModel.Delete(l.ctx, uint64(appId))
		if err != nil {
			l.Errorf("删除应用失败，应用ID: %d, 错误: %v", appId, err)
			return &cmpool.DelHostsAppResp{
				Success: false,
				Message: fmt.Sprintf("删除应用失败: %v", err),
			}, nil
		}
	}

	l.Infof("成功删除 %d 个应用", len(in.AppIds))
	return &cmpool.DelHostsAppResp{
		Success: true,
		Message: fmt.Sprintf("成功删除 %d 个应用", len(in.AppIds)),
	}, nil
}
