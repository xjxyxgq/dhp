# 统一外部资源同步 API 文档

## 接口概述

本文档描述了统一的外部资源同步 API，该 API 提供了一套标准化的接口来管理和执行来自不同数据源（Elasticsearch 和 CMSys）的资源数据同步任务。

### 设计理念

1. **统一接口设计**：所有外部资源同步操作使用统一的 API 接口
2. **数据源参数化**：通过 `data_source` 参数控制使用哪种数据源
3. **向后兼容**：保留原有的 ES 和 CMSys 专用接口，确保现有系统无缝升级
4. **前端友好**：减少前端适配成本，统一的请求/响应格式

### 支持的数据源

- **Elasticsearch (ES)**：值为 `"elasticsearch"` 或 `"es"`
- **CMSys**：值为 `"cmsys"`

## 认证说明

所有接口都需要在请求头中携带认证 Token：

```
Authorization: Bearer YOUR_AUTH_TOKEN
```

## API 接口列表

### 1. 任务管理接口

#### 1.1 创建同步任务

**接口地址**
```
POST /api/cmdb/v1/external-sync-tasks
```

**请求体**
```json
{
  "task_name": "每日数据同步",
  "description": "每天凌晨2点同步数据库组主机数据",
  "data_source": "elasticsearch",
  "es_endpoint": "http://es.example.com:9200",
  "es_index_pattern": "metricbeat-*",
  "cron_expression": "0 2 * * *",
  "query_time_range": "7d",
  "created_by": "admin"
}
```

**参数说明**

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| task_name | string | 是 | 任务名称 |
| description | string | 否 | 任务描述 |
| data_source | string | 是 | 数据源类型："elasticsearch"/"es" 或 "cmsys" |
| es_endpoint | string | 条件 | ES接口地址（data_source=es时使用） |
| es_index_pattern | string | 条件 | ES索引模式（data_source=es时使用） |
| cron_expression | string | 是 | Cron表达式，如 "0 2 * * *" 表示每天凌晨2点 |
| query_time_range | string | 否 | 查询时间范围，如 "7d"、"30d" |
| created_by | string | 否 | 创建人 |

**响应示例**
```json
{
  "success": true,
  "message": "任务创建成功",
  "task_id": 123
}
```

**curl 示例**

Elasticsearch数据源：
```bash
curl -X POST 'http://localhost:8888/api/cmdb/v1/external-sync-tasks' \
  -H 'Content-Type: application/json' \
  -H 'Authorization: Bearer YOUR_TOKEN' \
  -d '{
    "task_name": "ES每日同步",
    "description": "同步ES中的主机监控数据",
    "data_source": "elasticsearch",
    "es_endpoint": "http://es.example.com:9200",
    "es_index_pattern": "metricbeat-*",
    "cron_expression": "0 2 * * *",
    "query_time_range": "7d"
  }'
```

CMSys数据源：
```bash
curl -X POST 'http://localhost:8888/api/cmdb/v1/external-sync-tasks' \
  -H 'Content-Type: application/json' \
  -H 'Authorization: Bearer YOUR_TOKEN' \
  -d '{
    "task_name": "CMSys每日同步",
    "description": "同步CMSys中的主机负载数据",
    "data_source": "cmsys",
    "cron_expression": "0 3 * * *",
    "query_time_range": "30d"
  }'
```

#### 1.2 更新同步任务

**接口地址**
```
PUT /api/cmdb/v1/external-sync-tasks/:task_id
```

**请求体**
```json
{
  "task_name": "每日数据同步（已更新）",
  "description": "更新后的描述",
  "data_source": "elasticsearch",
  "es_endpoint": "http://es.example.com:9200",
  "es_index_pattern": "metricbeat-*",
  "cron_expression": "0 3 * * *",
  "query_time_range": "14d"
}
```

**curl 示例**
```bash
curl -X PUT 'http://localhost:8888/api/cmdb/v1/external-sync-tasks/123' \
  -H 'Content-Type: application/json' \
  -H 'Authorization: Bearer YOUR_TOKEN' \
  -d '{
    "task_name": "ES每日同步（更新）",
    "cron_expression": "0 3 * * *",
    "query_time_range": "14d"
  }'
```

#### 1.3 删除同步任务

**接口地址**
```
DELETE /api/cmdb/v1/external-sync-tasks/:task_id
```

**curl 示例**
```bash
curl -X DELETE 'http://localhost:8888/api/cmdb/v1/external-sync-tasks/123' \
  -H 'Authorization: Bearer YOUR_TOKEN'
```

