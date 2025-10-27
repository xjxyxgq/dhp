package model

import (
	"context"
	"fmt"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
	"strings"
	"time"
)

var _ HostsPoolModel = (*customHostsPoolModel)(nil)

type (
	// HostsPoolModel is an interface to be customized, add more methods here,
	// and implement the added methods in customHostsPoolModel.
	HostsPoolModel interface {
		hostsPoolModel
		FindAllActiveHosts(ctx context.Context) ([]*HostInfo, error)
		FindByIP(ctx context.Context, ip string) (*HostsPool, error)
		InsertIfNotExists(ctx context.Context, hostName, hostIP, hostType string) (int64, error)
		InsertOrUpdateWithRemark(ctx context.Context, hostName, hostIP, hostType, remark string) (int64, error)
		FindHostsPoolDetailWithFilter(ctx context.Context, ipList []string) ([]*HostPoolDetailRow, error)
		FindAllHostIPs(ctx context.Context) ([]string, error)
		UpdateHostHardwareInfo(ctx context.Context, hostInfo *HostsPool) error
		UpdateHostIdcInfo(ctx context.Context, hostIp string, idcId int64) error
		FindAllHostsForIdcUpdate(ctx context.Context, hostIpList []string) ([]*HostsPool, error)
	}

	// HostInfo 用于监控数据核对的主机信息结构体
	HostInfo struct {
		HostIp     string `db:"host_ip"`
		HostName   string `db:"host_name"`
		PoolName   string `db:"pool_name"`
		CreateTime string `db:"create_time"`
	}

	// HostPoolDetailRow 主机池详情查询结果
	HostPoolDetailRow struct {
		Id              int64   `db:"id"`
		HostName        string  `db:"host_name"`
		HostIp          string  `db:"host_ip"`
		HostType        *string `db:"host_type"`
		H3cId           *string `db:"h3c_id"`
		H3cStatus       *string `db:"h3c_status"`
		DiskSize        int32   `db:"disk_size"`
		Ram             int32   `db:"ram"`
		Vcpus           int32   `db:"vcpus"`
		IfH3cSync       string  `db:"if_h3c_sync"`
		H3cImgId        string  `db:"h3c_img_id"`
		H3cHmName       string  `db:"h3c_hm_name"`
		LeafNumber      string  `db:"leaf_number"`
		RackNumber      string  `db:"rack_number"`
		RackHeight      int32   `db:"rack_height"`
		RackStartNumber int32   `db:"rack_start_number"`
		FromFactor      int32   `db:"from_factor"`
		SerialNumber    string  `db:"serial_number"`
		Remark          string  `db:"remark"`
		IsDeleted       bool    `db:"is_deleted"`
		IsStatic        bool    `db:"is_static"`
		CreateTime      string  `db:"create_time"`
		UpdateTime      string  `db:"update_time"`
	}

	customHostsPoolModel struct {
		*defaultHostsPoolModel
	}
)

// NewHostsPoolModel returns a model for the database table.
func NewHostsPoolModel(conn sqlx.SqlConn) HostsPoolModel {
	return &customHostsPoolModel{
		defaultHostsPoolModel: newHostsPoolModel(conn),
	}
}

// FindAllActiveHosts 获取所有未删除的主机信息
func (m *customHostsPoolModel) FindAllActiveHosts(ctx context.Context) ([]*HostInfo, error) {
	query := `
		SELECT host_ip, host_name, 'hosts_pool' as pool_name, create_time
		FROM hosts_pool 
		WHERE is_deleted = 0
		ORDER BY host_ip
	`

	var hosts []*HostInfo
	err := m.conn.QueryRowsCtx(ctx, &hosts, query)
	return hosts, err
}

// FindByIP 根据IP查找主机记录
func (m *customHostsPoolModel) FindByIP(ctx context.Context, ip string) (*HostsPool, error) {
	query := "SELECT " + hostsPoolRows + " FROM " + m.table + " WHERE host_ip = ? LIMIT 1"
	var resp HostsPool
	err := m.conn.QueryRowCtx(ctx, &resp, query, ip)
	if err != nil {
		return nil, err
	}
	return &resp, nil
}

