package logic

import (
	"context"

	"cmdb-rpc/cmpool"
	"cmdb-rpc/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
)

type CmdbSyncHostsLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewCmdbSyncHostsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *CmdbSyncHostsLogic {
	return &CmdbSyncHostsLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 数据库主机池数据维护
func (l *CmdbSyncHostsLogic) CmdbSyncHosts(in *cmpool.SyncHostsReq) (*cmpool.SyncHostsResp, error) {
	l.Logger.Infof("开始同步主机数据，数据库类型: %d", in.DBType)

	// 同步hosts_pool数据
	err := l.svcCtx.ExternalAPI.SyncHostsPoolData()
	if err != nil {
		l.Logger.Errorf("同步hosts_pool数据失败: %v", err)
		return &cmpool.SyncHostsResp{
			Success: false,
			Message: "同步hosts_pool数据失败: " + err.Error(),
		}, nil
	}

	// 同步host_application数据
	err = l.svcCtx.DataSync.SyncHostApplications()
	if err != nil {
		l.Logger.Errorf("同步host_application数据失败: %v", err)
		return &cmpool.SyncHostsResp{
			Success: false,
			Message: "同步host_application数据失败: " + err.Error(),
		}, nil
	}

	// 同步cluster_group数据
	err = l.svcCtx.DataSync.SyncClusterGroups()
	if err != nil {
		l.Logger.Errorf("同步cluster_group数据失败: %v", err)
		return &cmpool.SyncHostsResp{
			Success: false,
			Message: "同步cluster_group数据失败: " + err.Error(),
		}, nil
	}

	l.Logger.Info("主机数据同步完成")
	return &cmpool.SyncHostsResp{
		Success: true,
		Message: "主机数据同步成功",
	}, nil
}