**响应示例**
```json
{
  "success": true,
  "message": "任务删除成功"
}
```

#### 1.4 启用/禁用任务

**接口地址**
```
PUT /api/cmdb/v1/external-sync-tasks/enable
```

**请求体**
```json
{
  "id": 123,
  "is_enabled": true
}
```

**curl 示例**
```bash
# 启用任务
curl -X PUT 'http://localhost:8888/api/cmdb/v1/external-sync-tasks/enable' \
  -H 'Content-Type: application/json' \
  -H 'Authorization: Bearer YOUR_TOKEN' \
  -d '{
    "id": 123,
    "is_enabled": true
  }'

# 禁用任务
curl -X PUT 'http://localhost:8888/api/cmdb/v1/external-sync-tasks/enable' \
  -H 'Content-Type: application/json' \
  -H 'Authorization: Bearer YOUR_TOKEN' \
  -d '{
    "id": 123,
    "is_enabled": false
  }'
```

#### 1.5 获取任务列表

**接口地址**
```
GET /api/cmdb/v1/external-sync-tasks
```

**查询参数**

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| data_source | string | 否 | 数据源过滤："elasticsearch" 或 "cmsys" |
| enabled_only | boolean | 否 | 是否只返回启用的任务 |

**curl 示例**
```bash
# 获取所有任务
curl 'http://localhost:8888/api/cmdb/v1/external-sync-tasks' \
  -H 'Authorization: Bearer YOUR_TOKEN'

# 只获取ES数据源的任务
curl 'http://localhost:8888/api/cmdb/v1/external-sync-tasks?data_source=elasticsearch' \
  -H 'Authorization: Bearer YOUR_TOKEN'

# 只获取启用的CMSys任务
curl 'http://localhost:8888/api/cmdb/v1/external-sync-tasks?data_source=cmsys&enabled_only=true' \
  -H 'Authorization: Bearer YOUR_TOKEN'
```

**响应示例**
```json
{
  "success": true,
  "message": "查询成功",
  "tasks": [
    {
      "id": 123,
      "task_name": "ES每日同步",
      "description": "同步ES中的主机监控数据",
      "data_source": "elasticsearch",
      "es_endpoint": "http://es.example.com:9200",
      "es_index_pattern": "metricbeat-*",
      "cron_expression": "0 2 * * *",
      "query_time_range": "7d",
      "is_enabled": true,
      "created_by": "admin",
      "created_at": "2025-01-21 10:00:00",
      "updated_at": "2025-01-21 10:00:00",
      "last_execution_time": "2025-01-22 02:00:00",
      "next_execution_time": "2025-01-23 02:00:00"
    }
  ]
}
```

#### 1.6 获取任务详情

**接口地址**
```
GET /api/cmdb/v1/external-sync-tasks/:task_id
```

**curl 示例**
```bash
curl 'http://localhost:8888/api/cmdb/v1/external-sync-tasks/123' \
  -H 'Authorization: Bearer YOUR_TOKEN'
```

### 2. 执行同步接口

#### 2.1 按IP列表执行同步

**接口地址**
```
POST /api/cmdb/v1/external-sync-execute
```

**请求体**

Elasticsearch数据源：
```json
{
  "data_source": "elasticsearch",
  "task_name": "手动同步指定IP",
  "host_ip_list": ["192.168.1.1", "192.168.1.2", "10.0.0.100"],
  "es_endpoint": "http://es.example.com:9200",
  "es_index_pattern": "metricbeat-*",
  "query_time_range": "7d"
}
```

CMSys数据源：
```json
{
  "data_source": "cmsys",
  "task_name": "手动同步指定IP",
  "host_ip_list": ["192.168.1.1", "192.168.1.2"],
  "query": "department=DB"
}
```

**curl 示例**

ES数据源：
```bash
curl -X POST 'http://localhost:8888/api/cmdb/v1/external-sync-execute' \
  -H 'Content-Type: application/json' \
  -H 'Authorization: Bearer YOUR_TOKEN' \
  -d '{
    "data_source": "elasticsearch",
    "task_name": "手动ES同步",
    "host_ip_list": ["192.168.1.1", "192.168.1.2"],
    "query_time_range": "7d"
  }'
```

CMSys数据源：
```bash
curl -X POST 'http://localhost:8888/api/cmdb/v1/external-sync-execute' \
  -H 'Content-Type: application/json' \
  -H 'Authorization: Bearer YOUR_TOKEN' \
  -d '{
    "data_source": "cmsys",
    "task_name": "手动CMSys同步",
    "host_ip_list": ["192.168.1.1", "192.168.1.2"]
  }'
```

