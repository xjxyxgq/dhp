# Mock ES接口使用指南

## 📖 简介

为了方便开发和测试ES数据同步功能，我们在API服务中创建了一个Mock ES接口。该接口模拟真实ES的查询响应，返回符合格式的监控数据，让开发过程不再依赖真实的ES环境。

## 🎯 功能特点

- ✅ 完全模拟ES查询响应格式
- ✅ 返回随机但合理的监控数据（CPU、内存、磁盘）
- ✅ 支持ES查询请求参数解析
- ✅ 无需外部依赖，启动即可用
- ✅ 便于调试和测试

## 🚀 快速开始

### 1. 启动服务

Mock ES接口已集成在API服务中，启动API服务即可使用：

```bash
cd cmdb_backend_v2/api
./cmdb-api -f etc/cmdb-api.yaml
```

### 2. Mock ES接口地址

```
http://localhost:8888/platform/query/es
```

**注意**: Mock接口路径 `/platform/query/es` 与真实ES路径一致，便于切换。

### 3. 配置切换

#### 开发/测试环境配置（使用Mock ES）

修改 `rpc/etc/cmpool.yaml`：

```yaml
ESDataSource:
  # 使用Mock ES - 指向本地API服务
  DefaultEndpoint: "http://localhost:8888/platform/query/es"
  DefaultIndexPattern: "cluster*:data-zabbix-host-monitor-*"
  TimeoutSeconds: 30
```

#### 生产环境配置（使用真实ES）

修改 `rpc/etc/cmpool.yaml`：

```yaml
ESDataSource:
  # 使用真实ES
  DefaultEndpoint: "http://phoenix.local.com/platform/query/es"
  DefaultIndexPattern: "cluster*:data-zabbix-host-monitor-*"
  TimeoutSeconds: 30
```

## 📊 Mock数据说明

### 返回的监控数据范围

Mock接口会生成以下范围的随机数据：

| 指标 | 最小值 | 最大值 | 说明 |
|------|--------|--------|------|
| CPU使用率 | 60% | 90% | 模拟中高负载 |
| 内存使用量 | 50GB | 90GB | 模拟典型服务器内存使用 |
| 磁盘空间 | 800GB | 1000GB | 模拟常见磁盘容量 |
| 数据点数量 | 7000 | 8640 | 模拟30天数据（每5分钟一个点）|

### 数据特点

- **合理性**: 平均值始终小于最大值（60-80%的关系）
- **一致性**: 同一次查询返回相同主机的数据保持一致
- **随机性**: 每次查询生成不同的数据，模拟真实环境变化

## 🔧 使用示例

### 示例1: 测试单个主机同步

```bash
curl -X POST http://localhost:8888/api/cmdb/v1/es-sync-execute \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer <token>" \
  -d '{
    "task_name": "测试同步",
    "host_ip_list": ["192.168.1.100", "192.168.1.101"],
    "query_time_range": "30d"
  }'
```

### 示例2: 测试定时任务

1. 创建任务：
```bash
curl -X POST http://localhost:8888/api/cmdb/v1/es-sync-tasks \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer <token>" \
  -d '{
    "task_name": "测试定时任务",
    "cron_expression": "0 */5 * * * ?",
    "query_time_range": "1h"
  }'
```

2. 启用任务：
```bash
curl -X PUT http://localhost:8888/api/cmdb/v1/es-sync-tasks/enable \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer <token>" \
  -d '{"id": 1, "is_enabled": true}'
```

### 示例3: 直接测试Mock ES接口

```bash
curl -X POST http://localhost:8888/platform/query/es \
  -H "Content-Type: application/json" \
  -d '{
    "index": "cluster*:data-zabbix-host-monitor-*",
    "query": {
      "bool": {
        "must": [
          {
            "term": {
              "hostIp": "192.168.1.100"
            }
          }
        ]
      }
    },
    "aggs": {
      "cpu_stats": {"stats": {"field": "cpu"}},
      "memory_stats": {"stats": {"field": "available_memory"}},
      "disk_stats": {"stats": {"field": "total_disk_space_all"}}
    }
  }'
```

## 📋 Mock ES响应格式

Mock接口返回标准的ES查询响应格式：

```json
{
  "took": 15,
  "timed_out": false,
  "_shards": {
    "total": 5,
    "successful": 5,
    "skipped": 0,
    "failed": 0
  },
  "hits": {
    "total": {
      "value": 8640,
      "relation": "eq"
    },
    "max_score": null,
    "hits": []
  },
  "aggregations": {
    "cpu_stats": {
      "count": 8640,
      "min": 10.5,
      "max": 85.23,
      "avg": 68.45,
      "sum": 591168.0
    },
    "memory_stats": {
      "count": 8640,
      "min": 20.0,
      "max": 78.56,
      "avg": 62.34,
      "sum": 538617.6
    },
    "disk_stats": {
      "count": 8640,
      "min": 500.0,
      "max": 950.12,
      "avg": 855.10,
      "sum": 7388064.0
    }
  }
}
```

