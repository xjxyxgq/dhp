package logic

import (
	"context"
	"strings"

	"cmdb-api/cmpool"
	"cmdb-api/internal/svc"
	"cmdb-api/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetDiskPredictionLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewGetDiskPredictionLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetDiskPredictionLogic {
	return &GetDiskPredictionLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetDiskPredictionLogic) GetDiskPrediction(req *types.ServerResourceRequest) (resp *types.DiskPredictionListResponse, err error) {
	// 调用RPC服务获取磁盘预测数据
	rpcReq := &cmpool.DiskPredictionReq{
		BeginTime:      req.StartDate,
		EndTime:        req.EndDate,
		Cluster:        req.Cluster,        // 添加集群参数支持
		DepartmentName: req.DepartmentName, // 添加部门参数支持
	}

	// 如果有IP参数，转换为IP列表
	if req.Ip != "" {
		// 支持逗号分隔的多个IP
		ips := strings.Split(req.Ip, ",")
		for i, ip := range ips {
			ips[i] = strings.TrimSpace(ip)
		}
		rpcReq.IpList = ips
		l.Logger.Infof("传递IP过滤条件: %s", req.Ip)
	}

	// 如果请求中有集群参数，则传递给RPC服务
	if req.Cluster != "" {
		l.Logger.Infof("传递集群过滤条件: %s", req.Cluster)
	}

	if req.Ip == "" && req.Cluster == "" {
		l.Logger.Info("未指定过滤条件，获取所有磁盘预测数据")
	}

	rpcResp, err := l.svcCtx.CmpoolRpc.GetDiskPrediction(l.ctx, rpcReq)
	if err != nil {
		l.Logger.Errorf("Failed to call RPC GetDiskPrediction: %v", err)
		return &types.DiskPredictionListResponse{
			Success: false,
			Message: "获取磁盘预测数据失败",
			List:    []types.DiskPrediction{},
		}, nil
	}

	if !rpcResp.Success {
		return &types.DiskPredictionListResponse{
			Success: false,
			Message: rpcResp.Message,
			List:    []types.DiskPrediction{},
		}, nil
	}

	// 转换RPC响应为API响应类型
	var diskPredictions []types.DiskPrediction
	for _, rpcPrediction := range rpcResp.DiskPrediction {
		diskPrediction := types.DiskPrediction{
			ID:                      int(rpcPrediction.Id),
			IP:                      rpcPrediction.Ip,
			CurrentDiskUsagePercent: rpcPrediction.CurrentDiskUsagePercent,
			TotalDisk:               rpcPrediction.TotalDisk,
			UsedDisk:                rpcPrediction.UsedDisk,
			DailyGrowthRate:         rpcPrediction.DailyGrowthRate,
			PredictedFullDate:       rpcPrediction.PredictedFullDate,
			DaysUntilFull:           rpcPrediction.DaysUntilFull,
			IsHighRisk:              rpcPrediction.IsHighRisk,
			CreateAt:                rpcPrediction.CreateAt,
			UpdateAt:                rpcPrediction.UpdateAt,
		}

		// 转换集群信息数组
		for _, cluster := range rpcPrediction.Clusters {
			diskPrediction.Clusters = append(diskPrediction.Clusters, types.HostClusterInfo{
				ClusterName:      cluster.ClusterName,
				ClusterGroupName: cluster.ClusterGroupName,
				DepartmentName:   cluster.DepartmentName,
			})
		}

		// 转换 IDC 信息
		if rpcPrediction.IdcInfo != nil {
			diskPrediction.IdcInfo = types.IdcInfo{
				ID:             int(rpcPrediction.IdcInfo.Id),
				IdcName:        rpcPrediction.IdcInfo.IdcName,
				IdcCode:        rpcPrediction.IdcInfo.IdcCode,
				IdcLocation:    rpcPrediction.IdcInfo.IdcLocation,
				IdcDescription: rpcPrediction.IdcInfo.IdcDescription,
			}
		}

		diskPredictions = append(diskPredictions, diskPrediction)
	}

	return &types.DiskPredictionListResponse{
		Success: true,
		Message: rpcResp.Message,
		List:    diskPredictions,
	}, nil
}
