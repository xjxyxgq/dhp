# HostsPoolModel.UpdateHostHardwareInfo 使用示例

## 函数签名
```go
UpdateHostHardwareInfo(ctx context.Context, hostInfo *HostsPool) error
```

## HostsPool 结构体（相关字段）
```go
type HostsPool struct {
    Id              uint64         `db:"id"`
    HostName        string         `db:"host_name"`         // 主机名
    HostIp          string         `db:"host_ip"`           // 主机IP （必填）
    HostType        sql.NullString `db:"host_type"`         // 主机类型
    H3cId           sql.NullString `db:"h3c_id"`            // H3C ID
    H3cStatus       sql.NullString `db:"h3c_status"`        // H3C状态
    DiskSize        sql.NullInt64  `db:"disk_size"`         // 磁盘大小(GB)
    Ram             sql.NullInt64  `db:"ram"`               // 内存大小(GB)
    Vcpus           sql.NullInt64  `db:"vcpus"`             // CPU核数
    IfH3cSync       sql.NullString `db:"if_h3c_sync"`       // 是否H3C同步
    H3cImgId        sql.NullString `db:"h3c_img_id"`        // H3C镜像ID
    H3cHmName       sql.NullString `db:"h3c_hm_name"`       // H3C主机名
    LeafNumber      sql.NullString `db:"leaf_number"`       // 叶子节点编号
    RackNumber      sql.NullString `db:"rack_number"`       // 机架号
    RackHeight      sql.NullInt64  `db:"rack_height"`       // 机架高度
    RackStartNumber sql.NullInt64  `db:"rack_start_number"` // 机架起始位置
    FromFactor      sql.NullInt64  `db:"from_factor"`       // 规格因子
    SerialNumber    sql.NullString `db:"serial_number"`     // 序列号
    IsDelete        sql.NullString `db:"is_delete"`         // 是否删除标记
    // ... 其他字段
}
```

## 使用示例

### 1. 只更新主机名
```go
err := hostsPoolModel.UpdateHostHardwareInfo(ctx, &model.HostsPool{
    HostIp:   "192.168.1.100",
    HostName: "new-hostname", // 非空字符串会被更新
})
```

### 2. 只更新硬件信息
```go
err := hostsPoolModel.UpdateHostHardwareInfo(ctx, &model.HostsPool{
    HostIp:   "192.168.1.100",
    DiskSize: sql.NullInt64{Int64: 500, Valid: true},  // Valid=true且值>0时更新
    Ram:      sql.NullInt64{Int64: 16, Valid: true},   // 16GB
    Vcpus:    sql.NullInt64{Int64: 8, Valid: true},    // 8核
})
```

### 3. 更新H3C相关信息
```go
err := hostsPoolModel.UpdateHostHardwareInfo(ctx, &model.HostsPool{
    HostIp:     "192.168.1.100",
    H3cId:      sql.NullString{String: "h3c_001", Valid: true},
    H3cStatus:  sql.NullString{String: "running", Valid: true},
    IfH3cSync:  sql.NullString{String: "yes", Valid: true},
    H3cImgId:   sql.NullString{String: "img_123", Valid: true},
    H3cHmName:  sql.NullString{String: "h3c_hostname", Valid: true},
})
```

### 4. 更新机架信息
```go
err := hostsPoolModel.UpdateHostHardwareInfo(ctx, &model.HostsPool{
    HostIp:          "192.168.1.100",
    LeafNumber:      sql.NullString{String: "leaf_01", Valid: true},
    RackNumber:      sql.NullString{String: "rack_A01", Valid: true},
    RackHeight:      sql.NullInt64{Int64: 42, Valid: true},      // 42U机架
    RackStartNumber: sql.NullInt64{Int64: 1, Valid: true},       // 从第1U开始
    FromFactor:      sql.NullInt64{Int64: 2, Valid: true},       // 2U设备
    SerialNumber:    sql.NullString{String: "SN123456", Valid: true},
})
```

### 5. 综合更新（包含多种字段）
```go
err := hostsPoolModel.UpdateHostHardwareInfo(ctx, &model.HostsPool{
    HostIp:     "192.168.1.100",
    HostName:   "production-server",
    HostType:   sql.NullString{String: "database", Valid: true},
    DiskSize:   sql.NullInt64{Int64: 1000, Valid: true}, // 1TB
    Ram:        sql.NullInt64{Int64: 32, Valid: true},   // 32GB
    Vcpus:      sql.NullInt64{Int64: 16, Valid: true},   // 16核
    RackNumber: sql.NullString{String: "rack_B02", Valid: true},
})
```

### 6. 部分字段更新（其他字段保持不变）
```go
// 只更新RAM，其他字段不变
err := hostsPoolModel.UpdateHostHardwareInfo(ctx, &model.HostsPool{
    HostIp: "192.168.1.100",
    Ram:    sql.NullInt64{Int64: 64, Valid: true}, // 升级到64GB
    // 其他字段使用默认值，不会被更新
})
```

## 智能更新规则

### 字符串字段（sql.NullString）
- `Valid=false` 或 `String=""` 时：不更新
- `Valid=true` 且 `String!=""` 时：更新

### 数字字段（sql.NullInt64）
- `Valid=false` 或 `Int64<=0` 时：不更新（除了 RackStartNumber 允许>=0）
- `Valid=true` 且符合条件时：更新

### 特殊字段
- `HostIp`：必填字段，用于定位要更新的记录
- `HostName`：普通字符串，空字符串时不更新
- `RackStartNumber`：允许>=0的值（0表示机架底部）

## 注意事项
- `HostIp` 字段必填，用于定位要更新的记录
- 所有 `sql.NullString` 和 `sql.NullInt64` 字段都需要正确设置 `Valid` 标志
- 如果没有任何字段需要更新，函数会返回错误
- 函数会自动更新 `update_time` 字段为当前时间
- 更新操作是原子性的，要么全部成功要么全部失败