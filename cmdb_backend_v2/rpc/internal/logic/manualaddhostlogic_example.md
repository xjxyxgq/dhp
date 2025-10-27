# ManualAddHostLogic.updateHostHardwareInfo 实现说明

## 函数功能
`updateHostHardwareInfo` 函数用于更新主机的硬件信息，它通过调用 `HostsPoolModel.UpdateHostHardwareInfo` 来实现智能的字段更新。

## 实现流程

### 1. 查找现有主机记录
```go
// 根据hostId查找主机记录，获取主机IP
existingHost, err := l.svcCtx.HostsPoolModel.FindOne(l.ctx, uint64(hostId))
if err != nil {
    return fmt.Errorf("查找主机记录失败: %v", err)
}
```

### 2. 构建更新数据结构
```go
// 构建要更新的主机信息，只设置有值的字段
updateHost := &model.HostsPool{
    HostIp: existingHost.HostIp, // 必填字段，用于定位记录
}
```

### 3. 智能字段映射
函数会检查 `ManualHostHardwareInfo` 中每个字段的值，只有非默认值才会被设置到更新结构中：

#### 硬件基本信息
- `DiskSize > 0` → 设置磁盘大小
- `Ram > 0` → 设置内存大小  
- `Vcpus > 0` → 设置CPU核数
- `HostType != ""` → 设置主机类型

#### H3C相关信息
- `H3CId != ""` → 设置H3C ID
- `H3CStatus != ""` → 设置H3C状态
- `IfH3CSync != ""` → 设置H3C同步状态
- `H3CImgId != ""` → 设置H3C镜像ID
- `H3CHmName != ""` → 设置H3C主机名

#### 机架信息
- `LeafNumber != ""` → 设置叶子节点编号
- `RackNumber != ""` → 设置机架编号
- `RackHeight > 0` → 设置机架高度
- `RackStartNumber >= 0` → 设置机架起始位置（允许0值）
- `FromFactor > 0` → 设置规格因子
- `SerialNumber != ""` → 设置序列号

### 4. 执行更新
```go
// 调用UpdateHostHardwareInfo进行更新
err = l.svcCtx.HostsPoolModel.UpdateHostHardwareInfo(l.ctx, updateHost)
if err != nil {
    return fmt.Errorf("更新主机硬件信息失败: %v", err)
}
```

## 使用示例

### 1. 完整硬件信息更新
```go
hardwareInfo := &cmpool.ManualHostHardwareInfo{
    DiskSize:        500,           // 500GB磁盘
    Ram:             32,            // 32GB内存
    Vcpus:           16,            // 16核CPU
    HostType:        "database",    // 数据库服务器
    H3CId:           "h3c_001",     // H3C ID
    H3CStatus:       "running",     // 运行中
    IfH3CSync:       "yes",         // 已同步
    H3CImgId:        "img_123",     // 镜像ID
    H3CHmName:       "db-server",   // H3C主机名
    LeafNumber:      "leaf_01",     // 叶子节点
    RackNumber:      "rack_A01",    // 机架编号
    RackHeight:      42,            // 42U机架
    RackStartNumber: 1,             // 从第1U开始
    FromFactor:      2,             // 2U设备
    SerialNumber:    "SN123456",    // 序列号
}

err := logic.updateHostHardwareInfo(hostId, hardwareInfo)
```

### 2. 部分信息更新
```go
// 只更新硬件核心信息
hardwareInfo := &cmpool.ManualHostHardwareInfo{
    DiskSize: 1000,  // 升级到1TB磁盘
    Ram:      64,    // 升级到64GB内存
    Vcpus:    32,    // 升级到32核CPU
    // 其他字段保持默认值，不会被更新
}

err := logic.updateHostHardwareInfo(hostId, hardwareInfo)
```

### 3. H3C信息更新
```go
// 只更新H3C相关信息
hardwareInfo := &cmpool.ManualHostHardwareInfo{
    H3CId:     "h3c_new_001",
    H3CStatus: "maintenance",
    IfH3CSync: "no",
    // 其他字段保持默认值，不会被更新
}

err := logic.updateHostHardwareInfo(hostId, hardwareInfo)
```

## 关键特性

### 1. 智能字段过滤
- 只有有效值才会被更新到数据库
- 空字符串和0值（除RackStartNumber外）会被忽略
- 避免意外覆盖现有数据

### 2. NULL值正确处理
- 使用 `sql.NullString` 和 `sql.NullInt64` 正确处理数据库NULL值
- 正确设置 `Valid` 标志

### 3. 错误处理
- 完整的错误传播和日志记录
- 清晰的错误消息

### 4. 日志记录
- 成功更新时记录详细信息
- 便于调试和监控

## 注意事项

1. **字段命名**：protobuf生成的字段名遵循Go命名约定（如 `H3CId` 而不是 `H3cId`）
2. **数据验证**：函数会先查找现有主机记录，确保hostId有效
3. **原子性**：所有字段更新在一个事务中完成
4. **灵活性**：可以选择性更新任意字段组合