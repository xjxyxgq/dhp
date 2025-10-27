package model

import "github.com/zeromicro/go-zero/core/stores/sqlx"

var _ ClusterResourceSummaryModel = (*customClusterResourceSummaryModel)(nil)

type (
	// ClusterResourceSummaryModel is an interface to be customized, add more methods here,
	// and implement the added methods in customClusterResourceSummaryModel.
	ClusterResourceSummaryModel interface {
		clusterResourceSummaryModel
		withSession(session sqlx.Session) ClusterResourceSummaryModel
	}

	customClusterResourceSummaryModel struct {
		*defaultClusterResourceSummaryModel
	}
)

// NewClusterResourceSummaryModel returns a model for the database table.
func NewClusterResourceSummaryModel(conn sqlx.SqlConn) ClusterResourceSummaryModel {
	return &customClusterResourceSummaryModel{
		defaultClusterResourceSummaryModel: newClusterResourceSummaryModel(conn),
	}
}

func (m *customClusterResourceSummaryModel) withSession(session sqlx.Session) ClusterResourceSummaryModel {
	return NewClusterResourceSummaryModel(sqlx.NewSqlConnFromSession(session))
}
