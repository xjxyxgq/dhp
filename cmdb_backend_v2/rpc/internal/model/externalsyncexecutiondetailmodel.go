package model

import (
	"context"
	"fmt"

	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

var _ ExternalSyncExecutionDetailModel = (*customExternalSyncExecutionDetailModel)(nil)

type (
	// ExternalSyncExecutionDetailModel is an interface to be customized, add more methods here,
	// and implement the added methods in customExternalSyncExecutionDetailModel.
	ExternalSyncExecutionDetailModel interface {
		externalSyncExecutionDetailModel
		withSession(session sqlx.Session) ExternalSyncExecutionDetailModel
		// Custom methods
		FindByExecutionId(ctx context.Context, executionId uint64) ([]*ExternalSyncExecutionDetail, error)
	}

	customExternalSyncExecutionDetailModel struct {
		*defaultExternalSyncExecutionDetailModel
	}
)

// NewExternalSyncExecutionDetailModel returns a model for the database table.
func NewExternalSyncExecutionDetailModel(conn sqlx.SqlConn) ExternalSyncExecutionDetailModel {
	return &customExternalSyncExecutionDetailModel{
		defaultExternalSyncExecutionDetailModel: newExternalSyncExecutionDetailModel(conn),
	}
}

func (m *customExternalSyncExecutionDetailModel) withSession(session sqlx.Session) ExternalSyncExecutionDetailModel {
	return NewExternalSyncExecutionDetailModel(sqlx.NewSqlConnFromSession(session))
}

// FindByExecutionId 根据执行记录ID查询详情列表
func (m *customExternalSyncExecutionDetailModel) FindByExecutionId(ctx context.Context, executionId uint64) ([]*ExternalSyncExecutionDetail, error) {
	query := fmt.Sprintf("SELECT %s FROM %s WHERE execution_id = ? ORDER BY created_at DESC", externalSyncExecutionDetailRows, m.table)
	var details []*ExternalSyncExecutionDetail
	err := m.conn.QueryRowsCtx(ctx, &details, query, executionId)
	if err != nil {
		return nil, err
	}
	return details, nil
}