**响应示例**（统一结构，所有数据源相同）

ES数据源响应：
```json
{
  "success": true,
  "message": "同步完成: 成功2个, 失败0个, 不在池中0个",
  "data_source": "elasticsearch",
  "execution_id": 456,
  "total_hosts": 2,
  "success_count": 2,
  "failed_count": 0,
  "not_in_pool_count": 0,
  "not_in_datasource_count": 0,
  "new_hosts_count": 0,
  "updated_hosts_count": 0,
  "success_ip_list": ["192.168.1.1", "192.168.1.2"],
  "failed_ip_list": [],
  "not_in_pool_ip_list": [],
  "not_in_datasource_ip_list": [],
  "new_host_ip_list": [],
  "updated_host_ip_list": []
}
```

CMSys数据源响应（字段完全相同）：
```json
{
  "success": true,
  "message": "同步完成: 成功2个, 失败0个, 数据源中不存在0个",
  "data_source": "cmsys",
  "execution_id": 457,
  "total_hosts": 2,
  "success_count": 2,
  "failed_count": 0,
  "not_in_pool_count": 0,
  "not_in_datasource_count": 0,
  "new_hosts_count": 0,
  "updated_hosts_count": 0,
  "success_ip_list": ["192.168.1.1", "192.168.1.2"],
  "failed_ip_list": [],
  "not_in_pool_ip_list": [],
  "not_in_datasource_ip_list": [],
  "new_host_ip_list": [],
  "updated_host_ip_list": []
}
```

#### 2.2 通过文件执行同步

**接口地址**
```
POST /api/cmdb/v1/external-sync-execute-file
```

**请求体**
```json
{
  "data_source": "elasticsearch",
  "task_name": "批量同步",
  "file_content": "192.168.1.1\n192.168.1.2\n10.0.0.100",
  "query_time_range": "7d"
}
```

**curl 示例**
```bash
curl -X POST 'http://localhost:8888/api/cmdb/v1/external-sync-execute-file' \
  -H 'Content-Type: application/json' \
  -H 'Authorization: Bearer YOUR_TOKEN' \
  -d '{
    "data_source": "elasticsearch",
    "task_name": "文件批量同步",
    "file_content": "192.168.1.1\n192.168.1.2\n10.0.0.100",
    "query_time_range": "7d"
  }'
```

#### 2.3 全量同步

**接口地址**
```
POST /api/cmdb/v1/external-sync-full-sync
```

**请求体**

Elasticsearch全量同步：
```json
{
  "data_source": "elasticsearch",
  "group_name": "DB组",
  "query_time_range": "30d",
  "task_name": "ES全量同步"
}
```

CMSys全量同步：
```json
{
  "data_source": "cmsys",
  "query": "department=DB",
  "task_name": "CMSys全量同步"
}
```

**curl 示例**

ES全量同步：
```bash
curl -X POST 'http://localhost:8888/api/cmdb/v1/external-sync-full-sync' \
  -H 'Content-Type: application/json' \
  -H 'Authorization: Bearer YOUR_TOKEN' \
  -d '{
    "data_source": "elasticsearch",
    "group_name": "DB组",
    "query_time_range": "30d"
  }'
```

CMSys全量同步：
```bash
curl -X POST 'http://localhost:8888/api/cmdb/v1/external-sync-full-sync' \
  -H 'Content-Type: application/json' \
  -H 'Authorization: Bearer YOUR_TOKEN' \
  -d '{
    "data_source": "cmsys",
    "query": "department=DB"
  }'
```

**响应示例**（统一结构，所有数据源相同）

ES全量同步响应：
```json
{
  "success": true,
  "message": "全量同步完成: 成功150个, 失败0个, 新增5个, 更新145个",
  "data_source": "elasticsearch",
  "execution_id": 458,
  "total_hosts": 150,
  "success_count": 0,
  "failed_count": 0,
  "not_in_pool_count": 0,
  "not_in_datasource_count": 0,
  "new_hosts_count": 5,
  "updated_hosts_count": 145,
  "success_ip_list": [],
  "failed_ip_list": [],
  "not_in_pool_ip_list": [],
  "not_in_datasource_ip_list": [],
  "new_host_ip_list": ["10.0.0.1", "10.0.0.2", "10.0.0.3", "10.0.0.4", "10.0.0.5"],
  "updated_host_ip_list": ["192.168.1.1", "192.168.1.2"]
}
```

