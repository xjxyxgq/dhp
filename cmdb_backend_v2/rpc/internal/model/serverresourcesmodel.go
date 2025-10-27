package model

import (
	"context"
	"database/sql"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
	"strings"
)

var _ ServerResourcesModel = (*customServerResourcesModel)(nil)

type (
	// ServerResourcesModel is an interface to be customized, add more methods here,
	// and implement the added methods in customServerResourcesModel.
	ServerResourcesModel interface {
		serverResourcesModel
		FindDistinctIPsInTimeRange(ctx context.Context, startTime, endTime string) ([]*MonitoringIP, error)
		FindDiskPredictionDataWithFilter(ctx context.Context, filter *DiskPredictionFilter) ([]*DiskPredictionData, error)
		FindDiskPredictionData(ctx context.Context, beginTime, endTime string, ipList []string) ([]*DiskPredictionData, error)
		FindDiskPredictionDataByCluster(ctx context.Context, beginTime, endTime string, clusterName string) ([]*DiskPredictionData, error)
		FindDiskPredictionDataByDepartment(ctx context.Context, beginTime, endTime string, departmentName string) ([]*DiskPredictionData, error)
		FindServerResourceMax(ctx context.Context, beginTime, endTime string, ipList []string, clusterName string) ([]*ServerResourceMaxData, error)
		FindServerResourcesWithFilter(ctx context.Context, beginTime, endTime string, ipList []string) ([]*ServerResourceRow, error)
		//FindServerResourceMaxByCluster(ctx context.Context, beginTime, endTime string, clusterName string) ([]*ServerResourceMaxData, error)
		FindServerResourcesWithClusterFilter(ctx context.Context, beginTime, endTime string, clusterName string) ([]*ServerResourceRow, error)
		FindClusterResources(ctx context.Context, beginTime, endTime, clusterName, groupName string) ([]*ClusterMemberResourceData, error)
		FindClusterResourcesMax(ctx context.Context, beginTime, endTime, clusterName, groupName string) ([]*ClusterResourceMaxData, error)
		FindClusterMemberResources(ctx context.Context, clusterName string) ([]*ClusterMemberResourceData, error)
		// ES数据同步相关方法 - 接受绝对值和百分比参数
		UpsertFromES(ctx context.Context, poolId uint64, ip string,
			usedMemory, usedDisk, cpuLoad float64,
			cpuPercentMax, cpuPercentAvg, cpuPercentMin,
			memPercentMax, memPercentAvg, memPercentMin,
			diskPercentMax, diskPercentAvg, diskPercentMin float64) error
		// CMSys数据同步相关方法 - 接受百分比参数
		UpsertFromCMSys(ctx context.Context, poolId uint64, ip string,
			cpuPercent, memPercent, diskPercent float64) error
	}
	// MonitoringIP 用于存储有监控数据的IP
	MonitoringIP struct {
		IP string `db:"ip"`
	}

	// DiskPredictionFilter 磁盘预测数据查询过滤条件
	DiskPredictionFilter struct {
		BeginTime      string   // 开始时间
		EndTime        string   // 结束时间
		IPList         []string // IP列表筛选
		ClusterName    string   // 集群名称筛选
		DepartmentName string   // 部门名称筛选
	}

	// ServerResourceRow 用于接收数据库查询结果（关联查询后的结果）
	ServerResourceRow struct {
		Id             int64           `db:"id"`
		CreateTime     string          `db:"create_time"`
		UpdateTime     string          `db:"update_time"`
		PoolId         sql.NullInt64   `db:"pool_id"`
		ClusterName    sql.NullString  `db:"cluster_name"` // 通过 LEFT JOIN 获取
		GroupName      sql.NullString  `db:"group_name"`   // 通过 LEFT JOIN 获取
		Ip             sql.NullString  `db:"ip"`
		Port           sql.NullInt32   `db:"port"`
		InstanceRole   sql.NullString  `db:"instance_role"` // 通过 LEFT JOIN 获取
		TotalMemory    sql.NullFloat64 `db:"total_memory"`
		UsedMemory     sql.NullFloat64 `db:"used_memory"`
		TotalDisk      sql.NullFloat64 `db:"total_disk"`
		UsedDisk       sql.NullFloat64 `db:"used_disk"`
		CPUCores       sql.NullInt32   `db:"cpu_cores"`
		CPULoad        sql.NullFloat64 `db:"cpu_load"`
		MonDate        string          `db:"mon_date"`
		DepartmentName sql.NullString  `db:"department_name"` // 通过 LEFT JOIN 获取
		// 百分比字段
		CpuPercentMax  sql.NullFloat64 `db:"cpu_percent_max"`
		CpuPercentAvg  sql.NullFloat64 `db:"cpu_percent_avg"`
		CpuPercentMin  sql.NullFloat64 `db:"cpu_percent_min"`
		MemPercentMax  sql.NullFloat64 `db:"mem_percent_max"`
		MemPercentAvg  sql.NullFloat64 `db:"mem_percent_avg"`
		MemPercentMin  sql.NullFloat64 `db:"mem_percent_min"`
		DiskPercentMax sql.NullFloat64 `db:"disk_percent_max"`
		DiskPercentAvg sql.NullFloat64 `db:"disk_percent_avg"`
		DiskPercentMin sql.NullFloat64 `db:"disk_percent_min"`
	}

	// DiskPredictionData 磁盘预测数据结构
	DiskPredictionData struct {
		ID             int64   `db:"id"`
		Ip             string  `db:"ip"`
		ClusterName    string  `db:"cluster_name"`
		GroupName      string  `db:"group_name"`
		DepartmentName string  `db:"department_line_name"`
		TotalDisk      float32 `db:"total_disk"`
		UsedDisk       float32 `db:"used_disk"`
		DiskPercentMax float32 `db:"disk_percent_max"` // 磁盘使用百分比
		MonDate        string  `db:"mon_date"`
		// IDC 信息
		IdcId          sql.NullInt64  `db:"idc_id"`
		IdcName        sql.NullString `db:"idc_name"`
		IdcCode        sql.NullString `db:"idc_code"`
		IdcLocation    sql.NullString `db:"idc_location"`
		IdcDescription sql.NullString `db:"idc_description"`
	}

	// ServerResourceMaxData 服务器资源最大值数据结构
	ServerResourceMaxData struct {
		Id             int64   `db:"id"`
		CreateTime     string  `db:"create_time"`
		UpdateTime     string  `db:"update_time"`
		PoolId         int64   `db:"pool_id"`
		ClusterName    string  `db:"cluster_name"` // 通过 LEFT JOIN 获取
		GroupName      string  `db:"group_name"`   // 通过 LEFT JOIN 获取
		DepartmentName string  `db:"department_line_name"`
		Ip             string  `db:"ip"`
		TotalMemory    float64 `db:"total_memory"`
		MaxUsedMemory  float64 `db:"max_used_memory"`
		TotalDisk      float64 `db:"total_disk"`
		MaxUsedDisk    float64 `db:"max_used_disk"`
		CPUCores       int32   `db:"cpu_cores"`
		MaxCPULoad     float64 `db:"max_cpu_load"`
		MaxDatetime    string  `db:"max_datetime"`
		HostName       string  `db:"host_name"` // 通过 LEFT JOIN 获取
		HostType       string  `db:"host_type"` // 通过 LEFT JOIN 获取
		// 百分比字段
		CpuPercentMax  float64 `db:"cpu_percent_max"`
		CpuPercentAvg  float64 `db:"cpu_percent_avg"`
		CpuPercentMin  float64 `db:"cpu_percent_min"`
		MemPercentMax  float64 `db:"mem_percent_max"`
		MemPercentAvg  float64 `db:"mem_percent_avg"`
		MemPercentMin  float64 `db:"mem_percent_min"`
		DiskPercentMax float64 `db:"disk_percent_max"`
		DiskPercentAvg float64 `db:"disk_percent_avg"`
		DiskPercentMin float64 `db:"disk_percent_min"`
	}

	// ClusterMemberResourceData 集群成员资源数据结构
	ClusterMemberResourceData struct {
		Id               int64           `db:"id"`
		ClusterName      string          `db:"cluster_name"`
		ClusterGroupName string          `db:"group_name"`
		Ip               string          `db:"ip"`
		HostName         string          `db:"host_name"`
		Port             int32           `db:"port"`
		InstanceRole     string          `db:"instance_role"`
		TotalMemory      sql.NullFloat64 `db:"total_memory"`
		UsedMemory       sql.NullFloat64 `db:"used_memory"`
		TotalDisk        sql.NullFloat64 `db:"total_disk"`
		UsedDisk         sql.NullFloat64 `db:"used_disk"`
		CPUCores         sql.NullInt32   `db:"cpu_cores"`
		CPULoad          sql.NullFloat64 `db:"cpu_load"`
		MonDate          string          `db:"mon_date"`
		DepartmentName   string          `db:"department_name"`
		// 百分比字段
		CpuPercentMax  sql.NullFloat64 `db:"cpu_percent_max"`
		CpuPercentAvg  sql.NullFloat64 `db:"cpu_percent_avg"`
		CpuPercentMin  sql.NullFloat64 `db:"cpu_percent_min"`
		MemPercentMax  sql.NullFloat64 `db:"mem_percent_max"`
		MemPercentAvg  sql.NullFloat64 `db:"mem_percent_avg"`
		MemPercentMin  sql.NullFloat64 `db:"mem_percent_min"`
		DiskPercentMax sql.NullFloat64 `db:"disk_percent_max"`
		DiskPercentAvg sql.NullFloat64 `db:"disk_percent_avg"`
		DiskPercentMin sql.NullFloat64 `db:"disk_percent_min"`
	}

	// ClusterResourceMaxData 集群资源最大值数据结构
	ClusterResourceMaxData struct {
		ClusterName      string `db:"cluster_name"`
		ClusterGroupName string `db:"group_name"`
		DepartmentName   string `db:"department_name"`
		NodeCount        int32  `db:"node_count"`
		// 平均值字段
		AvgCPULoad     sql.NullFloat64 `db:"avg_cpu_load"`
		AvgMemoryUsage sql.NullFloat64 `db:"avg_memory_usage"`
		AvgDiskUsage   sql.NullFloat64 `db:"avg_disk_usage"`
		// 最大值字段
		MaxCPULoad     sql.NullFloat64 `db:"max_cpu_load"`
		MaxMemoryUsage sql.NullFloat64 `db:"max_memory_usage"`
		MaxDiskUsage   sql.NullFloat64 `db:"max_disk_usage"`
		MaxDateTime    sql.NullString  `db:"max_datetime"`
		TotalMemory    sql.NullFloat64 `db:"total_memory"`
		TotalDisk      sql.NullFloat64 `db:"total_disk"`
		CPUCores       sql.NullInt32   `db:"cpu_cores"`
		BusLines       sql.NullString  `db:"bus_lines"`
		// 百分比字段
		CpuPercentMax  sql.NullFloat64 `db:"cpu_percent_max"`
		CpuPercentAvg  sql.NullFloat64 `db:"cpu_percent_avg"`
		CpuPercentMin  sql.NullFloat64 `db:"cpu_percent_min"`
		MemPercentMax  sql.NullFloat64 `db:"mem_percent_max"`
		MemPercentAvg  sql.NullFloat64 `db:"mem_percent_avg"`
		MemPercentMin  sql.NullFloat64 `db:"mem_percent_min"`
		DiskPercentMax sql.NullFloat64 `db:"disk_percent_max"`
		DiskPercentAvg sql.NullFloat64 `db:"disk_percent_avg"`
		DiskPercentMin sql.NullFloat64 `db:"disk_percent_min"`
	}

	customServerResourcesModel struct {
		*defaultServerResourcesModel
	}
)

