# 资源使用率百分比字段添加实施指南

## 需求概述

在 ES 同步接口获取到的是资源使用率的百分比值，需要在后端数据库的 `server_resources` 表中添加对应的字段（cpu_percent、mem_percent、disk_percent），每个指标包含 max、avg、min 三个值，共计9个字段，用于存储这些百分比值。

同时修改4个查询接口，返回这些新的百分比指标：
- `/api/v1/hardware-proxy/cmdb/v1/cluster-resources-max`
- `/api/v1/hardware-proxy/cmdb/v1/cluster-resources`
- `/api/v1/hardware-proxy/cmdb/v1/server-resources-max`
- `/api/v1/hardware-proxy/cmdb/v1/server-resources`

## 已完成的修改

### 1. 数据库层修改

#### ✅ Schema 定义更新
**文件**: `source/schema.sql`

已在 `server_resources` 表中添加9个百分比字段：
```sql
`cpu_percent_max` double DEFAULT NULL COMMENT 'CPU使用率最大值(%)',
`cpu_percent_avg` double DEFAULT NULL COMMENT 'CPU使用率平均值(%)',
`cpu_percent_min` double DEFAULT NULL COMMENT 'CPU使用率最小值(%)',
`mem_percent_max` double DEFAULT NULL COMMENT '内存使用率最大值(%)',
`mem_percent_avg` double DEFAULT NULL COMMENT '内存使用率平均值(%)',
`mem_percent_min` double DEFAULT NULL COMMENT '内存使用率最小值(%)',
`disk_percent_max` double DEFAULT NULL COMMENT '磁盘使用率最大值(%)',
`disk_percent_avg` double DEFAULT NULL COMMENT '磁盘使用率平均值(%)',
`disk_percent_min` double DEFAULT NULL COMMENT '磁盘使用率最小值(%)'
```

#### ✅ 数据库表结构更新
已执行 ALTER TABLE 语句在实际数据库中添加这些字段。

### 2. Model 层修改

#### ✅ Model 文件重新生成
**文件**: `rpc/internal/model/serverresourcesmodel_gen.go`

已重新生成，`ServerResources` 结构体现在包含：
```go
CpuPercentMax  sql.NullFloat64 `db:"cpu_percent_max"`  // CPU使用率最大值(%)
CpuPercentAvg  sql.NullFloat64 `db:"cpu_percent_avg"`  // CPU使用率平均值(%)
CpuPercentMin  sql.NullFloat64 `db:"cpu_percent_min"`  // CPU使用率最小值(%)
MemPercentMax  sql.NullFloat64 `db:"mem_percent_max"`  // 内存使用率最大值(%)
MemPercentAvg  sql.NullFloat64 `db:"mem_percent_avg"`  // 内存使用率平均值(%)
MemPercentMin  sql.NullFloat64 `db:"mem_percent_min"`  // 内存使用率最小值(%)
DiskPercentMax sql.NullFloat64 `db:"disk_percent_max"` // 磁盘使用率最大值(%)
DiskPercentAvg sql.NullFloat64 `db:"disk_percent_avg"` // 磁盘使用率平均值(%)
DiskPercentMin sql.NullFloat64 `db:"disk_percent_min"` // 磁盘使用率最小值(%)
```

#### ✅ UpsertFromES 方法更新
**文件**: `rpc/internal/model/serverresourcesmodel.go`

已更新接口签名和实现：
```go
UpsertFromES(ctx context.Context, poolId uint64, ip string, cpuLoad, usedMemory, totalDisk float64,
    cpuPercentMax, cpuPercentAvg, cpuPercentMin,
    memPercentMax, memPercentAvg, memPercentMin,
    diskPercentMax, diskPercentAvg, diskPercentMin float64) error
```

INSERT 和 UPDATE 语句已包含所有新字段。

## 需要完成的修改

### 3. RPC Proto 文件修改

**文件**: `rpc/proto/cmpool.proto`

需要在以下消息类型中添加百分比字段：