CMSys全量同步响应（字段完全相同）：
```json
{
  "success": true,
  "message": "全量同步完成: 成功120个, 失败2个, 新增3个, 更新117个",
  "data_source": "cmsys",
  "execution_id": 459,
  "total_hosts": 120,
  "success_count": 118,
  "failed_count": 2,
  "not_in_pool_count": 0,
  "not_in_datasource_count": 0,
  "new_hosts_count": 3,
  "updated_hosts_count": 117,
  "success_ip_list": ["192.168.1.1", "192.168.1.2"],
  "failed_ip_list": ["10.0.0.99", "10.0.0.100"],
  "not_in_pool_ip_list": [],
  "not_in_datasource_ip_list": [],
  "new_host_ip_list": ["10.0.0.6", "10.0.0.7", "10.0.0.8"],
  "updated_host_ip_list": ["192.168.1.3", "192.168.1.4"]
}
```

### 3. 执行日志接口

#### 3.1 获取执行日志列表

**接口地址**
```
GET /api/cmdb/v1/external-sync-execution-logs
```

**查询参数**

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| task_id | int64 | 否 | 任务ID（为空则查询所有） |
| data_source | string | 否 | 数据源过滤 |
| limit | int32 | 否 | 限制返回数量，默认50 |

**curl 示例**
```bash
# 获取所有执行日志
curl 'http://localhost:8888/api/cmdb/v1/external-sync-execution-logs' \
  -H 'Authorization: Bearer YOUR_TOKEN'

# 获取指定任务的执行日志
curl 'http://localhost:8888/api/cmdb/v1/external-sync-execution-logs?task_id=123' \
  -H 'Authorization: Bearer YOUR_TOKEN'

# 获取ES数据源的执行日志
curl 'http://localhost:8888/api/cmdb/v1/external-sync-execution-logs?data_source=elasticsearch&limit=20' \
  -H 'Authorization: Bearer YOUR_TOKEN'
```

**响应示例**
```json
{
  "success": true,
  "message": "查询成功",
  "execution_logs": [
    {
      "id": 456,
      "task_id": 123,
      "task_name": "ES每日同步",
      "data_source": "elasticsearch",
      "execution_time": "2025-01-22 02:00:00",
      "execution_status": "success",
      "total_hosts": 150,
      "success_count": 148,
      "failed_count": 2,
      "not_in_pool_count": 0,
      "duration_ms": 12500,
      "query_time_range": "7d",
      "created_at": "2025-01-22 02:00:00"
    }
  ]
}
```

#### 3.2 获取执行详情

**接口地址**
```
GET /api/cmdb/v1/external-sync-execution-detail/:execution_id
```

**curl 示例**
```bash
curl 'http://localhost:8888/api/cmdb/v1/external-sync-execution-detail/456' \
  -H 'Authorization: Bearer YOUR_TOKEN'
```

**响应示例**
```json
{
  "success": true,
  "message": "查询成功",
  "execution_log": {
    "id": 456,
    "task_name": "ES每日同步",
    "data_source": "elasticsearch",
    "execution_time": "2025-01-22 02:00:00",
    "execution_status": "success",
    "total_hosts": 150,
    "success_count": 148,
    "failed_count": 2
  },
  "execution_details": [
    {
      "host_ip": "192.168.1.1",
      "host_name": "db-server-01",
      "sync_status": "success",
      "max_cpu": 75.5,
      "avg_cpu": 65.2,
      "max_memory": 80.3,
      "avg_memory": 72.1,
      "max_disk": 65.8,
      "avg_disk": 58.4,
      "data_point_count": 168,
      "created_at": "2025-01-22 02:00:15"
    },
    {
      "host_ip": "192.168.1.2",
      "host_name": "db-server-02",
      "sync_status": "failed",
      "error_message": "ES查询超时",
      "created_at": "2025-01-22 02:00:20"
    }
  ]
}
```

## 数据源切换说明

### 切换方式

通过 `data_source` 参数控制使用哪种数据源：

- 使用 Elasticsearch：`"data_source": "elasticsearch"` 或 `"data_source": "es"`
- 使用 CMSys：`"data_source": "cmsys"`

### 请求参数差异

不同数据源需要的请求参数有所不同：

| 参数 | ES使用 | CMSys使用 | 说明 |
|------|--------|-----------|------|
| es_endpoint | 是 | 否 | ES接口地址（可选，有默认配置） |
| es_index_pattern | 是 | 否 | ES索引模式（可选，有默认配置） |
| group_name | 是 | 否 | 组名（全量同步时使用，默认"DB组"） |
| query | 否 | 是 | CMSys查询参数（可选） |

