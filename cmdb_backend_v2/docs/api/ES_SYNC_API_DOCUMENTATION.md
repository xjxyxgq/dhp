# ES数据同步功能API接口文档 - API文档：适合接口对接

本文档描述了ES数据同步功能的所有HTTP API接口。

## 基础信息

- **Base URL**: `http://localhost:8888`
- **Content-Type**: `application/json`
- **认证**: 需要在Header中携带Token (`Authorization: Bearer <token>`)

## 接口列表

### 1. 任务配置管理

#### 1.1 创建ES同步任务

**接口地址**: `POST /api/cmdb/v1/es-sync-tasks`

**请求参数**:
```json
{
  "task_name": "daily-sync-task",
  "description": "每日数据同步任务",
  "es_endpoint": "http://phoenix.local.com/platform/query/es",
  "es_index_pattern": "cluster*:data-zabbix-host-monitor-*",
  "cron_expression": "0 0 2 * * ?",
  "query_time_range": "30d",
  "created_by": "admin"
}
```

**参数说明**:
| 参数名 | 类型 | 必填 | 说明 |
|--------|------|------|------|
| task_name | string | 是 | 任务名称(唯一) |
| description | string | 否 | 任务描述 |
| es_endpoint | string | 否 | ES接口地址，为空则使用默认配置 |
| es_index_pattern | string | 否 | ES索引模式，为空则使用默认配置 |
| cron_expression | string | 是 | Cron表达式(支持秒级) |
| query_time_range | string | 否 | 查询时间范围(如7d,30d) |
| created_by | string | 否 | 创建人 |

**响应示例**:
```json
{
  "success": true,
  "message": "创建任务成功",
  "task_id": 1
}
```

#### 1.2 获取任务列表

**接口地址**: `GET /api/cmdb/v1/es-sync-tasks`

**请求参数**:
| 参数名 | 类型 | 必填 | 说明 |
|--------|------|------|------|
| enabled_only | boolean | 否 | 是否只返回启用的任务 |

**响应示例**:
```json
{
  "success": true,
  "message": "获取成功",
  "tasks": [
    {
      "id": 1,
      "task_name": "daily-sync-task",
      "description": "每日数据同步任务",
      "es_endpoint": "http://phoenix.local.com/platform/query/es",
      "es_index_pattern": "cluster*:data-zabbix-host-monitor-*",
      "cron_expression": "0 0 2 * * ?",
      "query_time_range": "30d",
      "is_enabled": true,
      "created_by": "admin",
      "created_at": "2025-01-01 10:00:00",
      "updated_at": "2025-01-01 10:00:00",
      "last_execution_time": "2025-01-02 02:00:00",
      "next_execution_time": "2025-01-03 02:00:00"
    }
  ]
}
```

#### 1.3 获取任务详情

**接口地址**: `GET /api/cmdb/v1/es-sync-tasks/:task_id`

**路径参数**:
| 参数名 | 类型 | 说明 |
|--------|------|------|
| task_id | int64 | 任务ID |

**响应示例**:
```json
{
  "success": true,
  "message": "获取成功",
  "task": {
    "id": 1,
    "task_name": "daily-sync-task",
    "description": "每日数据同步任务",
    "es_endpoint": "http://phoenix.local.com/platform/query/es",
    "es_index_pattern": "cluster*:data-zabbix-host-monitor-*",
    "cron_expression": "0 0 2 * * ?",
    "query_time_range": "30d",
    "is_enabled": true,
    "created_by": "admin",
    "created_at": "2025-01-01 10:00:00",
    "updated_at": "2025-01-01 10:00:00",
    "last_execution_time": "2025-01-02 02:00:00",
    "next_execution_time": "2025-01-03 02:00:00"
  }
}
```

#### 1.4 更新任务配置

**接口地址**: `PUT /api/cmdb/v1/es-sync-tasks/:task_id`

**路径参数**:
| 参数名 | 类型 | 说明 |
|--------|------|------|
| task_id | int64 | 任务ID |

**请求参数**:
```json
{
  "task_name": "daily-sync-task-updated",
  "description": "更新后的任务描述",
  "es_endpoint": "http://phoenix.local.com/platform/query/es",
  "es_index_pattern": "cluster*:data-zabbix-host-monitor-*",
  "cron_expression": "0 0 3 * * ?",
  "query_time_range": "7d"
}
```

**响应示例**:
```json
{
  "success": true,
  "message": "更新任务成功"
}
```

**注意**: 更新任务后，如果任务已启用，调度器会自动注销旧配置并重新注册新配置。

