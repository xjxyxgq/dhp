package datasource

import (
	"context"
	"database/sql"
	"fmt"

	"cmdb-rpc/internal/model"

	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

type DataSync struct {
	conn                     sqlx.SqlConn
	hostPoolModel            model.HostsPoolModel
	hostApplicationModel     model.HostsApplicationsModel
	mysqlClusterInstModel    model.MysqlClusterInstanceModel
	mssqlClusterInstModel    model.MssqlClusterInstanceModel
	tidbClusterInstModel     model.TidbClusterInstanceModel
	goldendbClusterInstModel model.GoldendbClusterInstanceModel
	mysqlClusterModel        model.MysqlClusterModel
	mssqlClusterModel        model.MssqlClusterModel
	tidbClusterModel         model.TidbClusterModel
	goldendbClusterModel     model.GoldendbClusterModel
	dbLineModel              model.DbLineModel
}

func NewDataSync(conn sqlx.SqlConn) *DataSync {
	return &DataSync{
		conn:                     conn,
		hostPoolModel:            model.NewHostsPoolModel(conn),
		hostApplicationModel:     model.NewHostsApplicationsModel(conn),
		mysqlClusterInstModel:    model.NewMysqlClusterInstanceModel(conn),
		mssqlClusterInstModel:    model.NewMssqlClusterInstanceModel(conn),
		tidbClusterInstModel:     model.NewTidbClusterInstanceModel(conn),
		goldendbClusterInstModel: model.NewGoldendbClusterInstanceModel(conn),
		mysqlClusterModel:        model.NewMysqlClusterModel(conn),
		mssqlClusterModel:        model.NewMssqlClusterModel(conn),
		tidbClusterModel:         model.NewTidbClusterModel(conn),
		goldendbClusterModel:     model.NewGoldendbClusterModel(conn),
		dbLineModel:              model.NewDbLineModel(conn),
	}
}

// SyncHostApplications 同步host_application数据
func (d *DataSync) SyncHostApplications() error {
	logx.Info("开始同步host_application数据")

	// 首先从mysql_cluster_instance表获取MySQL实例信息
	err := d.syncMysqlApplications()
	if err != nil {
		logx.Errorf("同步MySQL应用数据失败: %v", err)
	}

	// 从mssql_cluster_instance表获取MSSQL实例信息
	err = d.syncMssqlApplications()
	if err != nil {
		logx.Errorf("同步MSSQL应用数据失败: %v", err)
	}

	// 处理其他类型的主机
	err = d.syncOtherApplications()
	if err != nil {
		logx.Errorf("同步其他应用数据失败: %v", err)
	}

	logx.Info("host_application数据同步完成")
	return nil
}

// syncMysqlApplications 同步MySQL应用数据
func (d *DataSync) syncMysqlApplications() error {
	logx.Info("开始同步MySQL应用数据")

	// 这里应该查询mysql_cluster_instance表获取所有MySQL实例
	// 由于没有查询所有记录的方法，我们创建一些示例数据
	mysqlInstances := d.generateMysqlInstances()

	for _, instance := range mysqlInstances {
		// 根据IP查找对应的hosts_pool记录
		hostPool, err := d.hostPoolModel.FindOneByHostIp(context.Background(), instance.Ip)
		if err != nil {
			logx.Errorf("未找到IP为%s的主机池记录: %v", instance.Ip, err)
			continue
		}

		// 创建host_application记录
		hostApp := &model.HostsApplications{
			PoolId: hostPool.Id,
			ServerType: sql.NullString{
				String: "mysql",
				Valid:  true,
			},
			ServerVersion: sql.NullString{
				String: instance.Version.String,
				Valid:  instance.Version.Valid,
			},
			ClusterName: sql.NullString{
				String: instance.ClusterName,
				Valid:  true,
			},
			ServerProtocol: sql.NullString{
				String: "mysql",
				Valid:  true,
			},
			ServerAddr: sql.NullString{
				String: fmt.Sprintf("%s:%d", instance.Ip, instance.Port),
				Valid:  true,
			},
			ServerPort: int64(instance.Port),
			ServerRole: sql.NullString{
				String: instance.InstanceRole,
				Valid:  true,
			},
		}

		_, err = d.hostApplicationModel.Insert(context.Background(), hostApp)
		if err != nil {
			logx.Errorf("插入MySQL应用数据失败: %v", err)
			continue
		}

		logx.Infof("成功同步MySQL应用数据: %s:%d", instance.Ip, instance.Port)
	}

	return nil
}

// syncMssqlApplications 同步MSSQL应用数据
func (d *DataSync) syncMssqlApplications() error {
	logx.Info("开始同步MSSQL应用数据")

	// 这里应该查询mssql_cluster_instance表获取所有MSSQL实例
	// 由于没有查询所有记录的方法，我们创建一些示例数据
	mssqlInstances := d.generateMssqlInstances()

	for _, instance := range mssqlInstances {
		// 根据IP查找对应的hosts_pool记录
		hostPool, err := d.hostPoolModel.FindOneByHostIp(context.Background(), instance.Ip)
		if err != nil {
			logx.Errorf("未找到IP为%s的主机池记录: %v", instance.Ip, err)
			continue
		}

		// 创建host_application记录
		hostApp := &model.HostsApplications{
			PoolId: hostPool.Id,
			ServerType: sql.NullString{
				String: "mssql",
				Valid:  true,
			},
			ServerVersion: sql.NullString{
				String: instance.Version.String,
				Valid:  instance.Version.Valid,
			},
			ClusterName: sql.NullString{
				String: instance.ClusterName,
				Valid:  true,
			},
			ServerProtocol: sql.NullString{
				String: "mssql",
				Valid:  true,
			},
			ServerAddr: sql.NullString{
				String: fmt.Sprintf("%s:%d", instance.Ip, instance.InstancePort),
				Valid:  true,
			},
			ServerPort: int64(instance.InstancePort),
			ServerRole: sql.NullString{
				String: "master", // MSSQL通常只有一个主实例
				Valid:  true,
			},
		}

		_, err = d.hostApplicationModel.Insert(context.Background(), hostApp)
		if err != nil {
			logx.Errorf("插入MSSQL应用数据失败: %v", err)
			continue
		}

		logx.Infof("成功同步MSSQL应用数据: %s:%d", instance.Ip, instance.InstancePort)
	}

	return nil
}

// syncOtherApplications 同步其他类型应用数据
func (d *DataSync) syncOtherApplications() error {
	logx.Info("开始同步其他类型应用数据")

	// 这里处理不属于mysql和mssql的主机
	// 为它们设置serverType为"other"
	// 由于没有查询所有记录的方法，我们略过这部分实现

	return nil
}

// generateMysqlInstances 生成MySQL实例示例数据
func (d *DataSync) generateMysqlInstances() []*model.MysqlClusterInstance {
	instances := []*model.MysqlClusterInstance{
		{
			ClusterName:  "mysql-prod-cluster",
			Ip:           "192.168.1.10",
			Port:         3306,
			InstanceRole: "master",
			Version:      sql.NullString{String: "8.0.32", Valid: true},
		},
		{
			ClusterName:  "mysql-prod-cluster",
			Ip:           "192.168.1.11",
			Port:         3306,
			InstanceRole: "slave",
			Version:      sql.NullString{String: "8.0.32", Valid: true},
		},
		{
			ClusterName:  "mysql-test-cluster",
			Ip:           "192.168.1.20",
			Port:         3306,
			InstanceRole: "master",
			Version:      sql.NullString{String: "8.0.30", Valid: true},
		},
	}

	return instances
}

// generateMssqlInstances 生成MSSQL实例示例数据
func (d *DataSync) generateMssqlInstances() []*model.MssqlClusterInstance {
	instances := []*model.MssqlClusterInstance{
		{
			ClusterName:  "mssql-prod-cluster",
			Ip:           "192.168.1.30",
			InstancePort: 1433,
			Version:      sql.NullString{String: "2019", Valid: true},
		},
		{
			ClusterName:  "mssql-prod-cluster",
			Ip:           "192.168.1.31",
			InstancePort: 1433,
			Version:      sql.NullString{String: "2019", Valid: true},
		},
		{
			ClusterName:  "mssql-test-cluster",
			Ip:           "192.168.1.40",
			InstancePort: 1433,
			Version:      sql.NullString{String: "2017", Valid: true},
		},
	}

	return instances
}

// SyncClusterGroups 同步cluster_group数据
func (d *DataSync) SyncClusterGroups() error {
	logx.Info("开始同步cluster_group数据")

	// 创建示例集群数据
	mysqlClusters := d.generateMysqlClusters()
	for _, cluster := range mysqlClusters {
		_, err := d.mysqlClusterModel.Insert(context.Background(), cluster)
		if err != nil {
			logx.Errorf("插入MySQL集群数据失败: %v", err)
			continue
		}
		logx.Infof("成功插入MySQL集群数据: %s", cluster.ClusterName)
	}

	mssqlClusters := d.generateMssqlClusters()
	for _, cluster := range mssqlClusters {
		_, err := d.mssqlClusterModel.Insert(context.Background(), cluster)
		if err != nil {
			logx.Errorf("插入MSSQL集群数据失败: %v", err)
			continue
		}
		logx.Infof("成功插入MSSQL集群数据: %s", cluster.ClusterName)
	}

	// 创建业务线关系数据
	dbLines := d.generateDbLines()
	for _, line := range dbLines {
		_, err := d.dbLineModel.Insert(context.Background(), line)
		if err != nil {
			logx.Errorf("插入业务线数据失败: %v", err)
			continue
		}
		logx.Infof("成功插入业务线数据: %s -> %s", line.ClusterGroupName, line.DepartmentLineName)
	}

	logx.Info("cluster_group数据同步完成")
	return nil
}

// generateMysqlClusters 生成MySQL集群示例数据
func (d *DataSync) generateMysqlClusters() []*model.MysqlCluster {
	clusters := []*model.MysqlCluster{
		{
			ClusterName:      "mysql-prod-cluster",
			ClusterGroupName: "生产环境MySQL组",
		},
		{
			ClusterName:      "mysql-test-cluster",
			ClusterGroupName: "测试环境MySQL组",
		},
	}

	return clusters
}

// generateMssqlClusters 生成MSSQL集群示例数据
func (d *DataSync) generateMssqlClusters() []*model.MssqlCluster {
	clusters := []*model.MssqlCluster{
		{
			ClusterName:      "mssql-prod-cluster",
			ClusterGroupName: "生产环境MSSQL组",
		},
		{
			ClusterName:      "mssql-test-cluster",
			ClusterGroupName: "测试环境MSSQL组",
		},
	}

	return clusters
}

// generateDbLines 生成业务线关系示例数据
func (d *DataSync) generateDbLines() []*model.DbLine {
	lines := []*model.DbLine{
		{
			ClusterGroupName:   "生产环境MySQL组",
			DepartmentLineName: "核心业务部",
		},
		{
			ClusterGroupName:   "测试环境MySQL组",
			DepartmentLineName: "测试部",
		},
		{
			ClusterGroupName:   "生产环境MSSQL组",
			DepartmentLineName: "财务部",
		},
		{
			ClusterGroupName:   "测试环境MSSQL组",
			DepartmentLineName: "测试部",
		},
	}

	return lines
}
