package model

import (
	"context"
	"fmt"

	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

var _ DbLineModel = (*customDbLineModel)(nil)

type (
	// DbLineModel is an interface to be customized, add more methods here,
	// and implement the added methods in customDbLineModel.
	DbLineModel interface {
		dbLineModel
		FindDepartmentByGroupName(ctx context.Context, groupName string) (string, error)
	}

	customDbLineModel struct {
		*defaultDbLineModel
	}
)

// NewDbLineModel returns a model for the database table.
func NewDbLineModel(conn sqlx.SqlConn) DbLineModel {
	return &customDbLineModel{
		defaultDbLineModel: newDbLineModel(conn),
	}
}

// FindDepartmentByGroupName 根据集群组名称查询部门名称
func (m *customDbLineModel) FindDepartmentByGroupName(ctx context.Context, groupName string) (string, error) {
	query := `SELECT department_line_name FROM db_line WHERE cluster_group_name = ? LIMIT 1`

	var departmentName string
	err := m.conn.QueryRowCtx(ctx, &departmentName, query, groupName)
	if err != nil {
		return "", fmt.Errorf("查询部门名称失败: %w", err)
	}

	return departmentName, nil
}
