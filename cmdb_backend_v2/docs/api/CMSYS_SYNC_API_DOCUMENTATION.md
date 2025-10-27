# CMSys HTTP 接口数据同步 API 文档

## 功能概述

CMSys HTTP 接口数据同步功能允许从外部 HTTP 接口获取主机监控数据，并自动同步到 CMDB 系统的 `hosts_pool` 和 `server_resources` 表中。

与 ES 数据同步不同，CMSys 同步：
- 使用 HTTPS 协议访问外部 API 接口
- 需要进行认证（获取 token）
- 支持 remark（备注）字段的同步
- 自动插入新主机到 `hosts_pool` 表

## 数据流程

```
[CMSys 认证接口] → [获取 Token]
         ↓
[CMSys 数据接口] → [获取主机数据] → [解析 JSON]
         ↓
[检查 hosts_pool] → [插入/更新主机] → [更新 remark]
         ↓
[更新 server_resources] → [记录执行日志]
```

## API 接口

### 1. CMSys 数据同步接口

从 CMSys HTTP 接口同步主机监控数据。

**接口地址**
```
POST /api/cmdb/v1/cmsys-sync
```

**请求参数**

| 参数名 | 类型 | 必填 | 说明 |
|--------|------|------|------|
| query | string | 否 | 查询参数（通过 URL query 参数传递给 CMSys 接口） |
| task_name | string | 否 | 任务名称（用于执行日志记录） |

**请求示例**

```bash
curl -X POST 'http://localhost:8888/api/cmdb/v1/cmsys-sync' \
  -H 'Content-Type: application/json' \
  -d '{
    "query": "department=DB&status=active",
    "task_name": "手动同步DB部门主机数据"
  }'
```

**响应参数**

| 参数名 | 类型 | 说明 |
|--------|------|------|
| success | boolean | 同步是否成功 |
| message | string | 同步结果消息 |
| execution_id | int64 | 执行记录ID（可用于查询详细执行日志） |
| total_hosts | int32 | 总主机数 |
| success_count | int32 | 同步成功的主机数 |
| failed_count | int32 | 同步失败的主机数 |
| not_in_datasource_count | int32 | 数据源中不存在的主机数 |
| success_ip_list | array | 同步成功的主机IP列表 |
| failed_ip_list | array | 同步失败的主机IP列表 |
| not_in_datasource_ip_list | array | 数据源中不存在的主机IP列表 |

**响应示例**

```json
{
  "success": true,
  "message": "同步完成: 成功15个, 失败2个, 数据源中不存在0个",
  "execution_id": 123,
  "total_hosts": 17,
  "success_count": 15,
  "failed_count": 2,
  "not_in_datasource_count": 0,
  "success_ip_list": [
    "192.168.1.1",
    "192.168.1.2",
    ...
  ],
  "failed_ip_list": [
    "192.168.1.100",
    "192.168.1.101"
  ],
  "not_in_datasource_ip_list": []
}
```

## CMSys 数据源配置

### 配置文件

在 `rpc/etc/cmpool.yaml` 中配置 CMSys 数据源信息：

```yaml
CMSysDataSource:
  AuthEndpoint: "https://api.cmsys.example.com/auth"     # 认证接口地址
  DataEndpoint: "https://api.cmsys.example.com/data"     # 数据接口地址
  AppCode: "DB"                                          # 应用代码
  AppSecret: "your-app-secret-here"                      # 应用密钥
  Operator: "admin"                                      # 操作员标识
  TimeoutSeconds: 60                                     # 请求超时时间(秒)
```

### 配置说明

| 配置项 | 说明 |
|--------|------|
| AuthEndpoint | 认证接口地址，用于获取 access token |
| DataEndpoint | 数据接口地址，用于获取主机监控数据 |
| AppCode | 应用代码，认证时使用 |
| AppSecret | 应用密钥，认证时使用 |
| Operator | 操作员标识，数据请求时在 header 中传递 |
| TimeoutSeconds | HTTP 请求超时时间 |

