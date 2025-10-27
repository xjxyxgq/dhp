package model

import (
	"context"
	"database/sql"
	"fmt"
	"strings"

	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

var _ HardwareResourceVerificationModel = (*customHardwareResourceVerificationModel)(nil)

type (
	// HardwareResourceVerificationModel is an interface to be customized, add more methods here,
	// and implement the added methods in customHardwareResourceVerificationModel.
	HardwareResourceVerificationModel interface {
		hardwareResourceVerificationModel
		InsertVerification(ctx context.Context, data *HardwareResourceVerification) (sql.Result, error)
		FindByTaskId(ctx context.Context, taskId string) ([]*HardwareResourceVerification, error)
		FindByHostIp(ctx context.Context, hostIp string, resourceType string, limit int32) ([]*HardwareResourceVerification, error)
		FindByHostIpList(ctx context.Context, hostIpList []string, resourceType string) ([]*HardwareResourceVerification, error)
		UpdateVerificationStatus(ctx context.Context, id int64, status, startTime, endTime string, exitCode sql.NullInt64, stdoutLog, stderrLog, resultSummary, sshError sql.NullString) error
		TerminateRunningTasks(ctx context.Context, hostIp, resourceType string) error
		HasRunningTask(ctx context.Context, hostIp, resourceType string) (bool, error)
	}

	customHardwareResourceVerificationModel struct {
		*defaultHardwareResourceVerificationModel
	}
)

// NewHardwareResourceVerificationModel returns a model for the database table.
func NewHardwareResourceVerificationModel(conn sqlx.SqlConn) HardwareResourceVerificationModel {
	return &customHardwareResourceVerificationModel{
		defaultHardwareResourceVerificationModel: newHardwareResourceVerificationModel(conn),
	}
}

// InsertVerification 插入硬件资源验证记录
func (m *customHardwareResourceVerificationModel) InsertVerification(ctx context.Context, data *HardwareResourceVerification) (sql.Result, error) {
	query := `INSERT INTO hardware_resource_verification 
		(task_id, host_ip, resource_type, target_percent, duration, script_params, execution_status) 
		VALUES (?, ?, ?, ?, ?, ?, ?)`
	return m.conn.ExecCtx(ctx, query, data.TaskId, data.HostIp, data.ResourceType, data.TargetPercent,
		data.Duration, data.ScriptParams, data.ExecutionStatus)
}

// FindByTaskId 根据任务ID查询验证记录
func (m *customHardwareResourceVerificationModel) FindByTaskId(ctx context.Context, taskId string) ([]*HardwareResourceVerification, error) {
	query := `SELECT id, create_time, update_time, delete_time, is_deleted, task_id, host_ip, resource_type, 
		target_percent, duration, script_params, execution_status, start_time, end_time, 
		exit_code, stdout_log, stderr_log, result_summary, ssh_error 
		FROM hardware_resource_verification 
		WHERE task_id = ? AND is_deleted = 0
		ORDER BY create_time DESC`

	var resp []*HardwareResourceVerification
	err := m.conn.QueryRowsCtx(ctx, &resp, query, taskId)
	return resp, err
}

// FindByHostIp 根据主机IP查询历史记录
func (m *customHardwareResourceVerificationModel) FindByHostIp(ctx context.Context, hostIp string, resourceType string, limit int32) ([]*HardwareResourceVerification, error) {
	var args []interface{}
	var conditions []string

	conditions = append(conditions, "host_ip = ?")
	args = append(args, hostIp)

	if resourceType != "" {
		conditions = append(conditions, "resource_type = ?")
		args = append(args, resourceType)
	}

	conditions = append(conditions, "1=1")

	limitClause := ""
	if limit > 0 {
		limitClause = fmt.Sprintf(" LIMIT %d", limit)
	} else {
		limitClause = " LIMIT 50" // 默认限制50条
	}

	query := fmt.Sprintf(`SELECT id, create_time, update_time, delete_time, is_deleted, task_id, host_ip, resource_type, 
		target_percent, duration, script_params, execution_status, start_time, end_time, 
		exit_code, stdout_log, stderr_log, result_summary, ssh_error 
		FROM hardware_resource_verification 
		WHERE %s AND is_deleted = 0
		ORDER BY create_time DESC%s`, strings.Join(conditions, " AND "), limitClause)

	var resp []*HardwareResourceVerification
	err := m.conn.QueryRowsCtx(ctx, &resp, query, args...)
	return resp, err
}

// FindByHostIpList 根据主机IP列表查询最新状态
func (m *customHardwareResourceVerificationModel) FindByHostIpList(ctx context.Context, hostIpList []string, resourceType string) ([]*HardwareResourceVerification, error) {
	if len(hostIpList) == 0 {
		// 如果没有指定IP列表，返回所有记录的最新状态
		var conditions []string
		var args []interface{}

		if resourceType != "" {
			conditions = append(conditions, "resource_type = ?")
			args = append(args, resourceType)
		}

		conditions = append(conditions, "1=1")

		whereClause := ""
		if len(conditions) > 0 {
			whereClause = "WHERE " + strings.Join(conditions, " AND ")
		}

		query := fmt.Sprintf(`SELECT v1.id, v1.create_time, v1.update_time, v1.delete_time, v1.is_deleted, v1.task_id, 
			v1.host_ip, v1.resource_type, v1.target_percent, v1.duration, v1.script_params, 
			v1.execution_status, v1.start_time, v1.end_time, v1.exit_code, v1.stdout_log, 
			v1.stderr_log, v1.result_summary, v1.ssh_error 
			FROM hardware_resource_verification v1
			INNER JOIN (
				SELECT host_ip, resource_type, MAX(create_time) as max_create_time
				FROM hardware_resource_verification 
				%s AND is_deleted = 0
				GROUP BY host_ip, resource_type
			) v2 ON v1.host_ip = v2.host_ip AND v1.resource_type = v2.resource_type 
			AND v1.create_time = v2.max_create_time AND v1.is_deleted = 0
			ORDER BY v1.create_time DESC`, whereClause)

		var resp []*HardwareResourceVerification
		err := m.conn.QueryRowsCtx(ctx, &resp, query, args...)
		return resp, err
	}

	// 构建IN子句的占位符
	placeholders := make([]string, len(hostIpList))
	args := make([]interface{}, 0, len(hostIpList)+1)

	for i, ip := range hostIpList {
		placeholders[i] = "?"
		args = append(args, ip)
	}

	var conditions []string
	conditions = append(conditions, fmt.Sprintf("host_ip IN (%s)", strings.Join(placeholders, ",")))

	if resourceType != "" {
		conditions = append(conditions, "resource_type = ?")
		args = append(args, resourceType)
	}

	conditions = append(conditions, "1=1")

	query := fmt.Sprintf(`SELECT v1.id, v1.create_time, v1.update_time, v1.delete_time, v1.is_deleted, v1.task_id, 
		v1.host_ip, v1.resource_type, v1.target_percent, v1.duration, v1.script_params, 
		v1.execution_status, v1.start_time, v1.end_time, v1.exit_code, v1.stdout_log, 
		v1.stderr_log, v1.result_summary, v1.ssh_error 
		FROM hardware_resource_verification v1
		INNER JOIN (
			SELECT host_ip, resource_type, MAX(create_time) as max_create_time
			FROM hardware_resource_verification 
			WHERE %s AND is_deleted = 0
			GROUP BY host_ip, resource_type
		) v2 ON v1.host_ip = v2.host_ip AND v1.resource_type = v2.resource_type 
		AND v1.create_time = v2.max_create_time AND v1.is_deleted = 0
		ORDER BY v1.create_time DESC`, strings.Join(conditions, " AND "))

	var resp []*HardwareResourceVerification
	err := m.conn.QueryRowsCtx(ctx, &resp, query, args...)
	return resp, err
}

// UpdateVerificationStatus 更新验证状态
func (m *customHardwareResourceVerificationModel) UpdateVerificationStatus(ctx context.Context, id int64, status, startTime, endTime string,
	exitCode sql.NullInt64, stdoutLog, stderrLog, resultSummary, sshError sql.NullString) error {

	// 处理endTime为空字符串的情况，转换为NULL
	var endTimeVal sql.NullString
	if endTime != "" {
		endTimeVal = sql.NullString{String: endTime, Valid: true}
	} else {
		endTimeVal = sql.NullString{Valid: false}
	}

	// 处理startTime为空字符串的情况，转换为NULL
	var startTimeVal sql.NullString
	if startTime != "" {
		startTimeVal = sql.NullString{String: startTime, Valid: true}
	} else {
		startTimeVal = sql.NullString{Valid: false}
	}

	query := `UPDATE hardware_resource_verification 
		SET execution_status = ?, start_time = ?, end_time = ?, exit_code = ?, 
		stdout_log = ?, stderr_log = ?, result_summary = ?, ssh_error = ?, 
		update_time = CURRENT_TIMESTAMP 
		WHERE id = ?`
	_, err := m.conn.ExecCtx(ctx, query, status, startTimeVal, endTimeVal, exitCode, stdoutLog, stderrLog, resultSummary, sshError, id)
	return err
}

func (m *customHardwareResourceVerificationModel) TerminateRunningTasks(ctx context.Context, hostIp, resourceType string) error {
	// 更新数据库状态为失败
	updateQuery := `
		UPDATE hardware_resource_verification 
		SET execution_status = 'failed',
			result_summary = '{"terminated": true, "reason": "被新任务强制终止"}',
			end_time = NOW()
		WHERE host_ip = ? 
		AND resource_type = ? 
		AND execution_status IN ('pending', 'running')
	`

	_, err := m.conn.ExecCtx(ctx, updateQuery, hostIp, resourceType)
	return err
}

func (m *customHardwareResourceVerificationModel) HasRunningTask(ctx context.Context, hostIp, resourceType string) (bool, error) {
	query := `
		SELECT COUNT(*) as count
		FROM hardware_resource_verification 
		WHERE host_ip = ? 
		AND resource_type = ? 
		AND execution_status IN ('pending', 'running')
	`

	var count int
	err := m.conn.QueryRowCtx(ctx, &count, query, hostIp, resourceType)
	if err != nil {
		return false, err
	}
	if count > 0 {
		return true, nil
	}
	return false, nil
}
