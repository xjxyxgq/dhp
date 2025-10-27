package model

import (
	"context"
	"fmt"

	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

var _ ExternalSyncTaskConfigModel = (*customExternalSyncTaskConfigModel)(nil)

type (
	// ExternalSyncTaskConfigModel is an interface to be customized, add more methods here,
	// and implement the added methods in customExternalSyncTaskConfigModel.
	ExternalSyncTaskConfigModel interface {
		externalSyncTaskConfigModel
		withSession(session sqlx.Session) ExternalSyncTaskConfigModel
		// Custom methods
		FindAll(ctx context.Context, enabledOnly bool) ([]*ExternalSyncTaskConfig, error)
		FindByDataSource(ctx context.Context, dataSource string, enabledOnly bool) ([]*ExternalSyncTaskConfig, error)
		SoftDelete(ctx context.Context, id uint64) error
		UpdateEnabledStatus(ctx context.Context, id uint64, isEnabled bool) error
		CheckTaskNameExists(ctx context.Context, taskName string, excludeId uint64) (bool, error)
	}

	customExternalSyncTaskConfigModel struct {
		*defaultExternalSyncTaskConfigModel
	}
)

// NewExternalSyncTaskConfigModel returns a model for the database table.
func NewExternalSyncTaskConfigModel(conn sqlx.SqlConn) ExternalSyncTaskConfigModel {
	return &customExternalSyncTaskConfigModel{
		defaultExternalSyncTaskConfigModel: newExternalSyncTaskConfigModel(conn),
	}
}

func (m *customExternalSyncTaskConfigModel) withSession(session sqlx.Session) ExternalSyncTaskConfigModel {
	return NewExternalSyncTaskConfigModel(sqlx.NewSqlConnFromSession(session))
}

// FindAll 查询所有任务（支持只查询启用的）
func (m *customExternalSyncTaskConfigModel) FindAll(ctx context.Context, enabledOnly bool) ([]*ExternalSyncTaskConfig, error) {
	var query string
	if enabledOnly {
		query = fmt.Sprintf("SELECT %s FROM %s WHERE deleted_at IS NULL AND is_enabled = 1 ORDER BY created_at DESC", externalSyncTaskConfigRows, m.table)
	} else {
		query = fmt.Sprintf("SELECT %s FROM %s WHERE deleted_at IS NULL ORDER BY created_at DESC", externalSyncTaskConfigRows, m.table)
	}

	var tasks []*ExternalSyncTaskConfig
	err := m.conn.QueryRowsCtx(ctx, &tasks, query)
	if err != nil {
		return nil, err
	}
	return tasks, nil
}

// FindByDataSource 按数据源类型查询任务
func (m *customExternalSyncTaskConfigModel) FindByDataSource(ctx context.Context, dataSource string, enabledOnly bool) ([]*ExternalSyncTaskConfig, error) {
	var query string
	if enabledOnly {
		query = fmt.Sprintf("SELECT %s FROM %s WHERE deleted_at IS NULL AND is_enabled = 1 AND data_source = ? ORDER BY created_at DESC", externalSyncTaskConfigRows, m.table)
	} else {
		query = fmt.Sprintf("SELECT %s FROM %s WHERE deleted_at IS NULL AND data_source = ? ORDER BY created_at DESC", externalSyncTaskConfigRows, m.table)
	}

	var tasks []*ExternalSyncTaskConfig
	err := m.conn.QueryRowsCtx(ctx, &tasks, query, dataSource)
	if err != nil {
		return nil, err
	}
	return tasks, nil
}

// SoftDelete 软删除任务
func (m *customExternalSyncTaskConfigModel) SoftDelete(ctx context.Context, id uint64) error {
	query := fmt.Sprintf("UPDATE %s SET deleted_at = NOW() WHERE id = ? AND deleted_at IS NULL", m.table)
	_, err := m.conn.ExecCtx(ctx, query, id)
	return err
}

// UpdateEnabledStatus 更新任务启用状态
func (m *customExternalSyncTaskConfigModel) UpdateEnabledStatus(ctx context.Context, id uint64, isEnabled bool) error {
	enabledValue := 0
	if isEnabled {
		enabledValue = 1
	}
	query := fmt.Sprintf("UPDATE %s SET is_enabled = ?, updated_at = NOW() WHERE id = ? AND deleted_at IS NULL", m.table)
	_, err := m.conn.ExecCtx(ctx, query, enabledValue, id)
	return err
}

// CheckTaskNameExists 检查任务名是否存在（可排除指定ID）
func (m *customExternalSyncTaskConfigModel) CheckTaskNameExists(ctx context.Context, taskName string, excludeId uint64) (bool, error) {
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
