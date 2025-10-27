# Mock ES Group 查询功能实现文档

## 概述

为支持全量同步功能，扩展了 Mock ES 接口，使其能够模拟按 group 查询多台主机的场景。此功能完全兼容原有的单主机查询，且符合真实 ES 聚合查询的响应格式。

## 修改内容

### 文件位置
`cmdb_backend_v2/api/internal/handler/mockesqueryhandler.go`

### 主要变更

#### 1. 增强参数提取逻辑
- 新增 `groupName` 参数提取
- 支持 `term.group.keyword` 和 `term.group` 两种格式
- 保持对原有 `hostIp` 参数的支持

```go
// 提取 group.keyword
if group, ok := term["group.keyword"].(string); ok {
    groupName = group
}
// 也尝试提取 group（兼容不带.keyword的情况）
if groupName == "" {
    if group, ok := term["group"].(string); ok {
        groupName = group
    }
}
```

#### 2. 查询类型判断
根据是否存在 `groupName` 参数，自动选择合适的响应生成函数：
- 有 `groupName`：调用 `generateGroupMockESResponse()` 生成多主机聚合响应
- 无 `groupName`：调用 `generateSingleHostMockESResponse()` 生成单主机响应

#### 3. 新增函数

##### `generateGroupMockESResponse(groupName string, hostCount int)`
生成包含多台主机的 ES 聚合查询响应。

**功能特性**：
- 生成指定数量的主机数据（默认100台）
- IP 范围：`10.0.1.1` ~ `10.0.1.100`
- 主机名格式：`db-server-001` ~ `db-server-100`
- 每台主机包含完整的监控统计数据

**数据范围**：
- CPU：15-35% (min), 60-95% (max)
- 内存：30-50 GB (min), 60-95 GB (max)
- 磁盘：500-800 GB (min), 900-2000 GB (max)
- 数据点数：7000-8640（模拟30天监控数据）

**响应格式**：
```json
{
  "took": 25,
  "timed_out": false,
  "_shards": {...},
  "hits": {
    "total": {"value": 860000, "relation": "eq"},
    "max_score": null,
    "hits": []
  },
  "aggregations": {
    "hosts": {
      "doc_count_error_upper_bound": 0,
      "sum_other_doc_count": 0,
      "buckets": [
        {
          "key": "10.0.1.1",
          "doc_count": 8234,
          "hostname": {
            "buckets": [{"key": "db-server-001", "doc_count": 8234}]
          },
          "cpu_stats": {
            "count": 8234,
            "min": 18.5,
            "max": 82.3,
            "avg": 62.1,
            "sum": 511394.14
          },
          "memory_stats": {...},
          "disk_stats": {...}
        },
        // ... 其他 99 台主机
      ]
    }
  }
}
```

##### `generateSingleHostMockESResponse(hostIP string)`
原 `generateMockESResponse` 函数重命名而来，功能保持不变。

## 测试

### 测试脚本
提供了完整的测试脚本：`test_mock_es_group.sh`

### 运行测试

确保 API 服务在运行中：
```bash
cd cmdb_backend_v2/api
go run cmdb.go -f etc/cmdb-api.yaml
```

在另一个终端运行测试：
```bash
cd cmdb_backend_v2
./test_mock_es_group.sh
```

### 测试内容
1. **Group 查询测试**：验证返回 100 台主机
2. **数据结构验证**：检查第一台主机的完整数据
3. **边界验证**：检查最后一台主机（第100台）
4. **兼容性测试**：验证单主机查询仍然正常工作

### 手动测试命令

#### 测试 Group 查询
```bash
curl -X POST "http://localhost:8888/platform/query/es" \
  -H "Content-Type: application/json" \
  -d '{
    "index": "cluster*:data-zabbix-host-monitor-*",
    "query": {
      "bool": {
        "must": [
          {"term": {"group.keyword": "DB组"}},
          {"range": {"@timestamp": {"gte": "now-30d", "lte": "now"}}}
        ]
      }
    },
    "aggs": {
      "hosts": {
        "terms": {"field": "hostIp.keyword", "size": 10000},
        "aggs": {
          "hostname": {"terms": {"field": "hostName.keyword", "size": 1}},
          "cpu_stats": {"stats": {"field": "cpu"}},
          "memory_stats": {"stats": {"field": "available_memory"}},
          "disk_stats": {"stats": {"field": "total_disk_space_all"}}
        }
      }
    },
    "size": 0
  }' | jq '.aggregations.hosts.buckets | length'
```

