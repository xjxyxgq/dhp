package model

import (
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

var _ ResourceAnalysisReportsModel = (*customResourceAnalysisReportsModel)(nil)

type (
	// ResourceAnalysisReportsModel is an interface to be customized, add more methods here,
	// and implement the added methods in customResourceAnalysisReportsModel.
	ResourceAnalysisReportsModel interface {
		resourceAnalysisReportsModel
	}

	customResourceAnalysisReportsModel struct {
		*defaultResourceAnalysisReportsModel
	}
)

// NewResourceAnalysisReportsModel returns a model for the database table.
func NewResourceAnalysisReportsModel(conn sqlx.SqlConn) ResourceAnalysisReportsModel {
	return &customResourceAnalysisReportsModel{
		defaultResourceAnalysisReportsModel: newResourceAnalysisReportsModel(conn),
	}
}
