package model

import (
	"context"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

var _ MssqlClusterInstanceModel = (*customMssqlClusterInstanceModel)(nil)

type (
	// MssqlClusterInstanceModel is an interface to be customized, add more methods here,
	// and implement the added methods in customMssqlClusterInstanceModel.
	MssqlClusterInstanceModel interface {
		mssqlClusterInstanceModel
		FindByHostIp(ctx context.Context, ip string) ([]*MssqlClusterInstance, error)
		FindInstancesWithCluster(ctx context.Context) ([]*ClusterInstanceInfo, error)
	}

	customMssqlClusterInstanceModel struct {
		*defaultMssqlClusterInstanceModel
	}
)

// NewMssqlClusterInstanceModel returns a model for the database table.
func NewMssqlClusterInstanceModel(conn sqlx.SqlConn) MssqlClusterInstanceModel {
	return &customMssqlClusterInstanceModel{
		defaultMssqlClusterInstanceModel: newMssqlClusterInstanceModel(conn),
	}
}

// FindByHostIp 根据主机IP查找MSSQL集群实例
func (m *customMssqlClusterInstanceModel) FindByHostIp(ctx context.Context, ip string) ([]*MssqlClusterInstance, error) {
	query := `SELECT ` + mssqlClusterInstanceRows + ` FROM ` + m.table + ` WHERE ip = ? AND is_deleted = 0`

	var instances []*MssqlClusterInstance
	err := m.conn.QueryRowsCtx(ctx, &instances, query, ip)
	return instances, err
}

// FindInstancesWithCluster 查询所有MSSQL实例及其集群信息
func (m *customMssqlClusterInstanceModel) FindInstancesWithCluster(ctx context.Context) ([]*ClusterInstanceInfo, error) {
	query := `
		SELECT
			mc.cluster_name,
			mc.cluster_group_name,
			mci.ip,
			mci.instance_port as port,
			COALESCE(mci.instance_role, 'primary') as instance_role,
			COALESCE(mci.version, 'unknown') as version,
			COALESCE(mci.instance_status, 'running') as instance_status
		FROM mssql_cluster_instance mci
		JOIN mssql_cluster mc ON mci.cluster_name = mc.cluster_name
		WHERE mci.ip IS NOT NULL AND mci.ip != '' AND mci.is_deleted = 0
	`

	var instances []*ClusterInstanceInfo
	err := m.conn.QueryRowsCtx(ctx, &instances, query)
	return instances, err
}