// NewServerResourcesModel returns a model for the database table.
func NewServerResourcesModel(conn sqlx.SqlConn) ServerResourcesModel {
	return &customServerResourcesModel{
		defaultServerResourcesModel: newServerResourcesModel(conn),
	}
}

// FindDistinctIPsInTimeRange 查询指定时间范围内有监控数据的不重复IP
func (m *customServerResourcesModel) FindDistinctIPsInTimeRange(ctx context.Context, startTime, endTime string) ([]*MonitoringIP, error) {
	query := `
		SELECT DISTINCT ip
		FROM server_resources
		WHERE mon_date >= ? AND mon_date <= ?
	`

	var ips []*MonitoringIP
	err := m.conn.QueryRowsCtx(ctx, &ips, query, startTime, endTime)
	return ips, err
}

// FindDiskPredictionData 查询磁盘预测数据（通过IP列表筛选）
func (m *customServerResourcesModel) FindDiskPredictionData(ctx context.Context, beginTime, endTime string, ipList []string) ([]*DiskPredictionData, error) {
	filter := &DiskPredictionFilter{
		BeginTime: beginTime,
		EndTime:   endTime,
		IPList:    ipList,
	}
	return m.FindDiskPredictionDataWithFilter(ctx, filter)
}

// FindServerResourceMax 查询服务器资源最大值数据
func (m *customServerResourcesModel) FindServerResourceMax(ctx context.Context, beginTime, endTime string, ipList []string, clusterName string) ([]*ServerResourceMaxData, error) {
	args := []interface{}{}

	// 构建时间过滤的JOIN条件（移到ON子句，避免过滤掉没有资源记录的主机）
	timeJoinCondition := ""
	if beginTime != "" && endTime != "" {
		timeJoinCondition = " AND sr.mon_date BETWEEN ? AND ?"
		args = append(args, beginTime, endTime)
	} else if beginTime != "" {
		timeJoinCondition = " AND sr.mon_date >= ?"
		args = append(args, beginTime)
	} else if endTime != "" {
		timeJoinCondition = " AND sr.mon_date <= ?"
		args = append(args, endTime)
	}

	// 构建WHERE子句（仅包含主机和集群过滤条件）
	whereClause := " WHERE hp.is_deleted = 0"

	// 添加IP过滤
	if len(ipList) > 0 {
		placeholders := make([]string, len(ipList))
		for i, ip := range ipList {
			placeholders[i] = "?"
			args = append(args, ip)
		}
		whereClause += " AND hp.host_ip IN (" + strings.Join(placeholders, ",") + ")"
	}

	// 添加集群名称过滤（如果指定）
	if clusterName != "" {
		whereClause += " AND ha.cluster_name = ?"
		args = append(args, clusterName)
	}

	// 修改查询策略：按主机IP分组，每个集群信息作为独立记录返回
	// 这样可以保证每个主机的每个集群都有一条记录，便于在logic层进行合并
	// 关键修复：将时间过滤条件移到 LEFT JOIN ON 子句中，确保没有资源记录的主机也会返回
	hostResourceQuery := `SELECT
	hp.id AS id,
	COALESCE ( MIN( sr.create_time ), '' ) AS create_time,
	COALESCE ( MAX( sr.update_time ), '' ) AS update_time,
	hp.id AS pool_id,
	COALESCE (ha.cluster_name, '未分配集群') as cluster_name,
	COALESCE (cg.group_name, '未分配集群组') as group_name,
	COALESCE( dl.department_line_name, '未知业务线') as department_line_name,
	hp.host_ip AS ip,
	COALESCE ( hp.ram, 0 ) AS total_memory,
	COALESCE ( MAX( sr.used_memory ), 0 ) AS max_used_memory,
	COALESCE ( hp.disk_size, 0 ) AS total_disk,
	COALESCE ( MAX( sr.used_disk ), 0 ) AS max_used_disk,
	COALESCE ( hp.vcpus, 0 ) AS cpu_cores,
	COALESCE ( MAX( sr.cpu_load ), 0 ) AS max_cpu_load,
	COALESCE ( MAX( sr.mon_date ), '' ) AS max_datetime,
	COALESCE ( hp.host_name, '' ) AS host_name,
	COALESCE ( hp.host_type, '' ) AS host_type,
	COALESCE ( MAX( sr.cpu_percent_max ), 0 ) AS cpu_percent_max,
	COALESCE ( AVG( sr.cpu_percent_avg ), 0 ) AS cpu_percent_avg,
	COALESCE ( MIN( sr.cpu_percent_min ), 0 ) AS cpu_percent_min,
	COALESCE ( MAX( sr.mem_percent_max ), 0 ) AS mem_percent_max,
	COALESCE ( AVG( sr.mem_percent_avg ), 0 ) AS mem_percent_avg,
	COALESCE ( MIN( sr.mem_percent_min ), 0 ) AS mem_percent_min,
	COALESCE ( MAX( sr.disk_percent_max ), 0 ) AS disk_percent_max,
	COALESCE ( AVG( sr.disk_percent_avg ), 0 ) AS disk_percent_avg,
	COALESCE ( MIN( sr.disk_percent_min ), 0 ) AS disk_percent_min
FROM
	hosts_pool hp
	LEFT JOIN hosts_applications ha ON hp.id = ha.pool_id AND ha.is_deleted = 0
	LEFT JOIN cluster_groups cg ON ha.cluster_name = cg.cluster_name
	LEFT JOIN db_line dl on cg.group_name = dl.cluster_group_name
	LEFT JOIN server_resources sr ON hp.host_ip = sr.ip AND sr.is_deleted = 0` +
		timeJoinCondition +
		whereClause +
		` GROUP BY hp.host_ip, ha.cluster_name, cg.group_name
	ORDER BY hp.id, hp.host_ip, ha.cluster_name`

	var hostRows []*ServerResourceMaxData
	err := m.conn.QueryRowsCtx(ctx, &hostRows, hostResourceQuery, args...)
	if err != nil {
		return nil, err
	}
	return hostRows, nil
}