#### 1.5 删除任务配置

**接口地址**: `DELETE /api/cmdb/v1/es-sync-tasks/:task_id`

**路径参数**:
| 参数名 | 类型 | 说明 |
|--------|------|------|
| task_id | int64 | 任务ID |

**响应示例**:
```json
{
  "success": true,
  "message": "删除任务成功"
}
```

**注意**: 删除任务时，如果任务已启用，会自动从调度器注销。

#### 1.6 启用/禁用任务

**接口地址**: `PUT /api/cmdb/v1/es-sync-tasks/enable`

**请求参数**:
```json
{
  "id": 1,
  "is_enabled": true
}
```

**参数说明**:
| 参数名 | 类型 | 必填 | 说明 |
|--------|------|------|------|
| id | int64 | 是 | 任务ID |
| is_enabled | boolean | 是 | 是否启用 |

**响应示例**:
```json
{
  "success": true,
  "message": "任务已启用"
}
```

**注意**: 启用任务时会自动注册到调度器，禁用时会从调度器注销。

### 2. 数据同步执行

#### 2.1 根据主机列表立即执行同步

**接口地址**: `POST /api/cmdb/v1/es-sync-execute`

**请求参数**:
```json
{
  "task_name": "manual-sync",
  "host_ip_list": ["10.1.1.1", "10.1.1.2", "10.1.1.3"],
  "query_time_range": "30d",
  "es_endpoint": "",
  "es_index_pattern": ""
}
```

**参数说明**:
| 参数名 | 类型 | 必填 | 说明 |
|--------|------|------|------|
| task_name | string | 是 | 任务名称(用于记录) |
| host_ip_list | array | 是 | 主机IP列表 |
| query_time_range | string | 否 | 查询时间范围，默认30d |
| es_endpoint | string | 否 | ES接口地址(为空则使用默认配置) |
| es_index_pattern | string | 否 | ES索引模式(为空则使用默认配置) |

**响应示例**:
```json
{
  "success": true,
  "message": "同步完成: 成功98个, 失败2个",
  "execution_id": 100
}
```

**注意**:
- 此接口会立即执行同步，并返回执行记录ID
- 同步过程中会检查主机是否在 hosts_pool 中
- 同步结果会写入 server_resources 表

#### 2.2 根据文件立即执行同步

**接口地址**: `POST /api/cmdb/v1/es-sync-execute-file`

**请求参数**:
```json
{
  "task_name": "file-sync",
  "file_content": "10.1.1.1\n10.1.1.2\n10.1.1.3",
  "query_time_range": "30d",
  "es_endpoint": ""
}
```

**参数说明**:
| 参数名 | 类型 | 必填 | 说明 |
|--------|------|------|------|
| task_name | string | 是 | 任务名称 |
| file_content | string | 是 | 文件内容(每行一个IP) |
| query_time_range | string | 否 | 查询时间范围 |
| es_endpoint | string | 否 | ES接口地址 |

**响应示例**:
```json
{
  "success": true,
  "message": "同步完成: 成功98个, 失败2个",
  "execution_id": 101
}
```

**文件格式要求**:
- 每行一个IP地址
- 支持空行（会被忽略）
- 支持注释行（#开头）

#### 2.3 执行ES全量同步

**接口地址**: `POST /api/cmdb/v1/es-sync-full-sync`

**请求参数**:
```json
{
  "group_name": "DB组",
  "query_time_range": "30d",
  "es_endpoint": "",
  "task_name": "ES全量同步"
}
```

**参数说明**:
| 参数名 | 类型 | 必填 | 说明 |
|--------|------|------|------|
| group_name | string | 否 | ES中的group字段值，默认"DB组" |
| query_time_range | string | 否 | 查询时间范围，默认"30d" |
| es_endpoint | string | 否 | ES接口地址(为空则使用默认配置) |
| task_name | string | 否 | 任务名称(用于记录)，默认"ES全量同步" |

**响应示例**:
```json
{
  "success": true,
  "message": "同步完成: 新增15个, 更新120个, 失败3个",
  "execution_id": 102,
  "total_hosts": 138,
  "new_hosts_count": 15,
  "updated_hosts_count": 120,
  "failed_count": 3,
  "new_host_ip_list": ["10.1.2.1", "10.1.2.2", "10.1.2.3"],
  "updated_host_ip_list": ["10.1.1.1", "10.1.1.2", "..."],
  "failed_ip_list": ["10.1.3.1", "10.1.3.2", "10.1.3.3"]
}
```

