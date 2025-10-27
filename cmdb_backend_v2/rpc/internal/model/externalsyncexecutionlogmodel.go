package model

import (
	"context"

	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

var _ ExternalSyncExecutionLogModel = (*customExternalSyncExecutionLogModel)(nil)

type (
	// ExternalSyncExecutionLogModel is an interface to be customized, add more methods here,
	// and implement the added methods in customExternalSyncExecutionLogModel.
	ExternalSyncExecutionLogModel interface {
		externalSyncExecutionLogModel
		withSession(session sqlx.Session) ExternalSyncExecutionLogModel
		// Custom methods
		FindByTaskId(ctx context.Context, taskId uint64, limit int32) ([]*ExternalSyncExecutionLog, error)
		FindLatest(ctx context.Context, limit int32) ([]*ExternalSyncExecutionLog, error)
		FindByDataSource(ctx context.Context, dataSource string, limit int32) ([]*ExternalSyncExecutionLog, error)
		FindByTaskIdAndDataSource(ctx context.Context, taskId uint64, dataSource string, limit int32) ([]*ExternalSyncExecutionLog, error)
		UpdateExecutionResult(ctx context.Context, id uint64, status string, successCount, failedCount, notInPoolCount int, durationMs int64) error
	}

	customExternalSyncExecutionLogModel struct {
		*defaultExternalSyncExecutionLogModel
	}
)

// NewExternalSyncExecutionLogModel returns a model for the database table.
func NewExternalSyncExecutionLogModel(conn sqlx.SqlConn) ExternalSyncExecutionLogModel {
	return &customExternalSyncExecutionLogModel{
		defaultExternalSyncExecutionLogModel: newExternalSyncExecutionLogModel(conn),
	}
}

func (m *customExternalSyncExecutionLogModel) withSession(session sqlx.Session) ExternalSyncExecutionLogModel {
	return NewExternalSyncExecutionLogModel(sqlx.NewSqlConnFromSession(session))
}

// FindByTaskId 根据任务ID查询执行记录
func (m *customExternalSyncExecutionLogModel) FindByTaskId(ctx context.Context, taskId uint64, limit int32) ([]*ExternalSyncExecutionLog, error) {
	var query string
	var logs []*ExternalSyncExecutionLog
	var err error

	if taskId > 0 {
		query = "SELECT id, task_id, task_name, data_source, execution_time, execution_status, total_hosts, success_count, failed_count, not_in_pool_count, error_message, duration_ms, query_time_range, created_at FROM external_sync_execution_log WHERE task_id = ? ORDER BY execution_time DESC LIMIT ?"
		err = m.conn.QueryRowsCtx(ctx, &logs, query, taskId, limit)
	} else {
		query = "SELECT id, task_id, task_name, data_source, execution_time, execution_status, total_hosts, success_count, failed_count, not_in_pool_count, error_message, duration_ms, query_time_range, created_at FROM external_sync_execution_log ORDER BY execution_time DESC LIMIT ?"
		err = m.conn.QueryRowsCtx(ctx, &logs, query, limit)
	}

	if err != nil {
		return nil, err
	}
	return logs, nil
}

// FindLatest 查询最新的执行记录
func (m *customExternalSyncExecutionLogModel) FindLatest(ctx context.Context, limit int32) ([]*ExternalSyncExecutionLog, error) {
	query := "SELECT id, task_id, task_name, data_source, execution_time, execution_status, total_hosts, success_count, failed_count, not_in_pool_count, error_message, duration_ms, query_time_range, created_at FROM external_sync_execution_log ORDER BY execution_time DESC LIMIT ?"
	var logs []*ExternalSyncExecutionLog
	err := m.conn.QueryRowsCtx(ctx, &logs, query, limit)
	if err != nil {
		return nil, err
	}
	return logs, nil
}

// UpdateExecutionResult 更新执行结果
func (m *customExternalSyncExecutionLogModel) UpdateExecutionResult(ctx context.Context, id uint64, status string, successCount, failedCount, notInPoolCount int, durationMs int64) error {
	query := `
		UPDATE external_sync_execution_log
		SET execution_status = ?,
		    success_count = ?,
		    failed_count = ?,
		    not_in_pool_count = ?,
		    duration_ms = ?
		WHERE id = ?`

	_, err := m.conn.ExecCtx(ctx, query, status, successCount, failedCount, notInPoolCount, durationMs, id)
	return err
}

// FindByDataSource 根据数据源查询执行记录（使用 data_source 字段，无需 JOIN）
func (m *customExternalSyncExecutionLogModel) FindByDataSource(ctx context.Context, dataSource string, limit int32) ([]*ExternalSyncExecutionLog, error) {
	query := `
		SELECT id, task_id, task_name, data_source, execution_time,
		       execution_status, total_hosts, success_count,
		       failed_count, not_in_pool_count, error_message,
		       duration_ms, query_time_range, created_at
		FROM external_sync_execution_log
		WHERE data_source = ?
		ORDER BY execution_time DESC
		LIMIT ?`

	var logs []*ExternalSyncExecutionLog
	err := m.conn.QueryRowsCtx(ctx, &logs, query, dataSource, limit)
	if err != nil {
		return nil, err
	}
	return logs, nil
}

// FindByTaskIdAndDataSource 根据任务ID和数据源查询执行记录（使用 data_source 字段，无需 JOIN）
func (m *customExternalSyncExecutionLogModel) FindByTaskIdAndDataSource(ctx context.Context, taskId uint64, dataSource string, limit int32) ([]*ExternalSyncExecutionLog, error) {
	query := `
		SELECT id, task_id, task_name, data_source, execution_time,
		       execution_status, total_hosts, success_count,
		       failed_count, not_in_pool_count, error_message,
		       duration_ms, query_time_range, created_at
		FROM external_sync_execution_log
		WHERE task_id = ? AND data_source = ?
		ORDER BY execution_time DESC
		LIMIT ?`

	var logs []*ExternalSyncExecutionLog
	err := m.conn.QueryRowsCtx(ctx, &logs, query, taskId, dataSource, limit)
	if err != nil {
		return nil, err
	}
	return logs, nil
}