// FindServerResourcesWithFilter 查询服务器资源数据（支持时间和IP过滤）
func (m *customServerResourcesModel) FindServerResourcesWithFilter(ctx context.Context, beginTime, endTime string, ipList []string) ([]*ServerResourceRow, error) {
	// 构建基础查询语句
	// 保持 hosts_pool 为主表，确保所有主机都被统计（即使没有监控数据）
	// 使用子查询获取集群信息，避免因多集群导致的笛卡尔积
	// 子查询优先获取非空的集群名
	query := `SELECT sr.id, sr.create_time, sr.update_time, hp.id as pool_id,
			  COALESCE((SELECT ha.cluster_name FROM hosts_applications ha
			            WHERE ha.pool_id = hp.id AND ha.cluster_name IS NOT NULL AND ha.cluster_name != ''
			            ORDER BY ha.id LIMIT 1), '未分配集群') as cluster_name,
			  COALESCE((SELECT cg.group_name FROM hosts_applications ha
			            LEFT JOIN cluster_groups cg ON ha.cluster_name = cg.cluster_name
			            WHERE ha.pool_id = hp.id AND ha.cluster_name IS NOT NULL AND ha.cluster_name != ''
			            ORDER BY ha.id LIMIT 1), '未分配集群组') as group_name,
			  hp.host_ip as ip,
			  sr.port,
			  COALESCE((SELECT ha.server_role FROM hosts_applications ha
			            WHERE ha.pool_id = hp.id AND ha.cluster_name IS NOT NULL AND ha.cluster_name != ''
			            ORDER BY ha.id LIMIT 1), '') as instance_role,
			  COALESCE(hp.ram, 0) as total_memory, sr.used_memory, COALESCE(hp.disk_size, 0) as total_disk,
			  sr.used_disk, COALESCE(hp.vcpus, 0) as cpu_cores, sr.cpu_load, sr.mon_date,
			  COALESCE((SELECT cg.department_line_name FROM hosts_applications ha
			            LEFT JOIN cluster_groups cg ON ha.cluster_name = cg.cluster_name
			            WHERE ha.pool_id = hp.id AND ha.cluster_name IS NOT NULL AND ha.cluster_name != ''
			            ORDER BY ha.id LIMIT 1), '无业务线归属') as department_name,
			  sr.cpu_percent_max, sr.cpu_percent_avg, sr.cpu_percent_min,
			  sr.mem_percent_max, sr.mem_percent_avg, sr.mem_percent_min,
			  sr.disk_percent_max, sr.disk_percent_avg, sr.disk_percent_min
			  FROM hosts_pool hp
			  LEFT JOIN server_resources sr ON hp.host_ip = sr.ip AND sr.is_deleted = 0`

	args := []interface{}{}

	// 构建WHERE子句
	whereClause := " WHERE hp.is_deleted = 0"

	// 添加时间范围过滤
	if beginTime != "" && endTime != "" {
		whereClause += " AND sr.mon_date BETWEEN ? AND ?"
		args = append(args, beginTime, endTime)
	} else if beginTime != "" {
		whereClause += " AND sr.mon_date >= ?"
		args = append(args, beginTime)
	} else if endTime != "" {
		whereClause += " AND sr.mon_date <= ?"
		args = append(args, endTime)
	}

	// 添加IP过滤
	if len(ipList) > 0 {
		placeholders := make([]string, len(ipList))
		for i, ip := range ipList {
			placeholders[i] = "?"
			args = append(args, ip)
		}
		whereClause += " AND hp.host_ip IN (" + strings.Join(placeholders, ",") + ")"
	}

	// 组装完整查询
	query += whereClause

	// 添加排序
	query += " ORDER BY sr.mon_date DESC, hp.host_ip"

	var rows []*ServerResourceRow
	err := m.conn.QueryRowsCtx(ctx, &rows, query, args...)
	return rows, err
}