**功能特点**:
1. **自动发现新主机**: 从ES中查询指定group的所有主机
2. **自动注册**: 将不在hosts_pool中的主机自动插入
3. **批量更新**: 更新所有主机的资源数据到server_resources表
4. **详细反馈**: 返回新增、更新、失败的主机IP列表

**与其他同步接口的区别**:
| 特性 | 全量同步 | 主机列表同步 | 文件同步 |
|------|---------|-------------|---------|
| 数据来源 | ES按group查询 | 用户提供IP列表 | 用户上传文件 |
| 新主机处理 | 自动插入hosts_pool | 跳过(not_in_pool) | 跳过(not_in_pool) |
| 适用场景 | 初始化/全量更新 | 指定主机同步 | 批量指定主机 |

**注意事项**:
- 此接口会自动将ES中存在但hosts_pool中不存在的主机添加到hosts_pool
- 适合用于初始化数据或定期全量更新
- 可能返回大量主机IP列表，建议关注统计数据

### 3. 执行记录查询

#### 3.1 获取执行记录列表

**接口地址**: `GET /api/cmdb/v1/es-sync-execution-logs`

**请求参数**:
| 参数名 | 类型 | 必填 | 说明 |
|--------|------|------|------|
| task_id | int64 | 否 | 任务ID(为0则返回所有) |
| limit | int32 | 否 | 限制返回记录数(默认50) |

**响应示例**:
```json
{
  "success": true,
  "message": "获取成功",
  "execution_logs": [
    {
      "id": 100,
      "task_id": 1,
      "task_name": "daily-sync-task",
      "execution_time": "2025-01-02 02:00:00",
      "execution_status": "success",
      "total_hosts": 100,
      "success_count": 98,
      "failed_count": 2,
      "not_in_pool_count": 0,
      "error_message": "",
      "duration_ms": 30000,
      "query_time_range": "30d",
      "created_at": "2025-01-02 02:00:00"
    }
  ]
}
```

**执行状态说明**:
| 状态 | 说明 |
|------|------|
| running | 正在运行 |
| success | 全部成功 |
| partial | 部分成功 |
| failed | 全部失败 |

#### 3.2 获取执行详情

**接口地址**: `GET /api/cmdb/v1/es-sync-execution-detail/:execution_id`

**路径参数**:
| 参数名 | 类型 | 说明 |
|--------|------|------|
| execution_id | int64 | 执行记录ID |

**响应示例**:
```json
{
  "success": true,
  "message": "获取成功",
  "execution_log": {
    "id": 100,
    "task_id": 1,
    "task_name": "daily-sync-task",
    "execution_time": "2025-01-02 02:00:00",
    "execution_status": "success",
    "total_hosts": 100,
    "success_count": 98,
    "failed_count": 2,
    "not_in_pool_count": 0,
    "error_message": "",
    "duration_ms": 30000,
    "query_time_range": "30d",
    "created_at": "2025-01-02 02:00:00"
  },
  "execution_details": [
    {
      "host_ip": "10.1.1.1",
      "host_name": "host-001",
      "sync_status": "success",
      "error_message": "",
      "max_cpu": 85.5,
      "avg_cpu": 45.2,
      "max_memory": 120.5,
      "avg_memory": 80.3,
      "max_disk": 500.0,
      "avg_disk": 350.0,
      "data_point_count": 2880,
      "created_at": "2025-01-02 02:05:00"
    },
    {
      "host_ip": "10.1.1.2",
      "host_name": "host-002",
      "sync_status": "no_data",
      "error_message": "ES中无数据",
      "max_cpu": 0,
      "avg_cpu": 0,
      "max_memory": 0,
      "avg_memory": 0,
      "max_disk": 0,
      "avg_disk": 0,
      "data_point_count": 0,
      "created_at": "2025-01-02 02:05:00"
    },
    {
      "host_ip": "10.1.1.3",
      "host_name": "host-003",
      "sync_status": "failed",
      "error_message": "ES查询超时",
      "max_cpu": 0,
      "avg_cpu": 0,
      "max_memory": 0,
      "avg_memory": 0,
      "max_disk": 0,
      "avg_disk": 0,
      "data_point_count": 0,
      "created_at": "2025-01-02 02:05:00"
    },
    {
      "host_ip": "10.1.1.99",
      "host_name": "unknown-host",
      "sync_status": "not_in_pool",
      "error_message": "主机不在hosts_pool中",
      "max_cpu": 0,
      "avg_cpu": 0,
      "max_memory": 0,
      "avg_memory": 0,
      "max_disk": 0,
      "avg_disk": 0,
      "data_point_count": 0,
      "created_at": "2025-01-02 02:05:00"
    }
  ]
}
```

