package model

import (
	"context"
	"fmt"

	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

var _ EsSyncExecutionDetailModel = (*customEsSyncExecutionDetailModel)(nil)

type (
	// EsSyncExecutionDetailModel is an interface to be customized, add more methods here,
	// and implement the added methods in customEsSyncExecutionDetailModel.
	EsSyncExecutionDetailModel interface {
		esSyncExecutionDetailModel
		withSession(session sqlx.Session) EsSyncExecutionDetailModel
		// Custom methods
		FindByExecutionId(ctx context.Context, executionId uint64) ([]*EsSyncExecutionDetail, error)
	}

	customEsSyncExecutionDetailModel struct {
		*defaultEsSyncExecutionDetailModel
	}
)

// NewEsSyncExecutionDetailModel returns a model for the database table.
func NewEsSyncExecutionDetailModel(conn sqlx.SqlConn) EsSyncExecutionDetailModel {
	return &customEsSyncExecutionDetailModel{
		defaultEsSyncExecutionDetailModel: newEsSyncExecutionDetailModel(conn),
	}
}

func (m *customEsSyncExecutionDetailModel) withSession(session sqlx.Session) EsSyncExecutionDetailModel {
	return NewEsSyncExecutionDetailModel(sqlx.NewSqlConnFromSession(session))
}

// FindByExecutionId 根据执行记录ID查询详情列表
func (m *customEsSyncExecutionDetailModel) FindByExecutionId(ctx context.Context, executionId uint64) ([]*EsSyncExecutionDetail, error) {
	query := fmt.Sprintf("SELECT %s FROM %s WHERE execution_id = ? ORDER BY created_at DESC", esSyncExecutionDetailRows, m.table)
	var details []*EsSyncExecutionDetail
	err := m.conn.QueryRowsCtx(ctx, &details, query, executionId)
	if err != nil {
		return nil, err
	}
	return details, nil
}