#### 修改 ServerResourceRow 消息
```protobuf
message ServerResourceRow {
    // ... 现有字段 ...

    // 新增百分比字段
    double cpu_percent_max = 20;
    double cpu_percent_avg = 21;
    double cpu_percent_min = 22;
    double mem_percent_max = 23;
    double mem_percent_avg = 24;
    double mem_percent_min = 25;
    double disk_percent_max = 26;
    double disk_percent_avg = 27;
    double disk_percent_min = 28;
}
```

#### 修改 ServerResourceMaxRow 消息
```protobuf
message ServerResourceMaxRow {
    // ... 现有字段 ...

    // 新增百分比字段
    double cpu_percent_max = 20;
    double cpu_percent_avg = 21;
    double cpu_percent_min = 22;
    double mem_percent_max = 23;
    double mem_percent_avg = 24;
    double mem_percent_min = 25;
    double disk_percent_max = 26;
    double disk_percent_avg = 27;
    double disk_percent_min = 28;
}
```

#### 修改 ClusterMemberResourceRow 消息
```protobuf
message ClusterMemberResourceRow {
    // ... 现有字段 ...

    // 新增百分比字段
    double cpu_percent_max = 20;
    double cpu_percent_avg = 21;
    double cpu_percent_min = 22;
    double mem_percent_max = 23;
    double mem_percent_avg = 24;
    double mem_percent_min = 25;
    double disk_percent_max = 26;
    double disk_percent_avg = 27;
    double disk_percent_min = 28;
}
```

#### 修改 ClusterResourceMaxRow 消息
```protobuf
message ClusterResourceMaxRow {
    // ... 现有字段 ...

    // 新增百分比字段
    double cpu_percent_max = 20;
    double cpu_percent_avg = 21;
    double cpu_percent_min = 22;
    double mem_percent_max = 23;
    double mem_percent_avg = 24;
    double mem_percent_min = 25;
    double disk_percent_max = 26;
    double disk_percent_avg = 27;
    double disk_percent_min = 28;
}
```

**重新生成 RPC 代码**:
```bash
cd rpc
/Users/xuguoqiang/LocalOthers/goctl/goctl rpc protoc proto/cmpool.proto --go_out=. --go-grpc_out=. --zrpc_out=.

# 复制客户端文件到 API 模块
cp cmpool/cmpool.pb.go ../api/cmpool/
cp cmpool/cmpool_grpc.pb.go ../api/cmpool/
```

### 4. RPC Logic 层修改

#### 修改 ES 同步逻辑

**文件**: `rpc/internal/logic/executeessyncbyhostlistlogic.go`

在 `syncToServerResources` 方法中（第238-255行），需要传递百分比值：

```go
func (l *ExecuteEsSyncByHostListLogic) syncToServerResources(poolId uint64, hostIP string, metrics *elasticsearch.HostMetrics) error {
    return l.svcCtx.ServerResourcesModel.UpsertFromES(
        l.ctx,
        poolId,
        hostIP,
        metrics.MaxCPU,
        metrics.MaxMemory,
        metrics.MaxDisk,
        // 新增百分比参数
        metrics.MaxCPU,    // cpu_percent_max - ES返回的已经是百分比
        metrics.AvgCPU,    // cpu_percent_avg
        metrics.MinCPU,    // cpu_percent_min - 需要确认ES是否返回MinCPU
        metrics.MaxMemory, // mem_percent_max
        metrics.AvgMemory, // mem_percent_avg
        metrics.MinMemory, // mem_percent_min - 需要确认
        metrics.MaxDisk,   // disk_percent_max
        metrics.AvgDisk,   // disk_percent_avg
        metrics.MinDisk,   // disk_percent_min - 需要确认
    )
}
```

**注意**: 需要确认 `elasticsearch.HostMetrics` 结构体是否包含 Min 值，如果没有可以设置为0或与Avg值相同。

**同样需要修改的文件**:
- `rpc/internal/scheduler/es_sync_scheduler.go` (定时同步)
- `rpc/internal/logic/executecmsyssynclogic.go` (CMSys同步，如果也需要支持)

