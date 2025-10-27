package logic

import (
	"context"
	"fmt"
	"sort"
	"time"

	"cmdb-rpc/cmpool"
	"cmdb-rpc/internal/model"
	"cmdb-rpc/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetServerResourceLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGetServerResourceLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetServerResourceLogic {
	return &GetServerResourceLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 查询主机资源使用率数据
func (l *GetServerResourceLogic) GetServerResource(in *cmpool.ServerResourceReq) (*cmpool.ServerResourceResp, error) {
	l.Logger.Infof("查询服务器资源数据，时间范围: %s - %s, IP列表: %v, 集群: %s", in.BeginTime, in.EndTime, in.IpList, in.Cluster)

	// 从数据库查询服务器资源数据
	serverResources, err := l.getServerResourcesFromDB(in)
	if err != nil {
		l.Logger.Errorf("查询服务器资源数据失败: %v", err)
		return &cmpool.ServerResourceResp{
			Success: false,
			Message: fmt.Sprintf("查询服务器资源数据失败: %v", err),
		}, nil
	}

	l.Logger.Infof("查询到%d条服务器资源数据", len(serverResources))
	return &cmpool.ServerResourceResp{
		Success:        true,
		Message:        "查询成功",
		ServerResource: serverResources,
	}, nil
}

// getServerResourcesFromDB 从数据库查询服务器资源数据
func (l *GetServerResourceLogic) getServerResourcesFromDB(req *cmpool.ServerResourceReq) ([]*cmpool.ServerResource, error) {
	var rows []*model.ServerResourceRow
	var err error

	// 根据参数选择不同的查询方法
	if req.Cluster != "" || req.IpList == nil || len(req.IpList) == 0 {
		// 使用集群名称查询
		rows, err = l.svcCtx.ServerResourcesModel.FindServerResourcesWithClusterFilter(l.ctx, req.BeginTime, req.EndTime, req.Cluster)
	} else {
		// 使用IP列表查询（原有逻辑）
		rows, err = l.svcCtx.ServerResourcesModel.FindServerResourcesWithFilter(l.ctx, req.BeginTime, req.EndTime, req.IpList)
	}

	if err != nil {
		return nil, fmt.Errorf("执行查询失败: %w", err)
	}

	// 对数据进行排序，确保数据的一致性
	// 按照 cluster_name、ip、host_name、mon_date、id 的顺序排序
	l.sortServerResourceRows(rows)

	// 将数据库结果转换为protobuf结构
	var resources []*cmpool.ServerResource
	for i, row := range rows {
		resource := &cmpool.ServerResource{
			Id:             row.Id,
			CreateAt:       row.CreateTime,
			UpdateAt:       row.UpdateTime,
			Datetime:       row.MonDate,
			SequenceNumber: int64(i + 1), // 添加序列号，从1开始
		}

		// 处理可为空的字段
		if row.PoolId.Valid {
			resource.PoolId = row.PoolId.Int64
		}
		if row.Ip.Valid {
			resource.Ip = row.Ip.String
		}
		if row.TotalMemory.Valid {
			resource.TotalMemory = float32(row.TotalMemory.Float64)
		}
		if row.UsedMemory.Valid {
			resource.UsedMemory = float32(row.UsedMemory.Float64)
		}
		if row.TotalDisk.Valid {
			resource.TotalDisk = float32(row.TotalDisk.Float64)
		}
		if row.UsedDisk.Valid {
			resource.UsedDisk = float32(row.UsedDisk.Float64)
		}
		if row.CPUCores.Valid {
			resource.CPUCores = row.CPUCores.Int32
		}
		if row.CPULoad.Valid {
			resource.CPULoad = float32(row.CPULoad.Float64)
		}

		// 百分比字段映射
		if row.CpuPercentMax.Valid {
			resource.CpuPercentMax = row.CpuPercentMax.Float64
		}
		if row.CpuPercentAvg.Valid {
			resource.CpuPercentAvg = row.CpuPercentAvg.Float64
		}
		if row.CpuPercentMin.Valid {
			resource.CpuPercentMin = row.CpuPercentMin.Float64
		}
		if row.MemPercentMax.Valid {
			resource.MemPercentMax = row.MemPercentMax.Float64
		}
		if row.MemPercentAvg.Valid {
			resource.MemPercentAvg = row.MemPercentAvg.Float64
		}
		if row.MemPercentMin.Valid {
			resource.MemPercentMin = row.MemPercentMin.Float64
		}
		if row.DiskPercentMax.Valid {
			resource.DiskPercentMax = row.DiskPercentMax.Float64
		}
		if row.DiskPercentAvg.Valid {
			resource.DiskPercentAvg = row.DiskPercentAvg.Float64
		}
		if row.DiskPercentMin.Valid {
			resource.DiskPercentMin = row.DiskPercentMin.Float64
		}

		// 4. 查询该主机所属的所有集群
		if row.Ip.Valid {
			clusters, err := l.svcCtx.HostsApplicationsModel.FindClustersByServerAddr(l.ctx, row.Ip.String)
			if err == nil {
				for _, cluster := range clusters {
					hostCluster := &cmpool.HostClusterInfo{
						ClusterName:      cluster.ClusterName,
						ClusterGroupName: cluster.ClusterGroupName,
					}
					resource.Clusters = append(resource.Clusters, hostCluster)
				}
			}
		}

		// 5. 查询该主机的IDC信息
		if row.Ip.Valid {
			idcConf, err := l.svcCtx.IdcConfModel.MatchIdcByIp(l.ctx, row.Ip.String)
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

		resources = append(resources, resource)
	}

	// 如果没有查询到数据，生成一些示例数据用于测试
	//if len(resources) == 0 {
	//	l.Logger.Info("数据库中暂无监控数据，生成示例数据")
	//	resources = l.generateSampleData(req.IpList)
	//}

	return resources, nil
}

// sortServerResourceRows 对服务器资源数据进行排序
// 按照 cluster_name、ip、host_name、mon_date、id 的顺序排序，确保数据的一致性
func (l *GetServerResourceLogic) sortServerResourceRows(rows []*model.ServerResourceRow) {
	if len(rows) <= 1 {
		return
	}

	// 使用Go的sort包进行排序
	sort.Slice(rows, func(i, j int) bool {
		// 首先按集群名称排序
		clusterNameI := ""
		clusterNameJ := ""
		if rows[i].ClusterName.Valid {
			clusterNameI = rows[i].ClusterName.String
		}
		if rows[j].ClusterName.Valid {
			clusterNameJ = rows[j].ClusterName.String
		}

		if clusterNameI != clusterNameJ {
			return clusterNameI < clusterNameJ
		}

		// 然后按IP排序
		ipI := ""
		ipJ := ""
		if rows[i].Ip.Valid {
			ipI = rows[i].Ip.String
		}
		if rows[j].Ip.Valid {
			ipJ = rows[j].Ip.String
		}

		if ipI != ipJ {
			return ipI < ipJ
		}

		// IP已经排序过了，这里可以跳过主机名排序
		// 或者如果有其他需要排序的字段可以在这里添加

		// 然后按时间排序
		if rows[i].MonDate != rows[j].MonDate {
			return rows[i].MonDate < rows[j].MonDate
		}

		// 最后按ID排序
		return rows[i].Id < rows[j].Id
	})
}

// generateSampleData 生成示例监控数据
func (l *GetServerResourceLogic) generateSampleData(ipList []string) []*cmpool.ServerResource {
	if len(ipList) == 0 {
		ipList = []string{"192.168.1.10", "192.168.1.11", "192.168.1.12"}
	}

	now := time.Now()
	var resources []*cmpool.ServerResource

	for i, ip := range ipList {
		resource := &cmpool.ServerResource{
			Id:          int64(i + 1),
			CreateAt:    now.Format("2006-01-02T15:04:05Z"),
			UpdateAt:    now.Format("2006-01-02T15:04:05Z"),
			PoolId:      int64(i + 1),
			Ip:          ip,
			TotalMemory: 64.0,
			UsedMemory:  32.5,
			TotalDisk:   500.0,
			UsedDisk:    250.2,
			CPUCores:    8,
			CPULoad:     45.6,
			Datetime:    now.Format("2006-01-02T15:04:05Z"),
		}
		resources = append(resources, resource)
	}

	return resources
}
