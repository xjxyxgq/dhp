# CMDB Backend V2

基于 go-zero 框架的配置管理数据库（CMDB）后端服务，采用 RPC + API 的标准微服务架构模式。

## 📑 目录

- [项目概述](#项目概述)
- [项目结构](#项目结构)
- [核心功能](#核心功能)
- [快速开始](#快速开始)
- [开发指南](#开发指南)
- [配置说明](#配置说明)
- [文档索引](#文档索引)

## 项目概述

**🚨 重要声明：前端已完全废弃**

- `cmdb_frontend_v2` 目录中的前端代码已经废弃，前端已迁移到其他项目
- **本项目仅包含后端服务**（RPC + API）
- **禁止对 `cmdb_frontend_v2` 目录下的任何代码进行修改**

### 服务架构

```
┌─────────────┐      HTTP       ┌─────────────┐      gRPC      ┌──────────────┐
│   客户端    │  ────────────>  │  API 服务   │  ───────────>  │  RPC 服务    │
│  (前端等)   │   (8888端口)    │ (接口代理)  │  (8080端口)    │ (业务逻辑)   │
└─────────────┘                 └─────────────┘                └──────────────┘
                                                                      │
                                                                      ├─> MySQL
                                                                      ├─> Redis
                                                                      └─> 外部数据源
```

### 技术栈

- **框架**: go-zero v1.8.4
- **语言**: Go 1.23.1+
- **数据库**: MySQL 5.7+
- **缓存**: Redis (可选)
- **RPC**: gRPC + Protobuf
- **工具**: 自定义 goctl (`/Users/xuguoqiang/LocalOthers/goctl/goctl`)

## 项目结构

```
cmdb_backend_v2/
├── rpc/                          # RPC 服务（业务逻辑层，端口 8080）
│   ├── proto/                    # Protobuf 定义文件
│   │   └── cmpool.proto          # 核心服务接口定义
│   ├── internal/
│   │   ├── config/               # 配置管理
│   │   ├── logic/                # 业务逻辑实现（150+ 个文件）
│   │   ├── server/               # gRPC 服务器
│   │   ├── svc/                  # 服务上下文
│   │   ├── model/                # 数据模型（goctl 生成）
│   │   ├── datasource/           # 外部数据源客户端
│   │   │   ├── elasticsearch/   # ES 客户端
│   │   │   └── cmsys/           # CMSys 客户端
│   │   ├── scheduler/            # 定时任务调度器
│   │   │   ├── task_scheduler.go           # 硬件资源验证调度器
│   │   │   └── external_sync_scheduler.go  # 外部资源同步调度器
│   │   └── global/              # 全局对象管理
│   ├── etc/                     # 配置文件
│   │   └── cmpool.yaml
│   ├── cmpool/                  # 生成的 gRPC 代码
│   └── cmpool.go                # RPC 服务入口
│
├── api/                         # API 服务（HTTP 接口层，端口 8888）
│   ├── cmdb.api                # API 定义文件
│   ├── internal/
│   │   ├── config/             # 配置管理
│   │   ├── handler/            # HTTP 处理器（goctl 生成）
│   │   ├── logic/              # 业务逻辑（调用 RPC）
│   │   ├── svc/                # 服务上下文
│   │   ├── types/              # 类型定义（goctl 生成）
│   │   └── middleware/         # 中间件（认证、日志等）
│   ├── cmpool/                 # RPC 客户端文件（从 rpc/ 复制）
│   ├── etc/
│   │   └── cmdb-api.yaml
│   └── cmdb.go                 # API 服务入口
│
├── docs/                        # 文档目录
│   ├── README.md               # 文档索引
│   ├── api/                    # API 接口文档
│   ├── implementation/         # 实现指南
│   ├── mock/                   # Mock 接口文档
│   ├── changelog/              # 变更记录
│   └── usage/                  # 使用说明
│
├── source/                     # 数据源文件
│   └── schema.sql              # 数据库表结构
│
├── start.sh                    # 服务启动脚本
└── README.md                   # 本文档
```

## 核心功能

### 1. 主机资源管理
- **主机池管理**: 主机基本信息、应用部署信息
- **集群组管理**: 数据库集群组信息维护
- **资源监控**: CPU、内存、磁盘等监控数据采集和查询

### 2. 外部资源同步 ⭐
统一的外部资源同步框架，支持多种数据源：

#### 支持的数据源
- **Elasticsearch (ES)**: 监控指标数据同步
- **CMSys**: 主机负载数据同步

#### 核心功能
- ✅ **定时任务配置**: 创建、更新、删除、启用/禁用定时同步任务
- ✅ **手动同步执行**: 支持全量同步、按IP列表同步、按文件同步
- ✅ **执行日志查询**: 查看同步执行历史和详情
- ✅ **统一调度器**: `ExternalSyncScheduler` 统一管理 ES 和 CMSys 定时任务
- ✅ **Cron 验证**: 6字段格式（秒 分 时 日 月 周）验证，创建时即时检查

#### 相关接口
```
POST   /api/cmdb/v1/external-sync-tasks              # 创建同步任务
PUT    /api/cmdb/v1/external-sync-tasks/:task_id     # 更新同步任务
DELETE /api/cmdb/v1/external-sync-tasks/:task_id     # 删除同步任务
GET    /api/cmdb/v1/external-sync-tasks              # 查询任务列表
GET    /api/cmdb/v1/external-sync-tasks/:task_id     # 查询任务详情
PUT    /api/cmdb/v1/external-sync-tasks/enable       # 启用/禁用任务
POST   /api/cmdb/v1/external-sync/execute            # 执行同步（按IP列表）
POST   /api/cmdb/v1/external-sync/execute-file       # 执行同步（按文件）
POST   /api/cmdb/v1/external-sync/execute-full       # 执行全量同步
GET    /api/cmdb/v1/external-sync/execution-logs     # 查询执行日志
GET    /api/cmdb/v1/external-sync/execution-detail/:id  # 查询执行详情
```

### 3. 定时硬件资源验证
- **验证任务管理**: CPU/内存/磁盘资源验证任务配置
- **定时调度执行**: 基于 Cron 表达式的定时验证
- **验证历史查询**: 执行历史和详情查看

相关接口：
```
POST   /api/v1/hardware-proxy/cmdb/v1/scheduled-tasks       # 创建验证任务
PUT    /api/v1/hardware-proxy/cmdb/v1/scheduled-tasks/:id   # 更新验证任务
DELETE /api/v1/hardware-proxy/cmdb/v1/scheduled-tasks/:id   # 删除验证任务
GET    /api/v1/hardware-proxy/cmdb/v1/scheduled-tasks       # 查询任务列表
```

### 4. 数据分析与报告
- **资源分析**: 集群资源使用情况统计分析
- **告警预测**: 磁盘满预测、资源告警
- **报告生成**: 集群组报告、IDC 报告

### 5. 认证与授权
- **用户登录**: 支持本地登录和 CAS 单点登录
- **Token 认证**: JWT Token 验证
- **会话管理**: 用户会话维护

## 快速开始

### 前置条件
- Go 1.23.1+
- MySQL 5.7+
- Redis (可选)

### 1. 数据库初始化

```bash
# 创建数据库
mysql -u root -p -e "CREATE DATABASE cmdb CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;"

# 导入表结构
mysql -u root -p cmdb < source/schema.sql
```

### 2. 配置文件

修改配置文件中的数据库连接信息：

**RPC 服务配置** (`rpc/etc/cmpool.yaml`):
```yaml
Name: cmpool
ListenOn: 0.0.0.0:8080
DataSource: root:password@tcp(localhost:3306)/cmdb?charset=utf8mb4&parseTime=True&loc=Local

# Elasticsearch 配置
ESDataSource:
  DefaultEndpoint: "http://localhost:9200"
  DefaultIndexPattern: "metricbeat-*"

# CMSys 配置
CMSysDataSource:
  BaseURL: "http://cmsys.example.com"
  Username: "your_username"
  Password: "your_password"
```

**API 服务配置** (`api/etc/cmdb-api.yaml`):
```yaml
Name: cmdb-api
Host: 0.0.0.0
Port: 8888

CmpoolRpc:
  Endpoints:
    - 127.0.0.1:8080
```

### 3. 启动服务

**使用启动脚本（推荐）**:
```bash
./start.sh
```

**手动启动**:
```bash
# 1. 先启动 RPC 服务
cd rpc
go run cmpool.go -f etc/cmpool.yaml

# 2. 再启动 API 服务
cd api
go run cmdb.go -f etc/cmdb-api.yaml
```

### 4. 测试接口

```bash
# 查询主机池
curl http://localhost:8888/api/cmdb/v1/get_hosts_pool_detail

# 查询集群组
curl http://localhost:8888/api/cmdb/v1/cluster-groups

# 查询外部同步任务（需要认证）
curl -H "Authorization: Bearer YOUR_TOKEN" \
  http://localhost:8888/api/cmdb/v1/external-sync-tasks?data_source=cmsys
```

## 开发指南

### 新增 RPC 接口的完整流程

1. **定义 Protobuf 接口** (`rpc/proto/cmpool.proto`):
```protobuf
service Cmpool {
  rpc YourNewMethod(YourRequest) returns (YourResponse);
}

message YourRequest {
  string field1 = 1;
}

message YourResponse {
  bool success = 1;
  string message = 2;
}
```

2. **生成 RPC 代码**:
```bash
cd rpc
/Users/xuguoqiang/LocalOthers/goctl/goctl rpc protoc proto/cmpool.proto --go_out=. --go-grpc_out=. --zrpc_out=.
```

3. **🚨 重要：复制客户端文件到 API 模块**:
```bash
cp cmpool/cmpool.pb.go ../api/cmpool/
cp cmpool/cmpool_grpc.pb.go ../api/cmpool/
```

4. **实现 RPC Logic**:
在 `rpc/internal/logic/` 目录下编写业务逻辑。

5. **添加 API 接口定义** (`api/cmdb.api`):
```go
type YourRequest {
    Field1 string `json:"field1"`
}

type YourResponse {
    Success bool   `json:"success"`
    Message string `json:"message"`
}

@server(
    group: your_group
)
service cmdb-api {
    @handler YourHandler
    post /api/cmdb/v1/your-path (YourRequest) returns (YourResponse)
}
```

6. **生成 API 代码**:
```bash
cd api
/Users/xuguoqiang/LocalOthers/goctl/goctl api go -api cmdb.api -dir .
```

7. **实现 API Logic**:
在 `api/internal/logic/` 中调用 RPC 服务。

### 数据库模型生成

当需要修改数据库结构时：

```bash
# 1. 更新 source/schema.sql
vim source/schema.sql

# 2. 重新生成模型
cd rpc
/Users/xuguoqiang/LocalOthers/goctl/goctl model mysql ddl \
  -src="../source/schema.sql" \
  -dir="internal/model" \
  --style=gozero
```

### Cron 表达式格式

所有定时任务使用 **6 字段格式**：

```
秒 分 时 日 月 周
```

示例：
- `0 0 2 * * *` - 每天凌晨 2 点执行
- `0 */30 * * * *` - 每 30 分钟执行一次
- `0 0 9 * * 1` - 每周一上午 9 点执行

## 配置说明

### RPC 服务完整配置示例

```yaml
Name: cmpool
ListenOn: 0.0.0.0:8080
Mode: dev  # dev/test/pro

# 数据库配置
DataSource: root:password@tcp(localhost:3306)/cmdb?charset=utf8mb4&parseTime=True&loc=Local

# Redis 缓存配置（可选）
Cache:
  - Host: localhost:6379
    Pass: ""
    Type: node

# Elasticsearch 数据源配置
ESDataSource:
  DefaultEndpoint: "http://localhost:9200"
  DefaultIndexPattern: "metricbeat-*"
  Username: ""
  Password: ""

# CMSys 数据源配置
CMSysDataSource:
  BaseURL: "http://cmsys.example.com"
  Username: "your_username"
  Password: "your_password"
  MockMode: false  # 是否使用 Mock 模式
```

### API 服务完整配置示例

```yaml
Name: cmdb-api
Host: 0.0.0.0
Port: 8888
Mode: dev

# RPC 服务连接配置
CmpoolRpc:
  Endpoints:
    - 127.0.0.1:8080
  Timeout: 10000  # 超时时间（毫秒）

# JWT 认证配置
Auth:
  AccessSecret: your-secret-key
  AccessExpire: 86400  # 24小时
```

## 文档索引

详细文档请参考 [docs/README.md](docs/README.md)

### API 文档
- [统一外部资源同步 API](docs/api/EXTERNAL_RESOURCE_SYNC_API.md) - 支持 ES 和 CMSys
- [资源查询 API](docs/api/API_RESOURCE_QUERY_DOCUMENTATION.md)

### 实现指南
- [ES 同步实现指南](docs/implementation/ES_SYNC_IMPLEMENTATION_GUIDE.md)
- [ES 数据字段映射](docs/implementation/ES_SYNC_DATA_MAPPING.md)

### 使用说明
- [CMSys 数据同步使用说明](docs/usage/usage_data_sync_from_cmsys.md)

## 注意事项

1. **服务启动顺序**: RPC 服务必须先于 API 服务启动
2. **代码生成工具**: 必须使用自定义 goctl (`/Users/xuguoqiang/LocalOthers/goctl/goctl`)
3. **客户端文件同步**: 每次重新生成 RPC 代码后，必须将 `rpc/cmpool/*.pb.go` 文件复制到 `api/cmpool/`
4. **生产环境**: 请修改默认密码和敏感配置
5. **Cron 格式**: 所有 Cron 表达式必须是 6 字段格式，创建时会验证

## 技术亮点

- ✅ **统一数据源接口**: ES 和 CMSys 使用统一的 API 和 RPC 接口
- ✅ **智能调度器**: 统一的 `ExternalSyncScheduler` 支持多数据源定时任务
- ✅ **Cron 验证**: 任务创建时即时验证 Cron 表达式，避免运行时错误
- ✅ **类型安全**: 基于 Protobuf 和 goctl 生成的类型安全代码
- ✅ **模块化设计**: RPC 和 API 完全分离，支持独立部署
- ✅ **缓存支持**: 数据模型自动生成 Redis 缓存逻辑

## 代码提交

使用项目提供的 git 提交脚本：

```bash
/Users/xuguoqiang/SynologyDrive/Backup/MI_office_notebook/D/myworkspace/nucc_workspace/program/src/nucc.com/cmpool_cursor/tools/gitcommiter/git_commit_and_tag.sh -m "提交信息"
```

## License

内部项目，保留所有权利。
