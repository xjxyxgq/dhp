# CMDB Backend V2 文档索引

本目录包含 CMDB Backend V2 项目的所有技术文档，按功能分类组织。

## 📂 目录结构

```
docs/
├── README.md              # 本文档（文档索引）
├── api/                   # API 接口文档
├── implementation/        # 实现指南与技术文档
├── mock/                  # Mock 接口文档（开发测试用）
├── changelog/             # 重要变更记录
└── usage/                 # 使用说明文档
```

## 📖 API 接口文档 (api/)

正式的 HTTP API 接口文档，描述接口定义、参数、响应格式等。

| 文档 | 描述 | 状态 |
|------|------|------|
| [EXTERNAL_RESOURCE_SYNC_API.md](api/EXTERNAL_RESOURCE_SYNC_API.md) | ⭐ **统一外部资源同步 API**<br>支持 ES 和 CMSys 的统一接口 | ✅ 主文档 |
| [API_RESOURCE_QUERY_DOCUMENTATION.md](api/API_RESOURCE_QUERY_DOCUMENTATION.md) | 资源查询 API 文档 | ✅ 活跃 |
| [ES_SYNC_API_DOCUMENTATION.md](api/ES_SYNC_API_DOCUMENTATION.md) | ⚠️ ES 同步 API（已被统一接口替代） | 📦 存档 |
| [CMSYS_SYNC_API_DOCUMENTATION.md](api/CMSYS_SYNC_API_DOCUMENTATION.md) | ⚠️ CMSys 同步 API（已被统一接口替代） | 📦 存档 |

**说明**：
- ⭐ **推荐使用统一的 EXTERNAL_RESOURCE_SYNC_API** 进行外部资源同步
- ES_SYNC_API 和 CMSYS_SYNC_API 保留作为历史参考，新功能请使用统一接口

## 🔧 实现指南 (implementation/)

技术实现文档，包含架构设计、数据映射、实现细节等。

| 文档 | 描述 | 状态 |
|------|------|------|
| [ES_SYNC_IMPLEMENTATION_GUIDE.md](implementation/ES_SYNC_IMPLEMENTATION_GUIDE.md) | ES 同步功能实现指南 | ✅ 活跃 |
| [ES_SYNC_IMPLEMENTATION_SUMMARY.md](implementation/ES_SYNC_IMPLEMENTATION_SUMMARY.md) | ES 同步功能实现总结 | ✅ 活跃 |
| [ES_SYNC_DATA_MAPPING.md](implementation/ES_SYNC_DATA_MAPPING.md) | ES 数据字段映射说明 | ✅ 活跃 |
| [RESOURCE_PERCENT_IMPLEMENTATION_GUIDE.md](implementation/RESOURCE_PERCENT_IMPLEMENTATION_GUIDE.md) | 资源百分比实现指南 | ✅ 活跃 |

## 🧪 Mock 接口文档 (mock/)

开发测试用的 Mock 接口文档，模拟外部数据源。

| 文档 | 描述 | 用途 |
|------|------|------|
| [MOCK_ES_GUIDE.md](mock/MOCK_ES_GUIDE.md) | Mock Elasticsearch 接口指南 | 开发测试 |
| [MOCK_ES_GROUP_QUERY.md](mock/MOCK_ES_GROUP_QUERY.md) | Mock ES 分组查询文档 | 开发测试 |
| [MOCK_ES_TEST_RESULTS.md](mock/MOCK_ES_TEST_RESULTS.md) | Mock ES 测试结果 | 开发测试 |
| [CMSYS_MOCK_INTERFACES.md](mock/CMSYS_MOCK_INTERFACES.md) | CMSys Mock 接口文档 | 开发测试 |
| [CMSYS_MOCK_IMPLEMENTATION_SUMMARY.md](mock/CMSYS_MOCK_IMPLEMENTATION_SUMMARY.md) | CMSys Mock 实现总结 | 开发测试 |

## 📝 变更记录 (changelog/)

重要的数据库结构变更、字段重命名等记录。

