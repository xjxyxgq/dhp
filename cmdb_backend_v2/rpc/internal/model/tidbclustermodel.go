package model

import (
	"context"
	"fmt"

	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

var _ TidbClusterModel = (*customTidbClusterModel)(nil)

type (
	// TidbClusterModel is an interface to be customized, add more methods here,
	// and implement the added methods in customTidbClusterModel.
	TidbClusterModel interface {
		tidbClusterModel
		FindAllClusters(ctx context.Context) ([]*ClusterInfo, error)
	}



	customTidbClusterModel struct {
		*defaultTidbClusterModel
	}
)

// NewTidbClusterModel returns a model for the database table.
func NewTidbClusterModel(conn sqlx.SqlConn) TidbClusterModel {
	return &customTidbClusterModel{
		defaultTidbClusterModel: newTidbClusterModel(conn),
	}
}

// FindAllClusters 查询所有TiDB集群信息
func (m *customTidbClusterModel) FindAllClusters(ctx context.Context) ([]*ClusterInfo, error) {
	query := `SELECT cluster_name, cluster_group_name FROM tidb_cluster`

	var clusters []*ClusterInfo
	err := m.conn.QueryRowsCtx(ctx, &clusters, query)
	if err != nil {
		return nil, fmt.Errorf("查询TiDB集群失败: %w", err)
	}

	return clusters, nil
}
