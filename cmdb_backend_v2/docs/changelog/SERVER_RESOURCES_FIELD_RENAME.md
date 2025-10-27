# server_resources 表字段重命名记录

## 变更概述

将 `server_resources` 表的 `date_time` 字段（datetime类型）重命名为 `mon_date`（date类型），同时更新所有相关代码。

## 变更原因

1. **语义更清晰**：`mon_date` 更准确地表达"监控日期"的含义
2. **数据类型优化**：从 `datetime` 改为 `date`，每天只存储一条记录，不需要时分秒
3. **唯一索引优化**：使用 `uk_ip_mon_date (ip, mon_date)` 唯一索引，确保每个IP每天只有一条记录

## 数据库变更

### Schema 定义
```sql
-- 字段定义
`mon_date` date NOT NULL COMMENT '监控日期（每天一条记录）',

-- 唯一键
UNIQUE KEY `uk_ip_mon_date` (`ip`, `mon_date`),

-- 索引
KEY `idx_mon_date` (`mon_date`),
```

### 原字段定义
```sql
`date_time` datetime DEFAULT CURRENT_TIMESTAMP COMMENT '监控时间',
KEY `idx_date_time` (`date_time`),
```

## 代码变更清单

### 1. Model 层 (serverresourcesmodel.go) ✅

#### 结构体字段修改
- **ServerResourceRow**: `DateTime string db:"date_time"` → `MonDate string db:"mon_date"`
- **DiskPredictionData**: `Datetime string db:"date_time"` → `MonDate string db:"mon_date"`
- **ClusterMemberResourceData**: `DateTime string db:"datetime"` → `MonDate string db:"mon_date"`

#### SQL 查询修改（共20+处）
所有 SQL 查询中的字段引用已从 `sr.date_time` 更新为 `sr.mon_date`：

1. **FindDistinctIPsInTimeRange**: WHERE 条件时间过滤
2. **FindServerResourceMax**: WHERE 时间过滤、MAX(date_time) 聚合
3. **FindServerResourcesWithFilter**: SELECT 字段、WHERE 时间过滤、ORDER BY
4. **FindServerResourcesWithClusterFilter**: SELECT 字段、WHERE 时间过滤
5. **FindClusterResources**: SELECT 别名、WHERE 时间过滤、ORDER BY
6. **FindClusterResourcesMax**: 两个子查询的时间过滤和 MAX 聚合
7. **FindClusterMemberResources**: SELECT 别名
8. **FindDiskPredictionDataWithFilter**: SELECT 别名、WHERE 时间过滤、ORDER BY
9. **UpsertFromES**: INSERT 字段名、UPDATE 字段名（使用 CURDATE() 而非 NOW()）

**关键变更**：
- UpsertFromES 方法：`NOW()` → `CURDATE()`，因为现在使用 date 类型
- 所有别名：`as datetime` → `as mon_date`

### 2. Logic 层修改 ✅

#### getserverresourcelogic.go
- 第81行：`Datetime: row.DateTime` → `Datetime: row.MonDate`
- 第236-237行：排序逻辑 `rows[i].DateTime` → `rows[i].MonDate`
- 第71行、196行：注释更新，`date_time` → `mon_date`

#### getserverresourcemaxlogic.go
- 第37行：WHERE 条件 `date_time BETWEEN` → `mon_date BETWEEN`

#### getclusterresourceslogic.go
- 第50行：字段映射 `DateTime: data.DateTime` → `DateTime: data.MonDate`

#### getdiskpredictionlogic.go
- 第48行：WHERE 条件 `sr.date_time BETWEEN` → `sr.mon_date BETWEEN`
- 第273-274行：时间解析格式 `2006-01-02 15:04:05` → `2006-01-02`
- 时间字段引用：`dataList[0].Datetime` → `dataList[0].MonDate`

