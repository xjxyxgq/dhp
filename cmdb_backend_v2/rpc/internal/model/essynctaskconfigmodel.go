package model

import (
	"context"
	"fmt"

	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

var _ EsSyncTaskConfigModel = (*customEsSyncTaskConfigModel)(nil)

type (
	// EsSyncTaskConfigModel is an interface to be customized, add more methods here,
	// and implement the added methods in customEsSyncTaskConfigModel.
	EsSyncTaskConfigModel interface {
		esSyncTaskConfigModel
		withSession(session sqlx.Session) EsSyncTaskConfigModel
		// Custom methods
		FindAll(ctx context.Context, enabledOnly bool) ([]*EsSyncTaskConfig, error)
		SoftDelete(ctx context.Context, id uint64) error
		UpdateEnabledStatus(ctx context.Context, id uint64, isEnabled bool) error
		CheckTaskNameExists(ctx context.Context, taskName string, excludeId uint64) (bool, error)
	}

	customEsSyncTaskConfigModel struct {
		*defaultEsSyncTaskConfigModel
	}
)

// NewEsSyncTaskConfigModel returns a model for the database table.
func NewEsSyncTaskConfigModel(conn sqlx.SqlConn) EsSyncTaskConfigModel {
	return &customEsSyncTaskConfigModel{
		defaultEsSyncTaskConfigModel: newEsSyncTaskConfigModel(conn),
	}
}

func (m *customEsSyncTaskConfigModel) withSession(session sqlx.Session) EsSyncTaskConfigModel {
	return NewEsSyncTaskConfigModel(sqlx.NewSqlConnFromSession(session))
}

// FindAll 查询所有任务（支持只查询启用的）
func (m *customEsSyncTaskConfigModel) FindAll(ctx context.Context, enabledOnly bool) ([]*EsSyncTaskConfig, error) {
	var query string
	if enabledOnly {
		query = fmt.Sprintf("SELECT %s FROM %s WHERE deleted_at IS NULL AND is_enabled = 1 ORDER BY created_at DESC", esSyncTaskConfigRows, m.table)
	} else {
		query = fmt.Sprintf("SELECT %s FROM %s WHERE deleted_at IS NULL ORDER BY created_at DESC", esSyncTaskConfigRows, m.table)
	}

	var tasks []*EsSyncTaskConfig
	err := m.conn.QueryRowsCtx(ctx, &tasks, query)
	if err != nil {
		return nil, err
	}
	return tasks, nil
}

// SoftDelete 软删除任务
func (m *customEsSyncTaskConfigModel) SoftDelete(ctx context.Context, id uint64) error {
	query := fmt.Sprintf("UPDATE %s SET deleted_at = NOW() WHERE id = ? AND deleted_at IS NULL", m.table)
	_, err := m.conn.ExecCtx(ctx, query, id)
	return err
}

// UpdateEnabledStatus 更新任务启用状态
func (m *customEsSyncTaskConfigModel) UpdateEnabledStatus(ctx context.Context, id uint64, isEnabled bool) error {
	enabledValue := 0
	if isEnabled {
		enabledValue = 1
	}
	query := fmt.Sprintf("UPDATE %s SET is_enabled = ?, updated_at = NOW() WHERE id = ? AND deleted_at IS NULL", m.table)
	_, err := m.conn.ExecCtx(ctx, query, enabledValue, id)
	return err
}

// CheckTaskNameExists 检查任务名是否存在（可排除指定ID）
func (m *customEsSyncTaskConfigModel) CheckTaskNameExists(ctx context.Context, taskName string, excludeId uint64) (bool, error) {
	var count int
	var query string
	var err error

	if excludeId > 0 {
		query = fmt.Sprintf("SELECT COUNT(*) FROM %s WHERE task_name = ? AND id != ? AND deleted_at IS NULL", m.table)
		err = m.conn.QueryRowCtx(ctx, &count, query, taskName, excludeId)
	} else {
		query = fmt.Sprintf("SELECT COUNT(*) FROM %s WHERE task_name = ? AND deleted_at IS NULL", m.table)
		err = m.conn.QueryRowCtx(ctx, &count, query, taskName)
	}

	if err != nil {
		return false, err
	}
	return count > 0, nil
}
