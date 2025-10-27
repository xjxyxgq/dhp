# ES数据同步字段映射说明

## 问题描述

在实现 ES 数据同步功能时，发现初始实现中使用的字段名与实际的 `server_resources` 表结构不匹配。

## 实际的 server_resources 表结构

```sql
CREATE TABLE `server_resources` (
  `id` int(10) unsigned NOT NULL AUTO_INCREMENT,
  `created_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `updated_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  `deleted_at` datetime DEFAULT NULL,
  `pool_id` int(10) unsigned NOT NULL COMMENT '主机池ID',
  `cluster_name` varchar(64) DEFAULT NULL COMMENT '集群名称',
  `group_name` varchar(100) DEFAULT NULL COMMENT '组名称',
  `ip` varchar(50) DEFAULT NULL COMMENT 'IP地址',
  `port` int(10) unsigned DEFAULT NULL COMMENT '端口',
  `instance_role` varchar(50) DEFAULT NULL COMMENT '实例角色',
  `total_memory` double DEFAULT NULL COMMENT '总内存(GB)',
  `used_memory` double DEFAULT NULL COMMENT '已用内存(GB)',
  `total_disk` double DEFAULT NULL COMMENT '总磁盘(GB)',
  `used_disk` double DEFAULT NULL COMMENT '已用磁盘(GB)',
  `cpu_cores` int(11) DEFAULT NULL COMMENT 'CPU核数',
  `cpu_load` double DEFAULT NULL COMMENT 'CPU负载(%)',
  `date_time` datetime NOT NULL COMMENT '监控时间',
  PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='服务器资源监控表';
```

## ES数据结构

ES中的数据字段（来自 cluster*:data-zabbix-host-monitor-* 索引）：
- `hostIp`: 主机IP
- `hostName`: 主机名
- `cpu`: CPU使用率（聚合数据，包含 max, avg, min 等）
- `available_memory`: 可用内存（聚合数据）
- `total_disk_space_all`: 总磁盘空间（聚合数据）

## 正确的字段映射关系

### ES → server_resources

| ES 字段 | ES 聚合值 | server_resources 字段 | 说明 |
|---------|----------|---------------------|------|
| hostIp | - | ip | 主机IP地址 |
| - | - | pool_id | 从 hosts_pool 表查询获取 |
| cpu | max | cpu_load | 最大CPU负载(%) |
| available_memory | max | used_memory | 可用内存转为已用内存 |
| total_disk_space_all | max | total_disk | 总磁盘空间(GB) |
| - | - | date_time | 使用查询时间范围的结束时间或当前时间 |

### 字段转换逻辑

1. **pool_id**: 通过 hostIp 查询 hosts_pool 表获取对应的 id
2. **ip**: 直接使用 ES 的 hostIp
3. **cpu_load**: 使用 ES 聚合数据中的 cpu.max 值
4. **used_memory**:
   - 如果 ES 提供 total_memory，则 used_memory = total_memory - available_memory
   - 否则，暂时使用 available_memory 的反向值（需要进一步确认）
5. **total_disk**: 使用 ES 聚合数据中的 total_disk_space_all.max 值
6. **date_time**: 使用当前时间戳
7. **其他字段**: 保持为 NULL 或默认值

## 修改建议

### 1. 修改 HostMetrics 结构体

原来的结构体：
```go
type HostMetrics struct {
    HostIP         string
    HostName       string
    MaxCPU         float64
    AvgCPU         float64
    MaxMemory      float64
    AvgMemory      float64
    MaxDisk        float64
    AvgDisk        float64
    DataPointCount int
}
```

应修改为与 server_resources 表匹配的结构。

### 2. 修改同步逻辑

需要修改 `ExecuteEsSyncByHostList` 中的数据写入逻辑：
- 使用正确的字段名
- 使用 server_resources 表的实际字段
- 正确映射 ES 数据到表字段

### 3. ServerResourcesModel 添加方法

需要在 ServerResourcesModel 中添加一个用于 ES 同步的 upsert 方法：

```go
// UpsertFromES ES数据同步专用的upsert方法
// 根据 pool_id + ip + date_time 判断是否存在，存在则更新，不存在则插入
func (m *customServerResourcesModel) UpsertFromES(ctx context.Context, poolId uint64, ip string, cpuLoad, usedMemory, totalDisk float64, dateTime time.Time) error
```

## 实施步骤

1. 更新 ES Client 的 HostMetrics 结构体
2. 在 ServerResourcesModel 中添加 UpsertFromES 方法
3. 修改 ExecuteEsSyncByHostList 逻辑中的数据同步部分
4. 更新相关的测试用例

## 注意事项

1. **数据类型**: 确保数据类型转换正确（ES 的浮点数 → 数据库的 double）
2. **单位转换**: 注意内存和磁盘的单位（ES 可能是字节，数据库是 GB）
3. **时区处理**: 确保时间字段的时区处理正确
4. **Null 值处理**: server_resources 中很多字段允许 NULL，需要使用 sql.NullXXX 类型
