package logic

import (
	"context"
	"fmt"

	"cmdb-rpc/cmpool"
	"cmdb-rpc/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetClusterResourcesLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGetClusterResourcesLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetClusterResourcesLogic {
	return &GetClusterResourcesLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 获取集群资源详细信息
func (l *GetClusterResourcesLogic) GetClusterResources(in *cmpool.ClusterResourceReq) (*cmpool.ClusterResourceResp, error) {
	// 调用 model 层的方法查询集群资源数据
	clusterResourcesData, err := l.svcCtx.ServerResourcesModel.FindClusterResources(l.ctx, in.BeginTime, in.EndTime, in.ClusterName, in.GroupName)
	if err != nil {
		logx.Errorf("查询集群资源数据失败: %v", err)
		return &cmpool.ClusterResourceResp{
			Success: false,
			Message: fmt.Sprintf("查询集群资源数据失败: %v", err),
		}, nil
	}

	// 转换数据结构为 protobuf 格式
	var clusterResources []*cmpool.ClusterMemberResource
	for _, data := range clusterResourcesData {
		resource := &cmpool.ClusterMemberResource{
			Id:               data.Id,
			ClusterName:      data.ClusterName,
			ClusterGroupName: data.ClusterGroupName,
			Ip:               data.Ip,
			HostName:         data.HostName,
			Port:             data.Port,
			InstanceRole:     data.InstanceRole,
			DateTime:         data.MonDate,
			DepartmentName:   []string{data.DepartmentName},
		}

		// 处理可能为 NULL 的字段
		if data.TotalMemory.Valid {
			resource.TotalMemory = float32(data.TotalMemory.Float64)
		}
		if data.UsedMemory.Valid {
			resource.UsedMemory = float32(data.UsedMemory.Float64)
		}
		if data.TotalDisk.Valid {
			resource.TotalDisk = float32(data.TotalDisk.Float64)
		}
		if data.UsedDisk.Valid {
			resource.UsedDisk = float32(data.UsedDisk.Float64)
		}
		if data.CPUCores.Valid {
			resource.CPUCores = data.CPUCores.Int32
		}
		if data.CPULoad.Valid {
			resource.CPULoad = float32(data.CPULoad.Float64)
		}

		// 处理百分比字段
		if data.CpuPercentMax.Valid {
			resource.CpuPercentMax = data.CpuPercentMax.Float64
		}
		if data.CpuPercentAvg.Valid {
			resource.CpuPercentAvg = data.CpuPercentAvg.Float64
		}
		if data.CpuPercentMin.Valid {
			resource.CpuPercentMin = data.CpuPercentMin.Float64
		}
		if data.MemPercentMax.Valid {
			resource.MemPercentMax = data.MemPercentMax.Float64
		}
		if data.MemPercentAvg.Valid {
			resource.MemPercentAvg = data.MemPercentAvg.Float64
		}
		if data.MemPercentMin.Valid {
			resource.MemPercentMin = data.MemPercentMin.Float64
		}
		if data.DiskPercentMax.Valid {
			resource.DiskPercentMax = data.DiskPercentMax.Float64
		}
		if data.DiskPercentAvg.Valid {
			resource.DiskPercentAvg = data.DiskPercentAvg.Float64
		}
		if data.DiskPercentMin.Valid {
			resource.DiskPercentMin = data.DiskPercentMin.Float64
		}

		clusterResources = append(clusterResources, resource)
	}

	return &cmpool.ClusterResourceResp{
		Success:          true,
		Message:          "查询成功",
		ClusterResources: clusterResources,
	}, nil
}