// FindDiskPredictionDataByCluster 根据集群名称查询磁盘预测数据
func (m *customServerResourcesModel) FindDiskPredictionDataByCluster(ctx context.Context, beginTime, endTime string, clusterName string) ([]*DiskPredictionData, error) {
	filter := &DiskPredictionFilter{
		BeginTime:   beginTime,
		EndTime:     endTime,
		ClusterName: clusterName,
	}
	return m.FindDiskPredictionDataWithFilter(ctx, filter)
}

// FindServerResourceMaxByCluster 根据集群名称查询服务器资源最大值数据
//func (m *customServerResourcesModel) FindServerResourceMaxByCluster(ctx context.Context, beginTime, endTime string, clusterName string) ([]*ServerResourceMaxData, error) {
//	// 构建查询语句，以 hosts_pool 为基础表，确保显示所有主机（即使没有资源数据）
//	baseQuery := `SELECT
//			COALESCE(MIN(hp.id), 0) as id,
//			COALESCE(MIN(sr.create_time), '') as create_time,
//			COALESCE(MAX(sr.update_time), '') as update_time,
//			hp.id as pool_id,
//			COALESCE(ha.cluster_name, '未分配集群') as cluster_name,
//			COALESCE(cg.group_name, '未分配集群组') as group_name,
//			hp.host_ip as ip,
//			COALESCE(MIN(sr.port), 0) as port,
//			COALESCE(ha.server_role, '') as instance_role,
//			COALESCE(MAX(sr.total_memory), 0) as total_memory,
//			COALESCE(MAX(sr.used_memory), 0) as max_used_memory,
//			COALESCE(MAX(sr.total_disk), 0) as total_disk,
//			COALESCE(MAX(sr.used_disk), 0) as max_used_disk,
//			COALESCE(MAX(sr.cpu_cores), 0) as cpu_cores,
//			COALESCE(MAX(sr.cpu_load), 0) as max_cpu_load,
//			COALESCE(MAX(sr.mon_date), '') as max_datetime,
//			COALESCE(cg.department_line_name, '无业务线归属') as department_name,
//			COALESCE(hp.host_name, '') as host_name,
//			COALESCE(hp.host_type, '') as os_type,
//			COALESCE(hp.host_type, '') as host_type,
//			'production' as deployment_environment,
//			'' as notes
//			FROM hosts_pool hp
//			LEFT JOIN hosts_applications ha ON hp.id = ha.pool_id
//			LEFT JOIN cluster_groups cg ON ha.cluster_name = cg.cluster_name
//			LEFT JOIN server_resources sr ON hp.host_ip = sr.ip AND sr.is_deleted = 0`
//
//	// 将时间过滤条件加入JOIN条件
//	if beginTime != "" && endTime != "" {
//		baseQuery += " AND sr.mon_date BETWEEN ? AND ?"
//	}
//
//	baseQuery += ` WHERE hp.is_deleted = 0
//			AND (ha.cluster_name = ? OR (? = '' AND (ha.cluster_name IS NULL OR ha.cluster_name = '')))
//			GROUP BY hp.host_ip, hp.id, ha.cluster_name, cg.group_name ORDER BY hp.host_ip`
//
//	var rows []*ServerResourceMaxData
//	err := m.conn.QueryRowsCtx(ctx, &rows, baseQuery, beginTime, endTime, clusterName, clusterName)
//	return rows, err
//}