| 文档 | 描述 | 日期 |
|------|------|------|
| [SERVER_RESOURCES_FIELD_RENAME.md](changelog/SERVER_RESOURCES_FIELD_RENAME.md) | server_resources 表字段重命名<br>`date_time` → `mon_date` | 2024 |
| [HOSTNAME_FIELD_ENHANCEMENT.md](changelog/HOSTNAME_FIELD_ENHANCEMENT.md) | 主机名字段增强变更 | 2024 |

## 📚 使用说明 (usage/)

面向用户的功能使用文档。

| 文档 | 描述 | 目标用户 |
|------|------|----------|
| [usage_data_sync_from_cmsys.md](usage/usage_data_sync_from_cmsys.md) | 从 CMSys 同步数据使用说明 | 运维人员 |
| [usage_data_async.md](usage/usage_data_async.md) | 异步数据同步使用说明 | 运维人员 |

## 🗂️ 根目录文档

| 文档 | 描述 | 状态 |
|------|------|------|
| [server_resources_refactor.md](server_resources_refactor.md) | server_resources 表重构说明 | ✅ 活跃 |
| [database_refactoring_plan.md](database_refactoring_plan.md) | 数据库重构计划 | ⚠️ 计划文档 |

## 🚀 快速查找

### 我想要...

#### 调用外部资源同步接口
→ [统一外部资源同步 API](api/EXTERNAL_RESOURCE_SYNC_API.md)

#### 实现新的 ES 同步逻辑
→ [ES 同步实现指南](implementation/ES_SYNC_IMPLEMENTATION_GUIDE.md)
→ [ES 数据字段映射](implementation/ES_SYNC_DATA_MAPPING.md)

#### 了解 CMSys 如何同步数据
→ [CMSys 数据同步使用说明](usage/usage_data_sync_from_cmsys.md)
→ [CMSys Mock 接口文档](mock/CMSYS_MOCK_INTERFACES.md)

#### 查看数据库变更历史
→ [变更记录 (changelog/)](changelog/)

#### 开发时使用 Mock 数据
→ [Mock 接口文档 (mock/)](mock/)

## 📋 文档维护规范

### 1. 文档分类原则

**✅ 应该保留的文档**:
- 正式的 API 接口文档
- 实现指南和技术设计文档
- 变更记录和迁移文档
- 使用说明和操作手册

**❌ 不应保留的文档**:
- 临时性的诊断报告（DIAGNOSIS、FIX、REPORT 等）
- 已完成的项目进度文档
- 个人笔记和草稿
- 过时的实验性文档

### 2. 文档更新规范

| 事件 | 操作 |
|------|------|
| API 接口变更 | 必须同步更新 API 文档 |
| 重大架构变更 | 更新实现指南 |
| Mock 接口变更 | 更新 Mock 文档 |
| 数据库/字段变更 | 在 changelog/ 中记录 |
| 功能使用方式变更 | 更新使用说明 |

### 3. 文档状态标记

- ✅ **活跃**: 当前正在使用和维护的文档
- 📦 **存档**: 历史参考文档，已有替代方案
- ⚠️ **计划**: 计划文档，可能未实施或已过时
- ❌ **废弃**: 已废弃的文档，应考虑删除

### 4. 文档编写规范

1. **使用 Markdown 格式**
2. **添加目录结构**（较长文档）
3. **包含代码示例**（API 文档）
4. **标注版本和日期**（变更记录）
5. **使用表格和图表**（提高可读性）

## 🔗 相关资源

- [项目主 README](../README.md) - 项目概述、快速开始
- [CLAUDE.md](../CLAUDE.md) - Claude Code Agent 使用指南
- [源代码目录](../rpc/) - RPC 服务源代码
- [API 定义文件](../api/cmdb.api) - API 接口定义

## 📮 文档贡献

如需添加或更新文档，请：

1. 确定文档应该放在哪个目录
2. 遵循相应的文档格式和命名规范
3. 更新本文档索引
4. 提交时使用清晰的 commit 信息

## ⚙️ 文档生成工具

- **API 文档**: 基于 `cmdb.api` 和 `cmpool.proto` 定义
- **数据库文档**: 基于 `source/schema.sql` 生成
- **代码注释**: 使用 godoc 格式

---

**最后更新**: 2025-10-22
**维护者**: CMDB 开发团队
