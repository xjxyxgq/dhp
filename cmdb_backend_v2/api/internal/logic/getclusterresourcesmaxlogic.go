package logic

import (
	"context"
	"fmt"

	"cmdb-api/cmpool"
	"cmdb-api/internal/svc"
	"cmdb-api/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetClusterResourcesMaxLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewGetClusterResourcesMaxLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetClusterResourcesMaxLogic {
	return &GetClusterResourcesMaxLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetClusterResourcesMaxLogic) GetClusterResourcesMax(req *types.ClusterResourceRequest) (resp *types.ClusterResourceMaxListResponse, err error) {
	// 构建 RPC 请求
	rpcReq := &cmpool.ClusterResourceReq{
		BeginTime:   req.StartDate,
		EndTime:     req.EndDate,
		ClusterName: req.ClusterName,
		GroupName:   req.GroupName,
	}

	// 调用 RPC 服务
	rpcResp, err := l.svcCtx.CmpoolRpc.GetClusterResourcesMax(l.ctx, rpcReq)
	if err != nil {
		logx.Errorf("调用GetClusterResourcesMax RPC失败: %v", err)
		return &types.ClusterResourceMaxListResponse{
			Success: false,
			Message: fmt.Sprintf("RPC调用失败: %v", err),
		}, nil
	}

	// 转换数据格式
	var clusterList []types.ClusterResourceMax
	for _, item := range rpcResp.ClusterResourcesMax {
		cluster := types.ClusterResourceMax{
			ClusterName:      item.ClusterName,
			ClusterGroupName: item.ClusterGroupName,
			DepartmentName:   item.DepartmentName,
			NodeCount:        int(item.NodeCount),
			// 平均值字段
			AvgCPULoad:       float64(item.AvgCPULoad),
			AvgMemoryUsage:   float64(item.AvgMemoryUsage),
			AvgDiskUsage:     float64(item.AvgDiskUsage),
			// 最大值字段
			MaxCPULoad:       float64(item.MaxCPULoad),
			MaxMemoryUsage:   float64(item.MaxMemoryUsage),
			MaxDiskUsage:     float64(item.MaxDiskUsage),
			TotalMemory:      float64(item.TotalMemory),
			TotalDisk:        float64(item.TotalDisk),
			MaxUsedMemory:    float64(item.MaxUsedMemory),
			MaxUsedDisk:      float64(item.MaxUsedDisk),
			MaxDateTime:      item.MaxDateTime,
			MemberNodes:      []types.ClusterMemberResource{}, // 暂时为空
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

		clusterList = append(clusterList, cluster)
	}

	return &types.ClusterResourceMaxListResponse{
		Success: rpcResp.Success,
		Message: rpcResp.Message,
		List:    clusterList,
	}, nil
}