## 🐛 调试技巧

### 1. 查看Mock ES日志

Mock接口会输出详细日志，便于调试：

```
收到Mock ES查询请求
Mock ES查询 - Index: cluster*:data-zabbix-host-monitor-*, HostIP: 192.168.1.100
生成Mock ES响应 - HostIP: 192.168.1.100, DataPoints: 8234, MaxCPU: 82.45, MaxMemory: 75.23, MaxDisk: 923.45
```

### 2. 验证同步结果

查询执行记录：
```bash
curl -X GET "http://localhost:8888/api/cmdb/v1/es-sync-execution-logs?limit=10" \
  -H "Authorization: Bearer <token>"
```

查询执行详情：
```bash
curl -X GET "http://localhost:8888/api/cmdb/v1/es-sync-execution-detail/1" \
  -H "Authorization: Bearer <token>"
```

### 3. 验证数据写入

检查 `server_resources` 表：
```sql
SELECT * FROM server_resources
WHERE ip IN ('192.168.1.100', '192.168.1.101')
ORDER BY date_time DESC
LIMIT 10;
```

## ⚠️ 注意事项

### 1. 仅用于开发测试

Mock ES接口仅供开发和测试使用，**不要在生产环境使用**！

### 2. 数据不持久

Mock接口每次请求都生成随机数据，不保存历史数据。

### 3. 性能考虑

Mock接口不会有真实ES的性能特征，不适合性能测试。

### 4. 功能限制

Mock接口仅模拟了最基本的ES聚合查询功能，不支持：
- 复杂的ES查询语法
- ES的所有高级特性
- 真实的分片和副本机制

## 🔄 环境切换流程

### 从Mock ES切换到真实ES

1. 停止RPC服务
2. 修改 `rpc/etc/cmpool.yaml` 中的 `ESDataSource.DefaultEndpoint`
3. 重启RPC服务
4. 验证连接：创建测试任务并执行

### 从真实ES切换到Mock ES

1. 停止RPC服务
2. 确保API服务正在运行（Mock ES需要）
3. 修改 `rpc/etc/cmpool.yaml` 中的 `ESDataSource.DefaultEndpoint` 为 `http://localhost:8888/platform/query/es`
4. 重启RPC服务
5. 验证Mock ES：检查日志中是否有Mock ES相关输出

## 📁 相关文件

- `api/internal/handler/mockesqueryhandler.go` - Mock ES实现
- `api/internal/handler/routes.go` - 路由注册（第333-342行）
- `rpc/etc/cmpool.yaml` - ES配置文件
- `rpc/internal/datasource/elasticsearch/esclient.go` - ES客户端

## 🎓 开发建议

### 推荐的开发流程

1. **本地开发**: 使用Mock ES进行功能开发
2. **联调测试**: 使用Mock ES进行接口联调
3. **集成测试**: 切换到测试环境的真实ES
4. **生产部署**: 使用生产环境的真实ES

### 测试用例建议

使用Mock ES时，建议测试：
- ✅ 单个主机同步
- ✅ 多个主机并发同步
- ✅ 文件上传同步
- ✅ 定时任务创建和执行
- ✅ 执行记录查询
- ✅ 错误处理（如主机不在pool中）

## 🆘 常见问题

### Q1: Mock ES接口无响应？

**A**: 检查API服务是否正在运行：
```bash
curl http://localhost:8888/api/auth/cas
```

### Q2: 同步失败，提示连接错误？

**A**: 确认配置文件中的 `DefaultEndpoint` 地址正确：
```bash
grep -A3 "ESDataSource" rpc/etc/cmpool.yaml
```

### Q3: 想要自定义Mock数据？

**A**: 修改 `mockesqueryhandler.go` 中的 `generateMockESResponse` 函数：
```go
maxCPU := 60.0 + rand.Float64()*30.0  // 调整这里的范围
```

### Q4: 如何验证正在使用Mock ES？

**A**: 查看RPC服务日志，如果看到API服务日志输出 "收到Mock ES查询请求"，说明正在使用Mock ES。

## 📞 支持

如有问题或建议，请查看：
- `ES_SYNC_API_DOCUMENTATION.md` - 完整API文档
- `ES_SYNC_IMPLEMENTATION_GUIDE.md` - 实现指南

---

*最后更新: 2025-10-13*
*版本: v1.0*