**注意**：响应结构在所有数据源中完全统一，请参见下面的"统一响应结构"章节。

### 统一响应结构

**重要说明**：所有数据源的响应都包含完全相同的字段集合，前端无需根据 `data_source` 判断响应结构。

#### 响应字段完整列表

| 字段名 | 类型 | 说明 | ES填充 | CMSys填充 |
|--------|------|------|--------|-----------|
| success | bool | 是否成功 | ✅ | ✅ |
| message | string | 响应消息 | ✅ | ✅ |
| data_source | string | 数据源类型 | "elasticsearch" | "cmsys" |
| execution_id | int64 | 执行记录ID | ✅ | ✅ |
| total_hosts | int32 | 总主机数 | ✅ | ✅ |
| success_count | int32 | 成功数量 | ✅ | ✅ |
| failed_count | int32 | 失败数量 | ✅ | ✅ |
| not_in_pool_count | int32 | 不在主机池中的数量 | ✅ | 0（固定） |
| not_in_datasource_count | int32 | 数据源中不存在的数量 | 0（固定） | ✅ |
| new_hosts_count | int32 | 新增主机数量 | ✅（全量同步） | ✅（全量同步） |
| updated_hosts_count | int32 | 更新主机数量 | ✅（全量同步） | ✅（全量同步） |
| success_ip_list | []string | 成功的IP列表 | ✅ | ✅ |
| failed_ip_list | []string | 失败的IP列表 | ✅ | ✅ |
| not_in_pool_ip_list | []string | 不在主机池中的IP列表 | ✅ | []（空数组） |
| not_in_datasource_ip_list | []string | 数据源中不存在的IP列表 | []（空数组） | ✅ |
| new_host_ip_list | []string | 新增主机IP列表 | ✅（全量同步） | ✅（全量同步） |
| updated_host_ip_list | []string | 更新主机IP列表 | ✅（全量同步） | ✅（全量同步） |

#### 字段填充规则

1. **所有响应都包含所有字段**，不会缺失任何字段
2. **数据源特定字段**：
   - ES数据源：`not_in_pool_count` 有值，`not_in_datasource_count` 固定为 0
   - CMSys数据源：`not_in_datasource_count` 有值，`not_in_pool_count` 固定为 0
3. **IP列表字段**：不适用的数据源返回空数组 `[]`，而不是 null
4. **全量同步特有字段**：`new_hosts_count`、`updated_hosts_count`、`new_host_ip_list`、`updated_host_ip_list` 仅在全量同步接口中有意义值，其他接口返回 0 或空数组

#### 前端处理建议

**✅ 推荐做法**：直接读取字段值，无需判断 `data_source`

```javascript
// 前端可以直接使用所有字段，无需条件判断
function displaySyncResult(response) {
  console.log(`总主机: ${response.total_hosts}`);
  console.log(`成功: ${response.success_count}`);
  console.log(`失败: ${response.failed_count}`);

  // 直接显示，值为0时自然不突出
  if (response.not_in_pool_count > 0) {
    console.log(`不在池中: ${response.not_in_pool_count}`);
  }
  if (response.not_in_datasource_count > 0) {
    console.log(`数据源中不存在: ${response.not_in_datasource_count}`);
  }

  // 全量同步结果
  if (response.new_hosts_count > 0) {
    console.log(`新增主机: ${response.new_hosts_count}`);
  }
  if (response.updated_hosts_count > 0) {
    console.log(`更新主机: ${response.updated_hosts_count}`);
  }
}
```

**❌ 不推荐做法**：根据 `data_source` 判断响应结构

```javascript
// 不需要这样做！
if (response.data_source === 'elasticsearch') {
  // 所有字段都存在，无需条件判断
  console.log(response.not_in_pool_count);
} else if (response.data_source === 'cmsys') {
  console.log(response.not_in_datasource_count);
}
```

## 完整测试流程示例

### 场景1：使用ES数据源的完整流程

