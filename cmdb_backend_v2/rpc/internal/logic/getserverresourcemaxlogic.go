package logic

import (
	"context"
	"fmt"
	"strings"

	"cmdb-rpc/cmpool"
	"cmdb-rpc/internal/model"
	"cmdb-rpc/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetServerResourceMaxLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGetServerResourceMaxLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetServerResourceMaxLogic {
	return &GetServerResourceMaxLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 查询主机资源最大利用率数据（支持集群数组）
func (l *GetServerResourceMaxLogic) GetServerResourceMax(in *cmpool.ServerResourceMaxReq) (*cmpool.ServerResourceMaxResp, error) {
	// 1. 构建查询条件
	whereClause := "1=1"
	var args []interface{}

	// 添加时间范围查询条件
	if in.BeginTime != "" && in.EndTime != "" {
		whereClause += " AND mon_date BETWEEN ? AND ?"
		args = append(args, in.BeginTime, in.EndTime)
	}

	// 添加IP列表查询条件
	if len(in.IpList) > 0 {
		placeholders := strings.Repeat("?,", len(in.IpList))
		placeholders = placeholders[:len(placeholders)-1]
		whereClause += " AND ip IN (" + placeholders + ")"
		for _, ip := range in.IpList {
			args = append(args, ip)
		}
	}

	// 添加集群查询条件
	if in.Cluster != "" {
		whereClause += " AND cluster_name = ?"
		args = append(args, in.Cluster)
	}

	// 移除未使用的查询代码，现在使用 model 层方法

	// 使用 model 层的方法查询主机资源最大值数据
	var serverResults []*model.ServerResourceMaxData
	var err error

	serverResults, err = l.svcCtx.ServerResourcesModel.FindServerResourceMax(l.ctx, in.BeginTime, in.EndTime, in.IpList, in.Cluster)

	if err != nil {
		logx.Errorf("查询主机资源最大值数据失败: %v", err)
		return &cmpool.ServerResourceMaxResp{
			Success: false,
			Message: fmt.Sprintf("查询主机资源最大值数据失败: %v", err),
		}, nil
	}

	// 3. 构建响应数据（按IP分组处理集群信息）
	resourceMap := make(map[string]*cmpool.ServerResourceMax)
	var serverResourceList []*cmpool.ServerResourceMax

	for _, result := range serverResults {
		// 获取IP字符串，用于映射键
		ip := result.Ip

		// 如果是该IP的第一条记录，创建资源对象
		resource, exists := resourceMap[ip]
		if !exists {
			resource = &cmpool.ServerResourceMax{
				Id:            result.Id,
				CreateAt:      result.CreateTime,
				UpdateAt:      result.UpdateTime,
				PoolId:        result.PoolId,
				Ip:            ip,
				HostName:      result.HostName,
				TotalMemory:   float32(result.TotalMemory),
				MaxUsedMemory: float32(result.MaxUsedMemory),
				TotalDisk:     float32(result.TotalDisk),
				MaxUsedDisk:   float32(result.MaxUsedDisk),
				CPUCores:      result.CPUCores,
				MaxCPULoad:    float32(result.MaxCPULoad),
				MaxDateTime:   result.MaxDatetime,
				HostType:      result.HostType,
				// 百分比字段映射
				CpuPercentMax:  result.CpuPercentMax,
				CpuPercentAvg:  result.CpuPercentAvg,
				CpuPercentMin:  result.CpuPercentMin,
				MemPercentMax:  result.MemPercentMax,
				MemPercentAvg:  result.MemPercentAvg,
				MemPercentMin:  result.MemPercentMin,
				DiskPercentMax: result.DiskPercentMax,
				DiskPercentAvg: result.DiskPercentAvg,
				DiskPercentMin: result.DiskPercentMin,
			}

			// 查询该主机的IDC信息
			if ip != "" {
				idcConf, err := l.svcCtx.IdcConfModel.MatchIdcByIp(l.ctx, ip)
				if err == nil {
					// 处理可空字段
					location := ""
					if idcConf.IdcLocation.Valid {
						location = idcConf.IdcLocation.String
					}
					description := ""
					if idcConf.IdcDescription.Valid {
						description = idcConf.IdcDescription.String
					}

					resource.IdcInfo = &cmpool.IdcConf{
						Id:             int64(idcConf.Id),
						IdcName:        idcConf.IdcName,
						IdcCode:        idcConf.IdcCode,
						IdcLocation:    location,
						IdcDescription: description,
						IdcIpRegexp:    idcConf.IdcIpRegexp,
						Priority:       int32(idcConf.Priority),
						IsActive:       idcConf.IsActive == 1,
						CreatedAt:      idcConf.CreatedAt.Format("2006-01-02 15:04:05"),
						UpdatedAt:      idcConf.UpdatedAt.Format("2006-01-02 15:04:05"),
					}
				}
			}

			resourceMap[ip] = resource
			serverResourceList = append(serverResourceList, resource)
		} else {
			// 对于已存在的IP，更新资源信息为最大值
			if result.TotalMemory > float64(resource.TotalMemory) {
				resource.TotalMemory = float32(result.TotalMemory)
			}
			if result.MaxUsedMemory > float64(resource.MaxUsedMemory) {
				resource.MaxUsedMemory = float32(result.MaxUsedMemory)
			}
			if result.TotalDisk > float64(resource.TotalDisk) {
				resource.TotalDisk = float32(result.TotalDisk)
			}
			if result.MaxUsedDisk > float64(resource.MaxUsedDisk) {
				resource.MaxUsedDisk = float32(result.MaxUsedDisk)
			}
			if result.CPUCores > resource.CPUCores {
				resource.CPUCores = result.CPUCores
			}
			if result.MaxCPULoad > float64(resource.MaxCPULoad) {
				resource.MaxCPULoad = float32(result.MaxCPULoad)
			}
			if result.MaxDatetime > resource.MaxDateTime {
				resource.MaxDateTime = result.MaxDatetime
			}

			// 更新百分比字段（取最大值）
			if result.CpuPercentMax > resource.CpuPercentMax {
				resource.CpuPercentMax = result.CpuPercentMax
			}
			if result.CpuPercentAvg > resource.CpuPercentAvg {
				resource.CpuPercentAvg = result.CpuPercentAvg
			}
			if result.CpuPercentMin > resource.CpuPercentMin {
				resource.CpuPercentMin = result.CpuPercentMin
			}
			if result.MemPercentMax > resource.MemPercentMax {
				resource.MemPercentMax = result.MemPercentMax
			}
			if result.MemPercentAvg > resource.MemPercentAvg {
				resource.MemPercentAvg = result.MemPercentAvg
			}
			if result.MemPercentMin > resource.MemPercentMin {
				resource.MemPercentMin = result.MemPercentMin
			}
			if result.DiskPercentMax > resource.DiskPercentMax {
				resource.DiskPercentMax = result.DiskPercentMax
			}
			if result.DiskPercentAvg > resource.DiskPercentAvg {
				resource.DiskPercentAvg = result.DiskPercentAvg
			}
			if result.DiskPercentMin > resource.DiskPercentMin {
				resource.DiskPercentMin = result.DiskPercentMin
			}
		}

		// 添加集群信息（避免重复）
		if result.ClusterName != "" && result.ClusterName != "未分配集群" {
			// 检查是否已存在该集群
			exists := false
			for _, cluster := range resource.Clusters {
				if cluster.ClusterName == result.ClusterName && cluster.ClusterGroupName == result.GroupName {
					exists = true
					break
				}
			}
			if !exists {
				hostCluster := &cmpool.HostClusterInfo{
					ClusterName:      result.ClusterName,
					ClusterGroupName: result.GroupName,
					DepartmentName:   result.DepartmentName,
				}
				resource.Clusters = append(resource.Clusters, hostCluster)
			}
		}
	}

	return &cmpool.ServerResourceMaxResp{
		Success:           true,
		Message:           "查询成功",
		ServerResourceMax: serverResourceList,
	}, nil
}