#### 修改查询接口 Logic

需要修改4个查询接口的Logic，将数据库查询结果中的百分比字段映射到Proto响应消息中。

##### (1) FindServerResourcesMax Logic

**文件**: `rpc/internal/logic/findserverresourcesmaxlogic.go`

在构建响应时添加百分比字段：
```go
// 在循环中构建每个 ServerResourceMaxRow 时
row := &cmpool.ServerResourceMaxRow{
    // ... 现有字段 ...

    // 新增百分比字段
    CpuPercentMax:  data.CpuPercentMax,
    CpuPercentAvg:  data.CpuPercentAvg,
    CpuPercentMin:  data.CpuPercentMin,
    MemPercentMax:  data.MemPercentMax,
    MemPercentAvg:  data.MemPercentAvg,
    MemPercentMin:  data.MemPercentMin,
    DiskPercentMax: data.DiskPercentMax,
    DiskPercentAvg: data.DiskPercentAvg,
    DiskPercentMin: data.DiskPercentMin,
}
```

**注意**: Model 层的查询方法也需要更新，在 SELECT 语句中添加这些字段：
- `FindServerResourceMax` 方法需要添加百分比字段到查询和结构体

##### (2) FindServerResourcesWith Filter Logic

**文件**: `rpc/internal/logic/findserverresourceswithfilterlogic.go`

同样需要在响应构建时添加百分比字段。

##### (3) FindClusterResourcesMax Logic

**文件**: `rpc/internal/logic/findclusterresourcesmaxlogic.go`

在响应中添加百分比字段。

##### (4) FindClusterResources Logic

**文件**: `rpc/internal/logic/findclusterresourceslogic.go`

在响应中添加百分比字段。

### 5. Model 层查询方法更新

需要更新 Model 层的查询结果结构体，添加百分比字段：

**文件**: `rpc/internal/model/serverresourcesmodel.go`

#### 更新 ServerResourceRow 结构体
```go
type ServerResourceRow struct {
    // ... 现有字段 ...

    // 新增百分比字段
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
```

#### 更新 ServerResourceMaxData 结构体
同样添加9个百分比字段。

#### 更新 ClusterMemberResourceData 结构体
同样添加9个百分比字段。

#### 更新 ClusterResourceMaxData 结构体
同样添加9个百分比字段。

#### 更新查询 SQL

在以下方法的 SELECT 语句中添加百分比字段：
- `FindServerResourceMax` - 添加 `sr.cpu_percent_max, sr.cpu_percent_avg, ...`
- `FindServerResourcesWithFilter` - 同上
- `FindClusterResources` - 同上
- `FindClusterResourcesMax` - 同上（可能需要计算聚合值）

### 6. API 层修改

#### 修改 API 定义

**文件**: `api/cmdb.api`

需要在相应的响应类型中添加百分比字段：

```go
type ServerResourceRow {
    // ... 现有字段 ...

    CpuPercentMax  float64 `json:"cpuPercentMax"`
    CpuPercentAvg  float64 `json:"cpuPercentAvg"`
    CpuPercentMin  float64 `json:"cpuPercentMin"`
    MemPercentMax  float64 `json:"memPercentMax"`
    MemPercentAvg  float64 `json:"memPercentAvg"`
    MemPercentMin  float64 `json:"memPercentMin"`
    DiskPercentMax float64 `json:"diskPercentMax"`
    DiskPercentAvg float64 `json:"diskPercentAvg"`
    DiskPercentMin float64 `json:"diskPercentMin"`
}

type ServerResourceMaxRow {
    // 同上添加9个字段
}

type ClusterMemberResourceRow {
    // 同上添加9个字段
}

type ClusterResourceMaxRow {
    // 同上添加9个字段
}
```

**重新生成 API 代码**:
```bash
cd api
/Users/xuguoqiang/LocalOthers/goctl/goctl api go -api cmdb.api -dir .
```