```bash
# 1. 创建ES同步任务
TASK_RESPONSE=$(curl -X POST 'http://localhost:8888/api/cmdb/v1/external-sync-tasks' \
  -H 'Content-Type: application/json' \
  -H 'Authorization: Bearer YOUR_TOKEN' \
  -d '{
    "task_name": "ES测试同步任务",
    "description": "用于测试的ES同步任务",
    "data_source": "elasticsearch",
    "cron_expression": "0 2 * * *",
    "query_time_range": "7d"
  }')

echo "创建任务响应: $TASK_RESPONSE"
TASK_ID=$(echo $TASK_RESPONSE | jq -r '.task_id')

# 2. 查看任务列表
curl 'http://localhost:8888/api/cmdb/v1/external-sync-tasks?data_source=elasticsearch' \
  -H 'Authorization: Bearer YOUR_TOKEN'

# 3. 手动执行同步（指定IP）
curl -X POST 'http://localhost:8888/api/cmdb/v1/external-sync-execute' \
  -H 'Content-Type: application/json' \
  -H 'Authorization: Bearer YOUR_TOKEN' \
  -d '{
    "data_source": "elasticsearch",
    "task_name": "手动测试同步",
    "host_ip_list": ["192.168.1.1", "192.168.1.2"],
    "query_time_range": "7d"
  }'

# 4. 查看执行日志
curl 'http://localhost:8888/api/cmdb/v1/external-sync-execution-logs?data_source=elasticsearch' \
  -H 'Authorization: Bearer YOUR_TOKEN'

# 5. 查看执行详情（假设execution_id=456）
curl 'http://localhost:8888/api/cmdb/v1/external-sync-execution-detail/456' \
  -H 'Authorization: Bearer YOUR_TOKEN'

# 6. 禁用任务
curl -X PUT 'http://localhost:8888/api/cmdb/v1/external-sync-tasks/enable' \
  -H 'Content-Type: application/json' \
  -H 'Authorization: Bearer YOUR_TOKEN' \
  -d '{
    "id": '$TASK_ID',
    "is_enabled": false
  }'

# 7. 删除任务
curl -X DELETE "http://localhost:8888/api/cmdb/v1/external-sync-tasks/$TASK_ID" \
  -H 'Authorization: Bearer YOUR_TOKEN'
```

### 场景2：使用CMSys数据源的完整流程

```bash
# 1. 创建CMSys同步任务
curl -X POST 'http://localhost:8888/api/cmdb/v1/external-sync-tasks' \
  -H 'Content-Type: application/json' \
  -H 'Authorization: Bearer YOUR_TOKEN' \
  -d '{
    "task_name": "CMSys测试同步任务",
    "description": "用于测试的CMSys同步任务",
    "data_source": "cmsys",
    "cron_expression": "0 3 * * *",
    "query_time_range": "30d"
  }'

# 2. 手动执行同步（全量）
curl -X POST 'http://localhost:8888/api/cmdb/v1/external-sync-full-sync' \
  -H 'Content-Type: application/json' \
  -H 'Authorization: Bearer YOUR_TOKEN' \
  -d '{
    "data_source": "cmsys",
    "task_name": "CMSys全量同步测试",
    "query": "department=DB"
  }'

# 3. 查看CMSys数据源的执行日志
curl 'http://localhost:8888/api/cmdb/v1/external-sync-execution-logs?data_source=cmsys' \
  -H 'Authorization: Bearer YOUR_TOKEN'
```

## 与原有接口的对比

### ES专用接口 vs 统一接口

| 功能 | 原ES专用接口 | 统一接口 |
|------|-------------|----------|
| 创建任务 | POST /api/cmdb/v1/es-sync-tasks | POST /api/cmdb/v1/external-sync-tasks (data_source=es) |
| 按IP执行 | POST /api/cmdb/v1/es-sync-execute | POST /api/cmdb/v1/external-sync-execute (data_source=es) |
| 全量同步 | POST /api/cmdb/v1/es-sync-full-sync | POST /api/cmdb/v1/external-sync-full-sync (data_source=es) |
| 执行日志 | GET /api/cmdb/v1/es-sync-execution-logs | GET /api/cmdb/v1/external-sync-execution-logs (data_source=es) |

### CMSys专用接口 vs 统一接口

| 功能 | 原CMSys专用接口 | 统一接口 | 说明 |
|------|----------------|----------|------|
| 按IP执行 | POST /api/cmdb/v1/cmsys-sync | POST /api/cmdb/v1/external-sync-execute (data_source=cmsys) | 统一接口支持 |
| 文件同步 | ❌ 不支持 | POST /api/cmdb/v1/external-sync-execute-file (data_source=cmsys) | ✅ 新增支持 |
| 全量同步 | ❌ 不支持 | POST /api/cmdb/v1/external-sync-full-sync (data_source=cmsys) | ✅ 新增支持 |

**重要更新**：
- CMSys现在支持文件同步和全量同步功能
- 所有同步操作的响应结构完全统一
- 前端可以使用相同的代码处理不同数据源的响应

### 兼容性说明

1. **原有接口保持不变**：所有原有的 ES 和 CMSys 专用接口继续可用
2. **推荐使用统一接口**：新开发的功能建议使用统一接口
3. **前端适配**：前端可以逐步迁移到统一接口，降低维护成本

