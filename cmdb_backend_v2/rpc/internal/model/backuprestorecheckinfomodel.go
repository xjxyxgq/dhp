package model

import (
	"context"
	"database/sql"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

var _ BackupRestoreCheckInfoModel = (*customBackupRestoreCheckInfoModel)(nil)

type (
	// BackupRestoreCheckInfoModel is an interface to be customized, add more methods here,
	// and implement the added methods in customBackupRestoreCheckInfoModel.
	BackupRestoreCheckInfoModel interface {
		backupRestoreCheckInfoModel
		FindOrderedByTime(ctx context.Context, limit int32) ([]*BackupRestoreCheckRecord, error)
	}

	// BackupRestoreCheckRecord 备份恢复检查记录
	BackupRestoreCheckRecord struct {
		Id                int64          `db:"id"`
		CheckSeq          string         `db:"check_seq"`
		CheckDb           string         `db:"check_db"`
		CheckSrcIP        sql.NullString `db:"check_src_ip"`
		BackupCheckResult sql.NullString `db:"backup_check_result"`
		CreatedAt         sql.NullString `db:"created_at"`
	}

	customBackupRestoreCheckInfoModel struct {
		*defaultBackupRestoreCheckInfoModel
	}
)

// NewBackupRestoreCheckInfoModel returns a model for the database table.
func NewBackupRestoreCheckInfoModel(conn sqlx.SqlConn) BackupRestoreCheckInfoModel {
	return &customBackupRestoreCheckInfoModel{
		defaultBackupRestoreCheckInfoModel: newBackupRestoreCheckInfoModel(conn),
	}
}

// FindOrderedByTime 按时间倒序查询备份恢复检查信息
func (m *customBackupRestoreCheckInfoModel) FindOrderedByTime(ctx context.Context, limit int32) ([]*BackupRestoreCheckRecord, error) {
	query := `SELECT id, check_seq, check_db, check_src_ip, backup_check_result, created_at 
			  FROM backup_restore_check_info 
			  ORDER BY created_at DESC 
			  LIMIT ?`

	var records []*BackupRestoreCheckRecord
	err := m.conn.QueryRowsCtx(ctx, &records, query, limit)
	return records, err
}
