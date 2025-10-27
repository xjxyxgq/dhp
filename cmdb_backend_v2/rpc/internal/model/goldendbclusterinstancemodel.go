package model

import (
	"context"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

var _ GoldendbClusterInstanceModel = (*customGoldendbClusterInstanceModel)(nil)

type (
	// GoldendbClusterInstanceModel is an interface to be customized, add more methods here,
	// and implement the added methods in customGoldendbClusterInstanceModel.
	GoldendbClusterInstanceModel interface {
		goldendbClusterInstanceModel
		FindInstancesWithCluster(ctx context.Context) ([]*ClusterInstanceInfo, error)
	}

	customGoldendbClusterInstanceModel struct {
		*defaultGoldendbClusterInstanceModel
	}
)

// NewGoldendbClusterInstanceModel returns a model for the database table.
func NewGoldendbClusterInstanceModel(conn sqlx.SqlConn) GoldendbClusterInstanceModel {
	return &customGoldendbClusterInstanceModel{
		defaultGoldendbClusterInstanceModel: newGoldendbClusterInstanceModel(conn),
	}
}

// FindInstancesWithCluster 查询所有GoldenDB实例及其集群信息
func (m *customGoldendbClusterInstanceModel) FindInstancesWithCluster(ctx context.Context) ([]*ClusterInstanceInfo, error) {
	query := `
		SELECT
			gc.cluster_name,
			gc.cluster_group_name,
			gci.ip,
			gci.port,
			COALESCE(gci.instance_role, 'storage') as instance_role,
			COALESCE(gci.version, 'unknown') as version,
			COALESCE(gci.instance_status, 'running') as instance_status
		FROM goldendb_cluster_instance gci
		JOIN goldendb_cluster gc ON gci.cluster_name = gc.cluster_name
		WHERE gci.ip IS NOT NULL AND gci.ip != '' AND gci.is_deleted = 0
	`

	var instances []*ClusterInstanceInfo
	err := m.conn.QueryRowsCtx(ctx, &instances, query)
	return instances, err
}