## 错误码说明

| 错误码 | 说明 | 解决方案 |
|--------|------|----------|
| 400 | 请求参数错误 | 检查请求参数格式和必填字段 |
| 401 | 未认证 | 提供有效的认证Token |
| 404 | 任务不存在 | 检查task_id是否正确 |
| 500 | 服务器内部错误 | 查看服务器日志，联系管理员 |

### 常见错误示例

**错误：data_source参数无效**
```json
{
  "success": false,
  "message": "无效的数据源类型，必须是 'elasticsearch'、'es' 或 'cmsys'"
}
```

**错误：ES数据源缺少必填参数**
```json
{
  "success": false,
  "message": "使用ES数据源时，es_endpoint 和 es_index_pattern 不能为空"
}
```

**错误：任务不存在**
```json
{
  "success": false,
  "message": "任务不存在: task_id=999"
}
```

## 前端适配注意事项

### 1. 数据源选择

前端需要提供数据源选择控件：

```javascript
const dataSourceOptions = [
  { label: 'Elasticsearch', value: 'elasticsearch' },
  { label: 'CMSys', value: 'cmsys' }
];
```

### 2. 动态表单字段

根据选择的数据源，动态显示/隐藏相关字段：

```javascript
// ES数据源：显示 es_endpoint, es_index_pattern, group_name
// CMSys数据源：显示 query 参数
if (dataSource === 'elasticsearch' || dataSource === 'es') {
  // 显示ES相关字段
  showField('es_endpoint');
  showField('es_index_pattern');
  hideField('query');
} else if (dataSource === 'cmsys') {
  // 显示CMSys相关字段
  hideField('es_endpoint');
  hideField('es_index_pattern');
  showField('query');
}
```

### 3. 响应数据处理

**推荐方式**：统一处理所有数据源的响应，无需条件判断

```javascript
function renderSyncResult(response) {
  // 基础统计信息（所有数据源都有）
  console.log(`总数: ${response.total_hosts}`);
  console.log(`成功: ${response.success_count}`);
  console.log(`失败: ${response.failed_count}`);

  // 直接检查值是否大于0，无需判断data_source
  if (response.not_in_pool_count > 0) {
    console.log(`不在池中: ${response.not_in_pool_count}`);
    console.log(`不在池中的IP: ${response.not_in_pool_ip_list.join(', ')}`);
  }

  if (response.not_in_datasource_count > 0) {
    console.log(`数据源中不存在: ${response.not_in_datasource_count}`);
    console.log(`不存在的IP: ${response.not_in_datasource_ip_list.join(', ')}`);
  }

  // 全量同步结果
  if (response.new_hosts_count > 0) {
    console.log(`新增主机: ${response.new_hosts_count}`);
  }
  if (response.updated_hosts_count > 0) {
    console.log(`更新主机: ${response.updated_hosts_count}`);
  }

  // 成功和失败的IP列表
  if (response.success_ip_list && response.success_ip_list.length > 0) {
    console.log(`成功的IP: ${response.success_ip_list.join(', ')}`);
  }
  if (response.failed_ip_list && response.failed_ip_list.length > 0) {
    console.log(`失败的IP: ${response.failed_ip_list.join(', ')}`);
  }
}
```

**❌ 避免的做法**：不要根据data_source进行条件判断

```javascript
// 不需要这样做！所有字段在所有响应中都存在
function renderSyncResult_BAD(response) {
  if (response.data_source === 'elasticsearch') {
    // 不需要这种判断
    if (response.not_in_pool_count > 0) {
      console.log(`不在池中: ${response.not_in_pool_count}`);
    }
  } else if (response.data_source === 'cmsys') {
    // 不需要这种判断
    if (response.not_in_datasource_count > 0) {
      console.log(`数据源中不存在: ${response.not_in_datasource_count}`);
    }
  }
}
```

### 4. 统一的请求封装

```javascript
// 统一的外部资源同步API封装
class ExternalSyncAPI {
  static async createTask(params) {
    return await http.post('/api/cmdb/v1/external-sync-tasks', params);
  }

  static async executeSync(params) {
    return await http.post('/api/cmdb/v1/external-sync-execute', params);
  }

  static async getExecutionLogs(filters = {}) {
    return await http.get('/api/cmdb/v1/external-sync-execution-logs', { params: filters });
  }
}

// 使用示例
const esTask = await ExternalSyncAPI.createTask({
  task_name: 'ES同步',
  data_source: 'elasticsearch',
  cron_expression: '0 2 * * *'
});

const cmsysTask = await ExternalSyncAPI.createTask({
  task_name: 'CMSys同步',
  data_source: 'cmsys',
  cron_expression: '0 3 * * *'
});
```

