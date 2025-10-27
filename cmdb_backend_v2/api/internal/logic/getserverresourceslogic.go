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

type GetServerResourcesLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewGetServerResourcesLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetServerResourcesLogic {
	return &GetServerResourcesLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetServerResourcesLogic) GetServerResources(req *types.ServerResourceRequest) (resp *types.ServerResourceListResponse, err error) {
	l.Logger.Info("开始调用RPC获取服务器资源信息")

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

	// 调用RPC服务获取服务器资源信息
	rpcReq := &cmpool.ServerResourceReq{
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
		l.Logger.Info("未指定过滤条件，获取所有服务器资源信息")
	}

	rpcResp, err := l.svcCtx.CmpoolRpc.GetServerResource(l.ctx, rpcReq)
	if err != nil {
		l.Logger.Errorf("调用RPC获取服务器资源信息失败: %v", err)
		return &types.ServerResourceListResponse{
			Success: false,
			Message: fmt.Sprintf("调用RPC服务失败: %v", err),
			List:    []types.ServerResource{},
		}, nil
	}

	if !rpcResp.Success {
		l.Logger.Errorf("RPC返回失败: %s", rpcResp.Message)
		return &types.ServerResourceListResponse{
			Success: false,
			Message: rpcResp.Message,
			List:    []types.ServerResource{},
		}, nil
	}

	// 转换RPC响应为API响应格式
	serverResources := make([]types.ServerResource, 0, len(rpcResp.ServerResource))
	for _, rpcResource := range rpcResp.ServerResource {
		resource := types.ServerResource{
			ID:          int(rpcResource.Id),
			CreateAt:    rpcResource.CreateAt,
			UpdateAt:    rpcResource.UpdateAt,
			PoolID:      int(rpcResource.PoolId),
			Ip:          rpcResource.Ip,
			TotalMemory: float64(rpcResource.TotalMemory),
			UsedMemory:  float64(rpcResource.UsedMemory),
			TotalDisk:   float64(rpcResource.TotalDisk),
			UsedDisk:    float64(rpcResource.UsedDisk),
			CPUCores:    int(rpcResource.CPUCores),
			CPULoad:     float64(rpcResource.CPULoad),
			DateTime:    rpcResource.Datetime,
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

		// 转换集群信息
		for _, cluster := range rpcResource.Clusters {
			resource.Clusters = append(resource.Clusters, types.HostClusterInfo{
				ClusterName:      cluster.ClusterName,
				ClusterGroupName: cluster.ClusterGroupName,
			})
		}

		serverResources = append(serverResources, resource)
	}

	l.Logger.Infof("成功获取%d条服务器资源信息", len(serverResources))
	return &types.ServerResourceListResponse{
		Success: true,
		Message: "查询成功",
		List:    serverResources,
	}, nil
}