**同步状态说明**:
| 状态 | 说明 |
|------|------|
| success | 同步成功 |
| failed | 同步失败（ES查询失败等错误） |
| no_data | ES查询成功但无数据（数据点为0） |
| not_in_pool | 主机不在hosts_pool中（仅手动同步） |

**注意**:
- `no_data` 和 `failed` 是两种不同的错误状态，便于问题定位
- `not_in_pool` 状态仅在手动同步（按IP列表或文件）时出现
- 全量同步会自动插入新主机，不会出现 `not_in_pool` 状态

## 错误码

| 错误码 | 说明 |
|--------|------|
| 200 | 成功 |
| 400 | 请求参数错误 |
| 401 | 未授权 |
| 404 | 资源不存在 |
| 500 | 服务器内部错误 |

## 通用响应格式

所有接口都遵循以下响应格式:

```json
{
  "success": true,   // 操作是否成功
  "message": "操作成功",  // 提示消息
  // ... 其他数据字段
}
```

## 使用示例

### 示例1: 创建定时任务

```bash
curl -X POST http://localhost:8888/api/cmdb/v1/es-sync-tasks \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer <token>" \
  -d '{
    "task_name": "hourly-sync",
    "description": "每小时同步一次",
    "es_endpoint": "http://phoenix.local.com/platform/query/es",
    "es_index_pattern": "cluster*:data-zabbix-host-monitor-*",
    "cron_expression": "0 0 * * * ?",
    "query_time_range": "1h",
    "created_by": "admin"
  }'
```

### 示例2: 手动执行同步

```bash
curl -X POST http://localhost:8888/api/cmdb/v1/es-sync-execute \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer <token>" \
  -d '{
    "task_name": "manual-test",
    "host_ip_list": ["10.1.1.1", "10.1.1.2"],
    "query_time_range": "7d"
  }'
```

### 示例3: 查询执行记录

```bash
curl -X GET "http://localhost:8888/api/cmdb/v1/es-sync-execution-logs?task_id=1&limit=10" \
  -H "Authorization: Bearer <token>"
```

### 示例4: 查看执行详情

```bash
curl -X GET http://localhost:8888/api/cmdb/v1/es-sync-execution-detail/100 \
  -H "Authorization: Bearer <token>"
```

### 示例5: 启用任务

```bash
curl -X PUT http://localhost:8888/api/cmdb/v1/es-sync-tasks/enable \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer <token>" \
  -d '{
    "id": 1,
    "is_enabled": true
  }'
```

### 示例6: 执行ES全量同步

```bash
curl -X POST http://localhost:8888/api/cmdb/v1/es-sync-full-sync \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer <token>" \
  -d '{
    "group_name": "DB组",
    "query_time_range": "30d",
    "task_name": "初始化同步"
  }'
```

**响应示例**:
```json
{
  "success": true,
  "message": "同步完成: 新增15个, 更新120个, 失败3个",
  "execution_id": 102,
  "total_hosts": 138,
  "new_hosts_count": 15,
  "updated_hosts_count": 120,
  "failed_count": 3,
  "new_host_ip_list": [
    "10.1.2.1", "10.1.2.2", "10.1.2.3", "10.1.2.4", "10.1.2.5",
    "10.1.2.6", "10.1.2.7", "10.1.2.8", "10.1.2.9", "10.1.2.10",
    "10.1.2.11", "10.1.2.12", "10.1.2.13", "10.1.2.14", "10.1.2.15"
  ],
  "updated_host_ip_list": [
    "10.1.1.1", "10.1.1.2", "10.1.1.3"
  ],
  "failed_ip_list": [
    "10.1.3.1", "10.1.3.2", "10.1.3.3"
  ]
}
```

## 数据流程说明

### 定时同步流程

1. 用户创建定时任务配置
2. 用户启用任务，系统自动注册到调度器
3. 调度器根据cron表达式定时触发任务
4. 任务执行时从hosts_pool获取主机列表
5. 对每个主机并发调用ES接口查询监控数据（最大并发数10）
6. 检查主机是否在hosts_pool中，只同步在pool中的主机
7. 将ES查询的最大CPU、最大内存、最大磁盘数据写入server_resources表
8. 记录执行结果到执行记录表和详情表

### 手动同步流程