// FindServerResourcesWithClusterFilter 根据集群名称查询服务器资源数据
func (m *customServerResourcesModel) FindServerResourcesWithClusterFilter(ctx context.Context, beginTime, endTime string, clusterName string) ([]*ServerResourceRow, error) {
	// 构建基础查询语句
	// 保持 hosts_pool 为主表，确保所有主机都被统计（即使没有监控数据）
	// 使用子查询获取集群信息，避免因多集群导致的笛卡尔积
	// 子查询优先获取非空的集群名
	query := `SELECT sr.id, sr.create_time, sr.update_time, hp.id as pool_id,
			  COALESCE((SELECT ha.cluster_name FROM hosts_applications ha
			            WHERE ha.pool_id = hp.id AND ha.cluster_name IS NOT NULL AND ha.cluster_name != ''
			            ORDER BY ha.id LIMIT 1), '未分配集群') as cluster_name,
			  COALESCE((SELECT cg.group_name FROM hosts_applications ha
			            LEFT JOIN cluster_groups cg ON ha.cluster_name = cg.cluster_name
			            WHERE ha.pool_id = hp.id AND ha.cluster_name IS NOT NULL AND ha.cluster_name != ''
			            ORDER BY ha.id LIMIT 1), '未分配集群组') as group_name,
			  hp.host_ip as ip,
			  sr.port,
			  COALESCE((SELECT ha.server_role FROM hosts_applications ha
			            WHERE ha.pool_id = hp.id AND ha.cluster_name IS NOT NULL AND ha.cluster_name != ''
			            ORDER BY ha.id LIMIT 1), '') as instance_role,
			  COALESCE(hp.ram, 0) as total_memory, sr.used_memory, COALESCE(hp.disk_size, 0) as total_disk,
			  sr.used_disk, COALESCE(hp.vcpus, 0) as cpu_cores, sr.cpu_load, sr.mon_date,
			  COALESCE((SELECT cg.department_line_name FROM hosts_applications ha
			            LEFT JOIN cluster_groups cg ON ha.cluster_name = cg.cluster_name
			            WHERE ha.pool_id = hp.id AND ha.cluster_name IS NOT NULL AND ha.cluster_name != ''
			            ORDER BY ha.id LIMIT 1), '无业务线归属') as department_name,
			  sr.cpu_percent_max, sr.cpu_percent_avg, sr.cpu_percent_min,
			  sr.mem_percent_max, sr.mem_percent_avg, sr.mem_percent_min,
			  sr.disk_percent_max, sr.disk_percent_avg, sr.disk_percent_min
			  FROM hosts_pool hp
			  LEFT JOIN server_resources sr ON hp.host_ip = sr.ip AND sr.is_deleted = 0
			  WHERE hp.is_deleted = 0 AND sr.mon_date BETWEEN ? AND ?
			  AND (EXISTS (SELECT 1 FROM hosts_applications ha WHERE ha.pool_id = hp.id AND ha.cluster_name = ?)
			       OR (? = '' AND NOT EXISTS (SELECT 1 FROM hosts_applications ha WHERE ha.pool_id = hp.id)))
			  ORDER BY sr.mon_date DESC, hp.host_ip`

	var rows []*ServerResourceRow
	err := m.conn.QueryRowsCtx(ctx, &rows, query, beginTime, endTime, clusterName, clusterName)
	return rows, err
}

// FindClusterResources 查询集群资源详细信息
func (m *customServerResourcesModel) FindClusterResources(ctx context.Context, beginTime, endTime, clusterName, groupName string) ([]*ClusterMemberResourceData, error) {
	// 构建查询条件
	whereClause := "hp.is_deleted = 0 AND sr.is_deleted = 0"
	var args []interface{}

	// 添加时间范围查询条件
	if beginTime != "" && endTime != "" {
		whereClause += " AND sr.mon_date BETWEEN ? AND ?"
		args = append(args, beginTime, endTime)
	}

	// 添加集群名称查询条件
	// 修改逻辑：避免笛卡尔积，使用 EXISTS 子查询
	if clusterName != "" {
		// 指定集群：查询属于该集群的主机
		whereClause += " AND EXISTS (SELECT 1 FROM hosts_applications ha WHERE ha.pool_id = hp.id AND ha.cluster_name = ? AND ha.is_deleted = 0)"
		args = append(args, clusterName)
	} else {
		// 空集群：只查询没有任何有效集群归属的主机（所有记录的cluster_name都为空或没有记录）
		whereClause += " AND NOT EXISTS (SELECT 1 FROM hosts_applications ha WHERE ha.pool_id = hp.id AND ha.cluster_name IS NOT NULL AND ha.cluster_name != '' AND ha.is_deleted = 0)"
	}

	// 添加集群组查询条件
	if groupName != "" {
		whereClause += " AND EXISTS (SELECT 1 FROM hosts_applications ha LEFT JOIN cluster_groups cg ON ha.cluster_name = cg.cluster_name WHERE ha.pool_id = hp.id AND cg.group_name = ? AND ha.is_deleted = 0)"
		args = append(args, groupName)
	}

	// 使用子查询获取第一个集群信息，避免笛卡尔积
	query := `
		SELECT
			hp.id,
			COALESCE((SELECT ha.cluster_name FROM hosts_applications ha
			          WHERE ha.pool_id = hp.id AND ha.is_deleted = 0
			          AND ha.cluster_name IS NOT NULL AND ha.cluster_name != ''
			          ORDER BY ha.id LIMIT 1), '') as cluster_name,
			COALESCE((SELECT cg.group_name FROM hosts_applications ha
			          LEFT JOIN cluster_groups cg ON ha.cluster_name = cg.cluster_name
			          WHERE ha.pool_id = hp.id AND ha.is_deleted = 0
			          AND ha.cluster_name IS NOT NULL AND ha.cluster_name != ''
			          ORDER BY ha.id LIMIT 1), '') as group_name,
			hp.host_ip as ip,
			COALESCE(hp.host_name, '') as host_name,
			0 as port,
			'' as instance_role,
			COALESCE(hp.ram, 0) as total_memory,
			sr.used_memory,
			COALESCE(hp.disk_size, 0) as total_disk,
			sr.used_disk,
			COALESCE(hp.vcpus, 0) as cpu_cores,
			sr.cpu_load,
			sr.mon_date as mon_date,
			'' as department_name,
			sr.cpu_percent_max,
			sr.cpu_percent_avg,
			sr.cpu_percent_min,
			sr.mem_percent_max,
			sr.mem_percent_avg,
			sr.mem_percent_min,
			sr.disk_percent_max,
			sr.disk_percent_avg,
			sr.disk_percent_min
		FROM hosts_pool hp
		LEFT JOIN server_resources sr ON hp.host_ip = sr.ip AND sr.is_deleted = 0
		WHERE ` + whereClause + `
		ORDER BY hp.host_ip, sr.mon_date DESC
	`

	var results []*ClusterMemberResourceData
	err := m.conn.QueryRowsCtx(ctx, &results, query, args...)
	return results, err
}

