package model

import (
	"context"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

var _ MysqlClusterInstanceModel = (*customMysqlClusterInstanceModel)(nil)

type (
	// MysqlClusterInstanceModel is an interface to be customized, add more methods here,
	// and implement the added methods in customMysqlClusterInstanceModel.
	MysqlClusterInstanceModel interface {
		mysqlClusterInstanceModel
		FindByHostIp(ctx context.Context, ip string) ([]*MysqlClusterInstance, error)
		FindInstancesWithCluster(ctx context.Context) ([]*ClusterInstanceInfo, error)
	}

	// ClusterInstanceInfo 集群实例信息（用于主机同步）
	ClusterInstanceInfo struct {
		ClusterName      string `db:"cluster_name"`
		ClusterGroupName string `db:"cluster_group_name"`
		IP               string `db:"ip"`
		Port             int    `db:"port"`
		InstanceRole     string `db:"instance_role"`
		Version          string `db:"version"`
		InstanceStatus   string `db:"instance_status"`
	}

	customMysqlClusterInstanceModel struct {
		*defaultMysqlClusterInstanceModel
	}
)

// NewMysqlClusterInstanceModel returns a model for the database table.
func NewMysqlClusterInstanceModel(conn sqlx.SqlConn) MysqlClusterInstanceModel {
	return &customMysqlClusterInstanceModel{
		defaultMysqlClusterInstanceModel: newMysqlClusterInstanceModel(conn),
	}
}

// FindByHostIp 根据主机IP查找MySQL集群实例
func (m *customMysqlClusterInstanceModel) FindByHostIp(ctx context.Context, ip string) ([]*MysqlClusterInstance, error) {
	query := `SELECT ` + mysqlClusterInstanceRows + ` FROM ` + m.table + ` WHERE ip = ? AND is_deleted = 0`

	var instances []*MysqlClusterInstance
	err := m.conn.QueryRowsCtx(ctx, &instances, query, ip)
	return instances, err
}

// FindInstancesWithCluster 查询所有MySQL实例及其集群信息
func (m *customMysqlClusterInstanceModel) FindInstancesWithCluster(ctx context.Context) ([]*ClusterInstanceInfo, error) {
	query := `
		SELECT
			mc.cluster_name,
			mc.cluster_group_name,
			mci.ip,
			mci.port,
			mci.instance_role,
			COALESCE(mci.version, 'unknown') as version,
			COALESCE(mci.instance_status, 'running') as instance_status
		FROM mysql_cluster_instance mci
		JOIN mysql_cluster mc ON mci.cluster_name = mc.cluster_name
		WHERE mci.ip IS NOT NULL AND mci.ip != '' AND mci.is_deleted = 0
	`

	var instances []*ClusterInstanceInfo
	err := m.conn.QueryRowsCtx(ctx, &instances, query)
	return instances, err
}