#### 修改 API Logic 层

在4个查询接口的 Logic 文件中，将 RPC 响应中的百分比字段传递给 API 响应：

##### (1) GetServerResourcesMaxLogic
**文件**: `api/internal/logic/getserverresourcesmaxlogic.go`

```go
// 在构建 API 响应时
apiRow := types.ServerResourceMaxRow{
    // ... 现有字段映射 ...

    CpuPercentMax:  rpcRow.CpuPercentMax,
    CpuPercentAvg:  rpcRow.CpuPercentAvg,
    CpuPercentMin:  rpcRow.CpuPercentMin,
    MemPercentMax:  rpcRow.MemPercentMax,
    MemPercentAvg:  rpcRow.MemPercentAvg,
    MemPercentMin:  rpcRow.MemPercentMin,
    DiskPercentMax: rpcRow.DiskPercentMax,
    DiskPercentAvg: rpcRow.DiskPercentAvg,
    DiskPercentMin: rpcRow.DiskPercentMin,
}
```

##### (2) GetServerResourcesLogic
**文件**: `api/internal/logic/getserverresourceslogic.go`

同样添加百分比字段映射。

##### (3) GetClusterResourcesMaxLogic
**文件**: `api/internal/logic/getclusterresourcesmaxlogic.go`

同样添加百分比字段映射。

##### (4) GetClusterResourcesLogic
**文件**: `api/internal/logic/getclusterresourceslogic.go`

同样添加百分比字段映射。

## 编译和验证

完成所有修改后，需要依次编译验证：

### 1. 编译 RPC 服务
```bash
cd rpc
go build -o cmdb-rpc .
```

### 2. 编译 API 服务
```bash
cd api
go build -o cmdb-api .
```

### 3. 启动服务测试
```bash
./start.sh
```

### 4. 测试4个查询接口

使用 curl 或 Postman 测试：

```bash
# 测试 server-resources-max
curl -X POST http://localhost:8888/api/v1/hardware-proxy/cmdb/v1/server-resources-max \
  -H "Content-Type: application/json" \
  -d '{"beginTime":"2025-01-01","endTime":"2025-01-20","ipList":[]}'

# 测试 server-resources
curl -X POST http://localhost:8888/api/v1/hardware-proxy/cmdb/v1/server-resources \
  -H "Content-Type: application/json" \
  -d '{"beginTime":"2025-01-01","endTime":"2025-01-20","ipList":[]}'

# 测试 cluster-resources-max
curl -X POST http://localhost:8888/api/v1/hardware-proxy/cmdb/v1/cluster-resources-max \
  -H "Content-Type: application/json" \
  -d '{"beginTime":"2025-01-01","endTime":"2025-01-20"}'

# 测试 cluster-resources
curl -X POST http://localhost:8888/api/v1/hardware-proxy/cmdb/v1/cluster-resources \
  -H "Content-Type: application/json" \
  -d '{"beginTime":"2025-01-01","endTime":"2025-01-20"}'
```

验证响应中包含新的百分比字段。

## API 文档

修改完成后的 API 响应示例：

### /api/v1/hardware-proxy/cmdb/v1/server-resources-max

**响应示例**:
```json
{
  "code": 0,
  "msg": "success",
  "data": {
    "rows": [
      {
        "id": 1,
        "poolId": 10,
        "clusterName": "payment-mysql-cluster",
        "groupName": "payment-mysql-group",
        "departmentName": "支付系统",
        "ip": "10.1.5.10",
        "hostName": "db-payment-01",
        "hostType": "physical",
        "totalMemory": 128.0,
        "maxUsedMemory": 96.5,
        "totalDisk": 1000.0,
        "maxUsedDisk": 650.0,
        "cpuCores": 32,
        "maxCpuLoad": 75.5,
        "maxDatetime": "2025-01-20 10:30:00",
        "cpuPercentMax": 75.5,
        "cpuPercentAvg": 65.2,
        "cpuPercentMin": 45.0,
        "memPercentMax": 85.5,
        "memPercentAvg": 72.3,
        "memPercentMin": 55.0,
        "diskPercentMax": 65.0,
        "diskPercentAvg": 60.5,
        "diskPercentMin": 50.0
      }
    ]
  }
}
```

