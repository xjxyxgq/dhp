package model

import (
	"context"
	"fmt"

	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

var _ GoldendbClusterModel = (*customGoldendbClusterModel)(nil)

type (
	// GoldendbClusterModel is an interface to be customized, add more methods here,
	// and implement the added methods in customGoldendbClusterModel.
	GoldendbClusterModel interface {
		goldendbClusterModel
		FindAllClusters(ctx context.Context) ([]*ClusterInfo, error)
	}



	customGoldendbClusterModel struct {
		*defaultGoldendbClusterModel
	}
)

// NewGoldendbClusterModel returns a model for the database table.
func NewGoldendbClusterModel(conn sqlx.SqlConn) GoldendbClusterModel {
	return &customGoldendbClusterModel{
		defaultGoldendbClusterModel: newGoldendbClusterModel(conn),
	}
}

// FindAllClusters 查询所有GoldenDB集群信息
func (m *customGoldendbClusterModel) FindAllClusters(ctx context.Context) ([]*ClusterInfo, error) {
	query := `SELECT cluster_name, cluster_group_name FROM goldendb_cluster`

	var clusters []*ClusterInfo
	err := m.conn.QueryRowsCtx(ctx, &clusters, query)
	if err != nil {
		return nil, fmt.Errorf("查询GoldenDB集群失败: %w", err)
	}

	return clusters, nil
}
