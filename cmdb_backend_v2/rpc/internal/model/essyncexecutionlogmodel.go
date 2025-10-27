package model

import (
	"context"

	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

var _ EsSyncExecutionLogModel = (*customEsSyncExecutionLogModel)(nil)

type (
	// EsSyncExecutionLogModel is an interface to be customized, add more methods here,
	// and implement the added methods in customEsSyncExecutionLogModel.
	EsSyncExecutionLogModel interface {
		esSyncExecutionLogModel
		withSession(session sqlx.Session) EsSyncExecutionLogModel
		// Custom methods
		FindByTaskId(ctx context.Context, taskId uint64, limit int32) ([]*EsSyncExecutionLog, error)
		FindLatest(ctx context.Context, limit int32) ([]*EsSyncExecutionLog, error)
		FindByDataSource(ctx context.Context, dataSource string, limit int32) ([]*EsSyncExecutionLog, error)
		FindByTaskIdAndDataSource(ctx context.Context, taskId uint64, dataSource string, limit int32) ([]*EsSyncExecutionLog, error)
		UpdateExecutionResult(ctx context.Context, id uint64, status string, successCount, failedCount, notInPoolCount int, durationMs int64) error
	}

	customEsSyncExecutionLogModel struct {
		*defaultEsSyncExecutionLogModel
	}
)

// NewEsSyncExecutionLogModel returns a model for the database table.
func NewEsSyncExecutionLogModel(conn sqlx.SqlConn) EsSyncExecutionLogModel {
	return &customEsSyncExecutionLogModel{
		defaultEsSyncExecutionLogModel: newEsSyncExecutionLogModel(conn),
	}
}

func (m *customEsSyncExecutionLogModel) withSession(session sqlx.Session) EsSyncExecutionLogModel {
	return NewEsSyncExecutionLogModel(sqlx.NewSqlConnFromSession(session))
}

// FindByTaskId 根据任务ID查询执行记录
func (m *customEsSyncExecutionLogModel) FindByTaskId(ctx context.Context, taskId uint64, limit int32) ([]*EsSyncExecutionLog, error) {
	var query string
	var logs []*EsSyncExecutionLog
	var err error

	if taskId > 0 {
		query = "SELECT id, task_id, task_name, execution_time, execution_status, total_hosts, success_count, failed_count, not_in_pool_count, error_message, duration_ms, query_time_range, created_at FROM es_sync_execution_log WHERE task_id = ? ORDER BY execution_time DESC LIMIT ?"
		err = m.conn.QueryRowsCtx(ctx, &logs, query, taskId, limit)
	} else {
		query = "SELECT id, task_id, task_name, execution_time, execution_status, total_hosts, success_count, failed_count, not_in_pool_count, error_message, duration_ms, query_time_range, created_at FROM es_sync_execution_log ORDER BY execution_time DESC LIMIT ?"
		err = m.conn.QueryRowsCtx(ctx, &logs, query, limit)
	}

	if err != nil {
		return nil, err
	}
	return logs, nil
}

// FindLatest 查询最新的执行记录
func (m *customEsSyncExecutionLogModel) FindLatest(ctx context.Context, limit int32) ([]*EsSyncExecutionLog, error) {
	query := "SELECT id, task_id, task_name, execution_time, execution_status, total_hosts, success_count, failed_count, not_in_pool_count, error_message, duration_ms, query_time_range, created_at FROM es_sync_execution_log ORDER BY execution_time DESC LIMIT ?"
	var logs []*EsSyncExecutionLog
	err := m.conn.QueryRowsCtx(ctx, &logs, query, limit)
	if err != nil {
		return nil, err
	}
	return logs, nil
}

// UpdateExecutionResult 更新执行结果
func (m *customEsSyncExecutionLogModel) UpdateExecutionResult(ctx context.Context, id uint64, status string, successCount, failedCount, notInPoolCount int, durationMs int64) error {
	query := `
		UPDATE es_sync_execution_log
		SET execution_status = ?,
		    success_count = ?,
		    failed_count = ?,
		    not_in_pool_count = ?,
		    duration_ms = ?
		WHERE id = ?`

	_, err := m.conn.ExecCtx(ctx, query, status, successCount, failedCount, notInPoolCount, durationMs, id)
	return err
}

// FindByDataSource 根据数据源查询执行记录（通过 JOIN 任务配置表）
func (m *customEsSyncExecutionLogModel) FindByDataSource(ctx context.Context, dataSource string, limit int32) ([]*EsSyncExecutionLog, error) {
	// 构建带表别名的字段列表
	query := `
		SELECT log.id, log.task_id, log.task_name, log.execution_time,
		       log.execution_status, log.total_hosts, log.success_count,
		       log.failed_count, log.not_in_pool_count, log.error_message,
		       log.duration_ms, log.query_time_range, log.created_at
		FROM es_sync_execution_log log
		INNER JOIN external_sync_task_config config ON log.task_id = config.id
		WHERE config.data_source = ? AND config.deleted_at IS NULL
		ORDER BY log.execution_time DESC
		LIMIT ?`

	var logs []*EsSyncExecutionLog
	err := m.conn.QueryRowsCtx(ctx, &logs, query, dataSource, limit)
	if err != nil {
		return nil, err
	}
	return logs, nil
}

// FindByTaskIdAndDataSource 根据任务ID和数据源查询执行记录
func (m *customEsSyncExecutionLogModel) FindByTaskIdAndDataSource(ctx context.Context, taskId uint64, dataSource string, limit int32) ([]*EsSyncExecutionLog, error) {
	// 构建带表别名的字段列表
	query := `
		SELECT log.id, log.task_id, log.task_name, log.execution_time,
		       log.execution_status, log.total_hosts, log.success_count,
		       log.failed_count, log.not_in_pool_count, log.error_message,
		       log.duration_ms, log.query_time_range, log.created_at
		FROM es_sync_execution_log log
		INNER JOIN external_sync_task_config config ON log.task_id = config.id
		WHERE log.task_id = ? AND config.data_source = ? AND config.deleted_at IS NULL
		ORDER BY log.execution_time DESC
		LIMIT ?`

	var logs []*EsSyncExecutionLog
	err := m.conn.QueryRowsCtx(ctx, &logs, query, taskId, dataSource, limit)
	if err != nil {
		return nil, err
	}
	return logs, nil
}