### /api/v1/hardware-proxy/cmdb/v1/cluster-resources-max

**响应示例**:
```json
{
  "code": 0,
  "msg": "success",
  "data": {
    "rows": [
      {
        "clusterName": "payment-mysql-cluster",
        "clusterGroupName": "payment-mysql-group",
        "departmentName": "支付系统",
        "nodeCount": 3,
        "avgCpuLoad": 65.2,
        "avgMemoryUsage": 72.3,
        "avgDiskUsage": 60.5,
        "maxCpuLoad": 75.5,
        "maxMemoryUsage": 85.5,
        "maxDiskUsage": 65.0,
        "cpuPercentMax": 75.5,
        "cpuPercentAvg": 65.2,
        "cpuPercentMin": 45.0,
        "memPercentMax": 85.5,
        "memPercentAvg": 72.3,
        "memPercentMin": 55.0,
        "diskPercentMax": 65.0,
        "diskPercentAvg": 60.5,
        "diskPercentMin": 50.0
      }
    ]
  }
}
```

## 注意事项

1. **ES 数据源确认**: 需要确认 `elasticsearch.HostMetrics` 结构体是否包含 Min 值（MinCPU、MinMemory、MinDisk）。如果没有，可以：
   - 在 ES client 中添加 Min 值的查询
   - 或者在调用 UpsertFromES 时使用0或Avg值代替

2. **数据类型一致性**: 确保所有层（数据库、Model、Proto、API）的字段类型都是 `double/float64`

3. **NULL 值处理**: Model 层使用 `sql.NullFloat64`，在转换为 Proto 和 API 时需要正确处理 NULL 值

4. **聚合查询**: `FindClusterResourcesMax` 方法中可能需要添加对百分比字段的聚合计算（AVG、MAX等）

5. **向后兼容**: 新增字段都设置为 `DEFAULT NULL`，确保旧数据不受影响

6. **索引考虑**: 如果需要频繁按百分比字段查询或排序，考虑添加索引

## 修改文件清单

### 已修改
- ✅ `source/schema.sql`
- ✅ `rpc/internal/model/serverresourcesmodel_gen.go` (重新生成)
- ✅ `rpc/internal/model/serverresourcesmodel.go` (UpsertFromES方法)

### 待修改
- ⏳ `rpc/proto/cmpool.proto` (4个消息类型)
- ⏳ `rpc/internal/model/serverresourcesmodel.go` (查询结构体和SQL)
- ⏳ `rpc/internal/logic/executeessyncbyhostlistlogic.go`
- ⏳ `rpc/internal/scheduler/es_sync_scheduler.go`
- ⏳ `rpc/internal/logic/findserverresourcesmaxlogic.go`
- ⏳ `rpc/internal/logic/findserverresourceslogic.go`
- ⏳ `rpc/internal/logic/findclusterresourcesmaxlogic.go`
- ⏳ `rpc/internal/logic/findclusterresourceslogic.go`
- ⏳ `api/cmdb.api` (4个类型定义)
- ⏳ `api/internal/logic/getserverresourcesmaxlogic.go`
- ⏳ `api/internal/logic/getserverresourceslogic.go`
- ⏳ `api/internal/logic/getclusterresourcesmaxlogic.go`
- ⏳ `api/internal/logic/getclusterresourceslogic.go`

## 下一步行动

建议按照以下顺序完成剩余修改：

1. 修改 Proto 文件并重新生成 RPC 代码
2. 更新 Model 层查询方法和结构体
3. 修改 RPC Logic 层（ES同步和查询接口）
4. 修改 API 定义并重新生成代码
5. 修改 API Logic 层
6. 编译测试
7. 生成 API 文档

每个步骤完成后都应该进行编译验证，确保没有错误再继续下一步。
