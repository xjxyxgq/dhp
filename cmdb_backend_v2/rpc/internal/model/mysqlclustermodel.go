package model

import (
	"context"
	"fmt"

	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

var _ MysqlClusterModel = (*customMysqlClusterModel)(nil)

type (
	// MysqlClusterModel is an interface to be customized, add more methods here,
	// and implement the added methods in customMysqlClusterModel.
	MysqlClusterModel interface {
		mysqlClusterModel
		FindAllClusters(ctx context.Context) ([]*ClusterInfo, error)
	}



	customMysqlClusterModel struct {
		*defaultMysqlClusterModel
	}
)

// NewMysqlClusterModel returns a model for the database table.
func NewMysqlClusterModel(conn sqlx.SqlConn) MysqlClusterModel {
	return &customMysqlClusterModel{
		defaultMysqlClusterModel: newMysqlClusterModel(conn),
	}
}

// FindAllClusters 查询所有MySQL集群信息
func (m *customMysqlClusterModel) FindAllClusters(ctx context.Context) ([]*ClusterInfo, error) {
	query := `SELECT cluster_name, cluster_group_name FROM mysql_cluster`

	var clusters []*ClusterInfo
	err := m.conn.QueryRowsCtx(ctx, &clusters, query)
	if err != nil {
		return nil, fmt.Errorf("查询MySQL集群失败: %w", err)
	}

	return clusters, nil
}