预期输出：`100`

#### 查看第一台主机详细信息
```bash
curl -X POST "http://localhost:8888/platform/query/es" \
  -H "Content-Type: application/json" \
  -d '{...}' | jq '.aggregations.hosts.buckets[0]'
```

#### 测试单主机查询（验证兼容性）
```bash
curl -X POST "http://localhost:8888/platform/query/es" \
  -H "Content-Type: application/json" \
  -d '{
    "index": "cluster*:data-zabbix-host-monitor-*",
    "query": {
      "bool": {
        "must": [
          {"term": {"hostIp": "192.168.1.100"}},
          {"range": {"@timestamp": {"gte": "now-30d", "lte": "now"}}}
        ]
      }
    },
    "aggs": {
      "cpu_stats": {"stats": {"field": "cpu"}},
      "memory_stats": {"stats": {"field": "available_memory"}},
      "disk_stats": {"stats": {"field": "total_disk_space_all"}},
      "hostname": {"terms": {"field": "hostName.keyword", "size": 1}}
    },
    "size": 0
  }' | jq '.aggregations'
```

## 日志输出

### Group 查询日志
```
Mock ES group查询 - Index: cluster*:data-zabbix-host-monitor-*, Group: DB组
Mock ES group查询 - Group: DB组, 生成 100 台主机数据
生成Mock ES group响应 - Group: DB组, 主机数: 100, 总数据点数: 812345
```

### 单主机查询日志
```
Mock ES单主机查询 - Index: cluster*:data-zabbix-host-monitor-*, HostIP: 192.168.1.100
生成Mock ES单主机响应 - HostIP: 192.168.1.100, DataPoints: 8234, MaxCPU: 75.23, MaxMemory: 82.45, MaxDisk: 945.12
```

## 与真实 ES 对比

### 查询条件
| 项目 | Mock ES | 真实 ES |
|------|---------|---------|
| Group 查询格式 | `term.group.keyword` | ✓ 一致 |
| 聚合字段 | `hostIp.keyword` | ✓ 一致 |
| 子聚合 | hostname, cpu_stats, memory_stats, disk_stats | ✓ 一致 |

### 响应格式
| 项目 | Mock ES | 真实 ES |
|------|---------|---------|
| 顶层结构 | took, timed_out, _shards, hits, aggregations | ✓ 一致 |
| Buckets 结构 | key, doc_count, hostname, *_stats | ✓ 一致 |
| Stats 字段 | count, min, max, avg, sum | ✓ 一致 |

### 数据特征
- Mock 数据使用随机值，但范围合理且符合生产环境特征
- 数据点数量模拟 30 天监控数据（每 5 分钟一个点）
- 确保 avg 值始终介于 min 和 max 之间

## 使用场景

### 1. 全量同步开发
开发 ES 全量同步功能时，不需要真实的 ES 环境，直接使用 Mock 接口即可获得大量测试数据。

### 2. 性能测试
使用 Mock 接口可以快速生成 100+ 台主机数据，用于测试前端和后端的性能表现。

### 3. 集成测试
在 CI/CD 流程中，可以使用 Mock 接口代替真实 ES，简化测试环境配置。

## 注意事项

1. **数据一致性**：每次请求生成的数据都是随机的，不适合需要数据一致性的测试场景
2. **并发安全**：使用 `rand.Seed(time.Now().UnixNano())` 在高并发下可能生成相同数据
3. **仅用于开发**：Mock 接口仅用于开发和测试环境，生产环境必须使用真实 ES
4. **IP 范围限制**：当前固定为 `10.0.1.1-100`，如需其他范围需修改代码

## 后续优化建议

1. **参数化主机数量**：从请求参数中读取需要生成的主机数量
2. **可配置 IP 范围**：通过配置文件设置 IP 段
3. **数据缓存**：对于相同的查询参数，返回缓存的数据以提高一致性
4. **更多 Group 支持**：支持不同 group 返回不同特征的数据
5. **时间范围感知**：根据查询的时间范围动态调整数据点数量

## 相关文件

- 实现文件：`cmdb_backend_v2/api/internal/handler/mockesqueryhandler.go`
- 测试脚本：`cmdb_backend_v2/test_mock_es_group.sh`
- 真实 ES 客户端参考：`cmdb_backend_v2/rpc/internal/datasource/elasticsearch/esclient.go`

## 参考资料

- Elasticsearch Aggregation API 文档
- go-zero 框架官方文档
- 项目架构文档：`CLAUDE.md`