// FindClusterResourcesMax 查询集群资源最大利用率信息
// 说明：一个主机可能有多个应用，每个应用可能属于不同集群
// 该查询统计每个集群中所有主机的最大资源利用率情况
// 兼容MySQL 5.7，不使用CTE语法
func (m *customServerResourcesModel) FindClusterResourcesMax(ctx context.Context, beginTime, endTime, clusterName, groupName string) ([]*ClusterResourceMaxData, error) {
	// 构建查询条件
	whereClause := "1=1"
	var args []interface{}

	if clusterName != "" {
		whereClause += " AND cluster_name = ?"
		args = append(args, clusterName)
	}

	if groupName != "" {
		whereClause += " AND group_name = ?"
		args = append(args, groupName)
	}

	// 构建时间过滤条件
	timeCondition := ""
	var timeArgs []interface{}
	if beginTime != "" && endTime != "" {
		timeCondition = " AND sr.mon_date BETWEEN ? AND ?"
		timeArgs = append(timeArgs, beginTime, endTime)
	} else if beginTime != "" {
		timeCondition = " AND sr.mon_date >= ?"
		timeArgs = append(timeArgs, beginTime)
	} else if endTime != "" {
		timeCondition = " AND sr.mon_date <= ?"
		timeArgs = append(timeArgs, endTime)
	}

	query := `
		SELECT
			cluster_name,
			group_name,
			-- 优先使用db_line表的部门信息，如果没有则使用cluster_groups的部门信息
			COALESCE(MAX(dl.department_line_name), MAX(department_name)) as department_name,
			COUNT(DISTINCT host_ip) as node_count,
			-- 计算平均值
			AVG(avg_cpu_load) as avg_cpu_load,
			AVG(avg_memory_usage) as avg_memory_usage,
			AVG(avg_disk_usage) as avg_disk_usage,
			-- 计算最大值
			MAX(max_cpu_load) as max_cpu_load,
			MAX(max_memory_usage) as max_memory_usage,
			MAX(max_disk_usage) as max_disk_usage,
			MAX(max_datetime) as max_datetime,
			max(total_memory) as total_memory,
			max(total_disk) as total_disk,
			max(cpu_cores) as cpu_cores,
			GROUP_CONCAT(DISTINCT department_name) as bus_lines,
			-- 百分比字段聚合
			MAX(max_cpu_percent) as cpu_percent_max,
			AVG(avg_cpu_percent) as cpu_percent_avg,
			MIN(min_cpu_percent) as cpu_percent_min,
			MAX(max_mem_percent) as mem_percent_max,
			AVG(avg_mem_percent) as mem_percent_avg,
			MIN(min_mem_percent) as mem_percent_min,
			MAX(max_disk_percent) as disk_percent_max,
			AVG(avg_disk_percent) as disk_percent_avg,
			MIN(min_disk_percent) as disk_percent_min
		FROM (
			-- 获取每个集群-主机的最大和平均资源利用率
			SELECT
				CASE
					WHEN ha.cluster_name IS NULL OR ha.cluster_name = '' THEN '未分配集群'
					ELSE ha.cluster_name
				END as cluster_name,
				COALESCE(cg.group_name, '未分配集群组') as group_name,
				COALESCE(cg.department_line_name, '未知业务线') as department_name,
				hp.host_ip,
				-- 最大值
				MAX(sr.cpu_load) as max_cpu_load,
				MAX(CASE
					WHEN hp.ram > 0 THEN (sr.used_memory / hp.ram * 100)
					ELSE 0
				END) as max_memory_usage,
				MAX(CASE
					WHEN hp.disk_size > 0 THEN (sr.used_disk / hp.disk_size * 100)
					ELSE 0
				END) as max_disk_usage,
				-- 平均值
				AVG(sr.cpu_load) as avg_cpu_load,
				AVG(CASE
					WHEN hp.ram > 0 THEN (sr.used_memory / hp.ram * 100)
					ELSE 0
				END) as avg_memory_usage,
				AVG(CASE
					WHEN hp.disk_size > 0 THEN (sr.used_disk / hp.disk_size * 100)
					ELSE 0
				END) as avg_disk_usage,
			    max(hp.ram) as total_memory,
			    max(hp.disk_size) as total_disk,
			    max(hp.vcpus) as cpu_cores,
				MAX(sr.mon_date) as max_datetime,
				-- 百分比字段聚合
				MAX(sr.cpu_percent_max) as max_cpu_percent,
				AVG(sr.cpu_percent_avg) as avg_cpu_percent,
				MIN(sr.cpu_percent_min) as min_cpu_percent,
				MAX(sr.mem_percent_max) as max_mem_percent,
				AVG(sr.mem_percent_avg) as avg_mem_percent,
				MIN(sr.mem_percent_min) as min_mem_percent,
				MAX(sr.disk_percent_max) as max_disk_percent,
				AVG(sr.disk_percent_avg) as avg_disk_percent,
				MIN(sr.disk_percent_min) as min_disk_percent
			FROM hosts_pool hp
			INNER JOIN hosts_applications ha ON hp.id = ha.pool_id
			LEFT JOIN cluster_groups cg ON (ha.cluster_name = cg.cluster_name AND ha.cluster_name != '' AND ha.cluster_name IS NOT NULL)
			LEFT JOIN server_resources sr ON hp.host_ip = sr.ip AND sr.is_deleted = 0` + timeCondition + `
			WHERE hp.is_deleted = 0
				AND ha.is_deleted = 0
			GROUP BY
				CASE
					WHEN ha.cluster_name IS NULL OR ha.cluster_name = '' THEN '未分配集群'
					ELSE ha.cluster_name
				END,
				group_name, department_line_name, hp.host_ip

			UNION ALL

			-- 处理没有应用分配的主机，归入"未分配集群"
			SELECT
				'未分配集群' as cluster_name,
				'未分配集群组' as group_name,
				'未知业务线' as department_name,
				hp.host_ip,
				-- 最大值
				MAX(sr.cpu_load) as max_cpu_load,
				MAX(CASE
					WHEN hp.ram > 0 THEN (sr.used_memory / hp.ram * 100)
					ELSE 0
				END) as max_memory_usage,
				MAX(CASE
					WHEN hp.disk_size > 0 THEN (sr.used_disk / hp.disk_size * 100)
					ELSE 0
				END) as max_disk_usage,
				-- 平均值
				AVG(sr.cpu_load) as avg_cpu_load,
				AVG(CASE
					WHEN hp.ram > 0 THEN (sr.used_memory / hp.ram * 100)
					ELSE 0
				END) as avg_memory_usage,
				AVG(CASE
					WHEN hp.disk_size > 0 THEN (sr.used_disk / hp.disk_size * 100)
					ELSE 0
				END) as avg_disk_usage,
			    max(hp.ram) as total_memory,
			    max(hp.disk_size) as total_disk,
			    max(hp.vcpus) as cpu_cores,
				MAX(sr.mon_date) as max_datetime,
				-- 百分比字段聚合
				MAX(sr.cpu_percent_max) as max_cpu_percent,
				AVG(sr.cpu_percent_avg) as avg_cpu_percent,
				MIN(sr.cpu_percent_min) as min_cpu_percent,
				MAX(sr.mem_percent_max) as max_mem_percent,
				AVG(sr.mem_percent_avg) as avg_mem_percent,
				MIN(sr.mem_percent_min) as min_mem_percent,
				MAX(sr.disk_percent_max) as max_disk_percent,
				AVG(sr.disk_percent_avg) as avg_disk_percent,
				MIN(sr.disk_percent_min) as min_disk_percent
			FROM hosts_pool hp
			LEFT JOIN server_resources sr ON hp.host_ip = sr.ip AND sr.is_deleted = 0` + timeCondition + `
			WHERE hp.is_deleted = 0
				AND hp.id NOT IN (
					SELECT DISTINCT pool_id
					FROM hosts_applications
					WHERE is_deleted = 0 AND pool_id IS NOT NULL
				)
			GROUP BY hp.host_ip
		) as cluster_host_resources
		LEFT JOIN db_line dl ON cluster_host_resources.group_name = dl.cluster_group_name AND dl.is_deleted = 0
		WHERE ` + whereClause + `
		GROUP BY cluster_name, group_name, department_name
		ORDER BY cluster_name, group_name
	`

	// 构建完整的参数列表：时间参数需要传递两次（对应两个UNION子查询中的时间过滤）
	fullArgs := make([]interface{}, 0)
	fullArgs = append(fullArgs, timeArgs...) // 第一个查询的时间参数
	fullArgs = append(fullArgs, timeArgs...) // 第二个查询的时间参数
	fullArgs = append(fullArgs, args...)     // WHERE子句的参数

	var results []*ClusterResourceMaxData
	err := m.conn.QueryRowsCtx(ctx, &results, query, fullArgs...)
	return results, err
}

