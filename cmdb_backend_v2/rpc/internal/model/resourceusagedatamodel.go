package model

import "github.com/zeromicro/go-zero/core/stores/sqlx"

var _ ResourceUsageDataModel = (*customResourceUsageDataModel)(nil)

type (
	// ResourceUsageDataModel is an interface to be customized, add more methods here,
	// and implement the added methods in customResourceUsageDataModel.
	ResourceUsageDataModel interface {
		resourceUsageDataModel
		withSession(session sqlx.Session) ResourceUsageDataModel
	}

	customResourceUsageDataModel struct {
		*defaultResourceUsageDataModel
	}
)

// NewResourceUsageDataModel returns a model for the database table.
func NewResourceUsageDataModel(conn sqlx.SqlConn) ResourceUsageDataModel {
	return &customResourceUsageDataModel{
		defaultResourceUsageDataModel: newResourceUsageDataModel(conn),
	}
}

func (m *customResourceUsageDataModel) withSession(session sqlx.Session) ResourceUsageDataModel {
	return NewResourceUsageDataModel(sqlx.NewSqlConnFromSession(session))
}
