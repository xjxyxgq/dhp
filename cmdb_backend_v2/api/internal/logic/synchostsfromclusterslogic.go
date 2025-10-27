package logic

import (
	"context"

	"cmdb-api/cmpool"
	"cmdb-api/internal/svc"
	"cmdb-api/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type SyncHostsFromClustersLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewSyncHostsFromClustersLogic(ctx context.Context, svcCtx *svc.ServiceContext) *SyncHostsFromClustersLogic {
	return &SyncHostsFromClustersLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *SyncHostsFromClustersLogic) SyncHostsFromClusters() (resp *types.SyncClusterGroupsResponse, err error) {
	l.Logger.Info("开始调用RPC同步集群实例主机数据")

	// 调用RPC服务同步主机数据
	rpcReq := &cmpool.SyncClusterGroupsReq{}

	rpcResp, err := l.svcCtx.CmpoolRpc.SyncHostsFromClusters(l.ctx, rpcReq)
	if err != nil {
		l.Logger.Errorf("调用RPC同步主机失败: %v", err)
		return &types.SyncClusterGroupsResponse{
			Success:     false,
			Message:     "调用主机同步服务失败",
			SyncedCount: 0,
			Details:     []types.DatabaseSyncDetail{},
		}, nil
	}

	// 转换RPC的详细信息为API类型
	var details []types.DatabaseSyncDetail
	for _, rpcDetail := range rpcResp.Details {
		detail := types.DatabaseSyncDetail{
			DatabaseType:  rpcDetail.DatabaseType,
			SyncedCount:   int(rpcDetail.SyncedCount),
			ClusterGroups: rpcDetail.ClusterGroups,
		}
		details = append(details, detail)
	}

	l.Logger.Infof("RPC主机同步完成，结果: %s，总同步主机数: %d", rpcResp.Message, rpcResp.SyncedCount)

	// 记录详细信息到日志
	for _, detail := range details {
		if detail.SyncedCount > 0 {
			l.Logger.Infof("API层 - %s: %d 个主机", detail.DatabaseType, detail.SyncedCount)
		}
	}

	return &types.SyncClusterGroupsResponse{
		Success:     rpcResp.Success,
		Message:     rpcResp.Message,
		SyncedCount: int(rpcResp.SyncedCount),
		Details:     details,
	}, nil
}