// InsertIfNotExists 如果主机不存在则插入，返回主机ID
// 使用 INSERT IGNORE 避免并发插入时的重复键错误
func (m *customHostsPoolModel) InsertIfNotExists(ctx context.Context, hostName, hostIP, hostType string) (int64, error) {
	// 使用 INSERT IGNORE 避免重复键错误
	query := `
		INSERT IGNORE INTO hosts_pool (host_name, host_ip, host_type, create_time, update_time)
		VALUES (?, ?, ?, NOW(), NOW())
	`

	result, err := m.conn.ExecCtx(ctx, query, hostName, hostIP, hostType)
	if err != nil {
		return 0, err
	}

	// 获取插入的ID
	lastId, err := result.LastInsertId()
	if err != nil {
		return 0, err
	}

	// 如果 lastId 为 0，说明记录已存在（INSERT IGNORE 未插入）
	// 需要查询现有记录的ID
	if lastId == 0 {
		existingHost, err := m.FindByIP(ctx, hostIP)
		if err != nil {
			return 0, fmt.Errorf("查询现有主机失败: %v", err)
		}
		if existingHost == nil {
			return 0, fmt.Errorf("主机 %s 不存在但无法插入", hostIP)
		}
		return int64(existingHost.Id), nil
	}

	return lastId, nil
}

// InsertOrUpdateWithRemark 插入或更新主机信息（包含remark），返回主机ID
// 使用 INSERT ... ON DUPLICATE KEY UPDATE 避免并发时的竞态条件
func (m *customHostsPoolModel) InsertOrUpdateWithRemark(ctx context.Context, hostName, hostIP, hostType, remark string) (int64, error) {
	// 使用 ON DUPLICATE KEY UPDATE 确保原子性操作
	query := `
		INSERT INTO hosts_pool (host_name, host_ip, host_type, remark, create_time, update_time)
		VALUES (?, ?, ?, ?, NOW(), NOW())
		ON DUPLICATE KEY UPDATE remark = VALUES(remark), update_time = NOW()
	`

	result, err := m.conn.ExecCtx(ctx, query, hostName, hostIP, hostType, remark)
	if err != nil {
		return 0, err
	}

	// 获取插入或更新的ID
	lastId, err := result.LastInsertId()
	if err != nil {
		return 0, err
	}

	// 如果是更新操作，lastId 可能不准确，需要查询
	if lastId == 0 {
		existingHost, err := m.FindByIP(ctx, hostIP)
		if err != nil {
			return 0, fmt.Errorf("查询现有主机失败: %v", err)
		}
		if existingHost == nil {
			return 0, fmt.Errorf("主机 %s 不存在但无法插入", hostIP)
		}
		return int64(existingHost.Id), nil
	}

	return lastId, nil
}

// FindHostsPoolDetailWithFilter 查询主机池详细信息（支持IP过滤）
func (m *customHostsPoolModel) FindHostsPoolDetailWithFilter(ctx context.Context, ipList []string) ([]*HostPoolDetailRow, error) {
	// 构建查询主机基本信息的SQL
	query := `SELECT id, host_name, host_ip, host_type,
       COALESCE(h3c_id, "") as h3c_id,
       COALESCE(h3c_status, "") as h3c_status,
	   COALESCE(disk_size, 0) as disk_size,
	   COALESCE(ram, 0) as ram,
	   COALESCE(vcpus, 0) as vcpus,
	   COALESCE(if_h3c_sync, '') as if_h3c_sync,
	   COALESCE(h3c_img_id, '') as h3c_img_id,
	   COALESCE(h3c_hm_name, '') as h3c_hm_name,
	   COALESCE(leaf_number, '') as leaf_number,
	   COALESCE(rack_number, '') as rack_number,
	   COALESCE(rack_height, 0) as rack_height,
	   COALESCE(rack_start_number, 0) as rack_start_number,
	   COALESCE(from_factor, 0) as from_factor,
	   COALESCE(serial_number, '') as serial_number,
	   COALESCE(remark, '') as remark,
	   is_deleted, is_static, create_time, update_time
	   FROM hosts_pool
	   WHERE is_deleted = 0`

	args := []interface{}{}
	if len(ipList) > 0 {
		placeholders := make([]string, len(ipList))
		for i, ip := range ipList {
			placeholders[i] = "?"
			args = append(args, ip)
		}
		query += " AND host_ip IN (" + strings.Join(placeholders, ",") + ")"
	}

	query += " ORDER BY id"

	var hostRows []*HostPoolDetailRow
	err := m.conn.QueryRowsCtx(ctx, &hostRows, query, args...)
	return hostRows, err
}

