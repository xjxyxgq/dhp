package logic

import (
	"context"
	"fmt"
	"time"

	"cmdb-api/cmpool"
	"cmdb-api/internal/svc"
	"cmdb-api/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetServerResourcesMaxLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewGetServerResourcesMaxLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetServerResourcesMaxLogic {
	return &GetServerResourcesMaxLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetServerResourcesMaxLogic) GetServerResourcesMax(req *types.ServerResourceRequest) (resp *types.ServerResourceMaxListResponse, err error) {
	l.Logger.Info("开始调用RPC获取服务器资源最大值信息")

	// 设置时间范围
	var beginTime, endTime time.Time

	// 如果请求中有时间参数，则使用请求的时间
	if req.StartDate != "" && req.EndDate != "" {
		beginTime, err = time.Parse("2006-01-02", req.StartDate)
		if err != nil {
			l.Logger.Errorf("解析开始时间失败: %v", err)
			// 如果解析失败，使用默认时间范围（最近3个月）
			endTime = time.Now()
			beginTime = endTime.AddDate(0, -3, 0)
		} else {
			endTime, err = time.Parse("2006-01-02", req.EndDate)
			if err != nil {
				l.Logger.Errorf("解析结束时间失败: %v", err)
				// 如果解析失败，使用默认时间范围（最近3个月）
				endTime = time.Now()
				beginTime = endTime.AddDate(0, -3, 0)
			}
		}
	} else {
		// 如果没有时间参数，使用默认时间范围（最近3个月）
		endTime = time.Now().AddDate(0, 0, 1)
		beginTime = endTime.AddDate(0, -3, 0)
	}

	l.Logger.Infof("查询时间范围: %s 到 %s", beginTime.Format("2006-01-02 15:04:05"), endTime.Format("2006-01-02 15:04:05"))

	// 调用RPC服务获取服务器资源最大值信息
	rpcReq := &cmpool.ServerResourceMaxReq{
		BeginTime: beginTime.Format("2006-01-02 15:04:05"),
		EndTime:   endTime.Format("2006-01-02 15:04:05"),
		IpList:    []string{},  // 先初始化为空列表
		Cluster:   req.Cluster, // 添加集群参数支持
	}

	// 如果请求中有IP参数，则传递给RPC服务
	if req.Ip != "" {
		rpcReq.IpList = []string{req.Ip}
		l.Logger.Infof("传递IP过滤条件: %s", req.Ip)
	}

	// 如果请求中有集群参数，则传递给RPC服务
	if req.Cluster != "" {
		l.Logger.Infof("传递集群过滤条件: %s", req.Cluster)
	}

	if req.Ip == "" && req.Cluster == "" {
		l.Logger.Info("未指定过滤条件，获取所有服务器资源最大值信息")
	}

	rpcResp, err := l.svcCtx.CmpoolRpc.GetServerResourceMax(l.ctx, rpcReq)
	if err != nil {
		l.Logger.Errorf("调用RPC获取服务器资源最大值信息失败: %v", err)
		return &types.ServerResourceMaxListResponse{
			Success: false,
			Message: fmt.Sprintf("调用RPC服务失败: %v", err),
			List:    []types.ServerResourceMax{},
		}, nil
	}

	if !rpcResp.Success {
		l.Logger.Errorf("RPC返回失败: %s", rpcResp.Message)
		return &types.ServerResourceMaxListResponse{
			Success: false,
			Message: rpcResp.Message,
			List:    []types.ServerResourceMax{},
		}, nil
	}

	// 转换RPC响应为API响应格式
	serverResourcesMax := make([]types.ServerResourceMax, 0, len(rpcResp.ServerResourceMax))
	for _, rpcResource := range rpcResp.ServerResourceMax {
		// 提取第一个集群信息作为主要集群信息
		clusterName := "未分配集群"
		groupName := "未分配集群组"
		if len(rpcResource.Clusters) > 0 {
			clusterName = rpcResource.Clusters[0].ClusterName
			groupName = rpcResource.Clusters[0].ClusterGroupName
		}

		resource := types.ServerResourceMax{
			ID:            int(rpcResource.Id),
			CreateAt:      rpcResource.CreateAt,
			UpdateAt:      rpcResource.UpdateAt,
			PoolID:        int(rpcResource.PoolId),
			ClusterName:   clusterName,
			GroupName:     groupName,
			Ip:            rpcResource.Ip,
			TotalMemory:   float64(rpcResource.TotalMemory),
			MaxUsedMemory: float64(rpcResource.MaxUsedMemory),
			TotalDisk:     float64(rpcResource.TotalDisk),
			MaxUsedDisk:   float64(rpcResource.MaxUsedDisk),
			CPUCores:      int(rpcResource.CPUCores),
			MaxCPULoad:    float64(rpcResource.MaxCPULoad),
			MaxDateTime:   rpcResource.MaxDateTime,
			HostName:      rpcResource.HostName,
			HostType:      rpcResource.HostType,
			// 百分比字段映射
			CpuPercentMax:  rpcResource.CpuPercentMax,
			CpuPercentAvg:  rpcResource.CpuPercentAvg,
			CpuPercentMin:  rpcResource.CpuPercentMin,
			MemPercentMax:  rpcResource.MemPercentMax,
			MemPercentAvg:  rpcResource.MemPercentAvg,
			MemPercentMin:  rpcResource.MemPercentMin,
			DiskPercentMax: rpcResource.DiskPercentMax,
			DiskPercentAvg: rpcResource.DiskPercentAvg,
			DiskPercentMin: rpcResource.DiskPercentMin,
		}

		// 转换IDC信息
		if rpcResource.IdcInfo != nil {
			resource.IdcInfo = types.IdcInfo{
				ID:             int(rpcResource.IdcInfo.Id),
				IdcName:        rpcResource.IdcInfo.IdcName,
				IdcCode:        rpcResource.IdcInfo.IdcCode,
				IdcLocation:    rpcResource.IdcInfo.IdcLocation,
				IdcDescription: rpcResource.IdcInfo.IdcDescription,
			}
		}

		// 转换集群信息数组
		if len(rpcResource.Clusters) > 0 {
			for _, cluster := range rpcResource.Clusters {
				resource.Clusters = append(resource.Clusters, types.HostClusterInfo{
					ClusterName:      cluster.ClusterName,
					ClusterGroupName: cluster.ClusterGroupName,
					DepartmentName:   cluster.DepartmentName,
				})
			}
		} else {
			resource.Clusters = append(resource.Clusters, types.HostClusterInfo{
				ClusterName:      "未分配集群",
				ClusterGroupName: "未分配集群组",
				DepartmentName:   "未知业务线",
			})
		}

		serverResourcesMax = append(serverResourcesMax, resource)
	}

	l.Logger.Infof("成功获取%d条服务器资源最大值信息", len(serverResourcesMax))
	return &types.ServerResourceMaxListResponse{
		Success: true,
		Message: "查询成功",
		List:    serverResourcesMax,
	}, nil
}