1. 用户提供主机IP列表或上传文件
2. 系统立即创建执行记录（状态为running）
3. 对每个主机并发查询ES数据
4. 检查主机是否在hosts_pool中
5. 写入server_resources表（使用upsert操作）
6. 保存执行详情
7. 更新执行记录状态

### 全量同步流程

1. 用户调用全量同步接口，指定group名称（默认"DB组"）
2. 系统向ES查询指定group的所有主机数据（按hostIp聚合）
3. 系统创建执行记录（状态为running）
4. 对每个主机并发处理（最大并发数10）:
   - 检查主机是否在hosts_pool中
   - 如果不在，自动插入新主机到hosts_pool
   - 将ES查询的资源数据写入server_resources表
   - 记录是新增还是更新
5. 保存执行详情（包含新增、更新、失败的主机信息）
6. 更新执行记录状态，返回详细统计信息

**全量同步与手动同步的区别**:
- **数据来源**: 全量同步从ES查询group数据，手动同步使用用户提供的IP列表
- **新主机处理**: 全量同步会自动插入新主机到hosts_pool，手动同步会跳过不在pool中的主机
- **返回信息**: 全量同步返回新增/更新/失败的IP列表，手动同步返回成功/失败/不在池中的IP列表
- **适用场景**: 全量同步适合初始化或全量更新，手动同步适合指定主机的定向更新

## 调度器说明

### 调度器特性

- **自动启动**: RPC服务启动时自动创建并启动调度器
- **动态注册**: 启用任务时自动注册到调度器，禁用时自动注销
- **配置更新**: 更新任务配置时，如果任务已启用，会自动重新注册
- **线程安全**: 使用RWMutex保证并发操作的线程安全
- **优雅停止**: RPC服务停止时自动停止调度器

### Cron表达式格式

支持标准Cron表达式，格式为6个字段（支持秒级精度）:

```
秒 分 时 日 月 周
0  0  2  *  *  ?    # 每天凌晨2点执行
0  0  */2 * * ?    # 每2小时执行一次
0  0  0  1  * ?    # 每月1号零点执行
```

## 数据同步详解

### ES数据查询

- **查询字段**: 从ES查询CPU、内存、磁盘使用率的最大值和平均值
- **时间范围**: 支持相对时间（如7d、30d、1h、24h）
- **数据点统计**: 记录查询到的数据点数量

### 数据写入

- **目标表**: `server_resources`
- **字段映射**:
  - `pool_id`: 从hosts_pool查询得到
  - `ip`: 主机IP
  - `cpu_load`: ES中的CPU最大值
  - `used_memory`: ES中的内存最大值
  - `total_disk`: ES中的磁盘总空间最大值
- **更新策略**: 使用UPSERT操作，相同pool_id的记录会被更新

### 并发控制

- 使用信号量控制最大并发数为10
- 使用WaitGroup等待所有协程完成
- 使用Mutex保护共享计数器

## 注意事项

1. **时间范围格式**: 支持的格式包括 `7d`(7天)、`30d`(30天)、`1h`(1小时)、`24h`(24小时)等
2. **Cron表达式**: 使用标准Cron表达式,支持秒级精度（6个字段）
3. **主机过滤**:
   - **手动同步**：只有在hosts_pool中的主机才会同步数据，不在pool中的主机会被标记为`not_in_pool`
   - **全量同步**：会自动将ES中的主机添加到hosts_pool，然后同步数据
4. **数据覆盖**: 同一主机的数据会被新数据覆盖（UPSERT操作）
5. **文件格式**: 上传的主机列表文件每行一个IP地址,支持空行和注释(#开头)
6. **调度器状态**: 任务启用/禁用状态变更时，调度器会自动注册/注销任务
7. **配置更新**: 更新已启用任务的配置时，调度器会自动重新加载新配置
8. **并发性能**: 单次同步最大并发查询10个主机，适当控制主机列表大小
9. **ES无数据处理**:
   - 如果ES中查询不到主机数据（数据点为0），会被标记为`no_data`状态
   - `no_data`和`failed`（查询失败）是两种不同的错误状态，便于问题排查
10. **全量同步建议**:
    - 首次使用或需要初始化数据时使用全量同步
    - 定期使用全量同步发现新主机并更新数据
    - 全量同步会返回详细的新增/更新/失败IP列表，便于审计和验证

## 相关资源

- 实现指南: `ES_SYNC_IMPLEMENTATION_GUIDE.md`
- 数据库表结构: `source/schema.sql`
- Proto接口定义: `rpc/proto/cmpool.proto`
- 实现总结: `ES_SYNC_IMPLEMENTATION_SUMMARY.md`