// FindAllHostIPs 获取所有未删除主机的IP列表
func (m *customHostsPoolModel) FindAllHostIPs(ctx context.Context) ([]string, error) {
	query := "SELECT host_ip FROM hosts_pool WHERE is_deleted = 0"
	var hosts []struct {
		HostIp string `db:"host_ip"`
	}
	err := m.conn.QueryRowsCtx(ctx, &hosts, query)
	if err != nil {
		return nil, err
	}

	var hostIpList []string
	for _, host := range hosts {
		hostIpList = append(hostIpList, host.HostIp)
	}
	return hostIpList, nil
}

// UpdateHostHardwareInfo 更新主机信息（只更新非默认空值的字段）,对于IsDeleted、IsStatic两个字段，不想更新值时需要设置为负数
func (m *customHostsPoolModel) UpdateHostHardwareInfo(ctx context.Context, hostInfo *HostsPool) error {
	if hostInfo == nil || hostInfo.HostIp == "" || hostInfo.Id != 0 {
		return fmt.Errorf("hostInfo 不能为空且必须包含 host_ip 或有效的 id")
	}

	updateFields := []string{}
	updateValues := []interface{}{}

	// 更新主机名（如果提供且不为空）
	if hostInfo.HostName != "" {
		updateFields = append(updateFields, "host_name = ?")
		updateValues = append(updateValues, hostInfo.HostName)
	}

	// 更新主机类型（如果提供且有效）
	if hostInfo.HostType.Valid && hostInfo.HostType.String != "" {
		updateFields = append(updateFields, "host_type = ?")
		updateValues = append(updateValues, hostInfo.HostType.String)
	}

	// 更新硬件信息（如果提供且有效）
	if hostInfo.DiskSize.Valid && hostInfo.DiskSize.Int64 > 0 {
		updateFields = append(updateFields, "disk_size = ?")
		updateValues = append(updateValues, hostInfo.DiskSize.Int64)
	}
	if hostInfo.Ram.Valid && hostInfo.Ram.Int64 > 0 {
		updateFields = append(updateFields, "ram = ?")
		updateValues = append(updateValues, hostInfo.Ram.Int64)
	}
	if hostInfo.Vcpus.Valid && hostInfo.Vcpus.Int64 > 0 {
		updateFields = append(updateFields, "vcpus = ?")
		updateValues = append(updateValues, hostInfo.Vcpus.Int64)
	}

	// 更新H3C相关信息（如果提供且有效）
	if hostInfo.H3cId.Valid && hostInfo.H3cId.String != "" {
		updateFields = append(updateFields, "h3c_id = ?")
		updateValues = append(updateValues, hostInfo.H3cId.String)
	}
	if hostInfo.H3cStatus.Valid && hostInfo.H3cStatus.String != "" {
		updateFields = append(updateFields, "h3c_status = ?")
		updateValues = append(updateValues, hostInfo.H3cStatus.String)
	}
	if hostInfo.IfH3cSync.Valid && hostInfo.IfH3cSync.String != "" {
		updateFields = append(updateFields, "if_h3c_sync = ?")
		updateValues = append(updateValues, hostInfo.IfH3cSync.String)
	}
	if hostInfo.H3cImgId.Valid && hostInfo.H3cImgId.String != "" {
		updateFields = append(updateFields, "h3c_img_id = ?")
		updateValues = append(updateValues, hostInfo.H3cImgId.String)
	}
	if hostInfo.H3cHmName.Valid && hostInfo.H3cHmName.String != "" {
		updateFields = append(updateFields, "h3c_hm_name = ?")
		updateValues = append(updateValues, hostInfo.H3cHmName.String)
	}

	// 更新机架信息（如果提供且有效）
	if hostInfo.LeafNumber.Valid && hostInfo.LeafNumber.String != "" {
		updateFields = append(updateFields, "leaf_number = ?")
		updateValues = append(updateValues, hostInfo.LeafNumber.String)
	}
	if hostInfo.RackNumber.Valid && hostInfo.RackNumber.String != "" {
		updateFields = append(updateFields, "rack_number = ?")
		updateValues = append(updateValues, hostInfo.RackNumber.String)
	}
	if hostInfo.RackHeight.Valid && hostInfo.RackHeight.Int64 > 0 {
		updateFields = append(updateFields, "rack_height = ?")
		updateValues = append(updateValues, hostInfo.RackHeight.Int64)
	}
	if hostInfo.RackStartNumber.Valid && hostInfo.RackStartNumber.Int64 >= 0 {
		updateFields = append(updateFields, "rack_start_number = ?")
		updateValues = append(updateValues, hostInfo.RackStartNumber.Int64)
	}
	if hostInfo.FromFactor.Valid && hostInfo.FromFactor.Int64 > 0 {
		updateFields = append(updateFields, "from_factor = ?")
		updateValues = append(updateValues, hostInfo.FromFactor.Int64)
	}
	if hostInfo.SerialNumber.Valid && hostInfo.SerialNumber.String != "" {
		updateFields = append(updateFields, "serial_number = ?")
		updateValues = append(updateValues, hostInfo.SerialNumber.String)
	}

	// 更新删除标记（如果提供）
	if hostInfo.IsDeleted >= 0 {
		updateFields = append(updateFields, "is_deleted = ?")
		updateValues = append(updateValues, hostInfo.IsDeleted)
	}

	if hostInfo.IsStatic >= 0 {
		updateFields = append(updateFields, "is_static = ?")
		updateValues = append(updateValues, hostInfo.IsStatic)
	}

	if len(updateFields) == 0 {
		return fmt.Errorf("没有需要更新的信息")
	}

	// 添加更新时间
	updateFields = append(updateFields, "update_time = ?")
	updateValues = append(updateValues, time.Now())

	// 添加WHERE条件的参数
	updateValues = append(updateValues, hostInfo.HostIp)

	// 构建查询语句
	query := "UPDATE hosts_pool SET "
	for i, field := range updateFields {
		if i > 0 {
			query += ", "
		}
		query += field
	}
	query += " WHERE host_ip = ?"

	_, err := m.conn.ExecCtx(ctx, query, updateValues...)
	if err != nil {
		return fmt.Errorf("执行更新SQL失败: %v", err)
	}

	return nil
}