// FindClusterMemberResources 查询指定集群的成员节点详细信息
func (m *customServerResourcesModel) FindClusterMemberResources(ctx context.Context, clusterName string) ([]*ClusterMemberResourceData, error) {
	query := `
		SELECT
			hp.id,
			COALESCE(ha.cluster_name, '') as cluster_name,
			COALESCE(cg.group_name, '') as group_name,
			hp.host_ip as ip,
			COALESCE(hp.host_name, '') as host_name,
			ha.server_port as port,
			ha.server_role as instance_role,
			COALESCE(hp.ram, 0) as total_memory,
			sr.used_memory,
			COALESCE(hp.disk_size, 0) as total_disk,
			sr.used_disk,
			COALESCE(hp.vcpus, 0) as cpu_cores,
			sr.cpu_load,
			sr.mon_date as mon_date,
			'' as department_name
		FROM hosts_pool hp
		LEFT JOIN hosts_applications ha ON hp.id = ha.pool_id
		LEFT JOIN cluster_groups cg ON ha.cluster_name = cg.cluster_name
		LEFT JOIN server_resources sr ON hp.host_ip = sr.ip AND sr.is_deleted = 0
		WHERE ha.cluster_name = ? AND hp.is_deleted = 0
	`

	var results []*ClusterMemberResourceData
	err := m.conn.QueryRowsCtx(ctx, &results, query, clusterName)
	return results, err
}

// FindDiskPredictionDataByDepartment 根据部门名称查询磁盘预测数据
func (m *customServerResourcesModel) FindDiskPredictionDataByDepartment(ctx context.Context, beginTime, endTime string, departmentName string) ([]*DiskPredictionData, error) {
	filter := &DiskPredictionFilter{
		BeginTime:      beginTime,
		EndTime:        endTime,
		DepartmentName: departmentName,
	}
	return m.FindDiskPredictionDataWithFilter(ctx, filter)
}

