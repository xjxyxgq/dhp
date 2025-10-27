# 资源查询接口 API 文档

本文档描述了 CMDB 系统中 4 个资源查询接口的详细信息，包括新增的资源利用率百分比字段。

## 目录

1. [主机资源查询接口](#1-主机资源查询接口)
2. [主机资源最大值查询接口](#2-主机资源最大值查询接口)
3. [集群资源查询接口](#3-集群资源查询接口)
4. [集群资源最大值查询接口](#4-集群资源最大值查询接口)
5. [百分比字段说明](#5-百分比字段说明)
6. [测试示例](#6-测试示例)

---

## 1. 主机资源查询接口

### 接口信息

- **接口路径**: `/api/cmdb/v1/server-resources`
- **请求方法**: `GET`
- **接口描述**: 查询主机资源使用率数据（详细的时间序列数据）

### 请求参数

| 参数名 | 类型 | 必填 | 说明 |
|--------|------|------|------|
| start_date | string | 否 | 开始日期，格式：YYYY-MM-DD，默认为3个月前 |
| end_date | string | 否 | 结束日期，格式：YYYY-MM-DD，默认为当前日期 |
| ip | string | 否 | 主机IP地址，用于过滤特定主机 |
| cluster | string | 否 | 集群名称，用于过滤特定集群 |

### 响应格式

```json
{
  "success": true,
  "message": "查询成功",
  "list": [
    {
      "id": 1,
      "create_at": "2024-01-01T00:00:00Z",
      "update_at": "2024-01-01T00:00:00Z",
      "pool_id": 1,
      "ip": "192.168.1.100",
      "total_memory": 64.0,
      "used_memory": 32.5,
      "total_disk": 500.0,
      "used_disk": 250.2,
      "cpu_cores": 8,
      "cpu_load": 4.5,
      "date_time": "2024-01-01T12:00:00Z",

      // 新增：CPU 利用率百分比
      "cpu_percent_max": 75.5,
      "cpu_percent_avg": 56.2,
      "cpu_percent_min": 35.0,

      // 新增：内存利用率百分比
      "mem_percent_max": 82.3,
      "mem_percent_avg": 70.5,
      "mem_percent_min": 55.0,

      // 新增：磁盘利用率百分比
      "disk_percent_max": 68.9,
      "disk_percent_avg": 50.1,
      "disk_percent_min": 40.2,

      "clusters": [
        {
          "cluster_name": "MySQL主集群",
          "cluster_group_name": "生产环境",
          "department_name": "核心业务线"
        }
      ],
      "idc_info": {
        "id": 1,
        "idc_name": "北京IDC",
        "idc_code": "BJ01",
        "idc_location": "北京市朝阳区",
        "idc_description": "主数据中心"
      }
    }
  ]
}
```

### 测试命令

```bash
# 查询所有主机资源（默认最近3个月）
curl -X GET "http://localhost:8888/api/cmdb/v1/server-resources"

# 查询特定时间范围
curl -X GET "http://localhost:8888/api/cmdb/v1/server-resources?start_date=2024-01-01&end_date=2024-03-31"

# 查询特定IP的主机资源
curl -X GET "http://localhost:8888/api/cmdb/v1/server-resources?ip=192.168.1.100"

# 查询特定集群的主机资源
curl -X GET "http://localhost:8888/api/cmdb/v1/server-resources?cluster=MySQL主集群"
```

---

## 2. 主机资源最大值查询接口

### 接口信息

- **接口路径**: `/api/cmdb/v1/server-resources-max`
- **请求方法**: `GET`
- **接口描述**: 查询主机资源最大利用率数据（聚合的最大值数据）

### 请求参数

| 参数名 | 类型 | 必填 | 说明 |
|--------|------|------|------|
| start_date | string | 否 | 开始日期，格式：YYYY-MM-DD，默认为3个月前 |
| end_date | string | 否 | 结束日期，格式：YYYY-MM-DD，默认为当前日期 |
| ip | string | 否 | 主机IP地址，用于过滤特定主机 |
| cluster | string | 否 | 集群名称，用于过滤特定集群 |

### 响应格式

```json
{
  "success": true,
  "message": "查询成功",
  "list": [
    {
      "id": 1,
      "create_at": "2024-01-01T00:00:00Z",
      "update_at": "2024-01-01T00:00:00Z",
      "pool_id": 1,
      "cluster_name": "MySQL主集群",
      "group_name": "生产环境",
      "ip": "192.168.1.100",
      "host_name": "db-server-01",
      "host_type": "物理机",
      "total_memory": 64.0,
      "max_used_memory": 52.8,
      "total_disk": 500.0,
      "max_used_disk": 344.5,
      "cpu_cores": 8,
      "max_cpu_load": 6.8,
      "max_date_time": "2024-01-15T14:30:00Z",

      // 新增：CPU 利用率百分比（最大值、平均值、最小值）
      "cpu_percent_max": 85.0,
      "cpu_percent_avg": 65.3,
      "cpu_percent_min": 45.2,

      // 新增：内存利用率百分比（最大值、平均值、最小值）
      "mem_percent_max": 82.5,
      "mem_percent_avg": 70.8,
      "mem_percent_min": 60.1,

      // 新增：磁盘利用率百分比（最大值、平均值、最小值）
      "disk_percent_max": 68.9,
      "disk_percent_avg": 55.2,
      "disk_percent_min": 42.0,

      "clusters": [
        {
          "cluster_name": "MySQL主集群",
          "cluster_group_name": "生产环境",
          "department_name": "核心业务线"
        }
      ],
      "idc_info": {
        "id": 1,
        "idc_name": "北京IDC",
        "idc_code": "BJ01",
        "idc_location": "北京市朝阳区",
        "idc_description": "主数据中心"
      }
    }
  ]
}
```

### 测试命令

```bash
# 查询所有主机资源最大值（默认最近3个月）
curl -X GET "http://localhost:8888/api/cmdb/v1/server-resources-max"

# 查询特定时间范围的最大值
curl -X GET "http://localhost:8888/api/cmdb/v1/server-resources-max?start_date=2024-01-01&end_date=2024-03-31"

# 查询特定IP的主机资源最大值
curl -X GET "http://localhost:8888/api/cmdb/v1/server-resources-max?ip=192.168.1.100"

# 查询特定集群的主机资源最大值
curl -X GET "http://localhost:8888/api/cmdb/v1/server-resources-max?cluster=MySQL主集群"
```

---

## 3. 集群资源查询接口

### 接口信息

- **接口路径**: `/api/cmdb/v1/cluster-resources`
- **请求方法**: `GET`
- **接口描述**: 获取集群资源详细信息（集群成员节点的资源数据）

### 请求参数

| 参数名 | 类型 | 必填 | 说明 |
|--------|------|------|------|
| start_date | string | 否 | 开始日期时间，格式：YYYY-MM-DD HH:MM:SS |
| end_date | string | 否 | 结束日期时间，格式：YYYY-MM-DD HH:MM:SS |
| cluster_name | string | 否 | 集群名称 |
| group_name | string | 否 | 集群组名称 |

### 响应格式

```json
{
  "success": true,
  "message": "查询成功",
  "list": [
    {
      "id": 1,
      "cluster_name": "MySQL主集群",
      "cluster_group_name": "生产环境",
      "ip": "192.168.1.100",
      "host_name": "db-server-01",
      "port": 3306,
      "instance_role": "master",
      "total_memory": 64.0,
      "used_memory": 32.5,
      "total_disk": 500.0,
      "used_disk": 250.2,
      "cpu_cores": 8,
      "cpu_load": 4.5,
      "date_time": "2024-01-01T12:00:00Z",

      // 新增：CPU 利用率百分比
      "cpu_percent_max": 75.5,
      "cpu_percent_avg": 56.2,
      "cpu_percent_min": 35.0,

      // 新增：内存利用率百分比
      "mem_percent_max": 82.3,
      "mem_percent_avg": 70.5,
      "mem_percent_min": 55.0,

      // 新增：磁盘利用率百分比
      "disk_percent_max": 68.9,
      "disk_percent_avg": 50.1,
      "disk_percent_min": 40.2
    },
    {
      "id": 2,
      "cluster_name": "MySQL主集群",
      "cluster_group_name": "生产环境",
      "ip": "192.168.1.101",
      "host_name": "db-server-02",
      "port": 3306,
      "instance_role": "slave",
      "total_memory": 64.0,
      "used_memory": 30.2,
      "total_disk": 500.0,
      "used_disk": 248.5,
      "cpu_cores": 8,
      "cpu_load": 3.8,
      "date_time": "2024-01-01T12:00:00Z",

      "cpu_percent_max": 72.0,
      "cpu_percent_avg": 48.5,
      "cpu_percent_min": 30.0,
      "mem_percent_max": 79.8,
      "mem_percent_avg": 68.2,
      "mem_percent_min": 52.5,
      "disk_percent_max": 66.5,
      "disk_percent_avg": 49.7,
      "disk_percent_min": 38.9
    }
  ]
}
```

### 测试命令

```bash
# 查询所有集群资源
curl -X GET "http://localhost:8888/api/cmdb/v1/cluster-resources"

# 查询特定集群的资源
curl -X GET "http://localhost:8888/api/cmdb/v1/cluster-resources?cluster_name=MySQL主集群"

# 查询特定集群组的资源
curl -X GET "http://localhost:8888/api/cmdb/v1/cluster-resources?group_name=生产环境"

# 查询特定时间范围的集群资源
curl -X GET "http://localhost:8888/api/cmdb/v1/cluster-resources?start_date=2024-01-01%2000:00:00&end_date=2024-01-31%2023:59:59"
```

---

## 4. 集群资源最大值查询接口

### 接口信息

- **接口路径**: `/api/cmdb/v1/cluster-resources-max`
- **请求方法**: `GET`
- **接口描述**: 获取集群资源最大利用率信息（聚合的集群级别统计数据）

### 请求参数

| 参数名 | 类型 | 必填 | 说明 |
|--------|------|------|------|
| start_date | string | 否 | 开始日期时间，格式：YYYY-MM-DD HH:MM:SS |
| end_date | string | 否 | 结束日期时间，格式：YYYY-MM-DD HH:MM:SS |
| cluster_name | string | 否 | 集群名称 |
| group_name | string | 否 | 集群组名称 |

### 响应格式

```json
{
  "success": true,
  "message": "查询成功",
  "list": [
    {
      "cluster_name": "MySQL主集群",
      "cluster_group_name": "生产环境",
      "department_name": "核心业务线",
      "node_count": 3,

      // 平均值字段
      "avg_cpu_load": 4.2,
      "avg_memory_usage": 30.5,
      "avg_disk_usage": 245.8,

      // 最大值字段
      "max_cpu_load": 6.8,
      "max_memory_usage": 52.8,
      "max_disk_usage": 344.5,

      "total_memory": 192.0,
      "total_disk": 1500.0,
      "max_used_memory": 158.4,
      "max_used_disk": 1033.5,
      "max_date_time": "2024-01-15T14:30:00Z",

      // 新增：CPU 利用率百分比（最大值、平均值、最小值）
      "cpu_percent_max": 85.0,
      "cpu_percent_avg": 52.5,
      "cpu_percent_min": 37.5,

      // 新增：内存利用率百分比（最大值、平均值、最小值）
      "mem_percent_max": 82.5,
      "mem_percent_avg": 68.2,
      "mem_percent_min": 55.3,

      // 新增：磁盘利用率百分比（最大值、平均值、最小值）
      "disk_percent_max": 68.9,
      "disk_percent_avg": 53.1,
      "disk_percent_min": 40.5,

      "member_nodes": []
    }
  ]
}
```

### 测试命令

```bash
# 查询所有集群资源最大值
curl -X GET "http://localhost:8888/api/cmdb/v1/cluster-resources-max"

# 查询特定集群的资源最大值
curl -X GET "http://localhost:8888/api/cmdb/v1/cluster-resources-max?cluster_name=MySQL主集群"

# 查询特定集群组的资源最大值
curl -X GET "http://localhost:8888/api/cmdb/v1/cluster-resources-max?group_name=生产环境"

# 查询特定时间范围的集群资源最大值
curl -X GET "http://localhost:8888/api/cmdb/v1/cluster-resources-max?start_date=2024-01-01%2000:00:00&end_date=2024-01-31%2023:59:59"
```

---

## 5. 百分比字段说明

所有接口都新增了 9 个百分比字段，用于表示资源利用率的百分比值：

### CPU 利用率百分比

- **cpu_percent_max** (float64): CPU 利用率最大值（%），范围 0-100
- **cpu_percent_avg** (float64): CPU 利用率平均值（%），范围 0-100
- **cpu_percent_min** (float64): CPU 利用率最小值（%），范围 0-100

### 内存利用率百分比

- **mem_percent_max** (float64): 内存利用率最大值（%），范围 0-100
- **mem_percent_avg** (float64): 内存利用率平均值（%），范围 0-100
- **mem_percent_min** (float64): 内存利用率最小值（%），范围 0-100

### 磁盘利用率百分比

- **disk_percent_max** (float64): 磁盘利用率最大值（%），范围 0-100
- **disk_percent_avg** (float64): 磁盘利用率平均值（%），范围 0-100
- **disk_percent_min** (float64): 磁盘利用率最小值（%），范围 0-100

### 字段来源

- **Elasticsearch 数据源**: 提供 Max 和 Avg 值，Min 值设为 0
- **CMSys 数据源**: 仅提供 Max 值，Avg 使用 Max 值填充，Min 值设为 0

### 字段用途

1. **资源容量规划**: 通过最大值了解峰值负载
2. **性能分析**: 通过平均值了解常态负载
3. **告警阈值设置**: 基于百分比设置告警规则
4. **趋势分析**: 跟踪资源利用率随时间的变化

---

## 6. 测试示例

### 前置条件

1. 确保 RPC 服务和 API 服务都已启动
2. 确保数据库中有监控数据

### 启动服务

```bash
# 启动 RPC 服务
cd /Users/xuguoqiang/SynologyDrive/Backup/MI_office_notebook/D/myworkspace/nucc_workspace/program/src/nucc.com/cmpool_cursor/cmdb_backend_v2/rpc
./cmdb-rpc -f etc/cmpool.yaml &

# 启动 API 服务
cd /Users/xuguoqiang/SynologyDrive/Backup/MI_office_notebook/D/myworkspace/nucc_workspace/program/src/nucc.com/cmpool_cursor/cmdb_backend_v2/api
./cmdb-api -f etc/cmdb-api.yaml &
```

### 测试场景 1: 查询主机资源详细数据

```bash
# 查询所有主机资源（带百分比字段）
curl -X GET "http://localhost:8888/api/cmdb/v1/server-resources" | jq

# 验证响应中包含百分比字段
curl -X GET "http://localhost:8888/api/cmdb/v1/server-resources" | jq '.list[0] | {ip, cpu_percent_max, cpu_percent_avg, mem_percent_max, disk_percent_max}'
```

### 测试场景 2: 查询主机资源最大值

```bash
# 查询主机资源最大值
curl -X GET "http://localhost:8888/api/cmdb/v1/server-resources-max" | jq

# 查看特定主机的百分比数据
curl -X GET "http://localhost:8888/api/cmdb/v1/server-resources-max?ip=192.168.1.100" | jq '.list[0] | {ip, host_name, cpu_percent_max, mem_percent_max, disk_percent_max}'
```

### 测试场景 3: 查询集群资源详细数据

```bash
# 查询特定集群的资源数据
curl -X GET "http://localhost:8888/api/cmdb/v1/cluster-resources?cluster_name=MySQL主集群" | jq

# 验证集群成员的百分比字段
curl -X GET "http://localhost:8888/api/cmdb/v1/cluster-resources?cluster_name=MySQL主集群" | jq '.list[] | {ip, instance_role, cpu_percent_avg, mem_percent_avg, disk_percent_avg}'
```

### 测试场景 4: 查询集群资源最大值

```bash
# 查询集群级别的聚合数据
curl -X GET "http://localhost:8888/api/cmdb/v1/cluster-resources-max" | jq

# 查看特定集群的百分比统计
curl -X GET "http://localhost:8888/api/cmdb/v1/cluster-resources-max?cluster_name=MySQL主集群" | jq '.list[0] | {cluster_name, node_count, cpu_percent_max, cpu_percent_avg, mem_percent_max, disk_percent_max}'
```

### 测试场景 5: 时间范围查询

```bash
# 查询最近一周的数据
START_DATE=$(date -v-7d +%Y-%m-%d)
END_DATE=$(date +%Y-%m-%d)
curl -X GET "http://localhost:8888/api/cmdb/v1/server-resources?start_date=${START_DATE}&end_date=${END_DATE}" | jq '.list | length'

# 查询指定月份的数据
curl -X GET "http://localhost:8888/api/cmdb/v1/server-resources-max?start_date=2024-01-01&end_date=2024-01-31" | jq
```

### 验证百分比字段的正确性

```bash
# 验证百分比值在合理范围内（0-100）
curl -X GET "http://localhost:8888/api/cmdb/v1/server-resources-max" | jq '.list[] | select(.cpu_percent_max > 100 or .cpu_percent_max < 0 or .mem_percent_max > 100 or .mem_percent_max < 0 or .disk_percent_max > 100 or .disk_percent_max < 0)'

# 如果没有输出，说明所有百分比值都在合理范围内
```

### 性能测试

```bash
# 测试接口响应时间
time curl -X GET "http://localhost:8888/api/cmdb/v1/server-resources-max"

# 使用 ab 工具进行压力测试（如果可用）
ab -n 100 -c 10 "http://localhost:8888/api/cmdb/v1/server-resources-max"
```

---

## 附录：错误响应格式

所有接口在出现错误时，都会返回以下格式的响应：

```json
{
  "success": false,
  "message": "错误描述信息",
  "list": []
}
```

常见错误：

- **时间解析失败**: "解析开始时间失败"
- **RPC 调用失败**: "调用RPC服务失败"
- **数据库查询失败**: "查询数据失败"

---

## 更新历史

- **2024-10-16**: 添加资源利用率百分比字段（cpu/mem/disk_percent_max/avg/min）
- **2024-10-16**: 更新所有4个查询接口以返回百分比数据

---

**文档生成时间**: 2024-10-16
**版本**: v2.0
**维护者**: CMDB 开发团队