// UpdateHostIdcInfo 更新主机的IDC信息
func (m *customHostsPoolModel) UpdateHostIdcInfo(ctx context.Context, hostIp string, idcId int64) error {
	query := "UPDATE hosts_pool SET idc_id = ?, update_time = ? WHERE host_ip = ?"
	_, err := m.conn.ExecCtx(ctx, query, idcId, time.Now(), hostIp)
	if err != nil {
		return fmt.Errorf("更新主机IDC信息失败: %v", err)
	}
	return nil
}

// FindAllHostsForIdcUpdate 查找需要更新IDC信息的主机列表
func (m *customHostsPoolModel) FindAllHostsForIdcUpdate(ctx context.Context, hostIpList []string) ([]*HostsPool, error) {
	var query string
	var args []interface{}

	if len(hostIpList) > 0 {
		// 查询指定IP的主机
		placeholders := strings.Repeat("?,", len(hostIpList))
		placeholders = placeholders[:len(placeholders)-1] // 去掉最后一个逗号
		query = fmt.Sprintf("select %s from %s where `is_deleted` = 0 and `host_ip` in (%s)", hostsPoolRows, m.table, placeholders)
		for _, ip := range hostIpList {
			args = append(args, ip)
		}
	} else {
		// 查询所有未删除的主机
		query = fmt.Sprintf("select %s from %s where `is_deleted` = 0", hostsPoolRows, m.table)
	}

	var resp []*HostsPool
	err := m.conn.QueryRowsCtx(ctx, &resp, query, args...)
	switch err {
	case nil:
		return resp, nil
	default:
		return nil, err
	}
}
