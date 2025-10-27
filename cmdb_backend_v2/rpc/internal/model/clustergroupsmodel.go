package model

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

var _ ClusterGroupsModel = (*customClusterGroupsModel)(nil)

type (
	// ClusterGroupsModel is an interface to be customized, add more methods here,
	// and implement the added methods in customClusterGroupsModel.
	ClusterGroupsModel interface {
		clusterGroupsModel
		FindByClusterName(ctx context.Context, clusterName string) (*ClusterGroups, error)
		FindAllClusterGroups(ctx context.Context) ([]*ClusterGroupRow, error)
		CheckClusterGroupExists(ctx context.Context, groupName, clusterName, clusterType string) (bool, error)
		InsertClusterGroup(ctx context.Context, groupName, clusterType, clusterName, departmentName string) error
		UpdateClusterGroup(ctx context.Context, groupName, clusterType, clusterName, departmentName string) error
	}

	// ClusterGroupRow 用于接收数据库查询结果
	ClusterGroupRow struct {
		Id                 int64          `db:"id"`
		CreateTime         string         `db:"create_time"`
		UpdateTime         string         `db:"update_time"`
		GroupName          sql.NullString `db:"group_name"`
		ClusterType        sql.NullString `db:"cluster_type"`
		ClusterName        sql.NullString `db:"cluster_name"`
		DepartmentLineName sql.NullString `db:"department_line_name"`
	}

	customClusterGroupsModel struct {
		*defaultClusterGroupsModel
	}
)

// NewClusterGroupsModel returns a model for the database table.
func NewClusterGroupsModel(conn sqlx.SqlConn) ClusterGroupsModel {
	return &customClusterGroupsModel{
		defaultClusterGroupsModel: newClusterGroupsModel(conn),
	}
}

// FindByClusterName 根据集群名称查找集群组信息
func (m *customClusterGroupsModel) FindByClusterName(ctx context.Context, clusterName string) (*ClusterGroups, error) {
	var resp ClusterGroups
	query := fmt.Sprintf("select %s from %s where `cluster_name` = ? limit 1", clusterGroupsRows, m.table)
	err := m.conn.QueryRowCtx(ctx, &resp, query, clusterName)
	switch err {
	case nil:
		return &resp, nil
	case sqlx.ErrNotFound:
		return nil, ErrNotFound
	default:
		return nil, err
	}
}

// FindAllClusterGroups 查询所有集群组
func (m *customClusterGroupsModel) FindAllClusterGroups(ctx context.Context) ([]*ClusterGroupRow, error) {
	query := `SELECT id, create_time, update_time, group_name, cluster_type, cluster_name, department_line_name 
			  FROM cluster_groups 
			  ORDER BY id`

	var rows []*ClusterGroupRow
	err := m.conn.QueryRowsCtx(ctx, &rows, query)
	if err != nil {
		return nil, fmt.Errorf("执行查询失败: %w", err)
	}

	return rows, nil
}

// CheckClusterGroupExists 检查cluster_groups表中是否已存在指定记录
func (m *customClusterGroupsModel) CheckClusterGroupExists(ctx context.Context, groupName, clusterName, clusterType string) (bool, error) {
	query := `SELECT COUNT(*) FROM cluster_groups WHERE group_name = ? AND cluster_name = ? AND cluster_type = ?`

	var count int
	err := m.conn.QueryRowCtx(ctx, &count, query, groupName, clusterName, clusterType)
	if err != nil {
		return false, fmt.Errorf("检查记录存在性失败: %w", err)
	}

	return count > 0, nil
}

// InsertClusterGroup 向cluster_groups表插入新记录
func (m *customClusterGroupsModel) InsertClusterGroup(ctx context.Context, groupName, clusterType, clusterName, departmentName string) error {
	insertQuery := `INSERT INTO cluster_groups (group_name, cluster_type, cluster_name, department_line_name, create_time, update_time) 
		VALUES (?, ?, ?, ?, NOW(), NOW())`

	_, err := m.conn.ExecCtx(ctx, insertQuery, groupName, clusterType, clusterName, departmentName)
	if err != nil {
		return fmt.Errorf("插入记录失败: %w", err)
	}

	return nil
}

// UpdateClusterGroup 更新cluster_groups表中的记录
func (m *customClusterGroupsModel) UpdateClusterGroup(ctx context.Context, groupName, clusterType, clusterName, departmentName string) error {
	updateQuery := `UPDATE cluster_groups SET cluster_name = ?, department_line_name = ?, update_time = NOW() 
		WHERE group_name = ? AND cluster_name = ? AND cluster_type = ?`

	_, err := m.conn.ExecCtx(ctx, updateQuery, clusterName, departmentName, groupName, clusterName, clusterType)
	if err != nil {
		return fmt.Errorf("更新记录失败: %w", err)
	}

	return nil
}