## CMSys 接口规范

### 认证接口

**请求方式**: POST
**请求体**:
```json
{
  "appCode": "DB",
  "secret": "your-app-secret"
}
```

**响应格式**:
```json
{
  "code": "A0000",
  "msg": "success",
  "data": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."
}
```

### 数据接口

**请求方式**: GET
**请求头**:
```
x-control-access-token: eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...
x-control-access-operator: admin
```

**响应格式**:
```json
{
  "code": "A000",
  "msg": "success",
  "data": [
    {
      "ipAddress": "192.168.1.1",
      "cpuMaxNew": "85.5",
      "memMaxNew": "72.3",
      "diskMaxNew": "65.8",
      "remark": "生产环境数据库服务器"
    },
    {
      "ipAddress": "192.168.1.2",
      "cpuMaxNew": "45.2",
      "memMaxNew": "68.9",
      "diskMaxNew": "78.4",
      "remark": "测试环境应用服务器"
    }
  ]
}
```

## 数据映射

### CMSys 到 hosts_pool

| CMSys 字段 | hosts_pool 字段 | 类型 | 说明 |
|------------|----------------|------|------|
| ipAddress | host_ip | string | 主机 IP 地址 |
| - | host_name | string | 主机名（从 IP 推导或留空） |
| remark | remark | string | 备注信息 |

### CMSys 到 server_resources

| CMSys 字段 | server_resources 字段 | 类型 | 说明 |
|------------|----------------------|------|------|
| cpuMaxNew | cpu_load | float | CPU 最大利用率（%） |
| memMaxNew | used_memory | float | 内存最大利用率（%） |
| diskMaxNew | total_disk | float | 磁盘最大利用率（%） |

## 同步状态说明

| 状态 | 说明 |
|------|------|
| success | 同步成功 |
| failed | 同步失败（查询失败、数据库操作失败等） |
| not_in_datasource | 数据源中无此主机数据（CMSys 查询成功但返回空数据） |

## 执行日志

同步执行记录保存在以下表中：

### es_sync_execution_log

主执行日志表（与 ES 同步共用）

| 字段 | 说明 |
|------|------|
| id | 执行记录ID |
| task_id | 任务ID（手动执行为0） |
| task_name | 任务名称 |
| execution_time | 执行时间 |
| execution_status | 执行状态（success/partial/failed/running） |
| total_hosts | 总主机数 |
| success_count | 成功数量 |
| failed_count | 失败数量 |
| not_in_pool_count | 数据源中不存在数量 |
| duration_ms | 执行耗时（毫秒） |

### es_sync_execution_detail

执行详情表（与 ES 同步共用）

| 字段 | 说明 |
|------|------|
| id | 详情记录ID |
| execution_id | 执行记录ID |
| host_ip | 主机IP |
| host_name | 主机名 |
| sync_status | 同步状态（success/failed/not_in_datasource） |
| error_message | 错误信息 |
| max_cpu | 最大CPU利用率 |
| avg_cpu | 平均CPU利用率 |
| max_memory | 最大内存利用率 |
| avg_memory | 平均内存利用率 |
| max_disk | 最大磁盘利用率 |
| avg_disk | 平均磁盘利用率 |
| data_point_count | 数据点数量 |

## 使用示例

### 示例 1: 基本同步

同步所有主机数据：

```bash
curl -X POST 'http://localhost:8888/api/cmdb/v1/cmsys-sync' \
  -H 'Content-Type: application/json' \
  -d '{
    "task_name": "全量同步"
  }'
```

### 示例 2: 带查询条件的同步

同步特定条件的主机数据：

```bash
curl -X POST 'http://localhost:8888/api/cmdb/v1/cmsys-sync' \
  -H 'Content-Type: application/json' \
  -d '{
    "query": "department=DB&environment=production",
    "task_name": "DB生产环境同步"
  }'
```

## 注意事项

1. **Token 管理**
   - Token 自动缓存，默认1小时有效期
   - Token 过期后自动重新获取
   - 无需手动管理 token