#### loadservermetricsfromcsvlogic.go
- 第307行：字段赋值 `DateTime: time.Now()` → `MonDate: time.Now()`
- 注释：`当前时间` → `当前日期`

#### csv_loader.go (datasource)
- 第88行：字段赋值 `DateTime: time.Now()` → `MonDate: time.Now()`

### 3. 生成的代码 (serverresourcesmodel_gen.go) ✅

已由 goctl 重新生成：
```go
MonDate sql.NullTime `db:"mon_date"`
```

## 影响范围

### 直接影响的表
- `server_resources` - 主表

### 影响的查询接口
1. `GetServerResource` - 主机资源用量查询
2. `GetServerResourceMax` - 主机资源最大值查询
3. `GetClusterResources` - 集群资源详情查询
4. `GetClusterResourcesMax` - 集群资源最大值查询
5. `GetDiskPrediction` - 磁盘预测查询
6. `LoadServerMetricsFromCsv` - CSV导入
7. `UpsertFromES` - ES数据同步

### 未修改的文件（不涉及该字段）
- API 层文件（仅调用 RPC，不直接访问数据库）
- 其他 Logic 层文件（不操作 server_resources 表）
- 调度器、中间件等（不涉及该字段）

## 测试建议

### 1. 数据库层测试
- 验证唯一索引 `uk_ip_mon_date` 是否生效
- 测试同一IP同一天插入多条记录（应失败）
- 测试 CURDATE() 在 UpsertFromES 中是否正常工作

### 2. 功能测试
- **主机资源查询**：
  - 时间范围过滤是否正常
  - 返回的 mon_date 格式是否正确（YYYY-MM-DD）

- **集群资源查询**：
  - 时间聚合是否正确
  - MAX(mon_date) 是否返回正确日期

- **磁盘预测**：
  - 时间跨度计算是否正确（date vs datetime格式）
  - 预测算法是否受影响

- **CSV导入**：
  - 导入时 mon_date 是否设置为当前日期

- **ES同步**：
  - UpsertFromES 是否正确设置日期
  - 重复同步是否正确覆盖当天数据

### 3. 性能测试
- 查询性能是否有提升（date vs datetime）
- 索引效率是否改善

## 回滚方案

如需回滚，执行以下步骤：

1. **数据库回滚**：
```sql
ALTER TABLE server_resources CHANGE COLUMN mon_date date_time datetime DEFAULT CURRENT_TIMESTAMP;
ALTER TABLE server_resources DROP INDEX uk_ip_mon_date;
ALTER TABLE server_resources ADD INDEX idx_date_time (date_time);
```

2. **代码回滚**：
   - 恢复所有 `MonDate` → `DateTime`
   - 恢复所有 `mon_date` → `date_time`
   - 恢复 UpsertFromES 中的 `CURDATE()` → `NOW()`
   - 恢复时间解析格式 `2006-01-02` → `2006-01-02 15:04:05`

3. **重新生成代码**：
```bash
cd cmdb_backend_v2/rpc
/Users/xuguoqiang/LocalOthers/goctl/goctl model mysql datasource -url="root:password@tcp(localhost:3306)/cmdb" -table="server_resources" -dir=./internal/model
```

## 执行时间
2025-10-20

## 执行人
Claude (AI Assistant)

## 审核状态
待审核

## 附加说明

1. **时间格式差异**：
   - 原格式：`2006-01-02 15:04:05` (datetime)
   - 新格式：`2006-01-02` (date)

2. **兼容性**：
   - Protobuf 接口未变更（仍使用 `DateTime` 字段名）
   - 前端无需修改（前端已废弃）

3. **数据迁移**：
   - 如果表中已有数据，需要先执行数据迁移SQL
   - 建议使用 `DATE(date_time)` 转换现有数据

4. **最佳实践**：
   - 使用 `CURDATE()` 而非 `NOW()` 插入当前日期
   - 唯一索引确保数据一致性
   - 减少存储空间和索引开销
