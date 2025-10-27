package logic

import (
	"context"
	"fmt"

	"cmdb-api/cmpool"
	"cmdb-api/internal/svc"
	"cmdb-api/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetClusterResourcesLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewGetClusterResourcesLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetClusterResourcesLogic {
	return &GetClusterResourcesLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetClusterResourcesLogic) GetClusterResources(req *types.ClusterResourceRequest) (resp *types.ClusterResourceListResponse, err error) {
	// 构建 RPC 请求
	rpcReq := &cmpool.ClusterResourceReq{
		BeginTime:   req.StartDate,
		EndTime:     req.EndDate,
		ClusterName: req.ClusterName,
		GroupName:   req.GroupName,
	}

	// 调用 RPC 服务
	rpcResp, err := l.svcCtx.CmpoolRpc.GetClusterResources(l.ctx, rpcReq)
	if err != nil {
		logx.Errorf("调用GetClusterResources RPC失败: %v", err)
		return &types.ClusterResourceListResponse{
			Success: false,
			Message: fmt.Sprintf("RPC调用失败: %v", err),
		}, nil
	}

	// 转换数据格式
	var clusterList []types.ClusterMemberResource
	for _, item := range rpcResp.ClusterResources {
		resource := types.ClusterMemberResource{
			ID:               int(item.Id),
			ClusterName:      item.ClusterName,
			ClusterGroupName: item.ClusterGroupName,
			Ip:               item.Ip,
			HostName:         item.HostName,
			Port:             int(item.Port),
			InstanceRole:     item.InstanceRole,
			TotalMemory:      float64(item.TotalMemory),
			UsedMemory:       float64(item.UsedMemory),
			TotalDisk:        float64(item.TotalDisk),
			UsedDisk:         float64(item.UsedDisk),
			CPUCores:         int(item.CPUCores),
			CPULoad:          float64(item.CPULoad),
			DateTime:         item.DateTime,
			// 百分比字段映射
			CpuPercentMax:    item.CpuPercentMax,
			CpuPercentAvg:    item.CpuPercentAvg,
			CpuPercentMin:    item.CpuPercentMin,
			MemPercentMax:    item.MemPercentMax,
			MemPercentAvg:    item.MemPercentAvg,
			MemPercentMin:    item.MemPercentMin,
			DiskPercentMax:   item.DiskPercentMax,
			DiskPercentAvg:   item.DiskPercentAvg,
			DiskPercentMin:   item.DiskPercentMin,
		}
		clusterList = append(clusterList, resource)
	}

	return &types.ClusterResourceListResponse{
		Success: rpcResp.Success,
		Message: rpcResp.Message,
		List:    clusterList,
	}, nil
}
