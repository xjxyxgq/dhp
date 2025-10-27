package logic

import (
	"context"

	"cmdb-api/cmpool"
	"cmdb-api/internal/svc"
	"cmdb-api/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetClusterGroupsLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewGetClusterGroupsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetClusterGroupsLogic {
	return &GetClusterGroupsLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetClusterGroupsLogic) GetClusterGroups() (resp *types.ClusterGroupListResponse, err error) {
	l.Logger.Info("开始调用RPC获取集群组信息")

	// 调用RPC服务获取集群组信息
	rpcReq := &cmpool.ClusterGroupsReq{}

	rpcResp, err := l.svcCtx.CmpoolRpc.GetClusterGroups(l.ctx, rpcReq)
	if err != nil {
		l.Logger.Errorf("调用RPC获取集群组信息失败: %v", err)
		return nil, err
	}

	if !rpcResp.Success {
		l.Logger.Errorf("RPC返回失败: %s", rpcResp.Message)
		return &types.ClusterGroupListResponse{
			List: []types.ClusterGroup{},
		}, nil
	}

	// 转换RPC响应为API响应格式
	clusterGroups := make([]types.ClusterGroup, 0, len(rpcResp.ClusterGroup))
	for _, rpcGroup := range rpcResp.ClusterGroup {
		group := types.ClusterGroup{
			ID:             int(rpcGroup.Id),
			ClusterName:    rpcGroup.ClusterName,
			GroupName:      rpcGroup.GroupName,
			DepartmentName: rpcGroup.DepartmentName,
		}
		clusterGroups = append(clusterGroups, group)
	}

	l.Logger.Infof("成功获取%d个集群组信息", len(clusterGroups))
	return &types.ClusterGroupListResponse{
		List: clusterGroups,
	}, nil
}
