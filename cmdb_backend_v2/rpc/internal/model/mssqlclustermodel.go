package model

import (
	"context"
	"fmt"

	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

var _ MssqlClusterModel = (*customMssqlClusterModel)(nil)

type (
	// MssqlClusterModel is an interface to be customized, add more methods here,
	// and implement the added methods in customMssqlClusterModel.
	MssqlClusterModel interface {
		mssqlClusterModel
		FindAllClusters(ctx context.Context) ([]*ClusterInfo, error)
	}



	customMssqlClusterModel struct {
		*defaultMssqlClusterModel
	}
)

// NewMssqlClusterModel returns a model for the database table.
func NewMssqlClusterModel(conn sqlx.SqlConn) MssqlClusterModel {
	return &customMssqlClusterModel{
		defaultMssqlClusterModel: newMssqlClusterModel(conn),
	}
}

// FindAllClusters 查询所有MSSQL集群信息
func (m *customMssqlClusterModel) FindAllClusters(ctx context.Context) ([]*ClusterInfo, error) {
	query := `SELECT cluster_name, cluster_group_name FROM mssql_cluster`

	var clusters []*ClusterInfo
	err := m.conn.QueryRowsCtx(ctx, &clusters, query)
	if err != nil {
		return nil, fmt.Errorf("查询MSSQL集群失败: %w", err)
	}

	return clusters, nil
}
