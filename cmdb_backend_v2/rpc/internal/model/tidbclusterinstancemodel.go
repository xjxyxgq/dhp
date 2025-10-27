package model

import (
	"context"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

var _ TidbClusterInstanceModel = (*customTidbClusterInstanceModel)(nil)

type (
	// TidbClusterInstanceModel is an interface to be customized, add more methods here,
	// and implement the added methods in customTidbClusterInstanceModel.
	TidbClusterInstanceModel interface {
		tidbClusterInstanceModel
		FindInstancesWithCluster(ctx context.Context) ([]*ClusterInstanceInfo, error)
	}

	customTidbClusterInstanceModel struct {
		*defaultTidbClusterInstanceModel
	}
)

// NewTidbClusterInstanceModel returns a model for the database table.
func NewTidbClusterInstanceModel(conn sqlx.SqlConn) TidbClusterInstanceModel {
	return &customTidbClusterInstanceModel{
		defaultTidbClusterInstanceModel: newTidbClusterInstanceModel(conn),
	}
}

// FindInstancesWithCluster 查询所有TiDB实例及其集群信息
func (m *customTidbClusterInstanceModel) FindInstancesWithCluster(ctx context.Context) ([]*ClusterInstanceInfo, error) {
	query := `
		SELECT
			tc.cluster_name,
			tc.cluster_group_name,
			tci.ip,
			tci.port,
			COALESCE(tci.instance_role, 'tikv') as instance_role,
			COALESCE(tci.version, 'unknown') as version,
			COALESCE(tci.instance_status, 'running') as instance_status
		FROM tidb_cluster_instance tci
		JOIN tidb_cluster tc ON tci.cluster_name = tc.cluster_name
		WHERE tci.ip IS NOT NULL AND tci.ip != '' AND tci.is_deleted = 0
	`

	var instances []*ClusterInstanceInfo
	err := m.conn.QueryRowsCtx(ctx, &instances, query)
	return instances, err
}
