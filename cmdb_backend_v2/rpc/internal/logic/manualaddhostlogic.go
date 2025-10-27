package logic

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"cmdb-rpc/cmpool"
	"cmdb-rpc/internal/model"
	"cmdb-rpc/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
)

type ManualAddHostLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewManualAddHostLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ManualAddHostLogic {
	return &ManualAddHostLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 手动添加主机到hosts_pool表并同步相关信息
func (l *ManualAddHostLogic) ManualAddHost(in *cmpool.ManualAddHostReq) (*cmpool.ManualAddHostResp, error) {
	logx.Infof("开始手动添加主机: %s", in.HostIp)

	// 1. 验证IP格式
	if in.HostIp == "" {
		return &cmpool.ManualAddHostResp{
			Success: false,
			Message: "主机IP不能为空",
		}, nil
	}

	// 2. 添加主机到 hosts_pool 表
	hostName := in.HostName
	if hostName == "" {
		hostName = in.HostIp // 如果没有提供主机名，使用IP作为主机名
	}

	hostId, err := l.svcCtx.HostsPoolModel.InsertIfNotExists(l.ctx, hostName, in.HostIp, "manual")
	if err != nil {
		logx.Errorf("添加主机到hosts_pool失败: %v", err)
		return &cmpool.ManualAddHostResp{
			Success: false,
			Message: fmt.Sprintf("添加主机失败: %v", err),
		}, nil
	}

	logx.Infof("主机 %s 已添加到hosts_pool，ID: %d", in.HostIp, hostId)

	// 3. 更新主机的IDC信息
	if in.IdcId > 0 {
		err = l.svcCtx.HostsPoolModel.UpdateHostIdcInfo(l.ctx, in.HostIp, in.IdcId)
		if err != nil {
			logx.Errorf("更新主机IDC信息失败: %v", err)
			// IDC信息更新失败不应阻止主机添加，只记录错误
			logx.Infof("主机 %s 的IDC信息更新失败，但主机已成功添加", in.HostIp)
		} else {
			logx.Infof("主机 %s 的IDC信息已更新，IDC ID: %d", in.HostIp, in.IdcId)
		}
	}

	// 4. 更新主机的硬件信息
	if in.HardwareInfo != nil {
		err = l.updateHostHardwareInfo(hostId, in.HardwareInfo)
		if err != nil {
			logx.Errorf("更新主机硬件信息失败: %v", err)
			return &cmpool.ManualAddHostResp{
				Success: false,
				Message: fmt.Sprintf("更新硬件信息失败: %v", err),
			}, nil
		}
	}

	// 5. 添加应用信息
	if len(in.Applications) > 0 {
		err = l.addHostApplications(hostId, in.Applications)
		if err != nil {
			logx.Errorf("添加应用信息失败: %v", err)
			return &cmpool.ManualAddHostResp{
				Success: false,
				Message: fmt.Sprintf("添加应用信息失败: %v", err),
			}, nil
		}
	}

	// 6. 如果需要自动获取信息
	var autoFetchResult *cmpool.AutoFetchResult
	if in.AutoFetchFromCmdb || in.AutoFetchFromClusters {
		fetchReq := &cmpool.AutoFetchHostInfoReq{
			HostIp:            in.HostIp,
			FetchFromCmdb:     in.AutoFetchFromCmdb,
			FetchFromClusters: in.AutoFetchFromClusters,
		}

		autoFetchLogic := NewAutoFetchHostInfoLogic(l.ctx, l.svcCtx)
		fetchResp, err := autoFetchLogic.AutoFetchHostInfo(fetchReq)
		if err != nil {
			logx.Errorf("自动获取主机信息失败: %v", err)
		} else if fetchResp.Success {
			autoFetchResult = &cmpool.AutoFetchResult{
				Success:      true,
				Message:      fetchResp.Message,
				HardwareInfo: fetchResp.HardwareInfo,
				Applications: fetchResp.Applications,
			}
		}
	}

	return &cmpool.ManualAddHostResp{
		Success:         true,
		Message:         "主机添加成功",
		AutoFetchResult: autoFetchResult,
	}, nil
}

// updateHostHardwareInfo 更新主机硬件信息
func (l *ManualAddHostLogic) updateHostHardwareInfo(hostId int64, hardwareInfo *cmpool.ManualHostHardwareInfo) error {
	// 1. 根据hostId查找主机记录，获取主机IP
	existingHost, err := l.svcCtx.HostsPoolModel.FindOne(l.ctx, uint64(hostId))
	if err != nil {
		return fmt.Errorf("查找主机记录失败: %v", err)
	}

	// 2. 构建要更新的主机信息，只设置有值的字段
	updateHost := &model.HostsPool{
		HostIp: existingHost.HostIp, // 必填字段，用于定位记录
	}

	// 更新硬件信息（只有值大于0时才更新）
	if hardwareInfo.DiskSize > 0 {
		updateHost.DiskSize = sql.NullInt64{Int64: int64(hardwareInfo.DiskSize), Valid: true}
	}
	if hardwareInfo.Ram > 0 {
		updateHost.Ram = sql.NullInt64{Int64: int64(hardwareInfo.Ram), Valid: true}
	}
	if hardwareInfo.Vcpus > 0 {
		updateHost.Vcpus = sql.NullInt64{Int64: int64(hardwareInfo.Vcpus), Valid: true}
	}

	// 更新主机类型（如果提供且不为空）
	if hardwareInfo.HostType != "" {
		updateHost.HostType = sql.NullString{String: hardwareInfo.HostType, Valid: true}
	}

	// 更新H3C相关信息（如果提供且不为空）
	if hardwareInfo.H3CId != "" {
		updateHost.H3cId = sql.NullString{String: hardwareInfo.H3CId, Valid: true}
	}
	if hardwareInfo.H3CStatus != "" {
		updateHost.H3cStatus = sql.NullString{String: hardwareInfo.H3CStatus, Valid: true}
	}
	if hardwareInfo.IfH3CSync != "" {
		updateHost.IfH3cSync = sql.NullString{String: hardwareInfo.IfH3CSync, Valid: true}
	}
	if hardwareInfo.H3CImgId != "" {
		updateHost.H3cImgId = sql.NullString{String: hardwareInfo.H3CImgId, Valid: true}
	}
	if hardwareInfo.H3CHmName != "" {
		updateHost.H3cHmName = sql.NullString{String: hardwareInfo.H3CHmName, Valid: true}
	}

	// 更新机架信息（如果提供且不为空）
	if hardwareInfo.LeafNumber != "" {
		updateHost.LeafNumber = sql.NullString{String: hardwareInfo.LeafNumber, Valid: true}
	}
	if hardwareInfo.RackNumber != "" {
		updateHost.RackNumber = sql.NullString{String: hardwareInfo.RackNumber, Valid: true}
	}
	if hardwareInfo.RackHeight > 0 {
		updateHost.RackHeight = sql.NullInt64{Int64: int64(hardwareInfo.RackHeight), Valid: true}
	}
	if hardwareInfo.RackStartNumber >= 0 { // 允许0值，表示从机架底部开始
		updateHost.RackStartNumber = sql.NullInt64{Int64: int64(hardwareInfo.RackStartNumber), Valid: true}
	}
	if hardwareInfo.FromFactor > 0 {
		updateHost.FromFactor = sql.NullInt64{Int64: int64(hardwareInfo.FromFactor), Valid: true}
	}
	if hardwareInfo.SerialNumber != "" {
		updateHost.SerialNumber = sql.NullString{String: hardwareInfo.SerialNumber, Valid: true}
	}

	// 3. 调用UpdateHostHardwareInfo进行更新
	err = l.svcCtx.HostsPoolModel.UpdateHostHardwareInfo(l.ctx, updateHost)
	if err != nil {
		return fmt.Errorf("更新主机硬件信息失败: %v", err)
	}

	logx.Infof("成功更新主机 %s 的硬件信息", existingHost.HostIp)
	return nil
}

// addHostApplications 批量添加主机应用信息
func (l *ManualAddHostLogic) addHostApplications(hostId int64, applications []*cmpool.HostApplicationInfo) error {
	for _, app := range applications {
		// 检查是否已存在相同的应用记录
		exists, err := l.svcCtx.HostsApplicationsModel.ExistsByHostIdAndCluster(l.ctx, hostId, app.ClusterName)
		if err != nil {
			return fmt.Errorf("检查应用记录失败: %v", err)
		}

		if exists {
			logx.Infof("主机ID %d 的集群 %s 应用记录已存在，跳过添加", hostId, app.ClusterName)
			continue
		}

		// 添加新的应用记录
		_, err = l.svcCtx.HostsApplicationsModel.Insert(l.ctx, &model.HostsApplications{
			PoolId:         uint64(hostId),
			ServerType:     sql.NullString{String: app.ServerType, Valid: app.ServerType != ""},
			ServerVersion:  sql.NullString{String: app.ServerVersion, Valid: app.ServerVersion != ""},
			ServerSubtitle: sql.NullString{String: app.ServerSubtitle, Valid: app.ServerSubtitle != ""},
			ClusterName:    sql.NullString{String: app.ClusterName, Valid: app.ClusterName != ""},
			ServerProtocol: sql.NullString{String: app.ServerProtocol, Valid: app.ServerProtocol != ""},
			ServerAddr:     sql.NullString{String: app.ServerAddr, Valid: app.ServerAddr != ""},
			ServerPort:     int64(app.ServerPort),
			ServerRole:     sql.NullString{String: app.ServerRole, Valid: app.ServerRole != ""},
			ServerStatus:   sql.NullString{String: app.ServerStatus, Valid: app.ServerStatus != ""},
			DepartmentName: sql.NullString{String: app.DepartmentName, Valid: app.DepartmentName != ""},
			CreateTime:     time.Now(),
			UpdateTime:     time.Now(),
		})

		if err != nil {
			return fmt.Errorf("插入应用记录失败: %v", err)
		}
	}

	return nil
}

// getDepartmentByCluster 通过集群名称获取部门业务线信息
func (l *ManualAddHostLogic) getDepartmentByCluster(clusterName string) (string, error) {
	// 使用 ClusterGroupsModel 查询集群组信息
	clusterGroup, err := l.svcCtx.ClusterGroupsModel.FindByClusterName(l.ctx, clusterName)
	if err != nil {
		return "", err
	}

	return clusterGroup.DepartmentLineName, nil
}