// FindDiskPredictionDataWithFilter 统一的磁盘预测数据查询函数，支持多种过滤条件
func (m *customServerResourcesModel) FindDiskPredictionDataWithFilter(ctx context.Context, filter *DiskPredictionFilter) ([]*DiskPredictionData, error) {
	// 基础查询SQL - 统一的表关联和字段选择
	baseQuery := `
		SELECT
			hp.id as id,
			hp.host_ip as ip,
			COALESCE(ha.cluster_name, "未分配集群") AS cluster_name,
			COALESCE(cg.group_name, "未分配集群组") AS group_name,
			COALESCE(dl.department_line_name, "未知业务线") AS department_line_name,
			COALESCE(hp.disk_size, 0) as total_disk,
			COALESCE(sr.used_disk, 0) as used_disk,
			COALESCE(sr.disk_percent_max, 0) as disk_percent_max,
			COALESCE(sr.mon_date, '') as mon_date,
			hp.idc_id as idc_id,
			ic.idc_name as idc_name,
			ic.idc_code as idc_code,
			ic.idc_location as idc_location,
			ic.idc_description as idc_description
		FROM hosts_pool hp
		LEFT JOIN hosts_applications ha ON hp.id = ha.pool_id
		LEFT JOIN cluster_groups cg ON ha.cluster_name = cg.cluster_name
		LEFT JOIN db_line dl ON cg.group_name = dl.cluster_group_name`

	// 如果需要部门筛选，则添加db_line表关联
	if filter.DepartmentName != "" {
		baseQuery += `
		LEFT JOIN db_line dl ON cg.group_name = dl.cluster_group_name`
	}

	baseQuery += `
		LEFT JOIN server_resources sr ON hp.host_ip = sr.ip AND sr.is_deleted = 0
		LEFT JOIN idc_conf ic ON hp.idc_id = ic.id
		WHERE hp.is_deleted = 0
		AND (sr.mon_date IS NULL OR sr.mon_date BETWEEN ? AND ?)`

	var args []interface{}
	args = append(args, filter.BeginTime, filter.EndTime)

	// 构建动态WHERE条件
	if len(filter.IPList) > 0 {
		placeholders := ""
		for i, ip := range filter.IPList {
			if i > 0 {
				placeholders += ","
			}
			placeholders += "?"
			args = append(args, ip)
		}
		baseQuery += " AND hp.host_ip IN (" + placeholders + ")"
	}

	if filter.ClusterName != "" {
		baseQuery += " AND ha.cluster_name = ?"
		args = append(args, filter.ClusterName)
	}

	if filter.DepartmentName != "" {
		baseQuery += " AND dl.department_line_name = ?"
		args = append(args, filter.DepartmentName)
	}

	baseQuery += " ORDER BY hp.host_ip, sr.mon_date"

	var diskData []*DiskPredictionData
	err := m.conn.QueryRowsCtx(ctx, &diskData, baseQuery, args...)
	return diskData, err
}

// UpsertFromES ES数据同步专用的upsert方法
// 使用 INSERT ... ON DUPLICATE KEY UPDATE 实现 upsert
// 根据 pool_id + ip 的唯一性来判断是否存在（需要确保表有相应的唯一索引）
func (m *customServerResourcesModel) UpsertFromES(ctx context.Context, poolId uint64, ip string,
	usedMemory, usedDisk, cpuLoad float64,
	cpuPercentMax, cpuPercentAvg, cpuPercentMin,
	memPercentMax, memPercentAvg, memPercentMin,
	diskPercentMax, diskPercentAvg, diskPercentMin float64) error {
	query := `
		INSERT INTO server_resources
		(pool_id, ip,
		 used_memory, used_disk, cpu_load,
		 cpu_percent_max, cpu_percent_avg, cpu_percent_min,
		 mem_percent_max, mem_percent_avg, mem_percent_min,
		 disk_percent_max, disk_percent_avg, disk_percent_min,
		 mon_date, is_deleted)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, CURDATE(), 0)
		ON DUPLICATE KEY UPDATE
		    used_memory = VALUES(used_memory),
		    used_disk = VALUES(used_disk),
		    cpu_load = VALUES(cpu_load),
		    cpu_percent_max = VALUES(cpu_percent_max),
		    cpu_percent_avg = VALUES(cpu_percent_avg),
		    cpu_percent_min = VALUES(cpu_percent_min),
		    mem_percent_max = VALUES(mem_percent_max),
		    mem_percent_avg = VALUES(mem_percent_avg),
		    mem_percent_min = VALUES(mem_percent_min),
		    disk_percent_max = VALUES(disk_percent_max),
		    disk_percent_avg = VALUES(disk_percent_avg),
		    disk_percent_min = VALUES(disk_percent_min),
		    mon_date = VALUES(mon_date),
		    is_deleted = 0
	`

	_, err := m.conn.ExecCtx(ctx, query, poolId, ip,
		usedMemory, usedDisk, cpuLoad,
		cpuPercentMax, cpuPercentAvg, cpuPercentMin,
		memPercentMax, memPercentAvg, memPercentMin,
		diskPercentMax, diskPercentAvg, diskPercentMin)
	return err
}

// UpsertFromCMSys CMSys数据同步专用的upsert方法
// 使用 INSERT ... ON DUPLICATE KEY UPDATE 实现 upsert
// CMSys只返回百分比数据，不返回绝对值
func (m *customServerResourcesModel) UpsertFromCMSys(ctx context.Context, poolId uint64, ip string,
	cpuPercent, memPercent, diskPercent float64) error {
	query := `
		INSERT INTO server_resources
		(pool_id, ip,
		 cpu_percent_max, cpu_percent_avg, cpu_percent_min,
		 mem_percent_max, mem_percent_avg, mem_percent_min,
		 disk_percent_max, disk_percent_avg, disk_percent_min,
		 mon_date, is_deleted)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, CURDATE(), 0)
		ON DUPLICATE KEY UPDATE
		    cpu_percent_max = VALUES(cpu_percent_max),
		    cpu_percent_avg = VALUES(cpu_percent_avg),
		    cpu_percent_min = VALUES(cpu_percent_min),
		    mem_percent_max = VALUES(mem_percent_max),
		    mem_percent_avg = VALUES(mem_percent_avg),
		    mem_percent_min = VALUES(mem_percent_min),
		    disk_percent_max = VALUES(disk_percent_max),
		    disk_percent_avg = VALUES(disk_percent_avg),
		    disk_percent_min = VALUES(disk_percent_min),
		    mon_date = VALUES(mon_date),
		    is_deleted = 0
	`

	_, err := m.conn.ExecCtx(ctx, query, poolId, ip,
		cpuPercent, cpuPercent, cpuPercent, // CPU百分比 max=avg=min（CMSys只有单个值）
		memPercent, memPercent, memPercent, // 内存百分比 max=avg=min
		diskPercent, diskPercent, diskPercent) // 磁盘百分比 max=avg=min
	return err
}