2. **并发控制**
   - 系统自动控制并发数（最多10个）
   - 大批量同步时会自动分批处理

3. **数据覆盖**
   - 如果主机已存在，只更新 remark 字段
   - server_resources 表执行 UPSERT 操作（存在则更新，不存在则插入）

4. **错误处理**
   - 单个主机同步失败不影响其他主机
   - 详细错误信息记录在执行详情表中

5. **Remark 字段**
   - Remark 信息会自动写入 hosts_pool 表
   - 可通过 `/api/v1/hardware-proxy/cmdb/v1/get_hosts_pool_detail` 接口查询

## 相关接口

### 查询主机池详情（包含 remark）

```
POST /api/v1/hardware-proxy/cmdb/v1/get_hosts_pool_detail
```

**请求示例**:
```json
{
  "ip_list": ["192.168.1.1", "192.168.1.2"]
}
```

**响应示例**:
```json
{
  "success": true,
  "message": "查询成功",
  "hosts_pool_detail": [
    {
      "id": 1,
      "hostname": "db-server-01",
      "host_ip": "192.168.1.1",
      "remark": "生产环境数据库服务器",
      ...
    }
  ]
}
```

## 与 ES 同步的对比

| 特性 | CMSys 同步 | ES 同步 |
|------|-----------|---------|
| 数据源 | HTTP API 接口 | ElasticSearch |
| 认证方式 | Token 认证 | 无需认证 |
| 数据格式 | JSON | ES 查询结果 |
| Remark 支持 | ✅ 支持 | ❌ 不支持 |
| 查询方式 | HTTP GET | ES Query DSL |
| 数据粒度 | 汇总数据 | 时序数据聚合 |

## 故障排查

### 常见问题

1. **认证失败**
   - 检查 AppCode 和 AppSecret 配置是否正确
   - 查看 RPC 服务日志中的认证错误信息

2. **数据同步失败**
   - 检查数据接口地址配置
   - 查看执行详情表中的错误信息
   - 确认网络连接和 SSL 证书

3. **主机未插入 hosts_pool**
   - 检查数据库连接
   - 查看 RPC 服务日志中的数据库操作错误

4. **Remark 未更新**
   - 确认 CMSys 接口返回了 remark 字段
   - 检查数据库 hosts_pool 表是否有 remark 列

## Mock 接口（开发测试用）

为了方便开发和测试，系统提供了 CMSys Mock 接口，无需连接真实的 CMSys 系统即可进行开发。

### Mock 接口地址

- **认证接口**: `http://localhost:8888/platform/cmsys/auth`
- **数据接口**: `http://localhost:8888/platform/cmsys/data`

### 使用 Mock 接口

修改 `rpc/etc/cmpool.yaml` 配置：

```yaml
CMSysDataSource:
  AuthEndpoint: "http://localhost:8888/platform/cmsys/auth"
  DataEndpoint: "http://localhost:8888/platform/cmsys/data"
  AppCode: "DB"
  AppSecret: "your-app-secret-here"
  Operator: "admin"
  TimeoutSeconds: 60
```

### 测试 Mock 接口

使用提供的测试脚本：

```bash
cd cmdb_backend_v2
./test_cmsys_mock.sh
```

### Mock 数据说明

- 自动生成 10 台虚拟主机数据（192.168.1.1 - 192.168.1.10）
- 随机生成资源利用率（CPU 40-90%, Memory 50-90%, Disk 30-90%）
- 包含真实场景的 remark 描述
- 支持 token 认证验证

详细的 Mock 接口使用说明，请参考：[CMSYS_MOCK_INTERFACES.md](./CMSYS_MOCK_INTERFACES.md)

## 更新日志

### v1.0.0 (2025-01-15)
- 首次发布
- 支持从 CMSys HTTP 接口同步主机数据
- 支持 remark 字段同步
- 自动插入新主机到 hosts_pool
- 复用 ES 同步的执行日志表
- 提供 Mock 接口用于开发测试
