package logic

import (
	"context"
	"database/sql"
	"fmt"
	"strings"

	"cmdb-rpc/cmpool"
	"cmdb-rpc/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
)

type AutoFetchHostInfoLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewAutoFetchHostInfoLogic(ctx context.Context, svcCtx *svc.ServiceContext) *AutoFetchHostInfoLogic {
	return &AutoFetchHostInfoLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 自动获取主机信息用于填充表单
func (l *AutoFetchHostInfoLogic) AutoFetchHostInfo(in *cmpool.AutoFetchHostInfoReq) (*cmpool.AutoFetchHostInfoResp, error) {
	logx.Infof("开始自动获取主机信息: %s", in.HostIp)

	if in.HostIp == "" {
		return &cmpool.AutoFetchHostInfoResp{
			Success: false,
			Message: "主机IP不能为空",
		}, nil
	}

	var hardwareInfo *cmpool.ManualHostHardwareInfo
	var applications []*cmpool.HostApplicationInfo
	var idcInfo *cmpool.IdcConf
	var messages []string

	// 1. 从CMDB获取硬件信息
	if in.FetchFromCmdb {
		hw, err := l.fetchHardwareInfoFromCMDB(in.HostIp)
		if err != nil {
			messages = append(messages, fmt.Sprintf("CMDB硬件信息获取失败: %v", err))
			logx.Errorf("从CMDB获取硬件信息失败: %v", err)
		} else if hw != nil {
			hardwareInfo = hw
			messages = append(messages, "成功从CMDB获取硬件信息")
		} else {
			messages = append(messages, "CMDB中未找到该主机的硬件信息")
		}
	}

	// 2. 从集群实例表获取应用信息
	if in.FetchFromClusters {
		apps, err := l.fetchApplicationInfoFromClusters(in.HostIp)
		if err != nil {
			messages = append(messages, fmt.Sprintf("集群应用信息获取失败: %v", err))
			logx.Errorf("从集群实例表获取应用信息失败: %v", err)
		} else if len(apps) > 0 {
			applications = apps
			messages = append(messages, fmt.Sprintf("成功从集群实例表获取 %d 个应用信息", len(apps)))
		} else {
			messages = append(messages, "集群实例表中未找到该主机的应用信息")
		}
	}

	// 3. 计算IDC信息（始终尝试计算，独立于硬件信息）
	idcConf, err := l.svcCtx.IdcConfModel.MatchIdcByIp(l.ctx, in.HostIp)
	if err != nil {
		if err != sql.ErrNoRows {
			logx.Errorf("匹配IDC信息失败: %v", err)
			messages = append(messages, fmt.Sprintf("IDC信息匹配失败: %v", err))
		} else {
			messages = append(messages, "未匹配到IDC机房信息")
		}
	} else {
		// 成功匹配到IDC信息，转换为protobuf格式
		idcInfo = &cmpool.IdcConf{
			Id:          int64(idcConf.Id),
			IdcName:     idcConf.IdcName,
			IdcCode:     idcConf.IdcCode,
			IdcIpRegexp: idcConf.IdcIpRegexp,
			IsActive:    idcConf.IsActive > 0,
			Priority:    int32(idcConf.Priority),
			CreatedAt:   idcConf.CreatedAt.Format("2006-01-02 15:04:05"),
			UpdatedAt:   idcConf.UpdatedAt.Format("2006-01-02 15:04:05"),
		}

		if idcConf.IdcLocation.Valid {
			idcInfo.IdcLocation = idcConf.IdcLocation.String
		}

		if idcConf.IdcDescription.Valid {
			idcInfo.IdcDescription = idcConf.IdcDescription.String
		}

		messages = append(messages, fmt.Sprintf("成功匹配到IDC机房: %s", idcConf.IdcName))
	}

	// 判断总体成功状态
	success := hardwareInfo != nil || len(applications) > 0 || idcInfo != nil
	message := "自动获取完成"
	if len(messages) > 0 {
		message = fmt.Sprintf("%s: %s", message, strings.Join(messages, "; "))
	}

	return &cmpool.AutoFetchHostInfoResp{
		Success:      success,
		Message:      message,
		HardwareInfo: hardwareInfo,
		Applications: applications,
		IdcInfo:      idcInfo, // IDC信息作为独立字段返回
	}, nil
}

// fetchHardwareInfoFromCMDB 从外部CMDB获取硬件信息
func (l *AutoFetchHostInfoLogic) fetchHardwareInfoFromCMDB(hostIp string) (*cmpool.ManualHostHardwareInfo, error) {
	// 使用现有的硬件信息获取逻辑
	hardwareLogic := NewFetchHostsHardwareInfoLogic(l.ctx, l.svcCtx)
	resp, err := hardwareLogic.FetchHostsHardwareInfo(&cmpool.FetchHostsHardwareInfoReq{
		HostIpList: []string{hostIp},
	})

	if err != nil {
		return nil, fmt.Errorf("调用硬件信息获取接口失败: %v", err)
	}

	if !resp.Success {
		return nil, fmt.Errorf("硬件信息获取失败: %s", resp.Message)
	}

	// 检查是否成功获取到硬件信息
	if len(resp.HardwareInfoList) > 0 {
		hardwareInfo := resp.HardwareInfoList[0]
		if !hardwareInfo.Success {
			return nil, fmt.Errorf("获取硬件信息失败: %s", hardwareInfo.Message)
		}

		// 转换为ManualHostHardwareInfo格式
		return &cmpool.ManualHostHardwareInfo{
			DiskSize:        int32(hardwareInfo.Disk),
			Ram:             int32(hardwareInfo.Ram),
			Vcpus:           int32(hardwareInfo.Vcpus),
			HostType:        "cmdb", // 从CMDB获取的标记为cmdb类型
			H3CId:           "",
			H3CStatus:       "",
			IfH3CSync:       "",
			H3CImgId:        "",
			H3CHmName:       hardwareInfo.HostName,
			LeafNumber:      "",
			RackNumber:      "",
			RackHeight:      0,
			RackStartNumber: 0,
			FromFactor:      0,
			SerialNumber:    "",
		}, nil
	}

	return nil, nil
}

// fetchApplicationInfoFromClusters 从集群实例表获取应用信息
func (l *AutoFetchHostInfoLogic) fetchApplicationInfoFromClusters(hostIp string) ([]*cmpool.HostApplicationInfo, error) {
	var applications []*cmpool.HostApplicationInfo

	// 从 MySQL 集群实例表查询
	mysqlClusters, err := l.svcCtx.MysqlClusterInstanceModel.FindByHostIp(l.ctx, hostIp)
	if err == nil && len(mysqlClusters) > 0 {
		for _, cluster := range mysqlClusters {
			// 通过集群名称查找业务线信息
			departmentName, _ := l.getDepartmentByCluster(cluster.ClusterName)

			applications = append(applications, &cmpool.HostApplicationInfo{
				ServerType:     "mysql",
				ServerVersion:  "",
				ServerSubtitle: "",
				ClusterName:    cluster.ClusterName,
				ServerProtocol: "mysql",
				ServerAddr:     cluster.Ip,
				ServerPort:     int32(cluster.Port),
				ServerRole:     cluster.InstanceRole,
				ServerStatus:   "running",
				DepartmentName: departmentName,
			})
		}
	}

	// 从 MSSQL 集群实例表查询
	mssqlClusters, err := l.svcCtx.MssqlClusterInstanceModel.FindByHostIp(l.ctx, hostIp)
	if err == nil && len(mssqlClusters) > 0 {
		for _, cluster := range mssqlClusters {
			// 通过集群名称查找业务线信息
			departmentName, _ := l.getDepartmentByCluster(cluster.ClusterName)

			applications = append(applications, &cmpool.HostApplicationInfo{
				ServerType:     "mssql",
				ServerVersion:  "",
				ServerSubtitle: "",
				ClusterName:    cluster.ClusterName,
				ServerProtocol: "mssql",
				ServerAddr:     cluster.Ip,
				ServerPort:     int32(cluster.InstancePort),
				ServerRole:     cluster.InstanceRole,
				ServerStatus:   "running",
				DepartmentName: departmentName,
			})
		}
	}

	// TODO: 添加TiDB和GoldenDB的FindByHostIp方法后再启用
	logx.Infof("主机 %s 从集群实例表获取到 %d 个应用信息", hostIp, len(applications))

	return applications, nil
}

// getDepartmentByCluster 通过集群名称获取部门业务线信息
func (l *AutoFetchHostInfoLogic) getDepartmentByCluster(clusterName string) (string, error) {
	// 从cluster_groups表查询业务线信息

	cg, err := l.svcCtx.ClusterGroupsModel.FindByClusterName(l.ctx, clusterName)
	if err != nil {
		return "", err
	}

	return cg.DepartmentLineName, nil
}
