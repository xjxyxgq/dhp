package model

import (
	"context"
	"database/sql"
	"fmt"
	"strings"

	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

var _ HostsApplicationsModel = (*customHostsApplicationsModel)(nil)

type (
	// HostsApplicationsModel is an interface to be customized, add more methods here,
	// and implement the added methods in customHostsApplicationsModel.
	HostsApplicationsModel interface {
		hostsApplicationsModel
		FindByPoolId(ctx context.Context, poolId uint64) ([]*HostsApplications, error)
		FindByPoolIdAndCluster(ctx context.Context, poolId int64, serverType string, serverPort int32, clusterName string) (*HostsApplications, error)
		UpsertApplication(ctx context.Context, poolId int64, serverType, version, clusterName, protocol, addr, role, status, departmentName string, port int32) error
		FindByPoolIds(ctx context.Context, poolIds []string) ([]*ApplicationRow, error)
		ExistsByHostIdAndCluster(ctx context.Context, hostId int64, clusterName string) (bool, error)
		FindClustersByServerAddr(ctx context.Context, serverAddr string) ([]*HostClusterInfo, error)
	}

	// ApplicationRow 用于接收应用查询结果
	ApplicationRow struct {
		Id             int64          `db:"id"`         // 应用ID
		PoolId         int64          `db:"pool_id"`
		ServerType     sql.NullString `db:"server_type"`
		ServerVersion  sql.NullString `db:"server_version"`
		ServerSubtitle sql.NullString `db:"server_subtitle"`
		ClusterName    sql.NullString `db:"cluster_name"`
		ServerProtocol sql.NullString `db:"server_protocol"`
		ServerAddr     sql.NullString `db:"server_addr"`
		ServerPort     sql.NullInt32  `db:"server_port"`
		ServerRole     sql.NullString `db:"server_role"`
		ServerStatus   sql.NullString `db:"server_status"`
		DepartmentName sql.NullString `db:"department_name"`
	}

	// HostClusterInfo 主机集群信息
	HostClusterInfo struct {
		ClusterName     string `db:"cluster_name"`
		ClusterGroupName string `db:"cluster_group_name"`
	}

	customHostsApplicationsModel struct {
		*defaultHostsApplicationsModel
	}
)

// NewHostsApplicationsModel returns a model for the database table.
func NewHostsApplicationsModel(conn sqlx.SqlConn) HostsApplicationsModel {
	return &customHostsApplicationsModel{
		defaultHostsApplicationsModel: newHostsApplicationsModel(conn),
	}
}

// FindByPoolId 根据主机池ID查找关联的应用
func (m *customHostsApplicationsModel) FindByPoolId(ctx context.Context, poolId uint64) ([]*HostsApplications, error) {
	var resp []*HostsApplications
	query := fmt.Sprintf("select %s from %s where `pool_id` = ?", hostsApplicationsRows, m.table)
	err := m.conn.QueryRowsCtx(ctx, &resp, query, poolId)
	return resp, err
}

// FindByPoolIdAndCluster 查找特定主机和集群的应用记录
func (m *customHostsApplicationsModel) FindByPoolIdAndCluster(ctx context.Context, poolId int64, serverType string, serverPort int32, clusterName string) (*HostsApplications, error) {
	query := fmt.Sprintf("select %s from %s where `pool_id` = ? AND `server_type` = ? AND `server_port` = ? AND `cluster_name` = ? LIMIT 1", hostsApplicationsRows, m.table)
	var resp HostsApplications
	err := m.conn.QueryRowCtx(ctx, &resp, query, poolId, serverType, serverPort, clusterName)
	if err != nil {
		return nil, err
	}
	return &resp, nil
}

// UpsertApplication 插入或更新应用记录
func (m *customHostsApplicationsModel) UpsertApplication(ctx context.Context, poolId int64, serverType, version, clusterName, protocol, addr, role, status, departmentName string, port int32) error {
	// 先尝试查找现有记录
	existing, err := m.FindByPoolIdAndCluster(ctx, poolId, serverType, port, clusterName)
	if err == nil && existing != nil {
		// 记录存在，更新记录
		updateQuery := `
			UPDATE hosts_applications 
			SET server_version = ?, server_role = ?, server_status = ?, department_name = ?, update_time = NOW()
			WHERE id = ?
		`
		_, err = m.conn.ExecCtx(ctx, updateQuery, version, role, status, departmentName, existing.Id)
		return err
	}

	// 记录不存在，插入新记录
	insertQuery := `
		INSERT INTO hosts_applications (
			pool_id, server_type, server_version, cluster_name, 
			server_protocol, server_addr, server_port, server_role, 
			server_status, department_name, create_time, update_time
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, NOW(), NOW())
	`
	
	_, err = m.conn.ExecCtx(ctx, insertQuery,
		poolId, serverType, version, clusterName,
		protocol, addr, port, role,
		status, departmentName)
	
	return err
}

// FindByPoolIds 根据主机池ID列表查询应用信息
func (m *customHostsApplicationsModel) FindByPoolIds(ctx context.Context, poolIds []string) ([]*ApplicationRow, error) {
	// 查询应用信息
	query := `SELECT id, pool_id, server_type, server_version, server_subtitle, 
			 cluster_name, server_protocol, server_addr, server_port, 
			 server_role, server_status, department_name
			 FROM hosts_applications 
			 WHERE pool_id IN (` + strings.Join(poolIds, ",") + `)
			 ORDER BY pool_id, id`

	var appRows []*ApplicationRow
	err := m.conn.QueryRowsCtx(ctx, &appRows, query)
	return appRows, err
}

// ExistsByHostIdAndCluster 检查主机和集群的应用记录是否存在
func (m *customHostsApplicationsModel) ExistsByHostIdAndCluster(ctx context.Context, hostId int64, clusterName string) (bool, error) {
	query := `SELECT COUNT(*) as count FROM hosts_applications WHERE pool_id = ? AND cluster_name = ?`
	
	var count int
	err := m.conn.QueryRowCtx(ctx, &count, query, hostId, clusterName)
	if err != nil {
		return false, err
	}
	
	return count > 0, nil
}

// FindClustersByServerAddr 根据服务器地址查询该主机所属的所有集群
func (m *customHostsApplicationsModel) FindClustersByServerAddr(ctx context.Context, serverAddr string) ([]*HostClusterInfo, error) {
	query := `
		SELECT DISTINCT 
			ha.cluster_name, 
			COALESCE(cg.group_name, '') as cluster_group_name
		FROM hosts_applications ha
		LEFT JOIN cluster_groups cg ON ha.cluster_name = cg.cluster_name
		WHERE ha.server_addr = ?
	`

	var clusters []*HostClusterInfo
	err := m.conn.QueryRowsCtx(ctx, &clusters, query, serverAddr)
	return clusters, err
}