## 迁移指南

### 从原有接口迁移到统一接口

#### ES接口迁移

**原来（ES专用接口）**：
```javascript
// 创建ES任务
await http.post('/api/cmdb/v1/es-sync-tasks', {
  task_name: 'ES同步',
  es_endpoint: 'http://es.example.com',
  cron_expression: '0 2 * * *'
});

// 执行ES同步
await http.post('/api/cmdb/v1/es-sync-execute', {
  host_ip_list: ['192.168.1.1'],
  query_time_range: '7d'
});
```

**现在（统一接口）**：
```javascript
// 创建任务
await http.post('/api/cmdb/v1/external-sync-tasks', {
  task_name: 'ES同步',
  data_source: 'elasticsearch',  // 新增data_source参数
  es_endpoint: 'http://es.example.com',
  cron_expression: '0 2 * * *'
});

// 执行同步
await http.post('/api/cmdb/v1/external-sync-execute', {
  data_source: 'elasticsearch',  // 新增data_source参数
  host_ip_list: ['192.168.1.1'],
  query_time_range: '7d'
});
```

#### CMSys接口迁移

**原来（CMSys专用接口）**：
```javascript
await http.post('/api/cmdb/v1/cmsys-sync', {
  task_name: 'CMSys同步',
  query: 'department=DB'
});
```

**现在（统一接口）**：
```javascript
await http.post('/api/cmdb/v1/external-sync-execute', {
  data_source: 'cmsys',  // 指定数据源
  task_name: 'CMSys同步',
  host_ip_list: [],  // 全量同步时可为空
  query: 'department=DB'
});
```

### 渐进式迁移策略

1. **阶段1**：保留原有接口，新功能使用统一接口
2. **阶段2**：前端逐步迁移到统一接口
3. **阶段3**：弃用警告（在响应中添加 deprecated 标记）
4. **阶段4**：移除原有接口（可选，根据实际情况决定）

## 性能优化建议

1. **分批同步**：大量IP同步时，建议分批执行，每批不超过100个IP
2. **时间范围**：根据需要调整 query_time_range，避免查询过大时间范围
3. **并发控制**：系统内部已实现并发控制（最大10个并发），无需额外处理
4. **执行时间**：建议将定时任务安排在业务低峰期（如凌晨2-4点）

## 监控和告警

建议配置以下监控指标：

1. **任务执行成功率**：`success_count / total_hosts`
2. **任务执行耗时**：`duration_ms`
3. **失败IP数量**：`failed_count`
4. **数据源可用性**：监控 ES 和 CMSys 接口的响应时间

## 附录

### A. Cron 表达式示例

| 表达式 | 说明 |
|--------|------|
| `0 2 * * *` | 每天凌晨2点 |
| `0 */4 * * *` | 每4小时执行一次 |
| `0 0 * * 0` | 每周日凌晨0点 |
| `0 0 1 * *` | 每月1号凌晨0点 |

### B. 数据源配置示例

RPC服务配置（`rpc/etc/cmpool.yaml`）：

```yaml
# ES数据源配置
ElasticSearchDataSource:
  Endpoint: "http://es.example.com:9200"
  IndexPattern: "metricbeat-*"
  TimeoutSeconds: 60

# CMSys数据源配置
CMSysDataSource:
  AuthEndpoint: "http://cmsys.example.com/auth"
  DataEndpoint: "http://cmsys.example.com/data"
  AppCode: "DB"
  AppSecret: "your-secret"
  Operator: "admin"
  TimeoutSeconds: 60
```

### C. 更新日志

**v1.1.0 (2025-01-22)**
- ✅ **重大更新**：统一所有数据源的响应结构
- ✅ 所有同步接口返回完全相同的字段集合
- ✅ 新增 CMSys 文件同步功能
- ✅ 新增 CMSys 全量同步功能
- ✅ 添加 `data_source` 字段标识数据源类型
- ✅ 全量同步接口新增 `new_hosts_count`、`updated_hosts_count` 等字段
- ✅ 前端无需根据 `data_source` 进行条件判断
- 📝 更新 API 文档，强调统一响应结构

**v1.0.0 (2025-01-21)**
- 首次发布统一外部资源同步API
- 支持 Elasticsearch 和 CMSys 两种数据源
- 提供完整的任务管理和执行接口
- 保持与原有专用接口的向后兼容
