package model

import (
	"database/sql"
	"fmt"
	"github.com/zeromicro/go-zero/core/stores/sqlc"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

var _ ScheduledTaskExecutionHistoryModel = (*customScheduledTaskExecutionHistoryModel)(nil)

type (
	// ScheduledTaskExecutionHistoryModel is an interface to be customized, add more methods here,
	// and implement the added methods in customScheduledTaskExecutionHistoryModel.
	ScheduledTaskExecutionHistoryModel interface {
		scheduledTaskExecutionHistoryModel
		withSession(session sqlx.Session) ScheduledTaskExecutionHistoryModel
		FindByTaskId(scheduledTaskId int64, limit int32) ([]*ScheduledTaskExecutionHistory, error)
		FindByExecutionTaskId(executionTaskId string) (*ScheduledTaskExecutionHistory, error)
		UpdateExecutionStatus(id int64, status string, totalHosts, successHosts, failedHosts int32, errorMessage string) error
	}

	customScheduledTaskExecutionHistoryModel struct {
		*defaultScheduledTaskExecutionHistoryModel
	}
)

// NewScheduledTaskExecutionHistoryModel returns a model for the database table.
func NewScheduledTaskExecutionHistoryModel(conn sqlx.SqlConn) ScheduledTaskExecutionHistoryModel {
	return &customScheduledTaskExecutionHistoryModel{
		defaultScheduledTaskExecutionHistoryModel: newScheduledTaskExecutionHistoryModel(conn),
	}
}

func (m *customScheduledTaskExecutionHistoryModel) withSession(session sqlx.Session) ScheduledTaskExecutionHistoryModel {
	return NewScheduledTaskExecutionHistoryModel(sqlx.NewSqlConnFromSession(session))
}

// FindByTaskId 根据定时任务ID查找执行历史
func (m *customScheduledTaskExecutionHistoryModel) FindByTaskId(scheduledTaskId int64, limit int32) ([]*ScheduledTaskExecutionHistory, error) {
	if limit <= 0 {
		limit = 50 // 默认限制
	}

	query := fmt.Sprintf("select %s from %s where `scheduled_task_id` = ? order by `execution_time` desc limit ?", scheduledTaskExecutionHistoryRows, m.table)

	var resp []*ScheduledTaskExecutionHistory
	err := m.conn.QueryRows(&resp, query, scheduledTaskId, limit)
	return resp, err
}

// FindByExecutionTaskId 根据执行任务的ID查找执行历史
func (m *customScheduledTaskExecutionHistoryModel) FindByExecutionTaskId(executionTaskId string) (*ScheduledTaskExecutionHistory, error) {
	query := fmt.Sprintf("select %s from %s where `execution_task_id` = ? limit 1", scheduledTaskExecutionHistoryRows, m.table)
	var resp ScheduledTaskExecutionHistory
	err := m.conn.QueryRow(&resp, query, executionTaskId)
	switch err {
	case nil:
		return &resp, nil
	case sqlc.ErrNotFound:
		return nil, ErrNotFound
	default:
		return nil, err
	}
}

// UpdateExecutionStatus 更新执行状态
func (m *customScheduledTaskExecutionHistoryModel) UpdateExecutionStatus(id int64, status string, totalHosts, successHosts, failedHosts int32, errorMessage string) error {
	var errMsg sql.NullString
	if errorMessage != "" {
		errMsg = sql.NullString{String: errorMessage, Valid: true}
	}

	query := fmt.Sprintf("update %s set `execution_status` = ?, `total_hosts` = ?, `success_hosts` = ?, `failed_hosts` = ?, `error_message` = ? where `id` = ?", m.table)
	_, err := m.conn.Exec(query, status, totalHosts, successHosts, failedHosts, errMsg, id)
	return err
}
