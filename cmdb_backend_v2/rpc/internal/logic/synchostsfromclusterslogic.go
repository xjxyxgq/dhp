package logic

import (
	"cmdb-rpc/cmpool"
	"cmdb-rpc/internal/model"
	"cmdb-rpc/internal/svc"
	"context"
	"database/sql"
	"fmt"

	"github.com/zeromicro/go-zero/core/logx"
)

type SyncHostsFromClustersLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewSyncHostsFromClustersLogic(ctx context.Context, svcCtx *svc.ServiceContext) *SyncHostsFromClustersLogic {
	return &SyncHostsFromClustersLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 从集群实例表同步主机到hosts_pool和hosts_applications表
func (l *SyncHostsFromClustersLogic) SyncHostsFromClusters(in *cmpool.SyncClusterGroupsReq) (*cmpool.SyncClusterGroupsResp, error) {
	l.Logger.Info("开始从集群实例表同步主机数据")

	var details []*cmpool.DatabaseSyncDetail
	totalSyncedCount := 0

	// 1. 同步MySQL实例主机
	mysqlDetail, err := l.syncMysqlInstanceHosts()
	if err != nil {
		l.Logger.Errorf("同步MySQL实例主机失败: %v", err)
		return &cmpool.SyncClusterGroupsResp{
			Success:     false,
			Message:     fmt.Sprintf("同步MySQL实例主机失败: %v", err),
			SyncedCount: int32(totalSyncedCount),
			Details:     details,
		}, nil
	}
	details = append(details, mysqlDetail)
	totalSyncedCount += int(mysqlDetail.SyncedCount)

	// 2. 同步MSSQL实例主机
	mssqlDetail, err := l.syncMssqlInstanceHosts()
	if err != nil {
		l.Logger.Errorf("同步MSSQL实例主机失败: %v", err)
		return &cmpool.SyncClusterGroupsResp{
			Success:     false,
			Message:     fmt.Sprintf("同步MSSQL实例主机失败: %v", err),
			SyncedCount: int32(totalSyncedCount),
			Details:     details,
		}, nil
	}
	details = append(details, mssqlDetail)
	totalSyncedCount += int(mssqlDetail.SyncedCount)

	// 3. 同步TiDB实例主机
	tidbDetail, err := l.syncTidbInstanceHosts()
	if err != nil {
		l.Logger.Errorf("同步TiDB实例主机失败: %v", err)
		return &cmpool.SyncClusterGroupsResp{
			Success:     false,
			Message:     fmt.Sprintf("同步TiDB实例主机失败: %v", err),
			SyncedCount: int32(totalSyncedCount),
			Details:     details,
		}, nil
	}
	details = append(details, tidbDetail)
	totalSyncedCount += int(tidbDetail.SyncedCount)

	// 4. 同步GoldenDB实例主机
	goldendbDetail, err := l.syncGoldendbInstanceHosts()
	if err != nil {
		l.Logger.Errorf("同步GoldenDB实例主机失败: %v", err)
		return &cmpool.SyncClusterGroupsResp{
			Success:     false,
			Message:     fmt.Sprintf("同步GoldenDB实例主机失败: %v", err),
			SyncedCount: int32(totalSyncedCount),
			Details:     details,
		}, nil
	}
	details = append(details, goldendbDetail)
	totalSyncedCount += int(goldendbDetail.SyncedCount)

	// 记录汇总日志
	l.Logger.Infof("主机数据同步完成 - 总计同步: %d 个主机", totalSyncedCount)
	for _, detail := range details {
		if detail.SyncedCount > 0 {
			l.Logger.Infof("  - %s: %d 个主机", detail.DatabaseType, detail.SyncedCount)
		}
	}

	return &cmpool.SyncClusterGroupsResp{
		Success:     true,
		Message:     fmt.Sprintf("主机同步成功，共同步 %d 个主机", totalSyncedCount),
		SyncedCount: int32(totalSyncedCount),
		Details:     details,
	}, nil
}

// syncMysqlInstanceHosts 同步MySQL实例主机数据
func (l *SyncHostsFromClustersLogic) syncMysqlInstanceHosts() (*cmpool.DatabaseSyncDetail, error) {
	l.Logger.Infof("开始同步MySQL实例主机")

	// 调用Model层方法查询实例
	instances, err := l.svcCtx.MysqlClusterInstanceModel.FindInstancesWithCluster(l.ctx)
	if err != nil {
		l.Logger.Errorf("查询MySQL实例失败: %v", err)
		return &cmpool.DatabaseSyncDetail{
			DatabaseType:  "mysql",
			SyncedCount:   0,
			ClusterGroups: []string{},
		}, fmt.Errorf("查询MySQL实例失败: %w", err)
	}

	l.Logger.Infof("查询到 %d 个MySQL实例", len(instances))

	// 同步实例到主机表
	syncedCount, err := l.syncInstancesToHosts(instances, "mysql")
	if err != nil {
		l.Logger.Errorf("同步MySQL实例到主机表失败: %v", err)
		return &cmpool.DatabaseSyncDetail{
			DatabaseType:  "mysql",
			SyncedCount:   0,
			ClusterGroups: []string{},
		}, err
	}

	l.Logger.Infof("MySQL实例主机同步完成: %d 个主机", syncedCount)

	return &cmpool.DatabaseSyncDetail{
		DatabaseType:  "mysql",
		SyncedCount:   int32(syncedCount),
		ClusterGroups: []string{},
	}, nil
}

// syncMssqlInstanceHosts 同步MSSQL实例主机数据
func (l *SyncHostsFromClustersLogic) syncMssqlInstanceHosts() (*cmpool.DatabaseSyncDetail, error) {
	l.Logger.Infof("开始同步MSSQL实例主机")

	// 调用Model层方法查询实例
	instances, err := l.svcCtx.MssqlClusterInstanceModel.FindInstancesWithCluster(l.ctx)
	if err != nil {
		l.Logger.Errorf("查询MSSQL实例失败: %v", err)
		return &cmpool.DatabaseSyncDetail{
			DatabaseType:  "mssql",
			SyncedCount:   0,
			ClusterGroups: []string{},
		}, fmt.Errorf("查询MSSQL实例失败: %w", err)
	}

	l.Logger.Infof("查询到 %d 个MSSQL实例", len(instances))

	// 同步实例到主机表
	syncedCount, err := l.syncInstancesToHosts(instances, "mssql")
	if err != nil {
		l.Logger.Errorf("同步MSSQL实例到主机表失败: %v", err)
		return &cmpool.DatabaseSyncDetail{
			DatabaseType:  "mssql",
			SyncedCount:   0,
			ClusterGroups: []string{},
		}, err
	}

	l.Logger.Infof("MSSQL实例主机同步完成: %d 个主机", syncedCount)

	return &cmpool.DatabaseSyncDetail{
		DatabaseType:  "mssql",
		SyncedCount:   int32(syncedCount),
		ClusterGroups: []string{},
	}, nil
}

// syncTidbInstanceHosts 同步TiDB实例主机数据
func (l *SyncHostsFromClustersLogic) syncTidbInstanceHosts() (*cmpool.DatabaseSyncDetail, error) {
	l.Logger.Infof("开始同步TiDB实例主机")

	// 调用Model层方法查询实例
	instances, err := l.svcCtx.TidbClusterInstanceModel.FindInstancesWithCluster(l.ctx)
	if err != nil {
		l.Logger.Errorf("查询TiDB实例失败: %v", err)
		return &cmpool.DatabaseSyncDetail{
			DatabaseType:  "tidb",
			SyncedCount:   0,
			ClusterGroups: []string{},
		}, fmt.Errorf("查询TiDB实例失败: %w", err)
	}

	l.Logger.Infof("查询到 %d 个TiDB实例", len(instances))

	// 同步实例到主机表
	syncedCount, err := l.syncInstancesToHosts(instances, "tidb")
	if err != nil {
		l.Logger.Errorf("同步TiDB实例到主机表失败: %v", err)
		return &cmpool.DatabaseSyncDetail{
			DatabaseType:  "tidb",
			SyncedCount:   0,
			ClusterGroups: []string{},
		}, err
	}

	l.Logger.Infof("TiDB实例主机同步完成: %d 个主机", syncedCount)

	return &cmpool.DatabaseSyncDetail{
		DatabaseType:  "tidb",
		SyncedCount:   int32(syncedCount),
		ClusterGroups: []string{},
	}, nil
}

// syncGoldendbInstanceHosts 同步GoldenDB实例主机数据
func (l *SyncHostsFromClustersLogic) syncGoldendbInstanceHosts() (*cmpool.DatabaseSyncDetail, error) {
	l.Logger.Infof("开始同步GoldenDB实例主机")

	// 调用Model层方法查询实例
	instances, err := l.svcCtx.GoldendbClusterInstanceModel.FindInstancesWithCluster(l.ctx)
	if err != nil {
		l.Logger.Errorf("查询GoldenDB实例失败: %v", err)
		return &cmpool.DatabaseSyncDetail{
			DatabaseType:  "goldendb",
			SyncedCount:   0,
			ClusterGroups: []string{},
		}, fmt.Errorf("查询GoldenDB实例失败: %w", err)
	}

	l.Logger.Infof("查询到 %d 个GoldenDB实例", len(instances))

	// 同步实例到主机表
	syncedCount, err := l.syncInstancesToHosts(instances, "goldendb")
	if err != nil {
		l.Logger.Errorf("同步GoldenDB实例到主机表失败: %v", err)
		return &cmpool.DatabaseSyncDetail{
			DatabaseType:  "goldendb",
			SyncedCount:   0,
			ClusterGroups: []string{},
		}, err
	}

	l.Logger.Infof("GoldenDB实例主机同步完成: %d 个主机", syncedCount)

	return &cmpool.DatabaseSyncDetail{
		DatabaseType:  "goldendb",
		SyncedCount:   int32(syncedCount),
		ClusterGroups: []string{},
	}, nil
}

// syncInstancesToHosts 将实例数据同步到hosts_pool和hosts_applications表
func (l *SyncHostsFromClustersLogic) syncInstancesToHosts(instances []*model.ClusterInstanceInfo, clusterType string) (int, error) {
	syncedCount := 0

	for _, instance := range instances {
		// 1. 同步到hosts_pool表（调用Model层方法）
		hostPoolId, err := l.svcCtx.HostsPoolModel.InsertIfNotExists(l.ctx,
			fmt.Sprintf("%s-host", clusterType),
			instance.IP,
			"database")
		if err != nil {
			l.Logger.Errorf("同步IP %s 到hosts_pool失败: %v", instance.IP, err)
			continue
		}

		// 2. 为新插入的主机计算IDC信息
		l.updateHostIdcInfo(instance.IP)

		// 3. 同步到hosts_applications表（调用Model层方法）
		serverAddr := fmt.Sprintf("%s:%d", instance.IP, instance.Port)
		err = l.svcCtx.HostsApplicationsModel.UpsertApplication(l.ctx,
			hostPoolId,
			clusterType,
			instance.Version,
			instance.ClusterName,
			clusterType,
			serverAddr,
			instance.InstanceRole,
			instance.InstanceStatus,
			"未知", // 部门名称暂时使用默认值
			int32(instance.Port))

		if err != nil {
			l.Logger.Errorf("同步IP %s 到hosts_applications失败: %v", instance.IP, err)
			continue
		}

		syncedCount++
		l.Logger.Infof("成功同步主机: %s (%s:%d)", instance.IP, clusterType, instance.Port)
	}

	return syncedCount, nil
}

// updateHostIdcInfo 更新主机的IDC信息
func (l *SyncHostsFromClustersLogic) updateHostIdcInfo(hostIp string) {
	// 匹配IDC配置（调用Model层方法）
	idcConf, err := l.svcCtx.IdcConfModel.MatchIdcByIp(l.ctx, hostIp)
	if err != nil {
		if err == sql.ErrNoRows {
			l.Logger.Infof("主机%s未匹配到IDC配置", hostIp)
		} else {
			l.Logger.Errorf("匹配主机%s的IDC配置失败: %v", hostIp, err)
		}
		return
	}

	// 更新主机的IDC信息（调用Model层方法）
	err = l.svcCtx.HostsPoolModel.UpdateHostIdcInfo(l.ctx, hostIp, int64(idcConf.Id))
	if err != nil {
		l.Logger.Errorf("更新主机%s的IDC信息失败: %v", hostIp, err)
	} else {
		l.Logger.Infof("成功更新主机%s的IDC信息为%s", hostIp, idcConf.IdcName)
	}
}
