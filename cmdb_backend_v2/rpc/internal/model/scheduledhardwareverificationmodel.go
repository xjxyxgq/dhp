package model

import (
	"database/sql"
	"fmt"
	"github.com/zeromicro/go-zero/core/stores/sqlc"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
	"strings"
	"time"
)

var _ ScheduledHardwareVerificationModel = (*customScheduledHardwareVerificationModel)(nil)

type (
	// ScheduledHardwareVerificationModel is an interface to be customized, add more methods here,
	// and implement the added methods in customScheduledHardwareVerificationModel.
	ScheduledHardwareVerificationModel interface {
		scheduledHardwareVerificationModel
		withSession(session sqlx.Session) ScheduledHardwareVerificationModel
		FindAll(resourceType string, enabledOnly bool) ([]*ScheduledHardwareVerification, error)
		FindByExecutionTaskId(executionTaskId int64) (*ScheduledHardwareVerification, error)
		UpdateExecutionTime(id int64, lastExecutionTime, nextExecutionTime *time.Time) error
		EnableTask(id int64, enabled bool) error
	}

	customScheduledHardwareVerificationModel struct {
		*defaultScheduledHardwareVerificationModel
	}
)

// NewScheduledHardwareVerificationModel returns a model for the database table.
func NewScheduledHardwareVerificationModel(conn sqlx.SqlConn) ScheduledHardwareVerificationModel {
	return &customScheduledHardwareVerificationModel{
		defaultScheduledHardwareVerificationModel: newScheduledHardwareVerificationModel(conn),
	}
}

func (m *customScheduledHardwareVerificationModel) withSession(session sqlx.Session) ScheduledHardwareVerificationModel {
	return NewScheduledHardwareVerificationModel(sqlx.NewSqlConnFromSession(session))
}

// FindAll 查找所有定时任务
func (m *customScheduledHardwareVerificationModel) FindAll(resourceType string, enabledOnly bool) ([]*ScheduledHardwareVerification, error) {
	var conditions []string
	var args []interface{}

	if resourceType != "" {
		conditions = append(conditions, "`resource_type` = ?")
		args = append(args, resourceType)
	}

	if enabledOnly {
		conditions = append(conditions, "`is_enabled` = ?")
		args = append(args, true)
	}

	whereClause := ""
	if len(conditions) > 0 {
		whereClause = "WHERE " + strings.Join(conditions, " AND ")
	}

	query := fmt.Sprintf("select %s from %s %s order by `create_time` desc", scheduledHardwareVerificationRows, m.table, whereClause)

	var resp []*ScheduledHardwareVerification
	err := m.conn.QueryRows(&resp, query, args...)
	return resp, err
}

// UpdateExecutionTime 更新执行时间
func (m *customScheduledHardwareVerificationModel) UpdateExecutionTime(id int64, lastExecutionTime, nextExecutionTime *time.Time) error {
	var lastExecTime, nextExecTime sql.NullTime

	if lastExecutionTime != nil {
		lastExecTime = sql.NullTime{Time: *lastExecutionTime, Valid: true}
	}

	if nextExecutionTime != nil {
		nextExecTime = sql.NullTime{Time: *nextExecutionTime, Valid: true}
	}

	query := fmt.Sprintf("update %s set `last_execution_time` = ?, `next_execution_time` = ?, `update_time` = CURRENT_TIMESTAMP where `id` = ?", m.table)
	_, err := m.conn.Exec(query, lastExecTime, nextExecTime, id)
	return err
}

// FindByExecutionTaskId 根据执行任务的ID查找定时任务
func (m *customScheduledHardwareVerificationModel) FindByExecutionTaskId(executionTaskId int64) (*ScheduledHardwareVerification, error) {
	query := fmt.Sprintf("select %s from %s where `id` = ? limit 1", scheduledHardwareVerificationRows, m.table)
	var resp ScheduledHardwareVerification
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

// EnableTask 启用/禁用任务
func (m *customScheduledHardwareVerificationModel) EnableTask(id int64, enabled bool) error {
	query := fmt.Sprintf("update %s set `is_enabled` = ?, `update_time` = CURRENT_TIMESTAMP where `id` = ?", m.table)
	_, err := m.conn.Exec(query, enabled, id)
	return err
}
