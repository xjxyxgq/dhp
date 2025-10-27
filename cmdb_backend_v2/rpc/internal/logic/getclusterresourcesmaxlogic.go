package logic

import (
	"cmdb-rpc/cmpool"
	"cmdb-rpc/internal/svc"
	"context"
	"fmt"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetClusterResourcesMaxLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGetClusterResourcesMaxLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetClusterResourcesMaxLogic {
	return &GetClusterResourcesMaxLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 获取集群资源最大利用率信息
func (l *GetClusterResourcesMaxLogic) GetClusterResourcesMax(in *cmpool.ClusterResourceReq) (*cmpool.ClusterResourceMaxResp, error) {
	// 调用 model 层的方法查询集群资源最大值数据
	clusterMaxData, err := l.svcCtx.ServerResourcesModel.FindClusterResourcesMax(l.ctx, in.BeginTime, in.EndTime, in.ClusterName, in.GroupName)
	if err != nil {
		logx.Errorf("查询集群资源最大值失败: %v", err)
		return &cmpool.ClusterResourceMaxResp{
			Success: false,
			Message: fmt.Sprintf("查询集群资源最大值失败: %v", err),
		}, nil
	}

	// 构建响应数据
	var clusterMaxList []*cmpool.ClusterResourceMax
	for _, data := range clusterMaxData {
		clusterMax := &cmpool.ClusterResourceMax{
			ClusterName:      data.ClusterName,
			ClusterGroupName: data.ClusterGroupName,
			DepartmentName:   data.DepartmentName,
			NodeCount:        data.NodeCount,
			CPUCores:         data.CPUCores.Int32,
			// 平均值字段
			AvgCPULoad:       float32(data.AvgCPULoad.Float64),
			AvgMemoryUsage:   float32(data.AvgMemoryUsage.Float64),
			AvgDiskUsage:     float32(data.AvgDiskUsage.Float64),
			// 最大值字段
			MaxCPULoad:       float32(data.MaxCPULoad.Float64),
			MaxMemoryUsage:   float32(data.MaxMemoryUsage.Float64),
			MaxDiskUsage:     float32(data.MaxDiskUsage.Float64),
			TotalMemory:      float32(data.TotalMemory.Float64),
			TotalDisk:        float32(data.TotalDisk.Float64),
			MaxDateTime:      data.MaxDateTime.String,
			// 百分比字段映射
			CpuPercentMax:    data.CpuPercentMax.Float64,
			CpuPercentAvg:    data.CpuPercentAvg.Float64,
			CpuPercentMin:    data.CpuPercentMin.Float64,
			MemPercentMax:    data.MemPercentMax.Float64,
			MemPercentAvg:    data.MemPercentAvg.Float64,
			MemPercentMin:    data.MemPercentMin.Float64,
			DiskPercentMax:   data.DiskPercentMax.Float64,
			DiskPercentAvg:   data.DiskPercentAvg.Float64,
			DiskPercentMin:   data.DiskPercentMin.Float64,
		}
		clusterMaxList = append(clusterMaxList, clusterMax)
	}

	return &cmpool.ClusterResourceMaxResp{
		Success:             true,
		Message:             "查询成功",
		ClusterResourcesMax: clusterMaxList,
	}, nil
}
