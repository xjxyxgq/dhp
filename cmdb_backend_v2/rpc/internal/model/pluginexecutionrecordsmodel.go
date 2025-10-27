package model

import (
	"context"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

var _ PluginExecutionRecordsModel = (*customPluginExecutionRecordsModel)(nil)

type (
	// PluginExecutionRecordsModel is an interface to be customized, add more methods here,
	// and implement the added methods in customPluginExecutionRecordsModel.
	PluginExecutionRecordsModel interface {
		pluginExecutionRecordsModel
		FindRecentRecords(ctx context.Context, days int32) ([]*PluginRecord, error)
	}

	// PluginRecord 插件执行记录结构体
	PluginRecord struct {
		CheckSeq   string `db:"check_seq"`
		PluginName string `db:"plugin_name"`
		Result     string `db:"result"`
	}

	customPluginExecutionRecordsModel struct {
		*defaultPluginExecutionRecordsModel
	}
)

// NewPluginExecutionRecordsModel returns a model for the database table.
func NewPluginExecutionRecordsModel(conn sqlx.SqlConn) PluginExecutionRecordsModel {
	return &customPluginExecutionRecordsModel{
		defaultPluginExecutionRecordsModel: newPluginExecutionRecordsModel(conn),
	}
}

// FindRecentRecords 查询最近几天的插件执行记录
func (m *customPluginExecutionRecordsModel) FindRecentRecords(ctx context.Context, days int32) ([]*PluginRecord, error) {
	query := `SELECT check_seq, plugin_name, result 
			  FROM plugin_execution_records 
			  WHERE created_at >= DATE_SUB(NOW(), INTERVAL ? DAY)
			  ORDER BY created_at DESC`

	var records []*PluginRecord
	err := m.conn.QueryRowsCtx(ctx, &records, query, days)
	return records, err
}
